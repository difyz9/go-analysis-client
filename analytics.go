// Package analytics 提供轻量级、易用的分析统计客户端 SDK
//
// 快速开始:
//
//	client := analytics.NewClient("http://your-server.com", "YourApp")
//	defer client.Close()
//
//	// 可选：上报安装信息
//	client.ReportInstall()
//
//	client.Track("event_name", map[string]interface{}{
//	    "key": "value",
//	})
package analytics

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v4/host"
)

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	Enabled   bool
	SecretKey string
}

// Client 分析客户端
type Client struct {
	serverURL      string
	productName    string
	deviceID       string
	userID         string
	httpClient     *http.Client
	events         chan *Event
	quit           chan struct{}
	wg             sync.WaitGroup
	batchSize      int
	flushInterval  time.Duration
	bufferSize     int
	debug          bool
	logger         Logger
	sessionID      string
	sessionStarted time.Time
	encryption     *EncryptionConfig // 加密配置
}

// Event 表示一个分析事件
type Event struct {
	Name       string                 `json:"name"`
	Timestamp  int64                  `json:"timestamp"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	
	// 可选：Google Analytics 风格分类
	Category string  `json:"category,omitempty"`
	Action   string  `json:"action,omitempty"`
	Label    string  `json:"label,omitempty"`
	Value    float64 `json:"value,omitempty"`
}

// Logger 日志接口
type Logger interface {
	Printf(format string, v ...interface{})
}

// ClientOption 客户端配置选项
type ClientOption func(*Client)

// WithDeviceID 设置设备ID
func WithDeviceID(deviceID string) ClientOption {
	return func(c *Client) {
		c.deviceID = deviceID
	}
}

// WithUserID 设置用户ID
func WithUserID(userID string) ClientOption {
	return func(c *Client) {
		c.userID = userID
	}
}

// WithTimeout 设置HTTP超时
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithBatchSize 设置批量发送大小
func WithBatchSize(size int) ClientOption {
	return func(c *Client) {
		c.batchSize = size
	}
}

// WithFlushInterval 设置自动刷新间隔
func WithFlushInterval(interval time.Duration) ClientOption {
	return func(c *Client) {
		c.flushInterval = interval
	}
}

// WithBufferSize 设置事件缓冲区大小
func WithBufferSize(size int) ClientOption {
	return func(c *Client) {
		c.bufferSize = size
	}
}

// WithDebug 启用调试模式
func WithDebug(debug bool) ClientOption {
	return func(c *Client) {
		c.debug = debug
	}
}

// WithLogger 设置自定义日志器
func WithLogger(logger Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithEncryption 启用 AES 加密传输
// secretKey 必须是 16、24 或 32 字节长度，对应 AES-128、AES-192 或 AES-256
func WithEncryption(secretKey string) ClientOption {
	return func(c *Client) {
		c.encryption = &EncryptionConfig{
			Enabled:   true,
			SecretKey: secretKey,
		}
	}
}

// NewClient 创建新的分析客户端
//
// serverURL: 分析服务器地址，例如 "http://localhost:8080"
// productName: 产品名称，用于区分不同应用
// opts: 可选配置项
//
// 注意：NewClient 不会自动上报安装信息。如需上报，请调用 client.ReportInstall()
//
// 示例：
//
//	client := analytics.NewClient("http://localhost:8080", "MyApp")
//	defer client.Close()
//	
//	// 可选：上报安装信息
//	client.ReportInstall()
//	
//	// 发送事件
//	client.Track("button_click", map[string]interface{}{"button": "submit"})
func NewClient(serverURL, productName string, opts ...ClientOption) *Client {
	client := &Client{
		serverURL:     serverURL,
		productName:   productName,
		deviceID:      generateDeviceID(),
		httpClient:    &http.Client{Timeout: 10 * time.Second},
		batchSize:     20,
		flushInterval: 5 * time.Second,
		bufferSize:    1000,
		debug:         false,
		quit:          make(chan struct{}),
		sessionID:     uuid.New().String(),
		sessionStarted: time.Now(),
	}
	
	// 应用配置选项
	for _, opt := range opts {
		opt(client)
	}
	
	// 创建事件通道
	client.events = make(chan *Event, client.bufferSize)
	
	// 启动后台处理
	client.wg.Add(1)
	go client.processEvents()
	
	return client
}

// Track 发送一个简单事件（异步）
//
//	client.Track("button_click", map[string]interface{}{
//	    "button_name": "login",
//	})
func (c *Client) Track(eventName string, properties map[string]interface{}) {
	event := &Event{
		Name:       eventName,
		Timestamp:  time.Now().Unix(),
		Properties: properties,
	}
	
	select {
	case c.events <- event:
		// 成功加入队列
	default:
		if c.debug && c.logger != nil {
			c.logger.Printf("[Analytics] Event buffer full, dropping event: %s", eventName)
		}
	}
}

// TrackEvent 发送分类事件（Google Analytics 风格）
//
// Deprecated: Use Track instead for better flexibility.
// Migration example:
//
//	Old: client.TrackEvent("user", "login", "email", 1)
//	New: client.Track("user_login", map[string]interface{}{
//	    "category": "user",
//	    "action": "login",
//	    "label": "email",
//	    "value": 1,
//	})
func (c *Client) TrackEvent(category, action, label string, value float64) {
	event := &Event{
		Name:      action,
		Timestamp: time.Now().Unix(),
		Category:  category,
		Action:    action,
		Label:     label,
		Value:     value,
	}
	
	select {
	case c.events <- event:
		// 成功加入队列
	default:
		if c.debug && c.logger != nil {
			c.logger.Printf("[Analytics] Event buffer full, dropping event: %s/%s", category, action)
		}
	}
}

// TrackSync 同步发送事件（阻塞直到发送完成）
//
// Deprecated: Use Track followed by Flush for better control.
// Migration example:
//
//	Old: err := client.TrackSync("user_login", properties)
//	New: client.Track("user_login", properties)
//	     client.Flush()
func (c *Client) TrackSync(eventName string, properties map[string]interface{}) error {
	event := &Event{
		Name:       eventName,
		Timestamp:  time.Now().Unix(),
		Properties: properties,
	}
	
	return c.sendEvents([]*Event{event})
}

// TrackBatch 批量发送事件
func (c *Client) TrackBatch(events []Event) {
	for _, event := range events {
		evt := event
		evt.Timestamp = time.Now().Unix()
		
		select {
		case c.events <- &evt:
			// 成功加入队列
		default:
			if c.debug && c.logger != nil {
				c.logger.Printf("[Analytics] Event buffer full, dropping event: %s", event.Name)
			}
		}
	}
}

// Flush 立即发送所有缓冲的事件
func (c *Client) Flush() {
	// 发送信号通知立即刷新
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			return
		case <-ticker.C:
			if len(c.events) == 0 {
				return
			}
		}
	}
}

// Close 关闭客户端，确保所有事件发送完成
func (c *Client) Close() error {
	close(c.quit)
	c.wg.Wait()
	return nil
}

// processEvents 后台处理事件
func (c *Client) processEvents() {
	defer c.wg.Done()
	
	ticker := time.NewTicker(c.flushInterval)
	defer ticker.Stop()
	
	batch := make([]*Event, 0, c.batchSize)
	
	for {
		select {
		case <-c.quit:
			// 发送剩余事件
			if len(batch) > 0 {
				c.sendEvents(batch)
			}
			// 清空通道中的剩余事件
			for len(c.events) > 0 {
				event := <-c.events
				batch = append(batch, event)
				if len(batch) >= c.batchSize {
					c.sendEvents(batch)
					batch = make([]*Event, 0, c.batchSize)
				}
			}
			if len(batch) > 0 {
				c.sendEvents(batch)
			}
			return
			
		case event := <-c.events:
			batch = append(batch, event)
			if len(batch) >= c.batchSize {
				c.sendEvents(batch)
				batch = make([]*Event, 0, c.batchSize)
			}
			
		case <-ticker.C:
			if len(batch) > 0 {
				c.sendEvents(batch)
				batch = make([]*Event, 0, c.batchSize)
			}
		}
	}
}

// sendEvents 发送事件到服务器
func (c *Client) sendEvents(events []*Event) error {
	if len(events) == 0 {
		return nil
	}
	
	// 构建请求体
	payload := map[string]interface{}{
		"product":    c.productName,
		"device_id":  c.deviceID,
		"user_id":    c.userID,
		"session_id": c.sessionID,
		"events":     events,
	}
	
	data, err := json.Marshal(payload)
	if err != nil {
		if c.debug && c.logger != nil {
			c.logger.Printf("[Analytics] Failed to marshal events: %v", err)
		}
		return newClientError("sendEvents", fmt.Errorf("%w: %v", ErrMarshalFailed, err))
	}
	
	// 如果启用了加密，加密数据
	var requestBody []byte
	var contentType string
	
	if c.encryption != nil && c.encryption.Enabled {
		// 使用 AES 加密
		encrypted, err := AESEncrypt([]byte(c.encryption.SecretKey), data)
		if err != nil {
			if c.debug && c.logger != nil {
				c.logger.Printf("[Analytics] Failed to encrypt events: %v", err)
			}
			return newClientError("sendEvents", fmt.Errorf("%w: %v", ErrEncryptionFailed, err))
		}
		
		// 构建加密请求体
		encryptedPayload := map[string]string{
			"data": encrypted,
		}
		requestBody, err = json.Marshal(encryptedPayload)
		if err != nil {
			return newClientError("sendEvents", fmt.Errorf("%w: %v", ErrMarshalFailed, err))
		}
		contentType = "application/json"
		
		if c.debug && c.logger != nil {
			c.logger.Printf("[Analytics] Events encrypted, sending %d bytes", len(requestBody))
		}
	} else {
		// 不加密，直接发送
		requestBody = data
		contentType = "application/json"
	}
	
	// 发送请求
	url := fmt.Sprintf("%s/api/events/batch", c.serverURL)
	resp, err := c.httpClient.Post(url, contentType, bytes.NewReader(requestBody))
	if err != nil {
		if c.debug && c.logger != nil {
			c.logger.Printf("[Analytics] Failed to send events: %v", err)
		}
		return newNetworkError("POST", url, 0, fmt.Errorf("%w: %v", ErrNetworkFailure, err), true)
	}
	defer resp.Body.Close()
	
	// 检查 HTTP 状态码
	if resp.StatusCode >= 500 {
		// 5xx 错误，可以重试
		return newNetworkError("POST", url, resp.StatusCode, ErrServerResponse, true)
	} else if resp.StatusCode >= 400 {
		// 4xx 错误，通常不应该重试
		return newNetworkError("POST", url, resp.StatusCode, ErrServerResponse, false)
	}
	
	if c.debug && c.logger != nil {
		c.logger.Printf("[Analytics] Successfully sent %d events", len(events))
	}
	
	return nil
}

// generateDeviceID 生成设备ID
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

// GetDeviceID 获取当前设备ID
func (c *Client) GetDeviceID() string {
	return c.deviceID
}

// GetSessionID 获取当前会话ID
func (c *Client) GetSessionID() string {
	return c.sessionID
}

// SetUserID 设置用户ID
func (c *Client) SetUserID(userID string) {
	c.userID = userID
}

// marshalJSON 序列化JSON数据
func (c *Client) marshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// getLocalIPs 返回主机上所有非回环 IPv4 地址的列表
func getLocalIPs() []string {
	ips := make([]string, 0)
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	for _, iface := range ifaces {
		// 忽略 down 或 loopback 接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // 只收集 IPv4
			}
			ips = append(ips, ip.String())
		}
	}
	return ips
}

// getPublicIP 通过简单的外部服务获取公网 IP，失败则返回空字符串
func getPublicIP(client *http.Client) string {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	// 使用可靠的公共 IP 服务
	url := "https://api.ipify.org?format=text"
	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(bytes.TrimSpace(b))
}

// marshalSHA256 返回输入字符串的十六进制 SHA256 值
func marshalSHA256(s string) (string, error) {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:]), nil
}

// =============================================================================
// Installation & App Lifecycle Tracking
// =============================================================================

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

// reportInstallSync 同步上报安装信息
func (c *Client) reportInstallSync() error {
	// 获取主机信息
	info, err := host.Info()
	if err != nil {
		return newClientError("reportInstallSync", fmt.Errorf("get host info: %w", err))
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
		return newClientError("sendInstallInfo", fmt.Errorf("%w: %v", ErrMarshalFailed, err))
	}
	
	// 发送请求
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return newNetworkError("POST", url, 0, fmt.Errorf("%w: %v", ErrNetworkFailure, err), true)
	}
	defer resp.Body.Close()
	
	// 检查 HTTP 状态码
	if resp.StatusCode >= 500 {
		return newNetworkError("POST", url, resp.StatusCode, ErrServerResponse, true)
	} else if resp.StatusCode >= 400 {
		return newNetworkError("POST", url, resp.StatusCode, ErrServerResponse, false)
	}
	
	if c.debug && c.logger != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		c.logger.Printf("[Analytics] Install info response: %s", string(body))
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
	
	// 发送退出事件并立即刷新，确保在应用退出前完成
	c.Track("app_exit", properties)
	c.Flush()
}
