// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type customerObjectMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

// Point implement MeasurementV2.
func (m *customerObjectMeasurement) Point() *point.Point {
	opts := point.DefaultObjectOptions()
	if m.election {
		opts = append(opts,
			point.WithExtraTags(datakit.GlobalElectionTags()),
		)
		point.DefaultObjectOptions()
	}
	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (m *customerObjectMeasurement) Info() *inputs.MeasurementInfo { //nolint:funlen
	return &inputs.MeasurementInfo{
		Name:           "database",
		MetaDuplicated: true,
		Cat:            point.CustomObject,
		Fields: map[string]interface{}{
			"uptime": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationSecond,
				Desc:     "Current instance uptime",
			},

			"display_name": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "Displayed name in UI",
			},

			"version": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.Gauge,
				Unit:     inputs.NoUnit,
				Desc:     "Current version of the instance",
			},
		},
		Tags: map[string]interface{}{
			"name": &inputs.TagInfo{
				Desc: "Object uniq ID",
			},

			"col_co_status": &inputs.TagInfo{
				Desc: "Current status of collector on instance(`OK/NotOK`)",
			},

			"ip": &inputs.TagInfo{
				Desc: "Connection IP of the instance",
			},

			"host": &inputs.TagInfo{
				Desc: "The server host address",
			},
			"reason": &inputs.TagInfo{
				Desc: "If status not ok, we'll get some reasons about the status",
			},
		},
	}
}
