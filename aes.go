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
)

// AESClient 支持 AES 加密的客户端
type AESClient struct {
	BaseURL   string
	SecretKey string
	Client    *http.Client
}

// NewAESClient 创建新的 AES 客户端
func NewAESClient(baseURL, secretKey string) *AESClient {
	return &AESClient{
		BaseURL:   baseURL,
		SecretKey: secretKey,
		Client:    &http.Client{},
	}
}

// PKCS7Padding PKCS7 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// PKCS7UnPadding PKCS7 去填充
func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("invalid padding size")
	}
	unpadding := int(data[length-1])
	if unpadding > length {
		return nil, errors.New("invalid padding")
	}
	return data[:(length - unpadding)], nil
}

// AESEncrypt AES 加密
func AESEncrypt(key []byte, plaintext []byte) (string, error) {
	// 确保密钥长度为 16/24/32 字节
	keyLen := len(key)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		// 调整密钥长度
		if keyLen < 16 {
			key = append(key, bytes.Repeat([]byte{0}, 16-keyLen)...)
		} else if keyLen > 32 {
			key = key[:32]
		} else if keyLen > 24 {
			key = key[:24]
		} else if keyLen > 16 {
			key = key[:16]
		}
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// PKCS7 填充
	paddedData := pkcs7Padding(plaintext, block.BlockSize())

	// 使用前 16 字节作为 IV
	iv := key[:block.BlockSize()]

	ciphertext := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedData)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESDecrypt AES 解密
func AESDecrypt(key []byte, ciphertextBase64 string) ([]byte, error) {
	// 确保密钥长度
	keyLen := len(key)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		if keyLen < 16 {
			key = append(key, bytes.Repeat([]byte{0}, 16-keyLen)...)
		} else if keyLen > 32 {
			key = key[:32]
		} else if keyLen > 24 {
			key = key[:24]
		} else if keyLen > 16 {
			key = key[:16]
		}
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext)%block.BlockSize() != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	iv := key[:block.BlockSize()]
	mode := cipher.NewCBCDecrypter(block, iv)

	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// 去填充
	return pkcs7UnPadding(plaintext)
}

// PostEncrypted 发送加密的 POST 请求
func (c *AESClient) PostEncrypted(path string, data interface{}) ([]byte, error) {
	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal data error: %w", err)
	}

	// 加密数据
	encryptedData, err := AESEncrypt([]byte(c.SecretKey), jsonData)
	if err != nil {
		return nil, fmt.Errorf("encrypt data error: %w", err)
	}

	// 构造加密请求体
	encryptedRequest := map[string]string{
		"data": encryptedData,
	}

	reqBody, err := json.Marshal(encryptedRequest)
	if err != nil {
		return nil, fmt.Errorf("marshal encrypted request error: %w", err)
	}

	// 创建请求
	url := c.BaseURL + path
	req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Encrypted", "true")       // 告诉服务器请求已加密
	req.Header.Set("X-Response-Encrypt", "true") // 要求服务器加密响应

	// 发送请求
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request error: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response error: %w", err)
	}

	// 检查响应是否加密
	if resp.Header.Get("X-Encrypted") == "true" {
		// 解析加密响应
		var encryptedResp map[string]string
		if err := json.Unmarshal(respBody, &encryptedResp); err != nil {
			return nil, fmt.Errorf("unmarshal encrypted response error: %w", err)
		}

		// 解密响应数据
		decryptedData, err := AESDecrypt([]byte(c.SecretKey), encryptedResp["data"])
		if err != nil {
			return nil, fmt.Errorf("decrypt response error: %w", err)
		}

		return decryptedData, nil
	}

	return respBody, nil
}

// PostPlain 发送普通（未加密）的 POST 请求
func (c *AESClient) PostPlain(path string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal data error: %w", err)
	}

	url := c.BaseURL + path
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request error: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
