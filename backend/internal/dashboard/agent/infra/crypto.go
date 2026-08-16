package infra

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// Crypto 密码加解密(AES-GCM)。Agent 密码落库前加密,使用前解密。
type Crypto struct {
	key []byte
}

// NewCrypto 创建加解密器,key 自动 pad/截断到 32 字节
func NewCrypto(secret string) *Crypto {
	return &Crypto{key: padOrTrimKey([]byte(secret))}
}

// Encrypt 加密明文为 base64 密文(AES-GCM,随机 nonce)
func (c *Crypto) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 base64 密文
func (c *Crypto) Decrypt(encoded string) (string, error) {
	if encoded == "" {
		return "", errors.New("空密码")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("密文太短")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func padOrTrimKey(key []byte) []byte {
	const keySize = 32
	if len(key) >= keySize {
		return key[:keySize]
	}
	padded := make([]byte, keySize)
	copy(padded, key)
	return padded
}
