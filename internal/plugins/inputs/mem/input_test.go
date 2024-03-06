// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mem

import (
	"testing"

	"github.com/shirou/gopsutil/mem"
	"github.com/stretchr/testify/assert"
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

type mockTagger struct{}

func (t *mockTagger) HostTags() map[string]string {
	return map[string]string{
		"host": "me",
	}
}

func (t *mockTagger) ElectionTags() map[string]string {
	return nil
}
func (t *mockTagger) UpdateVersion() {}
func (t *mockTagger) Updated() bool  { return false }

func TestMemCollect(t *testing.T) {
	i := &Input{
		vmStat: VirtualMemoryStat4Test,
		tagger: &mockTagger{},
		Tags: map[string]string{
			"t1": "v1",
		},
	}

	i.platform = "linux" // runtime.GOOS

	i.setup()

	if err := i.collect(); err != nil {
		t.Error(err)
	}

	assert.Len(t, i.collectCache, 1)

	pt := i.collectCache[0]

	assert.Equalf(t,
		100*float64(vmStat.Used)/float64(vmStat.Total),
		pt.Get("used_percent").(float64), "get point: %s", pt.Pretty())

	assert.Equal(t,
		100*float64(vmStat.Available)/float64(vmStat.Total),
		pt.Get("available_percent").(float64))

	assert.Equal(t, vmStat.Active, pt.Get("active").(uint64))
	assert.Equal(t, vmStat.Available, pt.Get("available").(uint64))
	assert.Equal(t, vmStat.Buffers, pt.Get("buffered").(uint64))
	assert.Equal(t, vmStat.Cached, pt.Get("cached").(uint64))
	assert.Equal(t, vmStat.CommitLimit, pt.Get("commit_limit").(uint64))
	assert.Equal(t, vmStat.CommittedAS, pt.Get("committed_as").(uint64))
	assert.Equal(t, vmStat.Dirty, pt.Get("dirty").(uint64))
	assert.Equal(t, vmStat.Free, pt.Get("free").(uint64))
	assert.Equal(t, vmStat.HighFree, pt.Get("high_free").(uint64))
	assert.Equal(t, vmStat.HighTotal, pt.Get("high_total").(uint64))
	assert.Equal(t, vmStat.HugePageSize, pt.Get("huge_pages_size").(uint64))
	assert.Equal(t, vmStat.HugePagesFree, pt.Get("huge_pages_free").(uint64))
	assert.Equal(t, vmStat.HugePagesTotal, pt.Get("huge_pages_total").(uint64))
	assert.Equal(t, vmStat.Inactive, pt.Get("inactive").(uint64))
	assert.Equal(t, vmStat.LowFree, pt.Get("low_free").(uint64))
	assert.Equal(t, vmStat.LowTotal, pt.Get("low_total").(uint64))
	assert.Equal(t, vmStat.Mapped, pt.Get("mapped").(uint64))
	assert.Equal(t, vmStat.PageTables, pt.Get("page_tables").(uint64))
	assert.Equal(t, vmStat.Shared, pt.Get("shared").(uint64))
	assert.Equal(t, vmStat.Slab, pt.Get("slab").(uint64))
	assert.Equal(t, vmStat.SReclaimable, pt.Get("sreclaimable").(uint64))
	assert.Equal(t, vmStat.SUnreclaim, pt.Get("sunreclaim").(uint64))
	assert.Equal(t, vmStat.SwapCached, pt.Get("swap_cached").(uint64))
	assert.Equal(t, vmStat.SwapFree, pt.Get("swap_free").(uint64))
	assert.Equal(t, vmStat.SwapTotal, pt.Get("swap_total").(uint64))
	assert.Equal(t, vmStat.Total, pt.Get("total").(uint64))
	assert.Equal(t, vmStat.Used, pt.Get("used").(uint64))
	assert.Equal(t, vmStat.VMallocChunk, pt.Get("vmalloc_chunk").(uint64))
	assert.Equal(t, vmStat.VMallocTotal, pt.Get("vmalloc_total").(uint64))
	assert.Equal(t, vmStat.VMallocUsed, pt.Get("vmalloc_used").(uint64))
	assert.Equal(t, vmStat.Writeback, pt.Get("write_back").(uint64))
	assert.Equal(t, vmStat.WritebackTmp, pt.Get("write_back_tmp").(uint64))

	// check if tags attached
	assert.Equal(t, "me", pt.Get("host").(string))
	assert.Equal(t, "v1", pt.Get("t1").(string))
}
