import sys
import os

def deduplicate_file(input_file):
    """
    去除文本文件中的重复行，保持原有顺序，直接覆盖原文件
    """
    try:
        with open(input_file, 'r', encoding='utf-8') as f:
            lines = f.readlines()

        original_count = len(lines)
        print(f"📄 原始行数: {original_count}")

        seen = set()
        unique_lines = []
        duplicate_count = 0

        for line in lines:
            stripped_line = line.strip()
            if stripped_line:
                if stripped_line not in seen:
                    seen.add(stripped_line)
                    unique_lines.append(stripped_line + '\n')
                else:
                    duplicate_count += 1

        with open(input_file, 'w', encoding='utf-8') as f:
            f.writelines(unique_lines)

        unique_count = len(unique_lines)
        print(f"✅ 去重后行数: {unique_count}")
        print(f"🗑️ 删除重复: {duplicate_count} 行")
        print(f"💾 已覆盖原文件: {input_file}")

    except FileNotFoundError:
        print(f"❌ 错误: 找不到文件 {input_file}")
    except Exception as e:
        print(f"❌ 处理错误: {str(e)}")


def deduplicate_path(path):
    """
    支持两种输入：
    1. 文件：仅处理该文件
    2. 目录：递归处理目录下所有 .txt 文件
    """
    if os.path.isfile(path):
        deduplicate_file(path)
        return

    if os.path.isdir(path):
        txt_files = []
        for root, _, files in os.walk(path):
            for name in files:
                if name.lower().endswith('.txt'):
                    txt_files.append(os.path.join(root, name))

        if not txt_files:
            print(f"⚠️  目录下未找到 txt 文件: {path}")
            return

        print(f"📁 找到 {len(txt_files)} 个 txt 文件，开始批量去重...\n")
        for index, txt_file in enumerate(sorted(txt_files), start=1):
            print(f"[{index}/{len(txt_files)}] 处理: {txt_file}")
            deduplicate_file(txt_file)
            print()
        print("🎉 批量去重完成")
        return

    print(f"❌ 错误: 路径不存在 {path}")

def main():
    if len(sys.argv) < 2:
        print("❌ 错误: 必须指定输入文件或目录")
        print("用法: python .\\scripts\\deduplicate.py <文件或目录>")
        print("示例1: python .\\scripts\\deduplicate.py .\\data\\videos\\pc\\cosplay.txt")
        print("示例2: python .\\scripts\\deduplicate.py .\\data\\videos")
        return

    input_path = sys.argv[1]
    print(f"开始去重: {input_path}\n")
    deduplicate_path(input_path)

if __name__ == '__main__':
    main()
