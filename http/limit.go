package http

import (
	"net/http"
	"time"

	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
	"github.com/gin-gonic/gin"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
)

func limitHandler(lmt *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if lmt == nil {
			c.Next()
		} else {
			if err := tollbooth.LimitByRequest(lmt, c.Writer, c.Request); err != nil {
				uhttp.HttpErr(c, ErrReachLimit)
				lmt.ExecOnLimitReached(c.Writer, c.Request)
				c.Abort()

				// TODO: add metrics on limited requests
			} else {
				c.Next()
			}
		}
	}
}

func limitReach(w http.ResponseWriter, r *http.Request) {
	// TODO: export metrics here group by r.Method + r.URL
}

func setupLimiter(limit float64) *limiter.Limiter {
	return tollbooth.NewLimiter(limit, &limiter.ExpirableOptions{
		DefaultExpirationTTL: time.Second,
	}).SetOnLimitReached(limitReach)
}
