package funcs

import (
	_ "embed"

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

	// todo: parse function definition
	_ = "fn pt_kvs_get(name: str) -> any"
	_ = "fn pt_kvs_set(name: str, value: any, as_tag: bool = false) -> bool"
	_ = "fn pt_kvs_del(name: str)"
	_ = "fn pt_kvs_keys(tags: bool = true, fields: bool = true) -> list"

	FnPtKvsGet = NewFunc(
		"pt_kvs_get",
		[]*Param{
			{
				Name: "name",
				Type: []ast.DType{ast.String},
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

func ptKvsGet(ctx *runtime.Task, funcExpr *ast.CallExpr, vals ...any) *errchain.PlError {
	if val, dtype, err := getPtKey(ctx.InData(), vals[0].(string)); err != nil {
		ctx.Regs.ReturnAppend(nil, ast.Nil)
	} else {
		ctx.Regs.ReturnAppend(val, dtype)
	}

	return nil
}

func ptKvsSet(ctx *runtime.Task, funcExpr *ast.CallExpr, vals ...any) *errchain.PlError {
	name := vals[0].(string)
	asTag := vals[2].(bool)
	val := vals[1]

	pt, err := getPoint(ctx.InData())
	if err != nil {
		ctx.Regs.ReturnAppend(false, ast.Bool)
		return nil
	}

	if asTag {
		if ok := pt.SetTag(name, val, getValDtype(val)); !ok {
			ctx.Regs.ReturnAppend(false, ast.Bool)
			return nil
		}
	} else {
		if ok := pt.Set(name, val, getValDtype(val)); !ok {
			ctx.Regs.ReturnAppend(false, ast.Bool)
			return nil
		}
	}

	ctx.Regs.ReturnAppend(true, ast.Bool)
	return nil
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

	var elemCount int

	if tags {
		elemCount += len(pt.Tags())
	}
	if fields {
		elemCount += len(pt.Fields())
	}

	keyList := make([]any, 0, elemCount)

	if tags {
		for k := range pt.Tags() {
			keyList = append(keyList, k)
		}
	}
	if fields {
		for k := range pt.Fields() {
			keyList = append(keyList, k)
		}
	}

	ctx.Regs.ReturnAppend(keyList, ast.List)

	return nil
}
