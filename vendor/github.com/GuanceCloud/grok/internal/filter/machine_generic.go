//go:build !jitasm || (!amd64 && !arm64)
// +build !jitasm !amd64,!arm64

package filter

import "strings"

func archHasPrefix(s string, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func archStringEqual(s string, lit string) bool {
	return s == lit
}
