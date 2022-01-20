// Package hostobject collect host object.
package hostobject

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var _ inputs.ReadEnv = (*Input)(nil)

var l = logger.DefaultSLogger(InputName)

type Input struct {
	Name  string `toml:"name,omitempty"`        // deprecated
	Class string `toml:"class,omitempty"`       // deprecated
	Desc  string `toml:"description,omitempty"` // deprecated

	Pipeline string            `toml:"pipeline,omitempty"`
	Tags     map[string]string `toml:"tags,omitempty"`

	Interval                 *datakit.Duration `toml:"interval,omitempty"`
	IgnoreInputsErrorsBefore *datakit.Duration `toml:"ignore_inputs_errors_before,omitempty"`
	IOTimeout                *datakit.Duration `toml:"io_timeout,omitempty"`

	EnableNetVirtualInterfaces bool     `toml:"enable_net_virtual_interfaces"`
	IgnoreZeroBytesDisk        bool     `toml:"ignore_zero_bytes_disk"`
	IgnoreFS                   []string `toml:"ignore_fs"`

	CloudInfo map[string]string `toml:"cloud_info,omitempty"`

	p *pipeline.Pipeline

	collectData *hostMeasurement

	semStop    *cliutils.Sem // start stop signal
	isTestMode bool
}

func (ipt *Input) Catalog() string {
	return InputCat
}

func (ipt *Input) SampleConfig() string {
	return SampleConfig
}

const (
	maxInterval            = 30 * time.Minute
	minInterval            = 1 * time.Minute
	hostObjMeasurementName = "HOST"
)

func (ipt *Input) Run() {
	l = logger.SLogger(InputName)
	io.FeedEventLog(&io.Reporter{Message: "hostobject start ok, ready for collecting metrics.", Logtype: "event"})

	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	tick := time.NewTicker(ipt.Interval.Duration)
	n := 0
	defer tick.Stop()

	l.Debugf("starting %s(interval: %v)...", InputName, ipt.Interval)

	ipt.singleCollect(n) // 1st shot on datakit startup

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Infof("%s exit on sem", InputName)
			return

		case <-ipt.semStop.Wait():
			l.Infof("%s return on sem", InputName)
			return

		case <-tick.C:
			l.Debugf("start %d collecting...", n)
			ipt.singleCollect(n)
			n++
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

// ReadEnv , support envs：
//   ENV_INPUT_HOSTOBJECT_ENABLE_NET_VIRTUAL_INTERFACES: booler
func (ipt *Input) ReadEnv(envs map[string]string) {
	if enable, ok := envs["ENV_INPUT_HOSTOBJECT_ENABLE_NET_VIRTUAL_INTERFACES"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOSTOBJECT_ENABLE_NET_VIRTUAL_INTERFACES to bool: %s, ignore", err)
		} else {
			ipt.EnableNetVirtualInterfaces = b
		}
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

func (ipt *Input) singleCollect(n int) {
	l.Debugf("start %d collecting...", n)

	start := time.Now()
	if err := ipt.doCollect(); err != nil {
		io.FeedLastError(InputName, err.Error())
	} else if err := inputs.FeedMeasurement(InputName,
		datakit.Object,
		[]inputs.Measurement{ipt.collectData},
		&io.Option{CollectCost: time.Since(start)}); err != nil {
		io.FeedLastError(InputName, err.Error())
	}
}

type hostMeasurement struct {
	name   string
	fields map[string]interface{}
	tags   map[string]string
}

//nolint:lll
func (hm *hostMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: hostObjMeasurementName,
		Desc: "主机对象数据采集如下数据",
		Tags: map[string]interface{}{
			"os": &inputs.TagInfo{Desc: "主机操作系统类型"},
		},
		Fields: map[string]interface{}{
			"message":          &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "主机所有信息汇总"},
			"start_time":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "主机启动时间（Unix 时间戳）"},
			"datakit_ver":      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "采集器版本"},
			"cpu_usage":        &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "CPU 使用率"},
			"mem_used_percent": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "内存使用率"},
			"load":             &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "系统负载"},
			"state":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "主机状态"},
		},
	}
}

func (hm *hostMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(hm.name, hm.tags, hm.fields)
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&hostMeasurement{},
	}
}

func (ipt *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (ipt *Input) doCollect() error {
	message, err := ipt.getHostObjectMessage()
	if err != nil {
		return err
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		l.Errorf("json marshal err:%s", err.Error())
		return err
	}

	ipt.collectData = &hostMeasurement{
		name: hostObjMeasurementName,
		fields: map[string]interface{}{
			"message":          string(messageData),
			"start_time":       message.Host.HostMeta.BootTime,
			"datakit_ver":      datakit.Version,
			"cpu_usage":        message.Host.cpuPercent,
			"mem_used_percent": message.Host.Mem.usedPercent,
			"load":             message.Host.load5,
			"state":            "online",
		},

		tags: map[string]string{
			"name": message.Host.HostMeta.HostName,
			"os":   message.Host.HostMeta.OS,
		},
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

	if ipt.p != nil {
		if result, err := ipt.p.Run(string(messageData)).Result(); err == nil &&
			result != nil && !result.Dropped {
			for k, v := range result.Data {
				ipt.collectData.fields[k] = v
			}
			for k, v := range result.Tags {
				ipt.collectData.tags[k] = v
			}
			// ipt.collectData.tags
		} else {
			l.Debug("pipeline error: %s, ignored", err)
		}
	}

	return nil
}

func (ipt *Input) Collect() (map[string][]*io.Point, error) {
	ipt.isTestMode = true
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	if err := ipt.doCollect(); err != nil {
		return nil, err
	}

	var pts []*io.Point
	if pt, err := ipt.collectData.LineProto(); err != nil {
		return nil, err
	} else {
		pts = append(pts, pt)
	}

	mpts := make(map[string][]*io.Point)
	mpts[datakit.Object] = pts

	return mpts, nil
}

func DefaultHostObject() *Input {
	return &Input{
		Interval:                 &datakit.Duration{Duration: 5 * time.Minute},
		IgnoreInputsErrorsBefore: &datakit.Duration{Duration: 30 * time.Second},
		IOTimeout:                &datakit.Duration{Duration: 10 * time.Second},
		IgnoreFS: []string{
			"autofs",
			"tmpfs",
			"devtmpfs",
			"devfs",
			"iso9660",
			"overlay",
			"aufs",
			"squashfs",
		},
		semStop: cliutils.NewSem(),
		Tags:    make(map[string]string),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(InputName, func() inputs.Input {
		return DefaultHostObject()
	})
}

func SetLog() {
	l = logger.SLogger(InputName)
}
