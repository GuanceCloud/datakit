//go:build !linux && !windows
// +build !linux,!windows

package security

func initThirdPartCipher(cipherPath string) (err error) {
	panic("Unimplemented interface")
	return nil
}

func cipherGetCount() int {
	panic("Unimplemented interface")
	return 0
}

func cipherGetInfo(seqno, cipherId, cipherName, _type, blkSize, khSIze uintptr) {
	panic("Unimplemented interface")
}

func cipherEncryptInit(cipherId, key, keySize, cipherPara uintptr) {
	panic("Unimplemented interface")

}

func cipherGetCipherTextSize(cipherId, cipherPara, plainTextSize uintptr) uintptr {
	panic("Unimplemented interface")
	return 0
}

func cipherEncrypt(cipherId, cipherPara, plainText, plainTextSize, cipherText, cipherTextBufSize uintptr) uintptr {
	panic("Unimplemented interface")
	return 0
}

func cipherClean(cipherId, cipherPara uintptr) {
	panic("Unimplemented interface")
}

func cipherDecryptInit(cipherId, key, keySize, cipherPara uintptr) {
	panic("Unimplemented interface")

}

func cipherDecrypt(cipherId, cipherPara, cipherText, cipherTextSize, plainText, plainTextBufSize uintptr) uintptr {
	panic("Unimplemented interface")
	return 0
}
