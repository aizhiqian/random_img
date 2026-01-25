def deduplicate_file(input_file):
    """
    去除文本文件中的重复行，保持原有顺序，直接覆盖原文件
    """
    try:
        # 读取文件
        with open(input_file, 'r', encoding='utf-8') as f:
            lines = f.readlines()

        original_count = len(lines)
        print(f"📄 原始行数: {original_count}")

        # 去重，保持顺序
        seen = set()
        unique_lines = []
        duplicate_count = 0

        for line in lines:
            stripped_line = line.strip()
            if stripped_line:  # 跳过空行
                if stripped_line not in seen:
                    seen.add(stripped_line)
                    unique_lines.append(stripped_line + '\n')
                else:
                    duplicate_count += 1

        # 直接覆盖原文件
        with open(input_file, 'w', encoding='utf-8') as f:
            f.writelines(unique_lines)

        unique_count = len(unique_lines)
        print(f"✅ 去重后行数: {unique_count}")
        print(f"🗑️  删除重复: {duplicate_count} 行")
        print(f"💾 已覆盖原文件: {input_file}")

    except FileNotFoundError:
        print(f"❌ 错误: 找不到文件 {input_file}")
    except Exception as e:
        print(f"❌ 处理错误: {str(e)}")

def main():
    import sys

    # 必须指定输入文件
    if len(sys.argv) < 2:
        print("❌ 错误: 必须指定输入文件")
        print("用法: python deduplicate.py <文件名>")
        print("示例: python deduplicate.py output.txt")
        return

    input_file = sys.argv[1]
    print(f"开始去重: {input_file}\n")
    deduplicate_file(input_file)

if __name__ == '__main__':
    main()
