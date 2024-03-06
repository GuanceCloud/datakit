// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package prom

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
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
func (g *taggerMock) UpdateVersion() {}
func (g *taggerMock) Updated() bool  { return false }

func TestInputNoBatch(t *T.T) {
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

		pts, err := inp.collectFormSource(inp.URLs[0])
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has("instance"))
			assert.True(t, pt.Tags().Has("cause"))

			assert.True(t, pt.Tags().Has("project"))
			assert.True(t, pt.Tags().Has("cluster"))

			assert.Equal(t, float64(0.0), pt.Get("metric_handler_errors_total").(float64))

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

		pts, err := inp.collectFormSource(inp.URLs[0])
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has("instance"))
			assert.True(t, pt.Tags().Has("info1"))

			t.Logf("%s", pt.Pretty())
		}

		// info type disabled
		inp.DisableInfoTag = true
		inp.Init()

		pts, err = inp.collectFormSource(inp.URLs[0])
		assert.NoError(t, err)

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has("instance"))
			assert.False(t, pt.Tags().Has("info1"))

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

		pts, err := inp.collectFormSource(inp.URLs[0])
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has("instance"))
			assert.False(t, pt.Tags().Has("ignore_me"))

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

		pts, err := inp.collectFormSource(inp.URLs[0])
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has("instance"))
			assert.False(t, pt.Tags().Has("cause"))
			assert.True(t, pt.Tags().Has("__cause"))

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

		pts, err := inp.collectFormSource(inp.URLs[0])
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has("instance"))
			assert.True(t, pt.Tags().Has("cause"))
			assert.True(t, pt.Tags().Has("service"))
			assert.True(t, pt.Fields().Has("status"))

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

		pts, err := inp.collectFormSource(inp.URLs[0])
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has("instance"))
			assert.True(t, pt.Tags().Has("cause"))
			assert.Equal(t, pt.Name(), "some")

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

		pts, err := inp.collectFormSource(inp.URLs[0])
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has("instance"))
			assert.True(t, pt.Tags().Has("cause"))
			assert.Equal(t, pt.Name(), "some_promhttp")

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

		pts, err := inp.collectFormSource(inp.URLs[0])
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has("instance"))
			assert.True(t, pt.Tags().Has("cause"))
			assert.Equal(t, pt.Name(), "morp")

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

		pts, err := inp.collectFormSource(inp.URLs[0])
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			assert.False(t, pt.Tags().Has("instance"))
			assert.True(t, pt.Tags().Has("cause"))

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

		pts, err := inp.collectFormSource(inp.URLs[0])
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			causeKV := pt.Get("cause")
			if causeKV != nil {
				causeValue := causeKV.(string)
				if causeValue == "encoding-1" || causeValue == "encoding-2" {
					t.Errorf("got error KV %s:%s", "cause", causeValue)
				}
			}

			someKV := pt.Get("some")
			if someKV != nil {
				someValue := someKV.(string)
				if someValue == "foo-1" || someValue == "foo-3" {
					t.Errorf("got error KV %s:%s", "some", someValue)
				}
			}

			t.Logf("%s", pt.Pretty())
		}
	})
}

func TestInputBatch(t *T.T) {
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

		inp.StreamSize = 1
		ptCh := make(chan []*point.Point, 1)
		points := []*point.Point{}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for v := range ptCh {
				points = append(points, v...)
			}
			wg.Done()
		}()
		inp.callbackFunc = func(pts []*point.Point) error {
			for _, v := range inp.urlTags[inp.currentURL] {
				for _, pt := range pts {
					pt.AddTag(v.key, v.value)
				}
			}
			ptCh <- pts
			return nil
		}
		inp.Init()
		_, err := inp.collectFormSource(inp.URLs[0])
		close(ptCh)
		assert.NoError(t, err)
		wg.Wait()
		if len(points) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range points {
			assert.True(t, pt.Tags().Has("instance"))
			assert.True(t, pt.Tags().Has("cause"))
			assert.True(t, pt.Tags().Has("project"))
			assert.True(t, pt.Tags().Has("cluster"))

			assert.Equal(t, float64(0.0), pt.Get("metric_handler_errors_total").(float64))

			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run("info-type", func(t *T.T) {
		body := `# TYPE some_info info
some_info{info1="data1"} 0
# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
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

		inp.StreamSize = 1
		ptCh := make(chan []*point.Point, 1)
		points := []*point.Point{}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for v := range ptCh {
				points = append(points, v...)
			}
			wg.Done()
		}()
		inp.callbackFunc = func(pts []*point.Point) error {
			for _, v := range inp.urlTags[inp.currentURL] {
				for _, pt := range pts {
					pt.AddTag(v.key, v.value)
				}
			}
			ptCh <- pts
			return nil
		}
		inp.Init()
		_, err := inp.collectFormSource(inp.URLs[0])
		close(ptCh)
		assert.NoError(t, err)
		wg.Wait()
		if len(points) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range points {
			assert.True(t, pt.Tags().Has("instance"))
			assert.True(t, pt.Tags().Has("info1"))

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

		// info type disabled
		inp.DisableInfoTag = true

		inp.StreamSize = 1
		ptCh := make(chan []*point.Point, 1)
		points := []*point.Point{}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for v := range ptCh {
				points = append(points, v...)
			}
			wg.Done()
		}()
		inp.callbackFunc = func(pts []*point.Point) error {
			for _, v := range inp.urlTags[inp.currentURL] {
				for _, pt := range pts {
					pt.AddTag(v.key, v.value)
				}
			}
			ptCh <- pts
			return nil
		}
		inp.Init()
		_, err := inp.collectFormSource(inp.URLs[0])
		close(ptCh)
		assert.NoError(t, err)
		wg.Wait()
		if len(points) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range points {
			assert.True(t, pt.Tags().Has("instance"))
			assert.False(t, pt.Tags().Has("info1"))

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

		inp.StreamSize = 1
		ptCh := make(chan []*point.Point, 1)
		points := []*point.Point{}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for v := range ptCh {
				points = append(points, v...)
			}
			wg.Done()
		}()
		inp.callbackFunc = func(pts []*point.Point) error {
			for _, v := range inp.urlTags[inp.currentURL] {
				for _, pt := range pts {
					pt.AddTag(v.key, v.value)
				}
			}
			ptCh <- pts
			return nil
		}
		inp.Init()
		_, err := inp.collectFormSource(inp.URLs[0])
		close(ptCh)
		assert.NoError(t, err)
		wg.Wait()
		if len(points) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range points {
			assert.True(t, pt.Tags().Has("instance"))
			assert.False(t, pt.Tags().Has("ignore_me"))

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

		inp.StreamSize = 1
		ptCh := make(chan []*point.Point, 1)
		points := []*point.Point{}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for v := range ptCh {
				points = append(points, v...)
			}
			wg.Done()
		}()
		inp.callbackFunc = func(pts []*point.Point) error {
			for _, v := range inp.urlTags[inp.currentURL] {
				for _, pt := range pts {
					pt.AddTag(v.key, v.value)
				}
			}
			ptCh <- pts
			return nil
		}
		inp.Init()
		_, err := inp.collectFormSource(inp.URLs[0])
		close(ptCh)
		assert.NoError(t, err)
		wg.Wait()
		if len(points) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range points {
			assert.True(t, pt.Tags().Has("instance"))
			assert.False(t, pt.Tags().Has("cause"))
			assert.True(t, pt.Tags().Has("__cause"))

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

		inp.StreamSize = 1
		ptCh := make(chan []*point.Point, 1)
		points := []*point.Point{}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for v := range ptCh {
				points = append(points, v...)
			}
			wg.Done()
		}()
		inp.callbackFunc = func(pts []*point.Point) error {
			for _, v := range inp.urlTags[inp.currentURL] {
				for _, pt := range pts {
					pt.AddTag(v.key, v.value)
				}
			}
			ptCh <- pts
			return nil
		}
		inp.Init()
		_, err := inp.collectFormSource(inp.URLs[0])
		close(ptCh)
		assert.NoError(t, err)
		wg.Wait()
		if len(points) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range points {
			assert.True(t, pt.Tags().Has("instance"))
			assert.True(t, pt.Tags().Has("cause"))
			assert.True(t, pt.Tags().Has("service"))
			assert.True(t, pt.Fields().Has("status"))

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

		inp.StreamSize = 1
		ptCh := make(chan []*point.Point, 1)
		points := []*point.Point{}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for v := range ptCh {
				points = append(points, v...)
			}
			wg.Done()
		}()
		inp.callbackFunc = func(pts []*point.Point) error {
			for _, v := range inp.urlTags[inp.currentURL] {
				for _, pt := range pts {
					pt.AddTag(v.key, v.value)
				}
			}
			ptCh <- pts
			return nil
		}
		inp.Init()
		_, err := inp.collectFormSource(inp.URLs[0])
		close(ptCh)
		assert.NoError(t, err)
		wg.Wait()
		if len(points) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range points {
			assert.True(t, pt.Tags().Has("instance"))
			assert.True(t, pt.Tags().Has("cause"))
			assert.True(t, pt.Name() == "some")

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

		inp.StreamSize = 1
		ptCh := make(chan []*point.Point, 1)
		points := []*point.Point{}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for v := range ptCh {
				points = append(points, v...)
			}
			wg.Done()
		}()
		inp.callbackFunc = func(pts []*point.Point) error {
			for _, v := range inp.urlTags[inp.currentURL] {
				for _, pt := range pts {
					pt.AddTag(v.key, v.value)
				}
			}
			ptCh <- pts
			return nil
		}
		inp.Init()
		_, err := inp.collectFormSource(inp.URLs[0])
		close(ptCh)
		assert.NoError(t, err)
		wg.Wait()
		if len(points) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range points {
			assert.True(t, pt.Tags().Has("instance"))
			assert.True(t, pt.Tags().Has("cause"))
			assert.True(t, pt.Name() == "some_promhttp")

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

		inp.StreamSize = 1
		ptCh := make(chan []*point.Point, 1)
		points := []*point.Point{}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for v := range ptCh {
				points = append(points, v...)
			}
			wg.Done()
		}()
		inp.callbackFunc = func(pts []*point.Point) error {
			for _, v := range inp.urlTags[inp.currentURL] {
				for _, pt := range pts {
					pt.AddTag(v.key, v.value)
				}
			}
			ptCh <- pts
			return nil
		}
		inp.Init()
		_, err := inp.collectFormSource(inp.URLs[0])
		close(ptCh)
		assert.NoError(t, err)
		wg.Wait()
		if len(points) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range points {
			assert.True(t, pt.Tags().Has("instance"))
			assert.True(t, pt.Tags().Has("cause"))
			assert.True(t, pt.Name() == "morp")

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

		inp.StreamSize = 1
		ptCh := make(chan []*point.Point, 1)
		points := []*point.Point{}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for v := range ptCh {
				points = append(points, v...)
			}
			wg.Done()
		}()
		inp.callbackFunc = func(pts []*point.Point) error {
			for _, v := range inp.urlTags[inp.currentURL] {
				for _, pt := range pts {
					pt.AddTag(v.key, v.value)
				}
			}
			ptCh <- pts
			return nil
		}
		inp.Init()
		_, err := inp.collectFormSource(inp.URLs[0])
		close(ptCh)
		assert.NoError(t, err)
		wg.Wait()
		if len(points) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range points {
			assert.False(t, pt.Tags().Has("instance"))
			assert.True(t, pt.Tags().Has("cause"))

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

		inp.StreamSize = 1
		ptCh := make(chan []*point.Point, 1)
		points := []*point.Point{}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for v := range ptCh {
				points = append(points, v...)
			}
			wg.Done()
		}()
		inp.callbackFunc = func(pts []*point.Point) error {
			for _, v := range inp.urlTags[inp.currentURL] {
				for _, pt := range pts {
					pt.AddTag(v.key, v.value)
				}
			}
			ptCh <- pts
			return nil
		}
		inp.Init()
		_, err := inp.collectFormSource(inp.URLs[0])
		close(ptCh)
		assert.NoError(t, err)
		wg.Wait()
		if len(points) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range points {
			causeKV := pt.Get("cause")
			if causeKV != nil {
				causeValue := causeKV.(string)
				if causeValue == "encoding-1" || causeValue == "encoding-2" {
					t.Errorf("got error KV %s:%s", "cause", causeValue)
				}
			}

			someKV := pt.Get("some")
			if someKV != nil {
				someValue := someKV.(string)
				if someValue == "foo-1" || someValue == "foo-3" {
					t.Errorf("got error KV %s:%s", "some", someValue)
				}
			}

			t.Logf("%s", pt.Pretty())
		}
	})
}

func TestBatchParser(t *T.T) {
	t.Run("basic", func(t *T.T) {
		body := `# HELP dql_num_total dql metric time related indicators unit(ms).
		# TYPE dql_num_total counter
		dql_num_total{ws_uuid="wksp_xxxxxxx1b7c6a49a489bfc31ba629a6cc",data_type="object"} 1
		dql_num_total{ws_uuid="wksp_xxxxxxx51a7f24110941dad5d51ba29f1",data_type="logging"} 1
		dql_num_total{ws_uuid="wksp_xxxxxxx6d42af4c9596ecefb5e5dbfa3a",data_type="metric"} 1
		dql_num_total{ws_uuid="wksp_xxxxxxx51a7f24110941dad5d51ba29f1",data_type="metric"} 2
		dql_num_total{ws_uuid="wksp_xxxxxxx1e84ad45ac8c73869aab4abf72",data_type="logging"} 2
		dql_num_total{ws_uuid="wksp_xxxxxxx6c2b024301a0b1d139e756b61e",data_type="metric"} 2
		# HELP dql_query_fail_num_total dql metric time related indicators unit(ms).
		# TYPE dql_query_fail_num_total counter
		dql_query_fail_num_total{ws_uuid="wksp_xxxxxxx1b7c6a49a489bfc31ba629a6cc",data_type="object"} 0
		dql_query_fail_num_total{ws_uuid="wksp_xxxxxxx51a7f24110941dad5d51ba29f1",data_type="logging"} 0
		dql_query_fail_num_total{ws_uuid="wksp_xxxxxxx6d42af4c9596ecefb5e5dbfa3a",data_type="metric"} 0
		dql_query_fail_num_total{ws_uuid="wksp_xxxxxxx51a7f24110941dad5d51ba29f1",data_type="metric"} 0
		dql_query_fail_num_total{ws_uuid="wksp_xxxxxxx1e84ad45ac8c73869aab4abf72",data_type="logging"} 0
		dql_query_fail_num_total{ws_uuid="wksp_xxxxxxx6c2b024301a0b1d139e756b61e",data_type="metric"} 0
		# HELP dql_avg_cost dql metric time related indicators unit(ms).
		# TYPE dql_avg_cost gauge
		dql_avg_cost{ws_uuid="wksp_xxxxxxx1b7c6a49a489bfc31ba629a6cc",data_type="object"} 4
		dql_avg_cost{ws_uuid="wksp_xxxxxxx51a7f24110941dad5d51ba29f1",data_type="logging"} 31
		dql_avg_cost{ws_uuid="wksp_xxxxxxx6d42af4c9596ecefb5e5dbfa3a",data_type="metric"} 24
		dql_avg_cost{ws_uuid="wksp_xxxxxxx51a7f24110941dad5d51ba29f1",data_type="metric"} 16
		dql_avg_cost{ws_uuid="wksp_xxxxxxx1e84ad45ac8c73869aab4abf72",data_type="logging"} 340.5
		dql_avg_cost{ws_uuid="wksp_xxxxxxx6c2b024301a0b1d139e756b61e",data_type="metric"} 13
		# HELP dql_max_cost dql metric time related indicators unit(ms).
		# TYPE dql_max_cost gauge
		dql_max_cost{ws_uuid="wksp_xxxxxxx1b7c6a49a489bfc31ba629a6cc",data_type="object"} 4
		dql_max_cost{ws_uuid="wksp_xxxxxxx51a7f24110941dad5d51ba29f1",data_type="logging"} 31
		dql_max_cost{ws_uuid="wksp_xxxxxxx6d42af4c9596ecefb5e5dbfa3a",data_type="metric"} 24
		dql_max_cost{ws_uuid="wksp_xxxxxxx51a7f24110941dad5d51ba29f1",data_type="metric"} 20
		dql_max_cost{ws_uuid="wksp_xxxxxxx1e84ad45ac8c73869aab4abf72",data_type="logging"} 466
		dql_max_cost{ws_uuid="wksp_xxxxxxx6c2b024301a0b1d139e756b61e",data_type="metric"} 15
		# HELP dql_min_cost dql metric time related indicators unit(ms).
		# TYPE dql_min_cost gauge
		dql_min_cost{ws_uuid="wksp_xxxxxxx1b7c6a49a489bfc31ba629a6cc",data_type="object"} 4
		dql_min_cost{ws_uuid="wksp_xxxxxxx51a7f24110941dad5d51ba29f1",data_type="logging"} 31
		dql_min_cost{ws_uuid="wksp_xxxxxxx6d42af4c9596ecefb5e5dbfa3a",data_type="metric"} 24
		dql_min_cost{ws_uuid="wksp_xxxxxxx51a7f24110941dad5d51ba29f1",data_type="metric"} 12
		dql_min_cost{ws_uuid="wksp_xxxxxxx1e84ad45ac8c73869aab4abf72",data_type="logging"} 215
		dql_min_cost{ws_uuid="wksp_xxxxxxx6c2b024301a0b1d139e756b61e",data_type="metric"} 11
		# HELP dql_point_total dql metric time related indicators unit(ms).
		# TYPE dql_point_total counter
		dql_point_total{ws_uuid="wksp_xxxxxxx1b7c6a49a489bfc31ba629a6cc",data_type="object"} 2
		dql_point_total{ws_uuid="wksp_xxxxxxx51a7f24110941dad5d51ba29f1",data_type="logging"} 0
		dql_point_total{ws_uuid="wksp_xxxxxxx6d42af4c9596ecefb5e5dbfa3a",data_type="metric"} 8
		dql_point_total{ws_uuid="wksp_xxxxxxx51a7f24110941dad5d51ba29f1",data_type="metric"} 39
		dql_point_total{ws_uuid="wksp_xxxxxxx1e84ad45ac8c73869aab4abf72",data_type="logging"} 2
		dql_point_total{ws_uuid="wksp_xxxxxxx6c2b024301a0b1d139e756b61e",data_type="metric"} 48
		# HELP dql_parser_fail_num_total dql metric time related indicators unit(ms).
		# TYPE dql_parser_fail_num_total counter
		dql_parser_fail_num_total{ws_uuid="wksp_xxxxxxx1b7c6a49a489bfc31ba629a6cc",data_type="object"} 0
		dql_parser_fail_num_total{ws_uuid="wksp_xxxxxxx51a7f24110941dad5d51ba29f1",data_type="logging"} 0
		dql_parser_fail_num_total{ws_uuid="wksp_xxxxxxx6d42af4c9596ecefb5e5dbfa3a",data_type="metric"} 0
		dql_parser_fail_num_total{ws_uuid="wksp_xxxxxxx51a7f24110941dad5d51ba29f1",data_type="metric"} 0
		dql_parser_fail_num_total{ws_uuid="wksp_xxxxxxx1e84ad45ac8c73869aab4abf72",data_type="logging"} 0
		dql_parser_fail_num_total{ws_uuid="wksp_xxxxxxx6c2b024301a0b1d139e756b61e",data_type="metric"} 0
`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		t.Log(srv.URL)
		defer srv.Close()

		t.Logf("url: %s", srv.URL)

		inp := NewProm()
		inp.URLs = []string{srv.URL}

		inp.StreamSize = 1
		ptCh := make(chan []*point.Point, 1)
		points := []*point.Point{}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for v := range ptCh {
				points = append(points, v...)
			}
			wg.Done()
		}()
		inp.callbackFunc = func(pts []*point.Point) error {
			ptCh <- pts
			return nil
		}
		inp.Init()
		err := inp.doCollect()
		close(ptCh)
		assert.NoError(t, err)
		wg.Wait()
		if len(points) == 0 {
			t.Errorf("got nil pts error.")
		}

		want := []string{
			"dql,data_type=logging,ws_uuid=wksp_xxxxxxx1e84ad45ac8c73869aab4abf72 avg_cost=340.5",
			"dql,data_type=logging,ws_uuid=wksp_xxxxxxx1e84ad45ac8c73869aab4abf72 max_cost=466",
			"dql,data_type=logging,ws_uuid=wksp_xxxxxxx1e84ad45ac8c73869aab4abf72 min_cost=215",
			"dql,data_type=logging,ws_uuid=wksp_xxxxxxx1e84ad45ac8c73869aab4abf72 num_total=2",
			"dql,data_type=logging,ws_uuid=wksp_xxxxxxx1e84ad45ac8c73869aab4abf72 parser_fail_num_total=0",
			"dql,data_type=logging,ws_uuid=wksp_xxxxxxx1e84ad45ac8c73869aab4abf72 point_total=2",
			"dql,data_type=logging,ws_uuid=wksp_xxxxxxx1e84ad45ac8c73869aab4abf72 query_fail_num_total=0",
			"dql,data_type=logging,ws_uuid=wksp_xxxxxxx51a7f24110941dad5d51ba29f1 avg_cost=31",
			"dql,data_type=logging,ws_uuid=wksp_xxxxxxx51a7f24110941dad5d51ba29f1 max_cost=31",
			"dql,data_type=logging,ws_uuid=wksp_xxxxxxx51a7f24110941dad5d51ba29f1 min_cost=31",
			"dql,data_type=logging,ws_uuid=wksp_xxxxxxx51a7f24110941dad5d51ba29f1 num_total=1",
			"dql,data_type=logging,ws_uuid=wksp_xxxxxxx51a7f24110941dad5d51ba29f1 parser_fail_num_total=0",
			"dql,data_type=logging,ws_uuid=wksp_xxxxxxx51a7f24110941dad5d51ba29f1 point_total=0",
			"dql,data_type=logging,ws_uuid=wksp_xxxxxxx51a7f24110941dad5d51ba29f1 query_fail_num_total=0",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx51a7f24110941dad5d51ba29f1 avg_cost=16",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx51a7f24110941dad5d51ba29f1 max_cost=20",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx51a7f24110941dad5d51ba29f1 min_cost=12",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx51a7f24110941dad5d51ba29f1 num_total=2",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx51a7f24110941dad5d51ba29f1 parser_fail_num_total=0",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx51a7f24110941dad5d51ba29f1 point_total=39",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx51a7f24110941dad5d51ba29f1 query_fail_num_total=0",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx6c2b024301a0b1d139e756b61e avg_cost=13",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx6c2b024301a0b1d139e756b61e max_cost=15",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx6c2b024301a0b1d139e756b61e min_cost=11",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx6c2b024301a0b1d139e756b61e num_total=2",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx6c2b024301a0b1d139e756b61e parser_fail_num_total=0",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx6c2b024301a0b1d139e756b61e point_total=48",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx6c2b024301a0b1d139e756b61e query_fail_num_total=0",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx6d42af4c9596ecefb5e5dbfa3a avg_cost=24",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx6d42af4c9596ecefb5e5dbfa3a max_cost=24",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx6d42af4c9596ecefb5e5dbfa3a min_cost=24",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx6d42af4c9596ecefb5e5dbfa3a num_total=1",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx6d42af4c9596ecefb5e5dbfa3a parser_fail_num_total=0",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx6d42af4c9596ecefb5e5dbfa3a point_total=8",
			"dql,data_type=metric,ws_uuid=wksp_xxxxxxx6d42af4c9596ecefb5e5dbfa3a query_fail_num_total=0",
			"dql,data_type=object,ws_uuid=wksp_xxxxxxx1b7c6a49a489bfc31ba629a6cc avg_cost=4",
			"dql,data_type=object,ws_uuid=wksp_xxxxxxx1b7c6a49a489bfc31ba629a6cc max_cost=4",
			"dql,data_type=object,ws_uuid=wksp_xxxxxxx1b7c6a49a489bfc31ba629a6cc min_cost=4",
			"dql,data_type=object,ws_uuid=wksp_xxxxxxx1b7c6a49a489bfc31ba629a6cc num_total=1",
			"dql,data_type=object,ws_uuid=wksp_xxxxxxx1b7c6a49a489bfc31ba629a6cc parser_fail_num_total=0",
			"dql,data_type=object,ws_uuid=wksp_xxxxxxx1b7c6a49a489bfc31ba629a6cc point_total=2",
			"dql,data_type=object,ws_uuid=wksp_xxxxxxx1b7c6a49a489bfc31ba629a6cc query_fail_num_total=0",
		}
		got := []string{}
		for _, p := range points {
			s := p.LineProto()
			// remove timestamp
			s = s[:strings.LastIndex(s, " ")]
			got = append(got, s)
		}
		sort.Strings(got)

		if !reflect.DeepEqual(got, want) {
			t.Errorf("DeepEqual error, got = %v, want %v", got, want)
		}
	})
}

// info after point will useless
func TestBatchInfo(t *T.T) {
	t.Run("batch_info", func(t *T.T) {
		body := `# TYPE process_runtime_jvm_buffer_count gauge
# HELP process_runtime_jvm_buffer_count The number of buffers in the pool
process_runtime_jvm_buffer_count{pool="mapped - 'non-volatile memory'"} 0.0 1680231835149
# TYPE otel_scope_info info
# HELP otel_scope_info Scope metadata
otel_scope_info{otel_scope_name="otlp-server"} 1
# TYPE process_runtime_jvm_buffer_count gauge
# HELP process_runtime_jvm_buffer_count The number of buffers in the pool
process_runtime_jvm_buffer_count{pool="mapped - 'non-volatile memory'"} 0.0 1680231835149
# TYPE target info
# HELP target Target metadata
target_info{host_arch="amd64",host_name="DESKTOP-3JJLRI8",os_description="Windows 11 10.0",os_type="windows",process_command_line="D:\\software_installer\\java\\jdk-17\\bin\\java.exe -javaagent:D:/code_zy/opentelemetry-java-instrumentation/javaagent/build/libs/opentelemetry-javaagent-1.24.0-SNAPSHOT.jar -Dotel.traces.exporter=otlp -Dotel.exporter.otlp.endpoint=http://localhost:4317 -Dotel.resource.attributes=service.name=server,username=liu -Dotel.metrics.exporter=otlp -Dotel.propagators=b3 -Dotel.metrics.exporter=prometheus -Dotel.exporter.prometheus.port=10086 -Dotel.exporter.prometheus.resource_to_telemetry_conversion.enabled=true -XX:TieredStopAtLevel=1 -Xverify:none -Dspring.output.ansi.enabled=always -Dcom.sun.management.jmxremote -Dspring.jmx.enabled=true -Dspring.liveBeansView.mbeanDomain -Dspring.application.admin.enabled=true -javaagent:D:\\software_installer\\JetBrains\\IntelliJ IDEA 2022.1.4\\lib\\idea_rt.jar=55275:D:\\software_installer\\JetBrains\\IntelliJ IDEA 2022.1.4\\bin -Dfile.encoding=UTF-8",process_executable_path="D:\\software_installer\\java\\jdk-17\\bin\\java.exe",process_pid="23592",process_runtime_description="Oracle Corporation Java HotSpot(TM) 64-Bit Server VM 17.0.6+9-LTS-190",process_runtime_name="Java(TM) SE Runtime Environment",process_runtime_version="17.0.6+9-LTS-190",service_name="server",telemetry_auto_version="1.24.0-SNAPSHOT",telemetry_sdk_language="java",telemetry_sdk_name="opentelemetry",telemetry_sdk_version="1.23.1",username="liu"} 1
# TYPE process_runtime_jvm_buffer_count gauge
# HELP process_runtime_jvm_buffer_count The number of buffers in the pool
process_runtime_jvm_buffer_count{pool="mapped - 'non-volatile memory'"} 0.0 1680231835149
`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		t.Log(srv.URL)
		defer srv.Close()

		t.Logf("url: %s", srv.URL)

		inp := NewProm()
		inp.URLs = []string{srv.URL}

		inp.StreamSize = 1
		ptCh := make(chan []*point.Point, 1)
		points := []*point.Point{}
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for v := range ptCh {
				points = append(points, v...)
			}
			wg.Done()
		}()
		inp.callbackFunc = func(pts []*point.Point) error {
			ptCh <- pts
			return nil
		}
		inp.Init()
		err := inp.doCollect()
		close(ptCh)
		assert.NoError(t, err)
		wg.Wait()
		if len(points) == 0 {
			t.Errorf("got nil pts error.")
		}

		want := []string{
			"process,pool=mapped\\ -\\ 'non-volatile\\ memory' runtime_jvm_buffer_count=0",
			"process,otel_scope_name=otlp-server,pool=mapped\\ -\\ 'non-volatile\\ memory' runtime_jvm_buffer_count=0",
			"process,host_arch=amd64,host_name=DESKTOP-3JJLRI8,os_description=Windows\\ 11\\ 10.0,os_type=windows,otel_scope_name=otlp-server,pool=mapped\\ -\\ 'non-volatile\\ memory',process_command_line=D:\\software_installer\\java\\jdk-17\\bin\\java.exe\\ -javaagent:D:/code_zy/opentelemetry-java-instrumentation/javaagent/build/libs/opentelemetry-javaagent-1.24.0-SNAPSHOT.jar\\ -Dotel.traces.exporter\\=otlp\\ -Dotel.exporter.otlp.endpoint\\=http://localhost:4317\\ -Dotel.resource.attributes\\=service.name\\=server\\,username\\=liu\\ -Dotel.metrics.exporter\\=otlp\\ -Dotel.propagators\\=b3\\ -Dotel.metrics.exporter\\=prometheus\\ -Dotel.exporter.prometheus.port\\=10086\\ -Dotel.exporter.prometheus.resource_to_telemetry_conversion.enabled\\=true\\ -XX:TieredStopAtLevel\\=1\\ -Xverify:none\\ -Dspring.output.ansi.enabled\\=always\\ -Dcom.sun.management.jmxremote\\ -Dspring.jmx.enabled\\=true\\ -Dspring.liveBeansView.mbeanDomain\\ -Dspring.application.admin.enabled\\=true\\ -javaagent:D:\\software_installer\\JetBrains\\IntelliJ\\ IDEA\\ 2022.1.4\\lib\\idea_rt.jar\\=55275:D:\\software_installer\\JetBrains\\IntelliJ\\ IDEA\\ 2022.1.4\\bin\\ -Dfile.encoding\\=UTF-8,process_executable_path=D:\\software_installer\\java\\jdk-17\\bin\\java.exe,process_pid=23592,process_runtime_description=Oracle\\ Corporation\\ Java\\ HotSpot(TM)\\ 64-Bit\\ Server\\ VM\\ 17.0.6+9-LTS-190,process_runtime_name=Java(TM)\\ SE\\ Runtime\\ Environment,process_runtime_version=17.0.6+9-LTS-190,service_name=server,telemetry_auto_version=1.24.0-SNAPSHOT,telemetry_sdk_language=java,telemetry_sdk_name=opentelemetry,telemetry_sdk_version=1.23.1,username=liu runtime_jvm_buffer_count=0",
		}
		got := []string{}
		for _, p := range points {
			s := p.LineProto()
			// remove timestamp
			s = s[:strings.LastIndex(s, " ")]
			got = append(got, s)
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("DeepEqual error, got = %v, want %v", got, want)
		}
	})
}

// file2http file convert http service
func file2http(w http.ResponseWriter, r *http.Request) {
	fi, err := os.Open("large-metrics.txt")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		fmt.Fprint(w, string(a)+"\n")
	}
}

func TestLargeBatch(t *T.T) {
	t.Run("large_txt_batch", func(t *T.T) {
		var totalMetric int
		var mem runtime.MemStats

		srv := httptest.NewServer(http.HandlerFunc(file2http))
		defer srv.Close()

		feeder := dkio.NewMockedFeeder()
		inp := NewProm()
		inp.Feeder = feeder
		inp.URLs = []string{srv.URL}
		inp.StreamSize = 1

		start := time.Now()
		stopCh := make(chan bool)
		defer close(stopCh)
		go func() {
			for {
				select {
				case <-time.After(10 * time.Millisecond):
					pts, err := feeder.AnyPoints()
					if err != nil {
						assert.NoError(t, err)
						return
					}
					totalMetric += len(pts)
				case <-stopCh:
					return
				}
			}
		}()

		err := inp.collect()

		runtime.ReadMemStats(&mem)
		t.Logf("Alloc      = %v MiB", mem.Alloc/1024/1024)
		t.Logf("TotalAlloc = %v MiB", mem.TotalAlloc/1024/1024)
		t.Logf("Sys        = %v MiB", mem.Sys/1024/1024)
		t.Logf("NumGC      = %v", mem.NumGC)

		assert.NoError(t, err)
		time.Sleep(time.Millisecond * 200)

		if totalMetric < 1 {
			t.Errorf("got nil pts error.")
		}

		t.Logf("get %d metrics", totalMetric)
		t.Logf("time cost %v", time.Since(start))
	})
}

func TestLargeNoBatch(t *T.T) {
	t.Run("large_txt_no_batch", func(t *T.T) {
		var totalMetric int
		var mem runtime.MemStats

		srv := httptest.NewServer(http.HandlerFunc(file2http))
		defer srv.Close()

		feeder := dkio.NewMockedFeeder()
		inp := NewProm()
		inp.Feeder = feeder
		inp.URLs = []string{srv.URL}
		inp.StreamSize = 0

		start := time.Now()
		stopCh := make(chan bool)
		defer close(stopCh)
		go func() {
			for {
				select {
				case <-time.After(10 * time.Millisecond):
					pts, err := feeder.AnyPoints()
					if err != nil {
						assert.NoError(t, err)
						return
					}
					totalMetric += len(pts)
				case <-stopCh:
					return
				}
			}
		}()

		err := inp.collect()

		runtime.ReadMemStats(&mem)
		t.Logf("Alloc      = %v MiB", mem.Alloc/1024/1024)
		t.Logf("TotalAlloc = %v MiB", mem.TotalAlloc/1024/1024)
		t.Logf("Sys        = %v MiB", mem.Sys/1024/1024)
		t.Logf("NumGC      = %v", mem.NumGC)

		assert.NoError(t, err)
		time.Sleep(time.Millisecond * 200)

		if totalMetric < 1 {
			t.Errorf("got nil pts error.")
		}

		t.Logf("get %d metrics", totalMetric)
		t.Logf("time cost %v", time.Since(start))
	})
}

func TestLargeFileBatch(t *T.T) {
	t.Run("large_txt_file_batch", func(t *T.T) {
		var totalMetric int
		var mem runtime.MemStats

		feeder := dkio.NewMockedFeeder()
		inp := NewProm()
		inp.Feeder = feeder
		inp.URLs = []string{"large-metrics.txt"}
		inp.StreamSize = 1

		start := time.Now()
		stopCh := make(chan bool)
		defer close(stopCh)
		go func() {
			for {
				select {
				case <-time.After(10 * time.Millisecond):
					pts, err := feeder.AnyPoints()
					if err != nil {
						assert.NoError(t, err)
						return
					}
					totalMetric += len(pts)
				case <-stopCh:
					return
				}
			}
		}()

		err := inp.collect()
		runtime.ReadMemStats(&mem)
		t.Logf("Alloc      = %v MiB", mem.Alloc/1024/1024)
		t.Logf("TotalAlloc = %v MiB", mem.TotalAlloc/1024/1024)
		t.Logf("Sys        = %v MiB", mem.Sys/1024/1024)
		t.Logf("NumGC      = %v", mem.NumGC)

		assert.NoError(t, err)
		time.Sleep(time.Millisecond * 200)

		if totalMetric < 1 {
			t.Errorf("got nil pts error.")
		}

		t.Logf("get %d metrics", totalMetric)
		t.Logf("time cost %v", time.Since(start))
	})
}
