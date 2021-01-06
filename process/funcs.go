package process

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/tidwall/gjson"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/process/parser"

)
const (
	CONTENT = "__content"
)

type ProFunc func(p *Procedure, node parser.Node) (*Procedure, error)

var (
	funcsMap = map[string]ProFunc{
		"grok"       : Grok,
		"rename"     : Rename   ,
		"stringf"    : Stringf  ,
		"cast"       : Cast     ,
		"expr"       : Expr     ,

		"user_agent" : UserAgent,
		"url_decode" : UrlDecode,
		"geo_ip"     : GeoIp    ,
		"date"       : Date     ,
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

	old := funcExpr.Param[0].(*parser.Identifier).Name
	new := funcExpr.Param[1].(*parser.Identifier).Name

	data := make(map[string]interface{})
	err := json.Unmarshal(p.Content, &data)
	if err != nil {
		return p, err
	}

	r := gjson.GetBytes(p.Content, old)
	data[new] = getGjsonRst(&r)
	delete(data, old)

	if js, err := json.Marshal(data); err != nil {
		return p, err
	} else {
		p.Content = js
	}

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

	return p, nil
}

func Expr(p *Procedure, node parser.Node) (*Procedure, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 2 {
		return nil, fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	tag  := funcExpr.Param[1].(*parser.Identifier).Name
	expr := funcExpr.Param[0].(*parser.BinaryExpr)

	data := make(map[string]interface{})
	if err := json.Unmarshal(p.Content, &data); err != nil {
		return p, err
	}

	if v, err := Calc(expr, p.Content); err != nil {
		return p, err
	} else {
		data[tag] = v
	}

	if js, err := json.Marshal(data); err != nil {
		return p, err
	} else {
		p.Content = js
	}

	return p, nil
}

func Stringf(p *Procedure, node parser.Node) (*Procedure, error) {
	outdata := make([]interface{}, 0)

	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) < 2 {
		return nil, fmt.Errorf("func %s expected more than 2 args", funcExpr.Name)
	}

	tag  := funcExpr.Param[0].(*parser.Identifier).Name
	fmts := funcExpr.Param[1].(*parser.StringLiteral).Val

	data := make(map[string]interface{})
	err := json.Unmarshal(p.Content, &data)
	if err != nil {
		return p, err
	}

	for i := 2; i < len(funcExpr.Param); i++ {
		switch v := funcExpr.Param[i].(type) {
		case *parser.Identifier:
			gRst := gjson.GetBytes(p.Content, v.Name)
			outdata = append(outdata, getGjsonRst(&gRst))
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

	if js, err := json.Marshal(data); err != nil {
		return p, err
	} else {
		p.Content = js
	}

	return p, nil
}

func Cast(p *Procedure, node parser.Node) (*Procedure, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 3 {
		return nil, fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	nField := funcExpr.Param[0].(*parser.Identifier).Name
	field := funcExpr.Param[1].(*parser.Identifier).Name
	tInfo := funcExpr.Param[2].(*parser.StringLiteral).Val

	data := make(map[string]interface{})
	err := json.Unmarshal(p.Content, &data)
	if err != nil {
		return p, err
	}

	rst := gjson.GetBytes(p.Content, field)
	data[nField] = cast(&rst, tInfo)

	if js, err := json.Marshal(data); err != nil {
		return p, err
	} else {
		p.Content = js
	}

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

	nodes, err := parser.ParseFuncExpr(string(data))
	for _, node := range nodes {
		switch v := node.(type) {
		case *parser.FuncExpr:
			DebugNodesHelp(v, "")
		default:
		}
	}

	return nodes, err
}

func DebugNodesHelp(f *parser.FuncExpr, prev string)  {
	l.Debugf("%v%v", prev, f.Name)

	for _, node := range f.Param {
		switch v := node.(type) {
		case *parser.FuncExpr:
			DebugNodesHelp(v, prev+"    ")
		default:
			l.Debugf("%v%v", prev+"    |", node)
		}
	}
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

func getGjsonRst(g *gjson.Result) interface{} {
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