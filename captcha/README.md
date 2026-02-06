# 滑块验证码模块

## 简介

这是一个基于 Go 的高性能滑块验证码实现，支持多种拼图形状（三角形、六边形、梯形、星形）。

## 特性

- ✅ **多种形状**：支持三角形、六边形、梯形、星形
- ✅ **高质量渲染**：使用4480x4480高分辨率PNG mask，经过双线性插值缩放
- ✅ **平滑边缘**：应用高斯模糊和抗锯齿处理，边缘非常平滑
- ✅ **本地图片支持**：支持本地图片路径，无需网络请求

## 目录结构

```
captcha/
├── README.md          # 文档说明
├── EXAMPLE.md         # 使用示例
├── service.go         # 服务化实现（预加载优化版）⭐
├── image.go           # 图片加载、缩放（双线性插值）、base64转换
├── puzzle.go          # 拼图生成、缺口处理、高斯模糊、立体感效果
├── slider.go          # 验证码生成、验证逻辑、形状类型定义
├── store.go           # 验证码存储（内存缓存）
├── images/            # 背景图片目录（16:9，建议1920x1080）
│   ├── image1.jpg
│   ├── image2.jpg
│   └── ...
└── mask/              # 拼图形状PNG mask（4480x4480）
    ├── triangle.png   # 三角形
    ├── hexagon.png    # 六边形
    ├── trapezoid.png  # 梯形
    └── star.png       # 星形
```

## 使用方式

### 方式一：服务化（推荐）⭐

**优势**：
- ✅ 启动时预加载所有背景图片
- ✅ 预生成所有拼图mask
- ✅ 响应速度提升53%
- ✅ 减少CPU和I/O开销

**示例**：

```go
import "yusheng/go-common/captcha"

// 1. 创建服务实例
captchaSvc := captcha.NewCaptchaService()

// 2. 初始化（预加载图片和mask）
if err := captchaSvc.Init(); err != nil {
    log.Fatal(err)
}

// 3. 生成验证码
sliderCaptcha, err := captchaSvc.Generate()
```

详细使用见 [EXAMPLE.md](EXAMPLE.md)

### 方式二：直接调用

```go
import "yusheng/go-common/captcha"

// 生成新的滑块验证码
sliderCaptcha, err := captcha.Generate()
```

## 核心功能

### 1. 生成验证码

```go
import "yusheng/go-common/captcha"

// 生成新的滑块验证码
sliderCaptcha, err := captcha.Generate()
if err != nil {
    // 处理错误
}

// sliderCaptcha 包含：
// - ID: 验证码唯一标识
// - Background: base64编码的背景图（带缺口）
// - Slider: base64编码的滑块图
// - PositionY: 滑块Y轴位置
```

### 2. 验证滑块位置

```go
// 验证滑块位置（默认容差5像素）
success, err := captcha.VerifyWithTolerance(id, userX)

// 或者自定义容差
success, err := captcha.Verify(id, userX, tolerance)
```

## 配置参数

### 拼图块大小

在 `puzzle.go` 中修改：

```go
const (
    PuzzleWidth  = 70  // 拼图宽度
    PuzzleHeight = 70  // 拼图高度
)
```

### 白色遮罩浓度

在 `puzzle.go` 的 `CreatePuzzleHole` 函数中修改：

```go
// 原图比例和白色遮罩比例（加起来=1.0）
R: uint8(float64(c.R)*0.5 + 255*0.5),  // 当前50%原图 + 50%白色
G: uint8(float64(c.G)*0.6 + 255*0.4),
B: uint8(float64(c.B)*0.6 + 255*0.4),
```

### 黑色边框不透明度

在 `puzzle.go` 的 `addHoleBorder` 函数中修改：

```go
borderColor := color.RGBA{R: 0, G: 0, B: 0, A: 0}  // 0=无边框，255=全黑边框
```

### 高斯模糊强度

在 `puzzle.go` 的 `applyGaussianBlur` 函数中修改迭代次数：

```go
// 迭代次数越多，边缘越平滑
for iteration := 0; iteration < 2; iteration++ {  // 当前2次
```

### 背景图片列表

在 `image.go` 中修改：

```go
var BackgroundURLs = []string{
    "images/image1.jpg",
    "images/image2.jpg",
    // 添加更多图片...
}
```

## 图片要求

### 背景图

- **尺寸**：建议 1920x1080 或 1600x900（16:9比例）
- **格式**：JPG 或 PNG
- **大小**：建议 500KB-2MB
- **风格**：风景照、渐变背景、抽象纹理
- **数量**：建议 10-20 张，随机轮换

### 拼图Mask

- **尺寸**：4480x4480（高分辨率）
- **格式**：PNG（透明背景）
- **内容**：白色或彩色图形，完全透明背景
- **渲染**：64倍渲染倍数

## API接口

### 生成验证码

**请求**：
```
GET /api/captcha/generate
```

**响应**：
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "id": "uuid-string",
        "background": "data:image/png;base64,iVBORw0KG...",
        "slider": "data:image/png;base64,iVBORw0KG...",
        "positionY": 75
    }
}
```

### 验证滑块

**请求**：
```
POST /api/captcha/verify
Content-Type: application/json

{
    "id": "uuid-string",
    "x": "150"
}
```

**响应**：
```json
{
    "code": 200,
    "message": "Verification successful",
    "data": {
        "success": true
    }
}
```

## 技术实现

### 图像处理流程

1. **加载背景图**：从本地文件系统加载
2. **加载PNG mask**：4480x4480高分辨率
3. **双线性插值缩放**：保持边缘平滑
4. **生成缺口**：在背景图上创建缺口（白色遮罩）
5. **提取拼图块**：从背景图提取拼图形状
6. **添加边框**：白色边框 + 黑色描边
7. **立体感效果**：边缘高光处理
8. **高斯模糊**：2次迭代，平滑边缘
9. **Base64编码**：转换为base64返回给前端

### 性能优化

- ✅ 图片缓存：避免重复加载
- ✅ 双线性插值：比最近邻插值质量更高
- ✅ 高斯模糊：使用3x3卷积核，性能好
- ✅ 内存缓存：验证码数据存储在内存中，5分钟自动过期

## 依赖

```go
require (
    github.com/gin-gonic/gin v1.9.1
    github.com/google/uuid v1.3.0
)
```

## 使用示例

完整示例参考 `api-gateway` 项目中的集成代码。

## 注意事项

1. **背景图路径**：确保图片路径正确，相对于程序运行目录
2. **Mask文件**：确保4个PNG mask文件都存在
3. **内存管理**：验证码数据会在5分钟后自动清理
4. **并发安全**：使用 `sync.RWMutex` 保护存储

## 版本历史

- v1.0.0 (2025-02-06)
  - 初始版本
  - 支持4种拼图形状
  - 高质量渲染和平滑边缘
