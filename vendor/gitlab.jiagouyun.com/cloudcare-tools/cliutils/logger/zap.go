package logger

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	OPT_ENC_CONSOLE  = 1 // non-json
	OPT_SHORT_CALLER = 2
	OPT_SHUGAR       = 4

	DEBUG = "debug"
	INFO  = "info"
)

var (
	defaultRootLogger *zap.Logger
)

func SetGlobalRootLogger(fpath, level string, options int) {
	if defaultRootLogger != nil {
		panic(fmt.Sprintf("global root logger has been initialized: %+#v", defaultRootLogger))
	}

	var err error
	defaultRootLogger, err = NewRootLogger(fpath, level, options)
	if err != nil {
		panic(err)
	}
}

func Logger(name string) *zap.Logger {
	return GetLogger(defaultRootLogger, name)
}

func SLogger(name string) *zap.SugaredLogger {
	return GetSugarLogger(defaultRootLogger, name)
}

func GetLogger(root *zap.Logger, name string) *zap.Logger {
	return root.Named(name)
}

func GetSugarLogger(root *zap.Logger, name string) *zap.SugaredLogger {
	return root.Sugar().Named(name)
}

func NewRootLogger(fpath, level string, options int) (*zap.Logger, error) {

	cfg := &zap.Config{
		Encoding:    `json`,
		OutputPaths: []string{fpath},
		EncoderConfig: zapcore.EncoderConfig{
			NameKey:    "MOD",
			MessageKey: "MSG",

			LevelKey:    "LEV",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "TS",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey:    "POS",
			EncodeCaller: zapcore.FullCallerEncoder,
		},
	}

	switch strings.ToLower(level) {
	case DEBUG:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case INFO:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}

	if options&OPT_ENC_CONSOLE != 0 {
		cfg.Encoding = "console"
	}

	if options&OPT_SHORT_CALLER != 0 {
		cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return l, nil
}
