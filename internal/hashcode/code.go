// Package hashcode wrap internal hashing functions
package hashcode

import (
	"bytes"
	"encoding/hex"
	"math/rand"

	// nolint:gosec
	"crypto/md5"
	"fmt"
	"sort"
)

func GenMapHash(data map[string]string) string {
	var (
		i    = 0
		keys = make([]string, len(data))
	)
	for k := range data {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	buf := bytes.NewBuffer(nil)
	for i := range keys {
		buf.WriteString(data[keys[i]])
		buf.WriteString(data[keys[i]])
	}

	checksum := md5.Sum(buf.Bytes()) //nolint:gosec

	return hex.EncodeToString(checksum[:])
}

func GenStringsHash(ss ...string) string {
	var checksum [16]byte
	if len(ss) == 0 {
		buf := make([]byte, 30)
		rand.Read(buf)
		checksum = md5.Sum(buf)
	} else {
		buf := bytes.NewBuffer(nil)
		for i := range ss {
			buf.WriteString(ss[i])
		}
		checksum = md5.Sum(buf.Bytes())
	}

	return hex.EncodeToString(checksum[:])
}

func GetMD5String32(bt []byte) string {
	return fmt.Sprintf("%X", md5.Sum(bt)) //nolint:gosec
}
