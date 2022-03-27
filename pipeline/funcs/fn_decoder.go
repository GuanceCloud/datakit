package funcs

import (
	"errors"
	"fmt"
	"reflect"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
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

func Decode(ng *parser.EngineData, node parser.Node) interface{} {
	funcExpr := fexpr(node)

	var text, codeType parser.Node

	switch v := funcExpr.Param[0].(type) {
	case *parser.StringLiteral, *parser.Identifier, *parser.AttrExpr:
		text = v
	default:
		return fmt.Errorf("expect StringLiteral or Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
		codeType = v
	default:
		return fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	cont, err := ng.GetContentStr(text)
	if err != nil {
		l.Debug(err)
		return nil
	}

	codeTypeMode := codeType.String()
	fmt.Println(codeTypeMode)

	encode, err := NewDecoder(codeTypeMode)
	if err != nil {
		l.Warn(err)
		return nil
	}

	newcont, err := encode.decoder.String(cont)
	if err != nil {
		l.Warn(err)
		return nil
	}

	if err := ng.SetContent("changed", newcont); err != nil {
		l.Warn(err)
		return nil
	}

	return nil
}

func DecodeChecking(ng *parser.EngineData, node parser.Node) error {
	funcExpr := fexpr(node)
	if len(funcExpr.Param) < 2 || len(funcExpr.Param) > 2 {
		return fmt.Errorf("func %s expected 2", funcExpr.Name)
	}

	switch funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
	default:
		return fmt.Errorf("expect AttrExpr , Identifier or StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch funcExpr.Param[1].(type) {
	case *parser.AttrExpr, *parser.Identifier, *parser.StringLiteral:
	default:
		return fmt.Errorf("expect AttrExpr , Identifier or StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}
	return nil
}
