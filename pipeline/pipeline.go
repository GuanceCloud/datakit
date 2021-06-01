package pipeline

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	influxm "github.com/influxdata/influxdb1-client/models"
	conv "github.com/spf13/cast"
	vgrok "github.com/vjeantet/grok"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/geo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
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
	if p.ast == nil || len(p.ast.Functions) == 0 {
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
	for k, v := range p.Output {
		switch v.(type) {
		case int, uint64, uint32, uint16, uint8, int64, int32, int16, int8, bool, string, float32, float64:
		default:
			str, err := json.Marshal(v)
			if err != nil {
				l.Errorf("object type marshal error %v", err)
			}
			p.Output[k] = string(str)
		}
	}

	return p.Output, p.lastErr
}

func (p *Pipeline) LastError() error {
	return p.lastErr
}

func (p *Pipeline) getContent(key interface{}) (interface{}, error) {
	var k string

	switch t := key.(type) {
	case *parser.Identifier:
		k = t.String()
	case *parser.AttrExpr:
		k = t.String()
	case *parser.StringLiteral:
		k = t.Val
	case string:
		k = t
	default:
		return nil, fmt.Errorf("unsupported %v get", reflect.TypeOf(key).String())
	}

	if k == "_" {
		return p.Content, nil
	}

	v, ok := p.Output[k]
	if !ok {
		return nil, fmt.Errorf("%v no found", k)
	}

	return v, nil
}

func (p *Pipeline) getContentStr(key interface{}) (string, error) {
	c, err := p.getContent(key)

	switch v := reflect.ValueOf(c); v.Kind() {
	case reflect.Map:
		res, err := json.Marshal(v.Interface())
		return string(res), err
	default:
		return conv.ToString(v), err
	}
}

func (p *Pipeline) setContent(k, v interface{}) error {
	var key string

	switch t := k.(type) {
	case *parser.Identifier:
		key = t.String()
	case *parser.AttrExpr:
		key = t.String()
	case *parser.StringLiteral:
		key = t.Val
	case string:
		key = t
	default:
		return fmt.Errorf("unsupported %v set", reflect.TypeOf(key).String())
	}

	if p.Output == nil {
		p.Output = make(map[string]interface{})
	}

	if v == nil {
		return nil
	}

	p.Output[key] = v

	return nil
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

func Init(datadir string) error {
	l = logger.SLogger("pipeline")

	if err := geo.LoadIPLib(filepath.Join(datadir, "iploc.bin")); err != nil {
		return err
	}

	if err := ip2isp.Init(filepath.Join(datadir, "ip2isp.txt")); err != nil {
		return err
	}

	if err := loadPatterns(); err != nil {
		return err
	}

	return nil
}
