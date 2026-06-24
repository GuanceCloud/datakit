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

var xfsquotaTaggedby = []string{"project_id", "filesystem_path"}

//nolint:lll
func (*xfsquotaMetric) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "xfsquota",
		Cat:    point.Metric,
		Desc:   "XFS project quota block usage and configured block limits on Linux.",
		DescZh: "Linux 上 XFS project quota 的已用块数以及配置的块数限制。",
		Tags: map[string]interface{}{
			"project_id":      inputs.NewTagInfo("The Project ID in xfs_quota identifies a project or group for disk usage limits."),
			"filesystem_path": inputs.NewTagInfo("The file path of the XFS quota limit."),
		},
		Fields: map[string]interface{}{
			"used": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount, Desc: "The current disk block usage by the project.", Taggedby: xfsquotaTaggedby},
			"soft": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount, Desc: "The soft disk block usage limit for the project.", Taggedby: xfsquotaTaggedby},
			"hard": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount, Desc: "The hard disk block usage limit for the project.", Taggedby: xfsquotaTaggedby},
		},
	}
}
