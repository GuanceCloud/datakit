package expressjs

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = "expressjs"
	sampleCfg = `
[[inputs.expressjs]]
	path = "/v1/write/metrics"
	`
)

type ExpressJS struct{}

func (e *ExpressJS) Run()                 {}
func (e *ExpressJS) Catalog()             { return "expressjs" }
func (e *ExpressJS) SampleConfig() string { return sampleCfg }
func (e *ExpressJS) RegHttpHandler()      { /* TODO: regist global v1/write/metric handler */ }

func init() {
	inputs.Add(inputName, func() input.Input {
		return &ExpressJS{}
	})
}
