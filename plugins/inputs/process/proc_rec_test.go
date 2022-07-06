package process

import (
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func TestProcRec(t *testing.T) {
	rec := newProcRecorder()
	ipt := &Input{
		ObjectInterval: datakit.Duration{Duration: 1 * time.Millisecond},
		MetricInterval: datakit.Duration{Duration: 1 * time.Millisecond},
		semStop:        cliutils.NewSem(),
		Tags:           make(map[string]string),
	}

	procs := ipt.getProcesses(true)

	tn := time.Now()
	rec.flush(procs, tn)

	for _, v := range procs {
		t.Log(rec.calculatePercentTop(v, tn))
	}

	for i := 0; i < 10; i++ {
		time.Sleep(time.Millisecond * 2)

		procs := ipt.getProcesses(true)

		tn := time.Now()
		rec.flush(procs, tn)
		for _, v := range procs {
			t.Log(rec.calculatePercentTop(v, tn))
		}
	}
}
