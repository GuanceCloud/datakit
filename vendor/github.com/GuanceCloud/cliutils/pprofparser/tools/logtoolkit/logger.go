// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package logtoolkit is a collection of log utils.
package logtoolkit

import (
	"sync"

	"github.com/GuanceCloud/cliutils/logger"
)

var (
	defaultLogger  *logger.Logger
	logInitOnce    sync.Once
	loggerPool     = make(map[string]*logger.Logger)
	loggerPoolLock = &sync.Mutex{}
)

func Logger(name ...string) *logger.Logger {
	if len(name) == 0 {
		logInitOnce.Do(func() {
			defaultLogger = logger.SLogger("global")
		})
		return defaultLogger
	}

	logName := name[0]
	if logHandler, ok := loggerPool[logName]; ok {
		return logHandler
	}
	loggerPoolLock.Lock()
	defer loggerPoolLock.Unlock()
	if _, ok := loggerPool[logName]; !ok {
		loggerPool[logName] = logger.SLogger(logName)
	}
	return loggerPool[logName]
}

func Info(args ...interface{}) {
	Logger().Info(args...)
}

func Warn(args ...interface{}) {
	Logger().Warn(args...)
}

func Error(args ...interface{}) {
	Logger().Error(args...)
}

func Fatal(args ...interface{}) {
	Logger().Fatal(args...)
}

func Infof(format string, args ...interface{}) {
	Logger().Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	Logger().Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	Logger().Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	Logger().Fatalf(format, args...)
}
