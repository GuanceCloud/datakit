package tracer

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/go-sql-driver/mysql"
	"github.com/nsqio/go-nsq"
	nsqtracer "gitlab.jiagouyun.com/cloudcare-tools/cliutils/ddtrace/go-nsq"
	redistracer "gitlab.jiagouyun.com/cloudcare-tools/cliutils/ddtrace/go-redis"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	ddtsql "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
	ddtredis "gopkg.in/DataDog/dd-trace-go.v1/contrib/go-redis/redis"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var (
	l = logger.DefaultSLogger("ddtraces")
)

type DiscardLogger struct{}

func (*DiscardLogger) Log(msg string) {}

type SimpleLogger struct{}

func (*SimpleLogger) Log(msg string) {
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

func (this *Tracer) StartSpan(resource string, spanType string) ddtrace.Span {
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

func (this *Tracer) StartSpanFromContext(ctx context.Context, resource, operatoin string, spanType string, opts ...Option) (ddtrace.Span, context.Context) {
	if !this.Enabled {
		return nil, nil
	}

	ssopts := []ddtrace.StartSpanOption{
		tracer.ServiceName(this.Service),
		tracer.ResourceName(resource),
		tracer.SpanType(spanType),
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

func (this *Tracer) GinMiddleware(resource string, spanType string, opts ...Option) gin.HandlerFunc {
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
			tracer.SpanType(spanType),
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
	if this.Enabled {
		return nsqtracer.NewProducer(addr, config, nsqtracer.WithService(this.Service), nsqtracer.WithContext(context.Background()))
	} else {
		return nsq.NewProducer(addr, config)
	}
}

func (this *Tracer) NewConsumer(topic string, channel string, config *nsq.Config) (nsqtracer.TraceableConsumer, error) {
	if this.Enabled {
		return nsqtracer.NewConsumer(topic, channel, config, nsqtracer.WithService(this.Service), nsqtracer.WithContext(context.Background()))
	} else {
		return nsq.NewConsumer(topic, channel, config)
	}
}

func (this *Tracer) NewRedisClient(rdsopt *redis.Options) redistracer.TraceableClient {
	if this.Enabled {
		return ddtredis.NewClient(rdsopt, ddtredis.WithServiceName(this.Service))
	} else {
		return redis.NewClient(rdsopt)
	}
}

func (this *Tracer) OpenDB(driverName, dataSourceName string) (*sql.DB, error) {
	if this.Enabled {
		ddtsql.Register(driverName, &mysql.MySQLDriver{}, ddtsql.WithServiceName(this.Service))

		return ddtsql.Open(driverName, dataSourceName, ddtsql.WithServiceName(this.Service))
	} else {
		return sql.Open(driverName, dataSourceName)
	}
}
