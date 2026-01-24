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

// 全局变量存储图片和视频数据
var (
    imageStore = make(map[string][]string)
    videoStore = make(map[string][]string) // 视频存储
    storeMutex sync.RWMutex                // 读写锁
    imageDir   = "./data/images"           // 图片数据目录
    videoDir   = "./data/videos"           // 视频数据目录
    adminToken string
    apiDocHTML string
    previewHTML string
)

// 初始化
func init() {
    // 1. 加载 .env 文件 (如果存在)
    _ = godotenv.Load()

    // 2. 读取环境变量
    adminToken = getEnvOrDefault("ADMIN_TOKEN", "your-secret-admin-token")

    // 3. 设置随机数种子
    rand.Seed(time.Now().UnixNano())

    // 4. 确保数据目录存在
    for _, dir := range []string{imageDir, videoDir} {
        if _, err := os.Stat(dir); os.IsNotExist(err) {
            os.MkdirAll(dir, 0755)
            fmt.Printf("检测到 %s 目录不存在，已自动创建。\n", dir)
        }
    }
}

// ==================== 工具函数 ====================

// 获取环境变量，若为空则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
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

// 加载预览页面
func loadPreviewPage() error {
    templatePath := "./preview.html"
    content, err := os.ReadFile(templatePath)
    if err != nil {
        return fmt.Errorf("无法读取预览页面: %v", err)
    }
    previewHTML = string(content)
    fmt.Println("预览页面模板加载成功")
    return nil
}

// ==================== 数据加载函数 ====================

// 通用的从磁盘加载资源的函数
func loadResourcesFromDisk(dataDir string, store *map[string][]string, resourceType string) error {
    storeMutex.Lock()
    defer storeMutex.Unlock()

    // 1. 清空旧数据
    *store = make(map[string][]string)

    // 2. 读取目录
    files, err := os.ReadDir(dataDir)
    if err != nil {
        return fmt.Errorf("无法读取%s数据目录: %v", resourceType, err)
    }

    count := 0
    for _, file := range files {
        if file.IsDir() || !strings.HasSuffix(file.Name(), ".txt") {
            continue
        }

        // 获取分类名
        categoryName := strings.TrimSuffix(file.Name(), ".txt")
        filePath := filepath.Join(dataDir, file.Name())

        // 3. 读取文件内容
        f, err := os.Open(filePath)
        if err != nil {
            fmt.Printf("警告: 无法打开文件 %s, 错误: %v\n", file.Name(), err)
            continue
        }

        var urls []string
        scanner := bufio.NewScanner(f)
        for scanner.Scan() {
            line := strings.TrimSpace(scanner.Text())
            if line != "" {
                urls = append(urls, line)
            }
        }
        f.Close()

        // 4. 存入内存
        if len(urls) > 0 {
            (*store)[categoryName] = urls
            count += len(urls)
            fmt.Printf("加载成功: %s分类 [%s] 包含 %d 个资源\n", resourceType, categoryName, len(urls))
        }
    }

    fmt.Printf("%s初始化完成，共加载 %d 个分类，%d 个资源。\n", resourceType, len(*store), count)
    return nil
}

// loadImagesFromDisk 加载图片
func loadImagesFromDisk() error {
    return loadResourcesFromDisk(imageDir, &imageStore, "图片")
}

// loadVideosFromDisk 加载视频
func loadVideosFromDisk() error {
    return loadResourcesFromDisk(videoDir, &videoStore, "视频")
}

// ==================== 业务逻辑函数 ====================

// 通用的获取随机资源的函数
func getRandomResource(category string, store map[string][]string, resourceType string) (string, string, error) {
    storeMutex.RLock()
    defer storeMutex.RUnlock()

    if len(store) == 0 {
        return "", "", fmt.Errorf("%s库为空，请检查数据目录下是否有 txt 文件", resourceType)
    }

    var targetCategory string
    var urls []string

    // 处理分类逻辑
    if category == "" || category == "all" {
        keys := make([]string, 0, len(store))
        for k := range store {
            keys = append(keys, k)
        }
        targetCategory = keys[rand.Intn(len(keys))]
        urls = store[targetCategory]
    } else {
        var ok bool
        urls, ok = store[category]
        if !ok || len(urls) == 0 {
            return "", "", fmt.Errorf("%s分类 '%s' 不存在或为空", resourceType, category)
        }
        targetCategory = category
    }

    randomUrl := urls[rand.Intn(len(urls))]
    return randomUrl, targetCategory, nil
}

// getRandomImage 获取随机图片
func getRandomImage(category string) (string, string, error) {
    return getRandomResource(category, imageStore, "图片")
}

// getRandomVideo 获取随机视频
func getRandomVideo(category string) (string, string, error) {
    return getRandomResource(category, videoStore, "视频")
}

// ==================== HTTP 处理器函数 ====================

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

// 通用的处理随机资源 API 的函数
func handleRandomResourceAPI(c *gin.Context, getResourceFunc func(string) (string, string, error), resourceName string) {
    category := c.Query("category")
    countStr := c.DefaultQuery("count", "1")

    count, err := strconv.Atoi(countStr)
    if err != nil || count < 1 {
        count = 1
    }
    if count > 20 {
        count = 20
    }

    // 单个资源 - 保持简洁响应格式
    if count == 1 {
        url, cat, err := getResourceFunc(category)
        if err != nil {
            c.JSON(404, gin.H{
                "success": false,
                "message": err.Error(),
            })
            return
        }

        response := gin.H{
            "url":      url,
            "category": cat,
            "success":  true,
        }

        c.JSON(200, response)
        return
    }

    // 多个资源 - 返回数组格式
    var resources []gin.H
    for i := 0; i < count; i++ {
        url, cat, err := getResourceFunc(category)
        if err != nil {
            c.JSON(404, gin.H{
                "success": false,
                "message": err.Error(),
            })
            return
        }
        resources = append(resources, gin.H{
            "url":      url,
            "category": cat,
        })
    }

    c.JSON(200, gin.H{
        "success":    true,
        "count":      len(resources),
        resourceName: resources,
    })
}

// 通用的重定向处理函数
func handleResourceRedirect(c *gin.Context, getResourceFunc func(string) (string, string, error), resourceType string) {
    category := c.Query("category")

    url, _, err := getResourceFunc(category)
    if err != nil {
        c.String(404, "%s未找到: %v", resourceType, err)
        return
    }

    c.Redirect(302, url)
}

// 通用的获取分类列表函数
func handleCategoriesAPI(c *gin.Context, store map[string][]string) {
    storeMutex.RLock()
    keys := make([]string, 0, len(store))
    for k := range store {
        keys = append(keys, k)
    }
    storeMutex.RUnlock()

    c.JSON(200, gin.H{
        "categories": keys,
        "count":      len(keys),
    })
}

// ==================== 主函数 ====================

func main() {
    // 0. 开启发布模式 (关闭调试日志，提升性能)
    gin.SetMode(gin.ReleaseMode)

    // 1. 加载 API 文档
    if err := loadAPIDoc(); err != nil {
        fmt.Printf("警告: API 文档加载失败: %v\n", err)
        apiDocHTML = "<h1>API 文档加载失败</h1><p>请确保 api_doc.html 文件存在</p>"
    }

    // 2. 加载预览页面
    if err := loadPreviewPage(); err != nil {
        fmt.Printf("警告: 预览页面加载失败: %v\n", err)
        previewHTML = "<h1>预览页面加载失败</h1><p>请确保 preview.html 文件存在</p>"
    }

    // 3. 加载图片和视频数据
    if err := loadImagesFromDisk(); err != nil {
        fmt.Printf("图片加载失败: %v\n", err)
    }
    if err := loadVideosFromDisk(); err != nil {
        fmt.Printf("视频加载失败: %v\n", err)
    }

    // 4. 设置 Gin 路由
    r := gin.Default()

    // 5. 添加 favicon.ico 静态文件支持
    r.StaticFile("/favicon.ico", "./favicon.ico")

    // 6. 设置受信任的代理
    r.SetTrustedProxies([]string{"127.0.0.1"})

    // 7. 允许跨域 (前端调用必备)
    r.Use(cors.Default())

    // --- 接口区域 ---

    // 接口 0: API 文档首页
    r.GET("/", func(c *gin.Context) {
        c.Header("Content-Type", "text/html; charset=utf-8")
        c.String(200, apiDocHTML)
    })

    // 接口 1: 预览页面
    r.GET("/preview", func(c *gin.Context) {
        c.Header("Content-Type", "text/html; charset=utf-8")
        c.String(200, previewHTML)
    })

    // 接口 2: 健康检查端点
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status":    "ok",
            "timestamp": time.Now().Format(time.RFC3339),
        })
    })

    // 接口 3: 统计信息端点
    r.GET("/api/stats", func(c *gin.Context) {
        storeMutex.RLock()
        defer storeMutex.RUnlock()

        totalImages := 0
        imageDetails := make(map[string]int)
        for category, urls := range imageStore {
            imageDetails[category] = len(urls)
            totalImages += len(urls)
        }

        totalVideos := 0
        videoDetails := make(map[string]int)
        for category, urls := range videoStore {
            videoDetails[category] = len(urls)
            totalVideos += len(urls)
        }

        c.JSON(200, gin.H{
            "success": true,
            "images": gin.H{
                "categories": len(imageStore),
                "total":      totalImages,
                "details":    imageDetails,
            },
            "videos": gin.H{
                "categories": len(videoStore),
                "total":      totalVideos,
                "details":    videoDetails,
            },
        })
    })

    // 接口 4: 获取随机图片 (JSON)
    // 用法: /api/random/image?category=cat&count=5
    r.GET("/api/random/image", func(c *gin.Context) {
        handleRandomResourceAPI(c, getRandomImage, "images")
    })

    // 接口 5: 图片重定向 (直接显示图片)
    // 用法: /img?category=wallpaper
    r.GET("/img", func(c *gin.Context) {
        handleResourceRedirect(c, getRandomImage, "图片")
    })

    // 接口 6: 查看所有图片分类
    r.GET("/api/categories/image", func(c *gin.Context) {
        handleCategoriesAPI(c, imageStore)
    })

    // 接口 7: 获取随机视频 (JSON)
    // 用法: /api/random/video?category=movie&count=3
    r.GET("/api/random/video", func(c *gin.Context) {
        handleRandomResourceAPI(c, getRandomVideo, "videos")
    })

    // 接口 8: 视频重定向 (直接展示视频)
    // 用法: /video?category=movie
    r.GET("/video", func(c *gin.Context) {
        handleResourceRedirect(c, getRandomVideo, "视频")
    })

    // 接口 9: 查看所有视频分类
    r.GET("/api/categories/video", func(c *gin.Context) {
        handleCategoriesAPI(c, videoStore)
    })

    // 管理接口 (需要认证)
    admin := r.Group("/admin")
    admin.Use(authMiddleware())
    {
        // 接口 10: 重载图片和视频数据 (修改 txt 后调用)
        // 支持 POST 和 GET 方法，方便浏览器直接访问
        reloadHandler := func(c *gin.Context) {
            errImg := loadImagesFromDisk()
            errVid := loadVideosFromDisk()

            if errImg != nil && errVid != nil {
                c.JSON(500, gin.H{
                    "success": false,
                    "message": fmt.Sprintf("图片加载失败: %v; 视频加载失败: %v", errImg, errVid),
                })
            } else if errImg != nil {
                c.JSON(200, gin.H{
                    "success": true,
                    "message": "视频库已重新加载，但图片加载失败: " + errImg.Error(),
                })
            } else if errVid != nil {
                c.JSON(200, gin.H{
                    "success": true,
                    "message": "图片库已重新加载，但视频加载失败: " + errVid.Error(),
                })
            } else {
                c.JSON(200, gin.H{"success": true, "message": "图片和视频库已重新加载"})
            }
        }
        admin.POST("/reload", reloadHandler)
        admin.GET("/reload", reloadHandler)
    }

    // 启动服务，监听 38719 端口
    fmt.Println("服务启动成功，监听端口 38719 ...")
    r.Run(":38719")
}
