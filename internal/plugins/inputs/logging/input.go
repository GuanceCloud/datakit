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
)

type Input struct {
	LogFiles                   []string          `toml:"logfiles"`
	Sockets                    []string          `toml:"sockets,omitempty"`
	Ignore                     []string          `toml:"ignore"`
	Source                     string            `toml:"source"`
	StorageIndex               string            `toml:"storage_index"`
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
		tailer.WithStorageIndex(ipt.StorageIndex),
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
		if t, err := tailer.NewTailer(ipt.LogFiles, opts...); err != nil {
			l.Errorf("NewTailer(%q): %s", ipt.LogFiles, err.Error())
		} else {
			l.Infof("add tailf on %q", ipt.LogFiles)
			ipt.process = append(ipt.process, t)
		}
	}

	// we can start socket loggong the same time while file logging enbled.
	if len(ipt.Sockets) != 0 {
		if sock, err := tailer.NewSocketLogging(opts...); err != nil {
			l.Errorf("NewSocketLogging: %s", err.Error())
		} else {
			l.Infof("start socket logging on %q", ipt.Sockets)
			ipt.process = append(ipt.process, sock)
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
