// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package iotdb collects Apache IoTDB server metrics from its Prometheus endpoint.
package iotdb

import (
	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	_ inputs.InputV2       = (*Input)(nil)
	_ inputs.Input         = (*Input)(nil)
	_ inputs.ElectionInput = (*Input)(nil)
)

const (
	inputName = "iotdb"
)

var l = logger.DefaultSLogger(inputName)

const sampleConfig = `
[[inputs.iotdb]]
  ## IoTDB Prometheus metrics endpoint. Enable the PROMETHEUS reporter on IoTDB.
  endpoint = "http://127.0.0.1:9091/metrics"

  ## Optional Basic Auth for a reverse proxy in front of the metrics endpoint.
  # username = ""
  # password = ""

  ## Collection interval and request timeout.
  interval = "30s"
  timeout = "10s"

  ## Set true to enable election.
  election = true

  ## Additional tags.
  [inputs.iotdb.tags]
  # cluster_name = "testing"
  # some_tag = "some_value"
`

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Endpoint: defaultEndpoint,
			Interval: defaultInterval,
			Timeout:  defaultTimeout,
			Tags:     map[string]string{},
			Election: true,
			pauseCh:  make(chan bool, 8),
			semStop:  cliutils.NewSem(),
			feeder:   dkio.DefaultFeeder(),
			Tagger:   datakit.DefaultGlobalTagger(),
		}
	})
}
