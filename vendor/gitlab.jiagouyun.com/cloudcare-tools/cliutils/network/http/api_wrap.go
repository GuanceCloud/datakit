package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	ErrTooManyRequest = NewErr(errors.New("reach max API rate limit"), http.StatusTooManyRequests)
	HttpOK            = NewErr(nil, http.StatusOK)
	EnableTracing     bool
)

type WrapPlugins struct {
	Limiter  APIRateLimiter
	Reporter APIMetricReporter
	//Tracer   Tracer
}

type apiHandler func(http.ResponseWriter, *http.Request, ...interface{}) (interface{}, error)

func HTTPAPIWrapper(p *WrapPlugins, next apiHandler, any ...interface{}) func(*gin.Context) {
	return func(c *gin.Context) {
		var start time.Time
		var m *APIMetric

		if p != nil && p.Reporter != nil {
			start = time.Now()
			m = &APIMetric{
				API: c.Request.URL.Path + "@" + c.Request.Method,
			}
		}

		if p != nil && p.Limiter != nil {
			if p.Limiter.RequestLimited(c.Request) {
				HttpErr(c, ErrTooManyRequest)
				p.Limiter.LimitReadchedCallback(c.Request)
				if m != nil {
					m.StatusCode = ErrTooManyRequest.HttpCode
					m.Limited = true
				}
				c.Abort()
				goto feed
			}
		}

		if res, err := next(c.Writer, c.Request, any...); err != nil {
			HttpErr(c, err)
		} else {
			HttpOK.WriteBody(c, res)
		}

		if m != nil {
			m.StatusCode = c.Writer.Status()
			m.Latency = time.Since(start)
		}

	feed:
		if p != nil && m != nil {
			p.Reporter.Report(m)
		}
	}
}
