package pipeline

import (
	"fmt"
	"path/filepath"
	"reflect"

	vgrok "github.com/vjeantet/grok"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

var (
	grokCfg *vgrok.Grok
)

func Grok(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	var key, pattern string
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier:
		key = v.Name
	default:
		return p, fmt.Errorf("expect Identifier, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		pattern = v.Val
	default:
		return p, fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	val := p.getContentStr(key)
	m, err := p.grok.Parse(pattern, val)
	if err != nil {
		return p, err
	}

	for k, v := range m {
		p.setContent(k, v)
	}

	return p, nil
}

func loadPatterns() error {
	g, err := vgrok.NewWithConfig(&vgrok.Config{
		NamedCapturesOnly: true,
		PatternsDir: []string{
			filepath.Join(datakit.InstallDir, "pattern"),
		},
	})

	if err != nil {
		return err
	}

	grokCfg = g
	return nil
}
