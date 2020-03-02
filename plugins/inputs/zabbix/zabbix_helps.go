package zabbix

import (
	"time"
	"strings"
	"strconv"
)

const (
	nanosPerSecond     = int64(time.Second / time.Nanosecond)
	nanosPerMillisecond = int64(time.Millisecond / time.Nanosecond)
)

// Add a right pad
func RightPad(s string, padStr string, pLen int) string {
	if pLen > 0 {
		return s + strings.Repeat(padStr, pLen)
	}
	return s
}

// Convert milliseconds to Unix time
func NsToTime(ns string) (time.Time, error) {
	nsInt, err := strconv.ParseInt(ns, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(nsInt/nanosPerSecond,
		nsInt%nanosPerSecond), nil
}
