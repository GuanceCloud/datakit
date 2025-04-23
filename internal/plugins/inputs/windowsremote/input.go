// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package windowsremote collect Windows WMI and SNMP metrics
package windowsremote

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type Input struct {
	Election  bool   `toml:"election"`
	WorkerNum int    `toml:"worker_num"`
	Mode      string `toml:"mode"`

	IPList       []string      `toml:"ip_list"`
	CIDRs        []string      `toml:"cidrs"`
	ScanInterval time.Duration `toml:"scan_interval"`

	Wmi  *WmiConfig        `toml:"wmi"`
	Snmp *SnmpConfig       `toml:"snmp"`
	Tags map[string]string `toml:"tags"`

	instance    RemoteInstance
	targetPorts []int
	protocol    string

	feeder  dkio.Feeder
	chPause chan bool
	pause   bool
}

type WmiConfig struct {
	Port      int      `toml:"port"`
	LogEnable bool     `toml:"log_enable"`
	Auth      *WmiAuth `toml:"auth"`
	extraTags map[string]string
}

type WmiAuth struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type SnmpConfig struct {
	Ports     []int  `toml:"ports"`
	Community string `toml:"community"`
	extraTags map[string]string
	// Version   string `toml:"version"` // only supported v2c
}

func (*Input) SampleConfig() string                    { return sampleCfg }
func (*Input) Catalog() string                         { return inputName }
func (*Input) AvailableArchs() []string                { return []string{datakit.OSLabelWindows} }
func (*Input) Singleton()                              { /*nil*/ }
func (*Input) SampleMeasurement() []inputs.Measurement { return nil /* no measurement docs exported */ }
func (*Input) Terminate()                              { /* TODO */ }

func (ipt *Input) Run() {
	l = logger.SLogger("windows_remote")

	l.Info("windows_remote input start")
	if err := ipt.setup(); err != nil {
		l.Warnf("setup err: %s", err)
	}

	ipt.startCollect()
	l.Info("windows_remote input exit")
}

func (ipt *Input) setup() error {
	if ipt.WorkerNum <= 0 {
		ipt.WorkerNum = datakit.AvailableCPUs*2 + 1
	}
	if ipt.ScanInterval <= 0 {
		ipt.ScanInterval = time.Minute * 10
	}

	switch ipt.Mode {
	case "wmi":
		if ipt.Wmi == nil {
			return fmt.Errorf("invalid WMI config: %#v", ipt.Wmi)
		}
		ipt.Wmi.extraTags = ipt.Tags
		ipt.targetPorts = []int{ipt.Wmi.Port}
		ipt.protocol = "tcp"
		ipt.instance = newWmi(ipt.Wmi)
	case "snmp":
		if ipt.Snmp == nil {
			return fmt.Errorf("invalid SNMP config: %#v", ipt.Snmp)
		}
		ipt.Snmp.extraTags = ipt.Tags
		ipt.targetPorts = ipt.Snmp.Ports
		ipt.protocol = "udp"
		ipt.instance = newSnmp(ipt.Snmp)

	default:
		return fmt.Errorf("unexpected mode %s, only supported wmi and snmp", ipt.Mode)
	}
	return nil
}

func (ipt *Input) isPause() bool {
	return ipt.Election && ipt.pause
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	select {
	case ipt.chPause <- true:
		l.Info("pause %s paused", inputName)
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	select {
	case ipt.chPause <- false:
		l.Infof("resume %s false", inputName)
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func init() { //nolint:gochecknoinits
	setupMetrics()
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Tags:     map[string]string{},
			instance: nil,
			feeder:   dkio.DefaultFeeder(),
			chPause:  make(chan bool, inputs.ElectionPauseChannelLength),
			Election: true,
		}
	})
}
