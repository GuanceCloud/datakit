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
	"net/http"
	"net/http/httptest"
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

	return buf.Bytes(), err
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
