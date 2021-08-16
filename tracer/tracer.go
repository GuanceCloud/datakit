package tracer

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/ddtrace/tracer"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var (
	GlobalTracer *tracer.Tracer
	l            = logger.DefaultSLogger("dk_tracer")
)

func init() {
	GlobalTracer = tracer.NewTracer(false, tracer.WithLogger(&tracer.DiscardLogger{}))
}
