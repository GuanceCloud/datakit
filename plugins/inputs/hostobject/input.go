// Package hostobject collect host object.
package hostobject

import (
	"encoding/json"
	"strconv"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

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
	IgnoreFS                   []string `toml:"ignore_fs"`

	CloudInfo map[string]string `toml:"cloud_info,omitempty"`

	p *pipeline.Pipeline

	collectData *hostMeasurement
}

func (*Input) Catalog() string {
	return InputCat
}

func (*Input) SampleConfig() string {
	return SampleConfig
}

func (*Input) PipelineConfig() map[string]string {
	return map[string]string{
		InputName: pipelineSample,
	}
}

const (
	maxInterval            = 30 * time.Minute
	minInterval            = 1 * time.Minute
	hostObjMeasurementName = "HOST"
)

func (*Input) RunPipeline() {
	// TODO.
}

func (i *Input) Run() {
	l = logger.SLogger(InputName)

	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)
	i.p = i.getPipeline()
	tick := time.NewTicker(i.Interval.Duration)
	n := 0
	defer tick.Stop()

	l.Debugf("starting %s(interval: %v)...", InputName, i.Interval)

	i.singleCollect(n) // 1st shot on datakit startup

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Infof("%s exit on sem", InputName)
			return
		case <-tick.C:
			l.Debugf("start %d collecting...", n)
			i.singleCollect(n)
			n++
		}
	}
}

// ReadEnv support envs：
//   ENV_INPUT_HOSTOBJECT_ENABLE_NET_VIRTUAL_INTERFACES: booler
func (i *Input) ReadEnv(envs map[string]string) {
	if enable, ok := envs["ENV_INPUT_HOSTOBJECT_ENABLE_NET_VIRTUAL_INTERFACES"]; ok {
		b, err := strconv.ParseBool(enable)
		if err != nil {
			l.Warnf("parse ENV_INPUT_HOSTOBJECT_ENABLE_NET_VIRTUAL_INTERFACES to bool: %s, ignore", err)
		} else {
			i.EnableNetVirtualInterfaces = b
		}
	}
}

func (i *Input) singleCollect(n int) {
	l.Debugf("start %d collecting...", n)

	start := time.Now()
	if err := i.Collect(); err != nil {
		io.FeedLastError(InputName, err.Error())
	} else if err := inputs.FeedMeasurement(InputName,
		datakit.Object,
		[]inputs.Measurement{i.collectData},
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
func (*hostMeasurement) Info() *inputs.MeasurementInfo {
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

func (i *hostMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(i.name, i.tags, i.fields)
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&hostMeasurement{},
	}
}

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) Collect() error {
	message, err := i.getHostObjectMessage()
	if err != nil {
		return err
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		l.Errorf("json marshal err:%s", err.Error())
		return err
	}

	i.collectData = &hostMeasurement{
		name: hostObjMeasurementName,
		fields: map[string]interface{}{
			"message":          string(messageData),
			"start_time":       message.Host.HostMeta.BootTime,
			"datakit_ver":      datakit.Version,
			"cpu_usage":        message.Host.cpuPercent,
			"mem_used_percent": message.Host.Mem.usedPercent,
			"load":             message.Host.load5,
			"state":            "online",
			"Scheck":           message.Collectors[0].Version,
		},

		tags: map[string]string{
			"name": message.Host.HostMeta.HostName,
			"os":   message.Host.HostMeta.OS,
		},
	}

	// append extra cloud fields: all of them as tags
	for k, v := range message.Host.cloudInfo {
		switch tv := v.(type) {
		case string:
			if tv != Unavailable {
				i.collectData.tags[k] = tv
			}
		default:
			l.Warnf("ignore non-string cloud extra field: %s: %v, ignored", k, v)
		}
	}

	// merge custom tags: if conflict with fields, ignore the tag
	for k, v := range i.Tags {
		// 添加的 tag key 不能存在已有的 field key 中
		if _, ok := i.collectData.fields[k]; ok {
			l.Warnf("ignore tag `%s', exists in field", k)
			continue
		}

		// 用户 tag 无脑添加 tag(可能覆盖已有 tag)
		i.collectData.tags[k] = v
	}

	if i.p != nil {
		if result, err := i.p.Run(string(messageData)).Result(); err == nil {
			for k, v := range result {
				i.collectData.fields[k] = v
			}
		} else {
			l.Warnf("pipeline error: %s, ignored", err)
		}
	}

	return nil
}

func (i *Input) getPipeline() *pipeline.Pipeline {
	fname := i.Pipeline
	if fname == "" {
		fname = InputName + ".p"
	}

	p, err := pipeline.NewPipelineByScriptPath(fname)
	if err != nil {
		l.Warnf("%s", err)
		return nil
	}

	return p
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
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(InputName, func() inputs.Input {
		return DefaultHostObject()
	})
}

func SetLog() {
	l = logger.SLogger("hostobject")
}
