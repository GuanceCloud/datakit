package hostobject

import (
	"encoding/json"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l = logger.DefaultSLogger(InputName)
)

type Input struct {
	Name  string //deprecated
	Class string //deprecated
	Desc  string `toml:"description,omitempty"` //deprecated

	Interval datakit.Duration
	Pipeline string            `toml:"pipeline"`
	Tags     map[string]string `toml:"tags,omitempty"`

	IgnoreInputsErrorsBefore datakit.Duration `toml:"ignore_inputs_errors_before,omitempty"`
	IOTimeout                datakit.Duration `toml:"io_timeout,omitempty"`

	EnableNetVirtualInterfaces bool     `toml:"enable_net_virtual_interfaces"`
	IgnoreFS                   []string `toml:"ignore_fs"`

	CloudInfo map[string]string `toml:"cloud_info"`

	p *pipeline.Pipeline

	collectData *hostMeasurement
}

func (_ *Input) Catalog() string {
	return InputCat
}

func (_ *Input) SampleConfig() string {
	return SampleConfig
}

func (r *Input) PipelineConfig() map[string]string {
	return map[string]string{
		InputName: pipelineSample,
	}
}

const (
	maxInterval            = 30 * time.Minute
	minInterval            = 1 * time.Minute
	hostObjMeasurementName = "HOST"
)

func (c *Input) Run() {

	l = logger.SLogger(InputName)

	c.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, c.Interval.Duration)
	c.p = c.getPipeline()
	tick := time.NewTicker(c.Interval.Duration)
	n := 0
	defer tick.Stop()

	l.Debugf("starting %s(interval: %v)...", InputName, c.Interval)

	c.singleCollect(n) // 1st shot on datakit startup

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Infof("%s exit on sem", InputName)
			return
		case <-tick.C:
			l.Debugf("start %d collecting...", n)
			c.singleCollect(n)
			n++
		}
	}
}

func (c *Input) singleCollect(n int) {

	l.Debugf("start %d collecting...", n)

	start := time.Now()
	if err := c.Collect(); err != nil {
		io.FeedLastError(InputName, err.Error())
	} else {
		if err := inputs.FeedMeasurement(InputName,
			datakit.Object,
			[]inputs.Measurement{c.collectData},
			&io.Option{CollectCost: time.Since(start)}); err != nil {
			io.FeedLastError(InputName, err.Error())
		}
	}
}

type hostMeasurement struct {
	name   string
	fields map[string]interface{}
	tags   map[string]string
	ts     time.Time
}

func (x *hostMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: hostObjMeasurementName,
		Desc: "主机对象数据采集如下数据",
		Fields: map[string]interface{}{
			"message":          &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "主机所有信息汇总"},
			"os":               &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "主机操作系统类型"},
			"start_time":       &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "主机启动时间（Unix 时间戳）"},
			"datakit_ver":      &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "采集器版本"},
			"cpu_usage":        &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "CPU 使用率"},
			"mem_used_percent": &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.Percent, Desc: "内存使用率"},
			"load":             &inputs.FieldInfo{Type: inputs.Gauge, DataType: inputs.Float, Unit: inputs.UnknownUnit, Desc: "系统负载"},
			"state":            &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "主机状态"},
		},
	}
}

func (x *hostMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(x.name, x.tags, x.fields)
}

func (c *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&hostMeasurement{},
	}
}

func (c *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (c *Input) Collect() error {

	message, err := c.getHostObjectMessage()
	if err != nil {
		return err
	}

	messageData, err := json.Marshal(message)
	if err != nil {
		l.Errorf("json marshal err:%s", err.Error())
		return err
	}

	c.collectData = &hostMeasurement{
		name: hostObjMeasurementName,
		fields: map[string]interface{}{
			"message":          string(messageData),
			"os":               message.Host.HostMeta.OS,
			"start_time":       message.Host.HostMeta.BootTime,
			"datakit_ver":      git.Version,
			"cpu_usage":        message.Host.cpuPercent,
			"mem_used_percent": message.Host.Mem.usedPercent,
			"load":             message.Host.load5,
			"state":            "online",
		},

		tags: map[string]string{
			"name": message.Host.HostMeta.HostName,
		},
	}

	// append extra cloud fields: all of them as tags
	for k, v := range message.Host.cloudInfo {
		switch tv := v.(type) {
		case string:
			if tv != Unavailable {
				c.collectData.tags[k] = tv
			}
		default:
			l.Warnf("ignore non-string cloud extra field: %s: %v, ignored", k, v)
		}
	}

	// merge custom tags: if conflict with fields, ignore the tag
	for k, v := range c.Tags {
		if _, ok := c.collectData.fields[k]; !ok {
			c.collectData.fields[k] = v
		} else {
			l.Warnf("ignore tag %s: %s", k, v)
		}
	}

	if c.p != nil {
		if result, err := c.p.Run(string(messageData)).Result(); err == nil {
			for k, v := range result {
				c.collectData.fields[k] = v
			}
		} else {
			l.Warnf("pipeline error: %s, ignored", err)
		}
	}

	return nil
}

func (c *Input) getPipeline() *pipeline.Pipeline {

	fname := c.Pipeline
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

func init() {
	inputs.Add(InputName, func() inputs.Input {
		return &Input{
			Interval:                 datakit.Duration{Duration: 5 * time.Minute},
			IgnoreInputsErrorsBefore: datakit.Duration{Duration: 30 * time.Minute},
			IOTimeout:                datakit.Duration{Duration: 10 * time.Second},
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
	})
}
