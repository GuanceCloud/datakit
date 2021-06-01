package disk

import (
	"fmt"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
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
  ##
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

func (m *diskMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "disk",
		Fields: map[string]interface{}{
			"total": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeIByte,
				Desc: "Total disk size in bytes",
			},
			"free": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeIByte,
				Desc: "Free disk size in bytes",
			},
			"used": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.SizeIByte,
				Desc: "Used disk size in bytes",
			},
			"used_percent": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent,
				Desc: "Used disk size in percent",
			},
			"inodes_total": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Count,
				Desc: "Total inodes",
			},
			"inodes_free": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Count,
				Desc: "Free inodes",
			},
			"inodes_used": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.Count,
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
}

func (i *Input) appendMeasurement(name string, tags map[string]string, fields map[string]interface{}, ts time.Time) {
	tmp := &diskMeasurement{name: name, tags: tags, fields: fields, ts: ts}
	i.collectCache = append(i.collectCache, tmp)
	i.collectCacheLast1Ptr = tmp
}

func (i *Input) Catalog() string {
	return "host"
}

func (i *Input) SampleConfig() string {
	return sampleConfig
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&diskMeasurement{},
	}
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) Collect() error {
	i.collectCache = make([]inputs.Measurement, 0)
	disks, partitions, err := i.diskStats.FilterUsage(i.MountPoints, i.IgnoreFS)
	if err != nil {
		return fmt.Errorf("error getting disk usage info: %s", err)
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
			"device": strings.Replace(partitions[index].Device, "/dev/", "", -1),
			"fstype": du.Fstype,
			"mode":   mountOpts.Mode(),
		}
		for k, v := range i.Tags {
			tags[k] = v
		}
		var used_percent float64
		if du.Used+du.Free > 0 {
			used_percent = float64(du.Used) /
				(float64(du.Used) + float64(du.Free)) * 100
		}
		fields := map[string]interface{}{
			"total":        du.Total,
			"free":         du.Free,
			"used":         du.Used,
			"used_percent": used_percent,
			"inodes_total": du.InodesTotal,
			"inodes_free":  du.InodesFree,
			"inodes_used":  du.InodesUsed,
		}
		i.appendMeasurement(metricName, tags, fields, ts)
	}

	return nil
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("disk input started")
	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)
	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			start := time.Now()
			if err := i.Collect(); err == nil {
				if errFeed := inputs.FeedMeasurement(metricName, datakit.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since(start)}); errFeed != nil {
					io.FeedLastError(inputName, errFeed.Error())
				}
			} else {
				io.FeedLastError(inputName, err.Error())
				l.Error(err)
			}
		case <-datakit.Exit.Wait():
			l.Infof("disk input exit")
			return
		}
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			diskStats: &PSDisk{},
			Interval:  datakit.Duration{Duration: time.Second * 10},
		}
	})
}
