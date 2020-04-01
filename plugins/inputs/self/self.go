package self

import (
	"os"
	"runtime"
	"time"

	"github.com/influxdata/telegraf"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
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
	statMetric := s.stat.ToMetric()

	runnings := []string{}

	for _, input := range config.DKConfig.Inputs {
		if st, ok := input.Input.(internal.PluginStat); ok {
			if st.IsRunning() {
				runnings = append(runnings, input.Config.Name)
			}
			m := st.StatMetric()
			if m != nil {
				m.AddTag("datakit", config.DKConfig.MainCfg.UUID)
				acc.AddMetric(m)
			}
		}
	}

	//statMetric.AddField("running_inputs", strings.Join(runnings, ","))
	acc.AddMetric(statMetric)
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
