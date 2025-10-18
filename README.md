# Go Analysis Client

[![Go Reference](https://pkg.go.dev/badge/github.com/difyz9/go-analysis-client.svg)](https://pkg.go.dev/github.com/difyz9/go-analysis-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/difyz9/go-analysis-client)](https://goreportcard.com/report/github.com/difyz9/go-analysis-client)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go Analysis Client 是一个轻量级、高性能的 Go 语言数据分析 SDK，支持事件追踪、用户行为分析和业务数据收集。

## 特性

- 🚀 **高性能**: 异步事件上报，不影响主业务性能
- 🔒 **安全加密**: 支持 AES 加密保护数据传输
- 📊 **丰富事件**: 支持自定义事件、用户事件、设备信息等
- 🎯 **批量上报**: 支持事件批量上报，提高传输效率
- 🛡️ **错误处理**: 完善的错误处理和重试机制
- 📱 **多平台**: 支持 Web、移动端、服务端等多种场景
- 📈 **安装统计**: 自动收集安装信息和应用生命周期数据
- 🔄 **会话管理**: 自动管理用户会话和设备识别

## 快速开始

### 安装

```bash
go get github.com/difyz9/go-analysis-client
```

### 基础使用（30 秒快速开始）

```go
package main

import (
    "log"
    analytics "github.com/difyz9/go-analysis-client"
)

func main() {
    // 1. 创建客户端
    client := analytics.NewClient(
        "http://localhost:8080",  // 服务器地址
        "MyApp",                  // 产品名称
        analytics.WithDebug(true),
    )
    defer client.Close()

    // 2. 可选：上报安装信息
    client.ReportInstall()

    // 3. 发送事件
    client.Track("user_login", map[string]interface{}{
        "platform": "web",
        "version":  "1.0.0",
    })

    log.Println("事件已发送！")
}
```

### 完整示例（包含生命周期）

```go
package main

import (
    analytics "github.com/difyz9/go-analysis-client"
)

func main() {
    // 创建客户端
    client := analytics.NewClient(
        "http://localhost:8080",
        "MyApp",
        analytics.WithDebug(true),
        analytics.WithUserID("user123"),
    )
    
    // 确保退出时发送剩余事件
    defer func() {
        client.TrackAppExit(map[string]interface{}{
            "exit_reason": "normal",
        })
        client.Close()
    }()

    // 上报安装信息（可选）
    client.ReportInstall()

    // 记录应用启动
    client.TrackAppLaunch(map[string]interface{}{
        "version": "1.0.0",
    })

    // 发送业务事件
    client.Track("button_click", map[string]interface{}{
        "button": "submit",
        "screen": "home",
    })
}
```

## 详细文档

### 配置选项

```go
config := analytics.Config{
    ServerURL:     "https://analytics.example.com",  // 必填：分析服务器地址
    ProductID:     "your-product-id",                // 必填：产品ID
    APIKey:        "your-api-key",                   // 必填：API密钥
    Timeout:       30 * time.Second,                 // 可选：请求超时时间
    BatchSize:     50,                               // 可选：批量上报大小
    FlushInterval: 5 * time.Second,                  // 可选：刷新间隔
    EnableEncryption: true,                          // 可选：启用加密
    EncryptionKey:    "your-32-byte-encryption-key", // 可选：加密密钥
}
```

### 事件类型

#### 基础事件
```go
event := analytics.Event{
    Name:       "page_view",
    UserID:     "user123",
    SessionID:  "session456",
    Timestamp:  time.Now(),
    Properties: map[string]interface{}{
        "page": "/dashboard",
        "referrer": "https://google.com",
    },
}
```

#### 用户事件
```go
userEvent := analytics.UserEvent{
    UserID:    "user123",
    Action:    "login",
    Timestamp: time.Now(),
    UserProperties: map[string]interface{}{
        "name":     "张三",
        "email":    "zhangsan@example.com",
        "plan":     "premium",
        "created":  "2023-01-01",
    },
}
```

#### 设备信息
```go
device := analytics.DeviceInfo{
    DeviceID:      "device123",
    Platform:      "web",
    OSName:        "macOS",
    OSVersion:     "13.0",
    AppVersion:    "1.2.0",
    Language:      "zh-CN",
    Timezone:      "Asia/Shanghai",
    ScreenSize:    "1920x1080",
    UserAgent:     "Mozilla/5.0...",
}
```

### 批量上报

```go
events := []analytics.Event{
    {Name: "event1", UserID: "user1"},
    {Name: "event2", UserID: "user2"},
    {Name: "event3", UserID: "user3"},
}

err := client.TrackBatch(events)
if err != nil {
    log.Printf("批量上报失败: %v", err)
}
```

### 加密传输

启用加密可以保护敏感数据：

```go
config := analytics.Config{
    ServerURL:        "https://analytics.example.com",
    ProductID:        "your-product-id",
    APIKey:          "your-api-key",
    EnableEncryption: true,
    EncryptionKey:    "your-32-byte-encryption-key-here!",
}
```

### 错误处理

```go
client.SetErrorHandler(func(err error) {
    log.Printf("Analytics错误: %v", err)
    // 可以发送到错误监控系统
})
```

## 高级使用

### 自定义HTTP客户端

```go
httpClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:       10,
        IdleConnTimeout:    30 * time.Second,
        DisableCompression: true,
    },
}

config.HTTPClient = httpClient
```

### 中间件支持

对于Web应用，可以使用中间件自动收集请求信息：

```go
// Gin框架示例
func AnalyticsMiddleware(client *analytics.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        
        // 记录请求事件
        client.Track(analytics.Event{
            Name: "api_request",
            Properties: map[string]interface{}{
                "method":      c.Request.Method,
                "path":        c.Request.URL.Path,
                "status_code": c.Writer.Status(),
                "duration":    time.Since(start).Milliseconds(),
                "user_agent":  c.Request.UserAgent(),
                "ip":          c.ClientIP(),
            },
        })
    }
}
```

## 示例项目

- [Web应用示例](./example-gin/) - 使用Gin框架的Web应用
- [独立应用示例](./example-standalone/) - 独立Go应用
- [加密传输示例](./example-aes/) - 启用AES加密的示例
- [安装统计示例](./example-standalone/install_demo.go) - 完整的安装统计示例

## API 参考

### 推荐 API

我们推荐使用以下核心 API，它们简洁、灵活且功能完整：

#### 🎯 事件追踪

```go
// 推荐：使用 Track 发送所有事件（异步，高性能）
client.Track("user_login", map[string]interface{}{
    "method": "email",
    "platform": "web",
})

// 如需同步等待：使用 Track + Flush
client.Track("critical_event", properties)
client.Flush()  // 等待所有事件发送完成
```

#### 📦 批量发送

```go
// 批量发送多个事件
client.TrackBatch([]analytics.Event{
    {Name: "event1", Properties: map[string]interface{}{"key": "value1"}},
    {Name: "event2", Properties: map[string]interface{}{"key": "value2"}},
})
```

### 已废弃 API

以下方法仍然可用以保持向后兼容，但不推荐在新代码中使用：

- ⚠️ `TrackEvent(category, action, label, value)` - 已废弃，请使用 `Track` 替代
- ⚠️ `TrackSync(eventName, properties)` - 已废弃，请使用 `Track + Flush` 替代

### 完整方法列表

- `NewClient(serverURL, productName string, opts ...ClientOption) *Client` - 创建客户端
- `Track(eventName string, properties map[string]interface{})` - **推荐**：发送事件（异步）
- `TrackBatch(events []Event)` - **推荐**：批量发送事件
- `Flush()` - **推荐**：立即刷新缓冲区
- `ReportInstall()` - 上报安装信息（异步）
- `ReportInstallWithCallback(callback func(error))` - 上报安装信息并回调
- `TrackAppLaunch(properties map[string]interface{})` - 记录应用启动
- `TrackAppExit(properties map[string]interface{})` - 记录应用退出
- `SetUserID(userID string)` - 设置用户ID
- `GetDeviceID() string` - 获取设备ID
- `GetSessionID() string` - 获取会话ID
- `Close()` - 关闭客户端

### 配置结构

```go
type Config struct {
    ServerURL        string        // 服务器地址
    ProductID        string        // 产品ID
    APIKey           string        // API密钥
    Timeout          time.Duration // 请求超时
    BatchSize        int           // 批量大小
    FlushInterval    time.Duration // 刷新间隔
    EnableEncryption bool          // 启用加密
    EncryptionKey    string        // 加密密钥
    HTTPClient       *http.Client  // HTTP客户端
}
```

## 最佳实践

1. **性能优化**
   - 使用批量上报减少网络请求
   - 设置合适的刷新间隔
   - 避免在高频场景中同步发送事件

2. **错误处理**
   - 设置错误处理器记录失败事件
   - 实现重试机制
   - 监控上报成功率

3. **数据安全**
   - 启用加密传输保护敏感数据
   - 定期轮换API密钥
   - 避免在客户端存储敏感信息

4. **事件设计**
   - 使用清晰的事件名称
   - 保持属性结构一致
   - 避免过度收集用户数据

## 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 许可证

本项目使用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 支持

- 📖 [文档](https://github.com/difyz9/go-analysis-client/wiki)
- 🐛 [问题反馈](https://github.com/difyz9/go-analysis-client/issues)
- 💬 [讨论区](https://github.com/difyz9/go-analysis-client/discussions)

## 更新日志

### v1.1.0
- ✅ 新增安装信息统计功能
- ✅ 新增应用生命周期跟踪（启动/退出）
- ✅ 新增会话管理和设备识别
- ✅ 支持安装信息回调函数
- ✅ 优化设备ID生成算法

### v1.0.0
- 初始版本发布
- 支持基础事件追踪
- 支持批量上报
- 支持AES加密
- 支持多种事件类型
