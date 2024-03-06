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
	"github.com/GuanceCloud/cliutils/point"
	"github.com/shirou/gopsutil/disk"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export/doc"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
	inputName   = "diskio"
)

var (
	_        inputs.ReadEnv = (*Input)(nil)
	l                       = logger.DefaultSLogger(inputName)
	varRegex                = regexp.MustCompile(`\$(?:\w+|\{\w+\})`)
)

//nolint:unused,structcheck
type diskInfoCache struct {
	// Unix Nano timestamp of the last modification of the device.
	// This value is used to invalidate the cache
	modifiedAt int64

	udevDataPath string
	values       map[string]string
}

type Input struct {
	Interval         time.Duration
	Devices          []string
	DeviceTags       []string
	NameTemplates    []string
	SkipSerialNumber bool
	Tags             map[string]string

	collectCache []*point.Point
	lastStat     map[string]disk.IOCountersStat
	lastTime     time.Time
	diskIO       DiskIO
	feeder       dkio.Feeder
	mergedTags   map[string]string
	tagger       datakit.GlobalTagger

	infoCache    map[string]diskInfoCache //nolint:structcheck,unused
	deviceFilter *DevicesFilter

	semStop *cliutils.Sem // start stop signal
}

func (ipt *Input) Run() {
	ipt.setup()

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
			if err := ipt.feeder.Feed(inputName, point.Metric, ipt.collectCache,
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
	// set disk device filter
	ipt.deviceFilter = &DevicesFilter{}
	err := ipt.deviceFilter.Compile(ipt.Devices)
	if err != nil {
		return err
	}

	// disk io stat
	diskio, err := ipt.diskIO([]string{}...)
	if err != nil {
		return fmt.Errorf("error getting disk io info: %w", err)
	}

	ts := time.Now()
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(ts))
	for _, stat := range diskio {
		var kvs point.KVs

		match := false

		// match disk name
		if len(ipt.deviceFilter.filters) < 1 || ipt.deviceFilter.Match(stat.Name) {
			match = true
		}

		tagsName, devLinks := ipt.diskName(stat.Name)
		kvs = kvs.AddTag("name", tagsName)

		if !match {
			for _, devLink := range devLinks {
				if ipt.deviceFilter.Match(devLink) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		for t, v := range ipt.diskTags(stat.Name) {
			kvs = kvs.AddTag(t, v)
		}

		if !ipt.SkipSerialNumber {
			if len(stat.SerialNumber) != 0 {
				kvs = kvs.AddTag("serial", stat.SerialNumber)
			} else {
				kvs = kvs.AddTag("serial", "unknown")
			}
		}

		kvs = kvs.Add("reads", stat.ReadCount, false, true)
		kvs = kvs.Add("writes", stat.WriteCount, false, true)
		kvs = kvs.Add("read_bytes", stat.ReadBytes, false, true)
		kvs = kvs.Add("write_bytes", stat.WriteBytes, false, true)
		kvs = kvs.Add("read_time", stat.ReadTime, false, true)
		kvs = kvs.Add("write_time", stat.WriteTime, false, true)
		kvs = kvs.Add("io_time", stat.IoTime, false, true)
		kvs = kvs.Add("weighted_io_time", stat.WeightedIO, false, true)
		kvs = kvs.Add("iops_in_progress", stat.IopsInProgress, false, true)
		kvs = kvs.Add("merged_reads", stat.MergedReadCount, false, true)
		kvs = kvs.Add("merged_writes", stat.MergedWriteCount, false, true)

		if ipt.lastStat != nil {
			deltaTime := ts.Unix() - ipt.lastTime.Unix()
			if v, ok := ipt.lastStat[stat.Name]; ok && deltaTime > 0 {
				if stat.ReadBytes >= v.ReadBytes {
					kvs = kvs.Add("read_bytes/sec", int64(stat.ReadBytes-v.ReadBytes)/deltaTime, false, true)
				}
				if stat.WriteBytes >= v.WriteBytes {
					kvs = kvs.Add("write_bytes/sec", int64(stat.WriteBytes-v.WriteBytes)/deltaTime, false, true)
				}
			}
		}

		for k, v := range ipt.mergedTags {
			kvs = kvs.AddTag(k, v)
		}

		ipt.collectCache = append(ipt.collectCache, point.NewPointV2(inputName, kvs, opts...))
	}
	ipt.lastStat = diskio
	ipt.lastTime = ts
	return nil
}

func (ipt *Input) diskTags(devName string) map[string]string {
	if len(ipt.DeviceTags) == 0 {
		return nil
	}

	di, err := ipt.diskInfo(devName)
	if err != nil {
		l.Warnf("Error gathering disk info: %s", err)
		return nil
	}

	tags := map[string]string{}
	for _, dt := range ipt.DeviceTags {
		if v, ok := di[dt]; ok {
			tags[dt] = v
		}
	}

	return tags
}

func (ipt *Input) Singleton() {}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}
func (*Input) Catalog() string      { return "host" }
func (*Input) SampleConfig() string { return sampleCfg }
func (*Input) AvailableArchs() []string {
	return []string{
		datakit.OSLabelLinux, datakit.OSLabelWindows,
		datakit.LabelK8s, datakit.LabelDocker,
	}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&docMeasurement{},
	}
}

func (ipt *Input) GetENVDoc() []*inputs.ENVInfo {
	// nolint:lll
	infos := []*inputs.ENVInfo{
		{FieldName: "Interval"},
		{FieldName: "Devices", Type: doc.List, Example: "`^sda\\d,^sdb\\d,vd.*`", Desc: "Setting interfaces using regular expressions will collect these expected devices", DescZh: "使用正则表达式设置接口将收集这些预期的设备"},
		{FieldName: "DeviceTags", Type: doc.List, Example: `ID_FS_TYPE,ID_FS_USAGE`, Desc: "Device metadata added tags", DescZh: "设备附加标签"},
		{FieldName: "NameTemplates", Type: doc.List, Example: `$ID_FS_LABEL,$DM_VG_NAME/$DM_LV_NAME`, Desc: "Using the same metadata source as device_tags", DescZh: "使用与 device_ tags 相同的元数据源"},
		{FieldName: "SkipSerialNumber", Type: doc.Boolean, Default: `false`, Desc: "disk serial number is not required", DescZh: "不需要磁盘序列号"},
		{FieldName: "Tags"},
	}

	return doc.SetENVDoc("ENV_INPUT_DISKIO_", infos)
}

// ReadEnv support envs：
//
//	ENV_INPUT_DISKIO_SKIP_SERIAL_NUMBER : booler
//	ENV_INPUT_DISKIO_TAGS : "a=b,c=d"
//	ENV_INPUT_DISKIO_INTERVAL : time.Duration
//	ENV_INPUT_DISKIO_DEVICES : []string
//	ENV_INPUT_DISKIO_DEVICE_TAGS : []string
//	ENV_INPUT_DISKIO_NAME_TEMPLATES : []string
func (ipt *Input) ReadEnv(envs map[string]string) {
	if skip, ok := envs["ENV_INPUT_DISKIO_SKIP_SERIAL_NUMBER"]; ok {
		b, err := strconv.ParseBool(skip)
		if err != nil {
			l.Warnf("parse ENV_INPUT_DISKIO_SKIP_SERIAL_NUMBER to bool: %s, ignore", err)
		} else {
			ipt.SkipSerialNumber = b
		}
	}

	if tagsStr, ok := envs["ENV_INPUT_DISKIO_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
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
			ipt.Interval = config.ProtectedInterval(minInterval,
				maxInterval,
				da)
		}
	}

	if str, ok := envs["ENV_INPUT_DISKIO_DEVICES"]; ok {
		arrays := strings.Split(str, ",")
		l.Debugf("add ENV_INPUT_DISKIO_DEVICES from ENV: %v", arrays)
		ipt.Devices = append(ipt.Devices, arrays...)
	}

	if str, ok := envs["ENV_INPUT_DISKIO_DEVICE_TAGS"]; ok {
		arrays := strings.Split(str, ",")
		l.Debugf("add ENV_INPUT_DISKIO_DEVICE_TAGS from ENV: %v", arrays)
		ipt.DeviceTags = append(ipt.DeviceTags, arrays...)
	}

	if str, ok := envs["ENV_INPUT_DISKIO_NAME_TEMPLATES"]; ok {
		arrays := strings.Split(str, ",")
		l.Debugf("add ENV_INPUT_DISKIO_NAME_TEMPLATES from ENV: %v", arrays)
		ipt.NameTemplates = append(ipt.NameTemplates, arrays...)
	}
}

func (ipt *Input) diskName(devName string) (string, []string) {
	devName = "/dev/" + devName

	di, err := ipt.diskInfo(devName)

	devLinks := strings.Split(di["DEVLINKS"], " ")

	if err != nil {
		l.Warnf("Error gathering disk info: %s", err)
		return devName, devLinks
	}

	// diskInfo empty
	if len(ipt.NameTemplates) == 0 || len(di) == 0 {
		return devName, devLinks
	}

	// render name templates
	for _, nt := range ipt.NameTemplates {
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

func newDefaultInput() *Input {
	ipt := &Input{
		diskIO:   disk.IOCounters,
		Interval: time.Second * 10,
		Tags:     make(map[string]string),

		feeder: dkio.DefaultFeeder(),

		semStop:    cliutils.NewSem(),
		tagger:     datakit.DefaultGlobalTagger(),
		mergedTags: make(map[string]string),
	}
	return ipt
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return newDefaultInput()
	})
}
