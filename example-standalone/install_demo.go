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

// SimpleLogger 实现简单的日志记录器
type SimpleLogger struct{}

func (l *SimpleLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func main() {
	// 1. 创建分析客户端
	client := analytics.NewClient(
		"http://localhost:8080", // 替换为您的服务器地址
		"my-awesome-app",        // 产品名称
		analytics.WithDebug(true),
		analytics.WithLogger(&SimpleLogger{}),
		analytics.WithBatchSize(10),
		analytics.WithFlushInterval(5*time.Second),
		// 可选：启用加密
		// analytics.WithEncryption("your-32-byte-secret-key-here!"),
	)
	defer func() {
		// 确保在退出前发送所有事件
		client.TrackAppExit(map[string]interface{}{
			"exit_reason": "normal",
		})
		client.Close()
	}()

	// 2. 上报安装信息（首次启动或每次启动都可以调用）
	log.Println("上报安装信息...")
	client.ReportInstallWithCallback(func(err error) {
		if err != nil {
			log.Printf("❌ 上报安装信息失败: %v", err)
		} else {
			log.Println("✅ 上报安装信息成功")
		}
	})

	// 3. 记录应用启动事件
	log.Println("记录应用启动...")
	client.TrackAppLaunch(map[string]interface{}{
		"version":    "1.0.0",
		"build":      "100",
		"launch_via": "command_line",
		"env":        os.Getenv("ENV"),
	})

	// 4. 打印设备信息
	fmt.Printf("\n📱 设备信息:\n")
	fmt.Printf("   Device ID: %s\n", client.GetDeviceID())
	fmt.Printf("   Session ID: %s\n", client.GetSessionID())

	// 5. 模拟应用运行，记录各种事件
	fmt.Println("\n🚀 应用开始运行...")
	
	// 设置优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	// 启动事件模拟
	go simulateUserActivity(client)
	
	// 等待退出信号
	<-sigChan
	fmt.Println("\n\n⏹️  收到退出信号，正在清理...")
}

// simulateUserActivity 模拟用户活动
func simulateUserActivity(client *analytics.Client) {
	// 模拟用户登录
	time.Sleep(1 * time.Second)
	log.Println("📊 事件: 用户登录")
	client.Track("user_login", map[string]interface{}{
		"method":    "email",
		"device_id": client.GetDeviceID(),
	})

	// 模拟页面浏览
	time.Sleep(2 * time.Second)
	log.Println("📊 事件: 页面浏览")
	client.Track("page_view", map[string]interface{}{
		"page": "/dashboard",
		"from": "/home",
	})

	// 模拟功能使用
	time.Sleep(2 * time.Second)
	log.Println("📊 事件: 功能使用")
	client.Track("feature_used", map[string]interface{}{
		"feature":  "export_data",
		"format":   "csv",
		"duration": 1.5,
	})

	// 模拟按钮点击
	time.Sleep(1 * time.Second)
	log.Println("📊 事件: 按钮点击")
	client.Track("button_click", map[string]interface{}{
		"button_name": "settings",
		"page":        "/dashboard",
	})

	// 模拟错误发生
	time.Sleep(2 * time.Second)
	log.Println("📊 事件: 错误发生")
	client.Track("error_occurred", map[string]interface{}{
		"error_type": "network_timeout",
		"error_code": "ETIMEDOUT",
		"endpoint":   "/api/data",
	})

	// 批量发送事件示例
	time.Sleep(2 * time.Second)
	log.Println("📊 事件: 批量发送")
	batchEvents := []analytics.Event{
		{
			Name: "batch_event_1",
			Properties: map[string]interface{}{
				"type": "test",
			},
		},
		{
			Name: "batch_event_2",
			Properties: map[string]interface{}{
				"type": "test",
			},
		},
		{
			Name: "batch_event_3",
			Properties: map[string]interface{}{
				"type": "test",
			},
		},
	}
	client.TrackBatch(batchEvents)

	fmt.Println("\n✅ 模拟活动完成，按 Ctrl+C 退出程序")
}
