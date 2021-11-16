package tracer

import (
	"net"
)

type StartOption func(t *Tracer)

func WithTraceEnabled(enabled bool) StartOption {
	return func(t *Tracer) {
		t.TraceEnabled = enabled
	}
}

func WithAgentAddr(host, port string) StartOption {
	return func(t *Tracer) {
		t.Host = host
		t.Port = port
		t.agentAddr = net.JoinHostPort(host, port)
	}
}

func WithService(service string) StartOption {
	return func(t *Tracer) {
		t.Service = service
	}
}

func WithVersion(version string) StartOption {
	return func(t *Tracer) {
		t.Version = version
	}
}

func WithLogsStartup(startup bool) StartOption {
	return func(t *Tracer) {
		t.LogsStartup = startup
	}
}

func WithDebug(debug bool) StartOption {
	return func(t *Tracer) {
		t.Debug = debug
	}
}

func WithGlobalTag(key string, value interface{}) StartOption {
	return func(t *Tracer) {
		t.Tags[key] = value
	}
}

func WithEnv(env string) StartOption {
	return func(t *Tracer) {
		t.Env = env
	}
}
