// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rabbitmq

import (
	"time"

	"github.com/GuanceCloud/cliutils/logger"
)

var (
	inputName            = `rabbitmq`
	customObjectFeedName = inputName + "-CO"
	l                    = logger.DefaultSLogger(inputName)
	minInterval          = time.Second
	maxInterval          = time.Second * 30
	sample               = `
[[inputs.rabbitmq]]
  # rabbitmq url ,required
  url = "http://localhost:15672"

  # rabbitmq user, required
  username = "guest"

  # rabbitmq password, required
  password = "guest"

  # ##(optional) collection interval, default is 30s
  # interval = "30s"

  ## Optional TLS Config
  # tls_ca = "/xxx/ca.pem"
  # tls_cert = "/xxx/cert.cer"
  # tls_key = "/xxx/key.key"
  ## Use TLS but skip chain & host verification
  insecure_skip_verify = false

  ## Set true to enable election
  election = true

  # [inputs.rabbitmq.log]
  # files = []
  # #grok pipeline script path
  # pipeline = "rabbitmq.p"

  [inputs.rabbitmq.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...

`
	pipelineCfg = `
grok(_, "%{LOGLEVEL:status}%{DATA}====%{SPACE}%{DATA:time}%{SPACE}===%{SPACE}%{GREEDYDATA:msg}")

grok(_, "%{DATA:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}")

default_time(time)
`
)

const (
	overviewMeasurementName = "rabbitmq_overview"
	exchangeMeasurementName = "rabbitmq_exchange"
	nodeMeasurementName     = "rabbitmq_node"
	queueMeasurementName    = "rabbitmq_queue"
)
