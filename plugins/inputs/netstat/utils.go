// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

package netstat

import (
	"github.com/shirou/gopsutil/v3/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// NetConnections define a function type.
type NetConnections func() ([]net.ConnectionStat, error)

// GetNetConnections This function implements the NetConnes type. Call an external package to get data.
func GetNetConnections() ([]net.ConnectionStat, error) {
	return net.Connections("all")
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
