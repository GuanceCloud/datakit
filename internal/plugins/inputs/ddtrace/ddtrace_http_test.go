// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows

package ddtrace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"

	"github.com/ugorji/go/codec"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

var (
	msgpHandler codec.MsgpackHandle
	encoder     = codec.NewEncoder(nil, &msgpHandler)
	decoder     = codec.NewDecoder(nil, &msgpHandler)
)

func msgpackEncoder(ddtraces DDTraces) ([]byte, error) {
	return Marshal(ddtraces)
}

func Marshal(src interface{}) ([]byte, error) {
	buf := bufpool.GetBuffer()
	encoder.Reset(buf)
	err := encoder.Encode(src)
	b := buf.Bytes()
	bufpool.PutBuffer(buf)

	return b, err
}

type ddHandler struct{}

func (d *ddHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handleDDTraces(w, r)
}

func Test_handleDDTraces(t *testing.T) {
	mockFeed := dkio.NewMockedFeeder()
	afterGatherRun = itrace.NewAfterGather(itrace.WithLogger(log), itrace.WithPointOptions(), itrace.WithFeeder(mockFeed))
	ts := httptest.NewServer(&ddHandler{})
	buf, err := jsonEncoder(randomDDTraces(10, 10))
	if err != nil {
		t.Error(err.Error())

		return
	}

	req, err := http.NewRequest("post", ts.URL, bytes.NewBuffer(buf))
	if err != nil {
		t.Error(err.Error())

		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err.Error())

		return
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status code =%d", resp.StatusCode)
		return
	}
	pts, err := mockFeed.AnyPoints(time.Second)
	if err == nil {
		t.Logf("point len= %d", len(pts))
	}
}

func BenchmarkDDTrace_Msgsize(b *testing.B) {
	mockFeed := dkio.NewMockedFeeder()
	afterGatherRun = itrace.NewAfterGather(itrace.WithLogger(log), itrace.WithPointOptions(), itrace.WithFeeder(mockFeed))
	ts := httptest.NewServer(&ddHandler{})
	buf, err := msgpackEncoder(randomDDTraces(10, 10))
	if err != nil {
		b.Error(err.Error())

		return
	}
	for i := 0; i < b.N; i++ {
		req, err := http.NewRequest("post", ts.URL, bytes.NewBuffer(buf))
		if err != nil {
			b.Error(err.Error())

			return
		}
		// req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Type", "application/msgpack")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Error(err.Error())

			return
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("request failed with status code %d\n", resp.StatusCode)
		}
	}

	pts, err := mockFeed.AnyPoints(time.Second)
	if err == nil {
		b.Logf("point len= %d", len(pts))
	}
}

func TestDDTrace(t *testing.T) {
	mockFeed := dkio.NewMockedFeeder()
	gather := itrace.NewAfterGather(itrace.WithLogger(log), itrace.WithPointOptions(), itrace.WithFeeder(mockFeed))
	sample := &itrace.Sampler{
		SamplingRateGlobal: 1,
	}
	sample.Init()
	gather.AppendFilter(sample.Sample)
	gather.AppendFilter(itrace.PenetrateErrorTracing)
	afterGatherRun = gather
	tagSpan := randomDDSpan()
	tagSpan.Meta["http.url"] = "/tmall"
	tagSpan.Meta["span.kind"] = "server"
	tagSpan.Meta["process.env"] = "env"
	tagSpan.Meta["error.msg"] = "error message"
	tagSpan.Meta["error.type"] = "type"

	ts := httptest.NewServer(&ddHandler{})
	trace := DDTraces{DDTrace{tagSpan}}
	buf, err := msgpackEncoder(trace)
	if err != nil {
		t.Errorf("err=%v", err)
		return
	}
	req, err := http.NewRequest("post", ts.URL, bytes.NewBuffer(buf))
	if err != nil {
		t.Error(err.Error())

		return
	}
	req.Header.Set("Content-Type", "application/msgpack")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err.Error())

		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("request failed with status code %d\n", resp.StatusCode)
	}
	resp.Body.Close()

	pts, err := mockFeed.AnyPoints(time.Second)
	if err == nil {
		t.Logf("point len= %d", len(pts))
		if len(pts) != 1 {
			t.Error("points len !=1")
			return
		}
		assert.Equal(t, pts[0].GetTag("http_url"), "/tmall", "must be /tmall")
		assert.Equal(t, pts[0].GetTag("error_message"), "error message", "error_message")
		assert.NotEqual(t, pts[0].GetTag("process_env"), "env", "must empty")
	}

	sampleSpan := randomDDSpan()
	sampleSpan.Metrics[keyPriority] = -1
	sampleSpan2 := randomDDSpan()
	sampleSpan2.Metrics[keyPriority] = 2
	sampleTrace := DDTraces{DDTrace{sampleSpan, sampleSpan2}}

	buf, err = msgpackEncoder(sampleTrace)
	if err != nil {
		t.Errorf("err=%v", err)
		return
	}
	req, err = http.NewRequest("post", ts.URL, bytes.NewBuffer(buf))
	if err != nil {
		t.Error(err.Error())

		return
	}
	req.Header.Set("Content-Type", "application/msgpack")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err.Error())

		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("request failed with status code %d\n", resp.StatusCode)
	}
	resp.Body.Close()

	pts, err = mockFeed.AnyPoints(time.Second)
	if err == nil {
		assert.Len(t, pts, 1)
	}
}

func randomDDSpan() *DDSpan {
	return &DDSpan{
		Service:  testutils.RandString(10),
		Name:     testutils.RandString(10),
		Resource: testutils.RandString(10),
		TraceID:  uint64(testutils.RandInt64(10)),
		SpanID:   uint64(testutils.RandInt64(10)),
		ParentID: uint64(testutils.RandInt64(10)),
		Start:    testutils.RandTime().UnixNano(),
		Duration: testutils.RandInt64(6),
		Meta:     testutils.RandTags(10, 10, 20),
		Metrics:  testutils.RandMetrics(10, 10),
		Type: testutils.RandWithinStrings([]string{
			"consul", "cache", "memcached", "redis", "aerospike", "cassandra", "db", "elasticsearch", "leveldb",
			"", "mongodb", "sql", "http", "web", "benchmark", "build", "custom", "datanucleus", "dns", "graphql", "grpc", "hibernate", "queue", "rpc", "soap", "template", "test", "worker",
		}),
	}
}

func randomDDTrace(n int) DDTrace {
	ddtrace := make(DDTrace, n)
	for i := 0; i < n; i++ {
		ddtrace[i] = randomDDSpan()
	}

	return ddtrace
}

func randomDDTraces(n, m int) DDTraces {
	ddtraces := make(DDTraces, n)
	for i := 0; i < n; i++ {
		ddtraces[i] = randomDDTrace(m)
	}

	return ddtraces
}

func jsonEncoder(ddtraces DDTraces) ([]byte, error) {
	return json.Marshal(ddtraces)
}

func BenchmarkDecodeRequest(b *testing.B) {
	buf, err := msgpackEncoder(randomDDTraces(10, 10))
	if err != nil {
		b.Error(err)
		return
	}
	parm := &itrace.TraceParameters{
		URLPath: "/v0.4/traces",
		Media:   "application/msgpack",
		Encode:  "",
		Body:    bytes.NewBuffer(buf),
	}
	b.Logf("body len =%d", len(buf))
	b.ResetTimer()
	b.Run("with pool", func(b *testing.B) {
		ddtracePoolT := &sync.Pool{
			New: func() interface{} {
				return DDTraces{}
			},
		}
		for i := 0; i < b.N; i++ {
			dt := ddtracePoolT.Get().(DDTraces)
			decodeRequest(parm, &dt)
			dt.reset()
			ddtracePoolT.Put(dt) //nolint
		}
	})

	b.ResetTimer()
	b.Run("no pool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			decodeRequest(parm, &DDTraces{})
		}
	})
}

/*
// go test -benchmem -run=^$ -tags with_inputs -cpuprofile=cpu.prof -memprofile=mem.prof  -bench ^BenchmarkDecodeRequest$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ddtrace
goos: linux
goarch: amd64
pkg: gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ddtrace
cpu: AMD Ryzen 7 7700X 8-Core Processor
BenchmarkDecodeRequest/with_pool-16                 6388            180710 ns/op           41654 B/op       3402 allocs/op
BenchmarkDecodeRequest/no_pool-16                   6806            169198 ns/op          167056 B/op       3914 allocs/op
PASS
ok      gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ddtrace    3.555s
*/

func TestInt64ToPaddedString(t *testing.T) {
	type args struct {
		num uint64
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "123",
			args: args{num: 1234567890},
			want: 16,
		},
		{
			name: "1234",
			args: args{num: 6450066879287049030},
			want: 16,
		},
		{
			name: "max",
			args: args{num: math.MaxUint64},
			want: 16,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tid := Int64ToPaddedString(tt.args.num)
			t.Logf("tid=%s", tid)
			assert.Equalf(t, tt.want, len(tid), "Int64ToPaddedString(%v)", tt.args.num)
		})
	}
}
