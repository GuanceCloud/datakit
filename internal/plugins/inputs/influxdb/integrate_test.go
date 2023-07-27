// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package influxdb

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	dockertest "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

// ATTENTION: Docker version should use v20.10.18 in integrate tests. Other versions are not tested.

var mExpect = map[string]struct{}{
	influxdbCq:            {},
	influxdbHttpd:         {},
	influxdbMemstats:      {},
	influxdbQueryExecutor: {},
	influxdbRuntime:       {},
	influxdbSubscriber:    {},
	influxdbWrite:         {},
}

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
	return fmt.Sprintf("http://%s/debug/vars", net.JoinHostPort(host, port))
}

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name                      string // Also used as build image name:tag.
		conf                      string
		dockerFileText            string // Empty if not build image.
		exposedPorts              []string
		cmd                       []string
		optsInfluxdbCq            []inputs.PointCheckOption
		optsInfluxdbHttpd         []inputs.PointCheckOption
		optsInfluxdbMemstats      []inputs.PointCheckOption
		optsInfluxdbQueryExecutor []inputs.PointCheckOption
		optsInfluxdbRuntime       []inputs.PointCheckOption
		optsInfluxdbSubscriber    []inputs.PointCheckOption
		optsInfluxdbWrite         []inputs.PointCheckOption
	}{
		////////////////////////////////////////////////////////////////////////
		// influxdb:1.8.10
		////////////////////////////////////////////////////////////////////////
		{
			name: "influxdb:1.8.10-alpine",
			conf: `url = ""
			interval = '2s'
			timeout = "5s"
			election = true`, // set conf URL later.
			exposedPorts: []string{"8086/tcp"},
			optsInfluxdbCq: []inputs.PointCheckOption{
				inputs.WithExtraTags(map[string]string{
					"election": "1",
				}),
			},
			optsInfluxdbHttpd: []inputs.PointCheckOption{
				inputs.WithExtraTags(map[string]string{
					"election": "1",
				}),
			},
			optsInfluxdbMemstats: []inputs.PointCheckOption{
				inputs.WithExtraTags(map[string]string{
					"election": "1",
				}),
			},
			optsInfluxdbQueryExecutor: []inputs.PointCheckOption{
				inputs.WithExtraTags(map[string]string{
					"election": "1",
				}),
			},
			optsInfluxdbRuntime: []inputs.PointCheckOption{
				inputs.WithExtraTags(map[string]string{
					"election": "1",
				}),
			},
			optsInfluxdbSubscriber: []inputs.PointCheckOption{
				inputs.WithExtraTags(map[string]string{
					"election": "1",
				}),
			},
			optsInfluxdbWrite: []inputs.PointCheckOption{
				inputs.WithExtraTags(map[string]string{
					"election": "1",
				}),
			},
		},
		{
			name: "influxdb:1.8.10-alpine",
			conf: `url = ""
			interval = '2s'
			timeout = "5s"
			election = false`, // set conf URL later.
			exposedPorts: []string{"8086/tcp"},
		},
	}

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := io.NewMockedFeeder()

		ipt := defaultInput()
		ipt.feeder = feeder

		_, err := toml.Decode(base.conf, ipt)
		require.NoError(t, err)

		if ipt.Election {
			ipt.Tagger = testutils.NewTaggerElection()
		} else {
			ipt.Tagger = testutils.NewTaggerHost()
		}

		repoTag := strings.Split(base.name, ":")

		cases = append(cases, &caseSpec{
			t:       t,
			ipt:     ipt,
			name:    base.name,
			feeder:  feeder,
			repo:    repoTag[0],
			repoTag: repoTag[1],

			dockerFileText: base.dockerFileText,
			exposedPorts:   base.exposedPorts,
			cmd:            base.cmd,

			optsInfluxdbCq:            base.optsInfluxdbCq,
			optsInfluxdbHttpd:         base.optsInfluxdbHttpd,
			optsInfluxdbMemstats:      base.optsInfluxdbMemstats,
			optsInfluxdbQueryExecutor: base.optsInfluxdbQueryExecutor,
			optsInfluxdbRuntime:       base.optsInfluxdbRuntime,
			optsInfluxdbSubscriber:    base.optsInfluxdbSubscriber,
			optsInfluxdbWrite:         base.optsInfluxdbWrite,

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

	name                      string
	repo                      string
	repoTag                   string
	dockerFileText            string
	exposedPorts              []string
	serverPorts               []string
	optsInfluxdbCq            []inputs.PointCheckOption
	optsInfluxdbHttpd         []inputs.PointCheckOption
	optsInfluxdbMemstats      []inputs.PointCheckOption
	optsInfluxdbQueryExecutor []inputs.PointCheckOption
	optsInfluxdbRuntime       []inputs.PointCheckOption
	optsInfluxdbSubscriber    []inputs.PointCheckOption
	optsInfluxdbWrite         []inputs.PointCheckOption
	cmd                       []string
	mCount                    map[string]struct{}

	ipt    *Input
	feeder *io.MockedFeeder

	pool     *dockertest.Pool
	resource *dockertest.Resource

	cr *testutils.CaseResult
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		var opts []inputs.PointCheckOption
		opts = append(opts, inputs.WithExtraTags(cs.ipt.Tags))

		measurement := string(pt.Name())

		switch measurement {
		case influxdbCq:
			opts = append(opts, cs.optsInfluxdbCq...)
			opts = append(opts, inputs.WithDoc(&InfluxdbCqM{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[influxdbCq] = struct{}{}

		case influxdbHttpd:
			opts = append(opts, cs.optsInfluxdbHttpd...)
			opts = append(opts, inputs.WithDoc(&InfluxdbHttpdM{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[influxdbHttpd] = struct{}{}

		case influxdbMemstats:
			opts = append(opts, cs.optsInfluxdbMemstats...)
			opts = append(opts, inputs.WithDoc(&InfluxdbMemstatsM{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[influxdbMemstats] = struct{}{}

		case influxdbQueryExecutor:
			opts = append(opts, cs.optsInfluxdbQueryExecutor...)
			opts = append(opts, inputs.WithDoc(&InfluxdbQueryExecutorM{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[influxdbQueryExecutor] = struct{}{}

		case influxdbRuntime:
			opts = append(opts, cs.optsInfluxdbRuntime...)
			opts = append(opts, inputs.WithDoc(&InfluxdbRuntimeM{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[influxdbRuntime] = struct{}{}

		case influxdbSubscriber:
			opts = append(opts, cs.optsInfluxdbSubscriber...)
			opts = append(opts, inputs.WithDoc(&InfluxdbSubscriberM{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[influxdbSubscriber] = struct{}{}

		case influxdbWrite:
			opts = append(opts, cs.optsInfluxdbWrite...)
			opts = append(opts, inputs.WithDoc(&InfluxdbWriteM{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[influxdbWrite] = struct{}{}

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
	r := testutils.GetRemote()
	dockerTCP := r.TCPURL()

	cs.t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	start := time.Now()

	p, err := cs.getPool(dockerTCP)
	if err != nil {
		return err
	}

	dockerFileDir, dockerFilePath, err := cs.getDockerFilePath()
	if err != nil {
		return err
	}
	defer os.RemoveAll(dockerFileDir)

	uniqueContainerName := testutils.GetUniqueContainerName(inputName)

	var resource *dockertest.Resource

	if len(cs.dockerFileText) == 0 {
		// Just run a container from existing docker image.
		resource, err = p.RunWithOptions(
			&dockertest.RunOptions{
				Name: uniqueContainerName, // ATTENTION: not cs.name.

				Repository: cs.repo,
				Tag:        cs.repoTag,
				Cmd:        cs.cmd,

				ExposedPorts: cs.exposedPorts,
			},

			func(c *docker.HostConfig) {
				c.RestartPolicy = docker.RestartPolicy{Name: "no"}
				c.AutoRemove = true
			},
		)
	} else {
		// Build docker image from Dockerfile and run a container from it.
		resource, err = p.BuildAndRunWithOptions(
			dockerFilePath,

			&dockertest.RunOptions{
				ContainerName: uniqueContainerName,
				Name:          cs.name, // ATTENTION: not uniqueContainerName.

				Repository: cs.repo,
				Tag:        cs.repoTag,
				Cmd:        cs.cmd,

				ExposedPorts: cs.exposedPorts,
			},

			func(c *docker.HostConfig) {
				c.RestartPolicy = docker.RestartPolicy{Name: "no"}
				c.AutoRemove = true
			},
		)
	}

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
	pts, err := cs.feeder.NPoints(10, 5*time.Minute)
	if err != nil {
		return err
	}

	cs.cr.AddField("point_latency", int64(time.Since(start)))
	cs.cr.AddField("point_count", len(pts))

	// for k, v := range pts {
	// 	cs.t.Logf("pt [%02d] = %s", k, v.LineProto())
	// }

	cs.t.Logf("get %d points", len(pts))
	cs.mCount = make(map[string]struct{})
	if err := cs.checkPoint(pts); err != nil {
		return err
	}

	cs.t.Logf("stop input...")
	cs.ipt.Terminate()

	require.Equal(cs.t, mExpect, cs.mCount)

	cs.t.Logf("exit...")
	wg.Wait()

	return nil
}

func (cs *caseSpec) getPool(endpoint string) (*dockertest.Pool, error) {
	p, err := dockertest.NewPool(endpoint)
	if err != nil {
		return nil, err
	}
	err = p.Client.Ping()
	if err != nil {
		cs.t.Logf("Could not connect to Docker: %v", err)
		return nil, err
	}
	return p, nil
}

func (cs *caseSpec) getDockerFilePath() (dirName string, fileName string, err error) {
	if len(cs.dockerFileText) == 0 {
		return
	}

	tmpDir, err := ioutil.TempDir("", "dockerfiles_")
	if err != nil {
		cs.t.Logf("ioutil.TempDir failed: %s", err.Error())
		return "", "", err
	}

	tmpFile, err := ioutil.TempFile(tmpDir, "dockerfile_")
	if err != nil {
		cs.t.Logf("ioutil.TempFile failed: %s", err.Error())
		return "", "", err
	}

	_, err = tmpFile.WriteString(cs.dockerFileText)
	if err != nil {
		cs.t.Logf("TempFile.WriteString failed: %s", err.Error())
		return "", "", err
	}

	if err := os.Chmod(tmpFile.Name(), os.ModePerm); err != nil {
		cs.t.Logf("os.Chmod failed: %s", err.Error())
		return "", "", err
	}

	if err := tmpFile.Close(); err != nil {
		cs.t.Logf("Close failed: %s", err.Error())
		return "", "", err
	}

	return tmpDir, tmpFile.Name(), nil
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

////////////////////////////////////////////////////////////////////////////////
