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
	LogFiles      []string `toml:"logfiles"`
	Sockets       []string `toml:"sockets,omitempty"`
	Ignore        []string `toml:"ignore"`
	FromBeginning bool     `toml:"from_beginning,omitempty"`
	MaxOpenFiles  int      `toml:"max_open_files"`
	IgnoreDeadLog string   `toml:"ignore_dead_log"`

	Source                string   `toml:"source"`
	Service               string   `toml:"service"`
	Pipeline              string   `toml:"pipeline"`
	StorageIndex          string   `toml:"storage_index"`
	IgnoreStatus          []string `toml:"ignore_status"`
	FieldWhitelist        []string `toml:"field_white_list"`
	CharacterEncoding     string   `toml:"character_encoding"`
	RemoveAnsiEscapeCodes bool     `toml:"remove_ansi_escape_codes"`

	MultilineMatch             string   `toml:"multiline_match"`
	AutoMultilineDetection     bool     `toml:"auto_multiline_detection"`
	AutoMultilineExtraPatterns []string `toml:"auto_multiline_extra_patterns"`

	Tags map[string]string `toml:"tags"`
	Mode string            `toml:"mode,omitempty"`

	MinFlushInterval         time.Duration `toml:"-"`
	MaxMultilineLifeDuration time.Duration `toml:"-"`

	DeprecatedBlockingMode    bool   `toml:"blocking_mode"`
	DeprecatedEnableDiskCache bool   `toml:"enable_diskcache,omitempty"`
	DeprecatedPipeline        string `toml:"pipeline_path"`
	DeprecatedMultilineMatch  string `toml:"match"`
	DeprecatedMaximumLength   int    `toml:"maximum_length,omitempty"`

	processors []LogProcessor
	inputName  string
	semStop    *cliutils.Sem
	tagger     datakit.GlobalTagger
}

var l = logger.DefaultSLogger(inputName)

type LogProcessor interface {
	Start()
	Close()
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	ipt.handleDeprecatedConfig()
	fieldWhitelist := ipt.parseFieldWhitelist()
	ignoreDuration := ipt.parseIgnoreDuration()

	opts := ipt.buildTailerOptions(fieldWhitelist, ignoreDuration)
	multilinePatterns := ipt.setupMultilinePatterns()
	opts = append(opts, tailer.WithMultilinePatterns(multilinePatterns))

	ipt.startFileTailer(opts)
	ipt.startSocketLogger(opts)

	ipt.startProcessors()

	for {
		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Infof("logging input %s exiting", ipt.inputName)
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Infof("logging input %s terminated", ipt.inputName)
			return
		}
	}
}

func (ipt *Input) exit() {
	for _, processor := range ipt.processors {
		processor.Close()
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

func defaultInput() *Input {
	return &Input{
		Tags:      make(map[string]string),
		inputName: inputName,
		tagger:    datakit.DefaultGlobalTagger(),
		semStop:   cliutils.NewSem(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})

	inputs.Add(deprecatedInputName, func() inputs.Input {
		return &Input{
			AutoMultilineDetection: true,
			Tags:                   make(map[string]string),
			inputName:              deprecatedInputName,
			tagger:                 datakit.DefaultGlobalTagger(),
			semStop:                cliutils.NewSem(),
		}
	})
}

func (ipt *Input) handleDeprecatedConfig() {
	if ipt.Pipeline == "" && ipt.DeprecatedPipeline != "" {
		ipt.Pipeline = path.Base(ipt.DeprecatedPipeline)
	}

	if ipt.MultilineMatch == "" && ipt.DeprecatedMultilineMatch != "" {
		ipt.MultilineMatch = ipt.DeprecatedMultilineMatch
	}
}

func (ipt *Input) parseFieldWhitelist() []string {
	fieldWhitelist := []string{}
	if str := os.Getenv("ENV_LOGGING_FIELD_WHITE_LIST"); str != "" {
		if err := json.Unmarshal([]byte(str), &fieldWhitelist); err != nil {
			l.Warnf("failed to parse ENV_LOGGING_FIELD_WHITE_LIST: %v, ignoring", err)
		}
	}
	return fieldWhitelist
}

func (ipt *Input) parseIgnoreDuration() time.Duration {
	if dur, err := timex.ParseDuration(ipt.IgnoreDeadLog); err == nil {
		return dur
	}
	return 0
}

func (ipt *Input) buildTailerOptions(fieldWhitelist []string, ignoreDuration time.Duration) []tailer.Option {
	opts := []tailer.Option{
		tailer.WithStorageIndex(ipt.StorageIndex),
		tailer.WithIgnorePatterns(ipt.Ignore),
		tailer.WithSource(ipt.Source),
		tailer.WithService(ipt.Service),
		tailer.WithPipeline(ipt.Pipeline),
		tailer.EnableDebugFields(config.Cfg.EnableDebugFields),
		tailer.WithSockets(ipt.Sockets),
		tailer.WithIgnoredStatuses(ipt.IgnoreStatus),
		tailer.WithMaxOpenFiles(ipt.MaxOpenFiles),
		tailer.WithFromBeginning(ipt.FromBeginning),
		tailer.WithCharacterEncoding(ipt.CharacterEncoding),
		tailer.WithIgnoreDeadLog(ignoreDuration),
		tailer.EnableMultiline(ipt.AutoMultilineDetection),
		tailer.WithMaxMultilineLength(int64(float64(config.Cfg.Dataway.MaxRawBodySize) * 0.8)),
		tailer.WithExtraTags(inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")),
		tailer.WithRemoveAnsiEscapeCodes(ipt.RemoveAnsiEscapeCodes),
	}

	if len(fieldWhitelist) != 0 {
		opts = append(opts, tailer.WithFieldWhitelist(fieldWhitelist))
	} else if len(ipt.FieldWhitelist) != 0 {
		opts = append(opts, tailer.WithFieldWhitelist(ipt.FieldWhitelist))
	}

	switch ipt.Mode {
	case "docker":
		opts = append(opts, tailer.WithTextParserMode(tailer.DockerJSONLogMode))
	case "cri":
		opts = append(opts, tailer.WithTextParserMode(tailer.CriLogdMode))
	default:
		opts = append(opts, tailer.WithTextParserMode(tailer.FileMode))
	}

	return opts
}

func (ipt *Input) setupMultilinePatterns() []string {
	var multilinePatterns []string
	if ipt.MultilineMatch != "" {
		multilinePatterns = []string{ipt.MultilineMatch}
	} else if ipt.AutoMultilineDetection {
		multilinePatterns = ipt.AutoMultilineExtraPatterns
		multilinePatterns = append(multilinePatterns, multiline.GlobalPatterns...)
		l.Debugf("auto-multiline enabled for source %s, patterns: %v", ipt.Source, ipt.AutoMultilineExtraPatterns)
	}
	return multilinePatterns
}

func (ipt *Input) startFileTailer(opts []tailer.Option) {
	if len(ipt.LogFiles) == 0 {
		return
	}

	if tailer, err := tailer.NewTailer(ipt.LogFiles, opts...); err != nil {
		l.Errorf("failed to create tailer for files %v: %v", ipt.LogFiles, err)
	} else {
		l.Infof("started file tailing for %v", ipt.LogFiles)
		ipt.processors = append(ipt.processors, tailer)
	}
}

func (ipt *Input) startSocketLogger(opts []tailer.Option) {
	if len(ipt.Sockets) == 0 {
		return
	}

	if socketLogger, err := tailer.NewSocketLogging(opts...); err != nil {
		l.Errorf("failed to create socket logger for %v: %v", ipt.Sockets, err)
	} else {
		l.Infof("started socket logging on %v", ipt.Sockets)
		ipt.processors = append(ipt.processors, socketLogger)
	}
}

func (ipt *Input) startProcessors() {
	if len(ipt.processors) == 0 {
		l.Warnf("no logging processors configured")
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_logging"})
	for _, processor := range ipt.processors {
		func(processor LogProcessor) {
			g.Go(func(ctx context.Context) error {
				processor.Start()
				return nil
			})
		}(processor)
	}
}
