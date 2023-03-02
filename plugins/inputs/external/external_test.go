// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows

package external

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestInput(t *testing.T) {
	interval := ".1s"

	cases := []struct {
		name   string
		inputs []*ExternalInput
		notify chan interface{}
	}{
		{
			name:   "non-daemon-input-election-for-3-instance",
			notify: make(chan interface{}),
			inputs: []*ExternalInput{
				{
					Name:           "ipt1-instance1",
					Election:       true,
					Interval:       interval,
					Cmd:            "echo",
					Args:           []string{"ipt1-instance1"},
					semStop:        cliutils.NewSem(),
					semStopProcess: cliutils.NewSem(),
					pauseCh:        make(chan bool, inputs.ElectionPauseChannelLength),
				},

				{
					Name:           "ipt-instance2",
					Interval:       interval,
					Election:       true,
					Cmd:            "echo",
					Args:           []string{"ipt-instance2"},
					semStop:        cliutils.NewSem(),
					semStopProcess: cliutils.NewSem(),
					pauseCh:        make(chan bool, inputs.ElectionPauseChannelLength),
				},

				{
					Name:           "ipt-instance3",
					Interval:       interval,
					Election:       true,
					Cmd:            "echo",
					Args:           []string{"ipt-instance3"},
					semStop:        cliutils.NewSem(),
					semStopProcess: cliutils.NewSem(),
					pauseCh:        make(chan bool, inputs.ElectionPauseChannelLength),
				},
			},
		},
		{
			name:   "daemon-input-election-for-3-instance",
			notify: make(chan interface{}),
			inputs: []*ExternalInput{
				{
					Name:     "ipt1-instance1",
					Election: true,
					Daemon:   true,

					// sleep: run forever
					Cmd:  "sleep",
					Args: []string{"10000"},

					semStop:        cliutils.NewSem(),
					semStopProcess: cliutils.NewSem(),
					pauseCh:        make(chan bool, inputs.ElectionPauseChannelLength),
				},

				{
					Name:     "ipt-instance2",
					Daemon:   true,
					Election: true,

					// sleep: run forever
					Cmd:  "sleep",
					Args: []string{"10000"},

					semStop:        cliutils.NewSem(),
					semStopProcess: cliutils.NewSem(),
					pauseCh:        make(chan bool, inputs.ElectionPauseChannelLength),
				},

				{
					Name:     "ipt-instance3",
					Daemon:   true,
					Election: true,

					// sleep: run forever
					Cmd:  "sleep",
					Args: []string{"10000"},

					semStop:        cliutils.NewSem(),
					semStopProcess: cliutils.NewSem(),
					pauseCh:        make(chan bool, inputs.ElectionPauseChannelLength),
				},
			},
		},

		{
			name:   "mix-input-for-3-instance(no-election)",
			notify: make(chan interface{}),
			inputs: []*ExternalInput{
				{
					Name:           "ipt1-instance1",
					Interval:       interval,
					Cmd:            "echo",
					Args:           []string{"ipt1-instance1"},
					semStop:        cliutils.NewSem(),
					semStopProcess: cliutils.NewSem(),
					pauseCh:        make(chan bool, inputs.ElectionPauseChannelLength),
				},

				{
					Name:           "ipt-instance2",
					Daemon:         true,
					Cmd:            "echo",
					Args:           []string{"ipt-instance2"},
					semStop:        cliutils.NewSem(),
					semStopProcess: cliutils.NewSem(),
					pauseCh:        make(chan bool, inputs.ElectionPauseChannelLength),
				},

				{
					Name:           "ipt-instance3",
					Interval:       interval,
					Cmd:            "echo",
					Args:           []string{"ipt-instance3"},
					semStop:        cliutils.NewSem(),
					semStopProcess: cliutils.NewSem(),
					pauseCh:        make(chan bool, inputs.ElectionPauseChannelLength),
				},
			},
		},
	}

	wg := sync.WaitGroup{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// start input
			for i := range tc.inputs {
				ipt := tc.inputs[i]

				if err := ipt.precheck(); err != nil {
					t.Error(err)
					return
				}

				wg.Add(1)
				go func() {
					defer wg.Done()
					ipt.Run()
				}()
			}
		})

		// resume & pause & exit
		round := 0
		roundInterval := 1 * time.Second
		for {
			idx := rand.Int() % len(tc.inputs)
			for i := range tc.inputs {
				ipt := tc.inputs[i]
				if i == idx { // resume it
					fmt.Printf("resume %s\n", ipt.Name)
					ipt.Resume()
				} else {
					ipt.Pause()
				}
			}
			time.Sleep(roundInterval)
			round++
			if round >= 3 {
				fmt.Printf("terminat inputs...")
				for i := range tc.inputs {
					ipt := tc.inputs[i]
					ipt.Terminate()
				}
				break
			}
		}
	}

	wg.Wait()
}
