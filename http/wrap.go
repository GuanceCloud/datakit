// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
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

type APIHandler func(http.ResponseWriter, *http.Request, ...interface{}) (interface{}, error)

func rawHTTPWraper(lmt *limiter.Limiter, next APIHandler, other ...interface{}) gin.HandlerFunc {
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

		if isBlocked(lmt, c.Writer, c.Request) {
			uhttp.HttpErr(c, ErrReachLimit)
			lmt.ExecOnLimitReached(c.Writer, c.Request)
			status = http.StatusText(http.StatusTooManyRequests)

			c.Abort()
			return
		}

		if res, err := next(c.Writer, c.Request, other...); err != nil {
			uhttp.HttpErr(c, err)
			status = http.StatusText(getStatusCode(err))
		} else {
			status = http.StatusText(http.StatusOK)
			OK.HttpBody(c, res)
		}
	}
}

func limitReach(w http.ResponseWriter, r *http.Request) {
	// TODO: export metrics here group by r.Method + r.URL
	// or we can cache the request
}

func setupLimiter(limit float64) *limiter.Limiter {
	return tollbooth.NewLimiter(limit, &limiter.ExpirableOptions{
		DefaultExpirationTTL: time.Second,
	}).SetOnLimitReached(limitReach).SetBurst(1)
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
