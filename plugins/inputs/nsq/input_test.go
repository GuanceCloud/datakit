package nsq

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	cases := []struct {
		body string
	}{
		{
			`{"version":"1.2.0","health":"OK","start_time":1630393108,"topics":[{"topic_name":"df-billing","channels":[{"channel_name":"df-billing-channel","depth":0,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":0,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"df-calculate-metering","channels":[{"channel_name":"df-calculate-metering-channel","depth":0,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":0,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"df-trigger-metering","channels":[{"channel_name":"df-trigger-metering-channel","depth":0,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":0,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"influx-data","channels":[{"channel_name":"influx-data-channel","depth":0,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":0,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"memory":{"heap_objects":5781,"heap_idle_bytes":63447040,"heap_in_use_bytes":2842624,"heap_released_bytes":0,"gc_pause_usec_100":0,"gc_pause_usec_99":0,"gc_pause_usec_95":0,"next_gc_bytes":4473924,"gc_total_runs":0},"producers":[]}`,
		},
	}

	for _, tc := range cases {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, tc.body)
		}))

		pts, err := newInput().gatherEndpoint(ts.URL)
		assert.NoError(t, err)

		for _, pt := range pts {
			t.Log(pt.String())
		}
	}
}

func TestMan(t *testing.T) {
	i := &Input{}
	arr := i.SampleMeasurement()

	for _, elem := range arr {
		elem.LineProto()
		elem.Info()
	}
}
