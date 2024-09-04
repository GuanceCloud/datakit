// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package graphite config
package graphite

import (
	"regexp"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/graphite/mapper"
)

const (
	TCP               = "tcp"
	UDP               = "udp"
	minInterval       = time.Second
	maxInterval       = time.Second * 120
	defaultBufferSize = 100
	defaultPort       = ":9109"
	inputName         = "graphite"
	sampleConfig      = `
[[inputs.graphite]]
  ## Address to open UDP/TCP, default 9109
  address = ":9109"

  # Whether to open StrictMatch
  # strict_match = false

  ## Example Mapping Configuration
  #[inputs.graphite.metric_mapper]
  # name = "test"
  # [[inputs.graphite.metric_mapper.mappings]]
  # match = "test.dispatcher.*.*.*"
  # name = "dispatcher_events_total"

  # [inputs.graphite.metric_mapper.mappings.labels]
  # action = "$2"
  # job = "test_dispatcher"
  # outcome = "$3"
  # processor = "$1"

  # [[inputs.graphite.metric_mapper.mappings]]
  # match = "*.signup.*.*"
  # name = "signup_events_total"

  # [inputs.graphite.metric_mapper.mappings.labels]
  # job = "${1}_server"
  # outcome = "$3"
  # provider = "$2"

  # Regex Mapping Example
  # [[inputs.graphite.metric_mapper.mappings]]
  # match = '''servers\.(.*)\.networking\.subnetworks\.transmissions\.([a-z0-9-]+)\.(.*)'''
  # match_type = "regex"
  # name = "servers_networking_transmissions_${3}"

  # [inputs.graphite.metric_mapper.mappings.labels]
  # hostname = "${1}"
  # device = "${2}"
`
)

var (
	g                  = datakit.G("inputs_graphite")
	invalidMetricChars = regexp.MustCompile("[^a-zA-Z0-9_:]")
)

type Measurement struct {
	Name   string
	Tags   map[string]string
	Fields map[string]interface{}
	TS     time.Time
}

//nolint:lll
func (m *Measurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   inputName,
		Type:   "metric",
		Desc:   "Graphite exporter metrics",
		Fields: map[string]interface{}{},
		Tags:   map[string]interface{}{},
	}
}

func (m *Measurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.TS))

	return point.NewPointV2(m.Name,
		append(point.NewTags(m.Tags), point.NewKVs(m.Fields)...),
		opts...)
}

type graphiteMetric struct {
	OriginalName string
	Name         string
	Value        float64
	Labels       mapper.Labels
	Timestamp    time.Time
}
