// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/recorder"
)

// IOOption used to add various options to setup io module.
type IOOption func(x *dkIO)

// WithFilter disble consumer on IO feed.
func WithFilter(on bool) IOOption {
	return func(x *dkIO) {
		x.withFilter = on
	}
}

// WithRecorder setup record config for data points.
func WithRecorder(r *recorder.Recorder) IOOption {
	return func(x *dkIO) {
		if r != nil && r.Enabled {
			if r, err := recorder.SetupRecorder(r); err != nil {
				log.Warnf("invalid recorder: %s, ignored", err)
			} else {
				x.recorder = r
			}
		}
	}
}

// WithCompactor disble consumer on IO feed.
func WithCompactor(on bool) IOOption {
	return func(x *dkIO) {
		x.withCompactor = on
	}
}

// WithFeederOutputer used to set the output of feeder.
func WithFeederOutputer(fo FeederOutputer) IOOption {
	return func(x *dkIO) {
		x.fo = fo
	}
}

// WithDataway used to setup where data write to(dataway).
func WithDataway(dw dataway.IDataway) IOOption {
	return func(x *dkIO) {
		x.dw = dw
	}
}

// WithFilters used to setup point filter.
func WithFilters(filters map[string]filter.FilterConditions) IOOption {
	return func(x *dkIO) {
		if len(filters) > 0 {
			x.filters = filters
		}
	}
}

// WithCompactWorkers set IO flush workers.
func WithCompactWorkers(n int) IOOption {
	return func(x *dkIO) {
		if n > 0 {
			x.flushWorkers = n
		}
	}
}

// WithCompactInterval used to contol when to flush cached data.
func WithCompactInterval(d time.Duration) IOOption {
	return func(x *dkIO) {
		if int64(d) > 0 {
			x.flushInterval = d
		}
	}
}

// WithCompactAt used to set max cache size.
// The count used to control when to send the cached data.
func WithCompactAt(count int) IOOption {
	return func(x *dkIO) {
		if count > 0 {
			log.Debugf("set max cache count to %d", count)
			x.compactAt = count
		}
	}
}

// WithAvailableCPUs used to set concurrent uploader worker numbers.
func WithAvailableCPUs(cpus int) IOOption {
	return func(x *dkIO) {
		if cpus > 0 {
			x.availableCPUs = cpus
		}
	}
}
