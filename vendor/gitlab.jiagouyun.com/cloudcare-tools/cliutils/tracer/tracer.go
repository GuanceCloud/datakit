package tracer

import (
	"net"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gopkg.in/CodapeWild/dd-trace-go.v1/ddtrace/tracer"
)

var log = logger.DefaultSLogger("ddtrace")

type DiscardLogger struct{}

func (*DiscardLogger) Log(msg string) {}

type StdLogger struct{}

func (*StdLogger) Log(msg string) {
	log.Debug(msg)
}

type Tracer struct {
	TraceEnabled bool   `toml:"trace_enabled" yaml:"trace_enabled"` // env: DD_TRACE_ENABLED
	Host         string `toml:"host" yaml:"host"`                   // env: DD_AGENT_HOST
	Port         string `toml:"port" yaml:"port"`                   // env: DD_TRACE_AGENT_PORT
	agentAddr    string
	Service      string                 `toml:"service" yaml:"service"`           // env: DD_SERVICE
	Version      string                 `toml:"version" yaml:"version"`           // env: DD_VERSION
	LogsStartup  bool                   `toml:"logs_startup" yaml:"logs_startup"` // env: DD_TRACE_STARTUP_LOGS
	Debug        bool                   `toml:"debug" yaml:"debug"`               // env: DD_TRACE_DEBUG
	Env          string                 `toml:"env" yaml:"env"`                   // env: DD_ENV
	Tags         map[string]interface{} `toml:"tags" yaml:"tags"`                 // env: DD_TAGS
}

func (t *Tracer) GetStartOptions(opts ...StartOption) []tracer.StartOption {
	if t.Tags == nil {
		t.Tags = make(map[string]interface{})
	}

	for i := range opts {
		opts[i](t)
	}

	if t.Host == "" {
		t.Host = "127.0.0.1"
	}
	if t.Port == "" {
		t.Port = "9529"
	}
	t.agentAddr = net.JoinHostPort(t.Host, t.Port)

	startOpts := []tracer.StartOption{
		tracer.WithAgentAddr(t.agentAddr),
		tracer.WithService(t.Service),
		tracer.WithServiceVersion(t.Version),
		tracer.WithLogStartup(t.LogsStartup),
		tracer.WithDebugMode(t.Debug),
		tracer.WithEnv(t.Env),
	}
	for k, v := range t.Tags {
		startOpts = append(startOpts, tracer.WithGlobalTag(k, v))
	}

	return startOpts
}
