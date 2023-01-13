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

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/logtail/multiline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	timex "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/time"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
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
  # only two protocols are supported:TCP and UDP
  # sockets = [
  #	 "tcp://0.0.0.0:9530",
  #	 "udp://0.0.0.0:9531",
  # ]
  ## glob filteer
  ignore = [""]

  ## your logging source, if it's empty, use 'default'
  source = ""

  ## add service tag, if it's empty, use $source.
  service = ""

  ## grok pipeline script name
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

  auto_multiline_detection = true
  auto_multiline_extra_patterns = []

  ## removes ANSI escape codes from text strings
  remove_ansi_escape_codes = false

  ## if the data sent failure, will retry forevery
  blocking_mode = true

  ## if file is inactive, it is ignored
  ## time units are "ms", "s", "m", "h"
  ignore_dead_log = "1h"

  ## use disk cache of logging
  enable_diskcache = false

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
	EnableDiskCache            bool              `toml:"enable_diskcache,omitempty"`
	DockerMode                 bool              `toml:"docker_mode,omitempty"`
	IgnoreDeadLog              string            `toml:"ignore_dead_log"`
	MinFlushInterval           time.Duration     `toml:"-"`
	MaxMultilineLifeDuration   time.Duration     `toml:"-"`

	DeprecatedPipeline       string `toml:"pipeline_path"`
	DeprecatedMultilineMatch string `toml:"match"`
	DeprecatedMaximumLength  int    `toml:"maximum_length,omitempty"`

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
		Source:                ipt.Source,
		Service:               ipt.Service,
		Pipeline:              ipt.Pipeline,
		Sockets:               ipt.Sockets,
		IgnoreStatus:          ipt.IgnoreStatus,
		FromBeginning:         ipt.FromBeginning,
		CharacterEncoding:     ipt.CharacterEncoding,
		RemoveAnsiEscapeCodes: ipt.RemoveAnsiEscapeCodes,
		IgnoreDeadLog:         ignoreDuration,
		GlobalTags:            ipt.Tags,
		BlockingMode:          ipt.BlockingMode,
		EnableDiskCache:       ipt.EnableDiskCache,
		Done:                  ipt.semStop.Wait(),
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
		Name: "logging 日志采集",
		Type: "logging",
		Desc: "使用配置文件中的 `source` 字段值，如果该值为空，则默认为 `default`",
		Tags: map[string]interface{}{
			"filename": inputs.NewTagInfo(`此条日志来源的文件名，仅为基础文件名，并非带有全路径`),
			"host":     inputs.NewTagInfo(`主机名`),
			"service":  inputs.NewTagInfo("service 名称，对应配置文件中的 `service` 字段值"),
		},
		Fields: map[string]interface{}{
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志正文，默认存在，可以使用 pipeline 删除此字段"},
			"status":          &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "日志状态，默认为 `unknown`，采集器会该字段做支持映射，映射表见上述 pipelie 配置和使用[^1]"},
			"log_read_lines":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "采集到的行数计数，多行数据算成一行（[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)）"},
			"log_read_offset": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.UnknownUnit, Desc: "当前数据在文件中的偏移位置（[:octicons-tag-24: Version-1.4.8](changelog.md#cl-1.4.8) · [:octicons-beaker-24: Experimental](index.md#experimental)）"},
			"log_read_time":   &inputs.FieldInfo{DataType: inputs.DurationSecond, Unit: inputs.UnknownUnit, Desc: "数据从文件中读取到的这一刻的时间戳，单位是秒"},
			"message_length":  &inputs.FieldInfo{DataType: inputs.SizeByte, Unit: inputs.NCount, Desc: "message 字段的长度，单位字节"},
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
