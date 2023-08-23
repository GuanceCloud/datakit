// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package time

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

// Duration just wraps time.Duration.
type Duration struct {
	Duration time.Duration
}

// UnmarshalTOML parses the duration from the TOML config file.
func (d *Duration) UnmarshalTOML(b []byte) error {
	b = bytes.Trim(b, "'")

	// see if we can directly convert it
	if du, err := time.ParseDuration(string(b)); err == nil {
		d.Duration = du
		return nil
	}

	// Parse string duration, ie, "1s"
	if uq, err := strconv.Unquote(string(b)); err == nil && len(uq) > 0 {
		d.Duration, err = time.ParseDuration(uq)
		if err == nil {
			return nil
		}
	}

	// First try parsing as integer seconds
	if sI, err := strconv.ParseInt(string(b), 10, 64); err == nil {
		d.Duration = time.Second * time.Duration(sI)
		return nil
	}
	// Second try parsing as float seconds
	if sF, err := strconv.ParseFloat(string(b), 64); err == nil {
		d.Duration = time.Second * time.Duration(sF)
	} else {
		return err
	}

	return nil
}

func (d *Duration) UnitString(unit time.Duration) string {
	ts := fmt.Sprintf("%d", d.Duration/unit)
	switch unit {
	case time.Second:
		return ts + "s"
	case time.Millisecond:
		return ts + "ms"
	case time.Microsecond:
		return ts + "mics"
	case time.Minute:
		return ts + "m"
	case time.Hour:
		return ts + "h"
	case time.Nanosecond:
		return ts + "ns"
	default:
		return ts + "unknow"
	}
}

// Cost return time elapsed since start.
//
// You should call it within defer calling:
//
//		defer Cost(time.Now(), func(du time.Duration) {
//	   // do something on du...
//	 })
func Cost(start time.Time, f func(du time.Duration)) {
	if f != nil {
		f(time.Since(start))
	}
}
