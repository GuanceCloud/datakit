package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultTraceLoggerName = "trace_id"
	defaultSpanLoggerName  = "span_id"
)

type CtxOption interface {
	apply(config) config
}

type optionFunc func(config) config

func (fn optionFunc) apply(cfg config) config {
	return fn(cfg)
}

type Trace struct {
	TraceID string
	SpanID  string
}

type ExtractTrace func(ctx context.Context) Trace

type config struct {
	enabledTrace bool
	traceKey     string
	spanKey      string
	parseTrace   ExtractTrace
	core         zapcore.Core
	opts         []zap.Option
}

func newDefaultConfig() config {
	return config{
		traceKey: defaultTraceLoggerName,
		spanKey:  defaultSpanLoggerName,
		core:     zapcore.NewCore(zapcore.NewJSONEncoder(zapcore.EncoderConfig{MessageKey: "message"}), os.Stdout, zap.DebugLevel),
	}
}

func EnableTrace() CtxOption {
	return optionFunc(func(cfg config) config {
		cfg.enabledTrace = true
		return cfg
	})
}

func WithZapCore(core zapcore.Core) CtxOption {
	return optionFunc(func(cfg config) config {
		cfg.core = core
		return cfg
	})
}

func WithZapOpts(opts ...zap.Option) CtxOption {
	return optionFunc(func(cfg config) config {
		cfg.opts = opts
		return cfg
	})
}

func WithTraceKey(tn string, sn string) CtxOption {
	return optionFunc(func(cfg config) config {
		cfg.traceKey = tn
		cfg.spanKey = sn
		return cfg
	})
}

func WithParseTrace(f ExtractTrace) CtxOption {
	return optionFunc(func(cfg config) config {
		cfg.parseTrace = f
		return cfg
	})
}
