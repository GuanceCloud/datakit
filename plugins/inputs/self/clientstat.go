package self

import (
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
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
	s.HostName = config.Cfg.Hostname

	var memStatus runtime.MemStats
	runtime.ReadMemStats(&memStatus)

	s.NumGoroutines = int(runtime.NumGoroutine())
	s.HeapAlloc = int64(memStatus.HeapAlloc)
	s.HeapSys = int64(memStatus.HeapSys)
	s.HeapObjects = int64(memStatus.HeapObjects)
}

func (s *ClientStat) ToMetric() *io.Point {

	s.Uptime = int64(time.Now().Sub(StartTime) / time.Second)

	measurement := "datakit"

	tags := map[string]string{
		"uuid":    config.Cfg.UUID,
		"vserion": git.Version,
		"os":      s.OS,
		"arch":    s.Arch,
		"host":    s.HostName,
	}

	fields := map[string]interface{}{
		"pid":    s.PID,
		"uptime": s.Uptime,

		"num_goroutines": s.NumGoroutines,
		"heap_alloc":     s.HeapAlloc,
		"heap_sys":       s.HeapSys,
		"heap_objects":   s.HeapObjects,
	}

	pt, err := io.MakePoint(measurement, tags, fields)
	if err != nil {
		l.Error(err)
	}

	return pt
}
