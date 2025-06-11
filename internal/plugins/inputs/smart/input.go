// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package smart collects S.M.A.R.T metrics.
package smart

import (
	"os/exec"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	ipath "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const intelVID = "0x8086"

var (
	defSmartCmd     = "smartctl"
	defSmartCtlPath = "/usr/bin/smartctl"
	defNvmeCmd      = "nvme"
	defNvmePath     = "/usr/bin/nvme"
	defInterval     = datakit.Duration{Duration: 10 * time.Second}
	defTimeout      = datakit.Duration{Duration: 3 * time.Second}
)

var (
	inputName = "smart"
	//nolint:lll
	sampleConfig = `
[[inputs.smart]]
  ## The path to the smartctl executable
  # smartctl_path = "/usr/bin/smartctl"

  ## The path to the nvme-cli executable
  # path_nvme = "/usr/bin/nvme"

  ## Gathering interval
  # interval = "10s"

  ## Timeout for the cli command to complete.
  # timeout = "30s"

  ## Optionally specify if vendor specific attributes should be propagated for NVMe disk case
  ## ["auto-on"] - automatically find and enable additional vendor specific disk info
  ## ["vendor1", "vendor2", ...] - e.g. "Intel" enable additional Intel specific disk info
  # enable_extensions = ["auto-on"]

  ## On most platforms used cli utilities requires root access.
  ## Setting 'use_sudo' to true will make use of sudo to run smartctl or nvme-cli.
  ## Sudo must be configured to allow the telegraf user to run smartctl or nvme-cli
  ## without a password.
  # use_sudo = false

  ## Skip checking disks in this power mode. Defaults to "standby" to not wake up disks that have stopped rotating.
  ## See --nocheck in the man pages for smartctl.
  ## smartctl version 5.41 and 5.42 have faulty detection of power mode and might require changing this value to "never" depending on your disks.
  # no_check = "standby"

  ## Optionally specify devices to exclude from reporting if disks auto-discovery is performed.
  # excludes = [ "/dev/pass6" ]

  ## Optionally specify devices and device type, if unset a scan (smartctl --scan and smartctl --scan -d nvme) for S.M.A.R.T. devices will be done
  ## and all found will be included except for the excluded in excludes.
  # devices = [ "/dev/ada0 -d atacam", "/dev/nvme0"]

  ## Customer tags, if set will be seen with every metric.
  [inputs.smart.tags]
    # "key1" = "value1"
    # "key2" = "value2"
`
	l = logger.DefaultSLogger(inputName)
)

type nvmeDevice struct {
	name         string
	vendorID     string
	model        string
	serialNumber string
}

type Input struct {
	SmartCtlPath     string           `toml:"smartctl_path"`
	NvmePath         string           `toml:"nvme_path"`
	Interval         datakit.Duration `toml:"interval"`
	Timeout          datakit.Duration `toml:"timeout"`
	EnableExtensions []string         `toml:"enable_extensions"`
	UseSudo          bool             `toml:"use_sudo"`
	NoCheck          string           `toml:"no_check"`
	Excludes         []string         `toml:"excludes"`
	Devices          []string         `toml:"devices"`

	getter diskAttributeGetter

	Tags       map[string]string `toml:"tags"`
	mergedTags map[string]string

	ptsTime time.Time
	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	tagger  datakit.GlobalTagger
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) AvailableArchs() []string { return datakit.AllOS }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&smartMeasurement{}}
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	var err error
	if ipt.SmartCtlPath == "" || !ipath.IsFileExists(ipt.SmartCtlPath) {
		if ipt.SmartCtlPath, err = exec.LookPath(defSmartCmd); err != nil {
			l.Error("Can not find executable sensor command, install 'smartmontools' first.")

			return
		}
		l.Infof("Command fallback to %q due to invalide path provided in 'smart' input", ipt.SmartCtlPath)
	}
	if ipt.NvmePath == "" || !ipath.IsFileExists(ipt.NvmePath) {
		if ipt.NvmePath, err = exec.LookPath(defNvmeCmd); err != nil {
			ipt.NvmePath = ""
			l.Debug("Can not find executable sensor command, install 'nvme-cli' first.")
		} else {
			l.Infof("Command fallback to %q due to invalide path provided in 'smart' input", ipt.NvmePath)
		}
	}

	l.Info("smartctl input started")

	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()
	ipt.ptsTime = ntp.Now()

	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")

	// setup getter impl
	ipt.getter.(*smartctlGetter).nocheck = ipt.NoCheck
	ipt.getter.(*smartctlGetter).exePath = ipt.SmartCtlPath
	ipt.getter.(*smartctlGetter).timeout = ipt.Timeout.Duration
	ipt.getter.(*smartctlGetter).sudo = ipt.UseSudo

	for {
		if err := ipt.gather(); err != nil {
			l.Errorf("gagher: %s", err.Error())
			metrics.FeedLastError(inputName, err.Error())
		}

		select {
		case tt := <-tick.C:
			ipt.ptsTime = inputs.AlignTime(tt, ipt.ptsTime, ipt.Interval.Duration)
		case <-datakit.Exit.Wait():
			l.Info("smart input exits")
			return

		case <-ipt.semStop.Wait():
			l.Info("smart input return")
			return
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func defaultInput() *Input {
	return &Input{
		SmartCtlPath:     defSmartCtlPath,
		NvmePath:         defNvmePath,
		Interval:         defInterval,
		Timeout:          defTimeout,
		EnableExtensions: []string{"auto-on"},
		NoCheck:          "standby",
		getter:           &smartctlGetter{},

		semStop: cliutils.NewSem(),
		feeder:  dkio.DefaultFeeder(),
		tagger:  datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
