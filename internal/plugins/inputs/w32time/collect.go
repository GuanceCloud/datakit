// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package w32time

import (
	"fmt"

	"github.com/GuanceCloud/cliutils/point"
)

const w32TimeServiceName = "W32Time"

type (
	serviceStatusFunc  func() (bool, error)
	counterCollectFunc func() (map[string]float64, error)
)

func (ipt *Input) collectPoint(
	ptTS int64,
	serviceStatus serviceStatusFunc,
	collectCounters counterCollectFunc,
) (*point.Point, error) {
	running, err := serviceStatus()
	if err != nil {
		return nil, fmt.Errorf("query %s service: %w", w32TimeServiceName, err)
	}

	fields := map[string]interface{}{
		fieldServiceRunning: int64(0),
	}
	if running {
		fields[fieldServiceRunning] = int64(1)
	}

	var collectErr error
	if running {
		raw, err := collectCounters()
		for key, value := range normalizeCounterValues(raw) {
			fields[key] = value
		}
		collectErr = err
	}

	kvs := point.NewKVs(fields)
	for key, value := range ipt.mergedTags {
		kvs = kvs.AddTag(key, value)
	}

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(ptTS))

	pt := point.NewPoint(metricName, kvs, opts...)
	return pt, collectErr
}
