package main

import (
    "bufio"
    "fmt"
    "math/rand"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
)

// 定义返回数据的结构体
type ImageResponse struct {
    Url      string `json:"url"`
    Category string `json:"category"`
    Success  bool   `json:"success"`
    Message  string `json:"message,omitempty"`
}

// 全局变量存储图片数据
var (
    imageStore = make(map[string][]string)
    storeMutex sync.RWMutex // 读写锁
    dataDir    = "./data"   // 数据目录
    adminToken string
)

// 获取环境变量，若为空则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// 认证中间件
func authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("X-Admin-Token")
        if token == "" {
            token = c.Query("token") // 也支持 URL 参数传递
        }
        if token != adminToken {
            c.AbortWithStatusJSON(401, gin.H{
                "success": false,
                "message": "未授权访问，请提供有效的管理员 Token",
            })
            return
        }
        c.Next()
    }
}

// API 文档 HTML 模板
const apiDocHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Random Image API - 文档</title>
    <link rel="icon" href="./favicon.ico" type="image/x-icon">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background: #f1f5f9;
            min-height: 100vh;
            padding-bottom: 40px;
        }

        .container {
            max-width: 900px;
            margin: 40px auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 4px 20px rgba(0,0,0,0.05);
            overflow: hidden;
            padding: 0;
        }

        .hero {
            text-align: center;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            padding: 60px 40px;
            color: white;
        }

        .hero h1 {
            font-size: 3rem;
            margin-bottom: 16px;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.2);
        }

        .hero p {
            font-size: 1.1rem;
            opacity: 0.9;
            max-width: 600px;
            margin: 10px auto 0;
        }

        .hero .badge {
            display: inline-block;
            background: rgba(255,255,255,0.2);
            padding: 6px 14px;
            border-radius: 20px;
            margin-top: 15px;
            font-size: 0.85rem;
            backdrop-filter: blur(5px);
        }

        .cards {
            display: block;
            padding: 0;
        }

        .card {
            background: transparent;
            border-radius: 0;
            box-shadow: none;
            padding: 40px;
            margin: 0;
            border-bottom: 1px solid #f1f5f9;
            display: block;
        }

        .card:last-child {
            border-bottom: none;
        }

        .card:hover {
            transform: none;
            box-shadow: none;
            z-index: auto;
        }

        .card-header {
            display: flex;
            align-items: center;
            gap: 12px;
            margin-bottom: 16px;
            border-bottom: none;
            padding-bottom: 0;
        }

        .method {
            padding: 4px 10px;
            border-radius: 6px;
            font-weight: 700;
            font-size: 0.8rem;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .method.get { background: #d1fae5; color: #065f46; }
        .method.post { background: #fef3c7; color: #92400e; }

        .endpoint {
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 1.25rem;
            color: #1f2937;
            font-weight: 600;
        }

        .card p {
            color: #6b7280;
            margin-bottom: 16px;
            flex-grow: 1;
        }

        .params {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
            font-size: 0.9rem;
            border: 1px solid #e5e7eb;
            border-radius: 6px;
            overflow: hidden;
            border-spacing: 0;
        }

        .params th, .params td {
            padding: 12px 20px;
            text-align: left;
        }

        .params th {
            background: #f8fafc;
            color: #4b5563;
            text-transform: uppercase;
            font-size: 0.75rem;
            letter-spacing: 0.05em;
            border-bottom: 1px solid #e5e7eb;
            font-weight: 600;
        }

        .params td {
             border-bottom: 1px solid #f3f4f6;
             color: #374151;
        }

        .params tr:last-child td {
            border-bottom: none;
        }

        .code-block {
            background: #111827;
            border-radius: 8px;
            padding: 20px;
            padding-right: 80px;
            margin-top: 20px;
            position: relative;
            overflow-x: auto;
            border: 1px solid #374151;
        }

        .code-block code {
            color: #e5e7eb;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.85rem;
            white-space: pre-wrap;
            word-break: break-all;
            display: block;
        }

        .code-block .copy-btn {
            position: absolute;
            top: 10px;
            right: 10px;
            background: rgba(255, 255, 255, 0.1);
            color: #d1d5db;
            border: 1px solid rgba(255, 255, 255, 0.2);
            padding: 4px 10px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.75rem;
            transition: all 0.2s;
            z-index: 2;
        }

        .code-block .copy-btn:hover {
            background: #4b5563;
            color: white;
            border-color: #6b7280;
        }

        .tag {
            display: inline-block;
            padding: 4px 10px;
            border-radius: 12px;
            font-size: 0.75rem;
            font-weight: 500;
            margin-right: 6px;
        }

        .tag.required { background: #fee2e2; color: #dc2626; }
        .tag.optional { background: #e0e7ff; color: #4f46e5; }
        .tag.auth { background: #fef3c7; color: #d97706; }

        @media (max-width: 768px) {
            .container { margin: 0; border-radius: 0; }
            .card { padding: 30px 20px; }
            .hero h1 { font-size: 2rem; }
        }

        @media (max-width: 480px) {
            .container { padding: 12px; }
            .hero { padding: 40px 12px; }
            .hero h1 { font-size: 1.75rem; }
            .card-header { flex-wrap: wrap; }
        }

        .info-tip {
            max-width: 900px;
            margin: 0 auto 0 auto;
            background: #fef3c7;
            color: #92400e;
            padding: 16px 24px;
            font-size: 1rem;
            box-shadow: 0 2px 8px rgba(0,0,0,0.04);
            border: 1px solid #fde68a;
            display: block;
        }

        .info-tip code {
            background: #fde68a;
            padding: 2px 6px;
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="hero">
            <h1>🖼️ Random Image API</h1>
            <p>轻量级随机图片服务，支持分类管理、批量获取、热重载，零数据库设计</p>
            <div class="badge">✨ Go + Gin | 高性能 | 跨域支持</div>
        </div>

        <div class="info-tip">
            💡 数据存储于 <code>./data/*.txt</code> 文件，每行一个图片 URL，文件名即分类名
        </div>

        <div class="cards">
            <div class="card">
                <div class="card-header">
                    <span class="method get">GET</span>
                    <span class="endpoint">/api/random</span>
                </div>
                <p>获取随机图片信息，支持指定分类和批量获取</p>
                <table class="params">
                    <tr><th>参数</th><th>类型</th><th>说明</th></tr>
                    <tr><td><code>category</code></td><td><span class="tag optional">可选</span></td><td>分类名，留空或 "all" 随机分类</td></tr>
                    <tr><td><code>count</code></td><td><span class="tag optional">可选</span></td><td>数量 1-20，默认 1</td></tr>
                </table>
                <div class="code-block">
                    <button class="copy-btn" onclick="copyCode(this)">复制</button>
                    <code>GET /api/random?category=wallpaper&count=2</code>
                </div>
                <p style="margin-top: 20px; color: #374151; font-weight: 600;">响应示例：</p>
                <div class="code-block">
                    <button class="copy-btn" onclick="copyCode(this)">复制</button>
                    <code>{
    "success": true,
    "count": 2,
    "images": [
        {
            "url": "https://example.com/wallpaper1.jpg",
            "category": "wallpaper"
        },
        {
            "url": "https://example.com/wallpaper2.jpg",
            "category": "wallpaper"
        }
    ]
}</code>
                </div>
            </div>

            <div class="card">
                <div class="card-header">
                    <span class="method get">GET</span>
                    <span class="endpoint">/img</span>
                </div>
                <p>302 重定向到随机图片 URL，可直接嵌入 &lt;img&gt; 标签</p>
                <table class="params">
                    <tr><th>参数</th><th>类型</th><th>说明</th></tr>
                    <tr><td><code>category</code></td><td><span class="tag optional">可选</span></td><td>分类名</td></tr>
                </table>
                <div class="code-block">
                    <button class="copy-btn" onclick="copyCode(this)">复制</button>
                    <code>&lt;img src="/img?category=cat" alt="随机图片"&gt;</code>
                </div>
                <p style="margin-top: 20px; color: #374151; font-weight: 600;">响应：</p>
                <div class="code-block">
                    <button class="copy-btn" onclick="copyCode(this)">复制</button>
                    <code>HTTP/1.1 302 Found
Location: https://example.com/cat123.jpg</code>
                </div>
            </div>

            <div class="card">
                <div class="card-header">
                    <span class="method get">GET</span>
                    <span class="endpoint">/api/categories</span>
                </div>
                <p>获取所有已加载的图片分类列表</p>
                <div class="code-block">
                    <button class="copy-btn" onclick="copyCode(this)">复制</button>
                    <code>GET /api/categories</code>
                </div>
                <p style="margin-top: 20px; color: #374151; font-weight: 600;">响应示例：</p>
                <div class="code-block">
                    <button class="copy-btn" onclick="copyCode(this)">复制</button>
                    <code>{
    "categories": [
        "cat",
        "wallpaper",
        "anime"
    ],
    "count": 3
}</code>
                </div>
            </div>

            <div class="card">
                <div class="card-header">
                    <span class="method get">GET</span>
                    <span class="endpoint">/api/stats</span>
                </div>
                <p>获取图片库统计信息，包括分类数量和图片总数</p>
                <div class="code-block">
                    <button class="copy-btn" onclick="copyCode(this)">复制</button>
                    <code>GET /api/stats</code>
                </div>
                <p style="margin-top: 20px; color: #374151; font-weight: 600;">响应示例：</p>
                <div class="code-block">
                    <button class="copy-btn" onclick="copyCode(this)">复制</button>
                    <code>{
    "success": true,
    "categories": 3,
    "total_images": 250,
    "details": {
        "cat": 100,
        "wallpaper": 120,
        "anime": 30
    }
}</code>
                </div>
            </div>

            <div class="card">
                <div class="card-header">
                    <span class="method get">GET</span>
                    <span class="endpoint">/health</span>
                </div>
                <p>服务健康检查端点，用于监控和负载均衡</p>
                <div class="code-block">
                    <button class="copy-btn" onclick="copyCode(this)">复制</button>
                    <code>GET /health</code>
                </div>
                <p style="margin-top: 20px; color: #374151; font-weight: 600;">响应示例：</p>
                <div class="code-block">
                    <button class="copy-btn" onclick="copyCode(this)">复制</button>
                    <code>{
    "status": "ok",
    "timestamp": "2024-01-01T12:00:00Z"
}</code>
                </div>
            </div>

            <div class="card">
                <div class="card-header">
                    <span class="method post">POST</span>
                    <span class="endpoint">/admin/reload</span>
                    <span class="tag auth">🔐 需认证</span>
                </div>
                <p>重新加载 data 目录下的图片数据，无需重启服务</p>
                <table class="params">
                    <tr><th>请求头</th><th>说明</th></tr>
                    <tr><td><code>X-Admin-Token</code></td><td>管理员认证 Token，默认为 <code>your-secret-admin-token</code></td></tr>
                </table>
                <div class="code-block">
                    <button class="copy-btn" onclick="copyCode(this)">复制</button>
                    <code>curl -X POST -H "X-Admin-Token: your-secret-admin-token" /admin/reload
或
curl -X POST /admin/reload?token=your-secret-admin-token
</code>
                </div>
                <p style="margin-top: 20px; color: #374151; font-weight: 600;">响应示例：</p>
                <div class="code-block">
                    <button class="copy-btn" onclick="copyCode(this)">复制</button>
                    <code>{
    "success": true,
    "message": "图片库已重新加载"
}</code>
                </div>
            </div>
        </div>
    </div>

    <script>
        function copyCode(btn) {
            const code = btn.nextElementSibling.textContent;
            navigator.clipboard.writeText(code).then(() => {
                btn.textContent = '已复制!';
                setTimeout(() => btn.textContent = '复制', 2000);
            });
        }
    </script>
</body>
</html>`

// 初始化
func init() {
    // 1. 加载 .env 文件 (如果存在)
    _ = godotenv.Load()

    // 2. 读取环境变量
    adminToken = getEnvOrDefault("ADMIN_TOKEN", "your-secret-admin-token")

    rand.Seed(time.Now().UnixNano())

    // 3. 确保数据目录存在
    if _, err := os.Stat(dataDir); os.IsNotExist(err) {
        os.MkdirAll(dataDir, 0755)
        fmt.Println("检测到 data 目录不存在，已自动创建。")
    }
}

// loadImagesFromDisk 扫描 data 目录加载所有 txt 文件
func loadImagesFromDisk() error {
    storeMutex.Lock()
    defer storeMutex.Unlock()

    // 1. 清空旧数据
    imageStore = make(map[string][]string)

    // 2. 读取目录
    files, err := os.ReadDir(dataDir)
    if err != nil {
        return fmt.Errorf("无法读取数据目录: %v", err)
    }

    count := 0
    for _, file := range files {
        if file.IsDir() || !strings.HasSuffix(file.Name(), ".txt") {
            continue
        }

        // 获取分类名 (如 cat.txt -> cat)
        categoryName := strings.TrimSuffix(file.Name(), ".txt")
        filePath := filepath.Join(dataDir, file.Name())

        // 3. 读取文件内容
        f, err := os.Open(filePath)
        if err != nil {
            fmt.Printf("警告: 无法打开文件 %s, 错误: %v\n", file.Name(), err)
            continue
        }
        defer f.Close()

        var urls []string
        scanner := bufio.NewScanner(f)
        for scanner.Scan() {
            line := strings.TrimSpace(scanner.Text())
            if line != "" { // 跳过空行
                urls = append(urls, line)
            }
        }

        // 4. 存入内存
        if len(urls) > 0 {
            imageStore[categoryName] = urls
            count += len(urls)
            fmt.Printf("加载成功: 分类 [%s] 包含 %d 张图片\n", categoryName, len(urls))
        }
    }

    fmt.Printf("系统初始化完成，共加载 %d 个分类，%d 张图片。\n", len(imageStore), count)
    return nil
}

// getRandomImage 获取随机图片的逻辑
func getRandomImage(category string) (string, string, error) {
    storeMutex.RLock()
    defer storeMutex.RUnlock()

    if len(imageStore) == 0 {
        return "", "", fmt.Errorf("数据库为空，请检查 data 目录下是否有 txt 文件")
    }

    var targetCategory string
    var urls []string

    // 处理分类逻辑
    if category == "" || category == "all" {
        // 如果没有指定分类，或者指定为 all，则从所有分类中随机选一个
        keys := make([]string, 0, len(imageStore))
        for k := range imageStore {
            keys = append(keys, k)
        }
        targetCategory = keys[rand.Intn(len(keys))]
        urls = imageStore[targetCategory]
    } else {
        // 指定了分类
        var ok bool
        urls, ok = imageStore[category]
        if !ok || len(urls) == 0 {
            return "", "", fmt.Errorf("分类 '%s' 不存在或为空", category)
        }
        targetCategory = category
    }

    // 随机选 URL
    randomUrl := urls[rand.Intn(len(urls))]
    return randomUrl, targetCategory, nil
}

func main() {
    // 0. 开启发布模式 (关闭调试日志，提升性能)
    gin.SetMode(gin.ReleaseMode)

    // 1. 启动时加载
    if err := loadImagesFromDisk(); err != nil {
        fmt.Printf("初始化失败: %v\n", err)
        return
    }

    // 2. 设置 Gin 路由
    r := gin.Default()

    // 3. 添加 favicon.ico 静态文件支持
    r.StaticFile("/favicon.ico", "./favicon.ico")

    // 4. 设置受信任的代理
    r.SetTrustedProxies([]string{"127.0.0.1"})

    // 5. 允许跨域 (前端调用必备)
    r.Use(cors.Default())

    // --- 接口区域 ---

    // 接口 0: API 文档首页
    r.GET("/", func(c *gin.Context) {
        c.Header("Content-Type", "text/html; charset=utf-8")
        c.String(200, apiDocHTML)
    })

    // 接口 1: 健康检查端点
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status":    "ok",
            "timestamp": time.Now().Format(time.RFC3339),
        })
    })

    // 接口 2: 统计信息端点
    r.GET("/api/stats", func(c *gin.Context) {
        storeMutex.RLock()
        defer storeMutex.RUnlock()

        totalImages := 0
        details := make(map[string]int)
        for category, urls := range imageStore {
            details[category] = len(urls)
            totalImages += len(urls)
        }

        c.JSON(200, gin.H{
            "success":      true,
            "categories":   len(imageStore),
            "total_images": totalImages,
            "details":      details,
        })
    })

    // 接口 3: 获取随机图片 (JSON)
    // 用法: /api/random?category=cat&count=5
    r.GET("/api/random", func(c *gin.Context) {
        category := c.Query("category")
        countStr := c.DefaultQuery("count", "1")

        count, err := strconv.Atoi(countStr)
        if err != nil || count < 1 {
            count = 1
        }
        if count > 20 {
            count = 20 // 最大限制 20 张
        }

        // 单张图片 - 保持原有响应格式（向后兼容）
        if count == 1 {
            url, cat, err := getRandomImage(category)
            if err != nil {
                c.JSON(404, ImageResponse{
                    Success: false,
                    Message: err.Error(),
                })
                return
            }
            c.JSON(200, ImageResponse{
                Url:      url,
                Category: cat,
                Success:  true,
            })
            return
        }

        // 多张图片 - 返回数组格式
        var images []gin.H
        for i := 0; i < count; i++ {
            url, cat, err := getRandomImage(category)
            if err != nil {
                c.JSON(404, gin.H{
                    "success": false,
                    "message": err.Error(),
                })
                return
            }
            images = append(images, gin.H{
                "url":      url,
                "category": cat,
            })
        }

        c.JSON(200, gin.H{
            "success": true,
            "count":   len(images),
            "images":  images,
        })
    })

    // 接口 4: 图片重定向 (直接显示图片)
    // 用法: /img?category=wallpaper
    r.GET("/img", func(c *gin.Context) {
        category := c.Query("category")

        url, _, err := getRandomImage(category)
        if err != nil {
            c.String(404, "图片未找到: %v", err)
            return
        }

        // 302 跳转
        c.Redirect(302, url)
    })

    // 接口 5: 查看所有分类
    r.GET("/api/categories", func(c *gin.Context) {
        storeMutex.RLock()
        keys := make([]string, 0, len(imageStore))
        for k := range imageStore {
            keys = append(keys, k)
        }
        storeMutex.RUnlock()

        c.JSON(200, gin.H{
            "categories": keys,
            "count":      len(keys),
        })
    })

    // 管理接口 (需要认证)
    admin := r.Group("/admin")
    admin.Use(authMiddleware())
    {
        // 接口 6: 重载图片数据 (修改 txt 后调用)
        admin.POST("/reload", func(c *gin.Context) {
            if err := loadImagesFromDisk(); err != nil {
                c.JSON(500, gin.H{"success": false, "message": err.Error()})
            } else {
                c.JSON(200, gin.H{"success": true, "message": "图片库已重新加载"})
            }
        })
    }

    // 启动服务，监听 38719 端口
    fmt.Println("服务启动成功，监听端口 38719 ...")
    r.Run(":38719")
}