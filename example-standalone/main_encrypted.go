package main

import (
	"fmt"
	"log"
	"time"

	"github.com/yourusername/go-analysis/client"
)

func main() {
	fmt.Println("=== Go-Analysis Client - Encrypted Communication Example ===\n")

	// AES 密钥（必须与服务端配置相同，16/24/32字节）
	// ⚠️ 生产环境中应该从环境变量或配置文件读取
	secretKey := "go_analysis_aes_2024_key_v1.0" // 32字节密钥

	// 创建客户端，启用加密通讯
	fmt.Println("1. Creating analytics client with AES encryption...")
	analyticsClient := client.NewClient(
		"http://localhost:8080", // 服务端地址
		"EncryptedApp",          // 应用名称
		// 启用 AES 加密 - 所有通讯都会被加密
		client.WithEncryption(secretKey),
		// 其他可选配置
		client.WithBatchSize(20),
		client.WithFlushInterval(5*time.Second),
		client.WithDebug(true), // 开启调试日志查看加密过程
	)
	defer analyticsClient.Close()

	fmt.Println("   ✅ Client created with encryption enabled\n")

	// 示例1：发送敏感的支付事件
	fmt.Println("2. Tracking sensitive payment event (encrypted)...")
	analyticsClient.Track("payment_completed", map[string]interface{}{
		"user_id":     "user_12345",
		"amount":      199.99,
		"currency":    "USD",
		"card_last4":  "4242", // 敏感数据：信用卡后四位
		"card_brand":  "visa",
		"merchant_id": "merchant_789",
		"timestamp":   time.Now().Unix(),
	})
	fmt.Println("   ✅ Payment event tracked (encrypted)\n")

	// 示例2：发送用户个人信息
	fmt.Println("3. Tracking user profile update (encrypted)...")
	analyticsClient.Track("profile_updated", map[string]interface{}{
		"user_id": "user_12345",
		"email":   "user@example.com", // 敏感数据：邮箱
		"phone":   "+1234567890",      // 敏感数据：手机号
		"address": map[string]interface{}{ // 敏感数据：地址
			"street":  "123 Main St",
			"city":    "San Francisco",
			"state":   "CA",
			"zip":     "94105",
			"country": "US",
		},
		"updated_at": time.Now().Unix(),
	})
	fmt.Println("   ✅ Profile update tracked (encrypted)\n")

	// 示例3：发送登录事件（包含IP等敏感信息）
	fmt.Println("4. Tracking login event (encrypted)...")
	analyticsClient.Track("user_login", map[string]interface{}{
		"user_id":    "user_12345",
		"ip_address": "203.0.113.42", // 敏感数据：IP地址
		"user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
		"device": map[string]interface{}{
			"type":     "desktop",
			"os":       "macOS",
			"browser":  "Chrome",
			"version":  "120.0.0",
			"language": "en-US",
		},
		"location": map[string]interface{}{
			"country":   "US",
			"region":    "California",
			"city":      "San Francisco",
			"latitude":  37.7749,
			"longitude": -122.4194,
		},
		"login_method": "password",
		"timestamp":    time.Now().Unix(),
	})
	fmt.Println("   ✅ Login event tracked (encrypted)\n")

	// 示例4：批量发送多个事件
	fmt.Println("5. Tracking multiple events in batch (encrypted)...")
	for i := 1; i <= 5; i++ {
		analyticsClient.Track("page_view", map[string]interface{}{
			"user_id":  "user_12345",
			"page":     fmt.Sprintf("/products/item_%d", i),
			"duration": i * 1000,
			"referrer": "https://google.com",
			"session": map[string]interface{}{
				"id":         "session_abc123",
				"started_at": time.Now().Add(-10 * time.Minute).Unix(),
				"page_count": i,
			},
			"timestamp": time.Now().Unix(),
		})
	}
	fmt.Println("   ✅ 5 events tracked in batch (encrypted)\n")

	// 示例5：发送自定义结构的事件
	fmt.Println("6. Tracking custom structured event (encrypted)...")
	analyticsClient.Track("purchase_completed", map[string]interface{}{
		"user_id":  "user_12345",
		"order_id": "order_xyz789",
		"items": []map[string]interface{}{
			{
				"product_id": "prod_001",
				"name":       "Premium Subscription",
				"quantity":   1,
				"price":      99.99,
				"currency":   "USD",
			},
			{
				"product_id": "prod_002",
				"name":       "Additional Storage",
				"quantity":   2,
				"price":      9.99,
				"currency":   "USD",
			},
		},
		"subtotal": 119.97,
		"tax":      10.80,
		"total":    130.77,
		"payment": map[string]interface{}{
			"method":     "credit_card",
			"card_last4": "4242", // 敏感数据
			"card_brand": "visa",
		},
		"shipping": map[string]interface{}{
			"method":  "standard",
			"cost":    0.00,
			"address": "123 Main St, San Francisco, CA 94105", // 敏感数据
		},
		"timestamp": time.Now().Unix(),
	})
	fmt.Println("   ✅ Purchase event tracked (encrypted)\n")

	// 等待所有事件发送完成
	fmt.Println("7. Flushing remaining events...")
	time.Sleep(1 * time.Second)
	analyticsClient.Close()
	fmt.Println("   ✅ All events flushed and client closed\n")

	fmt.Println("=== Example completed successfully ===")
	fmt.Println("\n📝 Notes:")
	fmt.Println("   - All communication is encrypted with AES-256")
	fmt.Println("   - Sensitive data (PII, payment info) is protected")
	fmt.Println("   - Request body and response body are both encrypted")
	fmt.Println("   - Check server logs to see encrypted requests")
	fmt.Println("\n🔐 Security:")
	fmt.Println("   - Always use environment variables for secret keys in production")
	fmt.Println("   - Rotate keys regularly (every 3-6 months)")
	fmt.Println("   - Use HTTPS in production for additional transport security")
}
