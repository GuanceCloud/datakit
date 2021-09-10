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
	f := func(pos int, mult time.Duration) {
		if m[pos] == "" {
			return
		}

		n, _ := strconv.Atoi(m[pos])
		d := time.Duration(n)
		du += d * mult
	}

	f(2, 60*60*24*365*time.Second) // y
	f(4, 60*60*24*7*time.Second)   // w
	f(6, 60*60*24*time.Second)     // d
	f(8, 60*60*time.Second)        // h
	f(10, 60*time.Second)          // m
	f(12, time.Second)             // s
	f(14, time.Millisecond)        // ms
	f(16, time.Microsecond)        // us
	f(18, time.Nanosecond)         // ns
	return du, nil
}
