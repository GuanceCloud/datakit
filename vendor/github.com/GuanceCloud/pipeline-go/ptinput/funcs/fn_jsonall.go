// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/tidwall/gjson"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

type compiledJSONAllCall struct {
	srcKey           string
	includeKeys      map[string]struct{}
	includeStatic    bool
	includeProvided  bool
	keyPatterns      []*regexp.Regexp
	patternsStatic   bool
	patternsProvided bool
}

func JSONAllChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if err := normalizeFuncArgsDeprecated(funcExpr, []string{
		"input", "include_keys", "key_patterns",
	}, 1); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.NamePos)
	}

	srcKey, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	includeKeys, includeStatic, includeProvided, err := staticJSONAllIncludeKeys(funcExpr.Param[1])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[1].StartPos())
	}

	keyPatterns, patternsStatic, patternsProvided, err := staticJSONAllKeyPatterns(funcExpr.Param[2])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[2].StartPos())
	}

	funcExpr.PrivateData = &compiledJSONAllCall{
		srcKey:           srcKey,
		includeKeys:      includeKeys,
		includeStatic:    includeStatic,
		includeProvided:  includeProvided,
		keyPatterns:      keyPatterns,
		patternsStatic:   patternsStatic,
		patternsProvided: patternsProvided,
	}

	return nil
}

func JSONAll(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	opts := getCompiledJSONAllCall(funcExpr)
	if opts == nil {
		if len(funcExpr.Param) == 0 || funcExpr.Param[0] == nil {
			return runtime.NewRunError(ctx, "parameter input is required", funcExpr.NamePos)
		}
		srcKey, err := getKeyName(funcExpr.Param[0])
		if err != nil {
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
		}
		opts = &compiledJSONAllCall{srcKey: srcKey}
		if len(funcExpr.Param) > 1 && funcExpr.Param[1] != nil {
			includeKeys, includeStatic, includeProvided, err := staticJSONAllIncludeKeys(funcExpr.Param[1])
			if err != nil {
				return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[1].StartPos())
			}
			opts.includeKeys = includeKeys
			opts.includeStatic = includeStatic
			opts.includeProvided = includeProvided
		}
		if len(funcExpr.Param) > 2 && funcExpr.Param[2] != nil {
			keyPatterns, patternsStatic, patternsProvided, err := staticJSONAllKeyPatterns(funcExpr.Param[2])
			if err != nil {
				return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[2].StartPos())
			}
			opts.keyPatterns = keyPatterns
			opts.patternsStatic = patternsStatic
			opts.patternsProvided = patternsProvided
		}
	}

	includeKeys := opts.includeKeys
	includeProvided := opts.includeProvided
	if !opts.includeStatic && len(funcExpr.Param) > 1 && funcExpr.Param[1] != nil {
		var err *errchain.PlError
		includeKeys, includeProvided, err = runtimeJSONAllIncludeKeys(ctx, funcExpr.Param[1])
		if err != nil {
			return err
		}
	}

	keyPatterns := opts.keyPatterns
	patternsProvided := opts.patternsProvided
	if !opts.patternsStatic && len(funcExpr.Param) > 2 && funcExpr.Param[2] != nil {
		var err *errchain.PlError
		keyPatterns, patternsProvided, err = runtimeJSONAllKeyPatterns(ctx, funcExpr.Param[2])
		if err != nil {
			return err
		}
	}

	filterInclude := includeProvided || patternsProvided
	if !filterInclude {
		return nil
	}

	cont, err := ctx.GetKeyConv2Str(opts.srcKey)
	if err != nil {
		l.Debug(err)
		return nil
	}
	if !gjson.Valid(cont) {
		l.Debug("invalid json")
		return nil
	}

	state := &jsonAllWalkState{
		ctx:           ctx,
		includeKeys:   includeKeys,
		keyPatterns:   keyPatterns,
		filterInclude: filterInclude,
	}
	state.walkGJSON(gjson.Parse(cont))

	return nil
}

func getCompiledJSONAllCall(funcExpr *ast.CallExpr) *compiledJSONAllCall {
	if funcExpr == nil {
		return nil
	}
	if compiled, ok := funcExpr.PrivateData.(*compiledJSONAllCall); ok {
		return compiled
	}
	return nil
}

func staticJSONAllIncludeKeys(node *ast.Node) (map[string]struct{}, bool, bool, error) {
	keys, static, provided, err := staticJSONAllStringList(node, "include_keys")
	if err != nil {
		return nil, static, provided, err
	}
	return jsonAllStringSet(keys), static, provided, nil
}

func staticJSONAllKeyPatterns(node *ast.Node) ([]*regexp.Regexp, bool, bool, error) {
	patterns, static, provided, err := staticJSONAllStringList(node, "key_patterns")
	if err != nil {
		return nil, static, provided, err
	}
	if !static {
		return nil, static, provided, nil
	}
	compiled, err := compileJSONAllKeyPatterns(patterns)
	return compiled, static, provided, err
}

func staticJSONAllStringList(node *ast.Node, paramName string) ([]string, bool, bool, error) {
	if node == nil {
		return nil, true, false, nil
	}
	if node.NodeType == ast.TypeNilLiteral {
		return nil, true, false, nil
	}

	if node.NodeType != ast.TypeListLiteral {
		switch node.NodeType { //nolint:exhaustive
		case ast.TypeIdentifier, ast.TypeAttrExpr, ast.TypeCallExpr:
			return nil, false, true, nil
		default:
			return nil, false, true, fmt.Errorf("param %s expect ListLiteral, got %s", paramName, node.NodeType)
		}
	}

	keys := make([]string, 0, len(node.ListLiteral().List))
	for _, elem := range node.ListLiteral().List {
		if elem.NodeType != ast.TypeStringLiteral {
			return nil, true, true, fmt.Errorf("param %s element expect StringLiteral, got %s",
				paramName, elem.NodeType)
		}
		keys = append(keys, elem.StringLiteral().Val)
	}
	return keys, true, true, nil
}

func runtimeJSONAllIncludeKeys(ctx *runtime.Task, node *ast.Node) (map[string]struct{}, bool, *errchain.PlError) {
	keys, provided, err := runtimeJSONAllStringList(ctx, node, "include_keys")
	if err != nil {
		return nil, provided, err
	}
	return jsonAllStringSet(keys), provided, nil
}

func runtimeJSONAllKeyPatterns(ctx *runtime.Task, node *ast.Node) ([]*regexp.Regexp, bool, *errchain.PlError) {
	patterns, provided, err := runtimeJSONAllStringList(ctx, node, "key_patterns")
	if err != nil {
		return nil, provided, err
	}
	if !provided {
		return nil, false, nil
	}

	compiled, errCompile := compileJSONAllKeyPatterns(patterns)
	if errCompile != nil {
		return nil, provided, runtime.NewRunError(ctx, errCompile.Error(), node.StartPos())
	}
	return compiled, true, nil
}

func runtimeJSONAllStringList(ctx *runtime.Task, node *ast.Node, paramName string) ([]string, bool, *errchain.PlError) {
	val, dtype, err := runtime.RunStmt(ctx, node)
	if err != nil {
		return nil, false, err
	}
	if dtype == ast.Nil {
		return nil, false, nil
	}
	if dtype != ast.List {
		return nil, false, runtime.NewRunError(ctx, fmt.Sprintf("param %s expect list", paramName), node.StartPos())
	}

	keys, ok := val.([]any)
	if !ok {
		return nil, false, runtime.NewRunError(ctx, fmt.Sprintf("param %s expect list", paramName), node.StartPos())
	}

	strKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		keyStr, ok := key.(string)
		if !ok {
			return nil, true, runtime.NewRunError(ctx, fmt.Sprintf("param %s element expect string", paramName),
				node.StartPos())
		}
		strKeys = append(strKeys, keyStr)
	}
	return strKeys, true, nil
}

func jsonAllStringSet(keys []string) map[string]struct{} {
	if keys == nil {
		return nil
	}
	set := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		set[key] = struct{}{}
	}
	return set
}

func compileJSONAllKeyPatterns(patterns []string) ([]*regexp.Regexp, error) {
	if patterns == nil {
		return nil, nil
	}

	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		re, err := regexp.Compile("^" + ptKvsWildcardToRegexp(pattern) + "$")
		if err != nil {
			return nil, err
		}
		compiled = append(compiled, re)
	}
	return compiled, nil
}

type jsonAllWalkState struct {
	ctx           *runtime.Task
	includeKeys   map[string]struct{}
	keyPatterns   []*regexp.Regexp
	filterInclude bool
}

func (s *jsonAllWalkState) walkGJSON(root gjson.Result) {
	if root.Type != gjson.JSON {
		return
	}

	rootIsArray := root.IsArray()
	rootIsObject := root.IsObject()
	if !rootIsArray && !rootIsObject {
		return
	}

	root.ForEach(func(key, value gjson.Result) bool {
		path := key.String()
		if rootIsArray {
			path = jsonAllArrayIndex(int(key.Int()))
		}
		return s.add(path, value)
	})
}

func (s *jsonAllWalkState) add(path string, value gjson.Result) bool {
	if path == "" {
		return true
	}

	if s.filterInclude {
		if !s.match(path) {
			return true
		}
	}

	v, dtype, ok := jsonAllValue(value)
	if !ok {
		return true
	}

	addKey2PtWithVal(s.ctx.InData(), path, v, dtype, ptinput.KindPtDefault)
	return true
}

func (s *jsonAllWalkState) match(path string) bool {
	if _, ok := s.includeKeys[path]; ok {
		return true
	}
	for _, pattern := range s.keyPatterns {
		if pattern.MatchString(path) {
			return true
		}
	}
	return false
}

func jsonAllValue(value gjson.Result) (any, ast.DType, bool) {
	switch value.Type {
	case gjson.True, gjson.False:
		return value.Bool(), ast.Bool, true
	case gjson.Number:
		return value.Float(), ast.Float, true
	case gjson.String:
		return value.String(), ast.String, true
	default:
		return nil, ast.Invalid, false
	}
}

func jsonAllArrayIndex(idx int) string {
	return "[" + strconv.Itoa(idx) + "]"
}
