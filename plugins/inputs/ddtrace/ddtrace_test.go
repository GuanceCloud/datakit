package ddtrace

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DataDog/datadog-agent/pkg/trace/pb"
	"github.com/DataDog/datadog-agent/pkg/trace/test/testutil"
)

func TestTraceSample(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v0.3/traces", "/v0.4/traces", "/v0.5/traces":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Error(err)
			}
			t.Logf("body: %d", len(body))

		default:
			t.Logf("url: %s", r.URL.Path)
		}
	}))

	defer ts.Close()

	data := msgpTraces(t, pb.Traces{
		testutil.RandomTrace(10, 20),
		testutil.RandomTrace(10, 20),
		testutil.RandomTrace(10, 20),
	})

	for _, e := range []string{
		"/v0.3/traces",
		"/v0.4/traces",
		"/v0.5/traces",
	} {

		//t.Logf("body: %d", len(data))

		resp, err := http.Post(ts.URL+e, "application/msgpack", bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		_ = resp
	}
}

func msgpTraces(t *testing.T, traces pb.Traces) []byte {
	bts, err := traces.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}
	return bts
}
