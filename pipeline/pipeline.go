package pipeline

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	vgrok "github.com/vjeantet/grok"
	conv  "github.com/spf13/cast"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/patterns"
)

type Pipeline struct {
	Content  string
	Output   map[string]interface{}
	lastErr  error
	patterns map[string]string //存放自定义patterns
	nodes    []parser.Node
	grok     *vgrok.Grok
}

var (
	l = logger.DefaultSLogger("process")
)

func NewPipelineByScriptPath(path string) (*Pipeline, error) {

	scriptPath := filepath.Join(datakit.PipelineDir, path)
	data, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		return nil, err
	}
	return NewPipeline(string(data))
}

func NewPipeline(script string) (*Pipeline, error) {
	p := &Pipeline{
		Output: make(map[string]interface{}),
		grok:   grokCfg,
	}

	if err := p.parseScript(script); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Pipeline) Run(data string) *Pipeline {
	defer func() {
		r := recover()
		if r != nil {
			p.lastErr = fmt.Errorf("%v", r)
		}
	}()

	var err error

	//防止脚本解析错误
	if p.lastErr != nil {
		return p
	}

	p.Content = data
	p.Output = make(map[string]interface{})

	for _, node := range p.nodes {
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

			_, err = f(p, node)
			if err != nil {
				l.Errorf("Run func %v: %v", v.Name, err)
				p.lastErr = err
				return p
			}

		default:
			p.lastErr = fmt.Errorf("%v not function", v.String())
		}
	}
	return p
}

func (p *Pipeline) Result() (map[string]interface{}, error) {
	return p.Output, p.lastErr
}

func (p *Pipeline) LastError() error {
	return p.lastErr
}

func (p *Pipeline) getContent(key string) interface{} {
	if key == "_" {
		return p.Content
	}

	if v, ok := p.Output[key]; ok {
		return v
	}

	var m interface{}
	var nm interface{}

	m = p.Output
	keys := strings.Split(key, ".")
	for _, k := range keys {
		switch m.(type) {
		case map[string]interface{}:
			v := m.(map[string]interface{})
			nm = v[k]
			m = nm
		default:
			nm = nil
		}
	}

	return nm
}

func (p *Pipeline) getContentStr(key string) string {
	return conv.ToString(p.getContent(key))
}

func (p *Pipeline) setContent(k string, v interface{}) {
	if p.Output == nil {
		p.Output = make(map[string]interface{})
	}

	p.Output[k] = v
}

func (pl *Pipeline) parseScript(script string) error {

	nodes, err := parser.ParseFuncExpr(script)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		switch v := node.(type) {
		case *parser.FuncExpr:
			debugNodesHelp(v, "")
		default:
			return fmt.Errorf("should not been here")
		}
	}

	pl.nodes = nodes
	return nil
}

func debugNodesHelp(f *parser.FuncExpr, prev string) {
	l.Debugf("%v%v", prev, f.Name)

	for _, node := range f.Param {
		switch v := node.(type) {
		case *parser.FuncExpr:
			debugNodesHelp(v, prev+"    ")
		default:
			l.Debugf("%v%v", prev+"    |", node)
		}
	}
}

func Init() error {
	if err := patterns.InitPatternsFile(); err != nil {
		return err
	}

	if err := loadPatterns(); err != nil {
		return err
	}

	return nil
}
