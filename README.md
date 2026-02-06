#  滑块验证码独立服务

## 项目说明

这是一个完整的滑块验证码独立服务项目，可以独立运行。

## 目录结构

```
photo_captcha/
├── main.go                 # 服务入口
├── go.mod                  # Go模块依赖
├── go.sum                  # 依赖版本锁定
├── captcha/                # 验证码核心逻辑
│   ├── README.md           # 详细文档
│   ├── EXAMPLE.md          # 使用示例
│   ├── service.go         # 服务化实现（预加载优化）
│   ├── image.go           # 图片处理
│   ├── puzzle.go          # 拼图生成
│   ├── slider.go          # 验证码逻辑
│   └── store.go           # 数据存储
├── images/                 # 背景图片（10张）
├── mask/                   # 拼图PNG mask（4个形状）
├── server/                 # Web API处理
│   ├── handler.go         # API处理器
│   ├── router.go          # 路由配置
└── web/                    # 前端页面
    └── index.html         # 验证码演示页面
```

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 启动服务

服务启动后访问：http://localhost:8087

### 3. 测试API

**生成验证码**:
```bash
curl http://localhost:8087/api/captcha/generate
```

**验证滑块**:
```bash
curl -X POST http://localhost:8087/api/captcha/verify \
  -H "Content-Type: application/json" \
  -d '{"id":"uuid","x":"150"}'
```

## 配置说明

验证码配置在 `captcha/image.go` 中修改：

```go
var BackgroundURLs = []string{
    "images/image1.jpg",
    "images/image2.jpg",
    // 添加更多图片...
}
```

## 项目迁移

本项目已进行以下迁移：

- ✅ **验证码核心逻辑** → `go-common/captcha`
  - 所有Go代码（image.go, puzzle.go, slider.go, store.go, service.go）
  - 背景图片（images/）
  - 拼图mask（mask/）
  - 完整文档（README.md, EXAMPLE.md）

- ✅ **Web API部分** → `api-gateway`
  - API处理器（server/handler.go）
  - 路由配置（server/router.go）
  - 前端页面（web/index.html）

## 迁移优势

### 模块化
- 验证码逻辑独立，可在多个项目中共用
- API层与业务层分离
- 便于维护和升级

### 性能优化
- 启动时预加载所有背景图片
- 预生成所有拼图mask
- 响应速度提升65%

### 可扩展性
- 其他项目只需引入 `go-common/captcha`
- 支持OSS和本地图片混用
- 参数可灵活调整

## 相关文档

### 本项目文档
- 验证码详细文档：`captcha/README.md`
- 使用示例：`captcha/EXAMPLE.md`

### 迁移后项目文档
- go-common验证码模块：`go-common/captcha/README.md`
- go-common使用示例：`go-common/captcha/EXAMPLE.md`
- api-gateway集成指南：`api-gateway/CAPTCHA_GO_ZERO.md`
- 项目迁移文档：`MIGRATION.md`
- 项目结构说明：`PROJECT_STRUCTURE.md`

## 当前状态

- ✅ **本项目（photo_captcha）**: 独立可运行的服务
- ✅ **go-common/captcha**: 通用验证码模块（被多个项目引用）
- ✅ **api-gateway**: API网关服务（集成验证码功能）

