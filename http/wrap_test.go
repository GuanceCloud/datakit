package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestLimitWrap(t *testing.T) {
	var limit float64 = 1000.0
	reqLimiter = setupLimiter(limit)

	r := gin.New()
	apiHandler := func(c *gin.Context) {
		c.Data(200, "", nil)
	}

	r.GET("/", ginWraper(reqLimiter), apiHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()
	time.Sleep(time.Second)

	total := 0
	limited := 0
	passed := 0
	round := 0

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for {
		resp, err := http.Get(ts.URL)
		if err != nil {
			t.Error(err)
		}

		resp.Body.Close()
		time.Sleep(time.Microsecond)

		switch resp.StatusCode {
		case 200:
			passed++
		case 429:
			limited++
		}
		total++
		if total > 10000 {
			break
		}

		select {
		case <-tick.C:
			round++
			rate := float64(passed) / float64(round)
			tu.Assert(t, rate < limit, "expect %f < %f", rate, limit)

			t.Logf("rate: %f, passed: %d, limited: %d, total: %d", rate, passed, limited, total)
		default:
		}
	}
}
