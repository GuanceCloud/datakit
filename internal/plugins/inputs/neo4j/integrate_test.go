// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package neo4j

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
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

func TestIntegrate(t *testing.T) {
	if !testutils.CheckIntegrationTestingRunning() {
		t.Skip()
	}

	testutils.PurgeRemoteByName(inputName)
	defer testutils.PurgeRemoteByName(inputName)

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
		name         string
		conf         string
		exposedPorts []string
		opts         []inputs.PointCheckOption
	}{
		////////////////////////////////////////////////////////////////////////
		// neo4j 5.11.0
		////////////////////////////////////////////////////////////////////////
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/neo4j:5.11.0-enterprise-prom",
			conf: `interval = "10s"
		`,
			exposedPorts: []string{"2004/tcp"},
			opts: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalFields(mergeSlice([]string{}, getOptionalFields())...),
				inputs.WithOptionalTags(mergeSlice([]string{}, getOptionalTags())...),
			},
		},
		////////////////////////////////////////////////////////////////////////
		// neo4j 4.4.0
		////////////////////////////////////////////////////////////////////////
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/neo4j:4.4.0-enterprise-prom",
			conf: `interval = "10s"
election = true
[tags]
  tag1 = "some_value"
  tag2 = "some_other_value"
		`,
			exposedPorts: []string{"2004/tcp"},
			opts: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalFields(mergeSlice([]string{}, getOptionalFields())...),
				inputs.WithOptionalTags(mergeSlice([]string{}, getOptionalTags())...),
			},
		},
		////////////////////////////////////////////////////////////////////////
		// neo4j 3.4.0
		////////////////////////////////////////////////////////////////////////
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/neo4j:3.4.0-enterprise-prom",
			conf: `interval = "10s"
disable_host_tag = true
disable_instance_tag = true
		`,
			exposedPorts: []string{"2004/tcp"},
			opts: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithOptionalFields(mergeSlice([]string{}, getOptionalFields())...),
				inputs.WithOptionalTags(mergeSlice([]string{}, getOptionalTags())...),
			},
		},
	}

	var cases []*caseSpec

	for _, base := range bases {
		feeder := io.NewMockedFeeder()

		ipt := NewInput()
		ipt.Feeder = feeder

		// URL from ENV.
		_, err := toml.Decode(base.conf, ipt)
		require.NoError(t, err)

		envs := []string{
			"NEO4J_ACCEPT_LICENSE_AGREEMENT=yes",
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

	ipt    *Input
	feeder *io.MockedFeeder

	pool     *dt.Pool
	resource *dt.Resource

	cr *testutils.CaseResult // collect `go test -run` metric
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		// fmt.Printf("pt = %s\n", pt.LineProto())
		abc := getOptionalTags()
		_ = abc
		var opts []inputs.PointCheckOption
		opts = append(opts, inputs.WithExtraTags(cs.ipt.Tags))

		measurement := string(pt.Name())

		switch measurement {
		case inputName:
			opts = append(opts, cs.opts...)
			opts = append(opts, inputs.WithDoc(&TestMeasurement{}))

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
	cs.ipt.URLs = []string{getConfAccessPoint(r.Host, cs.serverPorts[0])}

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

// Testing measurement

type TestMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
}

func (t *TestMeasurement) LineProto() (*dkpt.Point, error) {
	return dkpt.NewPoint(t.name, t.tags, t.fields, dkpt.MOptElection())
}

func (t *TestMeasurement) Info() *inputs.MeasurementInfo {
	measurement := &inputs.MeasurementInfo{
		Name:   "neo4j",
		Fields: getTestFields(),
		Tags:   getTestTags(),
	}

	return measurement
}

func getTestFields() map[string]interface{} {
	fields := map[string]interface{}{}
	for _, field := range fieldNames {
		fields[field.name] = field.field
	}

	return fields
}

func getTestTags() map[string]interface{} {
	m := &Measurement{}
	tags := map[string]interface{}{}
	for k, v := range m.Info().Tags {
		tags[k] = v
	}

	return tags
}

func mergeSlice(a, b []string) []string {
	m := make(map[string]bool)
	for _, v := range a {
		m[v] = true
	}
	for _, v := range b {
		m[v] = true
	}
	c := make([]string, 0)
	for k := range m {
		c = append(c, k)
	}
	return c
}

func getOptionalFields() []string {
	m := &TestMeasurement{}
	info := m.Info()
	_ = info
	s := make([]string, 0)
	for k := range m.Info().Fields {
		s = append(s, k)
	}

	return s
}

func getOptionalTags() []string {
	m := &Measurement{}
	s := make([]string, 0)
	for k := range m.Info().Tags {
		s = append(s, k)
	}
	return s
}
