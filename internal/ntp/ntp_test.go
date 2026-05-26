// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2022-present Guance, Inc.

// Package ntp sync network time.
package ntp

import (
	T "testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockNSec struct {
	n int64
}

func (s *mockNSec) TimeDiff() int64 {
	return s.n // always with n sec diff
}

func TestNTPTime(t *T.T) {
	t.Cleanup(func() {
		localTimeSecDiff.Store(0)
	})

	t.Run("10s", func(t *T.T) {
		m := &mockNSec{
			n: 10,
		}
		StartNTP(m, time.Minute, 1)

		time.Sleep(time.Second) // wait worker ok

		local := LocalTime()
		ntpTime := Now()

		assert.Equalf(t, int64(10), ntpTime.Unix()-local.Unix(), "local: %d, ntp: %d", local.Unix(), ntpTime.Unix())
		t.Logf("local: %d, ntp: %d", local.Unix(), ntpTime.Unix())
	})

	t.Run("-10s", func(t *T.T) {
		m := &mockNSec{
			n: -10,
		}
		StartNTP(m, time.Minute, 1)

		time.Sleep(time.Second) // wait worker ok

		local := LocalTime()
		ntpTime := Now()

		assert.Equalf(t, int64(-10), ntpTime.Unix()-local.Unix(), "local: %d, ntp: %d", local.Unix(), ntpTime.Unix())
		t.Logf("local: %d, ntp: %d", local.Unix(), ntpTime.Unix())
	})
}

func TestDoSyncClearsRecoveredDiff(t *T.T) {
	localTimeSecDiff.Store(0)

	doSync(8*60*60+1, 30)
	assert.Equal(t, int64(8*60*60+1), localTimeSecDiff.Load())

	doSync(0, 30)
	assert.Zero(t, localTimeSecDiff.Load())

	local := LocalTime()
	ntpTime := Now()
	assert.LessOrEqual(t, abs64(ntpTime.Unix()-local.Unix()), int64(1))
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}

	return v
}
