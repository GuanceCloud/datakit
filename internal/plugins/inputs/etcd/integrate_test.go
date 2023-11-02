// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package etcd

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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

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

func getOptionalFields() []string {
	out := inputs.GetOptionalFieldsKeys(etcdFields)

	out = append(out, inputs.GetGolangGeneralOptionalFieldsKeys()...)

	return out
}

func getConfAccessPoint(host, port string) string {
	return fmt.Sprintf("http://%s/metrics", net.JoinHostPort(host, port))
}

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name         string
		conf         string
		exposedPorts []string
		opts         []inputs.PointCheckOption
	}{
		////////////////////////////////////////////////////////////////////////
		// etcd 3.5.7
		////////////////////////////////////////////////////////////////////////
		{
			name:         "pubrepo.guance.com/image-repo-for-testing/bitnami/etcd:3.5.7",
			conf:         `interval = "10s"`, // set conf URL later.
			exposedPorts: []string{"2379/tcp"},
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalFields(
					getOptionalFields()...,
				),
				inputs.WithOptionalTags(
					"action",
					"cluster_version",
					"code",
					"grpc_code",
					"grpc_method",
					"grpc_service",
					"grpc_type",
					"instance",
					"server_go_version",
					"server_id",
					"server_version",
					"version",
				),
				inputs.WithIgnoreTags("op", "success", "le", "quantile"),
			},
		},

		////////////////////////////////////////////////////////////////////////
		// etcd 3.4.24
		////////////////////////////////////////////////////////////////////////
		{
			name:         "pubrepo.guance.com/image-repo-for-testing/bitnami/etcd:3.4.24",
			conf:         `interval = "10s"`, // set conf URL later.
			exposedPorts: []string{"2379/tcp"},
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalFields(
					getOptionalFields()...,
				),
				inputs.WithOptionalTags(
					"action",
					"cluster_version",
					"code",
					"grpc_code",
					"grpc_method",
					"grpc_service",
					"grpc_type",
					"instance",
					"server_go_version",
					"server_id",
					"server_version",
					"version",
				),
				inputs.WithIgnoreTags("op", "success", "le", "quantile"),
			},
		},

		////////////////////////////////////////////////////////////////////////
		// etcd 3.3.27
		////////////////////////////////////////////////////////////////////////
		{
			name:         "pubrepo.guance.com/image-repo-for-testing/bitnami/etcd:3.3.27",
			conf:         `interval = "10s"`, // set conf URL later.
			exposedPorts: []string{"2379/tcp"},
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalFields(
					getOptionalFields()...,
				),
				inputs.WithOptionalTags(
					"action",
					"cluster_version",
					"code",
					"grpc_code",
					"grpc_method",
					"grpc_service",
					"grpc_type",
					"instance",
					"server_go_version",
					"server_id",
					"server_version",
					"version",
				),
				inputs.WithIgnoreTags("op", "success", "le", "quantile"),
			},
		},
	}

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := io.NewMockedFeeder()

		ipt := defaultInput()
		ipt.Feeder = feeder // Flush metric data to testing_metrics

		// URL from ENV.
		_, err := toml.Decode(base.conf, ipt)
		require.NoError(t, err)

		envs := []string{
			"ALLOW_NONE_AUTHENTICATION=yes",
		}

		repoTag := strings.Split(base.name, ":")

		cases = append(cases, &caseSpec{
			t:       t,
			ipt:     ipt,
			name:    base.name,
			feeder:  feeder,
			envs:    envs,
			repo:    repoTag[0],
			repoTag: repoTag[1],
			opts:    base.opts,

			exposedPorts: base.exposedPorts,

			// Test case result.
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

type caseSpec struct {
	t *testing.T

	name         string
	repo         string // docker name
	repoTag      string // docker tag
	envs         []string
	exposedPorts []string
	serverPorts  []string
	opts         []inputs.PointCheckOption
	mCount       map[string]struct{}

	ipt    *Input // This is real prom
	feeder *io.MockedFeeder

	pool     *dt.Pool
	resource *dt.Resource

	cr *testutils.CaseResult // collect `go test -run` metric
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		// fmt.Printf("pt = %s\n", pt.LineProto())

		var opts []inputs.PointCheckOption
		opts = append(opts, inputs.WithExtraTags(cs.ipt.Tags))

		measurement := pt.Name()

		switch measurement {
		case inputName:
			opts = append(opts, cs.opts...)
			opts = append(opts, inputs.WithDoc(&etcdMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[inputName] = struct{}{}

		default: // TODO: check other measurement
			panic("unknown measurement: " + measurement)
		}

		// check if tag appended
		if len(cs.ipt.Tags) != 0 {
			// cs.t.Logf("checking tags %+#v...", cs.ipt.Tags)

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
	cs.ipt.URLs = []string{getConfAccessPoint(r.Host, cs.serverPorts[0])} // set conf URL here.

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

	require.Equal(cs.t, 1, len(cs.mCount)) // Metric set count.

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
