//go:build amd64 && jitasm
// +build amd64,jitasm

package filter

import "strings"

const minAsmStringOpLenAMD64 = 32

//go:noescape
func archHasPrefixAsm(s string, prefix string) bool

//go:noescape
func archStringEqualAsm(s string, lit string) bool

func archHasPrefix(s string, prefix string) bool {
	if len(prefix) < minAsmStringOpLenAMD64 {
		return strings.HasPrefix(s, prefix)
	}
	return archHasPrefixAsm(s, prefix)
}

func archStringEqual(s string, lit string) bool {
	if len(lit) < minAsmStringOpLenAMD64 {
		return s == lit
	}
	return archStringEqualAsm(s, lit)
}
