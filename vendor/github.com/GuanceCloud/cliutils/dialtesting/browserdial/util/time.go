package util

import "time"

func NowISO() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func DurationUS(start time.Time) int64 {
	return time.Since(start).Microseconds()
}
