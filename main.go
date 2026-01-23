package main

import (
    "bufio"
    "fmt"
    "math/rand"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

// 定义返回数据的结构体
type ImageResponse struct {
    Url      string `json:"url"`
    Category string `json:"category"`
    Success  bool   `json:"success"`
    Message  string `json:"message,omitempty"` // omitempty 表示为空时不显示
}

// 全局变量存储图片数据
var (
    imageStore = make(map[string][]string)
    storeMutex sync.RWMutex // 读写锁
    dataDir    = "./data"   // 数据目录
)

// 初始化
func init() {
    rand.Seed(time.Now().UnixNano())
    
    // 确保数据目录存在
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
    // 1. 启动时加载
    if err := loadImagesFromDisk(); err != nil {
        fmt.Printf("初始化失败: %v\n", err)
        return
    }

    // 2. 设置 Gin 路由
    r := gin.Default()

    // 3. 允许跨域 (前端调用必备)
    r.Use(cors.Default())

    // --- 接口区域 ---

    // 接口 1: 获取随机图片 (JSON)
    // 用法: /api/random?category=cat
    r.GET("/api/random", func(c *gin.Context) {
        category := c.Query("category")

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
    })

    // 接口 2: 图片重定向 (直接显示图片)
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

    // 接口 3: 查看所有分类
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

    // 接口 4: 重载接口 (修改 txt 后调用)
    r.POST("/admin/reload", func(c *gin.Context) {
        if err := loadImagesFromDisk(); err != nil {
            c.JSON(500, gin.H{"status": "error", "message": err.Error()})
        } else {
            c.JSON(200, gin.H{"status": "success", "message": "图片库已更新"})
        }
    })

    // 启动服务，监听 38719 端口
    fmt.Println("服务启动成功，监听端口 38719 ...")
    r.Run(":38719")
}