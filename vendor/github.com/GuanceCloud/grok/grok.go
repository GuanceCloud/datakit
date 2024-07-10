// Package grok used to parses grok patterns in Go
package grok

import (
	"errors"
	"regexp"
	"strings"

	"github.com/spf13/cast"
)

var (
	valid    = regexp.MustCompile(`^\w+([-.]\w+)*(:([-.\w]+)(:(string|str|float|int|bool))?)?$`)
	normal   = regexp.MustCompile(`%{([\w-.]+(?::[\w-.]+(?::[\w-.]+)?)?)}`)
	symbolic = regexp.MustCompile(`\W`)
)

// Denormalized patterns as regular expressions.

type SubMatchName struct {
	name []string

	subexpIndex []int
	subexpCount int
}

type GrokRegexp struct {
	grokPattern   *GrokPattern
	re            *regexp.Regexp
	subMatchNames SubMatchName
}

var ErrNotCompiled = errors.New("not compiled")
var ErrMismatch = errors.New("mismatch")

func (g *GrokRegexp) GetValByName(k string, val []string) (string, bool) {
	if len(val) != len(g.subMatchNames.name) {
		return "", false
	}
	for i, name := range g.subMatchNames.name {
		if name == k {
			return val[i], true
		}
	}
	return "", false
}

func (g *GrokRegexp) MatchNames() []string {
	return g.subMatchNames.name
}

func (g *GrokRegexp) GetValAnyByName(k string, val []any) (any, bool) {
	if len(val) != len(g.subMatchNames.name) {
		return "", false
	}
	for i, name := range g.subMatchNames.name {
		if name == k {
			return val[i], true
		}
	}
	return "", false
}

func (g *GrokRegexp) GetValCastByName(k string, val []string) (any, bool) {
	if len(val) != len(g.subMatchNames.name) {
		return nil, false
	}

	for i, name := range g.subMatchNames.name {
		if name == k {
			if varType, ok := g.grokPattern.varbType[name]; ok {
				var dstV any
				switch varType {
				case GTypeInt:
					dstV, _ = cast.ToInt64E(val[i])
				case GTypeFloat:
					dstV, _ = cast.ToFloat64E(val[i])
				case GTypeBool:
					dstV, _ = cast.ToBoolE(val[i])
				case GTypeStr:
					dstV = val[i]
				default:
					return nil, false
				}
				return dstV, true
			} else {
				return val[i], true
			}
		}
	}
	return nil, false
}

func (g *GrokRegexp) Run(content string, trimSpace bool) ([]string, error) {
	if g.re == nil {
		return nil, ErrNotCompiled
	}

	match := g.re.FindStringSubmatchIndex(content)
	if len(match) == 0 {
		return nil, ErrMismatch
	}
	if g.subMatchNames.subexpCount*2 != len(match) {
		return nil, ErrMismatch
	}

	result := make([]string, len(g.subMatchNames.name))

	for i := range g.subMatchNames.name {
		idx := g.subMatchNames.subexpIndex[i]

		left := match[2*idx]
		right := match[2*idx+1]
		if left == -1 || right == -1 {
			continue
		}

		if trimSpace {
			result[i] = strings.TrimSpace(content[left:right])
		} else {
			result[i] = content[left:right]
		}
	}

	return result, nil
}

func (g *GrokRegexp) WithTypeInfo() bool {
	return len(g.grokPattern.varbType) > 0
}

func (g *GrokRegexp) RunWithTypeInfo(content string, trimSpace bool) ([]any, error) {
	ret, err := g.Run(content, trimSpace)
	if err != nil {
		return nil, err
	}

	castDst := make([]any, len(g.subMatchNames.name))

	for i, name := range g.subMatchNames.name {
		castDst[i], _ = g.GetValCastByName(name, ret)
	}

	return castDst, nil
}
