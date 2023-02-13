// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cloudprober

import (
	"net/http"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
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

func (m *Measurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOptElection())
}

func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{}
}
