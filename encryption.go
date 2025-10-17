package analytics

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	Enabled   bool   // 是否启用加密
	SecretKey string // AES密钥
}

// ClientOption 客户端配置选项
type ClientOption func(*Client)

// WithEncryption 启用加密通讯
func WithEncryption(secretKey string) ClientOption {
	return func(c *Client) {
		c.encryption = &EncryptionConfig{
			Enabled:   true,
			SecretKey: validateKeyLength(secretKey),
		}
	}
}

// encryptRequest 加密请求数据
func (c *Client) encryptRequest(data []byte) ([]byte, error) {
	if c.encryption == nil || !c.encryption.Enabled {
		return data, nil
	}

	// AES加密
	encryptedData, err := aesEncrypt(c.encryption.SecretKey, data)
	if err != nil {
		return nil, fmt.Errorf("encrypt request: %w", err)
	}

	// 构建加密请求结构
	encryptedReq := map[string]interface{}{
		"data":      encryptedData,
		"timestamp": time.Now().Unix(),
	}

	return json.Marshal(encryptedReq)
}

// decryptResponse 解密响应数据
func (c *Client) decryptResponse(body []byte) ([]byte, error) {
	if c.encryption == nil || !c.encryption.Enabled {
		return body, nil
	}

	// 解析加密响应结构
	var encryptedResp struct {
		Encrypted bool   `json:"encrypted"`
		Data      string `json:"data"`
		Timestamp int64  `json:"timestamp"`
	}

	if err := json.Unmarshal(body, &encryptedResp); err != nil {
		// 如果不是加密响应格式，直接返回原始数据
		return body, nil
	}

	if !encryptedResp.Encrypted {
		return body, nil
	}

	// AES解密
	decryptedData, err := aesDecrypt(c.encryption.SecretKey, encryptedResp.Data)
	if err != nil {
		return nil, fmt.Errorf("decrypt response: %w", err)
	}

	return decryptedData, nil
}

// addEncryptionHeaders 添加加密相关的请求头
func (c *Client) addEncryptionHeaders(req *http.Request) {
	if c.encryption != nil && c.encryption.Enabled {
		req.Header.Set("X-Encrypted", "1")
		req.Header.Set("X-Response-Encrypt", "1")
	}
}

// aesEncrypt AES加密
func aesEncrypt(keyStr string, data []byte) (string, error) {
	key := []byte(keyStr)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	encryptBytes := pkcs7Padding(data, blockSize)
	result := make([]byte, len(encryptBytes))

	// 使用CBC加密模式
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	blockMode.CryptBlocks(result, encryptBytes)

	return base64.StdEncoding.EncodeToString(result), nil
}

// aesDecrypt AES解密
func aesDecrypt(keyStr string, dataStr string) ([]byte, error) {
	key := []byte(keyStr)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	data, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	result := make([]byte, len(data))

	blockMode.CryptBlocks(result, data)

	// 去除填充
	result, err = pkcs7UnPadding(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// pkcs7Padding PKCS7填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7UnPadding PKCS7去填充
func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("empty encrypted data")
	}
	unPadding := int(data[length-1])
	if unPadding > length {
		return nil, errors.New("invalid padding")
	}
	return data[:(length - unPadding)], nil
}

// validateKeyLength 验证并调整密钥长度
func validateKeyLength(key string) string {
	keyBytes := []byte(key)
	keyLen := len(keyBytes)

	if keyLen >= 32 {
		return string(keyBytes[:32]) // AES-256
	} else if keyLen >= 24 {
		return string(keyBytes[:24]) // AES-192
	} else if keyLen >= 16 {
		return string(keyBytes[:16]) // AES-128
	} else {
		// 如果密钥长度不足，用0填充到16字节
		paddedKey := make([]byte, 16)
		copy(paddedKey, keyBytes)
		return string(paddedKey)
	}
}

// sendRequest 发送HTTP请求（带加密支持）
func (c *Client) sendRequest(url string, payload []byte) error {
	// 如果启用了加密，先加密数据
	requestData := payload
	var err error
	if c.encryption != nil && c.encryption.Enabled {
		requestData, err = c.encryptRequest(payload)
		if err != nil {
			return err
		}
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	
	// 添加加密相关的请求头
	c.addEncryptionHeaders(req)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 如果启用了加密，解密响应
	if c.encryption != nil && c.encryption.Enabled {
		body, err = c.decryptResponse(body)
		if err != nil {
			return err
		}
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
