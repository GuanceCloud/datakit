package logger

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

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

	STDOUT = "stdout" // log output to stdout
)

var (
	root                    *zap.Logger
	defaultStdoutRootLogger *zap.Logger // used for logging where root logger not setted

	totalSloggers int64

	slogs = &sync.Map{}

	mtx = &sync.Mutex{}

	defaultOption = &Option{
		Level: DEBUG,
		Flags: OPT_DEFAULT,
	}
)

var (
	StdoutColor bool
	StdoutLevel = DEBUG

	MaxSize    = 32 // MB
	MaxBackups = 5
	MaxAge     = 30 // day
)

type Logger struct {
	*zap.SugaredLogger
}

type Option struct {
	Path  string
	Level string

	Flags int
}

func init() {
	doInitStdoutLogger()
}

func doInitStdoutLogger() {
	flags := OPT_DEFAULT
	if StdoutColor {
		flags |= OPT_COLOR
	}

	var err error
	defaultStdoutRootLogger, err = stdoutLogger(StdoutLevel, flags)
	if err != nil {
		panic(fmt.Sprintf("should not been here: %s", err))
	}
}

func Reset() {
	mtx.Lock()
	defer mtx.Unlock()
	root = nil

	slogs = &sync.Map{}

	defaultStdoutRootLogger = nil

	totalSloggers = 0

	doInitStdoutLogger()
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

	if opt.Path != "" && (opt.Flags&OPT_STDOUT != 0) {
		return fmt.Errorf("set stdout logging with log path '%s', flag:%b", opt.Path, opt.Flags)
	}

	switch opt.Path {
	case "":
		// use default stdout logger
		return nil
	default:

		return doSetGlobalRootLogger(opt.Path, opt.Level, opt.Flags)
	}
}

func stdoutLogger(level string, options int) (*zap.Logger, error) {
	opt := options | OPT_STDOUT

	if rootlogger, err := newRootLogger("", level, opt); err != nil {
		return nil, err
	} else {
		return rootlogger, err
	}
}

func doSetGlobalRootLogger(fpath, level string, options int) error {

	mtx.Lock()
	defer mtx.Unlock()

	if root != nil {
		return nil
	}

	if _, err := os.Stat(fpath); err != nil { // create file if not exists
		if err := os.MkdirAll(filepath.Dir(fpath), 0700); err != nil {
			return err
		}

		// create empty log file
		if err := ioutil.WriteFile(fpath, nil, 0700); err != nil {
			return err
		}
	}

	var err error
	root, err = newRootLogger(fpath, level, options)
	if err != nil {
		return err
	}

	return nil
}

// Deprecated: use InitRoot() instead
func SetGlobalRootLogger(fpath, level string, options int) error {
	return doSetGlobalRootLogger(fpath, level, options)
}

func SLogger(name string) *Logger {
	if root == nil && defaultStdoutRootLogger == nil {
		panic("should not been here: root logger not set")
	}

	return &Logger{SugaredLogger: slogger(name)}
}

func DefaultSLogger(name string) *Logger {
	return &Logger{SugaredLogger: slogger(name)}
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

func TotalSLoggers() int64 {
	return atomic.LoadInt64(&totalSloggers)
}

func newWinFileSink(u *url.URL) (zap.Sink, error) {
	return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}

func newRotateRootLogger(fpath, level string, options int) (*zap.Logger, error) {
	if fpath == "" {
		fpath = filepath.Join(os.TempDir(), `logger.log`)
	}

	// use lumberjack.Logger for rotate
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fpath,
		MaxSize:    MaxSize,
		MaxBackups: MaxBackups,
		MaxAge:     MaxAge,
	})

	cfg := zapcore.EncoderConfig{
		NameKey:      "MOD",
		MessageKey:   "MSG",
		LevelKey:     "LEV",
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		TimeKey:      "TS",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		CallerKey:    "POS",
		EncodeCaller: zapcore.FullCallerEncoder,
	}

	if options&OPT_COLOR != 0 {
		cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if options&OPT_SHORT_CALLER != 0 {
		cfg.EncodeCaller = zapcore.ShortCallerEncoder
	}

	var enc zapcore.Encoder
	if options&OPT_ENC_CONSOLE != 0 {
		enc = zapcore.NewConsoleEncoder(cfg)
	} else {
		enc = zapcore.NewJSONEncoder(cfg)
	}

	lvl := zap.InfoLevel
	switch strings.ToLower(level) {
	case INFO: // pass
		lvl = zap.InfoLevel
	case DEBUG:
		lvl = zap.DebugLevel
	case WARN:
		lvl = zap.WarnLevel
	case ERROR:
		lvl = zap.ErrorLevel
	case PANIC:
		lvl = zap.PanicLevel
	case DPANIC:
		lvl = zap.DPanicLevel
	case FATAL:
		lvl = zap.FatalLevel
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
			NameKey:      "MOD",
			MessageKey:   "MSG",
			LevelKey:     "LEV",
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			TimeKey:      "TS",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			CallerKey:    "POS",
			EncodeCaller: zapcore.FullCallerEncoder,
		},
	}

	if fpath != "" {
		cfg.OutputPaths = []string{fpath}

		if runtime.GOOS == "windows" { // See: https://github.com/uber-go/zap/issues/621
			if err := zap.RegisterSink("winfile", newWinFileSink); err != nil {
				return nil, err
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
		cfg.OutputPaths = append(cfg.OutputPaths, STDOUT)
	}

	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return l, nil
}

func Close() {
	if root != nil {
		root.Sync()
	}
}
