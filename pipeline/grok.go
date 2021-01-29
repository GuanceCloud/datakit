package pipeline

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	vgrok "github.com/vjeantet/grok"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

var (
	grokCfg        *vgrok.Grok
	globalPatterns map[string]string
)

func Grok(p *Pipeline, node parser.Node) (*Pipeline, error) {
	if p.grok == nil {
		return p, fmt.Errorf("grok not init")
	}

	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	var key parser.Node
	var pattern string
	switch v := funcExpr.Param[0].(type) {
	case *parser.Identifier, *parser.AttrExpr:
		key = v
	default:
		return p, fmt.Errorf("expect Identifier or AttrExpr, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		pattern = v.Val
	default:
		return p, fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	val, err := p.getContentStr(key)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	m, err := p.grok.Parse(pattern, val)
	if err != nil {
		l.Warn(err)
		return p, nil
	}

	for k, v := range m {
		err := p.setContent(k, v)
		if err != nil {
			l.Warn(err)
			return p, nil
		}
	}

	return p, nil
}

func AddPattern(p *Pipeline, node parser.Node) (*Pipeline, error) {
	funcExpr := node.(*parser.FuncExpr)
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	var name, pattern string
	switch v := funcExpr.Param[0].(type) {
	case *parser.StringLiteral:
		name = v.Val
	default:
		return p, fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[0]).String())
	}

	switch v := funcExpr.Param[1].(type) {
	case *parser.StringLiteral:
		pattern = v.Val
	default:
		return p, fmt.Errorf("expect StringLiteral, got %s",
			reflect.TypeOf(funcExpr.Param[1]).String())
	}

	if p.patterns == nil {
		p.patterns = make(map[string]string)
		for n, pat := range globalPatterns {
			p.patterns[n] = pat
		}
	}
	p.patterns[name] = pattern
	p.grok = nil
	g, err := createGrok(p.patterns)
	if err != nil {
		l.Warn(err)
		return p, nil
	}
	p.grok = g

	return p, nil
}

func loadPatterns() error {
	p, err := readPatternsFromDir(filepath.Join(datakit.InstallDir, "pattern"))
	if err != nil {
		return err
	}
	globalPatterns = p

	g, err := createGrok(p)
	if err != nil {
		return err
	}
	grokCfg = g

	return nil
}

func readPatternsFromDir(path string) (map[string]string, error) {
	if fi, err := os.Stat(path); err == nil {
		if fi.IsDir() {
			path = path + "/*"
		}
	} else {
		return nil, fmt.Errorf("invalid path : %s", path)
	}

	files, _ := filepath.Glob(path)

	patterns := make(map[string]string)
	for _, fileName := range files {
		file, err := os.Open(fileName)
		if err != nil {
			return patterns, err
		}

		scanner := bufio.NewScanner(bufio.NewReader(file))

		for scanner.Scan() {
			l := scanner.Text()
			if len(l) > 0 && l[0] != '#' {
				names := strings.SplitN(l, " ", 2)
				patterns[names[0]] = names[1]
			}
		}

		file.Close()
	}

	return patterns, nil
}

func createGrok(pattern map[string]string) (*vgrok.Grok, error) {
	return vgrok.NewWithConfig(&vgrok.Config{
		SkipDefaultPatterns: true,
		NamedCapturesOnly:   true,
		Patterns:            pattern,
	})
}
