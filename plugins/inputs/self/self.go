package self

import (
	"os"
	"runtime"
	"time"

	"github.com/influxdata/telegraf"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type SelfInfo struct {
	stat *ClientStat
}

func getinfo() {

	var memStatus runtime.MemStats
	runtime.ReadMemStats(&memStatus)

	numGo := runtime.NumGoroutine
	_ = numGo

	//pprof.StartCPUProfile()
}

func (_ *SelfInfo) SampleConfig() string {
	return `Interval = '1m'`
}

func (_ *SelfInfo) Description() string {
	return ""
}

func (s *SelfInfo) Gather(acc telegraf.Accumulator) error {
	s.stat.Update()
	acc.AddMetric(s.stat.ToMetric())
	return nil
}

func init() {
	StartTime = time.Now()
	inputs.Add("self", func() telegraf.Input {
		return &SelfInfo{
			stat: &ClientStat{
				OS:   runtime.GOOS,
				Arch: runtime.GOARCH,
				PID:  os.Getpid(),
			},
		}
	})
}
