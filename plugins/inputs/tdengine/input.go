// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package tdengine is input for tdengine database
package tdengine

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	_ inputs.InputV2       = &Input{}
	_ inputs.Input         = &Input{}
	_ inputs.ElectionInput = &Input{}
)

var (
	l         = logger.DefaultSLogger("tdengine")
	inputName = "tdengine"
)

const (
	sampleConfig = `
[[inputs.tdengine]]
  ## adapter config (Required)
  adapter_endpoint = "<http://taosadapter.test.com>"
  user = "<userName>"
  password = "<pw>"

  ## add tag (optional)
  [inputs.tdengine.tags]
	## Different clusters can be distinguished by tag. Such as testing,product,local ,default is 'testing'
	## cluster_name = "testing"

    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)

type Input struct {
	AdapterEndpoint string            `toml:"adapter_endpoint"` // Adapter 访问地址，多个用逗号隔开
	User            string            `toml:"user"`
	Password        string            `toml:"password"`
	Tags            map[string]string `toml:"tags"`

	tdengine  *tdEngine
	inputName string

	pauseCh chan bool
	semStop *cliutils.Sem // start stop signal
}

func (i *Input) Catalog() string {
	return inputName
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (i *Input) SampleConfig() string {
	return sampleConfig
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&Measurement{}}
}

func (i *Input) exit() {
	i.tdengine.Stop()
}

func (i *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case i.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case i.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)

	i.tdengine = newTDEngine(i.User, i.Password, i.AdapterEndpoint)

	globalTags = i.Tags
	// 1 checkHealth: show databases
	if !i.tdengine.CheckHealth(checkHealthSQL) {
		return
	}

	go i.tdengine.run()
	l.Infof("TDEngine input started")

	for {
		select {
		case <-datakit.Exit.Wait():
			i.exit()
			l.Infof("%s exit", i.inputName)
			return
		case <-i.semStop.Wait():
			i.exit()
			l.Infof("%s return", i.inputName)
			return
		case f := <-i.pauseCh:
			i.tdengine.upstream = !f
		}
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Tags:      map[string]string{},
			inputName: inputName,
			pauseCh:   make(chan bool, inputs.ElectionPauseChannelLength),
			semStop:   cliutils.NewSem(),
		}
	})
}
