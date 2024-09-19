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
	t.Run("10s", func(t *T.T) {
		m := &mockNSec{
			n: 10,
		}
		StartNTP(m, time.Minute, 1)

		time.Sleep(time.Second) // wait worker ok

		local := LocalTime()
		ntpTime := NTPTime()

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
		ntpTime := NTPTime()

		assert.Equalf(t, int64(-10), ntpTime.Unix()-local.Unix(), "local: %d, ntp: %d", local.Unix(), ntpTime.Unix())
		t.Logf("local: %d, ntp: %d", local.Unix(), ntpTime.Unix())
	})
}
