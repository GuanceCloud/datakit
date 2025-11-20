// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"

	"github.com/GuanceCloud/cliutils"
)

const (
	// 禁用 JSON 形式输出.
	OPT_ENC_CONSOLE = 1 //nolint:golint,stylecheck
	// 显示代码路径时，不显示全路径.
	OPT_SHORT_CALLER = 2 //nolint:stylecheck,golint
	// 日志写到 stdout.
	OPT_STDOUT = 4 //nolint:stylecheck,golint
	// 日志内容中追加颜色.
	OPT_COLOR = 8 //nolint:stylecheck,golint
	// 日志自动切割.
	OPT_ROTATE = 32 //nolint:stylecheck,golint
	// 默认日志 flags.
	OPT_DEFAULT = OPT_ENC_CONSOLE | OPT_SHORT_CALLER | OPT_ROTATE //nolint:stylecheck,golint

	DEBUG  = "debug"
	INFO   = "info"
	WARN   = "warn"
	ERROR  = "error"
	PANIC  = "panic"
	DPANIC = "dpanic"
	FATAL  = "fatal"
)

var (
	MaxSize    = 32 // MB
	MaxBackups = 5
	MaxAge     = 30 // day

	mtx = &sync.Mutex{}
)

type Logger struct {
	name    string
	zsl     *zap.SugaredLogger
	rlimits []*rate.Limiter
	hints   []string
}

func (l *Logger) allowed(r float64) (string, bool) {
	if len(l.rlimits) == 0 { // no limiter, log them all
		return "", true
	}

	for idx, rl := range l.rlimits {
		if cliutils.FloatEquals(r, float64(rl.Limit())) {
			return l.hints[idx], rl.Allow()
		}
	}

	// @r not found, then no limiter on current log, we allow it.
	return "", true
}

func (l *Logger) Name() string {
	return l.name
}

func (l *Logger) Infof(fmt string, args ...any) {
	l.zsl.Infof(fmt, args...)
}

func (l *Logger) Infow(fmt string, args ...any) {
	l.zsl.Infow(fmt, args...)
}

func (l *Logger) Info(args ...any) {
	l.zsl.Info(args...)
}

func (l *Logger) RLInfof(r float64, fmt string, args ...any) {
	if h, ok := l.allowed(r); ok {
		l.zsl.Infof(h+fmt, args...)
	}
}

func (l *Logger) RLInfo(r float64, args ...any) {
	if h, ok := l.allowed(r); ok {
		if h != "" {
			xargs := []any{h}
			l.zsl.Info(append(xargs, args...)...)
		} else {
			l.zsl.Info(args...)
		}
	}
}

func (l *Logger) Warnf(fmt string, args ...any) {
	l.zsl.Warnf(fmt, args...)
}

func (l *Logger) Warnw(fmt string, args ...any) {
	l.zsl.Warnw(fmt, args...)
}

func (l *Logger) Warn(args ...any) {
	l.zsl.Warn(args...)
}

func (l *Logger) RLWarnf(r float64, fmt string, args ...any) {
	if h, ok := l.allowed(r); ok {
		l.zsl.Warnf(h+fmt, args...)
	}
}

func (l *Logger) RLWarn(r float64, args ...any) {
	if h, ok := l.allowed(r); ok {
		if h != "" {
			xargs := []any{h}
			l.zsl.Warn(append(xargs, args...)...)
		} else {
			l.zsl.Warn(args...)
		}
	}
}

func (l *Logger) Errorf(fmt string, args ...any) {
	l.zsl.Errorf(fmt, args...)
}

func (l *Logger) Errorw(fmt string, args ...any) {
	l.zsl.Errorw(fmt, args...)
}

func (l *Logger) Error(args ...any) {
	l.zsl.Error(args...)
}

func (l *Logger) RLErrorf(r float64, fmt string, args ...any) {
	if h, ok := l.allowed(r); ok {
		l.zsl.Errorf(h+fmt, args...)
	}
}

func (l *Logger) RLError(r float64, args ...any) {
	if h, ok := l.allowed(r); ok {
		if h != "" {
			xargs := []any{h}
			l.zsl.Error(append(xargs, args...)...)
		} else {
			l.zsl.Error(args...)
		}
	}
}

func (l *Logger) Debugf(fmt string, args ...any) {
	l.zsl.Debugf(fmt, args...)
}

func (l *Logger) Debugw(fmt string, args ...any) {
	l.zsl.Debugw(fmt, args...)
}

func (l *Logger) Debug(args ...any) {
	l.zsl.Debug(args...)
}

func (l *Logger) RLDebugf(r float64, fmt string, args ...any) {
	if h, ok := l.allowed(r); ok {
		l.zsl.Debugf(h+fmt, args...)
	}
}

func (l *Logger) RLDebug(r float64, args ...any) {
	if h, ok := l.allowed(r); ok {
		if h != "" {
			xargs := []any{h}
			l.zsl.Debug(append(xargs, args...))
		} else {
			l.zsl.Debug(args...)
		}
	}
}

func (l *Logger) Fatalf(fmt string, args ...any) {
	// fatal log not rate limited
	l.zsl.Fatalf(fmt, args...)
}

func (l *Logger) Fatalw(fmt string, args ...any) {
	// fatal log not rate limited
	l.zsl.Fatalw(fmt, args...)
}

func (l *Logger) Fatal(args ...any) {
	// fatal log not rate limited
	l.zsl.Fatal(args...)
}

func (l *Logger) Panicf(fmt string, args ...any) {
	// panic log not rate limited
	l.zsl.Panicf(fmt, args...)
}

func (l *Logger) Panicw(fmt string, args ...any) {
	// panic log not rate limited
	l.zsl.Panicw(fmt, args...)
}

func (l *Logger) Panic(args ...any) {
	// panic log not rate limited
	l.zsl.Panic(args...)
}

func (l *Logger) Level() zapcore.Level {
	return l.zsl.Level()
}

type Option struct {
	Path     string
	Level    string
	MaxSize  int
	Flags    int
	Compress bool
}

func Reset() {
	mtx.Lock()
	defer mtx.Unlock()
	root = nil

	slogs = &sync.Map{}

	defaultStdoutRootLogger = nil

	totalSloggers = 0

	if err := doInitStdoutLogger(); err != nil {
		panic(err.Error())
	}
}

func Close() {
	if root != nil {
		if err := root.Sync(); err != nil {
			_ = err // pass
		}
	}
}

//nolint:gochecknoinits
func init() {
	if err := doInitStdoutLogger(); err != nil {
		panic(err.Error())
	}

	if v, ok := os.LookupEnv("LOGGER_PATH"); ok {
		opt := &Option{
			Level: DEBUG,
			Flags: OPT_DEFAULT,
			Path:  v,
		}

		if err := setRootLoggerFromEnv(opt); err != nil {
			panic(err.Error())
		}
	}
}
