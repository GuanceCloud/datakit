package output

import (
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

type retrycliLogger struct{}

func (log *retrycliLogger) Error(msg string, kvs ...interface{}) {
	l.Errorf(msg, kvs...)
}

func (log *retrycliLogger) Info(msg string, kvs ...interface{}) {
	l.Infof(msg, kvs...)
}

func (log *retrycliLogger) Debug(msg string, kvs ...interface{}) {
	l.Debugf(msg, kvs...)
}

func (log *retrycliLogger) Warn(msg string, kvs ...interface{}) {
	l.Warnf(msg, kvs...)
}

func retryCallback(_ retryablehttp.Logger, r *http.Request, n int) {
	if n == 0 {
		return
	}

	l.Warnf("retry %d time on API %s", n, r.URL.Path)
}
