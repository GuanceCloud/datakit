//go:build amd64 && jitasm
// +build amd64,jitasm

#include "textflag.h"

TEXT ·archHasPrefixAsm(SB), NOSPLIT, $0-40
	MOVQ s_len+8(FP), AX
	MOVQ prefix_len+24(FP), BX
	CMPQ AX, BX
	JLT prefix_false
	TESTQ BX, BX
	JEQ prefix_true
	MOVQ prefix+16(FP), SI
	MOVQ s+0(FP), DI
	MOVQ BX, CX
	CLD
	REP
	CMPSB
	JNE prefix_false

prefix_true:
	MOVB $1, ret+32(FP)
	RET

prefix_false:
	MOVB $0, ret+32(FP)
	RET

TEXT ·archStringEqualAsm(SB), NOSPLIT, $0-40
	MOVQ s_len+8(FP), AX
	MOVQ lit_len+24(FP), BX
	CMPQ AX, BX
	JNE equal_false
	TESTQ BX, BX
	JEQ equal_true
	MOVQ lit+16(FP), SI
	MOVQ s+0(FP), DI
	MOVQ BX, CX
	CLD
	REP
	CMPSB
	JNE equal_false

equal_true:
	MOVB $1, ret+32(FP)
	RET

equal_false:
	MOVB $0, ret+32(FP)
	RET
