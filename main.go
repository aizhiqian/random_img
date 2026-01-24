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
    apiDocHTML string
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

// 加载 API 文档
func loadAPIDoc() error {
    templatePath := "./api_doc.html"
    content, err := os.ReadFile(templatePath)
    if err != nil {
        return fmt.Errorf("无法读取 API 文档: %v", err)
    }
    apiDocHTML = string(content)
    fmt.Println("API 文档模板加载成功")
    return nil
}

// 初始化
func init() {
    // 1. 加载 .env 文件 (如果存在)
    _ = godotenv.Load()

    // 2. 读取环境变量
    adminToken = getEnvOrDefault("ADMIN_TOKEN", "your-secret-admin-token")

    // 3. 设置随机数种子
    rand.Seed(time.Now().UnixNano())

    // 4. 确保数据目录存在
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

    // 1. 加载 API 文档
    if err := loadAPIDoc(); err != nil {
        fmt.Printf("警告: API 文档加载失败: %v\n", err)
        apiDocHTML = "<h1>API 文档加载失败</h1><p>请确保 api_doc.html 文件存在</p>"
    }

    // 2. 加载图片数据
    if err := loadImagesFromDisk(); err != nil {
        fmt.Printf("初始化失败: %v\n", err)
        return
    }

    // 3. 设置 Gin 路由
    r := gin.Default()

    // 4. 添加 favicon.ico 静态文件支持
    r.StaticFile("/favicon.ico", "./favicon.ico")

    // 5. 设置受信任的代理
    r.SetTrustedProxies([]string{"127.0.0.1"})

    // 6. 允许跨域 (前端调用必备)
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
        // 支持 POST 和 GET 方法，方便浏览器直接访问
        reloadHandler := func(c *gin.Context) {
            if err := loadImagesFromDisk(); err != nil {
                c.JSON(500, gin.H{"success": false, "message": err.Error()})
            } else {
                c.JSON(200, gin.H{"success": true, "message": "图片库已重新加载"})
            }
        }
        admin.POST("/reload", reloadHandler)
        admin.GET("/reload", reloadHandler)
    }

    // 启动服务，监听 38719 端口
    fmt.Println("服务启动成功，监听端口 38719 ...")
    r.Run(":38719")
}