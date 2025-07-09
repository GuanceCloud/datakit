// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"fmt"

	"github.com/GuanceCloud/grok"
	"github.com/GuanceCloud/platypus/pkg/ast"
)

type Varb struct {
	Value any
	DType ast.DType
}

type Stack struct {
	Data   map[string]*Varb
	Before *Stack

	CheckPattern map[string]*grok.GrokPattern
}

func NewStack() *Stack {
	return &Stack{
		Data: map[string]*Varb{},
	}
}

func (stack *Stack) SetPattern(patternAlias string, grokPattern *grok.GrokPattern) {
	if stack.CheckPattern == nil {
		stack.CheckPattern = make(map[string]*grok.GrokPattern)
	}
	stack.CheckPattern[patternAlias] = grokPattern
}

func (stack *Stack) GetPattern(pattern string) (*grok.GrokPattern, bool) {
	cur := stack

	for {
		// 在 cur stack
		if v, ok := cur.CheckPattern[pattern]; ok {
			return v, ok
		}
		// 尝试在上一级查找
		if cur.Before != nil {
			cur = cur.Before
		} else {
			break
		}
	}

	return nil, false
}

func (stack *Stack) Set(key string, value any, dType ast.DType) {
	cur := stack

	for {
		// 在 cur stack
		if v, ok := cur.Data[key]; ok {
			v.DType = dType
			v.Value = value
			return
		}
		// 尝试在上一级查找
		if cur.Before != nil {
			cur = cur.Before
		} else {
			break
		}
	}

	// new
	stack.Data[key] = &Varb{
		Value: value,
		DType: dType,
	}
}

// func (stack *PlProcStack) SetLocal(key string, value any, dType DType) {
// 	stack.data[key] = &Varb{
// 		Name:  key,
// 		Value: value,
// 		DType: dType,
// 	}
// }

func (stack *Stack) Get(key string) (*Varb, error) {
	cur := stack

	for {
		// 在 cur stack
		if v, ok := cur.Data[key]; ok {
			return v, nil
		}
		// 尝试在上一级查找
		if cur.Before != nil {
			cur = cur.Before
		} else {
			break
		}
	}

	return nil, fmt.Errorf("not found")
}

func (stack *Stack) Clear() {
	stack.CheckPattern = nil
	for k := range stack.Data {
		delete(stack.Data, k)
	}
}
