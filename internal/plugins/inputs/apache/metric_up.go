// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package apache

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type upMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

// Point implement MeasurementV2.
func (m *upMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *upMeasurement) Info() *inputs.MeasurementInfo { //nolint:funlen
	return &inputs.MeasurementInfo{
		Name: "collector",
		Type: "metric",
		Fields: map[string]interface{}{
			"up": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "",
			},
		},
		Tags: map[string]interface{}{
			"job": &inputs.TagInfo{
				Desc: "Server name",
			},
			"instance": &inputs.TagInfo{
				Desc: "Server addr",
			},
		},
	}
}
