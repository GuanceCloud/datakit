// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package hashcode wrap internal hashing functions
package hashcode

import (
	"bytes"

	// nolint:gosec
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
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

	checksum := md5.Sum(buf.Bytes()) // nolint:gosec

	return hex.EncodeToString(checksum[:])
}

func GenStringsHash(ss ...string) string {
	var checksum [16]byte
	if len(ss) == 0 {
		buf := make([]byte, 30)
		rand.Read(buf)          // nolint:gosec
		checksum = md5.Sum(buf) // nolint:gosec
	} else {
		buf := bytes.NewBuffer(nil)
		for i := range ss {
			buf.WriteString(ss[i])
		}
		checksum = md5.Sum(buf.Bytes()) // nolint:gosec
	}

	return hex.EncodeToString(checksum[:])
}

func GetMD5String32(bt []byte) string {
	return fmt.Sprintf("%X", md5.Sum(bt)) // nolint:gosec
}
