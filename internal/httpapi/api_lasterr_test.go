// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	T "testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

type mockAPIPutLastError struct {
	em *errMessage
}

func (m *mockAPIPutLastError) FeedLastError(err string, opts ...metrics.LastErrorOption) {
	m.em = &errMessage{
		ErrContent: err,
	}
}

func TestAPIPutLastError(t *T.T) {
	router := gin.New()
	m := &mockAPIPutLastError{}

	router.POST("/", RawHTTPWrapper(nil, apiPutLastError, m))

	ts := httptest.NewServer(router)
	time.Sleep(time.Second)

	defer ts.Close()

	cases := []struct {
		body []byte
		fail bool
	}{
		{
			[]byte(`{"input":"fakeCPU","err_content":"cpu has broken down"}`),
			false,
		},
		{
			[]byte(`{"input":"fakeCPU","err_content":""}`),
			true,
		},
		{
			[]byte(`{"input":"","err_content":"cpu has broken down"}`),
			true,
		},
		{
			[]byte(`{"input":"","err_content":""}`),
			true,
		},
		{
			[]byte(`{"":"fakeCPU","err_content":"cpu has broken down"}`),
			true,
		},
		{
			[]byte(`{"input":"fakeCPU","":"cpu has broken down"}`),
			true,
		},
		{
			[]byte(`{"":"fakeCPU","":"cpu has broken down"}`),
			true,
		},
		{
			[]byte(``),
			true,
		},
	}

	for i, fakeError := range cases {
		t.Run(fmt.Sprintf("%d", i), func(t *T.T) {
			resp, err := http.Post(fmt.Sprintf("%s/", ts.URL), "", bytes.NewReader(fakeError.body))
			assert.NoError(t, err)

			if fakeError.fail {
				assert.NotEqual(t, 2, resp.StatusCode/100)
			} else {
				assert.Equal(t, 2, resp.StatusCode/100)
				assert.NotNil(t, m.em)
			}
		})
	}
}
