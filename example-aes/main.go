package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go-analysis/client"
)

func main() {
	// 创建 AES 客户端
	// 注意：secretKey 必须与服务器配置的 secret_key 一致
	aesClient := client.NewAESClient(
		"http://localhost:8080",
		"go_analysis_aes_2024_key_v1.0",
	)

	// 示例 1: 发送加密的事件数据
	fmt.Println("=== 示例 1: 发送加密的事件数据 ===")
	eventData := map[string]interface{}{
		"event_id":    "user_login",
		"device_id":   "test-device-123",
		"user_id":     "user-456",
		"product":     "test-app",
		"app_version": "1.0.0",
		"properties": map[string]interface{}{
			"login_method": "email",
			"success":      true,
		},
		"timestamp": time.Now().Unix(),
	}

	resp, err := aesClient.PostEncrypted("/api/analytics/events", eventData)
	if err != nil {
		log.Fatalf("发送加密事件失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		log.Fatalf("解析响应失败: %v", err)
	}

	fmt.Printf("响应: %+v\n\n", result)

	// 示例 2: 发送普通（未加密）的事件数据
	fmt.Println("=== 示例 2: 发送普通的事件数据 ===")
	resp2, err := aesClient.PostPlain("/api/analytics/events", eventData)
	if err != nil {
		log.Fatalf("发送普通事件失败: %v", err)
	}

	if err := json.Unmarshal(resp2, &result); err != nil {
		log.Fatalf("解析响应失败: %v", err)
	}

	fmt.Printf("响应: %+v\n\n", result)

	// 示例 3: 验证许可（加密）
	fmt.Println("=== 示例 3: 验证许可（加密）===")
	licenseReq := map[string]interface{}{
		"product":   "test-app",
		"device_id": "test-device-123",
		"timestamp": time.Now().Unix(),
	}

	resp3, err := aesClient.PostEncrypted("/api/license/verify", licenseReq)
	if err != nil {
		log.Fatalf("验证许可失败: %v", err)
	}

	if err := json.Unmarshal(resp3, &result); err != nil {
		log.Fatalf("解析响应失败: %v", err)
	}

	fmt.Printf("响应: %+v\n\n", result)
}
