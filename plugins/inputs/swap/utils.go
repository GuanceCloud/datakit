package swap

import "github.com/shirou/gopsutil/mem"

type SwapStat func() (*mem.SwapMemoryStat, error)

func PSSwapStat() (*mem.SwapMemoryStat, error) {
	return mem.SwapMemory()
}
