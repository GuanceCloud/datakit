package logging

import (
	"path/filepath"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName           = "logging"
	deprecatedInputName = "tailf"

	sampleCfg = `
[[inputs.logging]]
  ## required, glob logfiles
  logfiles = ["/var/log/syslog"]

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
  match = '''^\S'''

  [inputs.logging.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
)

type Input struct {
	LogFiles                []string          `toml:"logfiles"`
	Ignore                  []string          `toml:"ignore"`
	Source                  string            `toml:"source"`
	Service                 string            `toml:"service"`
	Pipeline                string            `toml:"pipeline"`
	DeprecatedPipeline      string            `toml:"pipeline_path"`
	DeprecatedFromBeginning bool              `toml:"from_beginning"`
	IgnoreStatus            []string          `toml:"ignore_status"`
	CharacterEncoding       string            `toml:"character_encoding"`
	Match                   string            `toml:"match"`
	Tags                    map[string]string `toml:"tags"`
	FromBeginning           bool              `toml:"-"`

	tailer *inputs.Tailer

	// 在输出 log 内容时，区分是 tailf 还是 logging
	inputName string
}

var l = logger.DefaultSLogger(inputName)

// TODO
func (*Input) RunPipeline() {
}

func (this *Input) Run() {
	l = logger.SLogger(inputName)

	// 兼容旧版配置 pipeline_path
	if this.Pipeline == "" && this.DeprecatedPipeline != "" {
		this.Pipeline = this.DeprecatedPipeline
	}

	if this.Pipeline == "" {
		this.Pipeline = filepath.Join(datakit.PipelineDir, this.Source+".p")
	} else {
		this.Pipeline = filepath.Join(datakit.PipelineDir, this.Pipeline)
	}

	option := inputs.TailerOption{
		Files:             this.LogFiles,
		IgnoreFiles:       this.Ignore,
		Source:            this.Source,
		Service:           this.Service,
		Pipeline:          this.Pipeline,
		IgnoreStatus:      this.IgnoreStatus,
		FromBeginning:     this.FromBeginning,
		CharacterEncoding: this.CharacterEncoding,
		Match:             this.Match,
		Tags:              this.Tags,
	}

	var err error
	this.tailer, err = inputs.NewTailer(&option)
	if err != nil {
		l.Error(err)
		return
	}

	go this.tailer.Run()

	for {
		// 阻塞在此，用以关闭 tailer 资源
		select {
		case <-datakit.Exit.Wait():
			this.Stop()
			l.Infof("%s exit", this.inputName)
			return
		}
	}
}

func (this *Input) Stop() {
	this.tailer.Close()
}

func (this *Input) PipelineConfig() map[string]string {
	return nil
}

func (this *Input) Catalog() string {
	return "log"
}

func (this *Input) SampleConfig() string {
	return sampleCfg
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLinux}
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

func (this *loggingMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(this.name, this.tags, this.fields, this.ts)
}

func (this *loggingMeasurement) Info() *inputs.MeasurementInfo {
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

func init() {
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
