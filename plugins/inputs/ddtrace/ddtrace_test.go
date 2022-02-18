package ddtrace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/msgpack"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/trace"
)

var jsonContentTypes = []string{"", "xxx", "application/json", "text/json"}

func TestDDTraceHandlers(t *testing.T) {
	afterGatherRun = itrace.AfterGatherFunc(func(inputName string, dktrace itrace.DatakitTrace, strikMod bool) {})

	rand.Seed(time.Now().UnixNano())
	testJsonDDTraces(t)
	// testMsgPackDDTraces(t)
}

func testJsonDDTraces(t *testing.T) {
	t.Helper()

	for _, version := range []string{v3, v4} {
		tsvr := httptest.NewServer(handleDDTraces(version))
		for _, method := range []string{http.MethodPost, http.MethodPut} {
			buf, err := jsonEncoder(randomDDTraces(3, 10))
			if err != nil {
				t.Error(err.Error())

				return
			}

			req, err := http.NewRequest(method, tsvr.URL+version, bytes.NewBuffer(buf))
			if err != nil {
				t.Error(err.Error())

				return
			}

			for _, contentType := range jsonContentTypes {
				req.Header.Set("Content-Type", contentType)
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Error(err.Error())

					return
				}
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					fmt.Printf("request failed with status code %d\n", resp.StatusCode)
				}
			}
		}
	}
}

func testMsgPackDDTraces(t *testing.T) {
	t.Helper()

	tsvr := httptest.NewServer(handleDDTraces(v5))
	for _, method := range []string{http.MethodPost} {
		buf, err := msgpackEncoder(randomDDTraces(3, 10))
		if err != nil {
			t.Error(err.Error())

			return
		}

		req, err := http.NewRequest(method, tsvr.URL+v5, bytes.NewBuffer(buf))
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
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("request failed with status code %d\n", resp.StatusCode)
		}
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

func msgpackEncoder(ddtraces DDTraces) ([]byte, error) {
	return msgpack.Marshal(ddtraces)
}
