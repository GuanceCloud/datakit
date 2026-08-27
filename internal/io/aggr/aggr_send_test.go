// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package aggr

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GuanceCloud/cliutils/aggregate"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

func TestSendWorkerCount(t *testing.T) {
	origCPUs := datakit.AvailableCPUs
	t.Cleanup(func() { datakit.AvailableCPUs = origCPUs })

	ag := &Aggregator{}
	cases := []struct {
		name      string
		cpus      int
		taskCount int
		want      int
	}{
		{name: "cpu_zero", cpus: 0, taskCount: 5, want: 1},
		{name: "task_less_than_worker", cpus: 4, taskCount: 3, want: 3},
		{name: "cap_by_max_async_worker", cpus: 10, taskCount: 100, want: maxAsyncSendWorkers},
		{name: "task_zero", cpus: 8, taskCount: 0, want: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			datakit.AvailableCPUs = tc.cpus
			assert.Equal(t, tc.want, ag.sendWorkerCount(tc.taskCount))
		})
	}
}

func TestMaxRawBodySize(t *testing.T) {
	ag := &Aggregator{
		MaxRawBodySize: 4096,
		DW: &dataway.Dataway{
			MaxRawBodySize: 2048,
		},
	}
	assert.Equal(t, 4096, ag.maxRawBodySize())

	ag.MaxRawBodySize = 0
	assert.Equal(t, 2048, ag.maxRawBodySize())

	ag.DW = nil
	assert.Equal(t, dataway.DefaultMaxRawBodySize, ag.maxRawBodySize())
}

func TestRunAsyncSend(t *testing.T) {
	ag := &Aggregator{}
	var called int32
	expectErr := errors.New("send failed")

	err := ag.runAsyncSend(10, func(i int) error {
		atomic.AddInt32(&called, 1)
		if i == 5 {
			return expectErr
		}
		return nil
	})
	require.ErrorIs(t, err, expectErr)
	assert.Equal(t, int32(10), called)

	err = ag.runAsyncSend(0, func(int) error { return expectErr })
	require.NoError(t, err)
}

func TestSplitDataPacketBySize(t *testing.T) {
	pkg := buildTailSamplingDataPacket(10, 64)
	maxRawBodySize := pkg.Size() / 2
	require.Greater(t, maxRawBodySize, 0)

	parts := splitDataPacketBySize(pkg, maxRawBodySize)
	require.Greater(t, len(parts), 1)

	totalPoints := 0
	var reassembledPayload []byte
	for _, part := range parts {
		totalPoints += int(part.PointCount)
		reassembledPayload = append(reassembledPayload, part.PointsPayload...)
		assert.LessOrEqual(t, part.Size(), maxRawBodySize)
		assert.Equal(t, int(part.PointCount), countPBPointsPayload(t, part.PointsPayload))
	}
	assert.Equal(t, int(pkg.PointCount), totalPoints)
	assert.Equal(t, pkg.PointsPayload, reassembledPayload)
}

func TestSplitCompressedDataPacketPreservesCompressionAndSummaryVersion(t *testing.T) {
	pkg := buildTailSamplingDataPacket(20, 512)
	rawPayload := append([]byte(nil), pkg.PointsPayload...)
	compressed, compression, err := aggregate.CompressPointsPayload(rawPayload)
	require.NoError(t, err)
	require.Equal(t, aggregate.PayloadCompressionZstd, compression)
	pkg.PointsPayload = compressed
	pkg.PayloadCompression = compression
	pkg.PredicateSummaryVersion = aggregate.CurrentPredicateSummaryVersion

	parts := splitDataPacketBySize(pkg, pkg.Size()-1)
	require.Greater(t, len(parts), 1)

	var reassembled []byte
	var totalPoints int32
	for _, part := range parts {
		assert.Equal(t, aggregate.CurrentPredicateSummaryVersion, part.PredicateSummaryVersion)
		assert.LessOrEqual(t, part.Size(), pkg.Size()-1)
		decoded, decodeErr := aggregate.DecompressPointsPayload(part.PointsPayload, part.PayloadCompression)
		require.NoError(t, decodeErr)
		assert.Equal(t, int(part.PointCount), countPBPointsPayload(t, decoded))
		reassembled = append(reassembled, decoded...)
		totalPoints += part.PointCount
	}
	assert.Equal(t, pkg.PointCount, totalPoints)
	assert.Equal(t, rawPayload, reassembled)
}

func TestSplitDataPacketBySizeSinglePointOversize(t *testing.T) {
	pkg := buildTailSamplingDataPacket(2, 2048)

	base := cloneDataPacketMeta(pkg)
	base.PointsPayload = nil
	base.PointCount = 0
	base.MaxPointTimeUnixNano = 0
	singlePointSize := protoListElemSize(firstPBPointSize(t, pkg.PointsPayload))
	maxRawBodySize := base.Size() + singlePointSize - 1

	parts := splitDataPacketBySize(pkg, maxRawBodySize)
	require.Len(t, parts, 2)
	assert.Equal(t, int32(1), parts[0].PointCount)
	assert.Equal(t, int32(1), parts[1].PointCount)
	assert.Equal(t, 1, countPBPointsPayload(t, parts[0].PointsPayload))
	assert.Equal(t, 1, countPBPointsPayload(t, parts[1].PointsPayload))
}

func TestSplitDataPacketBySizeHonorsExactPacketSize(t *testing.T) {
	pkg := buildTailSamplingDataPacket(3, 64)
	maxRawBodySize := buildTailSamplingDataPacket(2, 64).Size() - 1
	require.LessOrEqual(t, buildTailSamplingDataPacket(1, 64).Size(), maxRawBodySize)

	parts := splitDataPacketBySize(pkg, maxRawBodySize)
	require.Greater(t, len(parts), 1)
	for _, part := range parts {
		assert.LessOrEqual(t, part.Size(), maxRawBodySize)
	}
}

func TestSplitDataPacketBySizeFallsBackForInvalidPayload(t *testing.T) {
	pkg := buildTailSamplingDataPacket(2, 64)
	pkg.PointsPayload = []byte{pbPointsArrayTag, 2, 0xff}
	base := cloneDataPacketMeta(pkg)
	base.PointsPayload = nil
	base.PointCount = 0
	base.MaxPointTimeUnixNano = 0

	parts := splitDataPacketBySize(pkg, base.Size()+1)
	require.Len(t, parts, 1)
	assert.Same(t, pkg, parts[0])
}

func TestWalkDataPacketPartsBySizeStopsWhenContextAlreadyCanceled(t *testing.T) {
	pkg := buildTailSamplingDataPacket(48, 16<<10)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	emitted := 0
	parts, err := walkDataPacketPartsBySize(ctx, pkg, 32<<10, func(*aggregate.DataPacket) error {
		emitted++
		return nil
	})
	require.ErrorIs(t, err, context.Canceled)
	assert.Zero(t, parts)
	assert.Zero(t, emitted)
}

func TestSplitBatchsBySize(t *testing.T) {
	batch := buildMetricBatchs(20, 64)
	maxRawBodySize := batch.Size() / 2
	require.Greater(t, maxRawBodySize, 0)

	parts := splitBatchsBySize(batch, maxRawBodySize)
	require.Greater(t, len(parts), 1)

	totalBatchCount := 0
	for _, part := range parts {
		totalBatchCount += len(part.Batchs)
		assert.LessOrEqual(t, part.Size(), maxRawBodySize)
	}
	assert.Equal(t, len(batch.Batchs), totalBatchCount)
}

func TestMarshalHelpers(t *testing.T) {
	_, err := marshalDataPacketWithPool(nil)
	require.Error(t, err)

	_, err = marshalBatchsWithPool(nil)
	require.Error(t, err)

	pkg := buildTailSamplingDataPacket(3, 32)
	body, err := marshalDataPacketWithPool(pkg)
	require.NoError(t, err)
	require.NotEmpty(t, body.buf)
	defer putMarshalBody(body)

	gotPkgV2 := &aggregate.DataPacket{}
	require.NoError(t, gotPkgV2.Unmarshal(body.buf))
	assert.Equal(t, pkg.RawGroupId, gotPkgV2.RawGroupId)
	assert.Equal(t, pkg.PointCount, gotPkgV2.PointCount)
	assert.Equal(t, int(pkg.PointCount), countPBPointsPayload(t, gotPkgV2.PointsPayload))

	batch := buildMetricBatchs(3, 32)
	batchBody, err := marshalBatchsWithPool(batch)
	require.NoError(t, err)
	require.NotEmpty(t, batchBody.buf)
	defer putMarshalBody(batchBody)

	gotBatch := &aggregate.Batchs{}
	require.NoError(t, gotBatch.Unmarshal(batchBody.buf))
	assert.Equal(t, batch.PickKey, gotBatch.PickKey)
	assert.Len(t, gotBatch.Batchs, len(batch.Batchs))
}

func TestSendTailSamplingPackageHeaders(t *testing.T) {
	var (
		gotContentLength int64
		gotContentType   string
		gotEncoding      string
		gotPickKey       string
		gotPayloadSize   string
		bodySize         int
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != datakit.TailSampling {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		bs, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		bodySize = len(bs)
		gotContentLength = r.ContentLength
		gotContentType = r.Header.Get("Content-Type")
		gotEncoding = r.Header.Get("Content-Encoding")
		gotPickKey = r.Header.Get(aggregate.GuancePickKey)
		gotPayloadSize = r.Header.Get(payloadSizeHeader)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ag := &Aggregator{
		Endpoints: []string{ts.URL + "?token=tkn_trace"},
	}
	ag.initHTTP()

	err := ag.sendTailSamplingPackageContext(context.Background(), 11, buildTailSamplingDataPacket(4, 32))
	require.NoError(t, err)

	assert.Equal(t, int64(bodySize), gotContentLength)
	assert.Equal(t, aggregatePayloadContentType, gotContentType)
	assert.Equal(t, identityContentEncoding, gotEncoding)
	assert.Equal(t, "11", gotPickKey)
	assert.Equal(t, strconv.Itoa(bodySize), gotPayloadSize)
}

func TestSendTailSamplingPackageFallsBackFromZstdV2(t *testing.T) {
	type receivedRequest struct {
		path        string
		token       string
		compression string
		packet      *aggregate.DataPacket
	}

	received := make(chan receivedRequest, 3)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		packet := &aggregate.DataPacket{}
		require.NoError(t, packet.Unmarshal(body))
		received <- receivedRequest{
			path:        r.URL.Path,
			token:       r.URL.Query().Get("token"),
			compression: r.Header.Get(aggregate.TailSamplingPayloadCompressionHeader),
			packet:      packet,
		}

		if r.URL.Path == datakit.TailSamplingV2 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ag := &Aggregator{Endpoints: []string{ts.URL + "?token=tkn_trace"}}
	ag.initHTTP()

	packet := buildTailSamplingDataPacket(8, 256)
	require.NoError(t, aggregate.SetDataPacketPayloadCompression(packet, aggregate.PayloadCompressionZstd))
	require.Equal(t, aggregate.PayloadCompressionZstd, packet.PayloadCompression)

	require.NoError(t, ag.sendTailSamplingPackageContext(context.Background(), 11, packet))
	require.NoError(t, ag.sendTailSamplingPackageContext(context.Background(), 11, packet))

	first := <-received
	assert.Equal(t, datakit.TailSamplingV2, first.path)
	assert.Equal(t, "tkn_trace", first.token)
	assert.Equal(t, aggregate.TailSamplingPayloadCompressionZstd, first.compression)
	assert.Equal(t, aggregate.PayloadCompressionZstd, first.packet.PayloadCompression)

	for range 2 {
		fallback := <-received
		assert.Equal(t, datakit.TailSampling, fallback.path)
		assert.Equal(t, "tkn_trace", fallback.token)
		assert.Empty(t, fallback.compression)
		assert.Equal(t, aggregate.PayloadCompressionNone, fallback.packet.PayloadCompression)
		assert.Equal(t, int(packet.PointCount), countPBPointsPayload(t, fallback.packet.PointsPayload))
	}
	assert.Len(t, received, 0, "unsupported endpoint capability should be cached")
}

func TestSendTailSamplingPackageFallsBackOnUnsupportedMediaType(t *testing.T) {
	receivedPaths := make(chan string, 2)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPaths <- r.URL.Path
		if r.URL.Path == datakit.TailSamplingV2 {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ag := &Aggregator{Endpoints: []string{ts.URL + "?token=tkn_trace"}}
	ag.initHTTP()

	packet := buildTailSamplingDataPacket(8, 256)
	require.NoError(t, aggregate.SetDataPacketPayloadCompression(packet, aggregate.PayloadCompressionZstd))
	require.NoError(t, ag.sendTailSamplingPackageContext(context.Background(), 11, packet))

	assert.Equal(t, datakit.TailSamplingV2, <-receivedPaths)
	assert.Equal(t, datakit.TailSampling, <-receivedPaths)
}

func TestSendTailSamplingPackagesSplitsDecodedPayloadBeforeLegacyFallback(t *testing.T) {
	const maxRawBodySize = 32 << 10

	var (
		legacyRequests int32
		receivedPoints int32
		oversized      int32
	)
	ag := newTailSamplingTestAggregator(t, maxRawBodySize, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		if r.URL.Path == datakit.TailSamplingV2 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		require.Equal(t, datakit.TailSampling, r.URL.Path)
		atomic.AddInt32(&legacyRequests, 1)
		if len(body) > maxRawBodySize {
			atomic.StoreInt32(&oversized, 1)
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}

		packet := &aggregate.DataPacket{}
		require.NoError(t, packet.Unmarshal(body))
		atomic.AddInt32(&receivedPoints, int32(countPBPointsPayload(t, packet.PointsPayload)))
		w.WriteHeader(http.StatusOK)
	}))

	packet := buildTailSamplingDataPacket(48, 16<<10)
	rawSize := packet.Size()
	require.NoError(t, aggregate.SetDataPacketPayloadCompression(packet, aggregate.PayloadCompressionZstd))
	require.Less(t, packet.Size(), maxRawBodySize, "fixture must fit on the compressed V2 wire path")
	require.Greater(t, rawSize, maxRawBodySize, "fixture must require splitting on the legacy raw path")

	err := ag.SendTailSamplingPackages(map[uint64]*aggregate.DataPacket{1: packet})
	require.NoError(t, err)
	assert.Zero(t, atomic.LoadInt32(&oversized))
	assert.Greater(t, atomic.LoadInt32(&legacyRequests), int32(1))
	assert.Equal(t, packet.PointCount, atomic.LoadInt32(&receivedPoints))
}

func TestSendMetricBatchHeaders(t *testing.T) {
	var (
		gotContentLength int64
		gotContentType   string
		gotEncoding      string
		gotPickKey       string
		gotPayloadSize   string
		gotRoutingKey    string
		bodySize         int
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != datakit.Aggregate {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		bs, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		bodySize = len(bs)
		gotContentLength = r.ContentLength
		gotContentType = r.Header.Get("Content-Type")
		gotEncoding = r.Header.Get("Content-Encoding")
		gotPickKey = r.Header.Get(aggregate.GuancePickKey)
		gotPayloadSize = r.Header.Get(payloadSizeHeader)
		gotRoutingKey = r.Header.Get(aggregate.GuanceRoutingKey)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ag := &Aggregator{
		Endpoints: []string{ts.URL + "?token=tkn_metric"},
	}
	ag.initHTTP()

	batch := buildMetricBatchs(4, 32)
	batch.PickKey = 7
	for _, item := range batch.Batchs {
		item.PickKey = 7
	}

	err := ag.sendMetricBatch(point.SMetric, 7, batch)
	require.NoError(t, err)

	assert.Equal(t, int64(bodySize), gotContentLength)
	assert.Equal(t, aggregatePayloadContentType, gotContentType)
	assert.Equal(t, identityContentEncoding, gotEncoding)
	assert.Equal(t, "7", gotPickKey)
	assert.Equal(t, strconv.Itoa(bodySize), gotPayloadSize)
	assert.Equal(t, "7", gotRoutingKey)
}

func TestSendAggrRequestsDoNotInheritDatawaySinkHeaders(t *testing.T) {
	paths := map[string]int32{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case datakit.Aggregate, datakit.TailSampling, datakit.TailSamplingConfig:
			paths[r.URL.Path]++
			assert.Emptyf(t, r.Header.Get(dataway.HeaderXGlobalTags), "failed on request %s", r.URL.Path)
			assert.Emptyf(t, r.Header.Get(dataway.HeaderXGlobalTagsV2), "failed on request %s", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	dw := dataway.NewDefaultDataway(dataway.WithGlobalTags(map[string]string{
		"host": "test-host",
	}))
	dw.EnableSinker = true
	require.NoError(t, dw.Init(dataway.WithURLs(ts.URL+"?token=tkn_sink")))

	ag := &Aggregator{
		DW:                  dw,
		tailSamplingConfig:  &aggregate.TailSamplingConfigs{Version: 1, Tracing: &aggregate.TraceTailSampling{}},
		tailSamplingEnabled: true,
	}
	ag.initHTTP()

	require.NoError(t, ag.sendMetricBatch(point.SMetric, 1, buildMetricBatchs(1, 32)))
	require.NoError(t, ag.sendTailSamplingPackageContext(context.Background(), 1, buildTailSamplingDataPacket(2, 32)))
	ag.sendTSConfigToDW()

	assert.Equal(t, int32(1), paths[datakit.Aggregate])
	assert.Equal(t, int32(1), paths[datakit.TailSampling])
	assert.Equal(t, int32(1), paths[datakit.TailSamplingConfig])
}

func TestSendMetricBatchesRunConcurrently(t *testing.T) {
	setAvailableCPUsForTest(t, 1)

	var (
		inflight    int32
		maxInflight int32
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != datakit.Aggregate {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		cur := atomic.AddInt32(&inflight, 1)
		for {
			prev := atomic.LoadInt32(&maxInflight)
			if cur <= prev || atomic.CompareAndSwapInt32(&maxInflight, prev, cur) {
				break
			}
		}

		time.Sleep(100 * time.Millisecond)
		atomic.AddInt32(&inflight, -1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ag := &Aggregator{
		Endpoints: []string{ts.URL + "?token=tkn_metric"},
	}
	ag.initHTTP()

	first := buildMetricBatchs(1, 32)
	second := buildMetricBatchs(1, 32)

	require.NoError(t, ag.SendMetricBatches(point.SMetric, map[uint64]*aggregate.Batchs{
		1: first,
		2: second,
	}))
	assert.Greater(t, atomic.LoadInt32(&maxInflight), int32(1))
}

func TestSendTailSamplingPackagesRunConcurrently(t *testing.T) {
	setAvailableCPUsForTest(t, 1)

	var (
		inflight    int32
		maxInflight int32
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != datakit.TailSampling {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		cur := atomic.AddInt32(&inflight, 1)
		for {
			prev := atomic.LoadInt32(&maxInflight)
			if cur <= prev || atomic.CompareAndSwapInt32(&maxInflight, prev, cur) {
				break
			}
		}

		time.Sleep(100 * time.Millisecond)
		atomic.AddInt32(&inflight, -1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ag := &Aggregator{
		Endpoints: []string{ts.URL + "?token=tkn_trace"},
	}
	ag.initHTTP()

	require.NoError(t, ag.SendTailSamplingPackages(map[uint64]*aggregate.DataPacket{
		1: buildTailSamplingDataPacket(2, 32),
		2: buildTailSamplingDataPacket(2, 32),
	}))
	assert.Greater(t, atomic.LoadInt32(&maxInflight), int32(1))
}

func TestSendTailSamplingPackagesKeepsSplitMemoryBounded(t *testing.T) {
	setAvailableCPUsForTest(t, 0)

	var requests int32
	ag := newTailSamplingTestAggregator(t, 32<<10, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		_, err := io.Copy(io.Discard, r.Body)
		if !assert.NoError(t, err) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	pkg := buildTailSamplingDataPacket(48, 16<<10)

	runtime.GC()
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	require.NoError(t, ag.SendTailSamplingPackages(map[uint64]*aggregate.DataPacket{1: pkg}))

	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	allocated := after.TotalAlloc - before.TotalAlloc
	t.Logf("payload=%d requests=%d allocated=%d", len(pkg.PointsPayload), requests, allocated)

	require.Greater(t, atomic.LoadInt32(&requests), int32(1), "fixture did not trigger splitting")
	assert.Less(t, allocated, uint64(32<<20), "split/send allocated more than 32 MiB")
}

func TestSendTailSamplingPackagesContinuesAfterFailure(t *testing.T) {
	setAvailableCPUsForTest(t, 0)

	var (
		requests       int32
		receivedPoints int32
		failedPoints   int32
		decodeErrors   int32
	)
	ag := newTailSamplingTestAggregator(t, 32<<10, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			atomic.AddInt32(&decodeErrors, 1)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		packet := &aggregate.DataPacket{}
		if err := packet.Unmarshal(body); err != nil {
			atomic.AddInt32(&decodeErrors, 1)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		atomic.AddInt32(&receivedPoints, packet.PointCount)
		if atomic.AddInt32(&requests, 1) == 1 {
			atomic.AddInt32(&failedPoints, packet.PointCount)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	pkg := buildTailSamplingDataPacket(48, 16<<10)
	lostBefore := counterValue(t, aggrLostPoints.WithLabelValues("tail_sampling", pkg.DataType, "server"))

	err := ag.SendTailSamplingPackages(map[uint64]*aggregate.DataPacket{
		1: pkg,
	})
	require.Error(t, err)
	assert.Zero(t, atomic.LoadInt32(&decodeErrors))
	assert.Greater(t, atomic.LoadInt32(&requests), int32(1))
	assert.Equal(t, pkg.PointCount, atomic.LoadInt32(&receivedPoints))
	assert.Equal(t, lostBefore+float64(atomic.LoadInt32(&failedPoints)),
		counterValue(t, aggrLostPoints.WithLabelValues("tail_sampling", pkg.DataType, "server")))
}

func TestSendTailSamplingPackagesStopsOnContextCancellation(t *testing.T) {
	setAvailableCPUsForTest(t, 0)

	started := make(chan struct{}, 1)
	release := make(chan struct{})
	ag := newTailSamplingTestAggregator(t, 32<<10, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		started <- struct{}{}
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	t.Cleanup(func() { close(release) })

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	result := make(chan error, 1)
	go func() {
		result <- ag.SendTailSamplingPackagesContext(ctx, map[uint64]*aggregate.DataPacket{
			1: buildTailSamplingDataPacket(48, 16<<10),
		})
	}()

	select {
	case <-started:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for the first request")
	}
	select {
	case err := <-result:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("send did not stop after context cancellation")
	}
}

func TestSendTailSamplingPackagesContextCancellationOverridesSendFailure(t *testing.T) {
	setAvailableCPUsForTest(t, 0)

	var requests int32
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	ag := newTailSamplingTestAggregator(t, 32<<10, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&requests, 1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		started <- struct{}{}
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	t.Cleanup(func() { close(release) })

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	result := make(chan error, 1)
	go func() {
		result <- ag.SendTailSamplingPackagesContext(ctx, map[uint64]*aggregate.DataPacket{
			1: buildTailSamplingDataPacket(48, 16<<10),
		})
	}()

	select {
	case <-started:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for the second request")
	}
	select {
	case err := <-result:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("send did not stop after context cancellation")
	}
}

func TestSendMetricBatchSelectsConfiguredEndpointByPickKey(t *testing.T) {
	var (
		firstReqs  int32
		secondReqs int32
	)

	first := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == datakit.Aggregate {
			atomic.AddInt32(&firstReqs, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer first.Close()

	second := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == datakit.Aggregate {
			atomic.AddInt32(&secondReqs, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer second.Close()

	ag := &Aggregator{
		Endpoints: []string{
			first.URL + "?token=tkn_metric_1",
			second.URL + "?token=tkn_metric_2",
		},
	}
	ag.initHTTP()

	batch := buildMetricBatchs(1, 32)
	require.NoError(t, ag.sendMetricBatch(point.SMetric, 1, batch))

	assert.Equal(t, int32(0), atomic.LoadInt32(&firstReqs))
	assert.Equal(t, int32(1), atomic.LoadInt32(&secondReqs))
}

func TestSendMetricBatchRecordsCompatibilityMetrics(t *testing.T) {
	category := "compat_metric_success"
	batch := buildMetricBatchs(2, 32)
	points := countSelectedMetricPointsInBatch(batch)

	successBefore := counterValue(t, aggrSendSuccess.WithLabelValues("metric", category))
	pointsBefore := counterValue(t, aggrSendPoints.WithLabelValues("metric", category))
	latencyBefore := summaryCount(t, aggrSendLatency.WithLabelValues("metric", category))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != datakit.Aggregate {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ag := &Aggregator{
		Endpoints: []string{ts.URL + "?token=tkn_metric"},
	}
	ag.initHTTP()

	require.NoError(t, ag.sendMetricBatch(category, 1, batch))

	assert.Equal(t, successBefore+1, counterValue(t, aggrSendSuccess.WithLabelValues("metric", category)))
	assert.Equal(t, pointsBefore+float64(points), counterValue(t, aggrSendPoints.WithLabelValues("metric", category)))
	assert.Equal(t, latencyBefore+1, summaryCount(t, aggrSendLatency.WithLabelValues("metric", category)))
}

func TestSendMetricBatchRecordsCompatibilityFailureMetrics(t *testing.T) {
	category := "compat_metric_failure"
	batch := buildMetricBatchs(2, 32)
	points := countSelectedMetricPointsInBatch(batch)

	failedBefore := counterValue(t, aggrSendFailed.WithLabelValues("metric", category, "other"))
	lostBefore := counterValue(t, aggrLostPoints.WithLabelValues("metric", category, "other"))
	latencyBefore := summaryCount(t, aggrSendLatency.WithLabelValues("metric", category))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != datakit.Aggregate {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	ag := &Aggregator{
		Endpoints: []string{ts.URL + "?token=tkn_metric"},
	}
	ag.initHTTP()

	require.Error(t, ag.sendMetricBatch(category, 1, batch))

	assert.Equal(t, failedBefore+1, counterValue(t, aggrSendFailed.WithLabelValues("metric", category, "other")))
	assert.Equal(t, lostBefore+float64(points), counterValue(t, aggrLostPoints.WithLabelValues("metric", category, "other")))
	assert.Equal(t, latencyBefore+1, summaryCount(t, aggrSendLatency.WithLabelValues("metric", category)))
}

func TestSendMetricBatchFallsBackToDatawayEndpoints(t *testing.T) {
	var (
		firstReqs  int32
		secondReqs int32
	)

	first := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == datakit.Aggregate {
			atomic.AddInt32(&firstReqs, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer first.Close()

	second := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == datakit.Aggregate {
			atomic.AddInt32(&secondReqs, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer second.Close()

	dw := dataway.NewDefaultDataway()
	require.NoError(t, dw.Init(dataway.WithURLs(
		first.URL+"?token=tkn_metric_1",
		second.URL+"?token=tkn_metric_2",
	)))

	ag := &Aggregator{DW: dw}
	ag.initHTTP()

	batch := buildMetricBatchs(1, 32)
	require.NoError(t, ag.sendMetricBatch(point.SMetric, 1, batch))

	assert.Equal(t, int32(1), atomic.LoadInt32(&firstReqs))
	assert.Equal(t, int32(1), atomic.LoadInt32(&secondReqs))
}

func TestSendTailSamplingPackageFallsBackToAllDatawayEndpoints(t *testing.T) {
	var (
		firstReqs  int32
		secondReqs int32
	)

	first := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == datakit.TailSampling {
			atomic.AddInt32(&firstReqs, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer first.Close()

	second := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == datakit.TailSampling {
			atomic.AddInt32(&secondReqs, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer second.Close()

	dw := dataway.NewDefaultDataway()
	require.NoError(t, dw.Init(dataway.WithURLs(
		first.URL+"?token=tkn_trace_1",
		second.URL+"?token=tkn_trace_2",
	)))

	ag := &Aggregator{DW: dw}
	ag.initHTTP()

	require.NoError(t, ag.sendTailSamplingPackageContext(context.Background(), 1, buildTailSamplingDataPacket(4, 32)))

	assert.Equal(t, int32(1), atomic.LoadInt32(&firstReqs))
	assert.Equal(t, int32(1), atomic.LoadInt32(&secondReqs))
}

func TestSendTailSamplingConfigHeaders(t *testing.T) {
	var gotPickKey string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != datakit.TailSamplingConfig {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		gotPickKey = r.Header.Get(aggregate.GuancePickKey)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ag := &Aggregator{
		Endpoints:           []string{ts.URL + "?token=tkn_trace"},
		tailSamplingConfig:  &aggregate.TailSamplingConfigs{Version: 1, Tracing: &aggregate.TraceTailSampling{}},
		tailSamplingEnabled: true,
	}
	ag.initHTTP()

	ag.sendTSConfigToDW()

	assert.Equal(t, "1", gotPickKey)
}

func buildTailSamplingDataPacket(pointCount int, payloadSize int) *aggregate.DataPacket {
	payload := strings.Repeat("x", payloadSize)

	var pointsPayload []byte
	var maxPointTimeUnixNano int64
	for i := 0; i < pointCount; i++ {
		pt := point.NewPoint("ddtrace", point.NewKVs(map[string]interface{}{
			"trace_id": "trace-001",
			"span_id":  "span-" + strconv.Itoa(i),
			"message":  payload,
			"duration": int64(1000 + i),
		}), point.CommonLoggingOptions()...)
		pointsPayload = point.AppendPointToPBPointsPayload(pointsPayload, pt)
		if ts := pt.Time().UnixNano(); ts > maxPointTimeUnixNano {
			maxPointTimeUnixNano = ts
		}
	}

	return &aggregate.DataPacket{
		GroupIdHash:            1,
		RawGroupId:             "trace-001",
		Token:                  "tok",
		Source:                 "ddtrace",
		DataType:               point.STracing,
		ConfigVersion:          1,
		PointCount:             int32(pointCount),
		PointsPayload:          pointsPayload,
		MaxPointTimeUnixNano:   maxPointTimeUnixNano,
		TraceStartTimeUnixNano: 1,
		TraceEndTimeUnixNano:   2,
	}
}

func countPBPointsPayload(t *testing.T, payload []byte) int {
	t.Helper()

	cnt := 0
	require.NoError(t, point.WalkPBPointsPayload(payload, func([]byte) bool {
		cnt++
		return true
	}))

	return cnt
}

func firstPBPointSize(t *testing.T, payload []byte) int {
	t.Helper()

	size := 0
	require.NoError(t, point.WalkPBPointsPayload(payload, func(raw []byte) bool {
		size = len(raw)
		return false
	}))
	require.Greater(t, size, 0)
	return size
}

func buildMetricBatchs(batchCount int, payloadSize int) *aggregate.Batchs {
	payload := strings.Repeat("y", payloadSize)

	arr := make([]*aggregate.AggregationBatch, 0, batchCount)
	for i := 0; i < batchCount; i++ {
		pt := point.NewPoint("opentelemetry", point.KVs{}.
			Add("jvm.buffer.memory.used", float64(100+i)).
			AddTag("service_name", "svc-"+payload).
			AddTag("id", "id-1"),
			point.DefaultMetricOptions()...)

		arr = append(arr, &aggregate.AggregationBatch{
			RoutingKey: uint64(i + 1),
			PickKey:    1,
			Points: &point.PBPoints{
				Arr: []*point.PBPoint{pt.PBPoint()},
			},
		})
	}

	return &aggregate.Batchs{
		PickKey: 1,
		Batchs:  arr,
	}
}

func BenchmarkWalkDataPacketPartsBySize64MiB(b *testing.B) {
	pkg := buildTailSamplingDataPacket(256, 256<<10)
	const maxRawBodySize = 1 << 20

	b.ReportAllocs()
	b.SetBytes(int64(len(pkg.PointsPayload)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parts, err := walkDataPacketPartsBySize(context.Background(), pkg, maxRawBodySize, func(*aggregate.DataPacket) error {
			return nil
		})
		if err != nil {
			b.Fatal(err)
		}
		if parts < 2 {
			b.Fatalf("fixture did not trigger splitting: parts=%d", parts)
		}
	}
}

func newTailSamplingTestAggregator(t *testing.T, maxRawBodySize int, handler http.Handler) *Aggregator {
	t.Helper()

	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	ag := &Aggregator{
		Endpoints:      []string{ts.URL + "?token=tkn_trace"},
		MaxRawBodySize: maxRawBodySize,
	}
	ag.initHTTP()
	return ag
}

func setAvailableCPUsForTest(t *testing.T, cpus int) {
	t.Helper()
	origCPUs := datakit.AvailableCPUs
	t.Cleanup(func() { datakit.AvailableCPUs = origCPUs })
	datakit.AvailableCPUs = cpus
}

func counterValue(t *testing.T, collector prometheus.Counter) float64 {
	t.Helper()

	m := &dto.Metric{}
	require.NoError(t, collector.Write(m))
	return m.GetCounter().GetValue()
}

func summaryCount(t *testing.T, observer prometheus.Observer) uint64 {
	t.Helper()

	metric, ok := observer.(prometheus.Metric)
	require.True(t, ok)

	m := &dto.Metric{}
	require.NoError(t, metric.Write(m))
	return m.GetSummary().GetSampleCount()
}
