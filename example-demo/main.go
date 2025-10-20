package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	analytics "github.com/difyz9/go-analysis-client"
)

func main() {
	// 创建分析客户端 - 连接到 go-analysis-server
	client := analytics.NewClient(
		"http://localhost:8097", // 使用服务器配置的端口
		"DemoApp",               // 产品名称
		analytics.WithDebug(true),
		analytics.WithLogger(log.Default()),
		analytics.WithUserID("demo-user-001"),
	)
	defer client.Close()

	log.Println("=== Go Analysis Client Demo Started ===")
	log.Printf("Device ID: %s", client.GetDeviceID())
	log.Printf("Session ID: %s", client.GetSessionID())
	log.Println()

	// 1. 上报安装信息（包含完整的设备信息）
	log.Println("📦 Reporting installation info...")
	client.ReportInstallWithCallback(func(err error) {
		if err != nil {
			log.Printf("❌ Failed to report install: %v", err)
		} else {
			log.Println("✅ Install info reported successfully")
		}
	})
	time.Sleep(1 * time.Second) // 等待安装信息上报完成

	// 2. 记录应用启动事件
	log.Println("\n🚀 Tracking app launch...")
	client.TrackAppLaunch(map[string]interface{}{
		"version":      "1.0.0",
		"build_number": "100",
		"environment":  "demo",
	})

	// 3. 模拟用户登录
	log.Println("\n👤 Simulating user login...")
	client.Track("user_login", map[string]interface{}{
		"method":     "email",
		"success":    true,
		"login_time": time.Now().Format(time.RFC3339),
	})

	// 4. 模拟页面浏览
	pages := []string{"/home", "/products", "/about", "/contact", "/pricing"}
	log.Println("\n📄 Simulating page views...")
	for _, page := range pages {
		client.Track("page_view", map[string]interface{}{
			"page":      page,
			"referrer":  "direct",
			"duration":  rand.Intn(30) + 5, // 5-35秒
			"timestamp": time.Now().Unix(),
		})
		log.Printf("  - Viewed: %s", page)
		time.Sleep(300 * time.Millisecond)
	}

	// 5. 模拟按钮点击事件
	log.Println("\n🖱️  Simulating button clicks...")
	buttons := []string{"submit", "cancel", "refresh", "download", "share"}
	for _, button := range buttons {
		client.Track("button_click", map[string]interface{}{
			"button_name": button,
			"screen":      "main",
			"x":           rand.Intn(1000),
			"y":           rand.Intn(800),
		})
		log.Printf("  - Clicked: %s", button)
		time.Sleep(200 * time.Millisecond)
	}

	// 6. 模拟功能使用
	log.Println("\n⚙️  Simulating feature usage...")
	features := []struct {
		name       string
		properties map[string]interface{}
	}{
		{"search", map[string]interface{}{"query": "golang analytics", "results": 42}},
		{"filter", map[string]interface{}{"type": "category", "value": "technology"}},
		{"export", map[string]interface{}{"format": "pdf", "size": "1024kb"}},
		{"share", map[string]interface{}{"platform": "twitter", "success": true}},
	}

	for _, feature := range features {
		client.Track("feature_use", feature.properties)
		feature.properties["feature"] = feature.name
		log.Printf("  - Used feature: %s", feature.name)
		time.Sleep(400 * time.Millisecond)
	}

	// 7. 模拟电商行为
	log.Println("\n🛒 Simulating e-commerce events...")
	
	// 浏览商品
	client.Track("product_view", map[string]interface{}{
		"product_id":   "PROD-001",
		"product_name": "Premium Package",
		"price":        99.99,
		"category":     "subscription",
	})
	log.Println("  - Viewed product")
	time.Sleep(500 * time.Millisecond)

	// 添加到购物车
	client.Track("add_to_cart", map[string]interface{}{
		"product_id": "PROD-001",
		"quantity":   1,
		"price":      99.99,
	})
	log.Println("  - Added to cart")
	time.Sleep(500 * time.Millisecond)

	// 开始结账
	client.Track("checkout_start", map[string]interface{}{
		"cart_value": 99.99,
		"items":      1,
	})
	log.Println("  - Started checkout")
	time.Sleep(500 * time.Millisecond)

	// 完成购买
	client.Track("purchase", map[string]interface{}{
		"order_id":     fmt.Sprintf("ORD-%d", time.Now().Unix()),
		"amount":       99.99,
		"currency":     "USD",
		"payment_method": "credit_card",
		"items":        1,
	})
	log.Println("  - Purchase completed")

	// 8. 模拟错误和异常
	log.Println("\n⚠️  Simulating errors...")
	client.Track("error", map[string]interface{}{
		"error_type":    "network",
		"error_message": "Connection timeout",
		"severity":      "warning",
		"retry_count":   3,
	})
	log.Println("  - Network error tracked")

	// 9. 批量发送性能指标
	log.Println("\n📊 Sending performance metrics...")
	performanceEvents := []analytics.Event{
		{
			Name: "performance",
			Properties: map[string]interface{}{
				"metric":   "page_load",
				"duration": 1230,
				"url":      "/home",
			},
		},
		{
			Name: "performance",
			Properties: map[string]interface{}{
				"metric":   "api_response",
				"duration": 450,
				"endpoint": "/api/users",
			},
		},
		{
			Name: "performance",
			Properties: map[string]interface{}{
				"metric":   "render_time",
				"duration": 180,
				"component": "ProductList",
			},
		},
	}
	client.TrackBatch(performanceEvents)
	log.Println("  - Performance metrics sent")

	// 10. 模拟用户交互流
	log.Println("\n🔄 Simulating user interaction flow...")
	interactions := []string{
		"scroll_down",
		"hover_menu",
		"open_dropdown",
		"select_option",
		"submit_form",
	}
	
	for i, interaction := range interactions {
		client.Track("interaction", map[string]interface{}{
			"type":     interaction,
			"sequence": i + 1,
			"timestamp": time.Now().Unix(),
		})
		log.Printf("  - Interaction %d: %s", i+1, interaction)
		time.Sleep(300 * time.Millisecond)
	}

	// 11. 发送自定义业务事件
	log.Println("\n💼 Sending custom business events...")
	client.Track("subscription_activated", map[string]interface{}{
		"plan":       "premium",
		"duration":   "monthly",
		"amount":     99.99,
		"trial_used": false,
	})
	log.Println("  - Subscription activated")

	client.Track("license_verified", map[string]interface{}{
		"license_key": "DEMO-1234-5678-90AB",
		"product":     "DemoApp",
		"valid":       true,
	})
	log.Println("  - License verified")

	// 12. 确保所有事件发送完成
	log.Println("\n⏳ Flushing all events...")
	client.Flush()
	time.Sleep(2 * time.Second) // 等待所有事件处理完成

	// 13. 记录应用退出
	log.Println("\n👋 Tracking app exit...")
	client.TrackAppExit(map[string]interface{}{
		"reason": "normal",
		"clean":  true,
	})

	// 14. 等待最后的事件发送
	client.Flush()
	time.Sleep(1 * time.Second)

	log.Println("\n=== Demo Completed Successfully ===")
	log.Println("📈 Check your analytics dashboard for the results!")
	log.Printf("🔗 Frontend URL: http://localhost:3000")
	log.Printf("📊 Events sent for product: DemoApp")
	log.Printf("🆔 Device ID: %s", client.GetDeviceID())
}
