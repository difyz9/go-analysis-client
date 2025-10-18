package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v4/host"
)

// generateDeviceID 生成设备ID（用于测试）
func generateDeviceID() string {
	// 尝试获取系统的唯一标识符
	if hostID, err := host.HostID(); err == nil && hostID != "" {
		return hostID
	}

	// 如果获取失败，使用机器信息组合生成稳定ID
	if info, err := host.Info(); err == nil {
		// 使用主机名、操作系统、平台等信息生成一个相对稳定的ID
		combined := fmt.Sprintf("%s-%s-%s-%s",
			info.Hostname,
			info.OS,
			info.Platform,
			info.PlatformVersion)
		return fmt.Sprintf("%x", uuid.NewSHA1(uuid.NameSpaceOID, []byte(combined)))
	}

	// 最后的回退方案：使用 UUID
	return uuid.New().String()
}

func main() {
	fmt.Println("测试设备ID生成功能")
	fmt.Println("===================")

	// 获取主机ID
	hostID, err := host.HostID()
	if err != nil {
		log.Printf("获取主机ID失败: %v", err)
	} else {
		fmt.Printf("主机ID: %s\n", hostID)
	}

	// 获取主机信息
	info, err := host.Info()
	if err != nil {
		log.Printf("获取主机信息失败: %v", err)
	} else {
		fmt.Printf("主机名: %s\n", info.Hostname)
		fmt.Printf("操作系统: %s\n", info.OS)
		fmt.Printf("平台: %s\n", info.Platform)
		fmt.Printf("平台版本: %s\n", info.PlatformVersion)
		fmt.Printf("内核版本: %s\n", info.KernelVersion)
		fmt.Printf("内核架构: %s\n", info.KernelArch)
	}

	fmt.Println("\n生成的设备ID:")
	deviceID := generateDeviceID()
	fmt.Printf("设备ID: %s\n", deviceID)

	// 多次调用确认稳定性
	fmt.Println("\n验证设备ID稳定性（调用3次）:")
	for i := 1; i <= 3; i++ {
		id := generateDeviceID()
		fmt.Printf("第%d次: %s\n", i, id)
	}
}
