package process

import (
	"encoding/json"
	"path/filepath"

	"github.com/tidwall/gjson"
	vgrok "github.com/vjeantet/grok"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/process/parser"
)

var (
	grokCfg *vgrok.Grok
)

func Grok(p *Procedure, node parser.Node) (*Procedure, error) {
	filedName := "__content"

	funcExpr := node.(*parser.FuncExpr)
	pattern  := funcExpr.Param[0].(*parser.StringLiteral).Val
	if len(funcExpr.Param) == 2 {
		filedName = funcExpr.Param[1].(*parser.Identifier).Name
	}

	value := gjson.GetBytes(p.Content, filedName)
	m, err := p.grok.Parse(pattern, value.String())
	if err != nil {
		return p, err
	}

	data := make(map[string]interface{})
	if err := json.Unmarshal(p.Content, &data); err != nil {
		return p, err
	}

	for k, v := range m {
		data[k] = v
	}

	if js, err := json.Marshal(data); err != nil {
		return p, err
	} else {
		p.Content = js
	}

    return p, nil
}

func LoadPatterns() {
	g, err := vgrok.NewWithConfig(&vgrok.Config{
		NamedCapturesOnly: true,
		PatternsDir:[]string{
			filepath.Join(datakit.InstallDir, "pattern"),
		},
	})

	if err != nil {
		l.Errorf("grok init err: %v", err)
	}

	grokCfg = g
}
