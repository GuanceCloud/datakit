package cloudprober

import (
	"net/http"
	"time"

	"github.com/influxdata/telegraf/plugins/common/tls"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName   = `cloudprober`
	l           = logger.DefaultSLogger(inputName)
	minInterval = time.Second
	maxInterval = time.Second * 30
	sample      = `
[[inputs.cloudprober]]
  # Cloudprober 默认指标路由（prometheus format）
  url = "http://localhost:9313/metrics"
  # ##(optional) collection interval, default is 30s
  # interval = "30s"
  ## Optional TLS Config
  # tls_ca = "/xxx/ca.pem"
  # tls_cert = "/xxx/cert.cer"
  # tls_key = "/xxx/key.key"
  ## Use TLS but skip chain & host verification
  insecure_skip_verify = false
  [inputs.cloudprober.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...`
)

type Input struct {
	URL      string           `toml:"url"`
	Interval datakit.Duration `toml:"interval"`
	tls.ClientConfig
	Tags map[string]string `toml:"tags"`

	client  *http.Client
	start   time.Time
	lastErr error

	semStop *cliutils.Sem // start stop signal
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
	return &inputs.MeasurementInfo{}
}
