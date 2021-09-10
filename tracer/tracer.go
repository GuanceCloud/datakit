package tracer

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/ddtrace/tracer"
)

var (
	GlobalTracer *tracer.Tracer = tracer.NewTracer(false, tracer.WithLogger(&tracer.DiscardLogger{}))
)
