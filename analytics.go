// Package analytics 提供轻量级、易用的分析统计客户端 SDK
//
// 快速开始:
//
//	client := analytics.NewClient("http://your-server.com", "YourApp")
//	defer client.Close()
//
//	client.Track("event_name", map[string]interface{}{
//	    "key": "value",
//	})
package analytics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v4/host"
)

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

// NewClient 创建新的分析客户端
//
// serverURL: 分析服务器地址，例如 "http://localhost:8080"
// productName: 产品名称，用于区分不同应用
// opts: 可选配置项
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
//	client.TrackEvent("user", "login", "email", 1)
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
		return fmt.Errorf("marshal events: %w", err)
	}
	
	// 使用加密发送请求
	url := fmt.Sprintf("%s/api/events/batch", c.serverURL)
	if err := c.sendRequest(url, data); err != nil {
		if c.debug && c.logger != nil {
			c.logger.Printf("[Analytics] Failed to send events: %v", err)
		}
		return err
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
