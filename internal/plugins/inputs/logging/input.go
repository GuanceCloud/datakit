// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package logging collects host logging data.
package logging

import (
	"context"
	"path"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/multiline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	timex "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
)

const (
	inputName           = "logging"
	deprecatedInputName = "tailf"

	sampleCfg = `
[[inputs.logging]]
  ## Required
  ## File names or a pattern to tail.
  logfiles = [
    "/var/log/syslog",
    "/var/log/messages",
  ]

  # Only two protocols are supported:TCP and UDP.
  # sockets = [
  #	 "tcp://0.0.0.0:9530",
  #	 "udp://0.0.0.0:9531",
  # ]
  ## glob filteer
  ignore = [""]

  ## Your logging source, if it's empty, use 'default'.
  source = ""

  ## Add service tag, if it's empty, use $source.
  service = ""

  ## Grok pipeline script name.
  pipeline = ""

  ## optional status:
  ##   "emerg","alert","critical","error","warning","info","debug","OK"
  ignore_status = []

  ## optional encodings:
  ##    "utf-8", "utf-16le", "utf-16le", "gbk", "gb18030" or ""
  character_encoding = ""

  ## The pattern should be a regexp. Note the use of '''this regexp'''.
  ## regexp link: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
  # multiline_match = '''^\S'''

  auto_multiline_detection = true
  auto_multiline_extra_patterns = []

  ## Removes ANSI escape codes from text strings.
  remove_ansi_escape_codes = false

  ## If the data sent failure, will retry forevery.
  blocking_mode = true

  ## If file is inactive, it is ignored.
  ## time units are "ms", "s", "m", "h"
  ignore_dead_log = "1h"

  ## Read file from beginning.
  from_beginning = false

  [inputs.logging.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
)

type Input struct {
	LogFiles                   []string          `toml:"logfiles"`
	Sockets                    []string          `toml:"sockets,omitempty"`
	Ignore                     []string          `toml:"ignore"`
	Source                     string            `toml:"source"`
	Service                    string            `toml:"service"`
	Pipeline                   string            `toml:"pipeline"`
	IgnoreStatus               []string          `toml:"ignore_status"`
	CharacterEncoding          string            `toml:"character_encoding"`
	MultilineMatch             string            `toml:"multiline_match"`
	AutoMultilineDetection     bool              `toml:"auto_multiline_detection"`
	AutoMultilineExtraPatterns []string          `toml:"auto_multiline_extra_patterns"`
	RemoveAnsiEscapeCodes      bool              `toml:"remove_ansi_escape_codes"`
	Tags                       map[string]string `toml:"tags"`
	BlockingMode               bool              `toml:"blocking_mode"`
	FromBeginning              bool              `toml:"from_beginning,omitempty"`
	DockerMode                 bool              `toml:"docker_mode,omitempty"`
	IgnoreDeadLog              string            `toml:"ignore_dead_log"`
	MinFlushInterval           time.Duration     `toml:"-"`
	MaxMultilineLifeDuration   time.Duration     `toml:"-"`

	DeprecatedEnableDiskCache bool   `toml:"enable_diskcache,omitempty"`
	DeprecatedPipeline        string `toml:"pipeline_path"`
	DeprecatedMultilineMatch  string `toml:"match"`
	DeprecatedMaximumLength   int    `toml:"maximum_length,omitempty"`

	process []LogProcessor
	// 在输出 log 内容时，区分是 tailf 还是 logging
	inputName string

	semStop *cliutils.Sem // start stop signal
}

var l = logger.DefaultSLogger(inputName)

type LogProcessor interface {
	Start()
	Close()
	// add func
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	// 兼容旧版配置 pipeline_path
	if ipt.Pipeline == "" && ipt.DeprecatedPipeline != "" {
		ipt.Pipeline = path.Base(ipt.DeprecatedPipeline)
	}

	if ipt.MultilineMatch == "" && ipt.DeprecatedMultilineMatch != "" {
		ipt.MultilineMatch = ipt.DeprecatedMultilineMatch
	}

	var ignoreDuration time.Duration
	if dur, err := timex.ParseDuration(ipt.IgnoreDeadLog); err == nil {
		ignoreDuration = dur
	}

	opt := &tailer.Option{
		Source:            ipt.Source,
		Service:           ipt.Service,
		Pipeline:          ipt.Pipeline,
		Sockets:           ipt.Sockets,
		IgnoreStatus:      ipt.IgnoreStatus,
		FromBeginning:     ipt.FromBeginning,
		CharacterEncoding: ipt.CharacterEncoding,
		IgnoreDeadLog:     ignoreDuration,
		GlobalTags:        ipt.Tags,
		BlockingMode:      ipt.BlockingMode,
		Done:              ipt.semStop.Wait(),
	}

	if ipt.DockerMode {
		opt.Mode = tailer.DockerMode
	}

	if ipt.MultilineMatch != "" {
		opt.MultilinePatterns = []string{ipt.MultilineMatch}
	} else if ipt.AutoMultilineDetection {
		if len(ipt.AutoMultilineExtraPatterns) != 0 {
			opt.MultilinePatterns = ipt.AutoMultilineExtraPatterns
			l.Infof("source %s automatic-multiline on, patterns %v", ipt.Source, ipt.AutoMultilineExtraPatterns)
		} else {
			opt.MultilinePatterns = multiline.GlobalPatterns
			l.Info("source %s automatic-multiline on, use default patterns", ipt.Source)
		}
	}

	ipt.process = make([]LogProcessor, 0)
	if len(ipt.LogFiles) != 0 {
		tailerL, err := tailer.NewTailer(ipt.LogFiles, opt, ipt.Ignore)
		if err != nil {
			l.Error(err)
		} else {
			ipt.process = append(ipt.process, tailerL)
		}
	}

	// 互斥：只有当logFile为空，socket不为空才开启socket采集日志
	if len(ipt.LogFiles) == 0 && len(ipt.Sockets) != 0 {
		socker, err := tailer.NewWithOpt(opt, ipt.Ignore)
		if err != nil {
			l.Error(err)
		} else {
			l.Infof("new socket logging")
			ipt.process = append(ipt.process, socker)
		}
	} else {
		l.Warn("socket len=0")
	}
	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_logging"})
	if ipt.process != nil && len(ipt.process) > 0 {
		// start all process
		for _, proce := range ipt.process {
			func(proce LogProcessor) {
				g.Go(func(ctx context.Context) error {
					proce.Start()
					return nil
				})
			}(proce)
		}
	} else {
		l.Warnf("There are no logging processors here")
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Infof("%s exit", ipt.inputName)
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Infof("%s terminate", ipt.inputName)
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
	if ipt.process != nil {
		for _, proce := range ipt.process {
			proce.Close()
		}
	}
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
}

func (ipt *loggingMeasurement) LineProto() (*point.Point, error) {
	return point.NewPoint(ipt.name, ipt.tags, ipt.fields, point.LOpt())
}

//nolint:lll
func (*loggingMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "logging collect",
		Type: "logging",
		Desc: "Use the `source` of the config，if empty then use `default`",
		Tags: map[string]interface{}{
			"filename": inputs.NewTagInfo(`The base name of the file.`),
			"host":     inputs.NewTagInfo(`Host name`),
			"service":  inputs.NewTagInfo("Use the `service` of the config."),
		},
		Fields: map[string]interface{}{
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The text of the logging."},
			"status":          &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The status of the logging, default is `unknown`[^1]."},
			"log_read_lines":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The lines of the read file ([:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6))."},
			"log_read_offset": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "The offset of the read file ([:octicons-tag-24: Version-1.4.8](changelog.md#cl-1.4.8) · [:octicons-beaker-24: Experimental](index.md#experimental))."},
			"log_read_time":   &inputs.FieldInfo{DataType: inputs.DurationSecond, Unit: inputs.UnknownUnit, Desc: "The timestamp of the read file."},
			"message_length":  &inputs.FieldInfo{DataType: inputs.SizeByte, Unit: inputs.NCount, Desc: "The length of the message content."},
			"`__docid`": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "Built-in extension fields added by server. The unique identifier for a log document, typically used for sorting and viewing details",
			},
			"__namespace": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "Built-in extension fields added by server. The unique identifier for a log document dataType.",
			},
			"__truncated_id": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "Built-in extension fields added by server. If the log is particularly large (usually exceeding 1M in size), the central system will split it and add three fields: `__truncated_id`, `__truncated_count`, and `__truncated_number` to define the splitting scenario. The __truncated_id field represents the unique identifier for the split log.",
			},
			"__truncated_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.UnknownUnit,
				Desc:     "Built-in extension fields added by server. If the log is particularly large (usually exceeding 1M in size), the central system will split it and add three fields: `__truncated_id`, `__truncated_count`, and `__truncated_number` to define the splitting scenario. The __truncated_count field represents the total number of logs resulting from the split.",
			},
			"__truncated_number": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.UnknownUnit,
				Desc:     "Built-in extension fields added by server. If the log is particularly large (usually exceeding 1M in size), the central system will split it and add three fields: `__truncated_id`, `__truncated_count`, and `__truncated_number` to define the splitting scenario. The __truncated_count field represents represents the current sequential identifier for the split logs.",
			},
			"date": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.DurationMS,
				Desc:     "Built-in extension fields added by server. The `date` field is set to the time when the log is collected by the collector by default, but it can be overridden using a pipeline.",
			},
			"date_ns": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.DurationNS,
				Desc:     "Built-in extension fields added by server. The `date_ns` field is set to the millisecond part of the time when the log is collected by the collector by default. Its maximum value is 1.0E+6 and its unit is nanoseconds. It is typically used for sorting.",
			},
			"create_time": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.DurationMS,
				Desc:     "Built-in extension fields added by server. The `create_time` field represents the time when the log is written to the storage engine.",
			},
			"df_metering_size": &inputs.FieldInfo{
				DataType: inputs.Int,
				Unit:     inputs.UnknownUnit,
				Desc:     "Built-in extension fields added by server. The `df_metering_size` field is used for logging cost statistics.",
			},
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
