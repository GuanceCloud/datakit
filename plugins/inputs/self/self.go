package self

import (
	"os"
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	inputName = datakit.DatakitInputName
	l         = logger.DefaultSLogger(inputName)
)

type SelfInfo struct {
	stat *ClientStat

	semStop          *cliutils.Sem // start stop signal
	semStopCompleted *cliutils.Sem // stop completed signal
}

func (*SelfInfo) Catalog() string {
	return inputName
}

func (*SelfInfo) SampleConfig() string {
	return ``
}

func (si *SelfInfo) Run() {
	tick := time.NewTicker(time.Second * 10)
	defer tick.Stop()

	l = logger.SLogger(inputName)
	l.Info("self input started...")

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("self exit")
			return

		case <-si.semStop.Wait():
			l.Info("self return")

			if si.semStopCompleted != nil {
				si.semStopCompleted.Close()
			}
			return

		case <-tick.C:
			si.stat.Update()
			pt := si.stat.ToMetric()
			_ = io.Feed(inputName, datakit.Metric, []*io.Point{pt}, nil)
		}
	}
}

func (si *SelfInfo) Terminate() {
	if si.semStop != nil {
		si.semStop.Close()

		// wait stop completed
		if si.semStopCompleted != nil {
			for range si.semStopCompleted.Wait() {
				return
			}
		}
	}
}

func init() { //nolint:gochecknoinits
	StartTime = time.Now()
	inputs.Add(inputName, func() inputs.Input {
		return &SelfInfo{
			stat: &ClientStat{
				OS:   runtime.GOOS,
				Arch: runtime.GOARCH,
				PID:  os.Getpid(),
			},
			semStop:          cliutils.NewSem(),
			semStopCompleted: cliutils.NewSem(),
		}
	})
}
