// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package logging collects host logging data.
package logging

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
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

  ## Only retain the fields specified in the whitelist.
  field_white_list = []

  ## optional encodings:
  ##    "utf-8", "utf-16le", "utf-16be", "gbk", "gb18030" or ""
  character_encoding = ""

  ## The pattern should be a regexp. Note the use of '''this regexp'''.
  ## regexp link: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
  # multiline_match = '''^\S'''

  auto_multiline_detection = true
  auto_multiline_extra_patterns = []

  ## Removes ANSI escape codes from text strings.
  remove_ansi_escape_codes = false

  ## The maximum allowed number of open files, default is 500. If it is -1, it means no limit.
  # max_open_files = 500

  ## If file is inactive, it is ignored.
  ## time units are "ms", "s", "m", "h"
  ignore_dead_log = "12h"

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
	FieldWhiteList             []string          `toml:"field_white_list"`
	CharacterEncoding          string            `toml:"character_encoding"`
	MultilineMatch             string            `toml:"multiline_match"`
	AutoMultilineDetection     bool              `toml:"auto_multiline_detection"`
	AutoMultilineExtraPatterns []string          `toml:"auto_multiline_extra_patterns"`
	RemoveAnsiEscapeCodes      bool              `toml:"remove_ansi_escape_codes"`
	Tags                       map[string]string `toml:"tags"`
	FromBeginning              bool              `toml:"from_beginning,omitempty"`
	MaxOpenFiles               int               `toml:"max_open_files"`
	IgnoreDeadLog              string            `toml:"ignore_dead_log"`

	MinFlushInterval         time.Duration `toml:"-"`
	MaxMultilineLifeDuration time.Duration `toml:"-"`
	Mode                     string        `toml:"mode,omitempty"`

	DeprecatedBlockingMode    bool   `toml:"blocking_mode"`
	DeprecatedEnableDiskCache bool   `toml:"enable_diskcache,omitempty"`
	DeprecatedPipeline        string `toml:"pipeline_path"`
	DeprecatedMultilineMatch  string `toml:"match"`
	DeprecatedMaximumLength   int    `toml:"maximum_length,omitempty"`

	process []LogProcessor
	// 在输出 log 内容时，区分是 tailf 还是 logging
	inputName string

	semStop *cliutils.Sem // start stop signal
	Tagger  datakit.GlobalTagger
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

	fieldWhiteList := []string{}
	if str := os.Getenv("ENV_LOGGING_FIELD_WHITE_LIST"); str != "" {
		if err := json.Unmarshal([]byte(str), &fieldWhiteList); err != nil {
			l.Warnf("parse ENV_INPUT_LOGGING_FIELD_WHITE_LIST to slice: %s, ignore", err)
		}
	}

	var ignoreDuration time.Duration
	if dur, err := timex.ParseDuration(ipt.IgnoreDeadLog); err == nil {
		ignoreDuration = dur
	}

	opts := []tailer.Option{
		tailer.WithIgnorePatterns(ipt.Ignore),
		tailer.WithSource(ipt.Source),
		tailer.WithService(ipt.Service),
		tailer.WithPipeline(ipt.Pipeline),
		tailer.EnableDebugFields(config.Cfg.EnableDebugFields),
		tailer.WithSockets(ipt.Sockets),
		tailer.WithIgnoreStatus(ipt.IgnoreStatus),
		tailer.WithFieldWhiteList(ipt.FieldWhiteList),
		tailer.WithMaxOpenFiles(ipt.MaxOpenFiles),
		tailer.WithFromBeginning(ipt.FromBeginning),
		tailer.WithCharacterEncoding(ipt.CharacterEncoding),
		tailer.WithIgnoreDeadLog(ignoreDuration),
		tailer.EnableMultiline(ipt.AutoMultilineDetection),
		tailer.WithMaxMultilineLength(int64(float64(config.Cfg.Dataway.MaxRawBodySize) * 0.8)),
		tailer.WithGlobalTags(inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, "")),
		tailer.WithRemoveAnsiEscapeCodes(ipt.RemoveAnsiEscapeCodes),
		tailer.WithFieldWhiteList(fieldWhiteList),
	}

	switch ipt.Mode {
	case "docker":
		opts = append(opts, tailer.WithTextParserMode(tailer.DockerJSONLogMode))
	case "cri":
		opts = append(opts, tailer.WithTextParserMode(tailer.CriLogdMode))
	default:
		opts = append(opts, tailer.WithTextParserMode(tailer.FileMode))
	}

	var multilinePatterns []string
	if ipt.MultilineMatch != "" {
		multilinePatterns = []string{ipt.MultilineMatch}
	} else if ipt.AutoMultilineDetection {
		multilinePatterns = ipt.AutoMultilineExtraPatterns
		multilinePatterns = append(multilinePatterns, multiline.GlobalPatterns...)
		l.Debugf("source %s automatic-multiline on, patterns %v", ipt.Source, ipt.AutoMultilineExtraPatterns)
	}
	opts = append(opts, tailer.WithMultilinePatterns(multilinePatterns))

	if len(ipt.LogFiles) != 0 {
		tailerL, err := tailer.NewTailer(ipt.LogFiles, opts...)
		if err != nil {
			l.Error(err)
		} else {
			ipt.process = append(ipt.process, tailerL)
		}
	} else {
		if len(ipt.Sockets) != 0 {
			socker, err := tailer.NewSocketLogWithOptions(opts...)
			if err != nil {
				l.Error(err)
			} else {
				l.Infof("new socket logging")
				ipt.process = append(ipt.process, socker)
			}
		} else {
			l.Warn("socket len=0")
		}
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
	if ipt.process != nil {
		for _, proce := range ipt.process {
			proce.Close()
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
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

type loggingMeasurement struct{}

//nolint:lll
func (*loggingMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:           "default",
		Cat:            point.Logging,
		MetaDuplicated: true, // input `tailf' and `logging' are the same.
		Desc:           "Use the `source` of the config，if empty then use `default`",
		Tags: map[string]interface{}{
			"host":     inputs.NewTagInfo(`Host name`),
			"service":  inputs.NewTagInfo("The name of the service, if `service` is empty then use `source`."),
			"filepath": inputs.NewTagInfo("The filepath to the log file on the host system where the log is stored."),
		},
		Fields: map[string]interface{}{
			"message":         &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The text of the logging."},
			"status":          &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "The status of the logging, default is `info`[^1]."},
			"log_read_lines":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The lines of the read file."},
			"log_file_inode":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The inode of the log file, which uniquely identifies it on the file system (requires enabling the global configuration `enable_debug_fields`)."},
			"log_read_offset": &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The current offset in the log file where reading has occurred, used to track progress during log collection (requires enabling the global configuration `enable_debug_fields`)."},
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
				Desc:     "Built-in extension fields added by server. The `date` field is set to the time when the log is collected by the collector by default, but it can be overridden using a Pipeline.",
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
			Tagger:    datakit.DefaultGlobalTagger(),
			semStop:   cliutils.NewSem(),
		}
	})

	inputs.Add(deprecatedInputName, func() inputs.Input {
		return &Input{
			Tags:      make(map[string]string),
			inputName: deprecatedInputName,
			Tagger:    datakit.DefaultGlobalTagger(),
			semStop:   cliutils.NewSem(),
		}
	})
}
