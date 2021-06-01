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
	// 禁用 JSON 形式输出
	OPT_ENC_CONSOLE = 1 //nolint:golint,stylecheck

	// 显示代码路径时，不显示全路径
	OPT_SHORT_CALLER = 2 //nolint:stylecheck,golint

	// 日志写到 stdout
	OPT_STDOUT = 4 //nolint:stylecheck,golint

	// 日志内容中追加颜色
	OPT_COLOR = 8 //nolint:stylecheck,golint

	// 开启 logger 模块自用 SugaredLogger
	OPT_RESERVED_LOGGER = 16 //nolint:stylecheck,golint

	// 日志自动切割
	OPT_ROTATE = 32 //nolint:stylecheck,golint

	// 默认日志 flags
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
	defaultRootLogger *zap.Logger
	stdoutRootLogger  *zap.Logger

	ll                  *zap.SugaredLogger // logger's logging
	reservedSLoggerName string             = "__reserved__"
	slogs               *sync.Map

	mtx = &sync.Mutex{}

	MaxSize    = 32 // megabytes
	MaxBackups = 5
	MaxAge     = 28 // day

	EnvRootLoggerPath = "ROOT_LOGGER_PATH"

	defaultOption = &Option{
		Level: DEBUG,
		Flags: OPT_DEFAULT,
	}
)

type Logger struct {
	*zap.SugaredLogger
}

type Option struct {
	Env   string
	Path  string
	Level string

	Flags int
}

func Reset() {

	mtx.Lock()
	defer mtx.Unlock()

	defaultRootLogger = nil
	stdoutRootLogger = nil
}

func InitRoot(opt *Option) error {

	if opt == nil {
		opt = defaultOption
	}

	switch opt.Level {
	case DEBUG, INFO, WARN, ERROR, PANIC, FATAL, DPANIC:
	case "": // 默认使用 DEBUG
		opt.Level = DEBUG

	default:
		return fmt.Errorf("invalid log level `%s'", opt.Level)
	}

	if opt.Flags == 0 {
		opt.Flags = OPT_DEFAULT
	}

	if opt.Env != "" {
		return SetEnvRootLogger(opt.Env, opt.Level, opt.Flags)
	}

	if opt.Path == "" {
		SetStdoutRootLogger(opt.Level, opt.Flags)
		return nil
	} else {
		SetGlobalRootLogger(opt.Path, opt.Level, opt.Flags)
		return nil
	}

	return nil
}

func SetEnvRootLogger(env, level string, options int) error {
	fpath, ok := os.LookupEnv(env)
	if !ok {
		return fmt.Errorf("ENV `%s' not set", env)
	}

	return SetGlobalRootLogger(fpath, level, options)
}

func SetStdoutRootLogger(level string, options int) {
	mtx.Lock()
	defer mtx.Unlock()

	opt := options | OPT_STDOUT
	if stdoutRootLogger != nil {
		return
	}

	var err error
	stdoutRootLogger, err = newRootLogger("", level, opt)
	if err != nil {
		panic(err)
	}
}

func SetGlobalRootLogger(fpath, level string, options int) error {
	mtx.Lock()
	defer mtx.Unlock()

	if defaultRootLogger != nil {
		if ll != nil {
			ll.Warnf("global root logger has been initialized %+#v", defaultRootLogger)
		}

		return nil
	}

	var err error
	defaultRootLogger, err = newRootLogger(fpath, level, options)
	if err != nil {
		return err
	}

	slogs = &sync.Map{}

	if options&OPT_RESERVED_LOGGER != 0 {
		ll = getSugarLogger(defaultRootLogger, reservedSLoggerName)
		slogs.Store(reservedSLoggerName, ll)

		ll.Info("root logger init ok")
	}
	return nil
}

func SLogger(name string) *Logger {
	if defaultRootLogger == nil && stdoutRootLogger == nil {
		panic("root logger not set")
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
		// try set root-logger via env
		if err := SetEnvRootLogger(EnvRootLoggerPath, DEBUG, OPT_DEFAULT); err == nil {
			root = defaultRootLogger
		} else {
			if runtime.GOOS != "windows" {
				SetStdoutRootLogger(DEBUG, OPT_DEFAULT|OPT_STDOUT|OPT_COLOR)
			} else {
				SetStdoutRootLogger(DEBUG, OPT_DEFAULT|OPT_STDOUT)
			}
			root = stdoutRootLogger
		}
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
	case WARN:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case ERROR:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case PANIC:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.PanicLevel)
	case DPANIC:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DPanicLevel)
	case FATAL:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
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
