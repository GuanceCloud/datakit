// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package config for crypto:AES and DES.
package config

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

// AESEncrypt using a private key of length 16.
func AESEncrypt(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// PKCS7 padding
	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	plaintext += string(padtext)

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], []byte(plaintext))

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESDecrypt using a private key of length 16.
func AESDecrypt(key []byte, cryptoText string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Remove PKCS7 padding
	padding := int(ciphertext[len(ciphertext)-1])
	if padding > aes.BlockSize {
		return "", fmt.Errorf("invalid padding")
	}
	ciphertext = ciphertext[:len(ciphertext)-padding]

	return string(ciphertext), nil
}

func readFromFile(filepath string) string {
	bts, err := os.ReadFile(path.Clean(filepath))
	if err != nil {
		l.Errorf("ENC read file err=%v", err)
		return ""
	}
	return string(bts)
}

func maskPassword(s string) string {
	if len(s) <= 5 {
		return s
	}
	stars := strings.Repeat("*", len(s)-5)
	return s[:5] + stars
}

func initCrypto(c *Config) {
	if c.Crypto == nil {
		c.Crypto = &configCrpto{}
		return
	}

	if c.Crypto.AESKeyFile != "" {
		datakit.ConfigAESKey = readFromFile(c.Crypto.AESKeyFile)
	}

	if c.Crypto.AESKey != "" {
		datakit.ConfigAESKey = c.Crypto.AESKey
	}
}
