package zabbix

import (
	"io/ioutil"
	"time"
)

func (z *ZabbixParam) getStartDate() time.Time {
	var err error
	var t time.Time

	content, err := ioutil.ReadFile(z.input.Registry)
	if err != nil {
		t, err = parseTimeStr(string(content))
		if err == nil {
			return t
		}
	}

	t, err = parseTimeStr(z.input.Startdate)
	if err == nil {
		return t
	}

	t, _ = parseTimeStr(defaultStartDate)
	return t
}

func (z *ZabbixParam) getStopDate(s time.Time) time.Time {
	var stopTime time.Time
	now := time.Now()

	stopTime = s.Add(time.Duration(z.input.Hoursperbatch) * time.Hour)
	if stopTime.After(now) {
		return now
	}
	return stopTime
}

func parseTimeStr(timeStr string) (time.Time, error) {
	startTime, err := time.Parse("2006-01-02T15:04:05", timeStr)
	if err != nil {
		startTime, err = time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return startTime, err
		}
	}
	return startTime, nil
}

func (z *ZabbixParam) updateRegistry(stop time.Time) {
	tStr := stop.Format("2006-01-02T15:04:05")
	ioutil.WriteFile(z.input.Registry, []byte(tStr), 0x666)
}
