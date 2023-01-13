// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package logger

import (
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
)

var (
	totalSloggers int64
	slogs         = &sync.Map{}
)

func SLogger(name string) *Logger {
	if root == nil && defaultStdoutRootLogger == nil {
		panic("should not been here: root logger not set")
	}

	return &Logger{SugaredLogger: slogger(name)}
}

func DefaultSLogger(name string) *Logger {
	return &Logger{SugaredLogger: slogger(name)}
}

func TotalSLoggers() int64 {
	return atomic.LoadInt64(&totalSloggers)
}

func slogger(name string) *zap.SugaredLogger {
	r := root // prefer root logger

	if r == nil {
		r = defaultStdoutRootLogger
	}

	if r == nil {
		panic("should not been here")
	}

	newlog := getSugarLogger(r, name)
	if root != nil {
		l, loaded := slogs.LoadOrStore(name, newlog)
		if !loaded {
			atomic.AddInt64(&totalSloggers, 1)
		}

		return l.(*zap.SugaredLogger)
	}

	return newlog
}

func getSugarLogger(l *zap.Logger, name string) *zap.SugaredLogger {
	return l.Sugar().Named(name)
}
