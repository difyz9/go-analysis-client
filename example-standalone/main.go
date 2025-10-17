package main

import (
	"log"
	"time"

	"github.com/yourusername/go-analysis/client"
)

func main() {
	// 创建分析客户端
	client := analytics.NewClient(
		"http://localhost:8080",
		"SimpleApp",
		analytics.WithDebug(true),
		analytics.WithLogger(log.Default()),
		analytics.WithUserID("demo-user"),
	)
	defer client.Close()

	log.Println("Analytics client started...")

	// 1. 发送简单事件
	client.Track("app_start", map[string]interface{}{
		"version": "1.0.0",
		"platform": "demo",
	})

	// 2. 发送用户行为事件
	client.Track("button_click", map[string]interface{}{
		"button_name": "start",
		"screen": "home",
	})

	// 3. 发送 Google Analytics 风格事件
	client.TrackEvent("user", "login", "email", 1)

	// 4. 批量发送事件
	events := []analytics.Event{
		{
			Name: "page_view",
			Properties: map[string]interface{}{
				"page": "/home",
				"referrer": "direct",
			},
		},
		{
			Name: "feature_use",
			Properties: map[string]interface{}{
				"feature": "search",
				"query": "golang",
			},
		},
	}
	client.TrackBatch(events)

	// 5. 同步发送重要事件
	log.Println("Sending critical event synchronously...")
	err := client.TrackSync("payment", map[string]interface{}{
		"amount": 99.99,
		"currency": "USD",
		"item": "premium_plan",
	})
	if err != nil {
		log.Printf("Failed to send payment event: %v", err)
	} else {
		log.Println("Payment event sent successfully")
	}

	// 6. 模拟一些用户活动
	for i := 0; i < 10; i++ {
		client.Track("activity", map[string]interface{}{
			"index": i,
			"timestamp": time.Now().Unix(),
		})
		time.Sleep(100 * time.Millisecond)
	}

	// 7. 手动刷新确保所有事件发送
	log.Println("Flushing remaining events...")
	client.Flush()

	log.Println("All events sent. Exiting...")
	time.Sleep(1 * time.Second) // 等待后台处理完成
}
