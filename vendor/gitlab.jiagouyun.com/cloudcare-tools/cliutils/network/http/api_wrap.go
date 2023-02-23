// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	ErrTooManyRequest = NewErr(errors.New("reach max API rate limit"), http.StatusTooManyRequests)
	HTTPOK            = NewErr(nil, http.StatusOK) //nolint:errname
	EnableTracing     bool
)

type WrapPlugins struct {
	Limiter  APIRateLimiter
	Reporter APIMetricReporter
}

type apiHandler func(http.ResponseWriter, *http.Request, ...interface{}) (interface{}, error)

func HTTPAPIWrapper(p *WrapPlugins, next apiHandler, args ...interface{}) func(*gin.Context) {
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

		if res, err := next(c.Writer, c.Request, args...); err != nil {
			HttpErr(c, err)
		} else {
			HTTPOK.WriteBody(c, res)
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
