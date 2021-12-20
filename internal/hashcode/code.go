// Package hashcode wrap internal hashing functions
package hashcode

import (
	"bytes"

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
	for _, key := range keys {
		buf.WriteString(key)
		buf.WriteString(data[key])
	}

	checksum := md5.Sum(buf.Bytes()) //nolint:gosec

	return string(checksum[:])
}

func GetMD5String32(bt []byte) string {
	return fmt.Sprintf("%X", md5.Sum(bt)) //nolint:gosec
}
