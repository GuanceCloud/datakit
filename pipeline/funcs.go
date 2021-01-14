package pipeline

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/tidwall/gjson"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

type PipelineFunc func(p *Pipeline, node parser.Node) (*Pipeline, error)

var (
	funcsMap = map[string]PipelineFunc{
		"grok":          Grok,
		"json":          Json,
		"rename":        Rename,
		"strfmt":        Strfmt,
		"cast":          Cast,
		"expr":          Expr,
		//"user_agent":    UserAgent,
		//"url_decode":    UrlDecode,
		//"geoip":         GeoIp,
		//"datetime":      DateTime,
		//"group_between": Group,
		//"group_in":      GroupIn,

		"uppercase":        Uppercase,
		"lowercase":        Lowercase,
		"drop_key":         Dropkey,
		"add_key":          Addkey,
		//"nullif":           NullIf,
		//"default_time":     DefaultTime,
		"drop_origin_data": DropOriginData,

		"add_pattern": AddPattern,
	}
)

func Json(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) < 2 || len(funcExpr.Param) > 3 {
		return p, fmt.Errorf("func %s expected 2 or 3 args", funcExpr.Name)
	}

	var key, old parser.Node

	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		key = v
	default:
		return p, fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		old = v
	default:
		return p, fmt.Errorf("expect AttrExpr or Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	newkey := old
	if len(funcExpr.Param) == 3 {
		switch v := funcExpr.Param[2].(type) {
		case *parser.AttrExpr, *parser.Identifier:
			newkey = v
		default:
			return p, fmt.Errorf("expect AttrExpr or Identifier, got %s",
				reflect.TypeOf(funcExpr.Param[2]).String())
		}
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	v, err := GsonGet(cont, old)
	if err != nil {
		return p, err
	}
	p.setContent(newkey, v)

	return p, nil
}

func Rename(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	var old, new parser.Node

	switch v := funcExpr.Param[0].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		new = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.AttrExpr, *parser.Identifier:
		old = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	v, err := p.getContent(old)
	if err != nil {
		return p, err
	}

	err = p.setContent(new, v)
	if err != nil {
		return p, err
	}

	delete(p.Output, old.String())

	return p, nil
}

func Expr(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func `%s' expected 2 args", funcExpr.Name)
	}

	var key parser.Node
	var expr *parser.BinaryExpr

	switch v := funcExpr.Param[0].(type) {
	case *parser.BinaryExpr:
		expr = v
	default:
		return p, fmt.Errorf("expect BinaryExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	if v, err := Calc(expr, p); err != nil {
		l.Warn(err)
		return p, nil
	} else {
		p.setContent(key, v)
	}

	return p, nil
}

func Strfmt(p *Pipeline, node parser.Node) (*Pipeline, error) {
	outdata := make([]interface{}, 0)

	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) < 2 {
		return p, fmt.Errorf("func `%s' expected more than 2 args", funcExpr.Name)
	}

	var key parser.Node
	var fmts string
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		fmts = v.Val
	default:
		return p, fmt.Errorf("expect StringLiteral, got `%s'",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	for i := 2; i < len(funcExpr.Param); i++ {
		switch v := funcExpr.Param[i].(type) {
		case *parser.Identifier:
			data, _ := p.getContent(v)
			outdata = append(outdata, data)
		case *parser.AttrExpr:
			data, _ := p.getContent(v)
			outdata = append(outdata, data)
		case *parser.NumberLiteral:
			if v.IsInt {
				outdata = append(outdata, v.Int)
			} else {
				outdata = append(outdata, v.Float)
			}
		case *parser.StringLiteral:
			outdata = append(outdata, v.Val)
		default:
			outdata = append(outdata, v)
		}
	}

	strfmt := fmt.Sprintf(fmts, outdata...)
	p.setContent(key, strfmt)

	return p, nil
}

func Cast(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func `%s' expected 2 args", funcExpr.Name)
	}

	var key parser.Node
	var castType string
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got `%s'",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		castType = v.Val
	default:
		return p, fmt.Errorf("expect StringLiteral, got `%s'",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	cont, err := p.getContent(key)
	if err != nil {
		l.Warn(err)
		return p, nil
	}
	val := cast(cont, castType)
	p.setContent(key, val)

	return p, nil
}

func Uppercase(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 1 {
		return p, fmt.Errorf("func %s expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	v := strings.ToUpper(cont)
	p.setContent(key, v)

	return p, nil
}

func Lowercase(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 1 {
		return p, fmt.Errorf("func %s expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	cont, err := p.getContentStr(key)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	v := strings.ToLower(cont)
	p.setContent(key, v)

	return p, nil
}
//
//func NullIf(p *Pipeline, node parser.Node) (*Pipeline, error) {
//	funcExpr := node.(*parser.FuncExpr)
//	if len(funcExpr.Param) != 2 {
//		return p, fmt.Errorf("func %s expected 2 args", funcExpr.Name)
//	}
//
//	var key string
//	switch v := funcExpr.Param[0].(type) {
//	case *parser.Identifier:
//		key = v.Name
//	default:
//		return p, fmt.Errorf("expect Identifier, got %s",
//			reflect.TypeOf(funcExpr.Param[0]).String())
//	}
//
//	if _, ok := p.getContentStrByCheck(key); !ok {
//		l.Warnf("key %v not exist", key)
//		return p, nil
//	}
//
//	var val interface{}
//	switch v := funcExpr.Param[1].(type) {
//	case *parser.StringLiteral:
//		val = v.Val
//
//	case *parser.NumberLiteral:
//		if v.IsInt {
//			val = v.Int
//		} else {
//			val = v.Float
//		}
//
//	case *parser.BoolLiteral:
//		val = v.Val
//
//	case *parser.NilLiteral:
//		val = nil
//	}
//
//	if reflect.DeepEqual(p.getContent(key), val) {
//		delete(p.Output, key)
//	}
//
//	return p, nil
//}

func Dropkey(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 1 {
		return p, fmt.Errorf("func %s expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	delete(p.Output, key.String())

	return p, nil
}

func DropOriginData(p *Pipeline, node parser.Node) (*Pipeline, error) {
	delete(p.Output, "message")
	return p, nil
}

func Addkey(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func %s expected 1 args", funcExpr.Name)
	}

	var key parser.Node
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	var val interface{}
	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		val = v.Val

	case *parser.NumberLiteral:
		if v.IsInt {
			val = v.Int
		} else {
			val = v.Float
		}

	case *parser.BoolLiteral:
		val = v.Val

	case *parser.NilLiteral:
		val = nil
	}

	p.setContent(key, val)

	return p, nil
}

func getGjsonResult(data, id string) interface{} {
	g := gjson.Get(data, id)
	switch g.Type {
	case gjson.Null:
		return nil

	case gjson.False:
		return false

	case gjson.Number:
		if strings.Contains(g.Raw, ".") {
			return g.Float()
		} else {
			return g.Int()
		}

	case gjson.String:
		return g.String()

	case gjson.True:
		return true

	case gjson.JSON:
		return g.Raw

	default:
		return nil
	}
}
