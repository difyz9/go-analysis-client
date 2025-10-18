package analytics

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/host"
)

// InstallInfo 安装信息
type InstallInfo struct {
	Product   string `json:"product"`    // 产品名称
	DeviceID  string `json:"device_id"`  // 设备ID
	Timestamp int64  `json:"timestamp"`  // 时间戳
	Sign      string `json:"sign"`       // 签名
	
	// 设备详细信息（可选）
	Hostname        string `json:"hostname,omitempty"`
	OS              string `json:"os,omitempty"`
	Platform        string `json:"platform,omitempty"`
	PlatformVersion string `json:"platform_version,omitempty"`
	KernelVersion   string `json:"kernel_version,omitempty"`
	KernelArch      string `json:"kernel_arch,omitempty"`
	Uptime          uint64 `json:"uptime,omitempty"`
}

// ReportInstall 上报安装信息（异步）
// 该方法会在后台goroutine中执行，不会阻塞主流程
func (c *Client) ReportInstall() {
	go func() {
		if err := c.reportInstallSync(); err != nil {
			if c.debug && c.logger != nil {
				c.logger.Printf("[Analytics] Failed to report install info: %v", err)
			}
		} else {
			if c.debug && c.logger != nil {
				c.logger.Printf("[Analytics] Successfully reported install info")
			}
		}
	}()
}

// reportInstallSync 同步上报安装信息
func (c *Client) reportInstallSync() error {
	// 获取主机信息
	info, err := host.Info()
	if err != nil {
		return fmt.Errorf("get host info: %w", err)
	}
	
	// 构建安装信息
	timestamp := time.Now().Unix()
	installInfo := &InstallInfo{
		Product:         c.productName,
		DeviceID:        c.deviceID,
		Timestamp:       timestamp,
		Sign:            c.generateInstallSign(c.productName, c.deviceID, timestamp),
		Hostname:        info.Hostname,
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		KernelArch:      info.KernelArch,
		Uptime:          info.Uptime,
	}
	
	// 发送到服务器
	return c.sendInstallInfo(installInfo)
}

// sendInstallInfo 发送安装信息到服务器
func (c *Client) sendInstallInfo(info *InstallInfo) error {
	// 构建请求URL
	url := fmt.Sprintf("%s/api/installs/push", c.serverURL)
	
	// 序列化数据
	data, err := c.marshalJSON(info)
	if err != nil {
		return fmt.Errorf("marshal install info: %w", err)
	}
	
	// 发送请求
	if err := c.sendRequest(url, data); err != nil {
		return fmt.Errorf("send install info: %w", err)
	}
	
	return nil
}

// generateInstallSign 生成安装信息签名
// 签名格式: SHA256(product#device_id#timestamp)
func (c *Client) generateInstallSign(product, deviceID string, timestamp int64) string {
	signStr := fmt.Sprintf("%s#%s#%d", product, deviceID, timestamp)
	hash := sha256.Sum256([]byte(signStr))
	return fmt.Sprintf("%x", hash)
}

// ReportInstallWithCallback 上报安装信息并执行回调
// 适用于需要知道上报结果的场景
func (c *Client) ReportInstallWithCallback(callback func(error)) {
	go func() {
		err := c.reportInstallSync()
		if callback != nil {
			callback(err)
		}
	}()
}

// TrackAppLaunch 记录应用启动事件
// 每次应用启动时调用，用于统计启动次数和启动时间
func (c *Client) TrackAppLaunch(properties map[string]interface{}) {
	if properties == nil {
		properties = make(map[string]interface{})
	}
	
	// 添加默认属性
	properties["session_id"] = c.sessionID
	properties["device_id"] = c.deviceID
	properties["session_started"] = c.sessionStarted.Unix()
	
	// 尝试获取系统信息
	if info, err := host.Info(); err == nil {
		properties["hostname"] = info.Hostname
		properties["os"] = info.OS
		properties["platform"] = info.Platform
		properties["uptime"] = info.Uptime
	}
	
	c.Track("app_launch", properties)
}

// TrackAppExit 记录应用退出事件
// 在应用退出前调用，用于统计会话时长
func (c *Client) TrackAppExit(properties map[string]interface{}) {
	if properties == nil {
		properties = make(map[string]interface{})
	}
	
	// 添加会话时长
	sessionDuration := time.Since(c.sessionStarted).Seconds()
	properties["session_duration"] = sessionDuration
	properties["session_id"] = c.sessionID
	properties["device_id"] = c.deviceID
	
	// 同步发送，确保在应用退出前完成
	c.TrackSync("app_exit", properties)
}
