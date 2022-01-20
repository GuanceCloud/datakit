package logger

import (
	"net/url"
	"os"
	"runtime"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	STDOUT = "stdout" // log output to stdout
)

func newNormalRootLogger(fpath, level string, options int) (*zap.Logger, error) {
	cfg := &zap.Config{
		Encoding: `json`,
		EncoderConfig: zapcore.EncoderConfig{
			NameKey:    NameKeyMod,
			MessageKey: NameKeyMsg,
			LevelKey:   NameKeyLevel,
			TimeKey:    NameKeyTime,
			CallerKey:  NameKeyPos,

			EncodeLevel:  zapcore.CapitalLevelEncoder,
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeCaller: zapcore.FullCallerEncoder,
		},
	}

	if fpath != "" {
		cfg.OutputPaths = []string{fpath}

		if runtime.GOOS == "windows" { // See: https://github.com/uber-go/zap/issues/621
			if err := zap.RegisterSink("winfile", newWinFileSink); err != nil {
				return nil, err
			}

			cfg.OutputPaths = []string{
				"winfile:///" + fpath,
			}
		}
	}

	switch strings.ToLower(level) {
	case DEBUG:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case INFO:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case WARN:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case ERROR:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case PANIC:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.PanicLevel)
	case DPANIC:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DPanicLevel)
	case FATAL:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}

	if options&OPT_COLOR != 0 {
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if options&OPT_ENC_CONSOLE != 0 {
		cfg.Encoding = "console"
	}

	if options&OPT_SHORT_CALLER != 0 {
		cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	if options&OPT_STDOUT != 0 || fpath == "" { // if no log file path set, default set to stdout
		cfg.OutputPaths = append(cfg.OutputPaths, STDOUT)
	}

	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return l, nil
}

func newWinFileSink(u *url.URL) (zap.Sink, error) {
	return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
}
