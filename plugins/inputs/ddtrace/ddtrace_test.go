package ddtrace

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DataDog/datadog-agent/pkg/trace/pb"
	"github.com/DataDog/datadog-agent/pkg/trace/test/testutil"
	"github.com/stretchr/testify/assert"
	vmsgp "github.com/vmihailenco/msgpack/v4"
)

var data = [2]interface{}{
	0: []string{
		0:  "baggage",
		1:  "item",
		2:  "elasticsearch.version",
		3:  "7.0",
		4:  "my-name",
		5:  "X",
		6:  "my-service",
		7:  "my-resource",
		8:  "_dd.sampling_rate_whatever",
		9:  "value whatever",
		10: "sql",
	},
	1: [][][12]interface{}{
		{
			{
				6,
				4,
				7,
				uint64(1),
				uint64(2),
				uint64(3),
				int64(123),
				int64(456),
				1,
				map[interface{}]interface{}{
					8: 9,
					0: 1,
					2: 3,
				},
				map[interface{}]float64{
					5: 1.2,
				},
				10,
			},
		},
	},
}

func genRamTraces(maxLevels, masSpans, count int) pb.Traces {
	traces := make(pb.Traces, count)
	for i := range traces {
		traces[i] = testutil.RandomTrace(maxLevels, masSpans)
	}

	return traces
}

func TestTracesDecodeV3V4(t *testing.T) {
	ramTraces := genRamTraces(10, 20, 10)

	tsrv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		traces, err := decodeRequest(req.URL.Path, req)
		if err != nil {
			t.Error(err)
			resp.WriteHeader(http.StatusBadRequest)
		}

		for i := range traces {
			for j := range traces[i] {
				if !assert.EqualValues(t, ramTraces[i][j], traces[i][j]) {
					t.Errorf("not equivalent span expect %v got %v", ramTraces[i][j], traces[i][j])
					resp.WriteHeader(http.StatusBadRequest)
				}
			}
		}

		resp.WriteHeader(http.StatusOK)
	}))
	defer tsrv.Close()

	bts, err := ramTraces.MarshalMsg(nil)
	if err != nil {
		t.Error(err)
	}

	for _, path := range []string{v3, v4} {
		resp, err := http.Post(tsrv.URL+path, "application/msgpack", bytes.NewBuffer(bts))
		if err != nil {
			t.Error(err)
		} else {
			if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
				t.Error(err)
			}
			defer resp.Body.Close() //nolint:errcheck
		}

		if resp.StatusCode != http.StatusOK {
			t.Error(resp.Status)
		}
	}
}

func TestTracesDecodeV5(t *testing.T) {
	tsrv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		traces, err := decodeRequest(req.URL.Path, req)
		if err != nil {
			t.Error(err)
		}

		for i := range traces {
			for j := range traces[i] {
				log.Info(traces[i][j])
			}
		}
	}))

	defer tsrv.Close() //nolint:errcheck

	bts, err := vmsgp.Marshal(data)
	if err != nil {
		t.Error(err)
	}

	resp, err := http.Post(tsrv.URL+v5, "application/msgpack", bytes.NewBuffer(bts)) //nolint:bodyclose
	if err != nil {
		t.Error(err)
	} else {
		if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
			t.Error(err)
		}
		defer resp.Body.Close() //nolint:errcheck
	}

	if resp.StatusCode != http.StatusOK {
		t.Error(err)
	}
}
