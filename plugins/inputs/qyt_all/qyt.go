package qyt_all

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/external"
)

const (
	configSample = `
[[inputs.external]]

	name = 'qyt_all'  # required

	# 是否以后台方式运行外部采集器
	daemon = false
	# 如果以非 daemon 方式运行外部采集器，则以该间隔多次运行外部采集器
	interval = '30s'

	# 外部采集器可执行程序路径(尽可能写绝对路径)
	cmd = "/usr/local/cloudcare/dataflux/datakit/externals/qyt_all/main.py" # required

	args = ["/usr/local/cloudcare/dataflux/datakit/externals/qyt_all/config.conf"]
	
`
)

var (
	inputName = "qyt_all"
)

type Qyt struct {
	external.ExernalInput
}

func (_ *Qyt) Catalog() string { return "qyt_all" }

func (_ *Qyt) SampleConfig() string { return configSample }

func (_ *Qyt) Test() (result *inputs.TestResult, err error) {
	return
}

func (o *Qyt) Run() {
	o.ExernalInput.Run()
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Qyt{}
	})
}
