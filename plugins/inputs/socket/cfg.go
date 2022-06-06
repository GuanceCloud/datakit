package socket

import (
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	KillGrace = 5 * time.Second
	TCP       = "tcp"
	UDP       = "udp"
)

var (
	inputName   = "socket"
	metricName  = inputName
	l           = logger.DefaultSLogger(inputName)
	minInterval = time.Second * 300
	maxInterval = time.Second * 30000
	sample      = `
# Gather indicators from established connections, using iproute2's ss command.
[[inputs.socket]]
  ## support tcp, udp, raw, unix, packet, dccp and sctp sockets
  ## if socket_types is null, default on udp and tcp
  dest_url = ["tcp:47.110.144.10:443", "udp:1.1.1.1:5555"]

  ## @param interval - number - optional - default: 900, min 300
  interval = "900s"

[inputs.socket.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
)

type Input struct {
	DestURL  []string         `toml:"dest_url"`
	Interval datakit.Duration `toml:"interval"` // 单位为秒

	curTasks map[string]*dialer
	wg       sync.WaitGroup

	collectCache []inputs.Measurement
	Tags         map[string]string `toml:"tags"`
	semStop      *cliutils.Sem     // start stop signal
	platform     string
}

type TCPMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

type UDPMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *TCPMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *UDPMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *TCPMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "tcp_dial_testing",
		Tags: map[string]interface{}{
			"dest_host": &inputs.TagInfo{Desc: "示例 wwww.baidu.com"},
			"dest_port": &inputs.TagInfo{Desc: "示例 80"},
			"proto":     &inputs.TagInfo{Desc: "示例 tcp"},
		},
		Fields: map[string]interface{}{
			"fail_reason": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "拨测失败原因",
			},
			"response_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "TCP 连接时间, 单位",
			},
			"response_time_with_dns": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationUS,
				Desc:     "连接时间（含DNS解析）, 单位",
			},
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "只有 1/-1 两种状态, 1 表示成功, -1 表示失败",
			},
		},
	}
}

func (m *UDPMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "udp",
		Fields: map[string]interface{}{
			"success": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "只有 1/-1 两种状态, 1 表示成功, -1 表示失败",
			},
			"fail_reason": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.UnknownUnit,
				Desc:     "拨测失败原因",
			},
		},
		Tags: map[string]interface{}{
			"dest_host": &inputs.TagInfo{Desc: "示例 wwww.baidu.com"},
			"dest_port": &inputs.TagInfo{Desc: "示例 80"},
			"proto":     &inputs.TagInfo{Desc: "示例 udp"},
		},
	}
}
