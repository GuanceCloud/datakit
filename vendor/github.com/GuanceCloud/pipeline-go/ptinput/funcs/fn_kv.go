// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

var (
	_defaultFieldSplitPattern = regexp.MustCompile(" ")
	_defaultValueSplitPattern = regexp.MustCompile("=")
)

var _regexpCache = reCache{
	m: map[string]*regexp.Regexp{},
}

type reCache struct {
	m map[string]*regexp.Regexp
	sync.RWMutex
}

func (c *reCache) set(p string) error {
	c.Lock()
	defer c.Unlock()

	if c.m == nil {
		return fmt.Errorf("nil ptr")
	}

	if r, err := regexp.Compile(p); err != nil {
		return err
	} else {
		c.m[p] = r
	}

	return nil
}

func (c *reCache) get(p string) (*regexp.Regexp, bool) {
	c.RLock()
	defer c.RUnlock()
	if c.m != nil {
		v, ok := c.m[p]
		return v, ok
	}

	return nil, false
}

func KVSplitChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := normalizeFuncArgsDeprecated(funcExpr, []string{
		"key", "field_split_pattern", "value_split_pattern",
		"trim_key", "trim_value", "include_keys", "prefix",
	}, 1); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	// field_split_pattern
	if funcExpr.Param[1] != nil {
		switch funcExpr.Param[1].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
			p := funcExpr.Param[1].StringLiteral().Val
			if err := _regexpCache.set(p); err != nil {
				return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[1].StartPos())
			}
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param field_split_pattern expect StringLiteral, got %s",
				funcExpr.Param[0].NodeType), funcExpr.NamePos)
		}
	}

	// value_split_pattern
	if funcExpr.Param[2] != nil {
		switch funcExpr.Param[2].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
			p := funcExpr.Param[2].StringLiteral().Val
			if err := _regexpCache.set(p); err != nil {
				return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[2].StartPos())
			}
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param value_split_pattern expect StringLiteral, got %s",
				funcExpr.Param[2].NodeType), funcExpr.NamePos)
		}
	}

	if funcExpr.Param[3] != nil {
		switch funcExpr.Param[3].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param trim_key expect StringLiteral, got %s",
				funcExpr.Param[3].NodeType), funcExpr.NamePos)
		}
	}

	if funcExpr.Param[4] != nil {
		switch funcExpr.Param[4].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param trim_value expect StringLiteral, got %s",
				funcExpr.Param[4].NodeType), funcExpr.NamePos)
		}
	}

	if funcExpr.Param[5] != nil {
		switch funcExpr.Param[5].NodeType { //nolint:exhaustive
		case ast.TypeListLiteral, ast.TypeIdentifier:
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param include_keys expect ListInitExpr or Identifier, got %s",
				funcExpr.Param[5].NodeType), funcExpr.NamePos)
		}
	}

	if funcExpr.Param[6] != nil {
		switch funcExpr.Param[6].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param prefix expect StringLiteral, got %s",
				funcExpr.Param[6].NodeType), funcExpr.NamePos)
		}
	}
	return nil
}

func KVSplit(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	val, err := ctx.GetKeyConv2Str(key)
	if err != nil {
		ctx.Regs.ReturnAppend(false, ast.Bool)
		return nil
	}

	var fieldSplit, valueSplit *regexp.Regexp

	// field_split_pattern
	if funcExpr.Param[1] != nil {
		switch funcExpr.Param[1].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
			p := funcExpr.Param[1].StringLiteral().Val
			var ok bool

			fieldSplit, ok = _regexpCache.get(p)
			if !ok {
				l.Debugf("field split pattern %s not found", p)
				ctx.Regs.ReturnAppend(false, ast.Bool)
				return nil
			}
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param field_split_pattern expect StringLiteral, got %s",
				funcExpr.Param[0].NodeType), funcExpr.NamePos)
		}
	}

	if funcExpr.Param[2] != nil {
		switch funcExpr.Param[2].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
			p := funcExpr.Param[2].StringLiteral().Val

			var ok bool

			valueSplit, ok = _regexpCache.get(p)
			if !ok {
				l.Debugf("value split pattern %s not found", p)
				ctx.Regs.ReturnAppend(false, ast.Bool)
				return nil
			}
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param value_split_pattern expect StringLiteral, got %s",
				funcExpr.Param[2].NodeType), funcExpr.NamePos)
		}
	}

	var trimKey, trimValue string
	if funcExpr.Param[3] != nil {
		switch funcExpr.Param[3].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
			trimKey = funcExpr.Param[3].StringLiteral().Val
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param trim_key expect StringLiteral, got %s",
				funcExpr.Param[3].NodeType), funcExpr.NamePos)
		}
	}

	if funcExpr.Param[4] != nil {
		switch funcExpr.Param[4].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
			trimValue = funcExpr.Param[4].StringLiteral().Val
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param trim_value expect StringLiteral, got %s",
				funcExpr.Param[4].NodeType), funcExpr.NamePos)
		}
	}

	var includeKeys []string
	if funcExpr.Param[5] != nil {
		switch funcExpr.Param[5].NodeType { //nolint:exhaustive
		case ast.TypeListLiteral, ast.TypeIdentifier:
			v, dt, err := runtime.RunStmt(ctx, funcExpr.Param[5])
			if err != nil {
				return err
			}
			if dt != ast.List {
				break
			}
			switch v := v.(type) {
			case []any:
				for _, k := range v {
					if k, ok := k.(string); ok {
						includeKeys = append(includeKeys, k)
					}
				}
			default:
			}

		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param include_keys expect ListInitExpr or Identifier, got %s",
				funcExpr.Param[5].NodeType), funcExpr.NamePos)
		}
	}

	if len(includeKeys) == 0 {
		ctx.Regs.ReturnAppend(false, ast.Bool)
		return nil
	}

	var prefix string
	if funcExpr.Param[6] != nil {
		switch funcExpr.Param[6].NodeType { //nolint:exhaustive
		case ast.TypeStringLiteral:
			prefix = funcExpr.Param[6].StringLiteral().Val
		default:
			return runtime.NewRunError(ctx, fmt.Sprintf("param prefix expect StringLiteral, got %s",
				funcExpr.Param[6].NodeType), funcExpr.NamePos)
		}
	}

	result := kvSplit(val, includeKeys, fieldSplit, valueSplit, trimKey, trimValue, prefix)
	if len(result) == 0 {
		ctx.Regs.ReturnAppend(false, ast.Bool)
		return nil
	}

	for k, v := range result {
		_ = addKey2PtWithVal(ctx.InData(), k, v, ast.String, ptinput.KindPtDefault)
	}

	ctx.Regs.ReturnAppend(true, ast.Bool)
	return nil
}

func kvSplit(str string, includeKeys []string, fieldSplit, valueSplit *regexp.Regexp,
	trimKey, trimValue, prefix string,
) map[string]string {
	if str == "" {
		return nil
	}

	if fieldSplit == nil {
		fieldSplit = _defaultFieldSplitPattern
	}

	if valueSplit == nil {
		valueSplit = _defaultValueSplitPattern
	}

	ks := map[string]struct{}{}

	for _, v := range includeKeys {
		ks[v] = struct{}{}
	}

	result := map[string]string{}
	fields := fieldSplit.Split(str, -1)
	for _, field := range fields {
		keyValue := valueSplit.Split(field, 2)

		if len(keyValue) == 2 {
			// trim key
			if tk := strings.Trim(keyValue[0], trimKey); tk != "" {
				keyValue[0] = tk
			} else {
				continue
			}

			// !include ? continue : ;
			if len(ks) > 0 {
				if _, ok := ks[keyValue[0]]; !ok {
					continue
				}
			}

			// trim value
			if trimValue != "" {
				keyValue[1] = strings.Trim(keyValue[1], trimValue)
			}

			// prefix + key
			if prefix != "" {
				keyValue[0] = prefix + keyValue[0]
			}

			// append to result
			result[keyValue[0]] = keyValue[1]
		}
	}
	return result
}
