// Package logging collects host logging data.
package logging

import (
	"path/filepath"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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
}

var l = logger.DefaultSLogger(inputName)

func (*Input) RunPipeline() {
	// nil
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)

	// 兼容旧版配置 pipeline_path
	if i.Pipeline == "" && i.DeprecatedPipeline != "" {
		i.Pipeline = i.DeprecatedPipeline
	}

	if i.MultilineMatch == "" && i.DeprecatedMultilineMatch != "" {
		i.MultilineMatch = i.DeprecatedMultilineMatch
	}

	var pipelinePath string

	if i.Pipeline == "" {
		pipelinePath = filepath.Join(datakit.PipelineDir, i.Source+".p")
	} else {
		pipelinePath = filepath.Join(datakit.PipelineDir, i.Pipeline)
	}

	opt := &tailer.Option{
		Source:                i.Source,
		Service:               i.Service,
		Pipeline:              pipelinePath,
		IgnoreStatus:          i.IgnoreStatus,
		FromBeginning:         i.FromBeginning,
		CharacterEncoding:     i.CharacterEncoding,
		MultilineMatch:        i.MultilineMatch,
		MultilineMaxLines:     i.MultilineMaxLines,
		RemoveAnsiEscapeCodes: i.RemoveAnsiEscapeCodes,
		GlobalTags:            i.Tags,
	}

	var err error
	i.tailer, err = tailer.NewTailer(i.LogFiles, opt, i.Ignore)
	if err != nil {
		l.Error(err)
		return
	}

	go i.tailer.Start()

	// 阻塞在此，用以关闭 tailer 资源
	<-datakit.Exit.Wait()
	i.Stop()
	l.Infof("%s exit", i.inputName)
}

func (i *Input) Stop() {
	i.tailer.Close() //nolint:errcheck
}

func (i *Input) PipelineConfig() map[string]string {
	return nil
}

func (i *Input) Catalog() string {
	return "log"
}

func (i *Input) SampleConfig() string {
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

func (i *loggingMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(i.name, i.tags, i.fields, i.ts)
}

//nolint:lll
func (i *loggingMeasurement) Info() *inputs.MeasurementInfo {
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
		}
	})
	inputs.Add(deprecatedInputName, func() inputs.Input {
		return &Input{
			Tags:      make(map[string]string),
			inputName: deprecatedInputName,
		}
	})
}
