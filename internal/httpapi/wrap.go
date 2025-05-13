// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"context"
	"net/http"
	"time"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
	tollbooth "github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
	"github.com/gin-gonic/gin"
)

var reqLimiter *limiter.Limiter

func isBlocked(lmt *limiter.Limiter, w http.ResponseWriter, r *http.Request) bool {
	if lmt == nil {
		return false
	}

	return tollbooth.LimitByRequest(lmt, w, r) != nil
}

func getStatusCode(err error) int {
	switch e := err.(type) { //nolint:errorlint
	case *uhttp.HttpError:
		return e.HttpCode
	case *uhttp.MsgError:
		return e.HttpError.HttpCode
	default:
		return http.StatusInternalServerError
	}
}

func ginLimiter(lmt *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		status := "unknown"

		defer func() {
			apiCountVec.WithLabelValues(
				c.Request.URL.Path,
				c.Request.Method,
				status).Inc()

			apiElapsedVec.WithLabelValues(
				c.Request.URL.Path,
				c.Request.Method,
				status).Observe(float64(time.Since(start)) / float64(time.Second))

			apiReqSizeVec.WithLabelValues(
				c.Request.URL.Path,
				c.Request.Method,
				status).Observe(float64(approximateRequestSize(c.Request)))
		}()

		if isBlocked(lmt, c.Writer, c.Request) {
			uhttp.HttpErr(c, ErrReachLimit)
			lmt.ExecOnLimitReached(c.Writer, c.Request)

			status = http.StatusText(http.StatusTooManyRequests)
			c.Abort()
			return
		}

		c.Next()
		status = http.StatusText(c.Writer.Status())
	}
}

type (
	APIHandler func(http.ResponseWriter, *http.Request, ...interface{}) (interface{}, error)

	Param string

	HandlerWrapper struct {
		// response wraped within a content field, such as `{"content":{origin-json ...}}`.
		// Or, response not wraped, like `origin-json`
		WrappedResponse bool
	}
)

// RawHTTPWrapper warp HTTP APIs that:
//   - with prometheus metric export
//   - with rate limit
func (hw *HandlerWrapper) RawHTTPWrapper(lmt *limiter.Limiter, next APIHandler, other ...interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		status := "unknown"

		// prometheus metrics
		defer func() {
			apiCountVec.WithLabelValues(
				c.Request.URL.Path,
				c.Request.Method,
				status).Inc()

			apiElapsedVec.WithLabelValues(
				c.Request.URL.Path,
				c.Request.Method,
				status).Observe(float64(time.Since(start)) / float64(time.Second))

			apiReqSizeVec.WithLabelValues(
				c.Request.URL.Path,
				c.Request.Method,
				status).Observe(float64(approximateRequestSize(c.Request)))
		}()

		if lmt != nil {
			if isBlocked(lmt, c.Writer, c.Request) {
				uhttp.HttpErr(c, ErrReachLimit)
				lmt.ExecOnLimitReached(c.Writer, c.Request)
				status = http.StatusText(http.StatusTooManyRequests)

				c.Abort()
				return
			}
		}

		ctx, cancel := context.WithCancel(c.Request.Context())
		defer cancel()
		for _, p := range c.Params {
			ctx = context.WithValue(ctx, Param(p.Key), p.Value)
		}

		if res, err := next(c.Writer, c.Request.WithContext(ctx), other...); err != nil {
			uhttp.HttpErr(c, err)
			status = http.StatusText(getStatusCode(err))
		} else {
			status = http.StatusText(http.StatusOK)
			if hw.WrappedResponse {
				OK.HttpBody(c, res)
			} else {
				OK.WriteBody(c, res)
			}
		}
	}
}

func limitReach(w http.ResponseWriter, r *http.Request) {
	// TODO: export metrics here group by r.Method + r.URL
	// or we can cache the request
	l.Warnf("request %s(%s) reached rate limit, dropped", r.URL.String(), r.Method)
}

func setupLimiter(limit float64, ttl time.Duration) *limiter.Limiter {
	return tollbooth.NewLimiter(limit,
		&limiter.ExpirableOptions{
			DefaultExpirationTTL: ttl,
		}).SetOnLimitReached(limitReach) // .SetBurst(2)
}

// From https://github.com/DanielHeckrath/gin-prometheus/blob/master/gin_prometheus.go
func approximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}
