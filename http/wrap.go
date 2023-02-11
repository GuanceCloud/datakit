// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"net/http"
	"time"

	tollbooth "github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
	"github.com/gin-gonic/gin"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
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

func ginWraper(lmt *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		m := &apiMetric{
			api: c.Request.URL.Path + "@" + c.Request.Method,
		}

		if isBlocked(lmt, c.Writer, c.Request) {
			uhttp.HttpErr(c, ErrReachLimit)
			lmt.ExecOnLimitReached(c.Writer, c.Request)
			m.statusCode = http.StatusTooManyRequests
			m.limited = true
			c.Abort()
			goto feed
		}

		c.Next()
		m.statusCode = c.Writer.Status()
		m.latency = time.Since(start) // only un-limit request logged the latency

	feed:
		feedMetric(m)
	}
}

type APIHandler func(http.ResponseWriter, *http.Request, ...interface{}) (interface{}, error)

func rawHTTPWraper(lmt *limiter.Limiter, next APIHandler, other ...interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		m := &apiMetric{
			api: c.Request.URL.Path + "@" + c.Request.Method,
		}

		if isBlocked(lmt, c.Writer, c.Request) {
			uhttp.HttpErr(c, ErrReachLimit)
			lmt.ExecOnLimitReached(c.Writer, c.Request)
			m.statusCode = http.StatusTooManyRequests
			m.limited = true
			c.Abort()
			goto feed
		}

		if res, err := next(c.Writer, c.Request, other...); err != nil {
			uhttp.HttpErr(c, err)
			l.Debugf("wrap next error: %s", err)

			m.statusCode = getStatusCode(err)
		} else {
			m.statusCode = http.StatusOK
			OK.HttpBody(c, res)
		}

		m.latency = time.Since(start) // only un-limit request logged the latency

	feed:
		feedMetric(m)
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
