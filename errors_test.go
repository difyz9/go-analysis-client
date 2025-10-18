package analytics

import (
	"errors"
	"testing"
)

// TestClientError 测试 ClientError 类型
func TestClientError(t *testing.T) {
	tests := []struct {
		name    string
		err     *ClientError
		wantMsg string
	}{
		{
			name: "基础错误",
			err: &ClientError{
				Op:  "Track",
				Err: ErrNetworkTimeout,
			},
			wantMsg: "Track: network timeout",
		},
		{
			name: "带上下文的错误",
			err: &ClientError{
				Op:  "SendEvents",
				Err: ErrMarshalFailed,
				Context: map[string]interface{}{
					"event_count": 10,
				},
			},
			wantMsg: "SendEvents: failed to marshal data (context: map[event_count:10])",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("ClientError.Error() = %v, want %v", got, tt.wantMsg)
			}
		})
	}
}

// TestClientError_Unwrap 测试 ClientError 的 Unwrap 功能
func TestClientError_Unwrap(t *testing.T) {
	baseErr := ErrNetworkTimeout
	clientErr := &ClientError{
		Op:  "Track",
		Err: baseErr,
	}
	
	// 测试 errors.Is
	if !errors.Is(clientErr, ErrNetworkTimeout) {
		t.Error("errors.Is failed for ClientError")
	}
	
	// 测试 Unwrap
	if unwrapped := clientErr.Unwrap(); unwrapped != baseErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, baseErr)
	}
}

// TestNetworkError 测试 NetworkError 类型
func TestNetworkError(t *testing.T) {
	tests := []struct {
		name    string
		err     *NetworkError
		wantMsg string
	}{
		{
			name: "无状态码",
			err: &NetworkError{
				Op:  "POST",
				URL: "http://example.com/api/events",
				Err: ErrNetworkFailure,
			},
			wantMsg: "POST http://example.com/api/events: network request failed",
		},
		{
			name: "带状态码",
			err: &NetworkError{
				Op:         "POST",
				URL:        "http://example.com/api/events",
				StatusCode: 500,
				Err:        ErrServerResponse,
			},
			wantMsg: "POST http://example.com/api/events: status 500: server response error",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("NetworkError.Error() = %v, want %v", got, tt.wantMsg)
			}
		})
	}
}

// TestNetworkError_Unwrap 测试 NetworkError 的 Unwrap 功能
func TestNetworkError_Unwrap(t *testing.T) {
	baseErr := ErrServerResponse
	netErr := &NetworkError{
		Op:         "POST",
		URL:        "http://example.com",
		StatusCode: 500,
		Err:        baseErr,
	}
	
	// 测试 errors.Is
	if !errors.Is(netErr, ErrServerResponse) {
		t.Error("errors.Is failed for NetworkError")
	}
	
	// 测试 Unwrap
	if unwrapped := netErr.Unwrap(); unwrapped != baseErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, baseErr)
	}
}

// TestIsRetryableError 测试可重试错误判断
func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil 错误",
			err:  nil,
			want: false,
		},
		{
			name: "网络超时错误",
			err:  ErrNetworkTimeout,
			want: true,
		},
		{
			name: "网络失败错误",
			err:  ErrNetworkFailure,
			want: true,
		},
		{
			name: "可重试的 NetworkError",
			err: &NetworkError{
				Op:        "POST",
				URL:       "http://example.com",
				Err:       ErrServerResponse,
				Retryable: true,
			},
			want: true,
		},
		{
			name: "不可重试的 NetworkError",
			err: &NetworkError{
				Op:        "POST",
				URL:       "http://example.com",
				Err:       ErrServerResponse,
				Retryable: false,
			},
			want: false,
		},
		{
			name: "普通错误",
			err:  errors.New("some error"),
			want: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetryableError(tt.err); got != tt.want {
				t.Errorf("isRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestWrapError 测试错误包装
func TestWrapError(t *testing.T) {
	tests := []struct {
		name   string
		op     string
		err    error
		wantOp string
	}{
		{
			name:   "包装普通错误",
			op:     "Track",
			err:    errors.New("test error"),
			wantOp: "Track",
		},
		{
			name:   "nil 错误",
			op:     "Track",
			err:    nil,
			wantOp: "",
		},
		{
			name: "已经是 ClientError",
			op:   "NewOp",
			err: &ClientError{
				Op:  "OriginalOp",
				Err: ErrNetworkTimeout,
			},
			wantOp: "OriginalOp", // 应该保持原来的操作名
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapError(tt.op, tt.err)
			
			if tt.err == nil {
				if got != nil {
					t.Errorf("wrapError() = %v, want nil", got)
				}
				return
			}
			
			var clientErr *ClientError
			if errors.As(got, &clientErr) {
				if clientErr.Op != tt.wantOp {
					t.Errorf("wrapError() Op = %v, want %v", clientErr.Op, tt.wantOp)
				}
			}
		})
	}
}

// TestNewClientError 测试创建 ClientError
func TestNewClientError(t *testing.T) {
	err := newClientError("TestOp", ErrNetworkTimeout)
	
	if err.Op != "TestOp" {
		t.Errorf("Op = %v, want TestOp", err.Op)
	}
	
	if !errors.Is(err, ErrNetworkTimeout) {
		t.Error("error should wrap ErrNetworkTimeout")
	}
}

// TestNewClientErrorWithContext 测试创建带上下文的 ClientError
func TestNewClientErrorWithContext(t *testing.T) {
	context := map[string]interface{}{
		"count": 5,
		"type":  "event",
	}
	
	err := newClientErrorWithContext("TestOp", ErrMarshalFailed, context)
	
	if err.Op != "TestOp" {
		t.Errorf("Op = %v, want TestOp", err.Op)
	}
	
	if err.Context == nil {
		t.Error("Context should not be nil")
	}
	
	if count, ok := err.Context["count"]; !ok || count != 5 {
		t.Error("Context should contain count = 5")
	}
}

// TestNewNetworkError 测试创建 NetworkError
func TestNewNetworkError(t *testing.T) {
	err := newNetworkError("POST", "http://example.com", 500, ErrServerResponse, true)
	
	if err.Op != "POST" {
		t.Errorf("Op = %v, want POST", err.Op)
	}
	
	if err.URL != "http://example.com" {
		t.Errorf("URL = %v, want http://example.com", err.URL)
	}
	
	if err.StatusCode != 500 {
		t.Errorf("StatusCode = %v, want 500", err.StatusCode)
	}
	
	if !err.Retryable {
		t.Error("Retryable should be true")
	}
	
	if !errors.Is(err, ErrServerResponse) {
		t.Error("error should wrap ErrServerResponse")
	}
}
