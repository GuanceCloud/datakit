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

type PlProcStack struct {
	data   map[string]*Varb
	before *PlProcStack

	checkPattern map[string]*grok.GrokPattern
}

func (stack *PlProcStack) SetPattern(patternAlias string, grokPattern *grok.GrokPattern) {
	if stack.checkPattern == nil {
		stack.checkPattern = make(map[string]*grok.GrokPattern)
	}
	stack.checkPattern[patternAlias] = grokPattern
}

func (stack *PlProcStack) GetPattern(pattern string) (*grok.GrokPattern, bool) {
	cur := stack

	for {
		// 在 cur stack
		if v, ok := cur.checkPattern[pattern]; ok {
			return v, ok
		}
		// 尝试在上一级查找
		if cur.before != nil {
			cur = cur.before
		} else {
			break
		}
	}

	return nil, false
}

func (stack *PlProcStack) Set(key string, value any, dType ast.DType) {
	cur := stack

	for {
		// 在 cur stack
		if v, ok := cur.data[key]; ok {
			v.DType = dType
			v.Value = value
			return
		}
		// 尝试在上一级查找
		if cur.before != nil {
			cur = cur.before
		} else {
			break
		}
	}

	// new
	stack.data[key] = &Varb{
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

func (stack *PlProcStack) Get(key string) (*Varb, error) {
	cur := stack

	for {
		// 在 cur stack
		if v, ok := cur.data[key]; ok {
			return v, nil
		}
		// 尝试在上一级查找
		if cur.before != nil {
			cur = cur.before
		} else {
			break
		}
	}

	return nil, fmt.Errorf("not found")
}

func (stack *PlProcStack) Clear() {
	stack.checkPattern = nil
	for k := range stack.data {
		delete(stack.data, k)
	}
}
