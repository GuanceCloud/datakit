package logger

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	// non-json
	OPT_ENC_CONSOLE = 1 //nolint:golint,stylecheck

	OPT_SHORT_CALLER    = 2                                               //nolint:stylecheck,golint
	OPT_STDOUT          = 4                                               //nolint:stylecheck,golint
	OPT_COLOR           = 8                                               //nolint:stylecheck,golint
	OPT_RESERVED_LOGGER = 16                                              //nolint:stylecheck,golint
	OPT_ROTATE          = 32                                              //nolint:stylecheck,golint
	OPT_DEFAULT         = OPT_ENC_CONSOLE | OPT_SHORT_CALLER | OPT_ROTATE //nolint:stylecheck,golint

	DEBUG = "debug"
	INFO  = "info"
)

var (
	defaultRootLogger *zap.Logger
	stdoutRootLogger  *zap.Logger

	ll                  *zap.SugaredLogger
	reservedSLoggerName string = "__reserved__"
	slogs               *sync.Map

	mtx = &sync.Mutex{}

	MaxSize    = 32 // megabytes
	MaxBackups = 5
	MaxAge     = 28 // day
)

type Logger struct {
	*zap.SugaredLogger
}

func SetStdoutRootLogger(level string, options int) {
	mtx.Lock()
	defer mtx.Unlock()

	if stdoutRootLogger != nil {
		return
	}

	var err error
	stdoutRootLogger, err = newRootLogger("", level, options)
	if err != nil {
		panic(err)
	}
}

func SetGlobalRootLogger(fpath, level string, options int) {
	mtx.Lock()
	defer mtx.Unlock()

	if defaultRootLogger != nil {
		if ll != nil {
			ll.Warnf("global root logger has been initialized %+#v", defaultRootLogger)
		}

		return
	}

	var err error
	defaultRootLogger, err = newRootLogger(fpath, level, options)
	if err != nil {
		panic(err)
	}

	slogs = &sync.Map{}

	if options&OPT_RESERVED_LOGGER != 0 {
		ll = getSugarLogger(defaultRootLogger, reservedSLoggerName)
		slogs.Store(reservedSLoggerName, ll)

		ll.Info("root logger init ok")
	}
}

const (
	rootNotInitialized = "you should init a root logger via SetGlobalRootLogger or SetStdoutRootLogger"
)

func SLogger(name string) *Logger {
	if defaultRootLogger == nil && stdoutRootLogger == nil {
		panic(rootNotInitialized)
	}

	return &Logger{SugaredLogger: slogger(name)}
}

func DefaultSLogger(name string) *Logger {
	return &Logger{SugaredLogger: slogger(name)}
}

func slogger(name string) *zap.SugaredLogger {
	root := defaultRootLogger // prefer defaultRootLogger
	if root == nil {
		root = stdoutRootLogger
	}

	if root == nil {
		if runtime.GOOS != "windows" {
			SetStdoutRootLogger(DEBUG, OPT_DEFAULT|OPT_STDOUT|OPT_COLOR)
		} else {
			SetStdoutRootLogger(DEBUG, OPT_DEFAULT|OPT_STDOUT)
		}
		root = stdoutRootLogger
	}

	newlog := getSugarLogger(root, name)

	if defaultRootLogger != nil {
		l, ok := slogs.LoadOrStore(name, newlog)
		if ll != nil {
			if ok {
				ll.Debugf("add new sloger `%s'", name)
			} else {
				ll.Debugf("reused exist sloger `%s'", name)
			}
		}

		return l.(*zap.SugaredLogger)
	}

	return newlog
}

func getLogger(root *zap.Logger, name string) *zap.Logger {
	return root.Named(name)
}

func getSugarLogger(root *zap.Logger, name string) *zap.SugaredLogger {
	return root.Sugar().Named(name)
}

func newWinFileSink(u *url.URL) (zap.Sink, error) {
	return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}

func newRotateRootLogger(fpath, level string, options int) (*zap.Logger, error) {
	if fpath == "" {
		fmt.Printf("default log file set to %s/logger.log\n", os.TempDir())
		fpath = filepath.Join(os.TempDir(), `logger.log`)
	}

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

	if options&OPT_COLOR != 0 {
		encodeCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if options&OPT_SHORT_CALLER != 0 {
		encodeCfg.EncodeCaller = zapcore.ShortCallerEncoder
	}

	var enc zapcore.Encoder
	if options&OPT_ENC_CONSOLE != 0 {
		enc = zapcore.NewConsoleEncoder(encodeCfg)
	} else {
		enc = zapcore.NewJSONEncoder(encodeCfg)
	}

	lvl := zap.InfoLevel
	switch strings.ToLower(level) {
	case INFO: // pass
	case DEBUG:
		lvl = zap.DebugLevel
	default:
		lvl = zap.DebugLevel
	}

	core := zapcore.NewCore(enc, w, lvl)
	l := zap.New(core, zap.AddCaller())
	return l, nil
}

func newRootLogger(fpath, level string, options int) (*zap.Logger, error) {

	if options&OPT_ROTATE != 0 && options&OPT_STDOUT == 0 {
		return newRotateRootLogger(fpath, level, options)
	}

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
			if err := zap.RegisterSink("winfile", newWinFileSink); err != nil {
				panic(err)
			}

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
