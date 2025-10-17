# 快速开始 - Client SDK

## 1分钟快速开始

### 安装

```bash
go get github.com/yourusername/go-analysis/client
```

### 使用（3行代码）

```go
client := analytics.NewClient("http://localhost:8080", "MyApp")
defer client.Close()
client.Track("event_name", map[string]interface{}{"key": "value"})
```

## 完整示例

### 基础使用

```go
package main

import (
    "github.com/yourusername/go-analysis/client"
)

func main() {
    // 1. 创建客户端
    client := analytics.NewClient("http://localhost:8080", "MyApp")
    defer client.Close()

    // 2. 跟踪事件
    client.Track("button_click", map[string]interface{}{
        "button_name": "login",
        "screen": "home",
    })

    // 3. 等待发送完成（可选）
    client.Flush()
}
```

### 高级配置

```go
client := analytics.NewClient(
    "http://localhost:8080",
    "MyApp",
    analytics.WithDeviceID("device-123"),
    analytics.WithUserID("user-456"),
    analytics.WithBatchSize(50),
    analytics.WithDebug(true),
)
```

## Gin 集成示例

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/yourusername/go-analysis/client"
)

var analyticsClient *analytics.Client

func main() {
    // 初始化分析客户端
    analyticsClient = analytics.NewClient("http://localhost:8080", "GinApp")
    defer analyticsClient.Close()

    r := gin.Default()
    
    // 使用中间件自动跟踪所有请求
    r.Use(AnalyticsMiddleware())

    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello"})
    })

    r.Run(":8000")
}

func AnalyticsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        
        analyticsClient.Track("http_request", map[string]interface{}{
            "method": c.Request.Method,
            "path": c.Request.URL.Path,
            "status": c.Writer.Status(),
            "duration_ms": time.Since(start).Milliseconds(),
        })
    }
}
```

## 运行示例

```bash
# 运行独立示例
cd client/example-standalone
go run main.go

# 运行 Gin 集成示例
cd client/example-gin
go run main.go
```

## API 参考

### 创建客户端
```go
NewClient(serverURL, productName string, opts ...ClientOption) *Client
```

### 跟踪事件
```go
Track(eventName string, properties map[string]interface{})    // 异步
TrackSync(eventName string, properties map[string]interface{}) error // 同步
TrackEvent(category, action, label string, value float64)      // GA风格
TrackBatch(events []Event)                                     // 批量
```

### 管理
```go
Flush()            // 强制发送所有缓冲事件
Close() error      // 关闭客户端
```

### 配置选项
```go
WithDeviceID(deviceID string)
WithUserID(userID string)
WithTimeout(timeout time.Duration)
WithBatchSize(size int)
WithFlushInterval(interval time.Duration)
WithBufferSize(size int)
WithDebug(debug bool)
WithLogger(logger Logger)
```

## 特性

- ✅ **开箱即用** - 3行代码即可开始
- ✅ **零依赖** - 仅依赖 uuid 包
- ✅ **异步发送** - 不阻塞主程序
- ✅ **批量处理** - 自动合并事件
- ✅ **线程安全** - 可并发使用
- ✅ **自动重试** - 网络错误自动处理
- ✅ **内存高效** - 可配置缓冲区大小
- ✅ **调试友好** - 内置 debug 模式

## 文档

- [完整文档](./README.md)
- [重构说明](../CLIENT_REFACTOR.md)
- [示例代码](./examples_test.go)

## 支持

如有问题，请查看：
1. [README.md](./README.md) - 完整使用指南
2. [CLIENT_REFACTOR.md](../CLIENT_REFACTOR.md) - 架构说明
3. [examples_test.go](./examples_test.go) - 更多示例
