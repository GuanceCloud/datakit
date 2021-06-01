package diskio

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/shirou/gopsutil/disk"
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
	inputName    = "diskio"
	metricName   = "diskio"
	l            = logger.DefaultSLogger(inputName)
	varRegex     = regexp.MustCompile(`\$(?:\w+|\{\w+\})`)
	sampleConfig = `
[[inputs.diskio]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'
  ##
  ## By default, gather stats for all devices including
  ## disk partitions.
  ## Setting interfaces using regular expressions will collect these expected devices.
  # devices = ['''^sda\d*''', '''^sdb\d*''', '''vd.*''']
  #
  ## If the disk serial number is not required, please uncomment the following line.
  # skip_serial_number = true
  #
  ## On systems which support it, device metadata can be added in the form of
  ## tags.
  ## Currently only Linux is supported via udev properties. You can view
  ## available properties for a device by running:
  ## 'udevadm info -q property -n /dev/sda'
  ## Note: Most, but not all, udev properties can be accessed this way. Properties
  ## that are currently inaccessible include DEVTYPE, DEVNAME, and DEVPATH.
  # device_tags = ["ID_FS_TYPE", "ID_FS_USAGE"]
  #
  ## Using the same metadata source as device_tags,
  ## you can also customize the name of the device through a template.
  ## The "name_templates" parameter is a list of templates to try to apply equipment.
  ## The template can contain variables of the form "$PROPERTY" or "${PROPERTY}".
  ## The first template that does not contain any variables that do not exist
  ## for the device is used as the device name label.
  ## A typical use case for LVM volumes is to obtain VG/LV names,
  ## not DM-0 names which are almost meaningless.
  ## In addition, "device" is reserved specifically to indicate the device name.
  # name_templates = ["$ID_FS_LABEL","$DM_VG_NAME/$DM_LV_NAME", "$device:$ID_FS_TYPE"]
  #

[inputs.diskio.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"`
)

type diskioMeasurement measurement

func (m *diskioMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

// https://www.kernel.org/doc/Documentation/ABI/testing/procfs-diskstats

func (m *diskioMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "diskio",
		Fields: map[string]interface{}{
			"reads":            newFieldsInfoCount("reads completed successfully"),
			"writes":           newFieldsInfoCount("writes completed"),
			"read_bytes":       newFieldsInfoBytes("read bytes"),
			"read_bytes/sec":   newFieldsInfoBytesPerSec("read bytes per second"),
			"write_bytes":      newFieldsInfoBytes("write bytes"),
			"write_bytes/sec":  newFieldsInfoBytesPerSec("write bytes per second"),
			"read_time":        newFieldsInfoMS("time spent reading"),
			"write_time":       newFieldsInfoMS("time spent writing"),
			"io_time":          newFieldsInfoMS("time spent doing I/Os"),
			"weighted_io_time": newFieldsInfoMS("weighted time spent doing I/Os"),
			"iops_in_progress": newFieldsInfoCount("I/Os currently in progress"),
			"merged_reads":     newFieldsInfoCount("reads merged"),
			"merged_writes":    newFieldsInfoCount("writes merged"),
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "主机名"},
			"name": &inputs.TagInfo{Desc: "磁盘设备名"},
		},
	}
}

type Input struct {
	Interval         datakit.Duration
	Devices          []string
	DeviceTags       []string
	NameTemplates    []string
	SkipSerialNumber bool
	Tags             map[string]string

	collectCache []inputs.Measurement
	lastStat     map[string]disk.IOCountersStat
	lastTime     time.Time
	diskIO       DiskIO

	infoCache    map[string]diskInfoCache
	deviceFilter *DevicesFilter
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) Catalog() string {
	return "host"
}

func (i *Input) SampleConfig() string {
	return sampleConfig
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&diskioMeasurement{},
	}
}

func (i *Input) Collect() error {
	// 设置 disk device 过滤器
	i.deviceFilter = &DevicesFilter{}
	err := i.deviceFilter.Compile(i.Devices)
	if err != nil {
		return err
	}

	// disk io stat
	diskio, err := i.diskIO([]string{}...)
	if err != nil {
		return fmt.Errorf("error getting disk io info: %s", err.Error())
	}

	ts := time.Now()
	for _, stat := range diskio {
		match := false

		// 匹配 disk name
		if len(i.deviceFilter.filters) < 1 || i.deviceFilter.Match(stat.Name) {
			match = true
		}

		tags := map[string]string{}
		// 用户自定义tags
		for k, v := range i.Tags {
			tags[k] = v
		}
		var devLinks []string

		tags["name"], devLinks = i.diskName(stat.Name)

		if !match {
			for _, devLink := range devLinks {
				if i.deviceFilter.Match(devLink) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		for t, v := range i.diskTags(stat.Name) {
			tags[t] = v
		}

		if !i.SkipSerialNumber {
			if len(stat.SerialNumber) != 0 {
				tags["serial"] = stat.SerialNumber
			} else {
				tags["serial"] = "unknown"
			}
		}

		fields := map[string]interface{}{
			"reads":            stat.ReadCount,
			"writes":           stat.WriteCount,
			"read_bytes":       stat.ReadBytes,
			"write_bytes":      stat.WriteBytes,
			"read_time":        stat.ReadTime,
			"write_time":       stat.WriteTime,
			"io_time":          stat.IoTime,
			"weighted_io_time": stat.WeightedIO,
			"iops_in_progress": stat.IopsInProgress,
			"merged_reads":     stat.MergedReadCount,
			"merged_writes":    stat.MergedWriteCount,
		}
		if i.lastStat != nil {
			deltaTime := ts.Unix() - i.lastTime.Unix()
			if v, ok := i.lastStat[stat.Name]; ok && deltaTime > 0 {
				if stat.ReadBytes >= v.ReadBytes {
					fields["read_bytes/sec"] = int64(stat.ReadBytes-v.ReadBytes) / deltaTime
				}
				if stat.WriteBytes >= v.WriteBytes {
					fields["write_bytes/sec"] = int64(stat.WriteBytes-v.WriteBytes) / deltaTime
				}
			}
		}
		i.collectCache = append(i.collectCache, &diskioMeasurement{name: metricName, tags: tags, fields: fields, ts: ts})
	}
	i.lastStat = diskio
	i.lastTime = ts
	return nil
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("diskio input started")
	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)

	l.Infof("diskio input started, collect interval: %v", i.Interval.Duration)

	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			start := time.Now()
			i.collectCache = make([]inputs.Measurement, 0)
			if err := i.Collect(); err == nil {
				if errFeed := inputs.FeedMeasurement(metricName, datakit.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since(start)}); errFeed != nil {

					l.Error(errFeed)
					io.FeedLastError(inputName, errFeed.Error())
				}
			} else {
				l.Error(err)
				io.FeedLastError(inputName, err.Error())
			}
		case <-datakit.Exit.Wait():
			l.Infof("diskio input exit")
			return
		}
	}
}

func (i *Input) diskName(devName string) (string, []string) {
	di, err := i.diskInfo(devName)

	devLinks := strings.Split(di["DEVLINKS"], " ")
	for i, devLink := range devLinks {
		devLinks[i] = strings.TrimPrefix(devLink, "/dev/")
	}

	if err != nil {
		l.Warnf("Error gathering disk info: %s", err)
		return devName, devLinks
	}

	// diskInfo empty
	if len(i.NameTemplates) == 0 || len(di) == 0 {
		return devName, devLinks
	}

	// render name templates
	for _, nt := range i.NameTemplates {
		miss := false
		name := varRegex.ReplaceAllStringFunc(nt, func(sub string) string {
			sub = sub[1:]
			if sub[0] == '{' {
				sub = sub[1 : len(sub)-1]
			}
			if v, ok := di[sub]; ok {
				return v
			}
			if sub == "device" {
				return devName
			}
			miss = true
			return ""
		})
		if !miss { // must match all variables
			return name, devLinks
		}
	}
	return devName, devLinks
}

func (i *Input) diskTags(devName string) map[string]string {
	if len(i.DeviceTags) == 0 {
		return nil
	}

	di, err := i.diskInfo(devName)
	if err != nil {
		l.Warnf("Error gathering disk info: %s", err)
		return nil
	}

	tags := map[string]string{}
	for _, dt := range i.DeviceTags {
		if v, ok := di[dt]; ok {
			tags[dt] = v
		}
	}

	return tags
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			diskIO:   disk.IOCounters,
			Interval: datakit.Duration{Duration: time.Second * 10},
		}
	})
}
