package analytics_test

import (
	"fmt"
	"log"
	"time"

	"github.com/yourusername/go-analysis/client"
)

// Example_quickStart 演示最简单的 3 行使用方式
func Example_quickStart() {
	client := analytics.NewClient("http://localhost:8080", "MyApp")
	defer client.Close()
	
	client.Track("page_view", map[string]interface{}{
		"page": "/home",
		"referrer": "google.com",
	})
}

// Example_fullOptions 演示完整配置选项
func Example_fullOptions() {
	client := analytics.NewClient(
		"http://localhost:8080",
		"MyApp",
		analytics.WithDeviceID("device-123"),
		analytics.WithUserID("user-456"),
		analytics.WithTimeout(15*time.Second),
		analytics.WithBatchSize(50),
		analytics.WithFlushInterval(10*time.Second),
		analytics.WithDebug(true),
		analytics.WithLogger(log.Default()),
	)
	defer client.Close()
	
	// 跟踪事件
	client.Track("button_click", map[string]interface{}{
		"button": "login",
		"screen": "home",
	})
	
	// Google Analytics 风格事件
	client.TrackEvent("user", "login", "email", 1)
	
	// 批量跟踪
	events := []analytics.Event{
		{Name: "event1", Properties: map[string]interface{}{"key": "value1"}},
		{Name: "event2", Properties: map[string]interface{}{"key": "value2"}},
	}
	client.TrackBatch(events)
	
	// 同步发送（阻塞）
	err := client.TrackSync("important_event", map[string]interface{}{
		"data": "must be sent immediately",
	})
	if err != nil {
		log.Printf("Failed to send event: %v", err)
	}
	
	// 手动刷新缓冲区
	client.Flush()
}

// Example_webApplication 演示在 Web 应用中使用
func Example_webApplication() {
	// 在应用启动时创建客户端
	analyticsClient := analytics.NewClient(
		"http://analytics.example.com",
		"WebApp",
		analytics.WithUserID("user-123"),
		analytics.WithBatchSize(100),
	)
	
	// 在关闭时清理
	defer analyticsClient.Close()
	
	// 在处理请求时跟踪事件
	handleRequest := func() {
		analyticsClient.Track("api_request", map[string]interface{}{
			"endpoint": "/api/users",
			"method":   "GET",
			"status":   200,
		})
	}
	
	handleRequest()
}

// Example_gaming 演示在游戏中使用
func Example_gaming() {
	client := analytics.NewClient(
		"http://game-analytics.example.com",
		"MyGame",
		analytics.WithDeviceID("switch-12345"),
		analytics.WithBatchSize(200), // 游戏可能产生大量事件
		analytics.WithFlushInterval(30*time.Second), // 降低网络频率
	)
	defer client.Close()
	
	// 跟踪游戏事件
	client.Track("level_complete", map[string]interface{}{
		"level":    5,
		"score":    1500,
		"duration": 120, // 秒
	})
	
	client.Track("item_purchase", map[string]interface{}{
		"item":     "sword",
		"currency": "gold",
		"amount":   100,
	})
	
	client.Track("achievement_unlock", map[string]interface{}{
		"achievement": "first_blood",
		"timestamp":   time.Now().Unix(),
	})
}

// Example_mobileApp 演示在移动应用中使用
func Example_mobileApp() {
	// 从持久化存储读取或生成设备ID
	deviceID := loadDeviceID() // 假设有这个函数
	
	client := analytics.NewClient(
		"http://mobile-analytics.example.com",
		"MobileApp",
		analytics.WithDeviceID(deviceID),
		analytics.WithBufferSize(500), // 移动设备可能间歇性联网
		analytics.WithTimeout(30*time.Second), // 移动网络可能较慢
	)
	defer client.Close()
	
	// 跟踪屏幕浏览
	client.Track("screen_view", map[string]interface{}{
		"screen_name": "home",
		"previous":    "splash",
	})
	
	// 跟踪用户行为
	client.Track("button_tap", map[string]interface{}{
		"button_id": "share",
		"content":   "article_123",
	})
	
	// 在应用进入后台时刷新
	client.Flush()
}

// Example_errorHandling 演示错误处理
func Example_errorHandling() {
	client := analytics.NewClient(
		"http://localhost:8080",
		"MyApp",
		analytics.WithDebug(true),
		analytics.WithLogger(log.Default()),
	)
	defer client.Close()
	
	// 异步发送（不阻塞，如果缓冲区满会丢弃）
	client.Track("normal_event", map[string]interface{}{
		"data": "value",
	})
	
	// 同步发送（可以捕获错误）
	err := client.TrackSync("critical_event", map[string]interface{}{
		"important": "data",
	})
	if err != nil {
		// 处理错误，例如记录到文件或重试
		log.Printf("Failed to send critical event: %v", err)
	}
}

// Example_userTracking 演示用户跟踪
func Example_userTracking() {
	client := analytics.NewClient("http://localhost:8080", "MyApp")
	defer client.Close()
	
	// 用户登录后设置用户ID
	onUserLogin := func(userID string) {
		client.SetUserID(userID)
		client.Track("user_login", map[string]interface{}{
			"method": "email",
		})
	}
	
	// 用户登出后清除用户ID
	onUserLogout := func() {
		client.Track("user_logout", nil)
		client.SetUserID("")
	}
	
	onUserLogin("user-123")
	
	// 跟踪用户行为
	client.Track("profile_view", map[string]interface{}{
		"user_id": client.GetDeviceID(),
	})
	
	onUserLogout()
}

// loadDeviceID 假设的辅助函数
func loadDeviceID() string {
	// 实际实现中应该从本地存储读取
	return "device-12345"
}

// 演示如何获取客户端信息
func Example_clientInfo() {
	client := analytics.NewClient("http://localhost:8080", "MyApp")
	defer client.Close()
	
	fmt.Printf("Device ID: %s\n", client.GetDeviceID())
	fmt.Printf("Session ID: %s\n", client.GetSessionID())
	
	// Output (示例输出，实际值会不同):
	// Device ID: abc-123-def-456
	// Session ID: xyz-789-uvw-012
}
