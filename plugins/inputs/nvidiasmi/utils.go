// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

package nvidiasmi

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// NewFieldInfoB new ibyte data type.
func NewFieldInfoB(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.SizeByte,
		Desc:     desc,
	}
}

// NewFieldInfoP new precent filed.
func NewFieldInfoP(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Float,
		Unit:     inputs.Percent,
		Desc:     desc,
	}
}

// NewFieldInfoC new count field.
func NewFieldInfoC(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		Type:     inputs.Gauge,
		DataType: inputs.Int,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}
