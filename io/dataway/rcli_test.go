// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	reflect "reflect"
	T "testing"
	"time"

	rhttp "github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
	ihttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/http"
)

func TestTimeout(t *T.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second)
	}))

	cli := newRetryCli(&ihttp.Options{}, time.Millisecond)

	req, err := http.NewRequest("POST", ts.URL, nil)
	require.NoError(t, err)

	x, err := rhttp.FromRequest(req)
	require.NoError(t, err)

	_, err = cli.Do(x)

	t.Logf("raw error: %s", err)

	//nolint:errorlint
	switch e := errors.Unwrap(err).(type) {
	case *url.Error:
		t.Logf("timeout: %v", e.Timeout())
	default:
		t.Logf("error: %s, type: %s", e.Error(), reflect.TypeOf(err))
	}
}

func Test5XX(t *T.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	cli := newRetryCli(&ihttp.Options{}, time.Millisecond)

	req, err := http.NewRequest("POST", ts.URL, nil)
	require.NoError(t, err)

	x, err := rhttp.FromRequest(req)
	require.NoError(t, err)

	_, err = cli.Do(x)

	if err != nil {
		switch e := errors.Unwrap(err).(type) { //nolint:errorlint
		case *url.Error:
			t.Logf("timeout: %v", e.Timeout())
		case nil:
			t.Logf("error: %s, type: %s", err.Error(), reflect.TypeOf(err))
		default:
			t.Logf("error: %s, type: %s", e.Error(), reflect.TypeOf(e))
		}
	}
}
