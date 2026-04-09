// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package aggr

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

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
	for _, part := range parts {
		totalPoints += int(part.PointCount)
		assert.LessOrEqual(t, part.Size(), maxRawBodySize)
		assert.Equal(t, int(part.PointCount), countPBPointsPayload(t, part.PointsPayload))
	}
	assert.Equal(t, int(pkg.PointCount), totalPoints)
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

	err := ag.sendTailSamplingPackage(11, buildTailSamplingDataPacket(4, 32))
	require.NoError(t, err)

	assert.Equal(t, int64(bodySize), gotContentLength)
	assert.Equal(t, aggregatePayloadContentType, gotContentType)
	assert.Equal(t, identityContentEncoding, gotEncoding)
	assert.Equal(t, "11", gotPickKey)
	assert.Equal(t, strconv.Itoa(bodySize), gotPayloadSize)
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

	err := ag.sendMetricBatch(7, batch)
	require.NoError(t, err)

	assert.Equal(t, int64(bodySize), gotContentLength)
	assert.Equal(t, aggregatePayloadContentType, gotContentType)
	assert.Equal(t, identityContentEncoding, gotEncoding)
	assert.Equal(t, "7", gotPickKey)
	assert.Equal(t, strconv.Itoa(bodySize), gotPayloadSize)
	assert.Equal(t, "7", gotRoutingKey)
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
