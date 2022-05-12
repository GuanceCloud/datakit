// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
)

type retrycliLogger struct{}

func (l *retrycliLogger) Error(msg string, kvs ...interface{}) {
	log.Errorf(msg, kvs...)
}

func (l *retrycliLogger) Info(msg string, kvs ...interface{}) {
	log.Infof(msg, kvs...)
}

func (l *retrycliLogger) Debug(msg string, kvs ...interface{}) {
	log.Debugf(msg, kvs...)
}

func (l *retrycliLogger) Warn(msg string, kvs ...interface{}) {
	log.Warnf(msg, kvs...)
}

func retryCallback(_ retryablehttp.Logger, r *http.Request, n int) {
	if n == 0 {
		return
	}

	log.Warnf("retry %d time on API %s", n, r.URL.Path)
}

func backoffCallback(min, max time.Duration, n int, resp *http.Response) time.Duration {
	switch n {
	case 1:
		return time.Second
	case 2:
		return time.Second * 2
	case 3:
		return time.Second * 3

	default: // should not been here
		return 0
	}
}

const maxretry = 3

func newRetryCli(opt *ihttp.Options, timeout time.Duration) *retryablehttp.Client {
	retrycli := retryablehttp.NewClient()

	retrycli.RetryWaitMin = time.Second
	retrycli.RetryWaitMax = time.Second * 3
	retrycli.RetryMax = maxretry

	retrycli.RequestLogHook = retryCallback
	retrycli.Backoff = backoffCallback

	retrycli.HTTPClient = ihttp.Cli(opt)
	retrycli.HTTPClient.Timeout = timeout
	retrycli.Logger = &retrycliLogger{}

	log.Debugf("httpCli: %p", retrycli.HTTPClient.Transport)
	return retrycli
}
