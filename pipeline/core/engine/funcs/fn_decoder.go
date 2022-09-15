// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"errors"
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/runtime"
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

func Decode(ng *runtime.Context, funcExpr *ast.CallExpr) runtime.PlPanic {
	var codeType string

	key, err := getKeyName(funcExpr.Param[0])
	if err != nil {
		return err
	}

	keyVal, err := ng.GetKeyConv2Str(key)
	if err != nil {
		l.Debug(err)
		return nil
	}
	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeAttrExpr, ast.TypeIdentifier:
		key, err := getKeyName(funcExpr.Param[1])
		if err != nil {
			return err
		}
		keyVal, err := ng.GetKeyConv2Str(key)
		if err != nil {
			l.Debug(err)
			return nil
		}
		codeType = keyVal
	case ast.TypeStringLiteral:
		codeType = funcExpr.Param[1].StringLiteral.Val
	default:
		return fmt.Errorf("expect AttrExpr, Identifier or StringLiteral, got %s",
			funcExpr.Param[1].NodeType)
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

	if err := ng.AddKey2PtWithVal(key, newcont, ast.String, runtime.KindPtDefault); err != nil {
		l.Debug(err)
		return nil
	}

	return nil
}

func DecodeChecking(ng *runtime.Context, funcExpr *ast.CallExpr) error {
	if len(funcExpr.Param) < 2 || len(funcExpr.Param) > 2 {
		return fmt.Errorf("func %s expected 2", funcExpr.Name)
	}

	if _, err := getKeyName(funcExpr.Param[0]); err != nil {
		return err
	}

	switch funcExpr.Param[1].NodeType { //nolint:exhaustive
	case ast.TypeAttrExpr, ast.TypeIdentifier, ast.TypeStringLiteral:
	default:
		return fmt.Errorf("expect AttrExpr, Identifier or StringLiteral, got %s",
			funcExpr.Param[1].NodeType)
	}
	return nil
}
