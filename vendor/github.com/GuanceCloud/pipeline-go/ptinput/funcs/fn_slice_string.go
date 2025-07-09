package funcs

import (
	_ "embed"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

// embed docs.
var (
	//go:embed md/slice_string.md
	docSliceString string

	//go:embed md/slice_string.en.md
	docSliceStringEN string

	// todo: parse function definition
	_ = "fn slice_string(name: str, start: int, end: int) -> str"

	FnSliceString = NewFunc(
		"slice_string",
		[]*Param{
			{
				Name: "name",
				Type: []ast.DType{ast.String},
			},
			{
				Name: "start",
				Type: []ast.DType{ast.Int},
			},
			{
				Name: "end",
				Type: []ast.DType{ast.Int},
			},
		},
		[]ast.DType{ast.String},
		[2]*PLDoc{
			{
				Language: langTagZhCN, Doc: docSliceString,
				FnCategory: map[string][]string{
					langTagZhCN: {cStringOp}},
			},
			{
				Language: langTagEnUS, Doc: docSliceStringEN,
				FnCategory: map[string][]string{
					langTagEnUS: {eStringOp}},
			},
		},
		sliceString,
	)
)

func sliceString(ctx *runtime.Task, funcExpr *ast.CallExpr, vals ...any) *errchain.PlError {
	errstring := ""
	if len(vals) != 3 {
		ctx.Regs.ReturnAppend(errstring, ast.String)
		return nil
	}
	name := vals[0].(string)
	start, ok := vals[1].(int64)
	if !ok {
		ctx.Regs.ReturnAppend(errstring, ast.String)
		return nil
	}
	end, ok := vals[2].(int64)
	if !ok {
		ctx.Regs.ReturnAppend(errstring, ast.String)
		return nil
	}

	runes := []rune(name)
	lenRunes := int64(len(runes))

	if end > lenRunes {
		end = lenRunes
	}

	if start < 0 || end > lenRunes || start > end {
		ctx.Regs.ReturnAppend(errstring, ast.String)
		return nil
	}

	substring := string(runes[start:end])

	ctx.Regs.ReturnAppend(substring, ast.String)
	return nil
}
