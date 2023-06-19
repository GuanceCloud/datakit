// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package prom

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	T "testing"

	"github.com/stretchr/testify/assert"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
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

		inp := NewProm()
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
			assert.True(t, pt.Tags().Has([]byte("instance")))
			assert.True(t, pt.Tags().Has([]byte("cause")))

			assert.True(t, pt.Tags().Has([]byte("project")))
			assert.True(t, pt.Tags().Has([]byte("cluster")))

			assert.Equal(t, float64(0.0), pt.Get([]byte("metric_handler_errors_total")).(float64))

			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run("info-type", func(t *T.T) {
		body := `# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
# TYPE some_info info
some_info{info1="data1"} 0
`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		t.Log(srv.URL)
		defer srv.Close()

		t.Logf("url: %s", srv.URL)

		inp := NewProm()
		inp.URLs = []string{srv.URL}
		inp.Init()

		pts, err := inp.Collect()
		assert.NoError(t, err)

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has([]byte("instance")))
			assert.True(t, pt.Tags().Has([]byte("info1")))

			t.Logf("%s", pt.Pretty())
		}

		// info type disabled
		inp.DisableInfoTag = true
		inp.Init()

		pts, err = inp.Collect()
		assert.NoError(t, err)

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has([]byte("instance")))
			assert.False(t, pt.Tags().Has([]byte("info1")))

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

		inp := NewProm()
		inp.URLs = []string{srv.URL}
		inp.TagsIgnore = []string{"ignore_me"}

		inp.Init()

		pts, err := inp.Collect()
		assert.NoError(t, err)

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has([]byte("instance")))
			assert.False(t, pt.Tags().Has([]byte("ignore_me")))

			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run("tags-rename", func(t *T.T) {
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

		inp := NewProm()
		inp.URLs = []string{srv.URL}
		inp.TagsRename = &iprom.RenameTags{
			Mapping: map[string]string{
				"cause": "__cause",
			},
		}

		inp.Init()

		pts, err := inp.Collect()
		assert.NoError(t, err)

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has([]byte("instance")))
			assert.False(t, pt.Tags().Has([]byte("cause")))
			assert.True(t, pt.Tags().Has([]byte("__cause")))

			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run("as-logging", func(t *T.T) {
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

		inp := NewProm()
		inp.URLs = []string{srv.URL}
		inp.AsLogging = &iprom.AsLogging{
			Enable:  true,
			Service: "as-logging",
		}

		inp.Init()

		pts, err := inp.Collect()
		assert.NoError(t, err)

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has([]byte("instance")))
			assert.True(t, pt.Tags().Has([]byte("cause")))
			assert.True(t, pt.Tags().Has([]byte("service")))
			assert.True(t, pt.Fields().Has([]byte("status")))

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

		inp := NewProm()
		inp.URLs = []string{srv.URL}
		inp.MeasurementName = "some"

		inp.Init()

		pts, err := inp.Collect()
		assert.NoError(t, err)

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has([]byte("instance")))
			assert.True(t, pt.Tags().Has([]byte("cause")))
			assert.True(t, string(pt.Name()) == "some")

			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run("set-measurement-prefix", func(t *T.T) {
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

		inp := NewProm()
		inp.URLs = []string{srv.URL}
		inp.MeasurementPrefix = "some_"

		inp.Init()

		pts, err := inp.Collect()
		assert.NoError(t, err)

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has([]byte("instance")))
			assert.True(t, pt.Tags().Has([]byte("cause")))
			assert.True(t, string(pt.Name()) == "some_promhttp")

			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run("filter-measurements", func(t *T.T) {
		body := `# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
promtcp_metric_handler_errors_total{cause="encoding"} 0
`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		t.Log(srv.URL)
		defer srv.Close()

		t.Logf("url: %s", srv.URL)

		inp := NewProm()
		inp.URLs = []string{srv.URL}
		inp.Measurements = []iprom.Rule{
			{
				Prefix: "prom",
				Name:   "morp",
			},
		}

		inp.Init()

		pts, err := inp.Collect()
		assert.NoError(t, err)

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has([]byte("instance")))
			assert.True(t, pt.Tags().Has([]byte("cause")))
			assert.True(t, string(pt.Name()) == "morp")

			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run("without-instance", func(t *T.T) {
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

		inp := NewProm()
		inp.URLs = []string{srv.URL}
		inp.DisableInstanceTag = true

		inp.Init()

		pts, err := inp.Collect()
		assert.NoError(t, err)

		for _, pt := range pts {
			assert.False(t, pt.Tags().Has([]byte("instance")))
			assert.True(t, pt.Tags().Has([]byte("cause")))

			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run("ignore-tag-kv", func(t *T.T) {
		body := `# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding-1",some="foo-1"} 0
promhttp_metric_handler_errors_total{cause="encoding-2",some="foo-2"} 0
promhttp_metric_handler_errors_total{cause="encoding-3",some="foo-3"} 0
`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		t.Log(srv.URL)
		defer srv.Close()

		t.Logf("url: %s", srv.URL)

		inp := NewProm()
		inp.URLs = []string{srv.URL}
		inp.IgnoreTagKV = map[string][]string{
			"cause": {"encoding-1", "encoding-2"}, // keep `encoding-3'
			"some":  {"foo-1", "foo-3"},           // keep `foo-2'
		}

		inp.Init()

		pts, err := inp.Collect()
		assert.NoError(t, err)

		for _, pt := range pts {
			assert.Equal(t, "encoding-3", string(pt.Get([]byte("cause")).([]byte)))
			assert.Equal(t, "foo-2", string(pt.Get([]byte("some")).([]byte)))

			t.Logf("%s", pt.Pretty())
		}
	})
}
