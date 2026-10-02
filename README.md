<p align="center">
    <h1 align="center">Random Media API</h1>
    <p align="center">轻量级、高性能、基于本地文件的随机图片和视频 API 接口 🎉</p>
    <p align="center">
        <img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go" alt="Go Version" />
        <a href="https://github.com/aizhiqian/random_img/tree/main?tab=MIT-1-ov-file" target="_blank" >
            <img src="https://img.shields.io/badge/license-MIT-green" />
        </a>
        <a href="https://github.com/aizhiqian/random_img/releases" target="_blank" >
            <img src="https://img.shields.io/github/v/release/aizhiqian/random_img" alt="releases" />
        </a>
        <a href="https://github.com/aizhiqian/random_img/stargazers" target="_blank">
            <img src="https://img.shields.io/github/stars/aizhiqian/random_img" alt="stargazers" />
        </a>
        <a href="https://github.com/aizhiqian/random_img/forks" target="_blank" >
            <img src="https://img.shields.io/github/forks/aizhiqian/random_img" alt="forks" />
        </a>
    </p>
</p>


---

## 📖 简介

这是一个基于 `Golang` + `Gin` 框架开发的轻量级随机图片和视频服务。它不需要复杂的数据库配置，仅需将图片/视频链接按分类存放在 `txt` 文件中即可运行。

非常适合搭建个人随机图床、二次元图片站、壁纸 API、随机视频或前端演示服务。

## ✨ 功能特性

-   🚀 **超高性能**：基于 Go 原生 HTTP 协议，内存占用极低，支持高并发。
-   📂 **零数据库**：数据来源于本地 `.txt` 文本文件，一行一个链接，管理极其简单。
-   🔄 **热重载**：支持运行时重新加载图片和视频数据，无需重启整个服务。
-   🌐 **多种模式**：支持 **JSON 数据返回** 和 **302 重定向** 两种模式。
-   🎨 **分类支持**：自动读取文件名作为分类，支持按分类随机抽取。
-   📏 **尺寸支持**：支持按资源尺寸（PC、手机、正方形等）筛选，灵活适配不同设备。
-   🔢 **批量获取**：支持一次获取 1-20 张随机图片/视频。
-   🛡️ **跨域支持**：内置 CORS 中间件，方便前端直接调用。
-   🎥 **视频支持**：支持随机视频 API，与图片服务同样的使用方式。

## 📚 API 文档

基础地址：`http://your-domain.com`

### 0. 在线预览页面

提供一个可视化的在线预览界面，方便快速查看随机图片和视频。

-   **访问地址**：`http://your-domain.com/preview`
-   **功能特性**：
    -   纯媒体全屏预览
    -   左侧筛选栏默认收起，悬停后可按类型、分类和尺寸筛选
    -   保留已浏览资源的历史记录，可回看上一项
    -   鼠标滚轮和键盘快捷键支持：
        -   向下滚动 / `→` / `↓`：下一项
        -   向上滚动 / `←` / `↑`：上一项

-   **使用场景**：
    -   快速测试 API 功能
    -   在线浏览图片/视频库
    -   向他人展示服务效果

### 1. 获取随机图片 (JSON)

返回 JSON 格式的图片数据，包含 URL 和分类信息。

-   **接口地址**：`GET /api/random/image`
-   **请求参数**：
    | 参数名   | 类型   | 必填 | 说明                                   |
    | :------- | :----- | :--- | :------------------------------------- |
    | category | string | 否   | 指定分类名（如 `cat`），不传则全库随机 |
    | size     | string | 否   | 指定尺寸类型（如 `pc`、`phone`、`square`），不传则随机尺寸 |
    | count    | number | 否   | 获取数量 1-20，默认 1                  |

-   **单张图片响应**：
    ```json
    {
        "url": "https://example.com/cat1.jpg",
        "category": "cat",
        "size": "pc",
        "success": true
    }
    ```

-   **多张图片响应** (count > 1)：
    ```json
    {
        "success": true,
        "count": 3,
        "images": [
            {
                "url": "https://example.com/cat1.jpg",
                "category": "cat",
                "size": "pc"
            },
            {
                "url": "https://example.com/cat2.jpg",
                "category": "cat",
                "size": "phone"
            },
            {
                "url": "https://example.com/cat3.jpg",
                "category": "cat",
                "size": "pc"
            }
        ]
    }
    ```

### 2. 图片重定向 (直接展示)

返回 302 重定向，直接跳转到图片地址。适合 `<img src="...">` 标签直接引用。

-   **接口地址**：`GET /img`
-   **请求参数**：
    | 参数名   | 类型   | 必填 | 说明                         |
    | :------- | :----- | :--- | :--------------------------- |
    | category | string | 否   | 指定分类名（如 `wallpaper`） |
    | size     | string | 否   | 指定尺寸类型（如 `pc`、`phone`） |
-   **响应**：
    ```
    HTTP/1.1 302 Found
    Location: https://example.com/wallpaper123.jpg
    ```
-   **使用方法**：
    ```html
    <img src="http://your-domain.com/img?category=wallpaper&size=pc" />
    ```

### 3. 获取随机视频 (JSON)

返回 JSON 格式的视频数据，包含 URL 和分类信息。

-   **接口地址**：`GET /api/random/video`
-   **请求参数**：
    | 参数名   | 类型   | 必填 | 说明                                   |
    | :------- | :----- | :--- | :------------------------------------- |
    | category | string | 否   | 指定分类名（如 `movie`），不传则全库随机 |
    | size     | string | 否   | 指定尺寸类型（如 `pc`、`phone`），不传则随机尺寸 |
    | count    | number | 否   | 获取数量 1-20，默认 1                  |

-   **单个视频响应**：
    ```json
    {
        "url": "https://example.com/video/movie1.mp4",
        "category": "movie",
        "size": "pc",
        "success": true
    }
    ```

-   **多个视频响应** (count > 1)：
    ```json
    {
        "success": true,
        "count": 3,
        "videos": [
            {
                "url": "https://example.com/video/movie1.mp4",
                "category": "movie",
                "size": "pc"
            },
            {
                "url": "https://example.com/video/movie2.mp4",
                "category": "movie",
                "size": "phone"
            },
            {
                "url": "https://example.com/video/movie3.mp4",
                "category": "movie",
                "size": "pc"
            }
        ]
    }
    ```

### 4. 视频重定向 (直接展示)

返回 302 重定向，直接跳转到视频地址。适合 `<video src="...">` 标签直接引用。

-   **接口地址**：`GET /video`
-   **请求参数**：
    | 参数名   | 类型   | 必填 | 说明                         |
    | :------- | :----- | :--- | :--------------------------- |
    | category | string | 否   | 指定分类名（如 `movie`） |
    | size     | string | 否   | 指定尺寸类型（如 `pc`、`phone`） |
-   **响应**：
    ```
    HTTP/1.1 302 Found
    Location: https://example.com/video/movie123.mp4
    ```
-   **使用方法**：
    ```html
    <video src="http://your-domain.com/video?category=movie&size=pc" controls></video>
    ```

### 5. 获取所有分类

获取当前系统中已加载的所有分类列表。

-   **图片分类接口**：`GET /api/categories/image`
-   **视频分类接口**：`GET /api/categories/video`
-   **响应示例**：

    ```json
    {
        "categories": ["cat", "scenery", "anime"],
        "count": 3
    }
    ```

### 6. 获取所有尺寸

获取当前系统中已加载的所有尺寸类型列表。

-   **图片尺寸接口**：`GET /api/sizes/image`
-   **视频尺寸接口**：`GET /api/sizes/video`
-   **响应示例**：

    ```json
    {
        "sizes": ["pc", "phone", "square"],
        "count": 3
    }
    ```

### 7. 获取统计信息

获取图片库的详细统计信息。

-   **接口地址**：`GET /api/stats`
-   **响应示例**：

    ```json
    {
        "success": true,
        "images": {
            "categories": 3,
            "sizes": 2,
            "total": 250,
            "details": {
                "cat": {
                    "pc": 50,
                    "phone": 50
                },
                "scenery": {
                    "pc": 120
                },
                "anime": {
                    "pc": 30
                }
            }
        },
        "videos": {
            "categories": 2,
            "sizes": 1,
            "total": 50,
            "details": {
                "movie": {
                    "pc": 30
                },
                "music": {
                    "pc": 20
                }
            }
        }
    }
    ```

### 8. 健康检查

用于监控服务状态的健康检查端点。

-   **接口地址**：`GET /health`
-   **响应示例**：

    ```json
    {
        "status": "ok",
        "timestamp": "2024-01-01T12:00:00Z"
    }
    ```

### 8. 重新加载数据 (管理接口)

当你手动修改了 `data/images` 或 `data/videos` 目录下的 txt 文件后,调用此接口刷新内存数据,无需重启服务。

-   **接口地址**:
    -   `POST /admin/reload`
    -   `GET /admin/reload`
-   **认证方式**:
    -   **方式 1**: 请求头 `X-Admin-Token`,默认为 `your-secret-admin-token`
    -   **方式 2**: URL 参数 `token`
-   **使用示例**:

    ```bash
    # 命令行 POST 方式
    curl -X POST -H "X-Admin-Token: your-secret-admin-token" http://your-domain.com/admin/reload

    # 命令行 GET 方式
    curl http://your-domain.com/admin/reload?token=your-secret-admin-token
    ```

-   **响应示例**:

    ```json
    {
        "success": true,
        "message": "图片和视频库已重新加载"
    }
    ```

---

## 📇 目录结构

```
random_go/
├── data/
│   ├── images/          # 图片链接存储目录
│   │   ├── pc/          # PC 端图片
│   │   │   ├── dongman.txt
│   │   │   ├── meinv.txt
│   │   │   └── nature.txt
│   │   ├── phone/       # 手机端图片
│   │   │   └── ...
│   │   └── square/      # 正方形图片
│   │       └── ...
│   └── videos/          # 视频链接存储目录
│       ├── pc/
│       │   ├── movie.txt
│       │   └── music.txt
│       └── phone/
│           └── ...
├── main.go
├── go.mod
├── .env
├── api_doc.html         # API 文档页面
├── preview.html         # 在线预览页面
├── favicon.ico
└── README.md
```

**文件格式说明**：
- 数据按 `{type}/{size}/{category}.txt` 结构组织，如 `data/images/pc/nature.txt`
- 尺寸文件夹名可自定义（如 pc、phone、square、ultrawide 等）
- 每个 `.txt` 文件代表一个分类，文件名即为分类名
- 每行一个 URL，支持图片和视频链接
- 空行会被自动忽略

---

## 🛠️ 宝塔面板部署指南

本程序编译后为二进制文件，在 Linux 服务器上运行极其稳定。

#### 第一步：下载文件

在 [Release](https://github.com/aizhiqian/random_img/releases/latest) 中下载对应平台的二进制文件，例如 `randimg-api-linux-amd64`

#### 第二步：上传文件

1.  在宝塔创建一个网站（目录如 `/www/wwwroot/api.yourdomain.com`）。
2.  将下载的 `randimg-api-linux-amd64、 data 文件夹、.env、favicon.ico、api_doc.html、preview.html` 上传至该目录。
3.  **关键**：将 `randimg-api-linux-amd64` 文件权限设置为 `755`。

#### 第三步：创建网站

1.  宝塔面板依次点击 `网站→Go项目→添加项目`。
2.  按下图配置

    ![](/img/Snipaste20260124-173221.png)

3.  启动后，查看日志确认显示 `服务启动成功，监听端口 38719 ...`。

    ![](/img/Snipaste20260124-174008.png)

---

## 📝 代码示例

### 在前端 JS 中调用

```javascript
// 获取单张图片（指定尺寸）
fetch('http://your-domain.com/api/random/image?category=cat&size=pc')
  .then(response => response.json())
  .then(data => {
    if(data.success) {
      console.log('图片地址:', data.url);
      console.log('尺寸:', data.size);
      document.getElementById('my-img').src = data.url;
    }
  });

// 获取多张图片（随机尺寸）
fetch('http://your-domain.com/api/random/image?category=wallpaper&count=5')
  .then(response => response.json())
  .then(data => {
    if(data.success) {
      data.images.forEach((img, index) => {
        console.log(`图片${index + 1}:`, img.url, '尺寸:', img.size);
      });
    }
  });

// 获取随机视频（手机尺寸）
fetch('http://your-domain.com/api/random/video?category=movie&size=phone')
  .then(response => response.json())
  .then(data => {
    if(data.success) {
      console.log('视频地址:', data.url);
      console.log('尺寸:', data.size);
      document.getElementById('my-video').src = data.url;
    }
  });
```

### 直接在 HTML 中使用

```html
<!-- 直接嵌入随机图片（PC 尺寸） -->
<img src="http://your-domain.com/img?category=anime&size=pc" alt="随机动漫图片" />

<!-- 随机壁纸（手机尺寸） -->
<img src="http://your-domain.com/img?category=wallpaper&size=phone" alt="随机壁纸" />

<!-- 全库随机图片（随机尺寸） -->
<img src="http://your-domain.com/img" alt="随机图片" />

<!-- 随机视频（PC 尺寸） -->
<video src="http://your-domain.com/video?category=movie&size=pc" controls></video>

<!-- 全库随机视频 -->
<video src="http://your-domain.com/video" controls autoplay></video>
```

---

## 📄 开源协议

本项目基于 [MIT License](LICENSE) 开源，免费供个人学习和商业使用。

---

<div align="center">
Made with ❤️ by AiGuoHou
</div>

