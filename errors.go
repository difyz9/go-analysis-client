package analytics

import (
	"errors"
	"fmt"
)

// =============================================================================
// 错误类型定义
// =============================================================================

// 预定义的错误类型
var (
	// ErrInvalidConfig 配置无效
	ErrInvalidConfig = errors.New("invalid configuration")
	
	// ErrInvalidServerURL 服务器地址无效
	ErrInvalidServerURL = errors.New("invalid server URL")
	
	// ErrInvalidProductName 产品名称无效
	ErrInvalidProductName = errors.New("invalid product name")
	
	// ErrNetworkTimeout 网络超时
	ErrNetworkTimeout = errors.New("network timeout")
	
	// ErrNetworkFailure 网络请求失败
	ErrNetworkFailure = errors.New("network request failed")
	
	// ErrEncryptionFailed 加密失败
	ErrEncryptionFailed = errors.New("encryption failed")
	
	// ErrDecryptionFailed 解密失败
	ErrDecryptionFailed = errors.New("decryption failed")
	
	// ErrInvalidKey 密钥无效
	ErrInvalidKey = errors.New("invalid encryption key")
	
	// ErrMarshalFailed JSON 序列化失败
	ErrMarshalFailed = errors.New("failed to marshal data")
	
	// ErrUnmarshalFailed JSON 反序列化失败
	ErrUnmarshalFailed = errors.New("failed to unmarshal data")
	
	// ErrServerResponse 服务器响应错误
	ErrServerResponse = errors.New("server response error")
	
	// ErrClientClosed 客户端已关闭
	ErrClientClosed = errors.New("client is closed")
	
	// ErrBufferFull 事件缓冲区已满
	ErrBufferFull = errors.New("event buffer is full")
)

// =============================================================================
// ClientError - 客户端操作错误
// =============================================================================

// ClientError 表示客户端操作中发生的错误
//
// 它包含了错误发生的操作上下文，便于调试和日志记录。
//
// 示例:
//
//	err := &ClientError{
//	    Op:  "Track",
//	    Err: ErrNetworkTimeout,
//	}
//	fmt.Println(err) // Output: Track: network timeout
type ClientError struct {
	// Op 是发生错误的操作名称（如 "Track", "Flush", "ReportInstall"）
	Op string
	
	// Err 是底层错误
	Err error
	
	// Context 包含额外的错误上下文信息（可选）
	Context map[string]interface{}
}

// Error 实现 error 接口
func (e *ClientError) Error() string {
	if e.Context != nil && len(e.Context) > 0 {
		return fmt.Sprintf("%s: %v (context: %v)", e.Op, e.Err, e.Context)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap 返回底层错误，支持 errors.Is 和 errors.As
func (e *ClientError) Unwrap() error {
	return e.Err
}

// =============================================================================
// NetworkError - 网络相关错误
// =============================================================================

// NetworkError 表示网络请求相关的错误
//
// 包含了请求的详细信息，便于重试和诊断。
//
// 示例:
//
//	err := &NetworkError{
//	    Op:         "POST",
//	    URL:        "http://example.com/api/events",
//	    StatusCode: 500,
//	    Err:        ErrServerResponse,
//	}
type NetworkError struct {
	// Op 是 HTTP 操作（GET, POST 等）
	Op string
	
	// URL 是请求的完整 URL
	URL string
	
	// StatusCode 是 HTTP 状态码（如果有响应）
	StatusCode int
	
	// Err 是底层错误
	Err error
	
	// Retryable 指示该错误是否可以重试
	Retryable bool
}

// Error 实现 error 接口
func (e *NetworkError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("%s %s: status %d: %v", e.Op, e.URL, e.StatusCode, e.Err)
	}
	return fmt.Sprintf("%s %s: %v", e.Op, e.URL, e.Err)
}

// Unwrap 返回底层错误
func (e *NetworkError) Unwrap() error {
	return e.Err
}

// =============================================================================
// 辅助函数
// =============================================================================

// newClientError 创建一个新的 ClientError
func newClientError(op string, err error) *ClientError {
	return &ClientError{
		Op:  op,
		Err: err,
	}
}

// newClientErrorWithContext 创建一个带上下文的 ClientError
func newClientErrorWithContext(op string, err error, context map[string]interface{}) *ClientError {
	return &ClientError{
		Op:      op,
		Err:     err,
		Context: context,
	}
}

// newNetworkError 创建一个新的 NetworkError
func newNetworkError(op, url string, statusCode int, err error, retryable bool) *NetworkError {
	return &NetworkError{
		Op:         op,
		URL:        url,
		StatusCode: statusCode,
		Err:        err,
		Retryable:  retryable,
	}
}

// isRetryableError 判断错误是否可以重试
//
// 以下情况被认为可以重试：
//   - 网络超时
//   - 连接失败
//   - 5xx 服务器错误
//   - NetworkError 且 Retryable 为 true
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	// 检查是否是 NetworkError 且标记为可重试
	var netErr *NetworkError
	if errors.As(err, &netErr) {
		return netErr.Retryable
	}
	
	// 检查是否是已知的可重试错误
	if errors.Is(err, ErrNetworkTimeout) || errors.Is(err, ErrNetworkFailure) {
		return true
	}
	
	return false
}

// wrapError 包装错误，添加操作上下文
//
// 如果 err 已经是 ClientError 或 NetworkError，直接返回。
// 否则创建一个新的 ClientError。
func wrapError(op string, err error) error {
	if err == nil {
		return nil
	}
	
	// 如果已经是我们的错误类型，直接返回
	var clientErr *ClientError
	var netErr *NetworkError
	if errors.As(err, &clientErr) || errors.As(err, &netErr) {
		return err
	}
	
	// 否则包装成 ClientError
	return newClientError(op, err)
}
