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

	PID    int
	Uptime int64
	OS     string
	Arch   string

	NumGoroutines int
	HeapAlloc     uint64
	HeapSys       uint64
	HeapObjects   uint64

	RunningInputs []string

	TotalGetMetrics    int64
	SuccessSendMetrics int64
	FailSendMetrics    int64
	DroppedMetrics     int64
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
		"uuid":    s.UUID,
		"vserion": s.Version,
		"os":      s.OS,
		"arch":    s.Arch,
	}

	fields := map[string]interface{}{
		"hostname":       s.HostName,
		"pid":            s.PID,
		"uptime":         s.Uptime,
		"running_inputs": strings.Join(s.RunningInputs, ","),

		"num_goroutines": s.NumGoroutines,
		"heap_alloc":     s.HeapAlloc,
		"heap_sys":       s.HeapSys,
		"heap_objects":   s.HeapObjects,

		"total_get_metrics":    s.TotalGetMetrics,
		"success_send_metrics": s.SuccessSendMetrics,
		"fail_send_metrics":    s.FailSendMetrics,
		"dropped_metrics":      s.DroppedMetrics,
	}

	running_inputs := ""
	if len(s.RunningInputs) > 0 {
		running_inputs = strings.Join(s.RunningInputs, ",")
	}
	fields["running_inputs"] = running_inputs

	m, _ := metric.New(
		measurement,
		tags,
		fields,
		time.Now(),
	)

	return m
}
