package analytics_test

import (
	"fmt"
	"time"

	analytics "github.com/difyz9/go-analysis-client"
)

// ExampleWithEncryptionOption 演示使用 WithEncryption 选项启用加密
func ExampleWithEncryptionOption() {
	// 创建支持加密的客户端
	client := analytics.NewClient(
		"http://localhost:8080",
		"MySecureApp",
		analytics.WithEncryption("go_analysis_aes_2024_key_v1.0"), // 32字节密钥
		analytics.WithDebug(true),
	)
	defer client.Close()

	// 发送事件 - 自动加密传输
	client.Track("payment", map[string]interface{}{
		"amount":      99.99,
		"currency":    "USD",
		"card_last4":  "4242", // 敏感数据
		"user_id":     "user123",
		"timestamp":   time.Now().Unix(),
	})

	fmt.Println("Event tracked with encryption")
	// Output: Event tracked with encryption
}

// Example_encryptionComparison 对比加密和非加密客户端
func Example_encryptionComparison() {
	serverURL := "http://localhost:8080"
	productName := "MyApp"

	// 场景1: 普通客户端（无加密）
	normalClient := analytics.NewClient(serverURL, productName)
	defer normalClient.Close()

	normalClient.Track("page_view", map[string]interface{}{
		"page": "/home",
	})

	// 场景2: 加密客户端
	encryptedClient := analytics.NewClient(
		serverURL,
		productName,
		analytics.WithEncryption("your-32-byte-secret-key-here!"),
	)
	defer encryptedClient.Close()

	// 敏感数据使用加密
	encryptedClient.Track("user_login", map[string]interface{}{
		"email":    "user@example.com",
		"ip":       "192.168.1.1",
		"password": "***", // 敏感数据
	})

	fmt.Println("Comparison complete")
	// Output: Comparison complete
}
