// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package refertable

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

type cliLogger struct{}

func (*cliLogger) Error(msg string, kv ...any) {
	l.Error(msg, kv)
}

func (*cliLogger) Warn(msg string, kv ...any) {
	l.Warn(msg, kv)
}

func (*cliLogger) Info(msg string, kv ...any) {
	l.Info(msg, kv)
}

func (*cliLogger) Debug(msg string, kv ...any) {
	l.Debug(msg, kv)
}

func newRetryCli(client *http.Client, timeout time.Duration) *retryablehttp.Client {
	cli := retryablehttp.NewClient()

	cli.RetryMax = 3
	cli.RetryWaitMin = time.Second
	cli.RetryWaitMax = time.Second * 5

	cli.RequestLogHook = func(_ retryablehttp.Logger, r *http.Request, n int) {
		if n > 0 {
			l.Warnf("retry %d time on API %s", n, r.URL.Path)
		}
	}

	cli.HTTPClient = client
	cli.HTTPClient.Timeout = timeout
	cli.Logger = &cliLogger{}

	return cli
}
