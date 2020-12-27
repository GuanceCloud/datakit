package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"strings"
)

func _padLength(text string, length int) string {
	count := len(text)
	addCount := length - (count % length)

	for i := 0; i < addCount; i++ {
		text = text + " "
	}
	return text
}

func CipherByAES(text, key string) string {
	iv := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	text = _padLength(text, 16)
	key = _padLength(key, 32)

	c, _ := aes.NewCipher([]byte(key))
	cbc := cipher.NewCBCEncrypter(c, iv)

	binText := make([]byte, len(text))
	cbc.CryptBlocks(binText, []byte(text))

	base64Text := base64.StdEncoding.EncodeToString(binText)
	return base64Text
}

func DecipherByAES(text, key string) string {
	iv := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	key = _padLength(key, 32)
	binText, _ := base64.StdEncoding.DecodeString(text)

	c, _ := aes.NewCipher([]byte(key))
	cbc := cipher.NewCBCDecrypter(c, iv)

	binData := make([]byte, len(binText))
	cbc.CryptBlocks(binData, binText)

	rawText := string(binData)
	rawText = strings.Trim(rawText, " ")
	return rawText
}
