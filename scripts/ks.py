import requests
import json
import time
from urllib.parse import quote, urlparse, urlunparse

def parse_video_url(url, max_retries=3):
    """
    解析单个视频URL，返回video_url，支持重试
    """
    for attempt in range(1, max_retries + 1):
        try:
            api_url = f"https://api.bugpk.com/api/kuaishou?url={quote(url)}"

            response = requests.get(api_url, timeout=30)
            response.raise_for_status()

            data = response.json()

            if data.get('code') == 200 and 'data' in data:
                video_url = data['data'].get('url')
                if video_url:
                    print(f"✅ 解析成功: {url[:50]}")
                    return video_url
                else:
                    print(f"❌ 未找到video_url: {url[:50]}")
                    return None
            else:
                print(f"❌ 解析失败: {url[:50]} - {data.get('msg', '未知错误')}")
                if attempt < max_retries:
                    print(f"   🔄 准备第 {attempt + 1} 次重试...")
                    time.sleep(2)
                    continue
                return None

        except requests.RequestException as e:
            print(f"❌ 请求错误: {url[:50]} - {str(e)}")
            if attempt < max_retries:
                print(f"   🔄 准备第 {attempt + 1} 次重试...")
                time.sleep(2)
                continue
            return None
        except json.JSONDecodeError as e:
            print(f"❌ JSON解析错误: {url[:50]} - {str(e)}")
            if attempt < max_retries:
                print(f"   🔄 准备第 {attempt + 1} 次重试...")
                time.sleep(2)
                continue
            return None
        except Exception as e:
            print(f"❌ 未知错误: {url[:50]} - {str(e)}")
            if attempt < max_retries:
                print(f"   🔄 准备第 {attempt + 1} 次重试...")
                time.sleep(2)
                continue
            return None

    return None

def replace_video_domain(video_url):
    """
    替换video_url的子域名为txmov2.a.kwimgs.com, 并去除查询参数
    """
    try:
        parsed = urlparse(video_url)
        new_netloc = 'txmov2.a.kwimgs.com'
        new_parsed = parsed._replace(netloc=new_netloc, query='', fragment='')
        return urlunparse(new_parsed)
    except Exception as e:
        print(f"⚠️ 域名替换失败: {str(e)}, 返回原URL")
        return video_url

def main():
    input_file = 'ks.txt'
    output_file = 'output.txt'

    # 读取输入文件
    try:
        with open(input_file, 'r', encoding='utf-8') as f:
            urls = [line.strip() for line in f if line.strip()]
    except FileNotFoundError:
        print(f"❌ 错误: 找不到文件 {input_file}")
        return
    except Exception as e:
        print(f"❌ 读取文件错误: {str(e)}")
        return

    if not urls:
        print("❌ 输入文件为空")
        return

    print(f"共找到 {len(urls)} 个URL需要解析\n")

    # 解析所有URL
    video_urls = []
    for i, url in enumerate(urls, 1):
        print(f"[{i}/{len(urls)}] 正在解析...")
        video_url = parse_video_url(url)
        if video_url:
            video_urls.append(video_url)

        # 添加延迟，避免请求过快
        if i < len(urls):
            time.sleep(1)

    # 写入输出文件
    try:
        with open(output_file, 'w', encoding='utf-8') as f:
            for video_url in video_urls:
                modified_url = replace_video_domain(video_url)
                f.write(modified_url + '\n')
        print(f"\n成功! 共解析 {len(video_urls)}/{len(urls)} 个视频URL")
        print(f"结果已保存到 {output_file}")
    except Exception as e:
        print(f"❌ 写入文件错误: {str(e)}")

if __name__ == '__main__':
    main()
