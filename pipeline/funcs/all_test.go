// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func NewTestingRunner(script string) (*parser.Engine, error) {
	name := "default.p"
	ret1, ret2 := parser.NewEngine(map[string]string{
		"default.p": script,
	}, nil, FuncsMap, FuncsCheckMap)
	if len(ret1) == 1 {
		return ret1[name], nil
	}
	if len(ret2) == 2 {
		return nil, ret2[name]
	}
	return nil, fmt.Errorf("parser func error")
}

func NewTestingRunner2(scripts map[string]string) (map[string]*parser.Engine, map[string]error) {
	return parser.NewEngine(scripts, nil, FuncsMap, FuncsCheckMap)
}
