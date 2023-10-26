// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package confd

import (
	"fmt"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/cpu"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/dk"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ipmi"
)

type args struct {
	data []map[string]string
}
type want struct {
	isHaveKey     bool
	mapKey        string
	isSliceNil    bool
	sliceLen      int
	sliceIdx      int
	inputInterval time.Duration
}

func checkGot(t *testing.T, confdInputs map[string][]*inputs.ConfdInfo, wants []want) {
	t.Helper()
	for _, want := range wants {
		if want.isHaveKey == false {
			if _, ok := confdInputs[want.mapKey]; ok {
				t.Errorf("want no mapKey : %s, but got.", want.mapKey)
			}
			continue
		}

		if _, ok := confdInputs[want.mapKey]; !ok {
			t.Errorf("want have mapKey : %s, but not got.", want.mapKey)
		}

		if want.isSliceNil == true {
			if len(confdInputs[want.mapKey]) > 0 {
				t.Errorf("want a nil slice, mapKey : %s, but got %d data.", want.mapKey, len(confdInputs[want.mapKey]))
			}
			continue
		}

		if len(confdInputs[want.mapKey]) != want.sliceLen {
			t.Errorf("want slice len : %d, mapKey : %s, but got %d data.", want.sliceLen, want.mapKey, len(confdInputs[want.mapKey]))
		}

		duration := time.Duration(1)
		switch want.mapKey {
		case "cpu":
			duration = confdInputs[want.mapKey][want.sliceIdx].Input.(*cpu.Input).Interval
		case "ipmi":
			duration = confdInputs[want.mapKey][want.sliceIdx].Input.(*ipmi.Input).Interval
		case "dk":
			duration = confdInputs[want.mapKey][want.sliceIdx].Input.(*dk.Input).Interval
		}
		if duration != want.inputInterval {
			t.Errorf("want Interval.Duration : %d, mapKey : %s, but got %d .", want.inputInterval, want.mapKey, duration)
		}
	}
}

func Test_handleConfdData(t *testing.T) {
	tests := []struct {
		name  string
		args  args
		wants []want
	}{
		{
			// will got 1 "cpu".
			// not got blank "mem", then will not delete "mem" input.
			// got blank "ipmi", then will delete existing input.
			name: "1-cpu",
			args: args{
				data: []map[string]string{{"any": `
[[inputs.cpu]]
  interval = '11s'
						`}},
			},
			wants: []want{
				{
					isHaveKey:     true,
					mapKey:        "cpu",
					isSliceNil:    false,
					sliceLen:      1,
					sliceIdx:      0,
					inputInterval: time.Duration(11000000000),
				},
				{
					isHaveKey: false,
					mapKey:    "mem",
				},
				{
					isHaveKey:  true,
					mapKey:     "ipmi",
					isSliceNil: true,
				},
			},
		},
		{
			// to test will got only 1 "cpu".
			name: "2-cpu",
			args: args{
				data: []map[string]string{{"any": `
[[inputs.cpu]]
  interval = '12s'
[[inputs.cpu]]
  interval = '13s'
				`}},
			},
			wants: []want{
				{
					isHaveKey:     true,
					mapKey:        "cpu",
					isSliceNil:    false,
					sliceLen:      1,
					sliceIdx:      0,
					inputInterval: time.Duration(12000000000),
				},
			},
		},
		{
			// to test will got 2 "ipmi".
			name: "2-ipmi",
			args: args{
				data: []map[string]string{{"any": `
[[inputs.ipmi]]
  interval = '12s'
[[inputs.ipmi]]
  interval = '13s'
				`}},
			},
			wants: []want{
				{
					isHaveKey:     true,
					mapKey:        "ipmi",
					isSliceNil:    false,
					sliceLen:      2,
					sliceIdx:      0,
					inputInterval: time.Duration(12000000000),
				},
				{
					isHaveKey:     true,
					mapKey:        "ipmi",
					isSliceNil:    false,
					sliceLen:      2,
					sliceIdx:      1,
					inputInterval: time.Duration(13000000000),
				},
			},
		},
		{
			// to test will not got "dk", then will not modify "dk" input.
			name: "1-dk",
			args: args{
				data: []map[string]string{{"any": `
[[inputs.dk]]
  interval = '12s'
				`}},
			},
			wants: []want{
				{
					isHaveKey: false,
					mapKey:    "dk",
				},
			},
		},
	}

	// cpu & mem inputs be default.
	config.Cfg.DefaultEnabledInputs = []string{"cpu", "mem"}
	// Existing inputs, "cpu" & "mem" be singleton, "dk" can't modify, "ipmi" be other.
	inputs.Inputs = map[string]inputs.Creator{
		"cpu": func() inputs.Input {
			return &cpu.Input{
				Interval:          time.Second * 10,
				EnableTemperature: true,
				Tags:              make(map[string]string),
			}
		},
		"dk": func() inputs.Input {
			return &dk.Input{
				Interval: time.Second * 10,
				Tags:     make(map[string]string),
			}
		},
		"ipmi": func() inputs.Input {
			return &ipmi.Input{
				Interval: time.Second * 10,
				Tags:     make(map[string]string),
			}
		},
	}

	// Add all inputs kind. (not all, only test.)
	inputs.AddInput("cpu", nil)
	inputs.AddInput("ipmi", nil)
	inputs.AddInput("dk", nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confdInputs = make(map[string][]*inputs.ConfdInfo)
			handleConfdData(tt.args.data)
			checkGot(t, confdInputs, tt.wants)
			fmt.Println(confdInputs)
		})
	}
}
