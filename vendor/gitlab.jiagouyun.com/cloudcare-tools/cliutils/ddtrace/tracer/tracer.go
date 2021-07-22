package tracer

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nsqio/go-nsq"
	nsqtracer "gitlab.jiagouyun.com/cloudcare-tools/cliutils/ddtrace/go-nsq"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var l = logger.DefaultSLogger("ddtraces")

type DDLog struct{}

func (ddl DDLog) Log(msg string) { // use exist logger for ddtrace log
	l.Debug(msg)
}

type Tracer struct {
	Enabled bool   `toml:"enabled" yaml:"enabled"`
	Host    string `toml:"host" yaml:"host"`
	Port    int    `toml:"port" yaml:"port"`
	addr    string
	Service string `toml:"service" yaml:"service"`
	Version string `toml:"version" yaml:"version"`
	Debug   bool   `toml:"debug" yaml:"debug"`
	env     string
	logger  ddtrace.Logger
}

func NewTracer(enabled bool, opts ...Option) *Tracer {
	tracer := &Tracer{Enabled: enabled}
	loadConfig(tracer, opts...)

	return tracer
}

func loadConfig(tracer *Tracer, opts ...Option) {
	for _, opt := range opts {
		opt(tracer)
	}
	tracer.addr = fmt.Sprintf("%s:%d", tracer.Host, tracer.Port)
}

func (this *Tracer) Start(opts ...Option) {
	if !this.Enabled {
		return
	}

	loadConfig(this, opts...)
	sopts := []tracer.StartOption{
		tracer.WithEnv(this.env),
		tracer.WithAgentAddr(this.addr),
		tracer.WithService(this.Service),
		tracer.WithServiceVersion(this.Version),
		tracer.WithDebugMode(this.Debug),
		tracer.WithLogger(this.logger),
	}

	tracer.Start(sopts...)
}

func (this *Tracer) StartSpan(resource string, spanType SpanType) ddtrace.Span {
	if !this.Enabled {
		return nil
	}

	ssopts := []ddtrace.StartSpanOption{
		tracer.ServiceName(this.Service),
		tracer.ResourceName(resource),
		tracer.SpanType(string(spanType)),
	}

	return tracer.StartSpan(resource, ssopts...)
}

func (this *Tracer) StartSpanFromContext(ctx context.Context, resource, operatoin string, spanType SpanType, opts ...Option) (ddtrace.Span, context.Context) {
	if !this.Enabled {
		return nil, nil
	}

	ssopts := []ddtrace.StartSpanOption{
		tracer.ServiceName(this.Service),
		tracer.ResourceName(resource),
		tracer.SpanType(string(spanType)),
		tracer.Measured(),
	}

	return tracer.StartSpanFromContext(ctx, operatoin, ssopts...)
}

func (this *Tracer) SetTag(span tracer.Span, key string, value interface{}) {
	if this.Enabled && span != nil {
		span.SetTag(key, value)
	}
}

func (this *Tracer) FinishSpan(span tracer.Span, opts ...ddtrace.FinishOption) {
	if this.Enabled && span != nil {
		span.Finish(opts...)
	}
}

func (this *Tracer) Inject(span ddtrace.Span, header http.Header) error {
	if this.Enabled && span != nil {
		return tracer.Inject(span.Context(), tracer.HTTPHeadersCarrier(header))
	}

	return nil
}

func (this *Tracer) GinMiddleware(resource string, spanType SpanType, opts ...Option) gin.HandlerFunc {
	if !this.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	loadConfig(this, opts...)

	return func(c *gin.Context) {
		ssopts := []ddtrace.StartSpanOption{
			tracer.ServiceName(this.Service),
			tracer.ResourceName(resource),
			tracer.SpanType(string(spanType)),
			tracer.Tag(ext.HTTPMethod, c.Request.Method),
			tracer.Tag(ext.HTTPURL, c.Request.URL.Path),
			tracer.Measured(),
		}
		if spanctx, err := tracer.Extract(tracer.HTTPHeadersCarrier(c.Request.Header)); err == nil {
			ssopts = append(ssopts, tracer.ChildOf(spanctx))
		}

		span, ctx := tracer.StartSpanFromContext(c.Request.Context(), c.Request.RequestURI, ssopts...)
		defer span.Finish(tracer.FinishTime(time.Now()))

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		status := c.Writer.Status()
		span.SetTag(ext.HTTPCode, strconv.Itoa(status))
		if status >= 500 && status < 600 {
			span.SetTag("http_request_error", fmt.Errorf("%d: %s", status, http.StatusText(status)))
		}
		if len(c.Errors) > 0 {
			span.SetTag("gin_request_errors", c.Errors.String())
		}
	}
}

func (this *Tracer) Stop() {
	if this.Enabled {
		tracer.Stop()
	}
}

func (this *Tracer) NewProducer(addr string, config *nsq.Config) (nsqtracer.TraceableProducer, error) {
	if !this.Enabled {
		return nsq.NewProducer(addr, config)
	} else {
		return nsqtracer.NewProducer(addr, config, nsqtracer.WithService(this.Service), nsqtracer.WithContext(context.Background()))
	}
}

func (this *Tracer) NewConsumer(topic string, channel string, config *nsq.Config) (nsqtracer.TraceableConsumer, error) {
	if !this.Enabled {
		return nsq.NewConsumer(topic, channel, config)
	} else {
		return nsqtracer.NewConsumer(topic, channel, config, nsqtracer.WithService(this.Service), nsqtracer.WithContext(context.Background()))
	}
}
