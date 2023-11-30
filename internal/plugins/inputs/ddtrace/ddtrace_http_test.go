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

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

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
		fmt.Printf("request failed with status code %d\n", resp.StatusCode)
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
	buf, err := jsonEncoder(randomDDTraces(10, 10))
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
		req.Header.Set("Content-Type", "application/json")
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
