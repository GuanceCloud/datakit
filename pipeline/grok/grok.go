// Package grok used to parses grok patterns in Go
package grok

import (
	"fmt"
	"regexp"
	"strings"

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

	// key: pattern name ->
	// key: unique index(such as call stack depth) of the same name pattern ->
	// value: denormalized pattern and regexp obj
	CompliedGrokRe map[string]map[string]*GrokRegexp
}

// Denormalized patterns as regular expressions.

type GrokRegexp struct {
	grokPattern *GrokPattern
	re          *regexp.Regexp
}

func (g *GrokRegexp) Run(content interface{}, trimSpace bool) (map[string]string, error) {
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
				if trimSpace {
					result[name] = strings.TrimSpace(string(match[i]))
				} else {
					result[name] = string(match[i])
				}
			}
		}
	case string:
		match := g.re.FindStringSubmatch(v)
		if len(match) == 0 {
			return nil, fmt.Errorf("no match")
		}
		for i, name := range g.re.SubexpNames() {
			if name != "" {
				if trimSpace {
					result[name] = strings.TrimSpace(match[i])
				} else {
					result[name] = match[i]
				}
			}
		}
	}
	return result, nil
}

func (g *GrokRegexp) RunWithTypeInfo(content interface{}, trimSpace bool) (map[string]interface{}, map[string]string, error) {
	castDst := map[string]interface{}{}
	castFail := map[string]string{}
	ret, err := g.Run(content, trimSpace)
	if err != nil {
		return nil, nil, err
	}
	var dstV interface{}
	for k, v := range ret {
		var err error
		dstV = v
		if varType, ok := g.grokPattern.varbType[k]; ok {
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
		// TODO: use the default value of the data type
		// cast 操作失败赋予默认值
		castDst[k] = dstV
		if err != nil {
			castFail[k] = v
		}
	}
	return castDst, castFail, nil
}
