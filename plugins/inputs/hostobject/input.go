// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package hostobject collect host object.
package hostobject

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/GuanceCloud/cliutils/point"
	dto "github.com/prometheus/client_model/go"
	diskutil "github.com/shirou/gopsutil/disk"
	netutil "github.com/shirou/gopsutil/net"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	_ inputs.ReadEnv   = (*Input)(nil)
	_ inputs.Singleton = (*Input)(nil)
	l                  = logger.DefaultSLogger(InputName)
)

type (
	NetIOCounters  func(bool) ([]netutil.IOCountersStat, error)
	DiskIOCounters func(names ...string) (map[string]diskutil.IOCountersStat, error)
)

type Input struct {
	Name  string `toml:"name,omitempty"`        // deprecated
	Class string `toml:"class,omitempty"`       // deprecated
	Desc  string `toml:"description,omitempty"` // deprecated

	PipelineDeprecated string `toml:"pipeline,omitempty"`

	Tags map[string]string `toml:"tags,omitempty"`

	Interval                 *datakit.Duration `toml:"interval,omitempty"`
	IgnoreInputsErrorsBefore *datakit.Duration `toml:"ignore_inputs_errors_before,omitempty"`
	DeprecatedIOTimeout      *datakit.Duration `toml:"io_timeout,omitempty"`

	EnableNetVirtualInterfaces bool     `toml:"enable_net_virtual_interfaces"`
	IgnoreZeroBytesDisk        bool     `toml:"ignore_zero_bytes_disk"`
	OnlyPhysicalDevice         bool     `toml:"only_physical_device"`
	ExtraDevice                []string `toml:"extra_device"`
	ExcludeDevice              []string `toml:"exclude_device"`

	DisableCloudProviderSync bool              `toml:"disable_cloud_provider_sync"`
	CloudInfo                map[string]string `toml:"cloud_info,omitempty"`
	lastSync                 time.Time

	netIOCounters  NetIOCounters
	diskIOCounters DiskIOCounters
	lastDiskIOInfo diskIOInfo
	lastNetIOInfo  netIOInfo

	collectData *hostMeasurement

	semStop    *cliutils.Sem // start stop signal
	isTestMode bool
	feeder     dkio.Feeder
	opt        point.Option

	mfs []*dto.MetricFamily
}

func (ipt *Input) Singleton() {
}

func (ipt *Input) Catalog() string {
	return InputCat
}

func (ipt *Input) SampleConfig() string {
	return SampleConfig
}

const (
	maxInterval            = 30 * time.Minute
	minInterval            = 10 * time.Second
	hostObjMeasurementName = "HOST"
)

func (ipt *Input) Run() {
	l = logger.SLogger(InputName)

	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	l.Debugf("starting %s(interval: %v)...", InputName, ipt.Interval)

	// no election.
	ipt.opt = point.WithExtraTags(dkpt.GlobalHostTags())

	for {
		l.Debugf("start collecting...")
		start := time.Now()
		if err := ipt.doCollect(); err != nil {
			ipt.feeder.FeedLastError(InputName, err.Error(), point.Object)
		} else if err := ipt.feeder.Feed(InputName,
			point.Object, []*point.Point{ipt.collectData.Point()},
			&dkio.Option{CollectCost: time.Since(start)}); err != nil {
			ipt.feeder.FeedLastError(InputName, err.Error(), point.Object)
		}

		select {
		case <-datakit.Exit.Wait():
			l.Infof("%s exit on sem", InputName)
			return

		case <-ipt.semStop.Wait():
			l.Infof("%s return on sem", InputName)
			return

		case <-tick.C:
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

// ReadEnv used to read ENVs while running under DaemonSet.
func (ipt *Input) ReadEnv(envs map[string]string) {
	if enable, ok := envs["ENV_INPUT_HOSTOBJECT_ENABLE_NET_VIRTUAL_INTERFACES"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOSTOBJECT_ENABLE_NET_VIRTUAL_INTERFACES to bool: %s, ignore", err)
		} else {
			ipt.EnableNetVirtualInterfaces = b
		}
	}

	if _, ok := envs["ENV_INPUT_HOSTOBJECT_ONLY_PHYSICAL_DEVICE"]; ok {
		l.Info("setup OnlyPhysicalDevice...")
		ipt.OnlyPhysicalDevice = true
	}
	if fsList, ok := envs["ENV_INPUT_HOSTOBJECT_EXTRA_DEVICE"]; ok {
		list := strings.Split(fsList, ",")
		l.Debugf("add extra_device from ENV: %v", fsList)
		ipt.ExtraDevice = append(ipt.ExtraDevice, list...)
	}
	if fsList, ok := envs["ENV_INPUT_HOSTOBJECT_EXCLUDE_DEVICE"]; ok {
		list := strings.Split(fsList, ",")
		l.Debugf("add exlude_device from ENV: %v", fsList)
		ipt.ExcludeDevice = append(ipt.ExcludeDevice, list...)
	}
	// https://gitlab.jiagouyun.com/cloudcare-tools/datakit/-/issues/505
	if enable, ok := envs["ENV_INPUT_HOSTOBJECT_ENABLE_ZERO_BYTES_DISK"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOSTOBJECT_ENABLE_ZERO_BYTES_DISK to bool: %s, ignore", err)
		} else {
			ipt.IgnoreZeroBytesDisk = b
		}
	}

	if tagsStr, ok := envs["ENV_INPUT_HOSTOBJECT_TAGS"]; ok {
		tags := config.ParseGlobalTags(tagsStr)
		for k, v := range tags {
			ipt.Tags[k] = v
		}
	}

	// ENV_CLOUD_PROVIDER 会覆盖 ENV_INPUT_HOSTOBJECT_TAGS 中填入的 cloud_provider
	if tagsStr, ok := envs["ENV_CLOUD_PROVIDER"]; ok {
		cloudProvider := dkstring.TrimString(tagsStr)
		cloudProvider = strings.ToLower(cloudProvider)
		switch cloudProvider {
		case "aliyun", "tencent", "aws", "hwcloud", "azure":
			ipt.Tags["cloud_provider"] = cloudProvider
		}
	} // ENV_CLOUD_PROVIDER
}

type hostMeasurement struct {
	name   string
	fields map[string]interface{}
	tags   map[string]string
	ts     time.Time
	ipt    *Input
}

// Point implement MeasurementV2.
func (m *hostMeasurement) Point() *point.Point {
	opts := point.DefaultObjectOptions()
	opts = append(opts, point.WithTime(m.ts), m.ipt.opt)

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

//nolint:lll
func (*hostMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: hostObjMeasurementName,
		Desc: "Host object metrics",
		Tags: map[string]interface{}{
			"host": &inputs.TagInfo{Desc: "Hostname. Required."},
			"name": &inputs.TagInfo{Desc: "Hostname"},
			"os":   &inputs.TagInfo{Desc: "Host OS type"},
		},
		Fields: map[string]interface{}{
			"message":                    &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Summary of all host information"},
			"start_time":                 &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationMS, Desc: "Host startup time (Unix timestamp)"},
			"datakit_ver":                &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "collector version"},
			"cpu_usage":                  &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "CPU usage"},
			"mem_used_percent":           &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "memory usage"},
			"load":                       &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "system load"},
			"state":                      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Host Status"},
			"disk_used_percent":          &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "disk usage"},
			"diskio_read_bytes_per_sec":  &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.BytesPerSec, Desc: "disk read rate"},
			"diskio_write_bytes_per_sec": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.BytesPerSec, Desc: "disk write rate"},
			"net_recv_bytes_per_sec":     &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.BytesPerSec, Desc: "network receive rate"},
			"net_send_bytes_per_sec":     &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Int, Unit: inputs.BytesPerSec, Desc: "network send rate"},
			"logging_level":              &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "log level"},
		},
	}
}

func (*hostMeasurement) LineProto() (*dkpt.Point, error) {
	// return dkpt.NewPoint(hm.name, hm.tags, hm.fields, dkpt.OOpt())
	return nil, fmt.Errorf("not implement")
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&hostMeasurement{},
	}
}

func (ipt *Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (ipt *Input) doCollect() error {
	if mfs, err := metrics.Gather(); err == nil {
		ipt.mfs = mfs
	}

	message, err := ipt.getHostObjectMessage()
	if err != nil {
		return err
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		l.Errorf("json marshal err:%s", err.Error())
		return err
	}

	l.Debugf("messageData len: %d", len(messageData))

	ipt.collectData = &hostMeasurement{
		name: hostObjMeasurementName,
		fields: map[string]interface{}{
			"message":                    string(messageData),
			"start_time":                 message.Host.HostMeta.BootTime * 1000,
			"datakit_ver":                datakit.Version,
			"cpu_usage":                  message.Host.cpuPercent,
			"mem_used_percent":           message.Host.Mem.usedPercent,
			"load":                       message.Host.load5,
			"state":                      "online",
			"disk_used_percent":          message.Host.diskUsedPercent,
			"diskio_read_bytes_per_sec":  message.Host.diskIOReadBytesPerSec,
			"diskio_write_bytes_per_sec": message.Host.diskIOWriteBytesPerSec,
			"net_recv_bytes_per_sec":     message.Host.netRecvBytesPerSec,
			"net_send_bytes_per_sec":     message.Host.netSendBytesPerSec,
			"logging_level":              message.Host.loggingLevel,
		},

		tags: map[string]string{
			"name": message.Host.HostMeta.HostName,
			"os":   message.Host.HostMeta.OS,
		},

		ts:  time.Now(),
		ipt: ipt,
	}

	if !ipt.isTestMode {
		ipt.collectData.fields["Scheck"] = message.Collectors[0].Version
	}

	// append extra cloud fields: all of them as tags
	for k, v := range message.Host.cloudInfo {
		switch tv := v.(type) {
		case string:
			if tv != Unavailable {
				ipt.collectData.tags[k] = tv
			}
		default:
			l.Warnf("ignore non-string cloud extra field: %s: %v, ignored", k, v)
		}
	}

	// merge custom tags: if conflict with fields, ignore the tag
	for k, v := range ipt.Tags {
		// 添加的 tag key 不能存在已有的 field key 中
		if _, ok := ipt.collectData.fields[k]; ok {
			l.Warnf("ignore tag `%s', exists in field", k)
			continue
		}

		// 用户 tag 无脑添加 tag(可能覆盖已有 tag)
		ipt.collectData.tags[k] = v
	}

	return nil
}

func (ipt *Input) Collect() (map[string][]*dkpt.Point, error) {
	ipt.isTestMode = true
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	if err := ipt.doCollect(); err != nil {
		return nil, err
	}

	var pts []*dkpt.Point
	if pt, err := ipt.collectData.LineProto(); err != nil {
		return nil, err
	} else {
		pts = append(pts, pt)
	}

	mpts := make(map[string][]*dkpt.Point)
	mpts[datakit.Object] = pts

	return mpts, nil
}

func defaultInput() *Input {
	return &Input{
		Interval:                 &datakit.Duration{Duration: 5 * time.Minute},
		IgnoreInputsErrorsBefore: &datakit.Duration{Duration: 30 * time.Second},
		IgnoreZeroBytesDisk:      true,
		diskIOCounters:           diskutil.IOCounters,
		netIOCounters:            netutil.IOCounters,

		semStop: cliutils.NewSem(),
		feeder:  dkio.DefaultFeeder(),
		Tags:    make(map[string]string),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(InputName, func() inputs.Input {
		return defaultInput()
	})
}

func SetLog() {
	l = logger.SLogger(InputName)
}
