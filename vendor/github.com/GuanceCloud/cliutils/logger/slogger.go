// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package logger

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/GuanceCloud/cliutils"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

var (
	totalSloggers int64
	slogs         = &sync.Map{}
)

type SLogerOpt func(*Logger)

// WithRateLimiter set rate limit on current sloger, n limits the
// max logs can be written per second. The hint will injected to
// the log message like this:
//
//	2025-11-13T11:11:40.168+0800 INFO basic logger/logger_test.go:44 [<your-hint-here>] <your-origin-log-message>
//
// If no hint set, the default seems like this(1 log/sec rate limited):
//
//	2025-11-13T11:11:40.168+0800 INFO basic logger/logger_test.go:44 [1.0-rl] <your-origin-log-message>
func WithRateLimiter(n float64, hint string) SLogerOpt {
	return func(sl *Logger) {
		if n > 0 {
			for _, rl := range sl.rlimits {
				if cliutils.FloatEquals(float64(rl.Limit()), n) {
					return // exist limit skipped
				}
			}

			// add new limiter
			sl.rlimits = append(sl.rlimits, rate.NewLimiter(rate.Limit(n), 1)) // no burst
			if len(hint) == 0 {
				sl.hints = append(sl.hints, fmt.Sprintf("[%.1f-rl] ", n))
			} else {
				sl.hints = append(sl.hints, fmt.Sprintf("[%s] ", hint))
			}
		}
	}
}

func SLogger(name string, opts ...SLogerOpt) *Logger {
	if root == nil && defaultStdoutRootLogger == nil {
		panic("should not been here: root logger not set")
	}

	sl := &Logger{
		name: name,
	}

	for _, opt := range opts {
		opt(sl)
	}

	sl.zsl = slogger(sl.name, 1)

	return sl
}

func DefaultSLogger(name string) *Logger {
	return &Logger{zsl: slogger(name, 1)}
}

func TotalSLoggers() int64 {
	return atomic.LoadInt64(&totalSloggers)
}

func slogger(name string, callerSkip int) *zap.SugaredLogger {
	r := root // prefer root logger

	if r == nil {
		r = defaultStdoutRootLogger
	}

	if r == nil {
		panic("should not been here")
	}

	newlog := getSugarLogger(r, name, callerSkip)
	if root != nil {
		l, loaded := slogs.LoadOrStore(name, newlog)
		if !loaded {
			atomic.AddInt64(&totalSloggers, 1)
		}

		return l.(*zap.SugaredLogger)
	}

	return newlog
}

func getSugarLogger(l *zap.Logger, name string, callerSkip int) *zap.SugaredLogger {
	if callerSkip > 0 {
		return l.WithOptions(zap.AddCallerSkip(callerSkip)).Sugar().Named(name)
	} else {
		return l.Sugar().Named(name)
	}
}
