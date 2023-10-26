// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package etcd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	T "testing"

	"github.com/stretchr/testify/assert"
)

type taggerMock struct {
	hostTags, electionTags map[string]string
}

func (m *taggerMock) HostTags() map[string]string {
	return m.hostTags
}

func (m *taggerMock) ElectionTags() map[string]string {
	return m.electionTags
}

// go test -v -timeout 30s -run ^TestInput$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/etcd
func TestInput(t *T.T) {
	t.Run("basic", func(t *T.T) {
		body := `# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		t.Log(srv.URL)
		defer srv.Close()

		t.Logf("url: %s", srv.URL)

		inp := defaultInput()
		inp.URLs = []string{srv.URL}

		inp.Tagger = &taggerMock{
			hostTags: map[string]string{
				"host":  "foo",
				"hello": "world",
			},

			electionTags: map[string]string{
				"project": "foo",
				"cluster": "bar",
			},
		}

		inp.Init()

		pts, err := inp.Collect()
		assert.NoError(t, err)

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has("cause"))
			assert.True(t, pt.Tags().Has("project"))
			assert.True(t, pt.Tags().Has("cluster"))

			assert.Equal(t, float64(0.0), pt.Get("metric_handler_errors_total").(float64))

			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run("ignore-tags", func(t *T.T) {
		body := `# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding",ignore_me="some"} 0
`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		t.Log(srv.URL)
		defer srv.Close()

		t.Logf("url: %s", srv.URL)

		inp := defaultInput()
		inp.URLs = []string{srv.URL}
		inp.TagsIgnore = []string{"ignore_me"}

		inp.Init()

		pts, err := inp.Collect()
		assert.NoError(t, err)

		for _, pt := range pts {
			assert.False(t, pt.Tags().Has("ignore_me"))

			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run("set-measurement", func(t *T.T) {
		body := `# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		t.Log(srv.URL)
		defer srv.Close()

		t.Logf("url: %s", srv.URL)

		inp := defaultInput()
		inp.URLs = []string{srv.URL}
		inp.MeasurementName = "some"

		inp.Init()

		pts, err := inp.Collect()
		assert.NoError(t, err)

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has("cause"))
			assert.True(t, pt.Name() == "some")

			t.Logf("%s", pt.Pretty())
		}
	})
}
