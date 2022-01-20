package cliutils

import (
	"crypto/aes"
	"crypto/cipher"

	// nolint:gosec
	"crypto/md5"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	mrand "math/rand"
	"sync"
	"time"

	"github.com/rs/xid"
)

type Sem struct {
	sem chan interface{}
}

func NewSem() *Sem {
	return &Sem{sem: make(chan interface{})}
}

func (s *Sem) Close() {
	select {
	case <-s.sem:
	// pass: s.sem has been closed before
	default:
		close(s.sem)
	}
}

func (s *Sem) Wait() <-chan interface{} {
	return s.sem
}

func WgWait(wg *sync.WaitGroup, timeout int) {
	c := make(chan interface{})

	go func() {
		defer close(c)
		wg.Wait()
	}()

	if timeout > 0 {
		select {
		case <-c:
		case <-time.After(time.Second * time.Duration(timeout)):
		}
	} else {
		<-c
	}
}

var (
	letterBytes   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = uint(6)              //nolint:gomnd // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   //nolint:gomnd
)

func CreateRandomString(n int) string {
	src := mrand.NewSource(time.Now().UnixNano())

	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & int64(letterIdxMask)); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func XID(p string) string {
	return p + xid.New().String()
}

func SizeFmt(n int64) string {
	f := float64(n)

	unit := []string{"", "K", "M", "G", "T", "P", "E", "Z"}
	for _, u := range unit {
		if math.Abs(f) < 1024.0 { //nolint:gomnd
			return fmt.Sprintf("%3.4f%sB", f, u)
		}
		f /= 1024.0 //nolint:gomnd
	}
	return fmt.Sprintf("%3.4fYB", f)
}

func Md5Hash(data []byte) string {
	hasher := md5.New() //nolint:gosec
	if _, err := hasher.Write(data); err != nil {
		panic(err)
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

func Encrypt(data, phrase []byte) ([]byte, error) {
	md5Cipher := Md5Hash(phrase)

	block, _ := aes.NewCipher([]byte(md5Cipher))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(crand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, data, nil), nil
}

func Decrypt(endata, phrase []byte) ([]byte, error) {
	md5Cipher := Md5Hash(phrase)

	block, err := aes.NewCipher([]byte(md5Cipher))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := endata[:nonceSize], endata[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
