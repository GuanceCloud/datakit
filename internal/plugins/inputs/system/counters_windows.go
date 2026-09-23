// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows && amd64

package system

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/win_utils/pdh"
)

func (ipt *Input) collectWindowsCounters(kvs point.KVs) point.KVs {
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

	const path = `\System\Processor Queue Length`
	var handle pdh.PDH_HCOUNTER
	if ret := pdh.PdhAddEnglishCounter(query, path, 0, &handle); ret != 0 {
		l.Warnf("add PDH counter %q: 0x%x", path, ret)
		return kvs
	}
	if ret := pdh.PdhCollectQueryData(query); ret != 0 {
		l.Warnf("collect PDH counter %q: 0x%x", path, ret)
		return kvs
	}
	var value pdh.PDH_RAW_COUNTER
	if ret := pdh.PdhGetRawCounterValue(handle, nil, &value); ret != 0 {
		l.Warnf("read PDH counter %q: 0x%x", path, ret)
		return kvs
	}
	if value.CStatus != pdh.PDH_CSTATUS_VALID_DATA && value.CStatus != pdh.PDH_CSTATUS_NEW_DATA {
		l.Warnf("PDH counter %q status: 0x%x", path, value.CStatus)
		return kvs
	}
	return kvs.Set("processor_queue_length", value.FirstValue)
}
