package socket

import (
	"bytes"
	"regexp"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName   = "socket"
	metricName  = inputName
	l           = logger.DefaultSLogger(inputName)
	minInterval = time.Second
	maxInterval = time.Second * 30
	sample      = `
# Gather indicators from established connections, using iproute2's ss command.
[[inputs.socket]]
  ## support tcp, udp, raw, unix, packet, dccp and sctp sockets
  ## if socket_types is null, default on udp and tcp
  socket_types = [ "tcp", "udp" ]
  ## The default timeout of 1s for ss execution can be overridden here:
  # interval = "1s"

[inputs.socket.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
)

type socketLister func(cmdName string, proto string) (*bytes.Buffer, error)

type Input struct {
	Interval    datakit.Duration `toml:"interval"`
	SocketProto []string         `toml:"socket_types"`

	isNewConnection *regexp.Regexp
	validValues     *regexp.Regexp
	cmdName         string
	lister          socketLister

	collectCache []inputs.Measurement
	Tags         map[string]string `toml:"tags"`
	semStop      *cliutils.Sem     // start stop signal
}

type Measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *Measurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: inputName,
		Fields: map[string]interface{}{
			"bytes_acked":    newCountFieldInfo("acked bytes"),
			"bytes_received": newCountFieldInfo("bytes received"),
			"segs_out":       newCountFieldInfo("segments out"),
			"segs_in":        newCountFieldInfo("segments in"),
			"data_segs_out":  newCountFieldInfo("data segmentsout"),
			"data_segs_in":   newCountFieldInfo("data segments in"),
		},
		Tags: map[string]interface{}{
			"proto":       inputs.NewTagInfo("the proto type"),
			"local_addr":  inputs.NewTagInfo("local addr"),
			"local_port":  inputs.NewTagInfo("local port"),
			"remote_addr": inputs.NewTagInfo("remote addr"),
			"remote_port": inputs.NewTagInfo("remote port"),
		},
	}
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}
