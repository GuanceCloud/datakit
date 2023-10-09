// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package consul

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

		opts []inputs.PointCheckOption
	}{
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/consul/consul:1.15.0",
			// selfBuild: true,
			conf: `
source = "consul"
metric_name_filter = ["consul_raft_leader", "consul_raft_peers", "consul_serf_lan_members", "consul_catalog_service", "consul_catalog_service_node_healthy", "consul_health_node_status", "consul_serf_lan_member_status"]
tags_ignore = ["check"]
interval = "10s"
url = ""

[tags]
  tag1 = "some_value"
  tag2 = "some_other_value"`, // set conf URL later.
			exposedPorts: []string{"9107/tcp"},
			mPathCount:   map[string]int{"/": 10},
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalTags("host", "check", "check_id", "check_name", "node", "tag", "key", "service_id", "service_name", "status", "member", "instance"),
				inputs.WithOptionalFields("up", "raft_peers", "raft_leader", "serf_lan_members", "serf_lan_member_status", "serf_wan_member_status", "catalog_services", "service_tag", "catalog_service_node_healthy", "health_node_status", "health_service_status", "service_checks", "catalog_kv"), // nolint:lll
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"tag1": "", "tag2": ""}),
			},
		},
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/consul/consul:1.14.4",
			// selfBuild: true,
			conf: `
source = "consul"
metric_name_filter = ["consul_raft_leader", "consul_raft_peers", "consul_serf_lan_members", "consul_catalog_service", "consul_catalog_service_node_healthy", "consul_health_node_status", "consul_serf_lan_member_status"]
tags_ignore = ["check"]
interval = "10s"
url = ""`, // set conf URL later.
			exposedPorts: []string{"9107/tcp"},
			mPathCount:   map[string]int{"/": 10},
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalTags("host", "check", "check_id", "check_name", "node", "tag", "key", "service_id", "service_name", "status", "member", "instance"),
				inputs.WithOptionalFields("up", "raft_peers", "raft_leader", "serf_lan_members", "serf_lan_member_status", "serf_wan_member_status", "catalog_services", "service_tag", "catalog_service_node_healthy", "health_node_status", "health_service_status", "service_checks", "catalog_kv"), // nolint:lll
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{}),
			},
		},
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/consul/consul:1.13.6",
			// selfBuild: true,
			conf: `
source = "consul"
metric_name_filter = ["consul_raft_leader", "consul_raft_peers", "consul_serf_lan_members", "consul_catalog_service", "consul_catalog_service_node_healthy", "consul_health_node_status", "consul_serf_lan_member_status"]
tags_ignore = ["check"]
interval = "10s"
url = ""

[tags]
  tag1 = "some_value"
  tag2 = "some_other_value"`, // set conf URL later.
			exposedPorts: []string{"9107/tcp"},
			mPathCount:   map[string]int{"/": 10},
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalTags("host", "check", "check_id", "check_name", "node", "tag", "key", "service_id", "service_name", "status", "member", "instance"),
				inputs.WithOptionalFields("up", "raft_peers", "raft_leader", "serf_lan_members", "serf_lan_member_status", "serf_wan_member_status", "catalog_services", "service_tag", "catalog_service_node_healthy", "health_node_status", "health_service_status", "service_checks", "catalog_kv"), // nolint:lll
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"tag1": "", "tag2": ""}),
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

			opts: base.opts,

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

	opts []inputs.PointCheckOption

	ipt    *prom.Input
	feeder *io.MockedFeeder

	pool     *dt.Pool
	resource *dt.Resource

	cr *testutils.CaseResult
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		measurement := string(pt.Name())
		opts := []inputs.PointCheckOption{}

		switch measurement {
		case "consul":
			opts = append(opts, cs.opts...)
			opts = append(opts, inputs.WithDoc(&ConsulMeasurement{}))

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
	r := testutils.GetRemote()
	dockerTCP := r.TCPURL()

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
	cs.t.Logf("get len(pts)...%d", len(pts))
	for _, v := range pts {
		cs.t.Logf("get v.LineProto()...%s", v.LineProto())
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
