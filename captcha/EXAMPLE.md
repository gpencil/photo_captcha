# 验证码服务使用示例

## 快速开始

### 1. 在服务启动时初始化

```go
package main

import (
    "log"

    "github.com/gin-gonic/gin"
    "yusheng/go-common/captcha"
)

var captchaService *captcha.CaptchaService

func main() {
    // 1. 创建验证码服务实例
    captchaService = captcha.NewCaptchaService()

    // 2. （可选）设置背景图片URL列表（OSS或本地）
    // 如果不设置，会使用 captcha.BackgroundURLs 的默认值
    captchaService.SetBackgroundURLs([]string{
        "https://your-bucket.oss-cn-hangzhou.aliyuncs.com/captcha/bg1.jpg",
        "https://your-bucket.oss-cn-hangzhou.aliyuncs.com/captcha/bg2.jpg",
        "images/bg3.jpg",  // 也可以混用本地路径
    })

    // 3. 初始化服务（预加载图片和生成mask）
    if err := captchaService.Init(); err != nil {
        log.Fatalf("Failed to initialize captcha service: %v", err)
    }

    // 4. 设置路由
    r := gin.Default()
    r.GET("/api/captcha/generate", generateHandler)
    r.POST("/api/captcha/verify", verifyHandler)

    // 5. 启动服务
    r.Run(":8087")
}

func generateHandler(c *gin.Context) {
    // 使用预加载的服务生成验证码
    sliderCaptcha, err := captchaService.Generate()
    if err != nil {
        c.JSON(500, gin.H{"code": 500, "message": err.Error()})
        return
    }

    c.JSON(200, gin.H{
        "code":    200,
        "message": "success",
        "data": gin.H{
            "id":         sliderCaptcha.ID,
            "background": sliderCaptcha.Background,
            "slider":     sliderCaptcha.Slider,
            "positionY":  sliderCaptcha.PositionY,
        },
    })
}
```

### 2. 使用OSS URL配置

#### 方式一：动态设置（推荐）

```go
// 从配置中心或环境变量获取OSS URL
ossURLs := []string{
    os.Getenv("CAPTCHA_BG_1"),
    os.Getenv("CAPTCHA_BG_2"),
    os.Getenv("CAPTCHA_BG_3"),
}

captchaService := captcha.NewCaptchaService()
captchaService.SetBackgroundURLs(ossURLs)
captchaService.Init()
```

#### 方式二：修改默认配置

编辑 `captcha/image.go`：

```go
var BackgroundURLs = []string{
    "https://your-bucket.oss-cn-hangzhou.aliyuncs.com/captcha/bg1.jpg",
    "https://your-bucket.oss-cn-hangzhou.aliyuncs.com/captcha/bg2.jpg",
    "https://your-bucket.oss-cn-hangzhou.aliyuncs.com/captcha/bg3.jpg",
}
```

### 3. 修改 api-gateway 集成

#### api-gateway/main.go

```go
package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "yusheng/go-common/captcha"
)

var captchaSvc *captcha.CaptchaService

func main() {
    // 创建验证码服务
    captchaSvc = captcha.NewCaptchaService()

    // 从环境变量或配置文件读取OSS URL
    captchaSvc.SetBackgroundURLs([]string{
        os.Getenv("OSS_CAPTCHA_BG_1"),
        os.Getenv("OSS_CAPTCHA_BG_2"),
        os.Getenv("OSS_CAPTCHA_BG_3"),
        // ... 更多图片
    })

    // 初始化（预加载OSS图片到内存）
    if err := captchaSvc.Init(); err != nil {
        log.Fatalf("Failed to initialize captcha service: %v", err)
    }

    // 设置路由并启动
    router := setupRouter()
    router.Run(":8087")
}
```

## 混合使用方案

支持同时使用OSS和本地图片：

```go
captchaService.SetBackgroundURLs([]string{
    // OSS图片（首选，利用CDN加速）
    "https://your-bucket.oss-cn-hangzhou.aliyuncs.com/images/bg1.jpg",
    "https://your-bucket.oss-cn-hangzhou.aliyuncs.com/images/bg2.jpg",

    // 本地图片（备用，无网络依赖）
    "images/backup1.jpg",
    "images/backup2.jpg",
})
```

## 注意事项

1. **必须先调用 Init()**
   ```go
   captchaService = captcha.NewCaptchaService()
   captchaService.Init() // 必须调用，否则会报错
   ```

## 完整示例

```go
package main

import (
    "log"
    "net/http"
    "os"

    "github.com/gin-gonic/gin"
    "yusheng/go-common/captcha"
)

var captchaSvc *captcha.CaptchaService

func main() {
    // 初始化验证码服务
    log.Println("Initializing captcha service...")
    captchaSvc = captcha.NewCaptchaService()

    // 设置OSS背景图URL
    ossURLs := []string{
        os.Getenv("OSS_CAPTCHA_BG_1"),
        os.Getenv("OSS_CAPTCHA_BG_2"),
        os.Getenv("OSS_CAPTCHA_BG_3"),
        os.Getenv("OSS_CAPTCHA_BG_4"),
        os.Getenv("OSS_CAPTCHA_BG_5"),
    }
    captchaSvc.SetBackgroundURLs(ossURLs)

    // 初始化（预加载OSS图片）
    if err := captchaSvc.Init(); err != nil {
        log.Fatalf("Failed to initialize captcha: %v", err)
    }
    log.Println("Captcha service initialized successfully")

    // 设置路由
    r := gin.Default()

    // 验证码 API
    r.GET("/api/captcha/generate", func(c *gin.Context) {
        result, err := captchaSvc.Generate()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "code": 500,
                "message": err.Error(),
            })
            return
        }
        c.JSON(http.StatusOK, gin.H{
            "code": 200,
            "message": "success",
            "data": result,
        })
    })

    // 启动服务
    log.Println("Server starting on :8087")
    r.Run(":8087")
}
```

