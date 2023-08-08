// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jenkins

import (
	"net/http"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var (
	inputName   = `jenkins`
	l           = logger.DefaultSLogger(inputName)
	minInterval = time.Second
	maxInterval = time.Second * 30
	sample      = `
[[inputs.jenkins]]
  ## Set true if you want to collect metric from url below.
  enable_collect = true

  ## The Jenkins URL in the format "schema://host:port",required
  url = "http://my-jenkins-instance:8080"

  ## Metric Access Key ,generate in your-jenkins-host:/configure,required
  key = ""

  # ##(optional) collection interval, default is 30s
  # interval = "30s"

  ## Set response_timeout
  # response_timeout = "5s"

  ## Set true to enable election
  # election = true

  ## Optional TLS Config
  # tls_ca = "/xx/ca.pem"
  # tls_cert = "/xx/cert.pem"
  # tls_key = "/xx/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## set true to receive jenkins CI event
  enable_ci_visibility = true

  ## which port to listen to jenkins CI event
  ci_event_port = ":9539"

  # [inputs.jenkins.log]
  # files = []
  # #grok pipeline script path
  # pipeline = "jenkins.p"

  [inputs.jenkins.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...

  [inputs.jenkins.ci_extra_tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`

	pipelineCfg = `
grok(_, "%{TIMESTAMP_ISO8601:time} \\[id=%{GREEDYDATA:id}\\]\t%{GREEDYDATA:status}\t")
default_time(time)
group_in(status, ["WARNING", "NOTICE"], "warning")
group_in(status, ["SEVERE", "ERROR"], "error")
group_in(status, ["INFO"], "info")

`
)

type jenkinslog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`
}

type Input struct {
	EnableCollect      bool              `toml:"enable_collect"`
	URL                string            `toml:"url"`
	Key                string            `toml:"key"`
	Interval           datakit.Duration  `toml:"interval"`
	ResponseTimeout    datakit.Duration  `toml:"response_timeout"`
	Election           bool              `toml:"election"`
	Log                *jenkinslog       `toml:"log"`
	Tags               map[string]string `toml:"tags"`
	EnableCIVisibility bool              `toml:"enable_ci_visibility"`
	CIEventPort        string            `toml:"ci_event_port"`
	CIExtraTags        map[string]string `toml:"ci_extra_tags"`

	tls.ClientConfig
	// HTTP client
	client *http.Client

	start time.Time
	tail  *tailer.Tailer

	srv *http.Server

	lastErr      error
	collectCache []*point.Point

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  dkpt.GlobalTagger
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}

func newRateFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Float,
		Type:     inputs.Gauge,
		Unit:     inputs.Percent,
		Desc:     desc,
	}
}
