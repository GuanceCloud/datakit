package self

import (
	"os"
	"runtime"
	"time"

	"github.com/influxdata/telegraf"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	sampleConfig = `
interval = '10s'
`
)

type SelfInfo struct {
	stat *ClientStat
}

func (s *SelfInfo) Init() error {
	s.stat.UUID = config.DKUUID
	s.stat.Version = git.Version
	return nil
}

func (_ *SelfInfo) SampleConfig() string {
	return sampleConfig
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

	inputs.InternalInputsData["self"] = []byte(sampleConfig)
}
