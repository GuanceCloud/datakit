package exporter

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSampling(t *testing.T) {
	Init(context.Background())

	s := newSampling(context.Background(), &cfg{
		samplingRate: "0.1",
	})

	var pts []*point.Point
	for i := 1; i <= 199; i++ {
		pt := point.NewPoint("a", point.NewKVs(map[string]any{"id": i}))
		pts = append(pts, pt)
	}

	newpts := s.sampling(point.Network.String(), pts)
	assert.Equal(t, 20, len(newpts))
}

func TestSamplingPtsLess1(t *testing.T) {
	Init(context.Background())

	ePtsVec.WithLabelValues("x", point.Network.String()).Add(1000)
	r, rmp := parseRate(
		"0.01", 1000)
	t.Logf("set samping rate %.2f pts/min", r)
	s := &sampling{
		rate:         map[string]float64{},
		ptsPerMinute: rmp,
	}
	lastStats, _ := getTotalPtsMetric()

	ePtsVec.WithLabelValues("y", point.Network.String()).Add(2000)

	_ = s.changeRate(lastStats)

	var pts []*point.Point
	for i := 1; i <= 11; i++ {
		pt := point.NewPoint("a", point.NewKVs(map[string]any{"id": i}))
		pts = append(pts, pt)
	}

	newpts := s.sampling(point.Network.String(), pts)
	assert.Equal(t, 1, len(newpts))
}

func TestSenderFeedBatchesLargePayload(t *testing.T) {
	sender := &Sender{
		ctx:    context.Background(),
		ch:     make(chan *task, 8),
		target: buildTarget("http://127.0.0.1:9529", "http://127.0.0.1:9529"),
	}

	pts := makeTestPoints(maxPtSendCount*2 + 7)
	err := sender.feed("bpf-netlog/netflow", point.Network, pts)
	require.NoError(t, err)

	expectedSizes := []int{maxPtSendCount, maxPtSendCount, 7}
	require.Len(t, sender.ch, len(expectedSizes))

	for _, size := range expectedSizes {
		task := <-sender.ch
		require.NotNil(t, task)
		assert.Len(t, task.data, size)
		assert.Equal(t,
			"http://127.0.0.1:9529/v1/write/network?input="+url.QueryEscape("bpf-netlog/netflow"),
			task.targetURL,
		)
	}
}

func TestSenderRequestPostsPayload(t *testing.T) {
	var (
		gotMethod string
		gotURL    string
		gotType   string
		gotBytes  int
	)

	sender := &Sender{
		httpCli: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				body, err := io.ReadAll(req.Body)
				require.NoError(t, err)
				_ = req.Body.Close()

				gotMethod = req.Method
				gotURL = req.URL.String()
				gotType = req.Header.Get("Content-Type")
				gotBytes = len(body)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
				}, nil
			}),
		},
	}

	targetURL := "http://127.0.0.1:9529/v1/write/network?input=bench"
	err := sender.request(context.Background(), &marshaler{}, &task{
		targetURL: targetURL,
		data:      makeTestPoints(3),
	})
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, targetURL, gotURL)
	assert.Equal(t, point.Protobuf.HTTPContentType(), gotType)
	assert.Greater(t, gotBytes, 0)
}

func TestSenderCollectTaskMergesSameTarget(t *testing.T) {
	sender := &Sender{
		ch: make(chan *task, 4),
	}

	current := &task{
		targetURL: "http://127.0.0.1:9529/v1/write/network?input=bench",
		data:      makeTestPoints(maxPtSendCount - 10),
	}
	sender.ch <- &task{
		targetURL: current.targetURL,
		data:      makeTestPoints(20),
	}
	sender.ch <- &task{
		targetURL: current.targetURL,
		data:      makeTestPoints(8),
	}

	merged, pending := sender.collectTask(current, nil)
	require.NotNil(t, merged)
	require.NotNil(t, pending)
	assert.Len(t, merged.data, maxPtSendCount)
	assert.Len(t, pending.data, 10)
	assert.Equal(t, current.targetURL, pending.targetURL)
}

func TestSenderCollectTaskStopsOnDifferentTarget(t *testing.T) {
	sender := &Sender{
		ch: make(chan *task, 2),
	}

	current := &task{
		targetURL: "http://127.0.0.1:9529/v1/write/network?input=bench",
		data:      makeTestPoints(10),
	}
	next := &task{
		targetURL: "http://127.0.0.1:9529/v1/write/logging?input=bench",
		data:      makeTestPoints(12),
	}
	sender.ch <- next

	merged, pending := sender.collectTask(current, nil)
	require.NotNil(t, merged)
	require.NotNil(t, pending)
	assert.Len(t, merged.data, 10)
	assert.Same(t, next, pending)
}

func BenchmarkSenderRequest(b *testing.B) {
	sender := &Sender{
		httpCli: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				_, _ = io.Copy(io.Discard, req.Body)
				_ = req.Body.Close()

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
				}, nil
			}),
		},
	}

	task := &task{
		targetURL: "http://127.0.0.1:9529/v1/write/network?input=bench",
		data:      makeTestPoints(maxPtSendCount),
	}
	m := &marshaler{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := sender.request(context.Background(), m, cloneTask(task)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSenderCollectTask(b *testing.B) {
	current := &task{
		targetURL: "http://127.0.0.1:9529/v1/write/network?input=bench",
		data:      makeTestPoints(64),
	}
	next1 := &task{
		targetURL: current.targetURL,
		data:      makeTestPoints(64),
	}
	next2 := &task{
		targetURL: current.targetURL,
		data:      makeTestPoints(64),
	}
	next3 := &task{
		targetURL: current.targetURL,
		data:      makeTestPoints(64),
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sender := &Sender{
			ch: make(chan *task, 3),
		}
		sender.ch <- cloneTask(next1)
		sender.ch <- cloneTask(next2)
		sender.ch <- cloneTask(next3)

		merged, pending := sender.collectTask(cloneTask(current), nil)
		if merged == nil || pending != nil || len(merged.data) != maxPtSendCount {
			b.Fatal("unexpected merge result")
		}
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func makeTestPoints(n int) []*point.Point {
	pts := make([]*point.Point, 0, n)
	for i := 1; i <= n; i++ {
		pts = append(pts, point.NewPoint("a", point.NewKVs(map[string]any{
			"id":      i,
			"service": "bench",
		})))
	}
	return pts
}

func cloneTask(src *task) *task {
	if src == nil {
		return nil
	}

	dup := &task{
		targetURL: src.targetURL,
		data:      make([]*point.Point, len(src.data)),
	}
	copy(dup.data, src.data)
	return dup
}
