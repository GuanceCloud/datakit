package nsq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint
func TestStatsPoint(t *testing.T) {
	cases := []struct {
		body string
	}{
		{
			`{"version":"1.2.0","health":"OK","start_time":1630393108,"topics":[{"topic_name":"topic-A","channels":[{"channel_name":"chan-A","depth":0,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":0,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-B","channels":[{"channel_name":"chan-B","depth":0,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":0,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-C","channels":[{"channel_name":"chan-C","depth":0,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":0,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-D","channels":[{"channel_name":"chan-D","depth":0,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":0,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"memory":{"heap_objects":5781,"heap_idle_bytes":63447040,"heap_in_use_bytes":2842624,"heap_released_bytes":0,"gc_pause_usec_100":0,"gc_pause_usec_99":0,"gc_pause_usec_95":0,"next_gc_bytes":4473924,"gc_total_runs":0},"producers":[]}`,
		},
		{
			`{"version":"1.2.0","health":"OK","start_time":1630393108,"topics":[{"topic_name":"topic-A","channels":[{"channel_name":"chan-A","depth":0,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":0,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-B","channels":[{"channel_name":"chan-B","depth":0,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":0,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-C","channels":[{"channel_name":"chan-C","depth":0,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":0,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-D","channels":[{"channel_name":"chan-D","depth":0,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"channel_name":"chan-E","depth":0,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":0,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"memory":{"heap_objects":2869,"heap_idle_bytes":63979520,"heap_in_use_bytes":2179072,"heap_released_bytes":63946752,"gc_pause_usec_100":888,"gc_pause_usec_99":327,"gc_pause_usec_95":225,"next_gc_bytes":4194304,"gc_total_runs":900},"producers":[]}`,
		},
	}

	st := newStats(nil)

	for _, tc := range cases {
		err := st.add([]byte(tc.body))
		assert.NoError(t, err)
	}

	pts, err := st.makePoint()
	assert.NoError(t, err)

	for _, pt := range pts {
		t.Log(pt.String())

	}
}
