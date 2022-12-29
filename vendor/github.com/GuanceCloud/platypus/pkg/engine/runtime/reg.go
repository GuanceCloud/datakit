// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"fmt"

	"github.com/GuanceCloud/platypus/pkg/ast"
)

type PlRegRange uint

const (
	RegR0 = iota
	RegR1
	RegR2
	RegR3
	RegR4
	RegR5
)

// 此处寄存器只存储返回值, 服务于函数的返回
// TODO: 行为变更，修改 pl 内置函数实现

type PlReg struct {
	count uint // less than 6

	regsValDType [6]ast.DType
	r0r5         [6]any
}

func (reg *PlReg) Reset() {
	for i := uint(0); i < reg.count; i++ {
		reg.regsValDType[i] = ast.Void
		reg.r0r5[i] = nil
	}
	reg.count = 0
}

func (reg *PlReg) ReturnAppend(val any, dtype ast.DType) bool {
	if reg.count < 6 {
		reg.regsValDType[reg.count] = dtype
		reg.r0r5[reg.count] = val
		reg.count++
		return true
	}
	return false
}

func (reg *PlReg) Count() int {
	return int(reg.count)
}

func (reg *PlReg) Get(i PlRegRange) (any, ast.DType, error) {
	if i < 6 {
		return reg.r0r5[i], reg.regsValDType[i], nil
	}
	return nil, ast.Invalid, fmt.Errorf("invalid r%d", i)
}
