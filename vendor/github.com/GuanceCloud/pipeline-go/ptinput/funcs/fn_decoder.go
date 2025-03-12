// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"errors"
	"fmt"

	"github.com/GuanceCloud/pipeline-go/ptinput"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
)

var errUnknownCharacterEncoding = errors.New("unknown character encoding")

type Decoder struct {
	decoder *encoding.Decoder
}

func NewDecoder(enc string) (*Decoder, error) {
	var decoder *encoding.Decoder

	switch enc {
	case "'utf-8'":
		decoder = unicode.UTF8.NewDecoder()
	case "utf-16le", "'utf-16le'":
		decoder = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
	case "utf-16be", "'utf-16be'":
		decoder = unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()
	case "gbk", "'gbk'":
		decoder = simplifiedchinese.GBK.NewDecoder()
	case "gb18030", "'gb18030'":
		decoder = simplifiedchinese.GB18030.NewDecoder()
	default:
		return nil, errUnknownCharacterEncoding
	}

	return &Decoder{decoder: decoder}, nil
}

func Decode(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	var codeType string

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	keyVal, err := ctx.GetKeyConv2Str(key)
	if err != nil {
		l.Debug(err)
		return nil
	}
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeAttrExpr, ast.TypeIdentifier:
		key, err := getKeyName(funcExpr.Param[1])
		if err != nil {
			return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[1].StartPos())
		}
		keyVal, err := ctx.GetKeyConv2Str(key)
		if err != nil {
			l.Debug(err)
			return nil
		}
		codeType = keyVal
	case ast.TypeStringLiteral:
		codeType = funcExpr.Param[1].StringLiteral().Val
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"expect AttrExpr, Identifier or StringLiteral, got %s",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}

	encode, err := NewDecoder(codeType)
	if err != nil {
		l.Debug(err)
		return nil
	}

	newcont, err := encode.decoder.String(keyVal)
	if err != nil {
		l.Debug(err)
		return nil
	}

	_ = addKey2PtWithVal(ctx.InData(), key, newcont,
		ast.String, ptinput.KindPtDefault)
	return nil
}

func DecodeChecking(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
	if len(funcExpr.Param) < 2 || len(funcExpr.Param) > 2 {
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"func %s expected 2", funcExpr.Name), funcExpr.NamePos)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return runtime.NewRunError(ctx, err.Error(), funcExpr.Param[0].StartPos())
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeAttrExpr, ast.TypeIdentifier, ast.TypeStringLiteral:
	default:
		return runtime.NewRunError(ctx, fmt.Sprintf(
			"expect AttrExpr, Identifier or StringLiteral, got %s",
			funcExpr.Param[1].NodeType), funcExpr.Param[1].StartPos())
	}
	return nil
}
