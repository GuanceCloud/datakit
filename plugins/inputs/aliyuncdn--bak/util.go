package aliyuncdn

import (
	"strconv"
	"time"
)

func DateISO8601(tt int64) string {
	return time.Unix(tt, 0).Format("2006-01-02T15:04:05Z07:00")
}

func RFC3339(tt string) time.Time {
	const layout = time.RFC3339
	tm, _ := time.Parse(layout, tt)
	return tm
}

func ConvertToNum(str string) int {
	num, _ := strconv.Atoi(str)

	return num
}

func ConvertToFloat(str string) float64 {
	value, _ := strconv.ParseFloat(str, 64)
	return value
}
