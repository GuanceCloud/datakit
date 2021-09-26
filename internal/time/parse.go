package time

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

//nolint:lll
var durationRE = regexp.MustCompile("^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?(([0-9]+)us)?(([0-9]+)ns)?$")

// ParseDuration 支持更多时间单位的解析
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
