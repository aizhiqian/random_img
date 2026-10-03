# scripts

用于维护 `data` 目录中的媒体链接。以下命令均在项目根目录执行

需先安装依赖：

```powershell
pip install requests
```

## check_urls.py

递归检查 `data/**/*.txt` 中的媒体链接，将无效链接和 429 限流链接分别记录到 `scripts/invalid_urls.txt`、`scripts/rate_limited_urls.txt`。确认无效的链接会从原文本文件中删除。

示例：

```powershell
python .\scripts\check_urls.py
```

## ks.py

读取 `scripts/ks.txt` 中的快手分享链接，解析为视频直链并写入 `scripts/output.txt`。

示例：

```powershell
python .\scripts\ks.py
```

## deduplicate.py

对一个 `.txt` 文件或一个目录下的全部 `.txt` 文件去重，保留首次出现的非空行。

示例：

```
python .\scripts\deduplicate.py .\data\videos\pc\cosplay.txt
python .\scripts\deduplicate.py .\data\videos
```

> 注意：链接检查和去重都会覆盖或修改数据文件；执行前请确认已备份或提交当前改动。
