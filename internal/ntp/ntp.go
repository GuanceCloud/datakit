// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package ntp sync network time.
package ntp

import (
	"context"
	"math"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	localTimeSecDiff atomic.Int64
	l                = logger.DefaultSLogger("ntp")
)

type syncer interface {
	TimeDiff() int64
}

func doSync(diffSec int64, abs uint64) {
	if uint64(math.Abs(float64(diffSec))) >= uint64(math.Abs(float64(abs))) {
		localTimeSecDiff.Store(diffSec)
		ntpSyncCount.Add(1)
		l.Infof("update local time diff %s", time.Duration(localTimeSecDiff.Load())*time.Second)
	}

	ntpSyncSummary.Observe(float64(diffSec))
}

// StartNTP start sync on dataway's timestamp.
func StartNTP(s syncer, syncInterval time.Duration, diffAbsRangeSecond uint64) {
	g := datakit.G("ntp")
	l = logger.SLogger("ntp")

	if syncInterval <= time.Minute {
		l.Warnf("invalid syncInterval(%v), reset to %v", syncInterval, time.Minute)
		syncInterval = time.Minute
	}

	if diffAbsRangeSecond < 5 {
		l.Warnf("invalid diffAbsRangeSecond(%d), reset to %d seconds", diffAbsRangeSecond, 5)
		diffAbsRangeSecond = 5
	}

	g.Go(func(_ context.Context) error {
		tick := time.NewTicker(syncInterval)
		defer tick.Stop()

		for {
			// sync ASAP
			doSync(s.TimeDiff(), diffAbsRangeSecond)

			select {
			case <-tick.C:
			case <-datakit.Exit.Wait():
				l.Infof("ntp exit")
				return nil
			}
		}
	})
}

// Now get synced network time.
func Now() time.Time {
	local := time.Now()

	// if ntp time > local time, then localTimeSecDiff > 0, so add the difference.
	// if ntp time < local time, localTimeSecDiff < 0, the minus the difference.
	return local.Add(time.Duration(localTimeSecDiff.Load()) * time.Second)
}

// CalibrateTime adjust time with NTP time.
func CalibrateTime(t time.Time) time.Time {
	return t.Add(time.Duration(localTimeSecDiff.Load()) * time.Second)
}

// LocalTime get local machine time.
func LocalTime() time.Time {
	return time.Now()
}
