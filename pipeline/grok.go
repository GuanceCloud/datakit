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

var grokCfg *vgrok.Grok

func mergePattners(global, local map[string]string) map[string]string {
	p := make(map[string]string)

	for k, v := range global {
		p[k] = v
	}

	for k, v := range local {
		if _, ok := p[k]; !ok {
			p[k] = v
			continue
		}
	}

	return p
}

func Grok(p *Pipeline, node parser.Node) (*Pipeline, error) {
	if p.grok == nil {
		curPattners := mergePattners(GlobalPatterns, p.patterns)

		g, err := createGrok(curPattners)
		if err != nil {
			l.Warn(err)
			return p, nil
		}
		p.grok = g
	}

	funcExpr := fexpr(node)
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
	funcExpr := fexpr(node)
	if funcExpr.RunOk {
		return p, nil
	}
	if len(funcExpr.Param) != 2 {
		return p, fmt.Errorf("func %s expected 2 args", funcExpr.Name)
	}

	defer func() {
		funcExpr.RunOk = true
	}()

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
		for n, pat := range GlobalPatterns {
			p.patterns[n] = pat
		}
	}

	p.patterns[name] = pattern
	p.grok = nil

	return p, nil
}

func loadPatterns() error {
	loadedPatterns, err := readPatternsFromDir(datakit.PipelinePatternDir)
	if err != nil {
		return err
	}

	for k, v := range loadedPatterns {
		if _, ok := GlobalPatterns[k]; !ok {
			GlobalPatterns[k] = v
		} else {
			l.Warnf("can not overwrite internal pattern `%s', skipped `%s'", k, k)
		}
	}

	g, err := createGrok(GlobalPatterns)
	if err != nil {
		return err
	}
	grokCfg = g

	return nil
}

func readPatternsFromDir(path string) (map[string]string, error) {
	if fi, err := os.Stat(path); err == nil {
		if fi.IsDir() {
			path += "/*"
		}
	} else {
		return nil, fmt.Errorf("invalid path : %s", path)
	}

	files, _ := filepath.Glob(path)

	patterns := make(map[string]string)
	for _, fileName := range files {
		file, err := os.Open(filepath.Clean(fileName))
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

		if err := file.Close(); err != nil {
			l.Warnf("Close: %s, ignored", err.Error())
		}
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
