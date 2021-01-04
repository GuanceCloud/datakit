package grok

import (
	"encoding/json"
	"fmt"

	vgrok "github.com/vjeantet/grok"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/process"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/process/parser"
)

var (
	grokCfg *vgrok.Grok
)

func Grok(p *process.Procedure, node parser.Node) (*process.Procedure, error) {
	filedName := "__content"

	funcExpr := node.(*parser.FuncExpr)
	pattern  := funcExpr.Param[0].(*parser.StringLiteral).Val
	if len(funcExpr.Param) == 2 {
		filedName = funcExpr.Param[1].(*parser.StringLiteral).Val
	}

	data := make(map[string]interface{})
	err := json.Unmarshal(p.Content, &data)
	if err != nil {
		return p, err
	}

	m, err := grok(pattern, filedName)
	if err != nil {
		return p, err
	}

	for k, v := range m {
		data[k] = v
	}

	js, err := json.Marshal(data)
	if err != nil {
		return p, err
	}

	p.Content = js
    return p, nil
}

func grok(pattern, text string) (map[string]string, error){
	if grokCfg == nil {
		return nil, fmt.Errorf("grok nil")
	}
	return grokCfg.Parse(pattern, text)
}

func init() {
	g, err := vgrok.NewWithConfig(&vgrok.Config{NamedCapturesOnly: true})
	if err != nil {
		fmt.Printf("grok init err: %v", err)
	}
	grokCfg = g
}
