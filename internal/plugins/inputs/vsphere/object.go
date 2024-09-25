// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package vsphere

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Object struct {
	class    string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

// Point implement MeasurementV2.
func (m *Object) Point() *point.Point {
	opts := point.DefaultObjectOptions()
	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPointV2(m.class,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (m *Object) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{}
}
