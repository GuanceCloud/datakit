// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package grok used to parses grok patterns in Go
package grok

import (
	"fmt"
	"regexp"

	"github.com/spf13/cast"
)

var (
	valid    = regexp.MustCompile(`^\w+([-.]\w+)*(:([-.\w]+)(:(string|float|int|bool))?)?$`)
	normal   = regexp.MustCompile(`%{([\w-.]+(?::[\w-.]+(?::[\w-.]+)?)?)}`)
	symbolic = regexp.MustCompile(`\W`)
)

type Grok struct {
	GlobalDenormalizedPatterns map[string]*GrokPattern

	DenormalizedPatterns map[string]*GrokPattern
	CompliedGrokRe       map[string]map[string]*GrokRegexp
}

type GrokRegexp struct {
	grokPattern *GrokPattern
	re          *regexp.Regexp
}

func (g *GrokRegexp) Run(content interface{}) (map[string]string, error) {
	if g.re == nil {
		return nil, fmt.Errorf("not complied")
	}
	result := map[string]string{}

	switch v := content.(type) {
	case []byte:
		match := g.re.FindSubmatch(v)
		if len(match) == 0 {
			return nil, fmt.Errorf("no match")
		}
		for i, name := range g.re.SubexpNames() {
			if name != "" {
				result[name] = string(match[i])
			}
		}
	case string:
		match := g.re.FindStringSubmatch(v)
		if len(match) == 0 {
			return nil, fmt.Errorf("no match")
		}
		for i, name := range g.re.SubexpNames() {
			if name != "" {
				result[name] = match[i]
			}
		}
	}
	return result, nil
}

func (g *GrokRegexp) RunWithTypeInfo(content interface{}) (map[string]interface{}, map[string]string, error) {
	castDst := map[string]interface{}{}
	castFail := map[string]string{}
	ret, err := g.Run(content)
	if err != nil {
		return nil, nil, err
	}

	for k, v := range ret {
		var err error
		var dstV interface{} = v
		if varType, ok := g.grokPattern.varType[k]; ok {
			switch varType {
			case GTypeInt:
				dstV, err = cast.ToInt64E(v)
			case GTypeFloat:
				dstV, err = cast.ToFloat64E(v)
			case GTypeBool:
				dstV, err = cast.ToBoolE(v)
			case GTypeString:
			default:
				err = fmt.Errorf("unsupported data type: %s", varType)
			}
		}
		// cast 操作失败赋予默认值
		castDst[k] = dstV
		if err != nil {
			castFail[k] = v
		}
	}
	return castDst, castFail, nil
}
