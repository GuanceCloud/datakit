package nsq

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

//nolint
func TestGather(t *testing.T) {
	bodyCases := []string{
		`{"version":"1.2.0","health":"OK","start_time":1630393108,"topics":[{"topic_name":"topic-A","channels":[{"channel_name":"chan-A","depth":10,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":10,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-B","channels":[{"channel_name":"chan-B","depth":20,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":20,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-C","channels":[{"channel_name":"chan-C","depth":30,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":30,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-D","channels":[{"channel_name":"chan-D","depth":40,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":40,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"memory":{"heap_objects":5781,"heap_idle_bytes":63447040,"heap_in_use_bytes":2842624,"heap_released_bytes":0,"gc_pause_usec_100":0,"gc_pause_usec_99":0,"gc_pause_usec_95":0,"next_gc_bytes":4473924,"gc_total_runs":0},"producers":[]}`,
		`{"version":"1.2.0","health":"OK","start_time":1630393108,"topics":[{"topic_name":"topic-A","channels":[{"channel_name":"chan-A","depth":11,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":11,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-B","channels":[{"channel_name":"chan-B","depth":21,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":21,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-C","channels":[{"channel_name":"chan-C","depth":31,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":31,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"topic_name":"topic-D","channels":[{"channel_name":"chan-D","depth":41,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}},{"channel_name":"chan-E","depth":51,"backend_depth":0,"in_flight_count":0,"deferred_count":0,"message_count":0,"requeue_count":0,"timeout_count":0,"client_count":0,"clients":[],"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"depth":92,"backend_depth":0,"message_count":0,"message_bytes":0,"paused":false,"e2e_processing_latency":{"count":0,"percentiles":null}}],"memory":{"heap_objects":2869,"heap_idle_bytes":63979520,"heap_in_use_bytes":2179072,"heap_released_bytes":63946752,"gc_pause_usec_100":888,"gc_pause_usec_99":327,"gc_pause_usec_95":225,"next_gc_bytes":4194304,"gc_total_runs":900},"producers":[]}`,
	}

	getNSQDEndpoint := func(s string) string {
		u, _ := url.Parse(s)
		return "http://" + u.Host
	}

	for _, body := range bodyCases {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))

		it := newInput()
		t.Log(ts.URL)
		it.NSQDs = []string{getNSQDEndpoint(ts.URL)}
		it.setup()

		pts, err := it.gather()
		assert.NoError(t, err)
		t.Logf("%v", pts)
	}
}

//nolint
func TestNSQDList(t *testing.T) {
	cases := []struct {
		body string
		out  []string
	}{
		{
			`{"producers":[{"remote_address":"172.19.0.4:55156","hostname":"5b44717bc03c","broadcast_address":"192.168.0.10","tcp_port":4150,"http_port":4151,"version":"1.2.0","tombstones":[],"topics":[]},{"remote_address":"172.19.0.2:47644","hostname":"0927e72b938b","broadcast_address":"192.168.0.10","tcp_port":14150,"http_port":14151,"version":"1.2.0","tombstones":[false,false,false,false],"topics":["influx-data","df-trigger-metering","df-calculate-metering","df-billing"]},{"remote_address":"172.19.0.5:48006","hostname":"702d89de2a23","broadcast_address":"192.168.0.10","tcp_port":14154,"http_port":14155,"version":"1.2.0","tombstones":[false,false,false,false],"topics":["df-billing","df-trigger-metering","df-calculate-metering","influx-data"]}]}`,
			[]string{"http://192.168.0.10:4151/stats?format=json", "http://192.168.0.10:14151/stats?format=json", "http://192.168.0.10:14155/stats?format=json"},
		},

		{
			`{"producers":[{"remote_address":"172.19.0.4:55512","hostname":"0927e72b938b","broadcast_address":"192.168.0.10","tcp_port":14150,"http_port":14151,"version":"1.2.0","tombstones":[false,false,false,false],"topics":["influx-data","df-trigger-metering","df-calculate-metering","df-billing"]},{"remote_address":"172.19.0.5:55134","hostname":"702d89de2a23","broadcast_address":"192.168.0.10","tcp_port":14154,"http_port":14155,"version":"1.2.0","tombstones":[false,false,false,false],"topics":["df-billing","df-trigger-metering","df-calculate-metering","influx-data"]}]}`,
			[]string{"http://192.168.0.10:14151/stats?format=json", "http://192.168.0.10:14155/stats?format=json"},
		},
		{
			`{"producers":[]}`,
			[]string{},
		},
	}

	for _, tc := range cases {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, tc.body)
		}))

		it := newInput()
		err := it.updateEndpointListByLookupd(ts.URL)
		assert.NoError(t, err)

		if len(tc.out) != len(it.nsqdEndpointList) {
			t.Errorf("shoud %d, got %d", len(tc.out), len(it.nsqdEndpointList))
			continue
		}

		for _, out := range tc.out {
			_, ok := it.nsqdEndpointList[out]
			if !ok {
				t.Errorf("not found %s", out)
			}
		}
	}
}

//nolint
func TestSetup(t *testing.T) {
	cases := []struct {
		in   *Input
		fail bool
	}{
		{
			func() *Input { it := newInput(); it.Lookupd = "http://dummy:1"; return it }(),
			true,
		},
		{
			func() *Input { it := newInput(); it.NSQDs = []string{"http://dummy:1"}; return it }(),
			false,
		},
		{
			func() *Input { it := newInput(); it.Interval = "10s"; return it }(),
			true,
		},
		{
			func() *Input { it := newInput(); it.Interval = "10s"; it.NSQDs = []string{"http://dummy:1"}; return it }(),
			false,
		},
		{
			func() *Input { it := newInput(); it.Interval = "dummy_interval"; return it }(),
			true,
		},
	}

	for _, tc := range cases {
		err := tc.in.setupDo()
		if tc.fail && assert.Error(t, err) {
			continue
		} else {
			assert.NoError(t, err)
		}
	}
}

//nolint
func TestBuildURL(t *testing.T) {
	cases := []struct {
		in   string
		fail bool
	}{
		{
			"http://dummy:1",
			false,
		},
		{
			":1",
			true,
		},
	}

	for _, tc := range cases {
		_, err := buildURL(tc.in)
		if tc.fail && assert.Error(t, err) {
			continue
		} else {
			assert.NoError(t, err)
		}
	}
}

//nolint
func TestRUn(t *testing.T) {
	minInterval = time.Second * 1
	updateEndpointListInterval = time.Second * 1

	it := newInput()
	it.Interval = "1s"
	it.NSQDs = []string{"http://dummy:1"}
	go it.Run()

	time.Sleep(time.Second * 3)
	datakit.Exit.Close()
}

func TestOther(t *testing.T) {
	it := newInput()
	assert.Equal(t, sampleCfg, it.SampleConfig())
	assert.Equal(t, catalog, it.Catalog())
	assert.Equal(t, datakit.AllArch, it.AvailableArchs())
}

func TestMan(t *testing.T) {
	i := &Input{}
	arr := i.SampleMeasurement()

	for _, elem := range arr {
		if _, err := elem.LineProto(); err != nil {
			t.Error(err)
		}
		elem.Info()
	}
}
