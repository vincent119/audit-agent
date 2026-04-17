// Package crypto 提供 AES-256-GCM 加解密功能，使用隨機 salt 衍生金鑰。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

const saltSize = 16

// Encrypt 使用 AES-256-GCM 加密明文，回傳 base64 編碼的密文。
func Encrypt(key, plaintext string) (string, error) {
	if key == "" {
		return "", errors.New("key must not be empty")
	}
	if plaintext == "" {
		return "", errors.New("plaintext must not be empty")
	}

	// 產生隨機 salt
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	// 用 sha256(key + salt) 衍生 32-byte AES key
	derivedKey := deriveKey(key, salt)

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 產生隨機 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密並組合: salt + nonce + ciphertext
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	result := make([]byte, 0, len(salt)+len(nonce)+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return base64.StdEncoding.EncodeToString(result), nil
}

// Decrypt 解密 base64 編碼的 AES-256-GCM 密文，回傳明文。
func Decrypt(key, ciphertext string) (string, error) {
	if key == "" {
		return "", errors.New("key must not be empty")
	}
	if ciphertext == "" {
		return "", errors.New("ciphertext must not be empty")
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// 分離 salt
	if len(data) < saltSize {
		return "", errors.New("ciphertext data too short")
	}
	salt := data[:saltSize]
	remaining := data[saltSize:]

	derivedKey := deriveKey(key, salt)

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 分離 nonce 與實際密文
	nonceSize := gcm.NonceSize()
	if len(remaining) < nonceSize {
		return "", errors.New("ciphertext data too short")
	}
	nonce := remaining[:nonceSize]
	encryptedData := remaining[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// deriveKey 使用 SHA-256 從 key 和 salt 衍生 32-byte 金鑰。
func deriveKey(key string, salt []byte) []byte {
	h := sha256.New()
	h.Write([]byte(key))
	h.Write(salt)
	return h.Sum(nil)
}
