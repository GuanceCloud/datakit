package self

import (
	"os"
	"runtime"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
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
	HeapAlloc     uint64
	HeapSys       uint64
	HeapObjects   uint64
}

func (s *ClientStat) Update() {
	s.HostName, _ = os.Hostname()

	var memStatus runtime.MemStats
	runtime.ReadMemStats(&memStatus)

	s.NumGoroutines = runtime.NumGoroutine()
	s.HeapAlloc = memStatus.HeapAlloc
	s.HeapSys = memStatus.HeapSys
	s.HeapObjects = memStatus.HeapObjects
}

func (s *ClientStat) ToMetric() telegraf.Metric {

	s.Uptime = int64(time.Now().Sub(StartTime) / time.Second)

	measurement := "datakit"

	tags := map[string]string{
		"uuid":    config.DKConfig.MainCfg.UUID,
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

	m, _ := metric.New(
		measurement,
		tags,
		fields,
		time.Now(),
	)

	return m
}
