package tracer

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var (
	GlobalTracer *Tracer
	l            = logger.DefaultSLogger("dk_tracer")
)

type DDLog struct{}

func (ddl DDLog) Log(msg string) { // use exist logger for ddtrace log
	l.Debug(msg)
}

type Option func(opt *Tracer)

func WithService(name, version string) Option {
	return func(opt *Tracer) {
		opt.Service = name
		opt.Version = version
	}
}

func WithEnable(enable bool) Option {
	return func(opt *Tracer) {
		opt.Enabled = enable
	}
}

func WithAgentAddr(host string, port int) Option {
	addr := fmt.Sprintf("%s:%d", host, port)
	return func(opt *Tracer) {
		opt.Host = host
		opt.Port = port
		opt.addr = addr
	}
}

func WithDebug(debug bool) Option {
	return func(opt *Tracer) {
		opt.Debug = debug
	}
}

func WithLogger(logger ddtrace.Logger) Option {
	return func(opt *Tracer) {
		opt.logger = logger
	}
}

func WithFinishTime(t time.Time) tracer.FinishOption {
	return tracer.FinishTime(t)
}

func WithError(err error) tracer.FinishOption {
	return tracer.WithError(err)
}

type Tracer struct {
	Service  string `toml:"service"`
	Version  string `toml:"version"`
	Enabled  bool   `toml:"enabled"`
	Resource string `toml:"resource"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	addr     string
	// StatsdPort          int    `toml:"statsd_port"`
	// EnableRuntimeMetric bool   `toml:"enable_runtime_metric"`
	Debug  bool `toml:"debug"`
	logger ddtrace.Logger
}

func newTracer(opts ...Option) *Tracer {
	tracer := &Tracer{}
	for _, opt := range opts {
		opt(tracer)
	}

	return tracer
}

func (t *Tracer) Start(opts ...Option) {
	for _, opt := range opts {
		opt(t)
	}

	if !t.Enabled {
		return
	}

	sopts := []tracer.StartOption{
		tracer.WithEnv("prod"),
		tracer.WithService(t.Service),
		tracer.WithServiceVersion(t.Version),
		tracer.WithAgentAddr(t.addr),
		tracer.WithDebugMode(t.Debug),
		tracer.WithLogger(t.logger),
	}

	l.Infof("starting ddtrace on datakit...")
	tracer.Start(sopts...)
}

func (t *Tracer) StartSpan(resource string) ddtrace.Span {
	opts := []ddtrace.StartSpanOption{
		tracer.SpanType(ext.SpanTypeHTTP),
		tracer.ServiceName(t.Service),
		tracer.ResourceName(resource),
	}

	return tracer.StartSpan(resource, opts...)
}

func (t *Tracer) Stop() {
	if t.Enabled {
		tracer.Stop()
	}
}

func init() {
	GlobalTracer = newTracer(WithService("datakit", git.Version), WithEnable(true), WithAgentAddr("10.200.7.21", 9529), WithDebug(true), WithLogger(DDLog{}))
}
