// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type upMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ipt    *Input
}

// Point implement MeasurementV2.
func (m *upMeasurement) Point() *point.Point {
	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		append(point.DefaultMetricOptions(), point.WithExtraTags(m.ipt.mergedTags))...)
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
