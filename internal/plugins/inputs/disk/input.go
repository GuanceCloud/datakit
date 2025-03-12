// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package disk collect host disk metrics.
package disk

import (
	"fmt"
	"regexp"
	"runtime"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/pcommon"
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

type DiskCacheEntry struct {
	Disks       []string
	LastUpdated time.Time
}

type Input struct {
	Interval time.Duration

	Tags          map[string]string `toml:"tags"`
	ExtraDevice   []string          `toml:"extra_device"`
	ExcludeDevice []string          `toml:"exclude_device"`

	IgnoreZeroBytesDisk bool `toml:"ignore_zero_bytes_disk"`

	IgnoreFSTypes    string `toml:"ignore_fstypes"`
	regIgnoreFSTypes *regexp.Regexp

	IgnoreMountpoints    string `toml:"ignore_mountpoints"`
	regIgnoreMountpoints *regexp.Regexp

	semStop      *cliutils.Sem
	collectCache []*point.Point
	diskStats    pcommon.DiskStats
	feeder       dkio.Feeder
	mergedTags   map[string]string
	tagger       datakit.GlobalTagger

	hostRoot string
}

func (ipt *Input) Run() {
	ipt.setup()

	tick := time.NewTicker(ipt.Interval)
	defer tick.Stop()

	start := time.Now()
	for {
		if err := ipt.collect(start.UnixNano()); err != nil {
			l.Errorf("collect: %s", err)
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Metric),
			)
		}

		if len(ipt.collectCache) > 0 {
			if err := ipt.feeder.FeedV2(point.Metric, ipt.collectCache,
				dkio.WithCollectCost(time.Since(start)),
				dkio.WithElection(false),
				dkio.WithInputName(metricName)); err != nil {
				ipt.feeder.FeedLastError(err.Error(),
					metrics.WithLastErrorInput(inputName),
					metrics.WithLastErrorCategory(point.Metric),
				)
				l.Errorf("feed measurement: %s", err)
			}
		}

		select {
		case tt := <-tick.C:
			start = time.UnixMilli(inputs.AlignTimeMillSec(tt, start.UnixMilli(), ipt.Interval.Milliseconds()))

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

	if ipt.IgnoreFSTypes != "" {
		if re, err := regexp.Compile(ipt.IgnoreFSTypes); err != nil {
			l.Warnf("regexp.Compile(%q): %s, ignored", ipt.IgnoreFSTypes, err.Error())
		} else {
			ipt.regIgnoreFSTypes = re
		}
	}

	if ipt.IgnoreMountpoints != "" {
		if re, err := regexp.Compile(ipt.IgnoreMountpoints); err != nil {
			l.Warnf("regexp.Compile(%q): %s, ignored", ipt.IgnoreMountpoints, err.Error())
		} else {
			ipt.regIgnoreMountpoints = re
		}
	}
}

func (ipt *Input) collect(ptTS int64) error {
	ipt.collectCache = make([]*point.Point, 0)
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(ptTS))

	fs, err := ipt.filterUsage()
	if err != nil {
		return fmt.Errorf("error getting disk usage info: %w", err)
	}

	for _, f := range fs {
		if f.Usage == nil {
			l.Infof("no usage available, skip partition %+#v", f.Part)
			continue
		}

		if f.Usage.Total == 0 {
			l.Infof("usage.Total == 0, skip partition %+#v", f.Part)
			continue
		}

		var kvs point.KVs
		kvs = kvs.Add("used_percent", float64(f.Usage.Used)/float64(f.Usage.Total)*100.0, false, true)
		kvs = kvs.Add("device", f.Part.Device, true, true)
		kvs = kvs.Add("fstype", f.Part.Fstype, true, true)

		kvs = kvs.Add("total", f.Usage.Total, false, true)
		kvs = kvs.Add("free", f.Usage.Free, false, true)
		kvs = kvs.Add("used", f.Usage.Used, false, true)

		switch runtime.GOOS {
		case datakit.OSLinux, datakit.OSDarwin:
			kvs = kvs.Add("inodes_total_mb", f.Usage.InodesTotal/1_000_000, false, true)
			kvs = kvs.Add("inodes_free_mb", f.Usage.InodesFree/1_000_000, false, true)
			kvs = kvs.Add("inodes_used_mb", f.Usage.InodesUsed/1_000_000, false, true)
			kvs = kvs.Add("inodes_used_percent", f.Usage.InodesUsedPercent, false, true) // float64

			kvs = kvs.Add("inodes_total", f.Usage.InodesTotal, false, true) // Deprecated
			kvs = kvs.Add("inodes_free", f.Usage.InodesFree, false, true)   // Deprecated
			kvs = kvs.Add("inodes_used", f.Usage.InodesUsed, false, true)   // Deprecated

			kvs = kvs.AddTag("mount_point", f.Part.Mountpoint)
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

func defaultInput() *Input {
	ipt := &Input{
		Interval: time.Second * 10,

		IgnoreFSTypes:     `^(tmpfs|autofs|binfmt_misc|devpts|fuse.lxcfs|overlay|proc|squashfs|sysfs)$`,
		IgnoreMountpoints: `^(/usr/local/datakit/.*|/run/containerd/.*)$`,

		semStop: cliutils.NewSem(),
		Tags:    make(map[string]string),
		feeder:  dkio.DefaultFeeder(),
		tagger:  datakit.DefaultGlobalTagger(),
	}

	ipt.diskStats = &pcommon.DiskStatsImpl{}

	return ipt
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
