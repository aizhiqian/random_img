import time
import random
import requests
import concurrent.futures
from pathlib import Path

SCRIPT_DIR = Path(__file__).resolve().parent
PROJECT_DIR = SCRIPT_DIR.parent
DATA_DIR = PROJECT_DIR / "data"

INPUT_FILES = sorted(
    (path for path in DATA_DIR.rglob("*.txt") if path.is_file()),
    key=lambda path: path.as_posix().lower(),
)
INVALID_FILE = SCRIPT_DIR / "invalid_urls.txt"
RATE_LIMITED_FILE = SCRIPT_DIR / "rate_limited_urls.txt"

TIMEOUT = 10
THREADS = 10                   # 并发过高更容易429
MAX_RETRIES_429 = 4            # 429重试次数
BACKOFF_BASE = 1               # 指数退避基数秒

HEADERS = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
                  "AppleWebKit/537.36 (KHTML, like Gecko) "
                  "Chrome/145.0.0.0 Safari/537.36"
}

def clean_url(line):
    """
    支持：
    1) https://example.com/a.jpg
    2) 1. https://example.com/a.jpg
    """
    line = line.strip()
    if not line:
        return None

    parts = line.split(" ", 1)
    if len(parts) == 2:
        prefix = parts[0]
        if prefix.replace(".", "", 1).isdigit():
            return parts[1].strip()

    return line

def load_urls_from_files(file_paths):
    items = []
    for p in file_paths:
        if not p.exists():
            print(f"警告：文件不存在，已跳过 -> {p}")
            continue

        with open(p, 'r', encoding='utf-8') as f:
            for idx, line in enumerate(f, start=1):
                url = clean_url(line)
                if url:
                    items.append({
                        "file": p,
                        "line_no": idx,
                        "url": url
                    })
    return items

def remove_invalid_urls(invalid_items):
    """从原始文本文件中删除已确认无效的链接行。"""
    invalid_lines_by_file = {}
    for item in invalid_items:
        invalid_lines_by_file.setdefault(item["file"], {})[item["line_no"]] = item["url"]

    removed_total = 0
    skipped_total = 0

    for file_path, invalid_lines in invalid_lines_by_file.items():
        with file_path.open("r", encoding="utf-8", newline="") as f:
            lines = f.readlines()

        kept_lines = []
        removed_in_file = 0
        for line_no, line in enumerate(lines, start=1):
            expected_url = invalid_lines.get(line_no)
            if expected_url is not None and clean_url(line) == expected_url:
                removed_in_file += 1
                continue
            kept_lines.append(line)

        skipped_in_file = len(invalid_lines) - removed_in_file
        if removed_in_file:
            with file_path.open("w", encoding="utf-8", newline="") as f:
                f.writelines(kept_lines)
            relative_path = file_path.relative_to(PROJECT_DIR)
            print(f"已从 {relative_path} 删除 {removed_in_file} 条无效链接。")

        if skipped_in_file:
            relative_path = file_path.relative_to(PROJECT_DIR)
            print(f"警告：{relative_path} 有 {skipped_in_file} 条无效链接未删除，文件内容可能已变更。")

        removed_total += removed_in_file
        skipped_total += skipped_in_file

    return removed_total, skipped_total

def check_media_url(url):
    url = url.strip()
    if not url:
        return "invalid", "🚫 Empty URL"

    timeout_retry = 1

    for attempt in range(MAX_RETRIES_429 + 1):
        try:
            with requests.get(
                url,
                timeout=TIMEOUT,
                headers=HEADERS,
                allow_redirects=True,
                stream=True,
            ) as resp:
                if resp.status_code == 429:
                    if attempt < MAX_RETRIES_429:
                        retry_after = resp.headers.get("Retry-After")
                        if retry_after and retry_after.isdigit():
                            sleep_s = int(retry_after)
                        else:
                            sleep_s = BACKOFF_BASE * (2 ** attempt) + random.uniform(0, 0.6)
                        time.sleep(sleep_s)
                        continue
                    else:
                        return "rate_limited", "⚠️ 429 Too Many Requests"

                if resp.status_code == 200:
                    content_type = (resp.headers.get("Content-Type") or "").lower()
                    if content_type.startswith(("video/", "image/", "application/octet-stream")):
                        return "valid", "✅ Media"
                    return "invalid", f"🚫 Not a media ({content_type})"

                return "invalid", f"❌ Status Code: {resp.status_code}"

        except requests.exceptions.Timeout:
            if timeout_retry > 0:
                timeout_retry -= 1
                time.sleep(0.8)
                continue
            return "invalid", "⏱️ Timeout"

        except requests.exceptions.RequestException as e:
            return "invalid", f"🌐 RequestException: {e}"

        except Exception as e:
            return "invalid", f"💥 Exception: {e}"

    return "invalid", "❓ Unknown error"

def main():
    print(f"脚本目录: {SCRIPT_DIR}")
    print("读取输入文件中...")

    url_items = load_urls_from_files(INPUT_FILES)
    total = len(url_items)

    if total == 0:
        print("没有可检查的链接（文件为空或路径无效）。")
        open(INVALID_FILE, "w", encoding="utf-8").close()
        open(RATE_LIMITED_FILE, "w", encoding="utf-8").close()
        return

    print(f"共发现 {total} 个链接，开始检查（线程数={THREADS}）...")

    invalid_rows = []
    invalid_items = []
    rate_limited_rows = []
    valid_count = 0

    with concurrent.futures.ThreadPoolExecutor(max_workers=THREADS) as executor:
        future_to_item = {
            executor.submit(check_media_url, item["url"]): item
            for item in url_items
        }

        for i, future in enumerate(concurrent.futures.as_completed(future_to_item), start=1):
            item = future_to_item[future]
            src_file = item["file"]
            relative_src_file = src_file.relative_to(PROJECT_DIR)
            line_no = item["line_no"]
            url = item["url"]

            category, msg = future.result()

            short_url = (url[:70] + "..") if len(url) > 70 else url

            if category == "valid":
                valid_count += 1
                print(f"[{i}/{total}] ✅ {short_url} - {msg}")
            elif category == "rate_limited":
                print(f"[{i}/{total}] ⚠️ {short_url} - {msg}")
                rate_limited_rows.append(
                    f"File: {relative_src_file} | Line: {line_no} | URL: {url} | Reason: {msg}"
                )
            else:
                print(f"[{i}/{total}] ❌ {short_url} - {msg}")
                invalid_items.append(item)
                invalid_rows.append(
                    f"File: {relative_src_file} | Line: {line_no} | URL: {url} | Reason: {msg}"
                )

    with open(INVALID_FILE, "w", encoding="utf-8") as f:
        for row in invalid_rows:
            f.write(row + "\n")

    with open(RATE_LIMITED_FILE, "w", encoding="utf-8") as f:
        for row in rate_limited_rows:
            f.write(row + "\n")

    removed_count, skipped_count = remove_invalid_urls(invalid_items)

    print("\n检查完毕：")
    print(f"有效链接: {valid_count}")
    print(f"无效链接: {len(invalid_rows)} -> {INVALID_FILE}")
    print(f"429限流: {len(rate_limited_rows)} -> {RATE_LIMITED_FILE}")
    print(f"已从原文件删除无效链接: {removed_count}")
    if skipped_count:
        print(f"因原文件内容变更而跳过: {skipped_count}")


if __name__ == "__main__":
    main()
