# 安装信息统计功能

## 功能概述

go-analysis-client 提供了完整的安装信息统计功能，可以追踪应用的安装、启动和使用情况。

## 核心功能

### 1. 安装信息上报

自动收集并上报设备信息，包括：
- 产品名称和设备ID
- 操作系统信息（OS、Platform、Version）
- 硬件信息（架构、内核版本）
- 时间戳和安全签名

### 2. 应用生命周期跟踪

- **应用启动统计** - 记录每次应用启动
- **应用退出统计** - 记录会话时长和退出原因
- **会话管理** - 自动生成和管理会话ID

### 3. 设备识别

- 基于系统唯一标识生成稳定的设备ID
- 跨重启保持一致的设备标识
- 支持自定义设备ID

## 快速开始

### 基础使用

```go
package main

import (
    "log"
    analytics "github.com/difyz9/go-analysis-client"
)

func main() {
    // 1. 创建客户端
    client := analytics.NewClient(
        "http://your-server.com",
        "your-app-name",
        analytics.WithDebug(true),
    )
    defer client.Close()

    // 2. 上报安装信息（异步，不阻塞）
    client.ReportInstall()

    // 3. 记录应用启动
    client.TrackAppLaunch(map[string]interface{}{
        "version": "1.0.0",
    })

    // 4. 您的业务逻辑
    // ...

    // 5. 应用退出前记录
    client.TrackAppExit(map[string]interface{}{
        "exit_reason": "normal",
    })
}
```

### 完整示例

```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    analytics "github.com/difyz9/go-analysis-client"
)

func main() {
    // 创建客户端（启用调试模式）
    client := analytics.NewClient(
        "http://localhost:8080",
        "geekai-plus",
        analytics.WithDebug(true),
        analytics.WithBatchSize(10),
        analytics.WithFlushInterval(5*time.Second),
    )

    // 确保退出时记录数据
    defer func() {
        client.TrackAppExit(map[string]interface{}{
            "exit_reason": "normal",
        })
        client.Close()
    }()

    // 上报安装信息（带回调）
    client.ReportInstallWithCallback(func(err error) {
        if err != nil {
            log.Printf("上报失败: %v", err)
        } else {
            log.Println("上报成功")
        }
    })

    // 记录应用启动
    client.TrackAppLaunch(map[string]interface{}{
        "version":    "1.0.0",
        "build":      "100",
        "launch_via": "desktop",
    })

    // 打印设备信息
    fmt.Printf("Device ID: %s\n", client.GetDeviceID())
    fmt.Printf("Session ID: %s\n", client.GetSessionID())

    // 模拟应用运行
    runApplication(client)

    // 等待退出信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan
}

func runApplication(client *analytics.Client) {
    // 记录各种事件
    client.Track("user_login", map[string]interface{}{
        "method": "email",
    })

    client.Track("page_view", map[string]interface{}{
        "page": "/dashboard",
    })

    client.Track("feature_used", map[string]interface{}{
        "feature": "export",
    })
}
```

## API 文档

### ReportInstall()

异步上报安装信息，不阻塞主流程。

```go
client.ReportInstall()
```

**特点**:
- 异步执行，立即返回
- 自动收集设备信息
- 生成安全签名
- 失败不影响主流程

### ReportInstallWithCallback(callback func(error))

上报安装信息并执行回调函数。

```go
client.ReportInstallWithCallback(func(err error) {
    if err != nil {
        log.Printf("上报失败: %v", err)
        // 可以记录到日志系统或重试
    } else {
        log.Println("上报成功")
    }
})
```

**参数**:
- `callback`: 回调函数，接收错误参数

**使用场景**:
- 需要知道上报结果
- 实现自定义错误处理
- 触发后续业务逻辑

### TrackAppLaunch(properties map[string]interface{})

记录应用启动事件。

```go
client.TrackAppLaunch(map[string]interface{}{
    "version":    "1.0.0",
    "build":      "100",
    "launch_via": "desktop_icon",
    "env":        "production",
})
```

**自动添加的属性**:
- `session_id`: 会话ID
- `device_id`: 设备ID
- `session_started`: 会话开始时间
- `hostname`: 主机名
- `os`: 操作系统
- `platform`: 平台信息
- `uptime`: 系统运行时间

### TrackAppExit(properties map[string]interface{})

记录应用退出事件（同步发送）。

```go
client.TrackAppExit(map[string]interface{}{
    "exit_reason": "user_quit",
    "has_error":   false,
})
```

**自动添加的属性**:
- `session_duration`: 会话时长（秒）
- `session_id`: 会话ID
- `device_id`: 设备ID

**注意**: 此方法会同步发送数据，确保在应用退出前完成上报。

## 数据结构

### InstallInfo

```go
type InstallInfo struct {
    Product         string `json:"product"`          // 产品名称
    DeviceID        string `json:"device_id"`        // 设备ID
    Timestamp       int64  `json:"timestamp"`        // 时间戳
    Sign            string `json:"sign"`             // 安全签名
    
    // 设备详细信息
    Hostname        string `json:"hostname"`
    OS              string `json:"os"`
    Platform        string `json:"platform"`
    PlatformVersion string `json:"platform_version"`
    KernelVersion   string `json:"kernel_version"`
    KernelArch      string `json:"kernel_arch"`
    Uptime          uint64 `json:"uptime"`
}
```

### 签名算法

```go
// SHA256(product#device_id#timestamp)
signStr := fmt.Sprintf("%s#%s#%d", product, deviceID, timestamp)
sign := sha256.Sum256([]byte(signStr))
```

## 服务端API

### 安装信息接收端点

**URL**: `POST /api/installs/push`

**请求体**:
```json
{
  "product": "geekai-plus",
  "device_id": "unique-device-id-here",
  "timestamp": 1697654400,
  "sign": "sha256-signature",
  "hostname": "user-macbook",
  "os": "darwin",
  "platform": "darwin",
  "platform_version": "13.0",
  "kernel_version": "22.1.0",
  "kernel_arch": "arm64",
  "uptime": 86400
}
```

**响应**:
```json
{
  "code": 200,
  "message": "success"
}
```

## 高级特性

### 1. 加密通讯

```go
client := analytics.NewClient(
    "http://localhost:8080",
    "geekai-plus",
    analytics.WithEncryption("your-32-byte-secret-key-here!"),
)

// 所有数据将自动加密传输
client.ReportInstall()
```

### 2. 自定义设备ID

```go
client := analytics.NewClient(
    "http://localhost:8080",
    "geekai-plus",
    analytics.WithDeviceID("custom-device-id"),
)
```

### 3. 自定义日志记录

```go
type MyLogger struct{}

func (l *MyLogger) Printf(format string, v ...interface{}) {
    // 自定义日志处理
    log.Printf("[MyApp] "+format, v...)
}

client := analytics.NewClient(
    "http://localhost:8080",
    "geekai-plus",
    analytics.WithLogger(&MyLogger{}),
    analytics.WithDebug(true),
)
```

### 4. 批量配置

```go
client := analytics.NewClient(
    "http://localhost:8080",
    "geekai-plus",
    analytics.WithBatchSize(50),           // 50个事件一批
    analytics.WithFlushInterval(10*time.Second), // 10秒自动刷新
    analytics.WithBufferSize(1000),        // 缓冲区大小
    analytics.WithTimeout(30*time.Second), // 请求超时
)
```

## 最佳实践

### 1. 应用启动时初始化

```go
func main() {
    client := initAnalyticsClient()
    defer cleanupAnalytics(client)
    
    // 上报安装信息
    client.ReportInstall()
    
    // 记录启动
    client.TrackAppLaunch(getAppInfo())
    
    // 运行应用
    runApp(client)
}

func cleanupAnalytics(client *analytics.Client) {
    client.TrackAppExit(map[string]interface{}{
        "exit_reason": "normal",
    })
    client.Close()
}
```

### 2. 错误处理

```go
client.ReportInstallWithCallback(func(err error) {
    if err != nil {
        // 记录到错误日志
        errorLogger.Printf("安装信息上报失败: %v", err)
        
        // 可选：触发重试机制
        retryReportInstall(client)
    }
})
```

### 3. 优雅退出

```go
// 捕获退出信号
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

go func() {
    <-sigChan
    
    // 记录退出事件
    client.TrackAppExit(map[string]interface{}{
        "exit_reason": "signal",
    })
    
    // 确保所有事件发送完成
    client.Flush()
    client.Close()
    
    os.Exit(0)
}()
```

### 4. 性能优化

```go
// 使用较大的批量大小和较长的刷新间隔
client := analytics.NewClient(
    serverURL,
    productName,
    analytics.WithBatchSize(100),          // 减少请求次数
    analytics.WithFlushInterval(30*time.Second), // 降低频率
    analytics.WithBufferSize(5000),        // 增大缓冲区
)
```

## 故障排查

### 1. 启用调试模式

```go
client := analytics.NewClient(
    serverURL,
    productName,
    analytics.WithDebug(true),
    analytics.WithLogger(&SimpleLogger{}),
)
```

### 2. 检查设备ID生成

```go
deviceID := client.GetDeviceID()
fmt.Printf("Device ID: %s\n", deviceID)
```

### 3. 验证网络连接

```go
err := client.TrackSync("test_event", map[string]interface{}{
    "test": true,
})
if err != nil {
    log.Printf("网络连接失败: %v", err)
}
```

## 常见问题

**Q: 安装信息会重复上报吗？**

A: 会。每次调用 `ReportInstall()` 都会上报，服务端需要根据 device_id 去重。

**Q: 如何确保退出时数据不丢失？**

A: 使用 `TrackAppExit()` 和 `defer client.Close()`，它们会同步发送数据。

**Q: 设备ID是如何生成的？**

A: 优先使用系统UUID（HostID），如果获取失败则基于主机信息生成稳定的哈希ID。

**Q: 是否支持离线缓存？**

A: 当前版本不支持，失败的请求会直接丢弃。可以在服务端实现重试逻辑。

## 示例项目

完整的示例代码位于：
- `example-standalone/install_demo.go` - 独立应用示例
- `example-gin/` - Gin Web 应用示例
- `example-aes/` - 加密传输示例

运行示例：

```bash
cd example-standalone
go run install_demo.go
```

## 更新日志

### v1.1.0
- ✅ 新增安装信息统计功能
- ✅ 新增应用生命周期跟踪
- ✅ 新增会话管理
- ✅ 优化设备ID生成算法
- ✅ 支持回调函数

## 相关文档

- [快速开始](./QUICKSTART.md)
- [API 参考](./README.md)
- [加密指南](./AES.md)
- [最佳实践](./BEST_PRACTICES.md)
