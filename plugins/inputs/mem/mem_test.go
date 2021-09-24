package mem

import (
	"testing"

	"github.com/shirou/gopsutil/mem"
)

var vmStat = &mem.VirtualMemoryStat{
	Total:          8106131456,
	Available:      992485376,
	Used:           6509629440,
	Free:           267042816,
	Active:         6045421568,
	Inactive:       1150640128,
	Slab:           436809728,
	Wired:          134,
	Buffers:        71581696,
	Cached:         1257877504,
	Shared:         300285952,
	CommitLimit:    6200545280,
	CommittedAS:    21432737792,
	Dirty:          40763392,
	HighFree:       0,
	HighTotal:      0,
	HugePageSize:   2097152,
	HugePagesFree:  0,
	HugePagesTotal: 0,
	LowFree:        0,
	LowTotal:       0,
	Mapped:         373067776,
	PageTables:     63610880,
	SReclaimable:   163459072,
	SUnreclaim:     273350656,
	SwapCached:     21561344,
	SwapFree:       667676672,
	SwapTotal:      2147479552,
	VMallocChunk:   0,
	VMallocTotal:   35184372087808,
	VMallocUsed:    49729536,
	Writeback:      0,
	WritebackTmp:   0,
}

func VirtualMemoryStat4Test() (*mem.VirtualMemoryStat, error) {
	return vmStat, nil
}

func TestMemCollect(t *testing.T) {
	i := &Input{vmStat: VirtualMemoryStat4Test}
	i.platform = "linux" // runtime.GOOS
	i.Collect()
	collect := i.collectCache[0].(*memMeasurement).fields

	assertEqualFloat64(t, 100*float64(vmStat.Used)/float64(vmStat.Total), collect["used_percent"].(float64), "used_percent")
	assertEqualFloat64(t, 100*float64(vmStat.Available)/float64(vmStat.Total), collect["available_percent"].(float64), "available_percent")

	assertEqualUint64(t, vmStat.Active, collect["active"].(uint64), "active")
	assertEqualUint64(t, vmStat.Available, collect["available"].(uint64), "available")
	assertEqualUint64(t, vmStat.Buffers, collect["buffered"].(uint64), "buffered")
	assertEqualUint64(t, vmStat.Cached, collect["cached"].(uint64), "cached")
	assertEqualUint64(t, vmStat.CommitLimit, collect["commit_limit"].(uint64), "commit_limit")
	assertEqualUint64(t, vmStat.CommittedAS, collect["committed_as"].(uint64), "committed_as")
	assertEqualUint64(t, vmStat.Dirty, collect["dirty"].(uint64), "dirty")
	assertEqualUint64(t, vmStat.Free, collect["free"].(uint64), "free")
	assertEqualUint64(t, vmStat.HighFree, collect["high_free"].(uint64), "high_free")
	assertEqualUint64(t, vmStat.HighTotal, collect["high_total"].(uint64), "high_total")
	assertEqualUint64(t, vmStat.HugePageSize, collect["huge_pages_size"].(uint64), "huge_pages_size")
	assertEqualUint64(t, vmStat.HugePagesFree, collect["huge_pages_free"].(uint64), "huge_pages_free")
	assertEqualUint64(t, vmStat.HugePagesTotal, collect["huge_pages_total"].(uint64), "huge_pages_total")
	assertEqualUint64(t, vmStat.Inactive, collect["inactive"].(uint64), "inactive")
	assertEqualUint64(t, vmStat.LowFree, collect["low_free"].(uint64), "low_free")
	assertEqualUint64(t, vmStat.LowTotal, collect["low_total"].(uint64), "low_total")
	assertEqualUint64(t, vmStat.Mapped, collect["mapped"].(uint64), "mapped")
	assertEqualUint64(t, vmStat.PageTables, collect["page_tables"].(uint64), "page_tables")
	assertEqualUint64(t, vmStat.Shared, collect["shared"].(uint64), "shared")
	assertEqualUint64(t, vmStat.Slab, collect["slab"].(uint64), "slab")
	assertEqualUint64(t, vmStat.SReclaimable, collect["sreclaimable"].(uint64), "sreclaimable")
	assertEqualUint64(t, vmStat.SUnreclaim, collect["sunreclaim"].(uint64), "sunreclaim")
	assertEqualUint64(t, vmStat.SwapCached, collect["swap_cached"].(uint64), "swap_cached")
	assertEqualUint64(t, vmStat.SwapFree, collect["swap_free"].(uint64), "swap_free")
	assertEqualUint64(t, vmStat.SwapTotal, collect["swap_total"].(uint64), "swap_total")
	assertEqualUint64(t, vmStat.Total, collect["total"].(uint64), "total")
	assertEqualUint64(t, vmStat.Used, collect["used"].(uint64), "used")
	assertEqualUint64(t, vmStat.VMallocChunk, collect["vmalloc_chunk"].(uint64), "vmalloc_chunk")
	assertEqualUint64(t, vmStat.VMallocTotal, collect["vmalloc_total"].(uint64), "vmalloc_total")
	assertEqualUint64(t, vmStat.VMallocUsed, collect["vmalloc_used"].(uint64), "vmalloc_used")
	assertEqualUint64(t, vmStat.Writeback, collect["write_back"].(uint64), "write_back")
	assertEqualUint64(t, vmStat.WritebackTmp, collect["write_back_tmp"].(uint64), "write_back_tmp")
}

func assertEqualFloat64(t *testing.T, expected, actual float64, mName string) {
	if expected != actual {
		t.Errorf("error: "+mName+" expected: %f \t actual %f", expected, actual)
	}
}

func assertEqualUint64(t *testing.T, expected, actual uint64, mName string) {
	if expected != actual {
		t.Errorf("error: "+mName+" expected: %d \t actual %d", expected, actual)
	}
}
