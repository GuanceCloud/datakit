// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package logger

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func newCustomizeRootLogger(level string, options int, ws io.Writer) (*zap.Logger, error) {
	// use lumberjack.Logger for rotate
	w := zapcore.AddSync(ws)

	cfg := zapcore.EncoderConfig{
		NameKey:    NameKeyMod,
		MessageKey: NameKeyMsg,
		LevelKey:   NameKeyLevel,
		TimeKey:    NameKeyTime,
		CallerKey:  NameKeyPos,

		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeCaller: zapcore.FullCallerEncoder,
	}

	if options&OPT_COLOR != 0 {
		cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if options&OPT_SHORT_CALLER != 0 {
		cfg.EncodeCaller = zapcore.ShortCallerEncoder
	}

	var enc zapcore.Encoder
	if options&OPT_ENC_CONSOLE != 0 {
		enc = zapcore.NewConsoleEncoder(cfg)
	} else {
		enc = zapcore.NewJSONEncoder(cfg)
	}

	var lvl zapcore.Level
	switch strings.ToLower(level) {
	case INFO: // pass
		lvl = zap.InfoLevel
	case DEBUG:
		lvl = zap.DebugLevel
	case WARN:
		lvl = zap.WarnLevel
	case ERROR:
		lvl = zap.ErrorLevel
	case PANIC:
		lvl = zap.PanicLevel
	case DPANIC:
		lvl = zap.DPanicLevel
	case FATAL:
		lvl = zap.FatalLevel
	default:
		lvl = zap.DebugLevel
	}

	core := zapcore.NewCore(enc, w, lvl)
	// NOTE: why need add another option while
	// zapcore.ShortCallerEncoder/FullCallerEncoder been set
	l := zap.New(core, zap.AddCaller())
	return l, nil
}

func newOnlyMessageRootLogger(ws io.Writer) (*zap.Logger, error) {
	// use lumberjack.Logger for rotate
	w := zapcore.AddSync(ws)

	cfg := zapcore.EncoderConfig{
		MessageKey: NameKeyMsg,

		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeCaller: zapcore.FullCallerEncoder,
	}

	enc := zapcore.NewConsoleEncoder(cfg)
	lvl := zap.DebugLevel

	core := zapcore.NewCore(enc, w, lvl)
	// NOTE: why need add another option while
	// zapcore.ShortCallerEncoder/FullCallerEncoder been set
	l := zap.New(core, zap.AddCaller())
	return l, nil
}

func newMultiWriterRootLogger(fpath, errorPath, level string, options int) (*zap.Logger, error) {
	// Create main log writer
	mainWriter := &lumberjack.Logger{
		Filename:   fpath,
		MaxSize:    MaxSize,
		MaxBackups: MaxBackups,
		MaxAge:     MaxAge,
	}

	// Create error log writer
	errorWriter := &lumberjack.Logger{
		Filename:   errorPath,
		MaxSize:    MaxSize,
		MaxBackups: MaxBackups,
		MaxAge:     MaxAge,
	}

	// Ensure error log directory exists
	if err := os.MkdirAll(filepath.Dir(errorPath), 0o600); err != nil {
		return nil, err
	}

	cfg := zapcore.EncoderConfig{
		NameKey:    NameKeyMod,
		MessageKey: NameKeyMsg,
		LevelKey:   NameKeyLevel,
		TimeKey:    NameKeyTime,
		CallerKey:  NameKeyPos,

		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeCaller: zapcore.FullCallerEncoder,
	}

	if options&OPT_COLOR != 0 {
		cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if options&OPT_SHORT_CALLER != 0 {
		cfg.EncodeCaller = zapcore.ShortCallerEncoder
	}

	var enc zapcore.Encoder
	if options&OPT_ENC_CONSOLE != 0 {
		enc = zapcore.NewConsoleEncoder(cfg)
	} else {
		enc = zapcore.NewJSONEncoder(cfg)
	}

	var lvl zapcore.Level
	switch strings.ToLower(level) {
	case INFO:
		lvl = zap.InfoLevel
	case DEBUG:
		lvl = zap.DebugLevel
	case WARN:
		lvl = zap.WarnLevel
	case ERROR:
		lvl = zap.ErrorLevel
	case PANIC:
		lvl = zap.PanicLevel
	case DPANIC:
		lvl = zap.DPanicLevel
	case FATAL:
		lvl = zap.FatalLevel
	default:
		lvl = zap.DebugLevel
	}

	// Combine cores with different level enablers
	// Main file gets all logs up to configured level
	// Error file gets only error+ level logs
	core := zapcore.NewTee(
		zapcore.NewCore(enc, zapcore.AddSync(mainWriter), zap.NewAtomicLevelAt(lvl)),                 // Main file
		zapcore.NewCore(enc, zapcore.AddSync(errorWriter), zap.NewAtomicLevelAt(zapcore.ErrorLevel)), // Error file
	)

	l := zap.New(core, zap.AddCaller())
	return l, nil
}
