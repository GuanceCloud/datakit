// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package recorder

import (
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecorder(t *T.T) {
	t.Run(`invalid-category`, func(t *T.T) {
		var (
			r = &Recorder{
				Enabled:    true,
				Path:       t.TempDir(),
				Categories: []string{""}, // not empty
			}
			err error
		)

		r, err = SetupRecorder(r)
		require.NoError(t, err)

		g := point.NewRander()
		pts := g.Rand(100)

		require.NoError(t, r.Record(pts, point.Metric, "some"))
		assert.Equal(t, int64(0), r.totalRecordedPoints.Load())
	})

	t.Run(`invalid-input`, func(t *T.T) {
		var (
			r = &Recorder{
				Enabled: true,
				Path:    t.TempDir(),
				Inputs:  []string{""}, // not empty
			}
			err error
		)

		r, err = SetupRecorder(r)
		require.NoError(t, err)

		g := point.NewRander()
		pts := g.Rand(100)

		require.NoError(t, r.Record(pts, point.Metric, "some"))
		assert.Equal(t, int64(0), r.totalRecordedPoints.Load())
	})

	t.Run(`record-lp-points`, func(t *T.T) {
		var (
			w = &mockw{}
			r = &Recorder{
				Path:    t.TempDir(),
				Enabled: true,
				Inputs:  []string{"some"},
			}
			err error
		)

		r, err = SetupRecorder(r)
		require.NoError(t, err)

		r.w = w // reset mocked writer

		g := point.NewRander()
		npts := 3
		rpts := g.Rand(npts)

		require.NoError(t, r.Record(rpts, point.Metric, "some"))
		assert.Equal(t, int64(npts), r.totalRecordedPoints.Load())
		assert.Len(t, w.data, 1)

		dec := point.GetDecoder(point.WithDecEncoding(point.LineProtocol))
		defer point.PutDecoder(dec)

		for k, arr := range w.data {
			t.Logf("key: %s", k)
			t.Logf("data: %s", string(arr[0]))

			pts, err := dec.Decode(arr[0])
			require.NoError(t, err)

			for idx, pt := range pts {
				assert.True(t, pt.Equal(rpts[idx]))
			}
		}
	})

	t.Run(`record-pb-points`, func(t *T.T) {
		var (
			w = &mockw{}
			r = &Recorder{
				Path:     t.TempDir(),
				Enabled:  true,
				Inputs:   []string{"some"},
				Encoding: point.Protobuf.String(),
			}
			err error
		)

		r, err = SetupRecorder(r)
		require.NoError(t, err)

		r.w = w // reset mocked writer

		g := point.NewRander()
		npts := 3
		rpts := g.Rand(npts)

		require.NoError(t, r.Record(rpts, point.Metric, "some"))
		assert.Equal(t, int64(npts), r.totalRecordedPoints.Load())
		assert.Len(t, w.data, 1)

		for k, arr := range w.data {
			t.Logf("key: %s", k)
			t.Logf("data: %s", string(arr[0]))

			pts, err := PBJson2pts(arr[0])
			require.NoError(t, err)

			for idx, pt := range pts {
				assert.True(t, pt.Equal(rpts[idx]))
			}
		}
	})

	t.Run(`record-json-points`, func(t *T.T) {
		var (
			w = &mockw{}
			r = &Recorder{
				Path:     t.TempDir(),
				Enabled:  true,
				Inputs:   []string{"some"},
				Encoding: point.JSON.String(),
			}
			err error
		)

		r, err = SetupRecorder(r)
		require.NoError(t, err)

		r.w = w // reset mocked writer

		g := point.NewRander()
		npts := 3
		rpts := g.Rand(npts)

		require.Error(t, r.Record(rpts, point.Metric, "some")) // not support
		assert.Equal(t, int64(0), r.totalRecordedPoints.Load())
		assert.Len(t, w.data, 0)
	})

	t.Run(`record-duration`, func(t *T.T) {
		var (
			w    = &mockw{}
			nsec = 3
			npts = 3
			r    = &Recorder{
				Path:     t.TempDir(),
				Enabled:  true,
				Inputs:   []string{"some"},
				Encoding: point.Protobuf.String(),
				Duration: time.Second * time.Duration(nsec),
			}
			err error
		)

		r, err = SetupRecorder(r)
		require.NoError(t, err)

		r.w = w // reset mocked writer

		n := 0
		for {
			g := point.NewRander()
			rpts := g.Rand(npts)

			require.NoError(t, r.Record(rpts, point.Metric, "some")) // not support
			n++
			time.Sleep(time.Second)

			if n >= nsec*2 {
				break
			}
		}

		assert.True(t, int64(npts*n) >= r.totalRecordedPoints.Load())

		t.Logf("record %d points, raw recorded %d points", r.totalRecordedPoints.Load(), npts*n)
	})
}

type mockw struct {
	data map[string][][]byte
}

func (w *mockw) write(f string, data []byte) error {
	if w.data == nil {
		w.data = map[string][][]byte{}
	}

	w.data[f] = append(w.data[f], data)
	return nil
}

func TestPBJson2pts(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		pbj := []byte(`
{
  "arr": [
    {
      "name": "cpu",
      "time": "1698043300404819000"
    },

    {
      "name": "disk",
      "time": "1698043387763427000"
    }
  ]
}`)
		pts, err := PBJson2pts(pbj)
		require.NoError(t, err)
		assert.Len(t, pts, 2)

		assert.Equal(t, int64(1698043300404819000), pts[0].Time().UnixNano())
		assert.Equal(t, int64(1698043387763427000), pts[1].Time().UnixNano())
	})
}
