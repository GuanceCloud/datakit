// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

type TestingRunner interface {
	Run(string) error
	GetContentStr(interface{}) (string, error)
	GetContent(interface{}) (interface{}, error)
	IsTag(k interface{}) bool
	Result() *parser.Output
}

func NewTestingRunner(script string) (TestingRunner, error) {
	return parser.NewEngine(script, FuncsMap, nil, false)
}
