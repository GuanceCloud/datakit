// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package logfwd

import (
	"errors"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

func buildTailerOptions(cfg *logConfig, fn tailer.ForwardFunc) []tailer.Option {
	opts := []tailer.Option{
		tailer.WithSource(cfg.Source),
		tailer.WithService(cfg.Service),
		tailer.WithStorageIndex(cfg.StorageIndex),
		tailer.WithCharacterEncoding(cfg.CharacterEncoding),
		tailer.WithPipeline(cfg.Pipeline),
		tailer.EnableMultiline(true),
		tailer.WithFromBeginning(cfg.FromBeginning),
		tailer.WithFileSizeThreshold(cfg.FromBeginningThresholdSize),
		tailer.WithRemoveAnsiEscapeCodes(cfg.RemoveAnsiEscapeCodes),
		tailer.WithIgnoreDeadLog(time.Hour * 12),
		tailer.WithExtraTags(cfg.Tags),
		tailer.WithForwardFunc(fn),
	}
	if cfg.autoMultiline {
		opts = append(opts, tailer.WithAutoMultilineExtraPatterns(nil))
	} else {
		opts = append(opts, tailer.WithMultilinePattern(cfg.multilinePattern))
	}

	switch cfg.Type {
	case "file":
		opts = append(opts, tailer.WithTextParserMode(tailer.FileMode))
	case runtime.DockerRuntime:
		opts = append(opts, tailer.WithTextParserMode(tailer.DockerJSONLogMode))
	default:
		opts = append(opts, tailer.WithTextParserMode(tailer.CriLogdMode))
	}

	return opts
}

type writeMessageFunc func([]byte) error

func forwardFunc(cfg *logConfig, fn writeMessageFunc) tailer.ForwardFunc {
	return func(filename, text string, fields map[string]interface{}) error {
		msg := message{
			Type:         "1",
			Source:       cfg.Source,
			StorageIndex: cfg.StorageIndex,
			Pipeline:     cfg.Pipeline,
			Log:          text,
			Tags:         make(map[string]string),
			Fields:       fields,
		}

		msg.Tags["filename"] = filename
		for k, v := range cfg.Tags {
			msg.Tags[k] = v
		}

		data, err := msg.json()
		if err != nil {
			log.Errorf("failed to marshal message: %v", err)
			return err
		}

		if err := fn(data); err != nil {
			if errors.Is(err, errWebsocketQueueFull) || errors.Is(err, errWebsocketClientClosed) {
				log.Debugf("drop log message: %v", err)
				return err
			}
			log.Errorf("failed to send message: %v", err)
			return err
		}

		return nil
	}
}
