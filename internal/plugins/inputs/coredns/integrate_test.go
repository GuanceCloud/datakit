// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package coredns

import (
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"os"
	"sync"
	T "testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	dt "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/prom"
	tu "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

type caseSpec struct {
	t *T.T

	name        string
	repo        string // docker name
	repoTag     string // docker tag
	envs        []string
	servicePort string // port (rand)）

	optsACL      []inputs.PointCheckOption
	optsCache    []inputs.PointCheckOption
	optsDNSSec   []inputs.PointCheckOption
	optsForward  []inputs.PointCheckOption
	optsGrpc     []inputs.PointCheckOption
	optsHosts    []inputs.PointCheckOption
	optsTemplate []inputs.PointCheckOption
	optsProm     []inputs.PointCheckOption

	ipt    *prom.Input // This is real prom
	feeder *io.MockedFeeder

	pool     *dt.Pool
	resource *dt.Resource

	cr *tu.CaseResult // collect `go test -run` metric
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		measurement := string(pt.Name())

		switch measurement {
		case "coredns_acl":
			msgs := inputs.CheckPoint(pt, inputs.WithDoc(&ACLMeasurement{}), inputs.WithExtraTags(cs.ipt.Tags))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}
		case "coredns_cache":
			msgs := inputs.CheckPoint(pt, inputs.WithDoc(&CacheMeasurement{}), inputs.WithExtraTags(cs.ipt.Tags))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}
		case "coredns_dnssec":
			msgs := inputs.CheckPoint(pt, inputs.WithDoc(&DNSSecMeasurement{}), inputs.WithExtraTags(cs.ipt.Tags))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}
		case "coredns_forward":
			msgs := inputs.CheckPoint(pt, inputs.WithDoc(&ForwardMeasurement{}), inputs.WithExtraTags(cs.ipt.Tags))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}
		case "coredns_grpc":
			msgs := inputs.CheckPoint(pt, inputs.WithDoc(&GrpcMeasurement{}), inputs.WithExtraTags(cs.ipt.Tags))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}
		case "coredns_hosts":
			msgs := inputs.CheckPoint(pt, inputs.WithDoc(&HostsMeasurement{}), inputs.WithExtraTags(cs.ipt.Tags))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}
		case "coredns_template":
			msgs := inputs.CheckPoint(pt, inputs.WithDoc(&TemplateMeasurement{}), inputs.WithExtraTags(cs.ipt.Tags))

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}
		case "coredns":
			var opts []inputs.PointCheckOption
			opts = append(opts, cs.optsProm...)
			opts = append(opts, inputs.WithDoc(&PromMeasurement{}))
			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}
		default: // TODO: check other measurement
			return nil
		}

		// check if tag appended
		if len(cs.ipt.Tags) != 0 {
			cs.t.Logf("checking tags %+#v...", cs.ipt.Tags)

			tags := pt.Tags()
			for k, expect := range cs.ipt.Tags {
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
	// start remote image server
	r := tu.GetRemote()
	dockerTCP := r.TCPURL() // got "tcp://" + net.JoinHostPort(i.Host, i.Port) 2375

	cs.t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	start := time.Now()

	p, err := dt.NewPool(dockerTCP)
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

	resource, err := p.RunWithOptions(&dt.RunOptions{
		// specify container image & tag
		Repository: cs.repo,
		Tag:        cs.repoTag,

		// port binding
		PortBindings: map[docker.Port][]docker.PortBinding{
			"9153/tcp": {{HostIP: "0.0.0.0", HostPort: cs.servicePort}},
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
	if err := cs.checkPoint(pts); err != nil {
		return err
	}

	cs.t.Logf("stop input...")
	cs.ipt.Terminate()

	cs.t.Logf("exit...")
	wg.Wait()

	return nil
}

func buildCases(t *T.T) ([]*caseSpec, error) {
	t.Helper()

	remote := tu.GetRemote()

	bases := []struct {
		name         string
		repo         string // docker name
		repoTag      string // docker tag
		conf         string
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
			name:    "remote-coredns",
			repo:    "pubrepo.jiagouyun.com/image-repo-for-testing/coredns/coredns",
			repoTag: "1.10.1",
			conf: fmt.Sprintf(`
source = "coredns"
metric_types = ["counter", "gauge"]
metric_name_filter = ["^coredns_(acl|cache|dnssec|forward|grpc|hosts|template|dns)_([a-z_]+)$"]
interval = "10s"
tls_open = false
url = "http://%s/metrics"
[[inputs.prom.measurements]]
  prefix = "coredns_acl_"
  name = "coredns_acl"
[[inputs.prom.measurements]]
  prefix = "coredns_cache_"
  name = "coredns_cache"
[[inputs.prom.measurements]]
  prefix = "coredns_dnssec_"
  name = "coredns_dnssec"
[[inputs.prom.measurements]]
  prefix = "coredns_forward_"
  name = "coredns_forward"
[[inputs.prom.measurements]]
  prefix = "coredns_grpc_"
  name = "coredns_grpc"
[[inputs.prom.measurements]]
  prefix = "coredns_hosts_"
  name = "coredns_hosts"
[[inputs.prom.measurements]]
  prefix = "coredns_template_"
  name = "coredns_template"
[[inputs.prom.measurements]]
  prefix = "coredns_dns_"
  name = "coredns"
[tags]
  tag1 = "some_value"
  tag2 = "some_other_value"`, net.JoinHostPort(remote.Host, fmt.Sprintf("%d", tu.RandPort("tcp")))),
			optsACL:      []inputs.PointCheckOption{},
			optsCache:    []inputs.PointCheckOption{},
			optsDNSSec:   []inputs.PointCheckOption{},
			optsForward:  []inputs.PointCheckOption{},
			optsGrpc:     []inputs.PointCheckOption{},
			optsHosts:    []inputs.PointCheckOption{},
			optsTemplate: []inputs.PointCheckOption{},
			optsProm: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"instance": "", "tag1": "", "tag2": ""}),
				inputs.WithOptionalTags("server", "zone", "type", "proto", "family", "rcode"),
				inputs.WithOptionalFields("dns_requests_total", "dns_request_duration_seconds", "dns_request_size_bytes", "dns_responses_total", "dns_response_size_bytes", "hosts_reload_timestamp_seconds", "forward_healthcheck_broken_total", "forward_max_concurrent_rejects_total"), // nolint:lll
			},
		},
		{
			name:    "remote-coredns",
			repo:    "pubrepo.jiagouyun.com/image-repo-for-testing/coredns/coredns",
			repoTag: "1.9.4",
			conf: fmt.Sprintf(`
source = "coredns"
metric_types = ["counter", "gauge"]
metric_name_filter = ["^coredns_(acl|cache|dnssec|forward|grpc|hosts|template|dns)_([a-z_]+)$"]
interval = "10s"
tls_open = false
url = "http://%s/metrics"
[[inputs.prom.measurements]]
  prefix = "coredns_acl_"
  name = "coredns_acl"
[[inputs.prom.measurements]]
  prefix = "coredns_cache_"
  name = "coredns_cache"
[[inputs.prom.measurements]]
  prefix = "coredns_dnssec_"
  name = "coredns_dnssec"
[[inputs.prom.measurements]]
  prefix = "coredns_forward_"
  name = "coredns_forward"
[[inputs.prom.measurements]]
  prefix = "coredns_grpc_"
  name = "coredns_grpc"
[[inputs.prom.measurements]]
  prefix = "coredns_hosts_"
  name = "coredns_hosts"
[[inputs.prom.measurements]]
  prefix = "coredns_template_"
  name = "coredns_template"
[[inputs.prom.measurements]]
  prefix = "coredns_dns_"
  name = "coredns"
[tags]
  tag1 = "some_value"
  tag2 = "some_other_value"`, net.JoinHostPort(remote.Host, fmt.Sprintf("%d", tu.RandPort("tcp")))),
			optsACL:      []inputs.PointCheckOption{},
			optsCache:    []inputs.PointCheckOption{},
			optsDNSSec:   []inputs.PointCheckOption{},
			optsForward:  []inputs.PointCheckOption{},
			optsGrpc:     []inputs.PointCheckOption{},
			optsHosts:    []inputs.PointCheckOption{},
			optsTemplate: []inputs.PointCheckOption{},
			optsProm: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"instance": "", "tag1": "", "tag2": ""}),
				inputs.WithOptionalTags("server", "zone", "type", "proto", "family", "rcode"),
				inputs.WithOptionalFields("dns_requests_total", "dns_request_duration_seconds", "dns_request_size_bytes", "dns_responses_total", "dns_response_size_bytes", "hosts_reload_timestamp_seconds", "forward_healthcheck_broken_total", "forward_max_concurrent_rejects_total"), // nolint:lll
			},
		},
		{
			name:    "remote-coredns",
			repo:    "pubrepo.jiagouyun.com/image-repo-for-testing/coredns/coredns",
			repoTag: "1.8.7",
			conf: fmt.Sprintf(`
source = "coredns"
metric_types = ["counter", "gauge"]
metric_name_filter = ["^coredns_(acl|cache|dnssec|forward|grpc|hosts|template|dns)_([a-z_]+)$"]
interval = "10s"
tls_open = false
url = "http://%s/metrics"
[[inputs.prom.measurements]]
  prefix = "coredns_acl_"
  name = "coredns_acl"
[[inputs.prom.measurements]]
  prefix = "coredns_cache_"
  name = "coredns_cache"
[[inputs.prom.measurements]]
  prefix = "coredns_dnssec_"
  name = "coredns_dnssec"
[[inputs.prom.measurements]]
  prefix = "coredns_forward_"
  name = "coredns_forward"
[[inputs.prom.measurements]]
  prefix = "coredns_grpc_"
  name = "coredns_grpc"
[[inputs.prom.measurements]]
  prefix = "coredns_hosts_"
  name = "coredns_hosts"
[[inputs.prom.measurements]]
  prefix = "coredns_template_"
  name = "coredns_template"
[[inputs.prom.measurements]]
  prefix = "coredns_dns_"
  name = "coredns"
[tags]
  tag1 = "some_value"
  tag2 = "some_other_value"`, net.JoinHostPort(remote.Host, fmt.Sprintf("%d", tu.RandPort("tcp")))),
			optsACL:      []inputs.PointCheckOption{},
			optsCache:    []inputs.PointCheckOption{},
			optsDNSSec:   []inputs.PointCheckOption{},
			optsForward:  []inputs.PointCheckOption{},
			optsGrpc:     []inputs.PointCheckOption{},
			optsHosts:    []inputs.PointCheckOption{},
			optsTemplate: []inputs.PointCheckOption{},
			optsProm: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"instance": "", "tag1": "", "tag2": ""}),
				inputs.WithOptionalTags("server", "zone", "type", "proto", "family", "rcode"),
				inputs.WithOptionalFields("dns_requests_total", "dns_request_duration_seconds", "dns_request_size_bytes", "dns_responses_total", "dns_response_size_bytes", "hosts_reload_timestamp_seconds", "forward_healthcheck_broken_total", "forward_max_concurrent_rejects_total"), // nolint:lll
			},
		},
	}

	// TODO: add per-image configs
	perImageCfgs := []interface{}{}
	_ = perImageCfgs

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := io.NewMockedFeeder()

		ipt := prom.NewProm() // This is real prom
		ipt.Feeder = feeder   // Flush metric data to testing_metrics

		// URL from ENV.
		_, err := toml.Decode(base.conf, ipt)
		assert.NoError(t, err)

		url, err := url.Parse(ipt.URL) // http://127.0.0.1:9153/metric --> 127.0.0.1:9153
		assert.NoError(t, err)

		ipport, err := netip.ParseAddrPort(url.Host)
		assert.NoError(t, err, "parse %s failed: %s", ipt.URL, err)

		cases = append(cases, &caseSpec{
			t:      t,
			ipt:    ipt,
			name:   base.name,
			feeder: feeder,
			// envs:   envs,

			repo:    base.repo,    // docker name
			repoTag: base.repoTag, // docker tag

			servicePort: fmt.Sprintf("%d", ipport.Port()),

			optsACL:      base.optsACL,
			optsCache:    base.optsCache,
			optsDNSSec:   base.optsDNSSec,
			optsForward:  base.optsForward,
			optsGrpc:     base.optsGrpc,
			optsHosts:    base.optsHosts,
			optsTemplate: base.optsTemplate,
			optsProm:     base.optsProm,

			// Test case result.
			cr: &tu.CaseResult{
				Name:        t.Name(),
				Case:        base.name,
				ExtraFields: map[string]any{},
				ExtraTags: map[string]string{
					"image":         base.repo,
					"image_tag":     base.repoTag,
					"remote_server": ipt.URL,
				},
			},
		})
	}
	return cases, nil
}

func TestCoreDnsInput(t *T.T) {
	if !tu.CheckIntegrationTestingRunning() {
		t.Skip()
	}
	start := time.Now()
	cases, err := buildCases(t)
	if err != nil {
		cr := &tu.CaseResult{
			Name:          t.Name(),
			Status:        tu.TestPassed,
			FailedMessage: err.Error(),
			Cost:          time.Since(start),
		}

		_ = tu.Flush(cr)
		return
	}

	t.Logf("testing %d cases...", len(cases))

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			caseStart := time.Now()

			t.Logf("testing %s...", tc.name)

			// Run a test case.
			if err := tc.run(); err != nil {
				tc.cr.Status = tu.TestFailed
				tc.cr.FailedMessage = err.Error()

				assert.NoError(t, err)
			} else {
				tc.cr.Status = tu.TestPassed
			}

			tc.cr.Cost = time.Since(caseStart)

			assert.NoError(t, tu.Flush(tc.cr))

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
