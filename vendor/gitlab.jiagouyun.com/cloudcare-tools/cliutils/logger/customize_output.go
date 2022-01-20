package logger

import (
	"io"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
