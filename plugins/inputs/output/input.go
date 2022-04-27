// Package output handle multiple output data.
package output

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

//------------------------------------------------------------------------------

func (ipt *Input) Run() {
	// l = logger.SLogger(inputName)

	// // 兼容旧版配置 pipeline_path
	// if ipt.Pipeline == "" && ipt.DeprecatedPipeline != "" {
	// 	ipt.Pipeline = path.Base(ipt.DeprecatedPipeline)
	// }

	// if ipt.MultilineMatch == "" && ipt.DeprecatedMultilineMatch != "" {
	// 	ipt.MultilineMatch = ipt.DeprecatedMultilineMatch
	// }

	// var ignoreDuration time.Duration
	// if dur, err := timex.ParseDuration(ipt.IgnoreDeadLog); err == nil {
	// 	ignoreDuration = dur
	// }

	// opt := &tailer.Option{
	// 	Source:                ipt.Source,
	// 	Service:               ipt.Service,
	// 	Pipeline:              ipt.Pipeline,
	// 	Sockets:               ipt.Sockets,
	// 	IgnoreStatus:          ipt.IgnoreStatus,
	// 	FromBeginning:         ipt.FromBeginning,
	// 	CharacterEncoding:     ipt.CharacterEncoding,
	// 	MaximumLength:         ipt.MaximumLength,
	// 	MultilineMatch:        ipt.MultilineMatch,
	// 	MultilineMaxLines:     ipt.MultilineMaxLines,
	// 	RemoveAnsiEscapeCodes: ipt.RemoveAnsiEscapeCodes,
	// 	IgnoreDeadLog:         ignoreDuration,
	// 	GlobalTags:            ipt.Tags,
	// }
	// ipt.process = make([]LogProcessor, 0)
	// if len(ipt.LogFiles) != 0 {
	// 	tailerL, err := tailer.NewTailer(ipt.LogFiles, opt, ipt.Ignore)
	// 	if err != nil {
	// 		l.Error(err)
	// 	} else {
	// 		ipt.process = append(ipt.process, tailerL)
	// 	}
	// }

	// // 互斥：只有当logFile为空，socket不为空才开启socket采集日志
	// if len(ipt.LogFiles) == 0 && len(ipt.Sockets) != 0 {
	// 	socker, err := tailer.NewWithOpt(opt, ipt.Ignore)
	// 	if err != nil {
	// 		l.Error(err)
	// 	} else {
	// 		l.Infof("new socket logging")
	// 		ipt.process = append(ipt.process, socker)
	// 	}
	// } else {
	// 	l.Warn("socket len =0")
	// }

	// if ipt.process != nil && len(ipt.process) > 0 {
	// 	// start all process
	// 	for _, proce := range ipt.process {
	// 		go proce.Start()
	// 	}
	// } else {
	// 	l.Warnf("There are no logging processors here")
	// }

	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info(inputName + " exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Infof(inputName + " return")
			return
		}
	}
}

func (ipt *Input) exit() {
	// ipt.Stop()
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

// func (ipt *Input) Stop() {
// 	// if ipt.process != nil {
// 	// 	for _, proce := range ipt.process {
// 	// 		proce.Close()
// 	// 	}
// 	// }
// }

//------------------------------------------------------------------------------

const (
	inputName = "output"

	sampleCfg = `
[[inputs.output]]
  # listen address, with protocol schema and port
  listen = "tcp://0.0.0.0:5044"

  ## action type, for now, we support: logstash
  action_type = "logstash"

  ## source, if it's empty, use 'default'
  source = ""

  ## add service tag, if it's empty, use $source.
  service = ""

  ## grok pipeline script name
  pipeline = ""

  [inputs.output.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
)

type Input struct {
	Listen     string            `toml:"listen"`
	ActionType string            `toml:"action_type"`
	Source     string            `toml:"source"`
	Service    string            `toml:"service"`
	Pipeline   string            `toml:"pipeline"`
	Tags       map[string]string `toml:"tags"`

	semStop *cliutils.Sem // start stop signal
}

var l = logger.DefaultSLogger(inputName)

func (*Input) Catalog() string { return "log" }

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) AvailableArchs() []string { return datakit.AllArch }

//------------------------------------------------------------------------------

type loggingMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&loggingMeasurement{},
	}
}

func (ipt *loggingMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(ipt.name, ipt.tags, ipt.fields, ipt.ts)
}

//nolint:lll
func (*loggingMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "接收器",
		Type: "logging",
		Desc: "使用配置文件中的 `source` 字段值，如果该值为空，则默认为 `default`",
		Tags: map[string]interface{}{
			"filepath": inputs.NewTagInfo(`此条记录来源的文件名，全路径`), // log.file.path
			"host":     inputs.NewTagInfo(`主机名`),            // host.name
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "记录正文，默认存在，可以使用 pipeline 删除此字段"}, // message
		},
	}
}

//------------------------------------------------------------------------------

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Tags:    make(map[string]string),
			semStop: cliutils.NewSem(),
		}
	})
}
