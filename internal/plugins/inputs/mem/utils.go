// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mem

import (
	"github.com/shirou/gopsutil/mem"
)

type VMStat func() (*mem.VirtualMemoryStat, error)

func VirtualMemoryStat() (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemory()
}
