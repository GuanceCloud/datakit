package logger

import (
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	OPT_ENC_CONSOLE  = 1 // non-json
	OPT_SHORT_CALLER = 2
	OPT_STDOUT       = 4
	OPT_SHUGAR       = 8
	OPT_COLOR        = 16

	DEBUG = "debug"
	INFO  = "info"
)

var (
	defaultRootLogger   *zap.Logger
	__l                 *zap.SugaredLogger
	reservedSLoggerName string = "__reserved__"
	slogs               *sync.Map

	MaxSize    = 32 // megabytes
	MaxBackups = 5
	MaxAge     = 28 // day
)

func SetGlobalRootLogger(fpath, level string, options int) error {
	if defaultRootLogger != nil {
		panic(fmt.Sprintf("global root logger has been initialized: %+#v", defaultRootLogger))
	}

	var err error
	defaultRootLogger, err = NewRootLogger(fpath, level, options)
	if err != nil {
		panic(err)
	}

	slogs = &sync.Map{}

	__l = GetSugarLogger(defaultRootLogger, reservedSLoggerName)
	slogs.Store(reservedSLoggerName, __l)

	__l.Info("root logger init ok")

	return nil
}

const (
	rootNotInitialized = "you should call SetGlobalRootLogger to initialize the global root logger"
)

func Logger(name string) *zap.Logger {
	if defaultRootLogger == nil {
		panic(rootNotInitialized)
	}

	return GetLogger(defaultRootLogger, name)
}

func SLogger(name string) *zap.SugaredLogger {
	if defaultRootLogger == nil {
		panic(rootNotInitialized)
	}

	newlog := GetSugarLogger(defaultRootLogger, name)

	l, ok := slogs.LoadOrStore(name, newlog)
	if ok {
		__l.Debugf("add new sloger `%s'", name)
	} else {
		__l.Debugf("reused exist sloger `%s'", name)
	}

	return l.(*zap.SugaredLogger)
}

func GetLogger(root *zap.Logger, name string) *zap.Logger {
	return root.Named(name)
}

func GetSugarLogger(root *zap.Logger, name string) *zap.SugaredLogger {
	return root.Sugar().Named(name)
}

func newWinFileSink(u *url.URL) (zap.Sink, error) {
	return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}

func _NewRotateRootLogger(fpath, level string, options int) (*zap.Logger, error) {
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fpath,
		MaxSize:    MaxSize,
		MaxBackups: MaxBackups,
		MaxAge:     MaxAge,
	})

	encodeCfg := zapcore.EncoderConfig{
		NameKey:    "MOD",
		MessageKey: "MSG",

		LevelKey:    "LEV",
		EncodeLevel: zapcore.CapitalLevelEncoder,

		TimeKey:    "TS",
		EncodeTime: zapcore.ISO8601TimeEncoder,

		CallerKey:    "POS",
		EncodeCaller: zapcore.FullCallerEncoder,
	}

	encoder := zapcore.NewConsoleEncoder(encodeCfg)

	core := zapcore.NewCore(encoder, w, zap.InfoLevel)
	l := zap.New(core)
	return l, nil
}

func NewRootLogger(fpath, level string, options int) (*zap.Logger, error) {

	cfg := &zap.Config{
		Encoding: `json`,
		EncoderConfig: zapcore.EncoderConfig{
			NameKey:    "MOD",
			MessageKey: "MSG",

			LevelKey:    "LEV",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "TS",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey:    "POS",
			EncodeCaller: zapcore.FullCallerEncoder,
		},
	}

	if fpath != "" {
		cfg.OutputPaths = []string{fpath}

		if runtime.GOOS == "windows" { // See: https://github.com/uber-go/zap/issues/621
			zap.RegisterSink("winfile", newWinFileSink)
			cfg.OutputPaths = []string{
				"winfile:///" + fpath,
			}
		}

	}

	switch strings.ToLower(level) {
	case DEBUG:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case INFO:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}

	if options&OPT_COLOR != 0 {
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if options&OPT_ENC_CONSOLE != 0 {
		cfg.Encoding = "console"
	}

	if options&OPT_SHORT_CALLER != 0 {
		cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	if options&OPT_STDOUT != 0 || fpath == "" { // if no log file path set, default set to stdout
		cfg.OutputPaths = append(cfg.OutputPaths, "stdout")
	}

	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return l, nil
}
