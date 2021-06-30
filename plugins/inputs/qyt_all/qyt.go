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
	daemon = true

	# 外部采集器可执行程序路径(尽可能写绝对路径)
	cmd = "python3" # required

	args = [
		"/usr/local/datakit/externals/qyt_all/main.py",
		"/usr/local/datakit/externals/qyt_all/config.conf"
	]

	envs = ['LD_LIBRARY_PATH=/path/to/lib:$LD_LIBRARY_PATH',]
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

func (o *Qyt) Run() {
	o.ExernalInput.Run()
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Qyt{}
	})
}
