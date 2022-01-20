package logger

import (
	"fmt"

	"go.uber.org/zap"
)

var (
	defaultStdoutRootLogger *zap.Logger // used for logging where root logger not setted
	StdoutColor             bool
	StdoutLevel             = DEBUG
)

func doInitStdoutLogger() error {
	flags := OPT_DEFAULT
	if StdoutColor {
		flags |= OPT_COLOR
	}

	var err error
	defaultStdoutRootLogger, err = stdoutLogger(StdoutLevel, flags)
	if err != nil {
		return err
	}
	return nil
}

func doSetStdoutLogger(opt *Option) error {
	// reset default stdout logger
	defaultStdoutRootLogger = nil
	var err error
	defaultStdoutRootLogger, err = stdoutLogger(opt.Level, opt.Flags)
	if err != nil {
		return fmt.Errorf("stdoutLogger: %w", err)
	}
	return nil
}

func stdoutLogger(level string, options int) (*zap.Logger, error) {
	opt := options | OPT_STDOUT

	if rootlogger, err := newRootLogger("", level, opt); err != nil {
		return nil, err
	} else {
		return rootlogger, err
	}
}
