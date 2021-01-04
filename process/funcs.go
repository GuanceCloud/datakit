package process

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/process/parser"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/process/grok"
)
const (
	CONTENT = "__content"
)

type ProFunc func(p *Procedure, node parser.Node) (*Procedure, error)

var (
	funcsMap = map[string]ProFunc{
		"grok"       : grok.Grok,
		"rename"     : Rename   ,
		"user_agent" : UserAgent,
		"url_decode" : UrlDecode,
		"geo_ip"     : GeoIp    ,
		"date"       : Date     ,
		"expr"       : Expr     ,
		"stringf"    : Stringf  ,
		"cast"       : Cast     ,
		"group"      : Group    ,
	}
)

func (p *Procedure) ProcessLog(data string, nodes []parser.Node) *Procedure {
	var err error

	p.Content = logStructed(data)

	for _, node := range nodes {
		switch v := node.(type) {
		case *parser.FuncExpr:
			fn := strings.ToLower(v.Name)
			f, ok := funcsMap[fn]
			if !ok {
				err := fmt.Errorf("unsupported func: %v", v.Name)
				l.Error(err)
				p.lastErr = err
				return p
			}

			p, err = f(p, node)
			if err != nil {
				l.Errorf("ProcessLog %v: %v", v.Name, err)
				p.lastErr = err
				return p
			}

		default:
			p.lastErr = fmt.Errorf("%v not function", v.String())
		}
	}
	return p
}

func Rename(p *Procedure, node parser.Node) (*Procedure, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 2 {
		return nil, fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	old := funcExpr.Param[0].(*parser.StringLiteral).Val
	new := funcExpr.Param[1].(*parser.StringLiteral).Val

	data := make(map[string]interface{})
	err := json.Unmarshal(p.Content, &data)
	if err != nil {
		return p, err
	}

	if v, ok := data[old]; ok {
		data[new] = v
	} else {
		return nil, fmt.Errorf("old tag %v not founded", old)
	}

	js, err := json.Marshal(data)
	if err != nil {
		return p, err
	}
	p.Content = js

	return p, nil
}

func UserAgent(p *Procedure, node parser.Node) (*Procedure, error) {
	return p, nil
}

func UrlDecode(p *Procedure, node parser.Node) (*Procedure, error) {
	return p, nil
}

func GeoIp(p *Procedure, node parser.Node) (*Procedure, error) {
	return p, nil
}

func Date(p *Procedure, node parser.Node) (*Procedure, error) {
	//funcExpr := node.(*parser.FuncExpr)
	//if len(funcExpr.Param) != 3 {
	//	return nil, fmt.Errorf("func %s expected 3 args", funcExpr.Name)
	//}
	//
	//tag   := funcExpr.Param[0].(*parser.StringLiteral).Val
	//fmts  := funcExpr.Param[1].(*parser.StringLiteral).Val
	//field := funcExpr.Param[2].(*parser.StringLiteral).Val
	//
	//data := make(map[string]interface{})
	//err := json.Unmarshal(p.Content, &data)
	//if err != nil {
	//	return p, err
	//}
	//
	//if v, ok := data[field]; ok {
	//	data[tag] = time.
	//} else {
	//	return nil, fmt.Errorf("%v not founded", field)
	//}
	//
	//js, err := json.Marshal(data)
	//if err != nil {
	//	return p, err
	//}
	//p.Content = js

	return p, nil
}

func Expr(p *Procedure, node parser.Node) (*Procedure, error) {
	return p, nil
}

func Stringf(p *Procedure, node parser.Node) (*Procedure, error) {
	outdata := make([]interface{}, 0)

	funcExpr := node.(*parser.FuncExpr)
	tag  := funcExpr.Param[0].(*parser.StringLiteral).Val
	fmts := funcExpr.Param[1].(*parser.StringLiteral).Val

	data := make(map[string]interface{})
	err := json.Unmarshal(p.Content, &data)
	if err != nil {
		return p, err
	}

	for i := 2; i < len(funcExpr.Param); i++ {
		switch v := funcExpr.Param[i].(type) {
		case *parser.Jspath:
			outdata = append(outdata, data[v.Jspath])
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

	s := fmt.Sprintf(fmts, outdata...)
	data[tag] = s

	js, err := json.Marshal(data)
	if err != nil {
		return p, err
	}
	p.Content = js

	return p, nil
}

func Cast(p *Procedure, node parser.Node) (*Procedure, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 2 {
		return nil, fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}
	field := funcExpr.Param[0].(*parser.StringLiteral).Val
	tInfo := funcExpr.Param[1].(*parser.StringLiteral).Val

	data := make(map[string]interface{})
	err := json.Unmarshal(p.Content, &data)
	if err != nil {
		return p, err
	}

	data[field] = cast(data[field], tInfo)
	js, err := json.Marshal(data)
	if err != nil {
		return p, err
	}
	p.Content = js

	return p, nil
}

func Group(p *Procedure, node parser.Node) (*Procedure, error) {
	return p, nil
}

func ParseScript(path string) ([]parser.Node, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return parser.ParseFuncExpr(string(data))
}

func logStructed(data string) []byte {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &m)
	if err == nil {
		return []byte(data)
	}

	m[CONTENT] = data
	js, _ := json.Marshal(m)
	return js
}
