package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"sync"
)

var (
	ErrInvalidCipherText = errors.New("无效的密文")
	ErrDecryptFailed     = errors.New("解密失败")
)

// RSAKeyPair RSA 密钥对
type RSAKeyPair struct {
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}

// Crypto RSA 加密工具
type Crypto struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	mu         sync.RWMutex
}

// NewCrypto 创建新的加密工具
func NewCrypto() (*Crypto, error) {
	c := &Crypto{}
	if err := c.GenerateKeyPair(); err != nil {
		return nil, err
	}
	return c, nil
}

// GenerateKeyPair 生成 RSA 密钥对
func (c *Crypto) GenerateKeyPair() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 生成 2048 位密钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	c.privateKey = privateKey
	c.publicKey = &privateKey.PublicKey
	return nil
}

// GetPublicKey 获取公钥（PEM 格式）
func (c *Crypto) GetPublicKey() (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.publicKey == nil {
		return "", errors.New("公钥未生成")
	}

	pubASN1, err := x509.MarshalPKIXPublicKey(c.publicKey)
	if err != nil {
		return "", err
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	})

	return base64.StdEncoding.EncodeToString(pubBytes), nil
}

// DecryptPassword 解密密码（Base64 + RSA）
func (c *Crypto) DecryptPassword(encryptedPassword string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.privateKey == nil {
		return "", errors.New("私钥未生成")
	}

	// Base64 解码
	cipherText, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", ErrInvalidCipherText
	}

	// RSA 解密
	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, c.privateKey, cipherText)
	if err != nil {
		return "", ErrDecryptFailed
	}

	return string(plainText), nil
}

// RefreshKeyPair 刷新密钥对（定期更换）
func (c *Crypto) RefreshKeyPair() error {
	return c.GenerateKeyPair()
}
