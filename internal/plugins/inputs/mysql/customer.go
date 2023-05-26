// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"time"

	gcPoint "github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type customerMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	ts       time.Time
	election bool
}

// 生成行协议.
func (m *customerMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOptElectionV2(m.election))
}

// Point implement MeasurementV2.
func (m *customerMeasurement) Point() *gcPoint.Point {
	opts := gcPoint.DefaultMetricOptions()

	if m.election {
		opts = append(opts, gcPoint.WithExtraTags(point.GlobalElectionTags()))
	}

	return gcPoint.NewPointV2([]byte(m.name),
		append(gcPoint.NewTags(m.tags), gcPoint.NewKVs(m.fields)...),
		opts...)
}

// 指定指标.
func (m *customerMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "mysql_customer",
		Type: "metric",
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{
				Desc: "Server addr",
			},
		},
	}
}
