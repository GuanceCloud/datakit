// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package time wraps time related functions
package time

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var (
	// MinUnixSeconds 10_0984_3200.
	MinUnixSeconds = time.Date(2002, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

	// MinUnixMilli 10_0984_3200_000.
	MinUnixMilli = MinUnixSeconds * 1000

	// MinUnixMicro 10_09843200_000_000.
	MinUnixMicro = MinUnixMilli * 1000

	// MinUnixNano 10_0984_3200_000_000_000.
	MinUnixNano = MinUnixMicro * 1000

	// MaxUnixSeconds 72_8965_4399.
	MaxUnixSeconds = time.Date(2200, 12, 31, 23, 59, 59, 0, time.UTC).Unix()

	// MaxUnixMilli 72_8965_4399_000.
	MaxUnixMilli = MaxUnixSeconds * 1000

	// MaxUnixMicro 72_8965_4399_000_000.
	MaxUnixMicro = MaxUnixMilli * 1000

	// MaxUnixNano 72_8965_4399_000_000_000.
	MaxUnixNano = MaxUnixMicro * 1000 //nolint:unused
)

//nolint:lll
var durationRE = regexp.MustCompile("^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?(([0-9]+)us)?(([0-9]+)ns)?$")

// ParseDuration 支持更多时间单位的解析
//
//nolint:gomnd
func ParseDuration(s string) (time.Duration, error) {
	switch s {
	case "0":
		return 0, nil
	case "":
		return 0, fmt.Errorf("empty duration string")
	}

	m := durationRE.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf("invalid duration string: %q", s)
	}

	var du time.Duration
	f := func(pos int, mult time.Duration) error {
		if m[pos] == "" {
			return nil
		}

		n, err := strconv.Atoi(m[pos])
		if err != nil {
			return err
		}

		d := time.Duration(n)
		du += d * mult //nolint:durationcheck

		return nil
	}

	if err := f(2, 60*60*24*365*time.Second); err != nil {
		return du, err
	} // y

	if err := f(4, 60*60*24*7*time.Second); err != nil {
		return du, err
	} // w

	if err := f(6, 60*60*24*time.Second); err != nil {
		return du, err
	} // d

	if err := f(8, 60*60*time.Second); err != nil {
		return du, err
	} // h

	if err := f(10, 60*time.Second); err != nil {
		return du, err
	} // m

	if err := f(12, time.Second); err != nil {
		return du, err
	} // s

	if err := f(14, time.Millisecond); err != nil {
		return du, err
	} // ms

	if err := f(16, time.Microsecond); err != nil {
		return du, err
	} // us

	if err := f(18, time.Nanosecond); err != nil {
		return du, err
	} // ns

	return du, nil
}

func UnixStampToTime(ts int64, loc *time.Location) time.Time {
	switch {
	case ts <= MaxUnixSeconds:
		return time.Unix(ts, 0).In(loc)
	case ts <= MaxUnixMilli:
		return time.UnixMilli(ts).In(loc)
	case ts <= MaxUnixMicro:
		return time.UnixMicro(ts).In(loc)
	default:
		return time.Unix(0, ts).In(loc)
	}
}
