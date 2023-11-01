// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package coredns

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	dt "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/prom"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

// ATTENTION: Docker version should use v20.10.18 in integrate tests. Other versions are not tested.

func TestIntegrate(t *testing.T) {
	if !testutils.CheckIntegrationTestingRunning() {
		t.Skip()
	}

	testutils.PurgeRemoteByName(inputName)       // purge at first.
	defer testutils.PurgeRemoteByName(inputName) // purge at last.

	start := time.Now()
	cases, err := buildCases(t)
	if err != nil {
		cr := &testutils.CaseResult{
			Name:          t.Name(),
			Status:        testutils.TestPassed,
			FailedMessage: err.Error(),
			Cost:          time.Since(start),
		}

		_ = testutils.Flush(cr)
		return
	}

	t.Logf("testing %d cases...", len(cases))

	for _, tc := range cases {
		func(tc *caseSpec) {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				caseStart := time.Now()

				t.Logf("testing %s...", tc.name)

				if err := testutils.RetryTestRun(tc.run); err != nil {
					tc.cr.Status = testutils.TestFailed
					tc.cr.FailedMessage = err.Error()

					panic(err)
				} else {
					tc.cr.Status = testutils.TestPassed
				}

				tc.cr.Cost = time.Since(caseStart)

				require.NoError(t, testutils.Flush(tc.cr))

				t.Cleanup(func() {
					// clean remote docker resources
					if tc.resource == nil {
						return
					}

					tc.pool.Purge(tc.resource)
				})
			})
		}(tc)
	}
}

func getConfAccessPoint(host, port string) string {
	return fmt.Sprintf("http://%s/metrics", net.JoinHostPort(host, port))
}

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name         string // Also used as build image name:tag.
		conf         string
		exposedPorts []string
		mPathCount   map[string]int

		optsACL      []inputs.PointCheckOption
		optsCache    []inputs.PointCheckOption
		optsDNSSec   []inputs.PointCheckOption
		optsForward  []inputs.PointCheckOption
		optsGrpc     []inputs.PointCheckOption
		optsHosts    []inputs.PointCheckOption
		optsTemplate []inputs.PointCheckOption
		optsProm     []inputs.PointCheckOption
	}{
		{
			name: "pubrepo.guance.com/image-repo-for-testing/coredns/coredns:1.10.1",
			// selfBuild: true,
			conf: `
source = "coredns"
metric_types = ["counter", "gauge"]
metric_name_filter = ["^coredns_(acl|cache|dnssec|forward|grpc|hosts|template|dns)_([a-z_]+)$"]
interval = "10s"
tls_open = false
url = ""
[[measurements]]
  prefix = "coredns_acl_"
  name = "coredns_acl"
[[measurements]]
  prefix = "coredns_cache_"
  name = "coredns_cache"
[[measurements]]
  prefix = "coredns_dnssec_"
  name = "coredns_dnssec"
[[measurements]]
  prefix = "coredns_forward_"
  name = "coredns_forward"
[[measurements]]
  prefix = "coredns_grpc_"
  name = "coredns_grpc"
[[measurements]]
  prefix = "coredns_hosts_"
  name = "coredns_hosts"
[[measurements]]
  prefix = "coredns_template_"
  name = "coredns_template"
[[measurements]]
  prefix = "coredns_dns_"
  name = "coredns"
`, // set conf URL later.
			exposedPorts: []string{"9153/tcp"},
			mPathCount:   map[string]int{"/": 10},

			optsACL: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("server", "zone", "instance", "host"),
				inputs.WithOptionalFields("dropped_requests_total", "blocked_requests_total", "filtered_requests_total", "allowed_requests_total"),
			},
			optsCache: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("type", "instance", "host", "server", "zones"),
				inputs.WithOptionalFields("prefetch_total", "drops_total", "served_stale_total", "evictions_total", "entries", "requests_total", "hits_total", "misses_total"),
			},
			optsDNSSec: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("server", "type", "instance", "host"),
				inputs.WithOptionalFields("cache_entries", "cache_hits_total", "cache_misses_total"),
			},
			optsForward: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("rcode", "proto", "instance", "host", "to"),
				inputs.WithOptionalFields("max_concurrent_rejects_total", "conn_cache_hits_total", "conn_cache_misses_total", "requests_total", "responses_total", "request_duration_seconds", "healthcheck_failures_total", "healthcheck_broken_total"),
			},
			optsGrpc: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("rcode", "instance", "host", "to"),
				inputs.WithOptionalFields("requests_total", "responses_total", "request_duration_seconds"),
			},
			optsHosts: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("host", "instance"),
				inputs.WithOptionalFields("entries", "reload_timestamp_seconds"),
			},
			optsTemplate: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("type", "host", "view", "zone", "class", "section", "template", "instance", "server"),
				inputs.WithOptionalFields("rr_failures_total", "matches_total", "failures_total"),
			},
			optsProm: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("service_kind", "view", "proto", "hash", "value", "host", "rcode", "name", "instance", "server", "goversion", "plugin", "status", "version", "revision", "zone"),
				inputs.WithOptionalFields("health_request_duration_seconds", "dns_https_responses_total", "reload_failed_total", "autopath_success_total", "health_request_failures_total", "local_localhost_requests_total", "build_info", "dns_do_requests_total", "dns_responses_total", "dns_plugin_enabled", "dns64_requests_translated_total", "kubernetes_dns_programming_duration_seconds", "dns_requests_total", "dns_request_size_bytes", "dns_panics_total", "dns_request_duration_seconds", "dns_response_size_bytes", "reload_version_info"),
			},
		},
		{
			name: "pubrepo.guance.com/image-repo-for-testing/coredns/coredns:1.9.4",
			// selfBuild: true,
			conf: `
source = "coredns"
metric_types = ["counter", "gauge"]
metric_name_filter = ["^coredns_(acl|cache|dnssec|forward|grpc|hosts|template|dns)_([a-z_]+)$"]
interval = "10s"
tls_open = false
url = ""
[[measurements]]
  prefix = "coredns_acl_"
  name = "coredns_acl"
[[measurements]]
  prefix = "coredns_cache_"
  name = "coredns_cache"
[[measurements]]
  prefix = "coredns_dnssec_"
  name = "coredns_dnssec"
[[measurements]]
  prefix = "coredns_forward_"
  name = "coredns_forward"
[[measurements]]
  prefix = "coredns_grpc_"
  name = "coredns_grpc"
[[measurements]]
  prefix = "coredns_hosts_"
  name = "coredns_hosts"
[[measurements]]
  prefix = "coredns_template_"
  name = "coredns_template"
[[measurements]]
  prefix = "coredns_dns_"
  name = "coredns"
`, // set conf URL later.
			exposedPorts: []string{"9153/tcp"},
			mPathCount:   map[string]int{"/": 10},

			optsACL: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("server", "zone", "instance", "host"),
				inputs.WithOptionalFields("dropped_requests_total", "blocked_requests_total", "filtered_requests_total", "allowed_requests_total"),
			},
			optsCache: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("type", "instance", "host", "server", "zones"),
				inputs.WithOptionalFields("prefetch_total", "drops_total", "served_stale_total", "evictions_total", "entries", "requests_total", "hits_total", "misses_total"),
			},
			optsDNSSec: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("server", "type", "instance", "host"),
				inputs.WithOptionalFields("cache_entries", "cache_hits_total", "cache_misses_total"),
			},
			optsForward: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("rcode", "proto", "instance", "host", "to"),
				inputs.WithOptionalFields("max_concurrent_rejects_total", "conn_cache_hits_total", "conn_cache_misses_total", "requests_total", "responses_total", "request_duration_seconds", "healthcheck_failures_total", "healthcheck_broken_total"),
			},
			optsGrpc: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("rcode", "instance", "host", "to"),
				inputs.WithOptionalFields("requests_total", "responses_total", "request_duration_seconds"),
			},
			optsHosts: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("host", "instance"),
				inputs.WithOptionalFields("entries", "reload_timestamp_seconds"),
			},
			optsTemplate: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("type", "host", "view", "zone", "class", "section", "template", "instance", "server"),
				inputs.WithOptionalFields("rr_failures_total", "matches_total", "failures_total"),
			},
			optsProm: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("service_kind", "view", "proto", "hash", "value", "host", "rcode", "name", "instance", "server", "goversion", "plugin", "status", "version", "revision", "zone"),
				inputs.WithOptionalFields("health_request_duration_seconds", "dns_https_responses_total", "reload_failed_total", "autopath_success_total", "health_request_failures_total", "local_localhost_requests_total", "build_info", "dns_do_requests_total", "dns_responses_total", "dns_plugin_enabled", "dns64_requests_translated_total", "kubernetes_dns_programming_duration_seconds", "dns_requests_total", "dns_request_size_bytes", "dns_panics_total", "dns_request_duration_seconds", "dns_response_size_bytes", "reload_version_info"),
			},
		},
		{
			name: "pubrepo.guance.com/image-repo-for-testing/coredns/coredns:1.8.7",
			// selfBuild: true,
			conf: `
source = "coredns"
metric_types = ["counter", "gauge"]
metric_name_filter = ["^coredns_(acl|cache|dnssec|forward|grpc|hosts|template|dns)_([a-z_]+)$"]
interval = "10s"
tls_open = false
url = ""
[[measurements]]
  prefix = "coredns_acl_"
  name = "coredns_acl"
[[measurements]]
  prefix = "coredns_cache_"
  name = "coredns_cache"
[[measurements]]
  prefix = "coredns_dnssec_"
  name = "coredns_dnssec"
[[measurements]]
  prefix = "coredns_forward_"
  name = "coredns_forward"
[[measurements]]
  prefix = "coredns_grpc_"
  name = "coredns_grpc"
[[measurements]]
  prefix = "coredns_hosts_"
  name = "coredns_hosts"
[[measurements]]
  prefix = "coredns_template_"
  name = "coredns_template"
[[measurements]]
  prefix = "coredns_dns_"
  name = "coredns"
`, // set conf URL later.
			exposedPorts: []string{"9153/tcp"},
			mPathCount:   map[string]int{"/": 10},

			optsACL: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("server", "zone", "instance", "host"),
				inputs.WithOptionalFields("dropped_requests_total", "blocked_requests_total", "filtered_requests_total", "allowed_requests_total"),
			},
			optsCache: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("type", "instance", "host", "server", "zones"),
				inputs.WithOptionalFields("prefetch_total", "drops_total", "served_stale_total", "evictions_total", "entries", "requests_total", "hits_total", "misses_total"),
			},
			optsDNSSec: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("server", "type", "instance", "host"),
				inputs.WithOptionalFields("cache_entries", "cache_hits_total", "cache_misses_total"),
			},
			optsForward: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("rcode", "proto", "instance", "host", "to"),
				inputs.WithOptionalFields("max_concurrent_rejects_total", "conn_cache_hits_total", "conn_cache_misses_total", "requests_total", "responses_total", "request_duration_seconds", "healthcheck_failures_total", "healthcheck_broken_total"),
			},
			optsGrpc: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("rcode", "instance", "host", "to"),
				inputs.WithOptionalFields("requests_total", "responses_total", "request_duration_seconds"),
			},
			optsHosts: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("host", "instance"),
				inputs.WithOptionalFields("entries", "reload_timestamp_seconds"),
			},
			optsTemplate: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("type", "host", "view", "zone", "class", "section", "template", "instance", "server"),
				inputs.WithOptionalFields("rr_failures_total", "matches_total", "failures_total"),
			},
			optsProm: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalTags("service_kind", "view", "proto", "hash", "value", "host", "rcode", "name", "instance", "server", "goversion", "plugin", "status", "version", "revision", "zone"),
				inputs.WithOptionalFields("health_request_duration_seconds", "dns_https_responses_total", "reload_failed_total", "autopath_success_total", "health_request_failures_total", "local_localhost_requests_total", "build_info", "dns_do_requests_total", "dns_responses_total", "dns_plugin_enabled", "dns64_requests_translated_total", "kubernetes_dns_programming_duration_seconds", "dns_requests_total", "dns_request_size_bytes", "dns_panics_total", "dns_request_duration_seconds", "dns_response_size_bytes", "reload_version_info"),
			},
		},
	}

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := io.NewMockedFeeder()

		ipt := prom.NewProm()
		ipt.Feeder = feeder

		_, err := toml.Decode(base.conf, ipt)
		require.NoError(t, err)

		// URL from ENV.
		envs := []string{
			"ALLOW_NONE_AUTHENTICATION=yes",
		}

		repoTag := strings.Split(base.name, ":")
		cases = append(cases, &caseSpec{
			t:            t,
			ipt:          ipt,
			name:         base.name,
			feeder:       feeder,
			envs:         envs,
			repo:         repoTag[0],
			repoTag:      repoTag[1],
			exposedPorts: base.exposedPorts,

			optsACL:      base.optsACL,
			optsCache:    base.optsCache,
			optsDNSSec:   base.optsDNSSec,
			optsForward:  base.optsForward,
			optsGrpc:     base.optsGrpc,
			optsHosts:    base.optsHosts,
			optsTemplate: base.optsTemplate,
			optsProm:     base.optsProm,

			cr: &testutils.CaseResult{
				Name:        t.Name(),
				Case:        base.name,
				ExtraFields: map[string]any{},
				ExtraTags: map[string]string{
					"image":       repoTag[0],
					"image_tag":   repoTag[1],
					"docker_host": remote.Host,
					"docker_port": remote.Port,
				},
			},
		})
	}

	return cases, nil
}

////////////////////////////////////////////////////////////////////////////////

// caseSpec.
type caseSpec struct {
	t *testing.T

	name         string
	repo         string
	repoTag      string
	envs         []string
	exposedPorts []string
	serverPorts  []string
	mCount       map[string]struct{}

	optsACL      []inputs.PointCheckOption
	optsCache    []inputs.PointCheckOption
	optsDNSSec   []inputs.PointCheckOption
	optsForward  []inputs.PointCheckOption
	optsGrpc     []inputs.PointCheckOption
	optsHosts    []inputs.PointCheckOption
	optsTemplate []inputs.PointCheckOption
	optsProm     []inputs.PointCheckOption

	ipt    *prom.Input
	feeder *io.MockedFeeder

	pool     *dt.Pool
	resource *dt.Resource

	cr *testutils.CaseResult
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		measurement := pt.Name()
		opts := []inputs.PointCheckOption{}

		switch measurement {
		case "coredns_acl":
			opts = append(opts, cs.optsACL...)
			opts = append(opts, inputs.WithDoc(&aCLMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[measurement] = struct{}{}
		case "coredns_cache":
			opts = append(opts, cs.optsCache...)
			opts = append(opts, inputs.WithDoc(&cacheMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[measurement] = struct{}{}
		case "coredns_dnssec":
			opts = append(opts, cs.optsDNSSec...)
			opts = append(opts, inputs.WithDoc(&dnsSecMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[measurement] = struct{}{}
		case "coredns_forward":
			opts = append(opts, cs.optsForward...)
			opts = append(opts, inputs.WithDoc(&forwardMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[measurement] = struct{}{}
		case "coredns_grpc":
			opts = append(opts, cs.optsGrpc...)
			opts = append(opts, inputs.WithDoc(&grpcMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[measurement] = struct{}{}
		case "coredns_hosts":
			opts = append(opts, cs.optsHosts...)
			opts = append(opts, inputs.WithDoc(&hostsMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[measurement] = struct{}{}
		case "coredns_template":
			opts = append(opts, cs.optsTemplate...)
			opts = append(opts, inputs.WithDoc(&templateMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[measurement] = struct{}{}
		case "coredns":
			opts = append(opts, cs.optsProm...)
			opts = append(opts, inputs.WithDoc(&promMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[measurement] = struct{}{}

		default: // TODO: check other measurement
			panic("unknown measurement: " + measurement)
		}

		// check if tag appended
		if len(cs.ipt.Tags) != 0 {
			cs.t.Logf("checking tags %+#v...", cs.ipt.Tags)

			tags := pt.Tags()
			for k, expect := range cs.ipt.Tags {
				if v := tags.Get(k); v != nil {
					got := v.GetS()
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
	// start remote image server
	r := testutils.GetRemote()
	dockerTCP := r.TCPURL() // got "tcp://" + net.JoinHostPort(i.Host, i.Port) 2375

	cs.t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	start := time.Now()

	p, err := dt.NewPool(dockerTCP)
	if err != nil {
		return err
	}

	uniqueContainerName := testutils.GetUniqueContainerName(inputName)

	resource, err := p.RunWithOptions(&dt.RunOptions{
		Name: uniqueContainerName, // ATTENTION: not cs.name.

		// specify container image & tag
		Repository: cs.repo,
		Tag:        cs.repoTag,

		ExposedPorts: cs.exposedPorts,

		// container run-time envs
		Env: cs.envs,
	}, func(c *docker.HostConfig) {
		c.RestartPolicy = docker.RestartPolicy{Name: "no"}
		c.AutoRemove = true
	})
	if err != nil {
		return err
	}

	cs.pool = p
	cs.resource = resource

	if err := cs.getMappingPorts(); err != nil {
		return err
	}
	cs.ipt.URL = getConfAccessPoint(r.Host, cs.serverPorts[0]) // set conf URL here.

	cs.t.Logf("check service(%s:%v)...", r.Host, cs.serverPorts)

	if err := cs.portsOK(r); err != nil {
		return err
	}

	cs.cr.AddField("container_ready_cost", int64(time.Since(start)))

	var wg sync.WaitGroup

	// start input
	cs.t.Logf("start input...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		cs.ipt.Run()
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
	cs.mCount = make(map[string]struct{})
	if err := cs.checkPoint(pts); err != nil {
		return err
	}

	cs.t.Logf("stop input...")
	cs.ipt.Terminate()

	require.GreaterOrEqual(cs.t, len(cs.mCount), 1) // At lest 1 Metric out.

	cs.t.Logf("exit...")
	wg.Wait()

	return nil
}

func (cs *caseSpec) getMappingPorts() error {
	cs.serverPorts = make([]string, len(cs.exposedPorts))
	for k, v := range cs.exposedPorts {
		mapStr := cs.resource.GetHostPort(v)
		_, port, err := net.SplitHostPort(mapStr)
		if err != nil {
			return err
		}
		cs.serverPorts[k] = port
	}
	return nil
}

func (cs *caseSpec) portsOK(r *testutils.RemoteInfo) error {
	for _, v := range cs.serverPorts {
		if !r.PortOK(docker.Port(v).Port(), time.Minute) {
			return fmt.Errorf("service checking failed")
		}
	}
	return nil
}
