// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLimitWrap(t *testing.T) {
	var (
		limit   = 100.0
		ttl     = time.Minute // ttl 需超过本测试的运行时长，这样得到的 limit 比较接近预期值 @expectLimited
		limiter = setupLimiter(limit, ttl)
	)

	r := gin.New()
	apiHandler := func(c *gin.Context) {
		c.Data(200, "", nil)
	}

	r.GET("/", ginLimiter(limiter), apiHandler)

	ts := httptest.NewServer(r)
	defer ts.Close()
	time.Sleep(time.Second)

	var (
		total   = 0
		limited = 0
		passed  = 0

		max   = 10000
		sleep = time.Millisecond // 总运行时长在 10s+ 左右
	)

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	start := time.Now()
	for {
		if total >= max {
			break
		}

		resp, err := http.Get(ts.URL)
		if err != nil {
			t.Error(err)
		}

		if resp != nil {
			resp.Body.Close()
			time.Sleep(sleep)

			switch resp.StatusCode {
			case 200:
				passed++
			case 429:
				limited++
			}
		}

		total++
	}

	expectLimited := float64(time.Since(start)) / float64(time.Second) * limit
	ratio := float64(passed) / expectLimited

	// 此处 passed 总会高出 expectLimited 一截，不清楚是不是 TTL 边界的原因，但不会超过 expectLimited 10%
	assert.Truef(t, ratio <= 1.1, "expected %d, passed %d", int(expectLimited), passed)
	// TTL 一旦小于 for 循环运行时长，此处的偏差就开始变大。当 TTL 大于 运行时长时，不管时 1min 还是 1hour，
	// 比例都在 10% 以内。

	t.Logf("cost %s, expected: %d, passed: %d(ratio: %f), limited: %d, total: %d",
		time.Since(start), int(expectLimited), passed, ratio, limited, total)
}
