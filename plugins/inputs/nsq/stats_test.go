package nsq

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//nolint
func TestStatsPoint(t *testing.T) {
	var pointTime = time.Unix(0, 1630911069947935363)

	bodyCases := []string{
		`{"version":"1.2.0","health":"OK","start_time":1630393108,"topics":[{"topic_name":"topic-A","channels":[{"channel_name":"chan-A","depth":10,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":10,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-B","channels":[{"channel_name":"chan-B","depth":20,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":20,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-C","channels":[{"channel_name":"chan-C","depth":30,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":30,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-D","channels":[{"channel_name":"chan-D","depth":40,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":40,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"memory":{"heap_objects":5781,"heap_idle_bytes":63447040,"heap_in_use_bytes":2842624,"heap_released_bytes":0,"gc_pause_usec_100":0,"gc_pause_usec_99":0,"gc_pause_usec_95":0,"next_gc_bytes":4473924,"gc_total_runs":0},"producers":[]}`,
		`{"version":"1.2.0","health":"OK","start_time":1630393108,"topics":[{"topic_name":"topic-A","channels":[{"channel_name":"chan-A","depth":11,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":11,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-B","channels":[{"channel_name":"chan-B","depth":21,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":21,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-C","channels":[{"channel_name":"chan-C","depth":31,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":31,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-D","channels":[{"channel_name":"chan-D","depth":41,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"channel_name":"chan-E","depth":51,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":92,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"memory":{"heap_objects":2869,"heap_idle_bytes":63979520,"heap_in_use_bytes":2179072,"heap_released_bytes":63946752,"gc_pause_usec_100":888,"gc_pause_usec_99":327,"gc_pause_usec_95":225,"next_gc_bytes":4194304,"gc_total_runs":900},"producers":[]}`,
	}

	ptsCases := map[string]interface{}{
		`nsq_topics,channel=chan-A,t_key=t_value,topic=topic-A backend_depth=0i,deferred_count=0i,depth=21i,in_flight_count=0i,message_count=0i,requeue_count=0i,timeout_count=0i 1630911069947935363`: nil,
		`nsq_topics,channel=chan-B,t_key=t_value,topic=topic-B backend_depth=0i,deferred_count=0i,depth=41i,in_flight_count=0i,message_count=0i,requeue_count=0i,timeout_count=0i 1630911069947935363`: nil,
		`nsq_topics,channel=chan-C,t_key=t_value,topic=topic-C backend_depth=0i,deferred_count=0i,depth=61i,in_flight_count=0i,message_count=0i,requeue_count=0i,timeout_count=0i 1630911069947935363`: nil,
		`nsq_topics,channel=chan-D,t_key=t_value,topic=topic-D backend_depth=0i,deferred_count=0i,depth=81i,in_flight_count=0i,message_count=0i,requeue_count=0i,timeout_count=0i 1630911069947935363`: nil,
		`nsq_topics,channel=chan-E,t_key=t_value,topic=topic-D backend_depth=0i,deferred_count=0i,depth=51i,in_flight_count=0i,message_count=0i,requeue_count=0i,timeout_count=0i 1630911069947935363`: nil,
		`nsq_nodes,server_host=testhost,t_key=t_value backend_depth=0i,depth=337i,message_count=0i 1630911069947935363`:                                                                                nil,
	}

	st := newStats()
	for _, body := range bodyCases {
		err := st.add("testhost", []byte(body))
		assert.NoError(t, err)
	}

	pts, err := st.makePoint(map[string]string{"t_key": "t_value"}, pointTime)
	assert.NoError(t, err)

	if len(pts) != len(ptsCases) {
		t.Errorf("shoud %d, got %d", len(pts), len(ptsCases))
	}

	for _, pt := range pts {
		s := pt.String()
		_, ok := ptsCases[s]
		if !ok {
			t.Errorf("not found %s", s)
		}
	}
}
