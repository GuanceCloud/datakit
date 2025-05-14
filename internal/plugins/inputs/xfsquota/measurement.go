// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package xfsquota implements the collection of quota information for the XFS file system.
package xfsquota

import (
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type xfsquotaMetric struct{}

//nolint:lll
func (*xfsquotaMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "xfsquota",
		Cat:  point.Metric,
		Desc: "The info of xfs_quota, only supported Linux system.",
		Tags: map[string]interface{}{
			"project_id":      inputs.NewTagInfo("The Project ID in xfs_quota identifies a project or group for disk usage limits."),
			"filesystem_path": inputs.NewTagInfo("The file path of the XFS quota limit."),
		},
		Fields: map[string]interface{}{
			"used": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The current disk usage by the project."},
			"soft": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The soft limit for disk usage."},
			"hard": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The hard limit for disk usage."},
		},
	}
}
