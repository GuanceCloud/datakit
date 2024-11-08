// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
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
	*zap.SugaredLogger
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
