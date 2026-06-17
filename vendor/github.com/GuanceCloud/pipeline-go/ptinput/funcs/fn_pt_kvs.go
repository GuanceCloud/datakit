package funcs

import (
	_ "embed"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

// embed docs.
var (
	//go:embed md/pt_kvs_get.md
	docPtKvsGet string

	//go:embed md/pt_kvs_get.en.md
	docPtKvsGetEN string

	//go:embed md/pt_kvs_set.md
	docKvsSet string

	//go:embed md/pt_kvs_set.en.md
	docPtKvsSetEN string

	//go:embed md/pt_kvs_del.md
	docKvsDel string

	//go:embed md/pt_kvs_del.en.md
	docPtKvsDelEN string

	//go:embed md/pt_kvs_keys.md
	docKvsKeys string

	//go:embed md/pt_kvs_keys.en.md
	docPtKvsKeysEN string

	//go:embed md/pt_kvs_set_map.md
	docPtKvsSetMap string

	//go:embed md/pt_kvs_set_map.en.md
	docPtKvsSetMapEN string

	// todo: parse function definition
	_ = "fn pt_kvs_get(name: str, raw: bool = false) -> any"
	_ = "fn pt_kvs_set(name: str, value: any, as_tag: bool = false, raw: bool = false) -> bool"
	_ = "fn pt_kvs_set_map(values: map, include_keys: list|nil = nil, key_patterns: list|nil = nil, as_tag: bool = false, raw: bool = false) -> int"
	_ = "fn pt_kvs_del(name: str)"
	_ = "fn pt_kvs_keys(tags: bool = true, fields: bool = true) -> list"

	FnPtKvsGet = NewFunc(
		"pt_kvs_get",
		[]*Param{
			{
				Name: "name",
				Type: []ast.DType{ast.String},
			},
			{
				Name:     "raw",
				Type:     []ast.DType{ast.Bool},
				Optional: true,
				DefaultVal: func() (any, ast.DType) {
					return false, ast.Bool
				},
			},
		},
		[]ast.DType{ast.Bool, ast.Int, ast.Float, ast.String,
			ast.List, ast.Map, ast.Nil},
		[2]*PLDoc{
			{
				Language: langTagZhCN, Doc: docPtKvsGet,
				FnCategory: map[string][]string{
					langTagZhCN: {cPointOp}},
			},
			{
				Language: langTagEnUS, Doc: docPtKvsGetEN,
				FnCategory: map[string][]string{
					langTagEnUS: {ePointOp}},
			},
		},
		ptKvsGet,
	)

	FnPtKvsSet = NewFunc(
		"pt_kvs_set",
		[]*Param{
			{
				Name: "name",
				Type: []ast.DType{ast.String},
			},
			{
				Name: "value",
				Type: []ast.DType{ast.Bool, ast.Int, ast.Float, ast.String,
					ast.List, ast.Map, ast.Nil},
			},
			{
				Name:     "as_tag",
				Type:     []ast.DType{ast.Bool},
				Optional: true,
				DefaultVal: func() (any, ast.DType) {
					return false, ast.Bool
				},
			},
			{
				Name:     "raw",
				Type:     []ast.DType{ast.Bool},
				Optional: true,
				DefaultVal: func() (any, ast.DType) {
					return false, ast.Bool
				},
			},
		},
		[]ast.DType{ast.Bool},
		[2]*PLDoc{
			{
				Language: langTagZhCN, Doc: docKvsSet,
				FnCategory: map[string][]string{
					langTagZhCN: {cPointOp}},
			},
			{
				Language: langTagEnUS, Doc: docPtKvsSetEN,
				FnCategory: map[string][]string{
					langTagEnUS: {ePointOp}},
			},
		},
		ptKvsSet,
	)

	FnPtKvsSetMap = newPtKvsSetMapFunc()

	FnPtKvsDel = NewFunc(
		"pt_kvs_del",
		[]*Param{
			{
				Name: "name",
				Type: []ast.DType{ast.String},
			},
		},
		nil,
		[2]*PLDoc{
			{
				Language: langTagZhCN, Doc: docKvsDel,
				FnCategory: map[string][]string{
					langTagZhCN: {cPointOp}},
			},
			{
				Language: langTagEnUS, Doc: docPtKvsDelEN,
				FnCategory: map[string][]string{
					langTagEnUS: {ePointOp}},
			},
		},
		ptKvsDel,
	)

	FnPtKvsKeys = NewFunc(
		"pt_kvs_keys",
		[]*Param{
			{
				Name:     "tags",
				Type:     []ast.DType{ast.Bool},
				Optional: true,
				DefaultVal: func() (any, ast.DType) {
					return true, ast.Bool
				},
			},
			{
				Name:     "fields",
				Type:     []ast.DType{ast.Bool},
				Optional: true,
				DefaultVal: func() (any, ast.DType) {
					return true, ast.Bool
				},
			},
		},
		[]ast.DType{ast.List},
		[2]*PLDoc{
			{
				Language: langTagZhCN, Doc: docKvsKeys,
				FnCategory: map[string][]string{
					langTagZhCN: {cPointOp}},
			},
			{
				Language: langTagEnUS, Doc: docPtKvsKeysEN,
				FnCategory: map[string][]string{
					langTagEnUS: {ePointOp}},
			},
		},
		ptKvsKeys,
	)
)

type compiledPtKvsSetMapCall struct {
	exactKeys      map[string]struct{}
	exactKeyList   []string
	filterExact    bool
	exactStatic    bool
	patterns       []*regexp.Regexp
	filterPatterns bool
	patternsStatic bool
}

func newPtKvsSetMapFunc() *Function {
	params := []*Param{
		{
			Name: "values",
			Type: []ast.DType{ast.Map},
		},
		{
			Name:     "include_keys",
			Type:     []ast.DType{ast.List, ast.Nil},
			Optional: true,
			DefaultVal: func() (any, ast.DType) {
				return nil, ast.Nil
			},
		},
		{
			Name:     "key_patterns",
			Type:     []ast.DType{ast.List, ast.Nil},
			Optional: true,
			DefaultVal: func() (any, ast.DType) {
				return nil, ast.Nil
			},
		},
		{
			Name:     "as_tag",
			Type:     []ast.DType{ast.Bool},
			Optional: true,
			DefaultVal: func() (any, ast.DType) {
				return false, ast.Bool
			},
		},
		{
			Name:     "raw",
			Type:     []ast.DType{ast.Bool},
			Optional: true,
			DefaultVal: func() (any, ast.DType) {
				return false, ast.Bool
			},
		},
	}

	fn := NewFunc(
		"pt_kvs_set_map",
		params,
		[]ast.DType{ast.Int},
		[2]*PLDoc{
			{
				Language: langTagZhCN, Doc: docPtKvsSetMap,
				FnCategory: map[string][]string{
					langTagZhCN: {cPointOp}},
			},
			{
				Language: langTagEnUS, Doc: docPtKvsSetMapEN,
				FnCategory: map[string][]string{
					langTagEnUS: {ePointOp}},
			},
		},
		ptKvsSetMap,
	)

	baseCheck := fn.Check
	fn.Check = func(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
		if err := baseCheck(ctx, funcExpr); err != nil {
			return err
		}
		return ptKvsSetMapChecking(ctx, funcExpr)
	}

	return fn
}

func ptKvsGet(ctx *runtime.Task, funcExpr *ast.CallExpr, vals ...any) *errchain.PlError {
	var (
		val   any
		dtype ast.DType
		err   error
	)

	if vals[1].(bool) {
		val, dtype, err = getPtKeyRaw(ctx.InData(), vals[0].(string))
	} else {
		val, dtype, err = getPtKey(ctx.InData(), vals[0].(string))
	}

	if err != nil {
		ctx.Regs.ReturnAppend(nil, ast.Nil)
	} else {
		ctx.Regs.ReturnAppend(val, dtype)
	}

	return nil
}

func ptKvsSet(ctx *runtime.Task, funcExpr *ast.CallExpr, vals ...any) *errchain.PlError {
	name := vals[0].(string)
	asTag := vals[2].(bool)
	raw := vals[3].(bool)
	val := vals[1]

	pt, err := getPoint(ctx.InData())
	if err != nil {
		ctx.Regs.ReturnAppend(false, ast.Bool)
		return nil
	}

	if ok := ptKvsSetValue(pt, name, val, asTag, raw); !ok {
		ctx.Regs.ReturnAppend(false, ast.Bool)
		return nil
	}

	ctx.Regs.ReturnAppend(true, ast.Bool)
	return nil
}

func ptKvsSetMapChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	includeKeys := ptKvsFuncArgValueNode(funcExpr.Param[1])
	exactKeys, exactKeyList, filterExact, exactStatic, err := ptKvsExactKeySetFromStaticNode(includeKeys)
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), includeKeys.StartPos())
	}

	keyPatterns := ptKvsFuncArgValueNode(funcExpr.Param[2])
	patterns, filterPatterns, patternsStatic, err := ptKvsPatternMatchersFromStaticNode(keyPatterns)
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), keyPatterns.StartPos())
	}

	funcExpr.PrivateData = &compiledPtKvsSetMapCall{
		exactKeys:      exactKeys,
		exactKeyList:   exactKeyList,
		filterExact:    filterExact,
		exactStatic:    exactStatic,
		patterns:       patterns,
		filterPatterns: filterPatterns,
		patternsStatic: patternsStatic,
	}
	return nil
}

func ptKvsFuncArgValueNode(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	if node.NodeType == ast.TypeAssignmentExpr {
		expr := node.AssignmentExpr()
		if len(expr.RHS) > 0 {
			return expr.RHS[0]
		}
	}
	return node
}

func ptKvsSetMap(ctx *runtime.Task, funcExpr *ast.CallExpr, vals ...any) *errchain.PlError {
	values, _ := vals[0].(map[string]any)
	includeKeys := vals[1]
	keyPatterns := vals[2]
	asTag := vals[3].(bool)
	raw := vals[4].(bool)

	compiled, _ := funcExpr.PrivateData.(*compiledPtKvsSetMapCall)

	var (
		exactKeys      map[string]struct{}
		exactKeyList   []string
		filterExact    bool
		patterns       []*regexp.Regexp
		filterPatterns bool
		err            error
	)

	if compiled != nil && compiled.exactStatic {
		exactKeys = compiled.exactKeys
		exactKeyList = compiled.exactKeyList
		filterExact = compiled.filterExact
	} else {
		exactKeys, exactKeyList, filterExact, err = ptKvsExactKeySet(includeKeys)
		if err != nil {
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[1].StartPos())
		}
	}

	if compiled != nil && compiled.patternsStatic {
		patterns = compiled.patterns
		filterPatterns = compiled.filterPatterns
	} else {
		patterns, filterPatterns, err = ptKvsPatternMatchers(keyPatterns)
		if err != nil {
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[2].StartPos())
		}
	}

	if !filterExact && !filterPatterns {
		ctx.Regs.ReturnAppend(int64(0), ast.Int)
		return nil
	}
	if len(exactKeyList) == 0 && len(patterns) == 0 {
		ctx.Regs.ReturnAppend(int64(0), ast.Int)
		return nil
	}

	pt, errPt := getPoint(ctx.InData())
	if errPt != nil {
		ctx.Regs.ReturnAppend(int64(0), ast.Int)
		return nil
	}

	var count int64
	if len(patterns) == 0 {
		for _, key := range exactKeyList {
			val, ok := values[key]
			if !ok {
				continue
			}
			if ok := ptKvsSetValue(pt, key, val, asTag, raw); ok {
				count++
			}
		}
		ctx.Regs.ReturnAppend(count, ast.Int)
		return nil
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if !ptKvsKeyMatched(key, exactKeys, patterns) {
			continue
		}
		if ok := ptKvsSetValue(pt, key, values[key], asTag, raw); ok {
			count++
		}
	}

	ctx.Regs.ReturnAppend(count, ast.Int)
	return nil
}

func ptKvsSetValue(pt ptinput.PlInputPt, name string, val any, asTag, raw bool) bool {
	if asTag {
		return pt.SetTag(name, val, getValDtype(val))
	}

	dtype := getValDtype(val)
	if !raw && (dtype == ast.List || dtype == ast.Map) {
		if s, err := ptinput.Conv2String(val, dtype); err == nil {
			val = s
			dtype = ast.String
		}
	}

	return pt.Set(name, val, dtype)
}

func ptKvsExactKeySet(v any) (map[string]struct{}, []string, bool, error) {
	keys, provided, err := ptKvsStringList(v, "include_keys")
	if err != nil || !provided {
		return nil, nil, provided, err
	}

	exactKeys := ptKvsStringSet(keys)
	return exactKeys, ptKvsSortedKeys(exactKeys), true, nil
}

func ptKvsExactKeySetFromStaticNode(node *ast.Node) (map[string]struct{}, []string, bool, bool, error) {
	if node == nil || node.NodeType == ast.TypeNilLiteral {
		return nil, nil, false, true, nil
	}
	if node.NodeType != ast.TypeListLiteral {
		return nil, nil, false, false, nil
	}

	keys, err := ptKvsStringListFromListLiteral(node, "include_keys")
	if err != nil {
		return nil, nil, true, true, err
	}

	exactKeys := ptKvsStringSet(keys)
	return exactKeys, ptKvsSortedKeys(exactKeys), true, true, nil
}

func ptKvsStringSet(keys []string) map[string]struct{} {
	set := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		set[key] = struct{}{}
	}
	return set
}

func ptKvsSortedKeys(set map[string]struct{}) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func ptKvsPatternMatchers(v any) ([]*regexp.Regexp, bool, error) {
	patterns, provided, err := ptKvsStringList(v, "key_patterns")
	if err != nil || !provided {
		return nil, provided, err
	}

	return ptKvsCompilePatterns(patterns)
}

func ptKvsPatternMatchersFromStaticNode(node *ast.Node) ([]*regexp.Regexp, bool, bool, error) {
	if node == nil || node.NodeType == ast.TypeNilLiteral {
		return nil, false, true, nil
	}
	if node.NodeType != ast.TypeListLiteral {
		return nil, false, false, nil
	}

	patterns, err := ptKvsPatternMatchersFromListLiteral(node)
	if err != nil {
		return nil, true, true, err
	}
	return patterns, true, true, nil
}

func ptKvsStringList(v any, paramName string) ([]string, bool, error) {
	if v == nil {
		return nil, false, nil
	}

	values, ok := v.([]any)
	if !ok {
		return nil, false, fmt.Errorf("param %s expect list", paramName)
	}

	keys := make([]string, 0, len(values))
	for _, key := range values {
		keyStr, ok := key.(string)
		if !ok {
			return nil, true, fmt.Errorf("param %s element expect string", paramName)
		}
		keys = append(keys, keyStr)
	}

	return keys, true, nil
}

func ptKvsStringListFromListLiteral(node *ast.Node, paramName string) ([]string, error) {
	keys := make([]string, 0, len(node.ListLiteral().List))
	for _, elem := range node.ListLiteral().List {
		if elem.NodeType != ast.TypeStringLiteral {
			return nil, fmt.Errorf("param %s element expect StringLiteral, got %s",
				paramName, elem.NodeType)
		}
		keys = append(keys, elem.StringLiteral().Val)
	}
	return keys, nil
}

func ptKvsPatternMatchersFromListLiteral(node *ast.Node) ([]*regexp.Regexp, error) {
	patterns, err := ptKvsStringListFromListLiteral(node, "key_patterns")
	if err != nil {
		return nil, err
	}
	matchers, _, err := ptKvsCompilePatterns(patterns)
	return matchers, err
}

func ptKvsCompilePatterns(patterns []string) ([]*regexp.Regexp, bool, error) {
	matchers := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		re, err := regexp.Compile("^" + ptKvsWildcardToRegexp(pattern) + "$")
		if err != nil {
			return nil, true, err
		}
		matchers = append(matchers, re)
	}
	return matchers, true, nil
}

func ptKvsKeyMatched(key string, exactKeys map[string]struct{}, patterns []*regexp.Regexp) bool {
	if _, ok := exactKeys[key]; ok {
		return true
	}
	for _, pattern := range patterns {
		if pattern.MatchString(key) {
			return true
		}
	}
	return false
}

func ptKvsWildcardToRegexp(pattern string) string {
	var b strings.Builder
	for _, r := range pattern {
		switch r {
		case '*':
			b.WriteString(".*")
		case '?':
			b.WriteByte('.')
		default:
			b.WriteString(regexp.QuoteMeta(string(r)))
		}
	}
	return b.String()
}

func ptKvsDel(ctx *runtime.Task, funcExpr *ast.CallExpr, vals ...any) *errchain.PlError {
	name := vals[0].(string)
	deletePtKey(ctx.InData(), name)
	return nil
}

func ptKvsKeys(ctx *runtime.Task, funcExpr *ast.CallExpr, vals ...any) *errchain.PlError {
	tags := vals[0].(bool)
	fields := vals[1].(bool)

	pt, err := getPoint(ctx.InData())
	if err != nil {
		ctx.Regs.ReturnAppend(false, ast.Bool)
		return nil
	}

	ctx.Regs.ReturnAppend(ptKvsKeyList(pt, tags, fields), ast.List)

	return nil
}

func ptKvsKeyList(pt ptinput.PlInputPt, tags, fields bool) []any {
	kvs := pt.Point().KVs()
	keyList := make([]any, 0, len(kvs))
	for _, kv := range kvs {
		if includePtKvsKey(kv, tags, fields) {
			keyList = append(keyList, kv.Key)
		}
	}
	return keyList
}

func includePtKvsKey(kv *point.Field, tags, fields bool) bool {
	if kv == nil {
		return false
	}
	if kv.IsTag {
		if !tags {
			return false
		}
		_, ok := kv.Val.(*point.Field_S)
		return ok
	}
	return fields
}
