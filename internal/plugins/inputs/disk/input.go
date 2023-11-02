// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package disk collect host disk metrics.
package disk

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
	inputName   = "disk"
	metricName  = inputName
)

var (
	_ inputs.ReadEnv   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
	l                  = logger.DefaultSLogger(inputName)
)

type Input struct {
	Interval time.Duration

	Tags          map[string]string `toml:"tags"`
	ExtraDevice   []string          `toml:"extra_device"`
	ExcludeDevice []string          `toml:"exclude_device"`

	IgnoreZeroBytesDisk bool `toml:"ignore_zero_bytes_disk"`
	OnlyPhysicalDevice  bool `toml:"only_physical_device"`

	semStop      *cliutils.Sem
	collectCache []*point.Point
	diskStats    PSDiskStats
	feeder       dkio.Feeder
	mergedTags   map[string]string
	tagger       datakit.GlobalTagger
}

func (ipt *Input) Run() {
	ipt.setup()

	ipt.ExtraDevice = unique(ipt.ExtraDevice)
	ipt.ExcludeDevice = unique(ipt.ExcludeDevice)

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	for {
		start := time.Now()
		if err := ipt.collect(); err != nil {
			l.Errorf("collect: %s", err)
			ipt.feeder.FeedLastError(err.Error(),
				dkio.WithLastErrorInput(inputName),
				dkio.WithLastErrorCategory(point.Metric),
			)
		}

		if len(ipt.collectCache) > 0 {
			if err := ipt.feeder.Feed(metricName, point.Metric, ipt.collectCache,
				&dkio.Option{CollectCost: time.Since(start)}); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					dkio.WithLastErrorInput(inputName),
					dkio.WithLastErrorCategory(point.Metric),
				)
				l.Errorf("feed measurement: %s", err)
			}
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Infof("%s input exit", inputName)
			return
		case <-ipt.semStop.Wait():
			l.Infof("%s input return", inputName)
			return
		}
	}
}

func (ipt *Input) setup() {
	l = logger.SLogger(inputName)

	l.Infof("%s input started", inputName)
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)
	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")
	l.Debugf("merged tags: %+#v", ipt.mergedTags)
}

func (ipt *Input) collect() error {
	ipt.collectCache = make([]*point.Point, 0)
	ts := time.Now()
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ts))

	disks, partitions, err := ipt.diskStats.FilterUsage()
	if err != nil {
		return fmt.Errorf("error getting disk usage info: %w", err)
	}
	for index, du := range disks {
		var kvs point.KVs

		if du.Total == 0 {
			// Skip dummy filesystem (procfs, cgroupfs, ...)
			continue
		}

		kvs = kvs.Add("device", partitions[index].Device, true, true)
		kvs = kvs.Add("fstype", du.Fstype, true, true)

		var usedPercent float64
		if du.Used+du.Free > 0 {
			usedPercent = float64(du.Used) /
				(float64(du.Used) + float64(du.Free)) * 100
		}
		kvs = kvs.Add("total", du.Total, false, true)
		kvs = kvs.Add("free", du.Free, false, true)
		kvs = kvs.Add("used", du.Used, false, true)
		kvs = kvs.Add("used_percent", usedPercent, false, true)

		switch runtime.GOOS {
		case datakit.OSLinux, datakit.OSDarwin:
			kvs = kvs.Add("inodes_total_mb", du.InodesTotal/1_000_000, false, true)
			kvs = kvs.Add("inodes_free_mb", du.InodesFree/1_000_000, false, true)
			kvs = kvs.Add("inodes_used_mb", du.InodesUsed/1_000_000, false, true)
			kvs = kvs.Add("inodes_used_percent", du.InodesUsedPercent, false, true) // float64
			kvs = kvs.Add("inodes_total", wrapUint64(du.InodesTotal), false, true)  // Deprecated
			kvs = kvs.Add("inodes_free", wrapUint64(du.InodesFree), false, true)    // Deprecated
			kvs = kvs.Add("inodes_used", wrapUint64(du.InodesUsed), false, true)    // Deprecated
			kvs = kvs.Add("mount_point", partitions[index].Mountpoint, true, true)
		}

		for k, v := range ipt.mergedTags {
			kvs = kvs.AddTag(k, v)
		}

		ipt.collectCache = append(ipt.collectCache, point.NewPointV2(inputName, kvs, opts...))
	}

	return nil
}

func (*Input) Singleton() {}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}
func (*Input) Catalog() string          { return "host" }
func (*Input) SampleConfig() string     { return sampleCfg }
func (*Input) AvailableArchs() []string { return datakit.AllOS }
func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&docMeasurement{},
	}
}

// ReadEnv support envs：
//
//	ENV_INPUT_DISK_EXCLUDE_DEVICE : []string
//	ENV_INPUT_DISK_EXTRA_DEVICE : []string
//	ENV_INPUT_DISK_TAGS : "a=b,c=d"
//	ENV_INPUT_DISK_ONLY_PHYSICAL_DEVICE : bool
//	ENV_INPUT_DISK_INTERVAL : time.Duration
func (ipt *Input) ReadEnv(envs map[string]string) {
	if fsList, ok := envs["ENV_INPUT_DISK_EXTRA_DEVICE"]; ok {
		list := strings.Split(fsList, ",")
		l.Debugf("add extra_device from ENV: %v", fsList)
		ipt.ExtraDevice = append(ipt.ExtraDevice, list...)
	}
	if fsList, ok := envs["ENV_INPUT_DISK_EXCLUDE_DEVICE"]; ok {
		list := strings.Split(fsList, ",")
		l.Debugf("add exlude_device from ENV: %v", fsList)
		ipt.ExcludeDevice = append(ipt.ExcludeDevice, list...)
	}

	if tagsStr, ok := envs["ENV_INPUT_DISK_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	if str := envs["ENV_INPUT_DISK_ONLY_PHYSICAL_DEVICE"]; str != "" {
		ipt.OnlyPhysicalDevice = true
	}

	//   ENV_INPUT_DISK_INTERVAL : time.Duration
	//   ENV_INPUT_DISK_MOUNT_POINTS : []string
	if str, ok := envs["ENV_INPUT_DISK_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_DISK_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}
}

func defaultInput() *Input {
	ipt := &Input{
		Interval: time.Second * 10,
		semStop:  cliutils.NewSem(),
		Tags:     make(map[string]string),
		feeder:   dkio.DefaultFeeder(),
		tagger:   datakit.DefaultGlobalTagger(),
	}

	x := &PSDisk{ipt: ipt}
	ipt.diskStats = x
	return ipt
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
