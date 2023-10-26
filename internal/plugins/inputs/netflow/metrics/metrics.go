// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package metrics contains metrics' definition.
package metrics

import (
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

// MetricType is the representation of an aggregator metric type.
type MetricType int

// metric type constants enumeration.
const (
	GaugeType MetricType = iota
	RateType
	CountType
	MonotonicCountType
	CounterType
	HistogramType
	HistorateType
	SetType
	DistributionType

	// NumMetricTypes is the number of metric types; must be the last item here.
	NumMetricTypes
)

////////////////////////////////////////////////////////////////////////////////

const DefaultSource = "netflow"

type NetflowMeasurement struct {
	Name   string
	Tags   map[string]string
	Fields map[string]interface{}
	TS     time.Time
}

// Point implement MeasurementV2.
func (m *NetflowMeasurement) Point() *point.Point {
	opts := point.DefaultLoggingOptions()
	opts = append(opts, point.WithTime(m.TS))

	return point.NewPointV2(m.Name,
		append(point.NewTags(m.Tags), point.NewKVs(m.Fields)...),
		opts...)
}

//nolint:lll
func (*NetflowMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: DefaultSource,
		Type: "logging",
		Desc: "Using `source` field in the config file, default is `default`.",
		Tags: map[string]interface{}{
			"ip":   inputs.NewTagInfo("Collector IP address."),
			"host": inputs.NewTagInfo("Hostname."),
		},
		Fields: map[string]interface{}{
			"message":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The text of the logging."},
			"status":      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The status of the logging, only supported `info/emerg/alert/critical/error/warning/debug/OK/unknown`."},
			"bytes":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.SizeByte, Desc: "Flow bytes."},
			"dest_ip":     &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Flow destination IP."},
			"dest_port":   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Flow destination port."},
			"device_ip":   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "NetFlow exporter IP."},
			"ip_protocol": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Flow network protocol."},
			"source_ip":   &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Flow source IP."},
			"source_port": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Flow source port."},
			"type":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Flow type."},
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
