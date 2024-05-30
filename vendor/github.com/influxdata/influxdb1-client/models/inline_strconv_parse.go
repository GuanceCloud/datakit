package models

import (
	"reflect"
	"strconv"
	"unsafe"
)

// parseIntBytes is a zero-alloc wrapper around strconv.ParseInt.
func parseIntBytes(b []byte, base int, bitSize int) (i int64, err error) {
	return strconv.ParseInt(toUnsafeString(b), base, bitSize)
}

// parseUintBytes is a zero-alloc wrapper around strconv.ParseUint.
func parseUintBytes(b []byte, base int, bitSize int) (i uint64, err error) {
	return strconv.ParseUint(toUnsafeString(b), base, bitSize)
}

// parseFloatBytes is a zero-alloc wrapper around strconv.ParseFloat.
func parseFloatBytes(b []byte, bitSize int) (float64, error) {
	return strconv.ParseFloat(toUnsafeString(b), bitSize)
}

// parseBoolBytes is a zero-alloc wrapper around strconv.ParseBool.
func parseBoolBytes(b []byte) (bool, error) {
	return strconv.ParseBool(toUnsafeString(b))
}

// toUnsafeString converts b to string without memory allocations.
// The returned string is valid only until b is reachable and unmodified.
func toUnsafeString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// toUnsafeBytes converts s to a byte slice without memory allocations.
// The returned byte slice is valid only until s is reachable and unmodified.
func toUnsafeBytes(s string) (b []byte) {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	slh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	slh.Data = sh.Data
	slh.Len = sh.Len
	slh.Cap = sh.Len
	return b
}
