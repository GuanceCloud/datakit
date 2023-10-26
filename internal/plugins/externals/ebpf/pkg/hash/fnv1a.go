// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package hash used to calculate hash
package hash

const (
	// pkg "hash/fnv".
	offset64 = 14695981039346656037
	prime64  = 1099511628211
)

func Fnv1aNew() uint64 {
	return offset64
}

func Fnv1aHashAdd(h uint64, s string) uint64 {
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime64
	}
	return h
}

func Fnv1aHashAddByte(h uint64, s []byte) uint64 {
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime64
	}
	return h
}

func Fnv1aStrHash(s string) uint64 {
	var h uint64 = offset64
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime64
	}
	return h
}

func Fnv1aU8Hash(s []byte) uint64 {
	var h uint64 = offset64
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime64
	}
	return h
}

func Fnv1aHash(v []string) uint64 {
	var h uint64 = offset64
	for _, s := range v {
		for j := 0; j < len(s); j++ {
			h ^= uint64(s[j])
			h *= prime64
		}
	}
	return h
}
