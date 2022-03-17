// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"net/http"
	"time"

	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
	"github.com/gin-gonic/gin"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
)

var reqLimiter *limiter.Limiter

func isBlocked(lmt *limiter.Limiter, w http.ResponseWriter, r *http.Request) bool {
	if lmt == nil {
		return false
	}

	return tollbooth.LimitByRequest(lmt, w, r) != nil
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
		m.latency = time.Since(start) // only un-limit request logged the latency
		m.statusCode = c.Writer.Status()
		m.latency = time.Since(start) // only un-limit request logged the latency

	feed:
		feedMetric(m)
	}
}

type apiHandler func(http.ResponseWriter, *http.Request, ...interface{}) (interface{}, error)

func rawHTTPWraper(lmt *limiter.Limiter, next apiHandler, any ...interface{}) gin.HandlerFunc {
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

		if res, err := next(c.Writer, c.Request, any...); err != nil {
			uhttp.HttpErr(c, err)
		} else {
			OK.HttpBody(c, res)
		}

		m.latency = time.Since(start) // only un-limit request logged the latency
		m.statusCode = c.Writer.Status()
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
	}).SetOnLimitReached(limitReach)
}
