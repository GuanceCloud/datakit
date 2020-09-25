package tcpdump

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	configSample = `
[inputs.tcpdump]
## 网卡
device = []
## 协议类型
protocol = ['tcp', 'udp']
## 指标集名称，默认值tcpdump
writeUrl = 'http://172.16.0.12:32758/v1/write/metrics?token=tkn_caba81680c8c4fb6b773e95b162623fe'
`
)

var (
	inputName = "tcpdump"
)

type Tcpdump struct {
}

func (_ *Tcpdump) Catalog() string {
	return "network"
}

func (_ *Tcpdump) SampleConfig() string {
	return configSample
}

func (o *Tcpdump) Run() {
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Tcpdump{}
	})
}
