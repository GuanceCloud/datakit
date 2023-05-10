// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package disk collect host disk metrics.
package disk

import (
	"fmt"
	"math"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/hostobject"
)

var (
	_ inputs.ReadEnv   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
)

var (
	inputName    = "disk"
	metricName   = "disk"
	l            = logger.DefaultSLogger(inputName)
	sampleConfig = `
[[inputs.disk]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

  # Physical devices only (e.g. hard disks, cd-rom drives, USB keys)
  # and ignore all others (e.g. memory partitions such as /dev/shm)
  only_physical_device = false

  ## Deprecated
  # ignore_mount_points = ["/"]

  ## Deprecated
  # mount_points = ["/"]


  ## Deprecated
  # ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"]

  ## Deprecated
  # fs = ["ext2", "ext3", "ext4", "NTFS"]

  ## We collect all devices prefixed with dev by default,If you want to collect additional devices, it's in extra_device add
  # extra_device = ["/nfsdata"]

  ## exclude some with dev prefix (We collect all devices prefixed with dev by default)
  # exclude_device = ["/dev/loop0","/dev/loop1"]
  [inputs.disk.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
)

type diskMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *diskMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOpt())
}

//nolint:lll
func (m *diskMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "disk",
		Fields: map[string]interface{}{
			"total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Total disk size in bytes.",
			},
			"free": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Free disk size in bytes.",
			},
			"used": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Used disk size in bytes.",
			},
			"used_percent": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "Used disk size in percent.",
			},
			"inodes_used_percent": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "Inode used percent",
			},
			"inodes_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Total Inode(**DEPRECATED: use inodes_total_mb instead**).",
			},
			"inodes_total_mb": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Total Inode(need to multiply by 10^6).",
			},
			"inodes_free": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Free Inode(**DEPRECATED: use inodes_free_mb instead**).",
			},
			"inodes_free_mb": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Free Inode(need to multiply by 10^6).",
			},
			"inodes_used_mb": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Used Inode(need to multiply by 10^6).",
			},
			"inodes_used": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Used Inode(**DEPRECATED: use inodes_used_mb instead**).",
			},
		},
		Tags: map[string]interface{}{
			"host":   &inputs.TagInfo{Desc: "System hostname."},
			"device": &inputs.TagInfo{Desc: "Disk device name."},
			"fstype": &inputs.TagInfo{Desc: "File system name."},
		},
	}
}

type Input struct {
	Interval datakit.Duration

	Tags          map[string]string `toml:"tags"`
	ExtraDevice   []string          `toml:"extra_device"`
	ExcludeDevice []string          `toml:"exclude_device"`

	IgnoreZeroBytesDisk bool `toml:"ignore_zero_bytes_disk"`
	OnlyPhysicalDevice  bool `toml:"only_physical_device"`

	collectCache         []inputs.Measurement
	collectCacheLast1Ptr inputs.Measurement
	diskStats            PSDiskStats

	semStop *cliutils.Sem // start stop signal
}

func (ipt *Input) Singleton() {
}

func (ipt *Input) appendMeasurement(name string,
	tags map[string]string,
	fields map[string]interface{}, ts time.Time,
) {
	tmp := &diskMeasurement{name: name, tags: tags, fields: fields, ts: ts}
	ipt.collectCache = append(ipt.collectCache, tmp)
	ipt.collectCacheLast1Ptr = tmp
}

func (*Input) Catalog() string {
	return "host"
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&diskMeasurement{},
	}
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (ipt *Input) Collect() error {
	ipt.collectCache = make([]inputs.Measurement, 0)
	disks, partitions, err := ipt.diskStats.FilterUsage()
	if err != nil {
		return fmt.Errorf("error getting disk usage info: %w", err)
	}
	ts := time.Now()
	for index, du := range disks {
		if du.Total == 0 {
			// Skip dummy filesystem (procfs, cgroupfs, ...)
			continue
		}

		tags := map[string]string{
			"device": partitions[index].Device,
			"fstype": du.Fstype,
		}
		for k, v := range ipt.Tags {
			tags[k] = v
		}
		var usedPercent float64
		if du.Used+du.Free > 0 {
			usedPercent = float64(du.Used) /
				(float64(du.Used) + float64(du.Free)) * 100
		}
		atomic.StoreUint64(&hostobject.DiskUsed, du.Used)
		atomic.StoreUint64(&hostobject.DiskFree, du.Free)

		fields := map[string]interface{}{
			"total":        du.Total,
			"free":         du.Free,
			"used":         du.Used,
			"used_percent": usedPercent,
		}

		switch runtime.GOOS {
		case datakit.OSLinux, datakit.OSDarwin:
			fields["inodes_total_mb"] = du.InodesTotal / 1_000_000
			fields["inodes_free_mb"] = du.InodesFree / 1_000_000
			fields["inodes_used_mb"] = du.InodesUsed / 1_000_000

			fields["inodes_used_percent"] = du.InodesUsedPercent // float64

			// Deprecated
			fields["inodes_total"] = wrapUint64(du.InodesTotal)
			fields["inodes_free"] = wrapUint64(du.InodesFree)
			fields["inodes_used"] = wrapUint64(du.InodesUsed)
		}

		ipt.appendMeasurement(metricName, tags, fields, ts)
	}

	return nil
}

func wrapUint64(x uint64) int64 {
	if x > uint64(math.MaxInt64) {
		return -1
	}
	return int64(x)
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("disk input started")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	ipt.ExtraDevice = unique(ipt.ExtraDevice)
	ipt.ExcludeDevice = unique(ipt.ExcludeDevice)
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	for {
		start := time.Now()
		if err := ipt.Collect(); err != nil {
			l.Errorf("Collect: %s", err)
			io.FeedLastError(inputName, err.Error())
		}

		if len(ipt.collectCache) > 0 {
			if errFeed := inputs.FeedMeasurement(metricName, datakit.Metric, ipt.collectCache,
				&io.Option{CollectCost: time.Since(start)}); errFeed != nil {
				io.FeedLastError(inputName, errFeed.Error())
			}
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Infof("disk input exit")

			return
		case <-ipt.semStop.Wait():
			l.Info("disk input return")

			return
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

// ReadEnv support envs：
//   ENV_INPUT_DISK_EXCLUDE_DEVICE : []string
//   ENV_INPUT_DISK_EXTRA_DEVICE : []string
//   ENV_INPUT_DISK_TAGS : "a=b,c=d"
//   ENV_INPUT_DISK_ONLY_PHYSICAL_DEVICE : bool
//   ENV_INPUT_DISK_INTERVAL : datakit.Duration
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

	//   ENV_INPUT_DISK_INTERVAL : datakit.Duration
	//   ENV_INPUT_DISK_MOUNT_POINTS : []string
	if str, ok := envs["ENV_INPUT_DISK_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_DISK_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			ipt.Interval.Duration = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}
}

func unique(strSlice []string) []string {
	keys := make(map[string]interface{})
	var list []string
	for _, entry := range strSlice {
		if _, ok := keys[entry]; !ok {
			keys[entry] = nil
			list = append(list, entry)
		}
	}
	return list
}

func newDefaultInput() *Input {
	ipt := &Input{
		Interval: datakit.Duration{Duration: time.Second * 10},
		semStop:  cliutils.NewSem(),
		Tags:     make(map[string]string),
	}

	x := &PSDisk{ipt: ipt}
	ipt.diskStats = x
	return ipt
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return newDefaultInput()
	})
}
