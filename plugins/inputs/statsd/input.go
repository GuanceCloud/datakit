package statsd

import (
	"bytes"
	"net"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	//"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
)

const (
	sampleConfig = `
[[inputs.statsd]]
	protocol           = "udp"
	bind               = ":8125"
	datadog_extentions = true
	`

	inputName = "statsd"

	udpMaxPktSize = 64 * 1024
)

var (
	l = logger.DefaultSLogger(inputName)
)

type input struct {
	Protocol          string `toml:"protocol"`
	Bind              string `toml:"bind"`
	DataDogExtensions bool   `toml:"datadog_extensions"`

	recvBufSize int

	udpListener *net.UDPConn
	wg          sync.WaitGroup
	in          chan *job

	drops uint64

	bufpool sync.Pool
}

func (x *input) SampleConfig() string {
	return sampleConfig
}

func (x *input) Catalog() string {
	return inputName
}

func (x *input) Run() {
	l = logger.SLogger(inputName)

	l.Info("start statsd...")

	x.bufpool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	switch x.Protocol {
	case "udp", "":
		x.setupUDPServer()
	}

	if x.Protocol == "udp" {
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &input{
			Protocol: "udp",
			Bind:     ":8125",
		}
	})
}
