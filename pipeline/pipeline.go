package pipeline

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	influxm "github.com/influxdata/influxdb1-client/models"
	conv "github.com/spf13/cast"
	vgrok "github.com/vjeantet/grok"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/geo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/patterns"
)

type Pipeline struct {
	Content  string
	Output   map[string]interface{}
	lastErr  error
	patterns map[string]string //存放自定义patterns
	ast      *parser.Ast
	grok     *vgrok.Grok
}

var (
	l = logger.DefaultSLogger("pipeline")
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
		return p, err
	}

	return p, nil
}

func NewPipelineFromFile(filename string) (*Pipeline, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewPipeline(string(b))
}

// PointToJSON, line protocol point to pipeline JSON
func (p *Pipeline) RunPoint(point influxm.Point) *Pipeline {
	defer func() {
		r := recover()
		if r != nil {
			p.lastErr = fmt.Errorf("%v", r)
		}
	}()

	m := map[string]interface{}{"measurement": string(point.Name())}

	if tags := point.Tags(); len(tags) > 0 {
		m["tags"] = map[string]string{}
		for _, tag := range tags {
			m["tags"].(map[string]string)[string(tag.Key)] = string(tag.Value)
		}
	}

	fields, err := point.Fields()
	if err != nil {
		p.lastErr = err
		return p
	}

	for k, v := range fields {
		m[k] = v
	}

	m["time"] = point.UnixNano()

	j, err := json.Marshal(m)
	if err != nil {
		p.lastErr = err
		return p
	}

	return p.Run(string(j))
}

func (p *Pipeline) Run(data string) *Pipeline {

	p.Content = data
	p.Output = make(map[string]interface{})
	p.Output["message"] = data

	//防止脚本解析错误
	if len(p.ast.Functions) == 0 {
		return p
	}

	//错误状态复位
	p.lastErr = nil

	var f rtpanic.RecoverCallback

	f = func(trace []byte, err error) {

		defer rtpanic.Recover(f, nil)

		if trace != nil {
			l.Error("panic: %s", string(trace))
			p.lastErr = fmt.Errorf("%s", trace)
			return
		}

		for _, fn := range p.ast.Functions {
			fname := strings.ToLower(fn.Name)
			plf, ok := funcsMap[fname]
			if !ok {
				p.lastErr = fmt.Errorf("unsupported func: `%v'", fn.Name)
				return
			}

			_, err = plf(p, fn)
			if err != nil {
				p.lastErr = fmt.Errorf("Run func %v: %v", fn.Name, err)
				return
			}
		}
	}

	f(nil, nil)
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

func (p *Pipeline) getContentStrByCheck(key string) (string, bool) {
	v := p.getContent(key)
	if v == nil {
		return "", false
	}

	return conv.ToString(v), true
}

func (p *Pipeline) setContent(k string, v interface{}) {
	if p.Output == nil {
		p.Output = make(map[string]interface{})
	}

	if v == nil {
		return
	}

	p.Output[k] = v
}

func (pl *Pipeline) parseScript(script string) error {

	node, err := parser.ParsePipeline(script)
	if err != nil {
		return err
	}

	switch ast := node.(type) {
	case *parser.Ast:
		pl.ast = ast
	default:
		return fmt.Errorf("should not been here")
	}

	return nil
}

func debugNodesHelp(f *parser.FuncExpr, prev string) {
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

	l = logger.SLogger("pipeline")

	if err := geo.Init(); err != nil {
		return err
	}

	if err := patterns.InitPatternsFile(); err != nil {
		return err
	}

	if err := loadPatterns(); err != nil {
		return err
	}

	return nil
}
