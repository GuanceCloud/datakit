// Package disk collect host disk metrics.
package disk

import (
	"fmt"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ReadEnv = (*Input)(nil)

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

  ## By default stats will be gathered for all mount points.
  ## Set mount_points will restrict the stats to only the specified mount points.
  # mount_points = ["/"]

  ## Ignore mount points by filesystem type.
  ignore_fs = ["tmpfs", "devtmpfs", "devfs", "iso9660", "overlay", "aufs", "squashfs"]

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

func (m *diskMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

//nolint:lll
func (m *diskMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "disk",
		Fields: map[string]interface{}{
			"total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Total disk size in bytes",
			},
			"free": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Free disk size in bytes",
			},
			"used": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeByte,
				Desc: "Used disk size in bytes",
			},
			"used_percent": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "Used disk size in percent",
			},
			"inodes_total": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Total inodes",
			},
			"inodes_free": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Free inodes",
			},
			"inodes_used": &inputs.FieldInfo{
				Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.NCount,
				Desc: "Used inodes",
			},
		},
		Tags: map[string]interface{}{
			"host":   &inputs.TagInfo{Desc: "主机名"},
			"device": &inputs.TagInfo{Desc: "磁盘设备名"},
			"fstype": &inputs.TagInfo{Desc: "文件系统名"},
			"mode":   &inputs.TagInfo{Desc: "读写模式"},
			"path":   &inputs.TagInfo{Desc: "磁盘挂载点"},
		},
	}
}

type Input struct {
	Interval    datakit.Duration
	MountPoints []string          `toml:"mount_points"`
	IgnoreFS    []string          `toml:"ignore_fs"`
	Tags        map[string]string `toml:"tags"`

	collectCache         []inputs.Measurement
	collectCacheLast1Ptr inputs.Measurement
	diskStats            PSDiskStats

	semStop *cliutils.Sem // start stop signal
}

func (ipt *Input) appendMeasurement(name string,
	tags map[string]string,
	fields map[string]interface{}, ts time.Time) {
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
	return datakit.AllArch
}

func (ipt *Input) Collect() error {
	ipt.collectCache = make([]inputs.Measurement, 0)
	disks, partitions, err := ipt.diskStats.FilterUsage(ipt.MountPoints, ipt.IgnoreFS)
	if err != nil {
		return fmt.Errorf("error getting disk usage info: %w", err)
	}
	ts := time.Now()
	for index, du := range disks {
		if du.Total == 0 {
			// Skip dummy filesystem (procfs, cgroupfs, ...)
			continue
		}
		mountOpts := parseOptions(partitions[index].Opts)
		tags := map[string]string{
			"path":   du.Path,
			"device": strings.ReplaceAll(partitions[index].Device, "/dev/", ""),
			"fstype": du.Fstype,
			"mode":   mountOpts.Mode(),
		}
		for k, v := range ipt.Tags {
			tags[k] = v
		}
		var usedPercent float64
		if du.Used+du.Free > 0 {
			usedPercent = float64(du.Used) /
				(float64(du.Used) + float64(du.Free)) * 100
		}
		fields := map[string]interface{}{
			"total":        du.Total,
			"free":         du.Free,
			"used":         du.Used,
			"used_percent": usedPercent,
			"inodes_total": du.InodesTotal,
			"inodes_free":  du.InodesFree,
			"inodes_used":  du.InodesUsed,
		}
		ipt.appendMeasurement(metricName, tags, fields, ts)
	}

	return nil
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("disk input started")
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	ipt.IgnoreFS = unique(ipt.IgnoreFS)

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
//   ENV_INPUT_DISK_IGNORE_FS : []string
func (ipt *Input) ReadEnv(envs map[string]string) {
	if fsList, ok := envs["ENV_INPUT_DISK_IGNORE_FS"]; ok {
		list := strings.Split(fsList, ",")
		l.Debugf("add ignore_fs from ENV: %v", fsList)
		ipt.IgnoreFS = append(ipt.IgnoreFS, list...)
	}

	if tagsStr, ok := envs["ENV_INPUT_DISK_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}
}

func unique(strSlice []string) []string {
	keys := make(map[string]interface{})
	list := []string{}
	for _, entry := range strSlice {
		if _, ok := keys[entry]; !ok {
			keys[entry] = nil
			list = append(list, entry)
		}
	}
	return list
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			diskStats: &PSDisk{},
			Interval:  datakit.Duration{Duration: time.Second * 10},

			semStop: cliutils.NewSem(),
			Tags:    make(map[string]string),
		}
	})
}
