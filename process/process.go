package process

import (
	"fmt"
	"strings"

	vgrok "github.com/vjeantet/grok"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/process/parser"
)

type Procedure struct {
	Content   []byte
	lastErr   error
	patterns  map[string]string    //存放自定义patterns
	nodes     []parser.Node
	grok      *vgrok.Grok
}

var (
	l = logger.DefaultSLogger("process")
)

func (p *Procedure) LastError() error {
	return p.lastErr
}

func (p *Procedure) GetContentByte() []byte {
	return p.Content
}

func (p *Procedure) GetContentStr() string {
	return string(p.Content)
}

func (p *Procedure) GetSubContent(id string) interface{} {
	return getContentById(p.Content, id)
}

func (p *Procedure) ProcessText(data string) *Procedure {
	var err error

	if p.lastErr != nil {
		return p
	}

	p.Content = logStructed(data)

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

			p, err = f(p, node)
			if err != nil {
				l.Errorf("ProcessLog func %v: %v", v.Name, err)
				p.lastErr = err
				return p
			}

		default:
			p.lastErr = fmt.Errorf("%v not function", v.String())
		}
	}
	return p
}

func NewProcedure(script string) *Procedure {
	p := &Procedure{}
	if script != "" {
		p.nodes, p.lastErr = ParseScript(script)
	}
	return p
}

func init() {
	GenScript()
}
