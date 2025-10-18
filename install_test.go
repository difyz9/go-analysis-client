package analytics_test

import (
	"fmt"
	"log"
	"time"

	analytics "github.com/difyz9/go-analysis-client"
)

// ExampleClient_ReportInstall 演示如何上报安装信息
func ExampleClient_ReportInstall() {
	// 创建客户端
	client := analytics.NewClient(
		"http://localhost:8080",
		"my-awesome-app",
		analytics.WithDebug(true),
	)
	defer client.Close()

	// 上报安装信息（异步）
	client.ReportInstall()

	// 继续执行其他业务逻辑
	fmt.Println("安装信息已提交上报")
	
	// Output: 安装信息已提交上报
}

// ExampleClient_ReportInstallWithCallback 演示带回调的安装信息上报
func ExampleClient_ReportInstallWithCallback() {
	client := analytics.NewClient(
		"http://localhost:8080",
		"my-awesome-app",
	)
	defer client.Close()

	// 上报安装信息并处理结果
	client.ReportInstallWithCallback(func(err error) {
		if err != nil {
			log.Printf("上报安装信息失败: %v", err)
		} else {
			log.Println("上报安装信息成功")
		}
	})

	// 等待回调执行
	time.Sleep(2 * time.Second)
}

// ExampleClient_TrackAppLaunch 演示应用启动统计
func ExampleClient_TrackAppLaunch() {
	client := analytics.NewClient(
		"http://localhost:8080",
		"my-awesome-app",
		analytics.WithDebug(true),
	)
	defer client.Close()

	// 记录应用启动
	client.TrackAppLaunch(map[string]interface{}{
		"version":    "1.0.0",
		"build":      "100",
		"launch_via": "desktop_icon",
	})

	fmt.Println("应用启动事件已记录")
	
	// Output: 应用启动事件已记录
}

// ExampleClient_TrackAppExit 演示应用退出统计
func ExampleClient_TrackAppExit() {
	client := analytics.NewClient(
		"http://localhost:8080",
		"my-awesome-app",
	)
	defer client.Close()

	// 模拟应用运行
	time.Sleep(1 * time.Second)

	// 记录应用退出（同步发送，确保数据不丢失）
	client.TrackAppExit(map[string]interface{}{
		"exit_reason": "user_quit",
	})

	fmt.Println("应用退出事件已记录")
	
	// Output: 应用退出事件已记录
}

// ExampleFullAppLifecycle 演示完整的应用生命周期统计
func ExampleFullAppLifecycle() {
	// 1. 初始化客户端
	client := analytics.NewClient(
		"http://localhost:8080",
		"my-awesome-app",
		analytics.WithDebug(true),
		analytics.WithUserID("user123"),
	)
	defer client.Close()

	// 2. 上报安装信息（首次启动时）
	client.ReportInstall()

	// 3. 记录应用启动
	client.TrackAppLaunch(map[string]interface{}{
		"version": "1.0.0",
	})

	// 4. 应用运行期间的事件统计
	client.Track("button_click", map[string]interface{}{
		"button_name": "login",
		"page":        "home",
	})

	client.Track("feature_used", map[string]interface{}{
		"feature": "export_data",
	})

	// 5. 应用退出前记录
	client.TrackAppExit(map[string]interface{}{
		"exit_reason": "user_quit",
	})

	fmt.Println("完整生命周期事件已记录")
	
	// Output: 完整生命周期事件已记录
}

// ExampleWithEncryption 演示加密通讯
func ExampleWithEncryption() {
	// 创建启用加密的客户端
	client := analytics.NewClient(
		"http://localhost:8080",
		"my-awesome-app",
		analytics.WithEncryption("your-32-byte-secret-key-here!"),
		analytics.WithDebug(true),
	)
	defer client.Close()

	// 所有数据将自动加密传输
	client.ReportInstall()
	client.TrackAppLaunch(nil)

	fmt.Println("加密通讯已启用")
	
	// Output: 加密通讯已启用
}
