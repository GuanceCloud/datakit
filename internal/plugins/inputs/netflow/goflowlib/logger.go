// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package goflowlib

import (
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/sirupsen/logrus"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/netflow/common"
	"go.uber.org/zap/zapcore"
)

var zapToLogrusLevel = map[zapcore.Level]logrus.Level{
	zapcore.DebugLevel:  logrus.DebugLevel,
	zapcore.InfoLevel:   logrus.InfoLevel,
	zapcore.WarnLevel:   logrus.WarnLevel,
	zapcore.ErrorLevel:  logrus.ErrorLevel,
	zapcore.DPanicLevel: logrus.PanicLevel,
	zapcore.PanicLevel:  logrus.PanicLevel,
	zapcore.FatalLevel:  logrus.FatalLevel,
}

// GetLogrusLevel returns logrus log level from zap.Level().
func GetLogrusLevel() *logrus.Logger {
	logLevel := l.Level()
	logrusLevel, ok := zapToLogrusLevel[logLevel]
	if !ok {
		l.Warnf("no matching logrus level for seelog level: %s", logLevel.String())
		logrusLevel = logrus.InfoLevel
	}
	logger := logrus.StandardLogger()
	logger.SetLevel(logrusLevel)
	return logger
}

var l = logger.DefaultSLogger(common.InputName)

func SetLogger(log *logger.Logger) {
	l = log
}
