// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	_ "embed"
	"encoding/hex"
	"unsafe"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

// embed docs.
var (
	//go:embed md/hash.md
	docHash string

	//go:embed md/hash.en.md
	docHashEN string

	_ = "fn hash(text: str, method: str) -> str"

	FnHash = NewFunc(
		"hash",
		[]*Param{
			{
				Name: "text",
				Type: []ast.DType{ast.String},
			},
			{
				Name: "method",
				Type: []ast.DType{ast.String},
			},
		},
		[]ast.DType{ast.String},
		[2]*PLDoc{
			{
				Language: langTagZhCN, Doc: docHash,
				FnCategory: map[string][]string{
					langTagZhCN: {cStringOp}},
			},
			{
				Language: langTagEnUS, Doc: docHashEN,
				FnCategory: map[string][]string{
					langTagEnUS: {eStringOp},
				},
			},
		},
		hashFn,
	)
)

func hashFn(ctx *runtime.Task, funcExpr *ast.CallExpr, vals ...any) *errchain.PlError {
	text := vals[0].(string)
	txt := *(*[]byte)(unsafe.Pointer(&text))
	// todo: go1.20+
	// textSlice := unsafe.Slice(unsafe.StringData(text), len(text))
	method := vals[1].(string)

	var sum string
	switch method {
	case "md5":
		b := md5.Sum(txt)
		sum = hex.EncodeToString(b[:])
	case "sha1":
		b := sha1.Sum(txt)
		sum = hex.EncodeToString(b[:])
	case "sha256":
		b := sha256.Sum256(txt)
		sum = hex.EncodeToString(b[:])
	case "sha512":
		b := sha512.Sum512(txt)
		sum = hex.EncodeToString(b[:])
	default:
	}

	ctx.Regs.ReturnAppend(sum, ast.String)
	return nil
}
