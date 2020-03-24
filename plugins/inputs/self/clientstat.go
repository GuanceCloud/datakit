package self

import (
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

var (
	StartTime time.Time
)

type ClientStat struct {
	UUID     string
	Version  string
	HostName string
	PID      int
	Uptime   int64
	OS       string
	Arch     string

	NumGoroutines int
	HeapAlloc     uint64
	HeapSys       uint64
	HeapObjects   uint64

	RunningInputs []string

	TotalGetMetrics    int64
	SuccessSendMetrics int64
	FailSendMetrics    int64
	DroppedMetrics     int64

	memStatus runtime.MemStats
}

func (s *ClientStat) Update() {
	s.NumGoroutines = runtime.NumGoroutine()
	s.HostName, _ = os.Hostname()
}

func (s *ClientStat) ToMetric() telegraf.Metric {

	s.Uptime = int64(time.Now().Sub(StartTime) / time.Second)

	measurement := "datakit_" + s.UUID

	tags := map[string]string{
		"uuid":    s.UUID,
		"vserion": s.Version,
		"os":      s.OS,
		"arch":    s.Arch,
	}

	fields := map[string]interface{}{
		"hostname":             s.HostName,
		"pid":                  s.PID,
		"uptime":               s.Uptime,
		"running_inputs":       strings.Join(s.RunningInputs, ","),
		"num_goroutines":       s.NumGoroutines,
		"total_get_metrics":    s.TotalGetMetrics,
		"success_send_metrics": s.SuccessSendMetrics,
		"fail_send_metrics":    s.FailSendMetrics,
		"dropped_metrics":      s.DroppedMetrics,
	}

	m, _ := metric.New(
		measurement,
		tags,
		fields,
		time.Now(),
	)

	return m
}
