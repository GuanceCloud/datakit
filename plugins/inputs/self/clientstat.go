package self

import (
	"runtime"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

var (
	StartTime time.Time
)

type ClientStat struct {
	HostName string

	PID    int
	Uptime int64
	OS     string
	Arch   string

	NumGoroutines int
	HeapAlloc     int64
	HeapSys       int64
	HeapObjects   int64
}

func (s *ClientStat) Update() {
	s.HostName = datakit.Cfg.MainCfg.Hostname

	var memStatus runtime.MemStats
	runtime.ReadMemStats(&memStatus)

	s.NumGoroutines = int(runtime.NumGoroutine())
	s.HeapAlloc = int64(memStatus.HeapAlloc)
	s.HeapSys = int64(memStatus.HeapSys)
	s.HeapObjects = int64(memStatus.HeapObjects)
}

func (s *ClientStat) ToMetric() *influxdb.Point {

	s.Uptime = int64(time.Now().Sub(StartTime) / time.Second)

	measurement := "datakit"

	tags := map[string]string{
		"uuid":    datakit.Cfg.MainCfg.UUID,
		"vserion": git.Version,
		"os":      s.OS,
		"arch":    s.Arch,
	}

	fields := map[string]interface{}{
		"hostname": s.HostName,
		"pid":      s.PID,
		"uptime":   s.Uptime,

		"num_goroutines": s.NumGoroutines,
		"heap_alloc":     s.HeapAlloc,
		"heap_sys":       s.HeapSys,
		"heap_objects":   s.HeapObjects,
	}

	m, _ := influxdb.NewPoint(measurement, tags, fields, time.Now().UTC())

	return m
}
