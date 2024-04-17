// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"net/http"
	"strconv"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	httpFailRatio      = 0 // %n
	httpFailStart      time.Time
	httpFailDuration   time.Duration
	httpMockedFailResp *http.Response
)

// nolint: gochecknoinits
func init() {
	if v := datakit.GetEnv("ENV_DEBUG_HTTP_FAIL_RATIO"); v != "" {
		if x, err := strconv.ParseInt(v, 10, 64); err == nil {
			httpFailRatio = int(x)
			httpFailStart = time.Now()

			httpMockedFailResp = &http.Response{
				Status:     http.StatusText(500),
				StatusCode: 500,
			}
		}
	}

	if v := datakit.GetEnv("ENV_DEBUG_HTTP_FAIL_DURATION"); v != "" {
		if x, err := time.ParseDuration(v); err == nil {
			httpFailDuration = x
		}
	}
}
