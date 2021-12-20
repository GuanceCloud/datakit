// Package logging collects host logging data.
package logging

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName           = "logging"
	deprecatedInputName = "tailf"

	sampleCfg = `
[[inputs.logging]]
  ## required
  logfiles = [
    "/var/log/syslog",
    "/var/log/message",
  ]

  ## glob filteer
  ignore = [""]

  ## your logging source, if it's empty, use 'default'
  source = ""

  ## add service tag, if it's empty, use $source.
  service = ""

  ## grok pipeline script path
  pipeline = ""

  ## optional status:
  ##   "emerg","alert","critical","error","warning","info","debug","OK"
  ignore_status = []

  ## optional encodings:
  ##    "utf-8", "utf-16le", "utf-16le", "gbk", "gb18030" or ""
  character_encoding = ""

  ## The pattern should be a regexp. Note the use of '''this regexp'''
  ## regexp link: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
  # multiline_match = '''^\S'''

  ## removes ANSI escape codes from text strings
  remove_ansi_escape_codes = false

  [inputs.logging.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
)

type Input struct {
	LogFiles              []string          `toml:"logfiles"`
	Ignore                []string          `toml:"ignore"`
	Source                string            `toml:"source"`
	Service               string            `toml:"service"`
	Pipeline              string            `toml:"pipeline"`
	IgnoreStatus          []string          `toml:"ignore_status"`
	CharacterEncoding     string            `toml:"character_encoding"`
	MultilineMatch        string            `toml:"multiline_match"`
	MultilineMaxLines     int               `toml:"multiline_maxlines"`
	RemoveAnsiEscapeCodes bool              `toml:"remove_ansi_escape_codes"`
	Tags                  map[string]string `toml:"tags"`
	FromBeginning         bool              `toml:"-"`

	DeprecatedPipeline       string `toml:"pipeline_path"`
	DeprecatedMultilineMatch string `toml:"match"`
	DeprecatedFromBeginning  bool   `toml:"from_beginning"`

	tailer *tailer.Tailer

	// 在输出 log 内容时，区分是 tailf 还是 logging
	inputName string

	semStop *cliutils.Sem // start stop signal
}

var l = logger.DefaultSLogger(inputName)

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	// 兼容旧版配置 pipeline_path
	if ipt.Pipeline == "" && ipt.DeprecatedPipeline != "" {
		ipt.Pipeline = ipt.DeprecatedPipeline
	}

	if ipt.MultilineMatch == "" && ipt.DeprecatedMultilineMatch != "" {
		ipt.MultilineMatch = ipt.DeprecatedMultilineMatch
	}

	var pipelinePath string

	if ipt.Pipeline == "" {
		ipt.Pipeline = ipt.Source + ".p"
	}

	pipelinePath, err := config.GetPipelinePath(ipt.Pipeline)
	if err != nil {
		l.Error(err)
		return
	}

	opt := &tailer.Option{
		Source:                ipt.Source,
		Service:               ipt.Service,
		Pipeline:              pipelinePath,
		IgnoreStatus:          ipt.IgnoreStatus,
		FromBeginning:         ipt.FromBeginning,
		CharacterEncoding:     ipt.CharacterEncoding,
		MultilineMatch:        ipt.MultilineMatch,
		MultilineMaxLines:     ipt.MultilineMaxLines,
		RemoveAnsiEscapeCodes: ipt.RemoveAnsiEscapeCodes,
		GlobalTags:            ipt.Tags,
	}

	ipt.tailer, err = tailer.NewTailer(ipt.LogFiles, opt, ipt.Ignore)
	if err != nil {
		l.Error(err)
		return
	}

	go ipt.tailer.Start()

	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Infof("%s exit", ipt.inputName)
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Infof("%s return", ipt.inputName)
			return
		}
	}
}

func (ipt *Input) exit() {
	ipt.Stop()
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) Stop() {
	ipt.tailer.Close()
}

func (*Input) Catalog() string {
	return "log"
}

func (*Input) SampleConfig() string {
	return sampleCfg
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&loggingMeasurement{},
	}
}

type loggingMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (ipt *loggingMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(ipt.name, ipt.tags, ipt.fields, ipt.ts)
}

//nolint:lll
func (*loggingMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "logging 日志采集",
		Desc: "使用配置文件中的 `source` 字段值，如果该值为空，则默认为 `default`",
		Tags: map[string]interface{}{
			"filename": inputs.NewTagInfo(`此条日志来源的文件名，仅为基础文件名，并非带有全路径`),
			"host":     inputs.NewTagInfo(`主机名`),
			"service":  inputs.NewTagInfo("service 名称，对应配置文件中的 `service` 字段值"),
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志正文，默认存在，可以使用 pipeline 删除此字段"},
			"status":  &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志状态，默认为 `info`，采集器会该字段做支持映射，映射表见上述 pipelie 配置和使用"},
		},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Tags:      make(map[string]string),
			inputName: inputName,

			semStop: cliutils.NewSem(),
		}
	})
	inputs.Add(deprecatedInputName, func() inputs.Input {
		return &Input{
			Tags:      make(map[string]string),
			inputName: deprecatedInputName,

			semStop: cliutils.NewSem(),
		}
	})
}
