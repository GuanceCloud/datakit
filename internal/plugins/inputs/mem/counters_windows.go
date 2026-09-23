// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows && amd64

package mem

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/win_utils/pdh"
)

func (ipt *Input) collectWindowsCounters(kvs point.KVs) point.KVs {
	if ipt.platform != "windows" {
		return kvs
	}
	var query pdh.PDH_HQUERY
	if ret := pdh.PdhOpenQuery(0, 0, &query); ret != 0 {
		l.Warnf("open PDH query: 0x%x", ret)
		return kvs
	}
	defer func() {
		if ret := pdh.PdhCloseQuery(query); ret != 0 {
			l.Warnf("close PDH query: 0x%x", ret)
		}
	}()

	counters := []struct {
		field  string
		path   string
		handle pdh.PDH_HCOUNTER
	}{
		{field: "free_system_page_table_entries", path: `\Memory\Free System Page Table Entries`},
		{field: "pages_total", path: `\Memory\Pages/sec`},
	}
	added := 0
	for i := range counters {
		counter := &counters[i]
		if ret := pdh.PdhAddEnglishCounter(query, counter.path, 0, &counter.handle); ret != 0 {
			l.Warnf("add PDH counter %q: 0x%x", counter.path, ret)
			counter.handle = 0
			continue
		}
		added++
	}
	if added == 0 {
		return kvs
	}
	if ret := pdh.PdhCollectQueryData(query); ret != 0 {
		l.Warnf("collect memory PDH counters: 0x%x", ret)
		return kvs
	}
	for _, counter := range counters {
		if counter.handle == 0 {
			continue
		}
		// Read the provider's raw count. In particular, Pages/sec is not formatted
		// into a rate here; the backend computes rate/irate from pages_total.
		var value pdh.PDH_RAW_COUNTER
		if ret := pdh.PdhGetRawCounterValue(counter.handle, nil, &value); ret != 0 {
			l.Warnf("read PDH counter %q: 0x%x", counter.path, ret)
			continue
		}
		if value.CStatus != pdh.PDH_CSTATUS_VALID_DATA && value.CStatus != pdh.PDH_CSTATUS_NEW_DATA {
			l.Warnf("PDH counter %q status: 0x%x", counter.path, value.CStatus)
			continue
		}
		kvs = kvs.Set(counter.field, value.FirstValue)
	}
	return kvs
}
