package log

type Logger interface {
	Errorf(format string, args ...interface{})
	Error(args ...interface{})
	Debugf(format string, args ...interface{})
	Debug(args ...interface{})
	Warnf(format string, args ...interface{})
	Warn(args ...interface{})
	Infof(format string, args ...interface{})
	Info(args ...interface{})
}
