package aliyuncdn

import (
	"errors"
	"strconv"
	"time"
)

func unixTimeStrISO8601(t time.Time) string {
	_, zoff := t.Zone()
	nt := t.Add(-(time.Duration(zoff) * time.Second))
	s := nt.Format(`2006-01-02T15:04:05Z`)
	return s
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

func CheckCfg(cfg *CDN) error {
	if cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" {
		return errors.New("accessKeyID or accessKeySecret is null")
	}

	// 间隔大于5m
	return nil
}
