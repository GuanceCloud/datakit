// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package process

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils"
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

// go test -run=^$ -bench=. -cpuprofile cpupprof.out
func BenchmarkWriteObject(b *testing.B) {
	in := &Input{}
	processList := in.getProcesses(false)
	procRecorder := newProcRecorder()
	tn := time.Now().UTC()
	in.WriteObject(processList, procRecorder, tn)
}
