// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package diskio collet disk IO metrics.
package diskio

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/shirou/gopsutil/disk"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
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

type diskioMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *diskioMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(m.name, m.tags, m.fields, point.MOpt())
}

// https://www.kernel.org/doc/Documentation/ABI/testing/procfs-diskstats

func (m *diskioMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "diskio",
		Fields: map[string]interface{}{
			"reads":            newFieldsInfoCount("The number of read requests."),
			"writes":           newFieldsInfoCount("The number of write requests."),
			"read_bytes":       newFieldsInfoBytes("The number of bytes read from the device."),
			"read_bytes/sec":   newFieldsInfoBytesPerSec("The number of bytes read from the per second."),
			"write_bytes":      newFieldsInfoBytes("The number of bytes written to the device."),
			"write_bytes/sec":  newFieldsInfoBytesPerSec("The number of bytes written to the device per second."),
			"read_time":        newFieldsInfoMS("Time spent reading."),
			"write_time":       newFieldsInfoMS("Time spent writing."),
			"io_time":          newFieldsInfoMS("Time spent doing I/Os."),
			"weighted_io_time": newFieldsInfoMS("Weighted time spent doing I/Os."),
			"iops_in_progress": newFieldsInfoCount("I/Os currently in progress."),
			"merged_reads":     newFieldsInfoCount("The number of merged read requests."),
			"merged_writes":    newFieldsInfoCount("The number of merged write requests."),
		},
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "System hostname."},
			"name": &inputs.TagInfo{Desc: "Device name."},
		},
	}
}

//nolint:unused,structcheck
type diskInfoCache struct {
	// Unix Nano timestamp of the last modification of the device.
	// This value is used to invalidate the cache
	modifiedAt int64

	udevDataPath string
	values       map[string]string
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

	infoCache    map[string]diskInfoCache //nolint:structcheck,unused
	deviceFilter *DevicesFilter

	semStop *cliutils.Sem // start stop signal
}

func (i *Input) Singleton() {
}

func (*Input) AvailableArchs() []string {
	return []string{
		datakit.OSLabelLinux, datakit.OSLabelWindows,
		datakit.LabelK8s, datakit.LabelDocker,
	}
}

func (*Input) Catalog() string {
	return "host"
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&diskioMeasurement{},
	}
}

func (i *Input) Collect() error {
	// set disk device filter
	i.deviceFilter = &DevicesFilter{}
	err := i.deviceFilter.Compile(i.Devices)
	if err != nil {
		return err
	}

	// disk io stat
	diskio, err := i.diskIO([]string{}...)
	if err != nil {
		return fmt.Errorf("error getting disk io info: %w", err)
	}

	ts := time.Now()
	for _, stat := range diskio {
		match := false

		// match disk name
		if len(i.deviceFilter.filters) < 1 || i.deviceFilter.Match(stat.Name) {
			match = true
		}

		tags := map[string]string{}
		// user-defined tags
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
		start := time.Now()
		i.collectCache = make([]inputs.Measurement, 0)
		if err := i.Collect(); err != nil {
			l.Errorf("Collect: %s", err)

			io.FeedLastError(inputName, err.Error())
		}

		if len(i.collectCache) > 0 {
			if errFeed := inputs.FeedMeasurement(metricName, datakit.Metric, i.collectCache,
				&io.Option{CollectCost: time.Since(start)}); errFeed != nil {
				l.Errorf("FeedMeasurement: %s", errFeed)

				io.FeedLastError(inputName, errFeed.Error())
			}
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Infof("diskio input exit")
			return

		case <-i.semStop.Wait():
			l.Info("diskio input return")
			return
		}
	}
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}

// ReadEnv support envs：
//   ENV_INPUT_DISKIO_SKIP_SERIAL_NUMBER : booler
//   ENV_INPUT_DISKIO_TAGS : "a=b,c=d"
//   ENV_INPUT_DISKIO_INTERVAL : datakit.Duration
//   ENV_INPUT_DISKIO_DEVICES : []string
//   ENV_INPUT_DISKIO_DEVICE_TAGS : []string
//   ENV_INPUT_DISKIO_NAME_TEMPLATES : []string
func (i *Input) ReadEnv(envs map[string]string) {
	if skip, ok := envs["ENV_INPUT_DISKIO_SKIP_SERIAL_NUMBER"]; ok {
		b, err := strconv.ParseBool(skip)
		if err != nil {
			l.Warnf("parse ENV_INPUT_DISKIO_SKIP_SERIAL_NUMBER to bool: %s, ignore", err)
		} else {
			i.SkipSerialNumber = b
		}
	}

	if tagsStr, ok := envs["ENV_INPUT_DISKIO_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			i.Tags[k] = v
		}
	}

	//   ENV_INPUT_DISKIO_INTERVAL : datakit.Duration
	//   ENV_INPUT_DISKIO_DEVICES : []string
	//   ENV_INPUT_DISKIO_DEVICE_TAGS : []string
	//   ENV_INPUT_DISKIO_NAME_TEMPLATES : []string
	if str, ok := envs["ENV_INPUT_DISKIO_INTERVAL"]; ok {
		da, err := time.ParseDuration(str)
		if err != nil {
			l.Warnf("parse ENV_INPUT_DISKIO_INTERVAL to time.Duration: %s, ignore", err)
		} else {
			i.Interval.Duration = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}

	if str, ok := envs["ENV_INPUT_DISKIO_DEVICES"]; ok {
		arrays := strings.Split(str, ",")
		l.Debugf("add ENV_INPUT_DISKIO_DEVICES from ENV: %v", arrays)
		i.Devices = append(i.Devices, arrays...)
	}

	if str, ok := envs["ENV_INPUT_DISKIO_DEVICE_TAGS"]; ok {
		arrays := strings.Split(str, ",")
		l.Debugf("add ENV_INPUT_DISKIO_DEVICE_TAGS from ENV: %v", arrays)
		i.DeviceTags = append(i.DeviceTags, arrays...)
	}

	if str, ok := envs["ENV_INPUT_DISKIO_NAME_TEMPLATES"]; ok {
		arrays := strings.Split(str, ",")
		l.Debugf("add ENV_INPUT_DISKIO_NAME_TEMPLATES from ENV: %v", arrays)
		i.NameTemplates = append(i.NameTemplates, arrays...)
	}
}

func (i *Input) diskName(devName string) (string, []string) {
	devName = "/dev/" + devName

	di, err := i.diskInfo(devName)

	devLinks := strings.Split(di["DEVLINKS"], " ")

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

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			diskIO:   disk.IOCounters,
			Interval: datakit.Duration{Duration: time.Second * 10},

			semStop: cliutils.NewSem(),
			Tags:    make(map[string]string),
		}
	})
}
