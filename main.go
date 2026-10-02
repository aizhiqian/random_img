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
    storeMutex sync.RWMutex                            // 读写锁
    imageStore = make(map[string]map[string][]string) // category -> size -> []urls
    videoStore = make(map[string]map[string][]string) // category -> size -> []urls
    imageSizes = make(map[string]bool)                 // 可用的图片尺寸
    videoSizes = make(map[string]bool)                 // 可用的视频尺寸
    imageDir   = "./data/images"                       // 图片数据目录
    videoDir   = "./data/videos"                       // 视频数据目录
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
func loadResourcesFromDisk(dataDir string, store *map[string]map[string][]string, sizes *map[string]bool, resourceType string) error {
    storeMutex.Lock()
    defer storeMutex.Unlock()

    // 1. 清空旧数据
    *store = make(map[string]map[string][]string)
    *sizes = make(map[string]bool)

    // 2. 读取尺寸目录
    sizeEntries, err := os.ReadDir(dataDir)
    if err != nil {
        return fmt.Errorf("无法读取%s数据目录: %v", resourceType, err)
    }

    totalCount := 0
    categoryCount := make(map[string]int)

    // 3. 遍历每个尺寸文件夹
    for _, sizeEntry := range sizeEntries {
        if !sizeEntry.IsDir() {
            continue
        }

        sizeName := sizeEntry.Name()
        (*sizes)[sizeName] = true
        sizeDir := filepath.Join(dataDir, sizeName)

        // 4. 读取该尺寸下的分类文件
        categoryFiles, err := os.ReadDir(sizeDir)
        if err != nil {
            fmt.Printf("警告: 无法读取尺寸目录 %s, 错误: %v\n", sizeName, err)
            continue
        }

        for _, file := range categoryFiles {
            if file.IsDir() || !strings.HasSuffix(file.Name(), ".txt") {
                continue
            }

            // 获取分类名
            categoryName := strings.TrimSuffix(file.Name(), ".txt")
            filePath := filepath.Join(sizeDir, file.Name())

            // 5. 读取文件内容
            f, err := os.Open(filePath)
            if err != nil {
                fmt.Printf("警告: 无法打开文件 %s/%s, 错误: %v\n", sizeName, file.Name(), err)
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

            // 6. 存入内存（category -> size -> urls）
            if len(urls) > 0 {
                if (*store)[categoryName] == nil {
                    (*store)[categoryName] = make(map[string][]string)
                }
                (*store)[categoryName][sizeName] = urls
                totalCount += len(urls)
                categoryCount[categoryName]++
                fmt.Printf("加载成功: %s [%s/%s] 包含 %d 个资源\n", resourceType, categoryName, sizeName, len(urls))
            }
        }
    }

    fmt.Printf("%s初始化完成，共加载 %d 个分类，%d 种尺寸，%d 个资源。\n", resourceType, len(*store), len(*sizes), totalCount)
    return nil
}

// loadImagesFromDisk 加载图片
func loadImagesFromDisk() error {
    return loadResourcesFromDisk(imageDir, &imageStore, &imageSizes, "图片")
}

// loadVideosFromDisk 加载视频
func loadVideosFromDisk() error {
    return loadResourcesFromDisk(videoDir, &videoStore, &videoSizes, "视频")
}

// ==================== 业务逻辑函数 ====================

// 通用的获取随机资源的函数
func getRandomResource(category string, size string, store map[string]map[string][]string, availableSizes map[string]bool, resourceType string) (string, string, string, error) {
    storeMutex.RLock()
    defer storeMutex.RUnlock()

    if len(store) == 0 {
        return "", "", "", fmt.Errorf("%s库为空，请检查数据目录下是否有 txt 文件", resourceType)
    }

    // 验证 size 参数
    if size != "" && size != "all" {
        if !availableSizes[size] {
            return "", "", "", fmt.Errorf("%s尺寸 '%s' 不存在", resourceType, size)
        }
    }

    var targetCategory string
    var targetSize string
    var urls []string

    // 处理分类逻辑
    if category == "" || category == "all" {
        // 从所有分类中随机选择
        keys := make([]string, 0, len(store))
        for k := range store {
            keys = append(keys, k)
        }
        targetCategory = keys[rand.Intn(len(keys))]
    } else {
        // 使用指定分类
        if _, ok := store[category]; !ok {
            return "", "", "", fmt.Errorf("%s分类 '%s' 不存在", resourceType, category)
        }
        targetCategory = category
    }

    // 处理尺寸逻辑
    categoryData := store[targetCategory]
    if size == "" || size == "all" {
        // 从该分类的所有尺寸中随机选择
        sizeKeys := make([]string, 0, len(categoryData))
        for k := range categoryData {
            sizeKeys = append(sizeKeys, k)
        }
        if len(sizeKeys) == 0 {
            return "", "", "", fmt.Errorf("%s分类 '%s' 下没有任何尺寸数据", resourceType, targetCategory)
        }
        targetSize = sizeKeys[rand.Intn(len(sizeKeys))]
        urls = categoryData[targetSize]
    } else {
        // 使用指定尺寸
        var ok bool
        urls, ok = categoryData[size]
        if !ok || len(urls) == 0 {
            return "", "", "", fmt.Errorf("%s分类 '%s' 下不存在尺寸 '%s' 或为空", resourceType, targetCategory, size)
        }
        targetSize = size
    }

    randomUrl := urls[rand.Intn(len(urls))]
    return randomUrl, targetCategory, targetSize, nil
}

// getRandomImage 获取随机图片
func getRandomImage(category string, size string) (string, string, string, error) {
    return getRandomResource(category, size, imageStore, imageSizes, "图片")
}

// getRandomVideo 获取随机视频
func getRandomVideo(category string, size string) (string, string, string, error) {
    return getRandomResource(category, size, videoStore, videoSizes, "视频")
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
func handleRandomResourceAPI(c *gin.Context, getResourceFunc func(string, string) (string, string, string, error), resourceName string) {
    category := c.Query("category")
    size := c.Query("size")
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
        url, cat, sz, err := getResourceFunc(category, size)
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
            "size":     sz,
            "success":  true,
        }

        c.JSON(200, response)
        return
    }

    // 多个资源 - 返回数组格式
    var resources []gin.H
    for i := 0; i < count; i++ {
        url, cat, sz, err := getResourceFunc(category, size)
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
            "size":     sz,
        })
    }

    c.JSON(200, gin.H{
        "success":    true,
        "count":      len(resources),
        resourceName: resources,
    })
}

// 通用的重定向处理函数
func handleResourceRedirect(c *gin.Context, getResourceFunc func(string, string) (string, string, string, error), resourceType string) {
    category := c.Query("category")
    size := c.Query("size")

    url, _, _, err := getResourceFunc(category, size)
    if err != nil {
        c.String(404, "%s未找到: %v", resourceType, err)
        return
    }

    c.Redirect(302, url)
}

// 通用的获取分类列表函数
func handleCategoriesAPI(c *gin.Context, store map[string]map[string][]string) {
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

// 通用的获取尺寸列表函数
func handleSizesAPI(c *gin.Context, sizes map[string]bool) {
    storeMutex.RLock()
    keys := make([]string, 0, len(sizes))
    for k := range sizes {
        keys = append(keys, k)
    }
    storeMutex.RUnlock()

    c.JSON(200, gin.H{
        "sizes": keys,
        "count": len(keys),
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

        // 统计图片
        totalImages := 0
        imageDetails := make(map[string]map[string]int) // category -> size -> count
        for category, sizeMap := range imageStore {
            imageDetails[category] = make(map[string]int)
            for size, urls := range sizeMap {
                count := len(urls)
                imageDetails[category][size] = count
                totalImages += count
            }
        }

        // 统计视频
        totalVideos := 0
        videoDetails := make(map[string]map[string]int) // category -> size -> count
        for category, sizeMap := range videoStore {
            videoDetails[category] = make(map[string]int)
            for size, urls := range sizeMap {
                count := len(urls)
                videoDetails[category][size] = count
                totalVideos += count
            }
        }

        c.JSON(200, gin.H{
            "success": true,
            "images": gin.H{
                "categories": len(imageStore),
                "sizes":      len(imageSizes),
                "total":      totalImages,
                "details":    imageDetails,
            },
            "videos": gin.H{
                "categories": len(videoStore),
                "sizes":      len(videoSizes),
                "total":      totalVideos,
                "details":    videoDetails,
            },
        })
    })

    // 接口 4.1: 获取随机图片 (JSON)
    // 用法: /api/random/image?category=cat&count=5
    r.GET("/api/random/image", func(c *gin.Context) {
        handleRandomResourceAPI(c, getRandomImage, "images")
    })

    // 接口 4.2: 图片重定向 (直接显示图片)
    // 用法: /img?category=wallpaper
    r.GET("/img", func(c *gin.Context) {
        handleResourceRedirect(c, getRandomImage, "图片")
    })

    // 接口 4.3: 查看所有图片分类
    r.GET("/api/categories/image", func(c *gin.Context) {
        handleCategoriesAPI(c, imageStore)
    })

    // 接口 4.4: 查看所有图片尺寸
    r.GET("/api/sizes/image", func(c *gin.Context) {
        handleSizesAPI(c, imageSizes)
    })

    // 接口 5.1: 获取随机视频 (JSON)
    // 用法: /api/random/video?category=movie&count=3
    r.GET("/api/random/video", func(c *gin.Context) {
        handleRandomResourceAPI(c, getRandomVideo, "videos")
    })

    // 接口 5.2: 视频重定向 (直接展示视频)
    // 用法: /video?category=movie
    r.GET("/video", func(c *gin.Context) {
        handleResourceRedirect(c, getRandomVideo, "视频")
    })

    // 接口 5.3: 查看所有视频分类
    r.GET("/api/categories/video", func(c *gin.Context) {
        handleCategoriesAPI(c, videoStore)
    })

    // 接口 5.4: 查看所有视频尺寸
    r.GET("/api/sizes/video", func(c *gin.Context) {
        handleSizesAPI(c, videoSizes)
    })

    // 管理接口 (需要认证)
    admin := r.Group("/admin")
    admin.Use(authMiddleware())
    {
        // 接口 6: 重载图片和视频数据 (修改 txt 后调用)
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
    fmt.Println("服务启动成功，访问地址：http://localhost:38719")
    r.Run(":38719")
}
