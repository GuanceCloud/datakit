// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nsq

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

//nolint:lll
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

//nolint:lll
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

//nolint:lll
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

//nolint:lll
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

//nolint:lll
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
	assert.Equal(t, datakit.AllOSWithElection, it.AvailableArchs())
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

// ---------------------------------------
/*
nsq容器启动命令：
1、docker run -d --name nsqd  -p 4150:4150 -p 4151:4151 -d nsqio/nsq:v1.2.1 /nsqd --broadcast-address=10.200.14.141  --lookupd-tcp-address=10.200.14.141:4150 --data-path=/data
2、docker run -d --name lookupd  -p 4160:4160 -p 4161:4161 -d nsqio/nsq:v1.2.1 /nsqlookupd --broadcast-address=10.200.14.141

*/
type caseSpec struct {
	t *testing.T

	name        string
	repo        string
	repoTag     string
	servicePort string

	in     *Input
	envs   []string
	feeder *io.MockedFeeder

	pool     *dockertest.Pool
	resource *dockertest.Resource

	cr *testutils.CaseResult
}

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	inputs := []struct {
		name string
		conf string
	}{
		{
			name: "nsq-lookupd",
			conf: fmt.Sprintf(`[[inputs.nsq]]
lookupd = "%s"
interval = "10s"
election = true`, testutils.GetRemote().Host+":1433"),
		},
		{
			name: "nsq-nsqd",
			conf: fmt.Sprintf(`[[inputs.nsq]]
nsqd = ["%s","%s"]
interval = "10s"
election = true`, testutils.GetRemote().Host+":2433", testutils.GetRemote().Host+":4433"),
		},
		{
			name: "nsq-tags",
			conf: fmt.Sprintf(`[[inputs.nsq]]
lookupd = "%s"
interval = "10s"
election = true
[inputs.nsq.tags]
	more_tag = "some_other_value"`, testutils.GetRemote().Host+":3433"),
		},
	}

	images := [][2]string{
		{"docker.io/nsqio/nsq", "v1.2.1"},
		// {"docker.io/nsqio/nsq", "v1.2.1"},
		// {"docker.io/nsqio/nsq", "v1.1.0"},
	}

	cases := make([]*caseSpec, 0, 10)

	for _, in := range inputs {
		for _, im := range images {
			ipt := newInput()

			_, err := toml.Decode(in.conf, ipt)
			assert.NoError(t, err)

			// envs := []string{
			// 	"ACCEPT_EULA=Y",
			// 	fmt.Sprintf("SA_PASSWORD=%s", ipt.Interval),
			// }

			cases = append(cases, &caseSpec{
				name:    in.name,
				repo:    im[0],
				repoTag: im[1],
				in:      ipt,
				envs:    []string{},
				feeder:  io.NewMockedFeeder(),

				cr: &testutils.CaseResult{
					Name:        t.Name(),
					Case:        in.name,
					ExtraFields: map[string]any{},
					ExtraTags: map[string]string{
						"image":         im[0],
						"image_tag":     im[1],
						"remote_server": ipt.Lookupd,
					},
				},
			})
		}
	}

	return cases, nil
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		switch string(pt.Name()) {
		case "nsq_performance":

			// TODO: check pt according to Performance

		default: // TODO: check other measurement
		}

		// check if tag appended
		if len(cs.in.Tags) != 0 {
			cs.t.Logf("checking tags %+#v...", cs.in.Tags)

			tags := pt.Tags()
			for k, expect := range cs.in.Tags {
				if v := tags.Get([]byte(k)); v != nil {
					got := string(v.GetD())
					if got != expect {
						return fmt.Errorf("expect tag value %s, got %s", expect, got)
					}
				} else {
					return fmt.Errorf("tag %s not found, got %v", k, tags)
				}
			}
		}
	}

	// TODO: some other checking on @pts, such as `if some required measurements exist'...

	return nil
}

func (cs *caseSpec) run() error {
	// start remote nsq
	r := testutils.GetRemote()
	dockerTCP := r.TCPURL()

	cs.t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	start := time.Now()

	p, err := dockertest.NewPool(dockerTCP)
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		cs.t.Logf("get hostname failed: %s, ignored", err)
		hostname = "unknown-hostname"
	}

	containerName := fmt.Sprintf("%s.%s", hostname, cs.name)

	// remove the container if exist.
	if err := p.RemoveContainerByName(containerName); err != nil {
		return err
	}

	resource, err := p.RunWithOptions(&dockertest.RunOptions{
		// specify container image & tag
		Repository: cs.repo,
		Tag:        cs.repoTag,

		// port binding
		PortBindings: map[docker.Port][]docker.PortBinding{
			"1433/tcp": {{HostIP: "0.0.0.0", HostPort: cs.servicePort}},
		},

		Name: containerName,

		// container run-time envs
		Env: cs.envs,
	}, func(c *docker.HostConfig) {
		c.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return err
	}

	cs.pool = p
	cs.resource = resource

	cs.t.Logf("check service(%s:%s)...", r.Host, cs.servicePort)
	if !r.PortOK(cs.servicePort, time.Minute) {
		return fmt.Errorf("service checking failed")
	}

	cs.cr.AddField("container_ready_cost", int64(time.Since(start)))

	var wg sync.WaitGroup

	// start input
	cs.t.Logf("start input...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		cs.in.Run()
	}()

	// wait data
	start = time.Now()
	cs.t.Logf("wait points...")
	pts, err := cs.feeder.AnyPoints()
	if err != nil {
		return err
	}

	cs.cr.AddField("point_latency", int64(time.Since(start)))
	cs.cr.AddField("point_count", len(pts))

	cs.t.Logf("get %d points", len(pts))
	if err := cs.checkPoint(pts); err != nil {
		return err
	}

	cs.t.Logf("stop input...")
	cs.in.Terminate()

	cs.t.Logf("exit...")
	wg.Wait()

	return nil
}

func TestIntergration(t *testing.T) {
	start := time.Now()
	cases, err := buildCases(t)
	if err != nil {
		cs := testutils.CaseResult{
			Name:          t.Name(),
			Status:        testutils.CasePassed,
			FailedMessage: err.Error(),
			Cost:          time.Since(start),
		}

		_ = cs.Flush()
		return
	}

	t.Logf("testing %d cases...", len(cases))

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			caseStart := time.Now()

			t.Logf("testing %s...", tc.name)

			if err := tc.run(); err != nil {
				tc.cr.Status = testutils.CaseFailed
				tc.cr.FailedMessage = err.Error()

				assert.NoError(t, err)
			} else {
				tc.cr.Status = testutils.CasePassed
			}

			tc.cr.Cost = time.Since(caseStart)

			assert.NoError(t, tc.cr.Flush())

			t.Cleanup(func() {
				// clean remote docker resources
				if tc.resource == nil {
					return
				}

				assert.NoError(t, tc.pool.Purge(tc.resource))
			})
		})
	}
}

func TestDockerNsq(t *testing.T) {
	r := testutils.GetRemote()
	r.Host = "10.200.14.141"
	dockerTCP := r.TCPURL()

	t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	p, err := dockertest.NewPool(dockerTCP)
	if err != nil {
		return
	}

	hostname, err := os.Hostname()
	if err != nil {
		t.Logf("get hostname failed: %s, ignored", err)
		hostname = "unknown-hostname"
	}

	// remove the container if exist.
	if err := p.RemoveContainerByName(hostname); err != nil {
		return
	}

	resource, err := p.RunWithOptions(&dockertest.RunOptions{
		// specify container image & tag
		Repository: "docker.io/nsqio/nsq",
		Tag:        "v1.2.1",

		// port binding
		PortBindings: map[docker.Port][]docker.PortBinding{
			"4170/tcp": {{HostIP: "0.0.0.0", HostPort: "4170"}},
			"4171/tcp": {{HostIP: "0.0.0.0", HostPort: "4171"}},
		},

		Name: hostname,

		// container run-time envs
		Env: []string{},
		Cmd: []string{
			"/nsqlookupd",
			"--broadcast-address=" + r.Host,
		},
	}, func(c *docker.HostConfig) {
		c.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return
	}

	assert.NoError(t, p.Purge(resource))
}
