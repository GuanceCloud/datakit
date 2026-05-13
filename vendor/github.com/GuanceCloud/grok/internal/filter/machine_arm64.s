//go:build arm64 && jitasm
// +build arm64,jitasm

#include "textflag.h"

TEXT ·archHasPrefixAsm(SB), NOSPLIT, $0-40
	MOVD s_len+8(FP), R0
	MOVD prefix_len+24(FP), R1
	CMP R1, R0
	BLT prefix_false
	MOVD s+0(FP), R2
	MOVD prefix+16(FP), R3

prefix_loop:
	CMP $0, R1
	BEQ prefix_true
	MOVBU (R2), R4
	MOVBU (R3), R5
	CMP R5, R4
	BNE prefix_false
	ADD $1, R2
	ADD $1, R3
	SUB $1, R1
	B prefix_loop

prefix_true:
	MOVD $1, R0
	MOVB R0, ret+32(FP)
	RET

prefix_false:
	MOVD $0, R0
	MOVB R0, ret+32(FP)
	RET

TEXT ·archStringEqualAsm(SB), NOSPLIT, $0-40
	MOVD s_len+8(FP), R0
	MOVD lit_len+24(FP), R1
	CMP R1, R0
	BNE equal_false
	MOVD s+0(FP), R2
	MOVD lit+16(FP), R3

equal_loop:
	CMP $0, R1
	BEQ equal_true
	MOVBU (R2), R4
	MOVBU (R3), R5
	CMP R5, R4
	BNE equal_false
	ADD $1, R2
	ADD $1, R3
	SUB $1, R1
	B equal_loop

equal_true:
	MOVD $1, R0
	MOVB R0, ret+32(FP)
	RET

equal_false:
	MOVD $0, R0
	MOVB R0, ret+32(FP)
	RET
