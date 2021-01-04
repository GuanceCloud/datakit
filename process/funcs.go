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
	return p, nil
}

func Stringf(p *Procedure, node parser.Node) (*Procedure, error) {
	return p, nil
}

func Cast(p *Procedure, node parser.Node) (*Procedure, error) {
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
