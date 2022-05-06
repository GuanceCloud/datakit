// Package tracer start global tracer
package tracer

import (
	"os"
	"strconv"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gopkg.in/CodapeWild/dd-trace-go.v1/ddtrace/tracer"
)

func Start() {
	if config.Cfg.Tracer == nil {
		return
	}

	if value, ok := os.LookupEnv("DD_TRACE_ENABLED"); ok {
		if enabled, err := strconv.ParseBool(value); err == nil {
			config.Cfg.Tracer.TraceEnabled = enabled
		}
	}
	if !config.Cfg.Tracer.TraceEnabled {
		return
	}

	startOpts := config.Cfg.Tracer.GetStartOptions()
	// one can add tracer startup options as err log like:
	// startOpts = append(startOpts, tracer.WithLogger(&itracer.StdLogger{}))
	tracer.Start(startOpts...)
}

func Stop() {
	tracer.Stop()
}
