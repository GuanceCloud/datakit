// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package swap

import "github.com/shirou/gopsutil/mem"

type SwapStat func() (*mem.SwapMemoryStat, error)

func PSSwapStat() (*mem.SwapMemoryStat, error) {
	return mem.SwapMemory()
}
