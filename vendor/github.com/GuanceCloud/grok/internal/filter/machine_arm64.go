//go:build arm64 && jitasm
// +build arm64,jitasm

package filter

import "strings"

const minAsmStringOpLenARM64 = 32

//go:noescape
func archHasPrefixAsm(s string, prefix string) bool

//go:noescape
func archStringEqualAsm(s string, lit string) bool

func archHasPrefix(s string, prefix string) bool {
	if len(prefix) < minAsmStringOpLenARM64 {
		return strings.HasPrefix(s, prefix)
	}
	return archHasPrefixAsm(s, prefix)
}

func archStringEqual(s string, lit string) bool {
	if len(lit) < minAsmStringOpLenARM64 {
		return s == lit
	}
	return archStringEqualAsm(s, lit)
}
