// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
)

func TestSinkerSetup(t *T.T) {
	t.Run(`token-not-set-in-URL`, func(t *T.T) {
		tkn := "tkn_xxxxxxxxxxxxxxxxxxxxxxxx"
		s := &Sinker{
			Categories:      []string{"T", "M"},
			Filters:         nil,
			URL:             "https://some.host.com",
			TokenDeprecated: tkn,
		}

		assert.NoError(t, s.Setup())
		assert.Equal(t, tkn, s.ep.token)
	})

	t.Run(`token-set-in-URL`, func(t *T.T) {
		tknX := "tkn_xxxxxxxxxxxxxxxxxxxxxxxx"
		tknY := "tkn_yyyyyyyyyyyyyyyyyyyyyyyy"
		s := &Sinker{
			Categories:      []string{"T", "M"},
			Filters:         nil,
			URL:             fmt.Sprintf("https://some.host.com?token=%s", tknY),
			TokenDeprecated: tknX,
		}

		assert.NoError(t, s.Setup())
		assert.Equal(t, tknY, s.ep.token)
	})
}

func TestSinkerWrite(t *T.T) {
	t.Run(`invalid-condition`, func(t *T.T) {
		s := &Sinker{
			Categories: []string{"L"},
			Filters:    []string{`some-invalid-filter-expr`},
		}

		err := s.Setup()
		assert.Error(t, err)
		t.Logf("Setup: %s", err)
	})

	t.Run(`no-condition`, func(t *T.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			defer r.Body.Close()

			x, err := uhttp.Unzip(body)
			assert.NoError(t, err)

			assert.NoError(t, err)
			assert.Equal(t, []byte(`test fa="a",fb="b" 123
test f1="1",f2="2" 123`), x)

			t.Logf("body: %q", x)

			w.WriteHeader(200)
		}))

		defer ts.Close()

		time.Sleep(time.Second)

		s := &Sinker{
			Categories: []string{"L"},
			Filters:    nil,
			URL:        ts.URL,
		}

		require.NoError(t, s.Setup())
		require.Len(t, s.conditions, 0)

		pts := []*dkpt.Point{
			dkpt.MustNewPoint("test", nil, map[string]any{"fa": "a", "fb": "b"},
				&dkpt.PointOption{Category: datakit.Logging, Time: time.Unix(0, 123)}),

			dkpt.MustNewPoint("test", nil, map[string]any{"f1": "1", "f2": "2"},
				&dkpt.PointOption{Category: datakit.Logging, Time: time.Unix(0, 123)}),
		}

		remains, err := s.sink(point.Logging, pts)
		require.NoError(t, err)

		assert.Lenf(t, remains, 0, "expect 0 remains, got %+#v", remains)

		mfs := metrics.MustGather()
		t.Logf("metric: %s", metrics.MetricFamily2Text(mfs))

		// ensure metrics ok
		m := metrics.GetMetricOnLabels(mfs, "datakit_io_dataway_sink_total", point.Logging.String())
		require.NotNil(t, m)
		assert.Equal(t, float64(1), m.GetCounter().GetValue())

		m = metrics.GetMetricOnLabels(mfs, "datakit_io_dataway_sink_point_total", point.Logging.String(), http.StatusText(http.StatusOK))
		require.NotNil(t, m)
		assert.Equal(t, float64(2), m.GetCounter().GetValue())

		t.Cleanup(func() {
			metricsReset()
		})
	})

	t.Run(`basic`, func(t *T.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)

			defer r.Body.Close()

			assert.NoError(t, err)

			x, err := uhttp.Unzip(body)
			assert.NoError(t, err)

			assert.Equal(t, []byte(`test fa="a",fb="b" 123`), x)

			t.Logf("body: %q", x)

			w.WriteHeader(200)
		}))

		time.Sleep(time.Second)

		s := &Sinker{
			Categories: []string{"L"},
			Filters:    []string{"{fa='a'}", "{fb='b'}"},
			URL:        ts.URL,
		}

		require.NoError(t, s.Setup())

		require.Equal(t, 2, len(s.conditions))

		mustSinked := dkpt.MustNewPoint("test", nil, map[string]any{"fa": "a", "fb": "b"},
			&dkpt.PointOption{Category: datakit.Logging, Time: time.Unix(0, 123)})

		notSinked := dkpt.MustNewPoint("test", nil, map[string]any{"f1": "1", "f2": "2"},
			&dkpt.PointOption{Category: datakit.Logging, Time: time.Unix(0, 123)})

		pts := []*dkpt.Point{mustSinked, notSinked}
		remains, err := s.sink(point.Logging, pts)
		require.NoError(t, err)

		assert.Lenf(t, remains, 1, "expect only 1 remains, got %+#v", remains)
	})
}
