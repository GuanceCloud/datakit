// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jaeger

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/gin-gonic/gin"
	dockertest "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
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
				// t.Parallel() // jaeger should not be paralleled due to code design. For example, afterGatherRun is global variable.
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

type AGENT_MODE int

const (
	AGENT_HTTP AGENT_MODE = iota + 1
	AGENT_UDP
)

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name string // Also used as build image name:tag.
		conf string
		mode AGENT_MODE
		opts []inputs.PointCheckOption
	}{
		{
			name: "pubrepo.guance.com/image-repo-for-testing/jaeger:jaeger-agent-http",
			conf: `endpoint = "/apis/traces"`,
			mode: AGENT_HTTP,
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalTags(
					itrace.TAG_ENV,
					itrace.TAG_PROJECT,
					itrace.VERSION,
					itrace.TAG_HTTP_STATUS_CODE,
					itrace.TAG_HTTP_METHOD,
					itrace.TAG_CONTAINER_HOST,
					itrace.TAG_ENDPOINT,
					itrace.TAG_HTTP_ROUTE,
					itrace.TAG_HTTP_URL,
				),
				inputs.WithOptionalFields(
					itrace.TAG_PID,
					itrace.FIELD_PRIORITY,
				),
				inputs.WithIgnoreTags(
					testutils.RUNTIME_ID,
				),
				inputs.WithIgnoreUnexpectedTags(true),
			},
		},

		{
			name: "pubrepo.guance.com/image-repo-for-testing/jaeger:jaeger-agent-udp",
			conf: "", // set conf later.
			mode: AGENT_UDP,
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalTags(
					itrace.TAG_ENV,
					itrace.TAG_PROJECT,
					itrace.VERSION,
					itrace.TAG_HTTP_STATUS_CODE,
					itrace.TAG_HTTP_METHOD,
					itrace.TAG_CONTAINER_HOST,
					itrace.TAG_ENDPOINT,
					itrace.TAG_HTTP_ROUTE,
					itrace.TAG_HTTP_URL,
				),
				inputs.WithOptionalFields(
					itrace.TAG_PID,
					itrace.FIELD_PRIORITY,
				),
				inputs.WithIgnoreTags(
					testutils.RUNTIME_ID,
				),
				inputs.WithIgnoreUnexpectedTags(true),
			},
		},
	}

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := io.NewMockedFeeder()

		ipt := defaultInput()
		ipt.feeder = feeder
		ipt.Tagger = testutils.NewTaggerHost()

		_, err := toml.Decode(base.conf, ipt)
		require.NoError(t, err)

		repoTag := strings.Split(base.name, ":")

		cases = append(cases, &caseSpec{
			t:       t,
			ipt:     ipt,
			name:    base.name,
			feeder:  feeder,
			repo:    repoTag[0],
			repoTag: repoTag[1],

			mode: base.mode,
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

	name           string
	repo           string
	repoTag        string
	dockerFileText string
	exposedPorts   []string
	opts           []inputs.PointCheckOption
	mode           AGENT_MODE
	mCount         map[string]struct{}
	done           chan struct{}
	cmd            []string

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
		opts = append(opts, cs.opts...)

		measurement := pt.Name()

		switch measurement {
		case inputName:
			opts = append(opts, inputs.WithDoc(&itrace.TraceMeasurement{Name: inputName}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[inputName] = struct{}{}

		default: // TODO: check other measurement
			panic("unknown measurement")
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

func (cs *caseSpec) handler(c *gin.Context) {
	handleJaegerTrace(c.Writer, c.Request)
}

func (cs *caseSpec) run() error {
	r := testutils.GetRemote()
	dockerTCP := r.TCPURL()

	cs.t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	extIP, err := testutils.ExternalIP()
	if err != nil {
		return err
	}

	////////////////////////////////////////////////////////////////////////////

	cs.ipt.RegHTTPHandler()

	var (
		srv         *http.Server
		randPortStr string
	)

	switch cs.mode {
	case AGENT_HTTP:
		gin.SetMode(gin.DebugMode)
		router := gin.Default()
		router.POST("apis/traces", cs.handler)

		var listener net.Listener

		for {
			randPort := testutils.RandPort("tcp")
			randPortStr = fmt.Sprintf("%d", randPort)
			listener, err = net.Listen("tcp", ":"+randPortStr)
			if err != nil {
				if strings.Contains(err.Error(), "bind: address already in use") {
					continue
				}
				cs.t.Logf("net.Listen failed: %v", err)
				return err
			}
			break
		}

		cs.t.Logf("listening port " + randPortStr + "...")

		srv = &http.Server{Handler: router}

		go func() {
			cs.done = make(chan struct{})
			if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
				panic(err)
			}
		}()

	case AGENT_UDP:
		conn, randPort, err := testutils.RandPortUDP()
		require.NoError(cs.t, err)
		randPortStr = fmt.Sprintf("%d", randPort)
		cs.ipt.Address = extIP + ":" + randPortStr
		cs.ipt.udpListener = conn
	}

	shutdownFunc := func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil && errors.Is(err, http.ErrServerClosed) {
			cs.t.Logf("Shutdown failed: %v", err)
		}
	}

	defer func() {
		//nolint:gocritic,exhaustive
		switch cs.mode {
		case AGENT_HTTP:
			shutdownFunc()
		}
	}()

	////////////////////////////////////////////////////////////////////////////

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
				Env: []string{
					"DATAKIT_HOST=" + extIP,
					"DATAKIT_PORT=" + randPortStr,
				},
				Cmd: cs.cmd,
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
				Env: []string{
					"DATAKIT_HOST=" + extIP,
					"DATAKIT_PORT=" + randPortStr,
				},
				Cmd: cs.cmd,
			},

			func(c *docker.HostConfig) {
				c.RestartPolicy = docker.RestartPolicy{Name: "no"}
				c.AutoRemove = true
			},
		)
	}

	if err != nil {
		cs.t.Logf("%s", err.Error())
		return err
	}

	cs.pool = p
	cs.resource = resource

	switch cs.mode {
	case AGENT_HTTP:
		cs.t.Logf("check service(%s:tcp:%s)...", r.Host, randPortStr)
	case AGENT_UDP:
		cs.t.Logf("check service(%s:udp:%s)...", r.Host, randPortStr)
	}

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
	pts, err := cs.feeder.AnyPoints(5 * time.Minute)
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

	require.Equal(cs.t, 1, len(cs.mCount))

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

func (cs *caseSpec) portsOK(r *testutils.RemoteInfo) error {
	for _, v := range cs.exposedPorts {
		if !r.PortOK(docker.Port(v).Port(), time.Minute) {
			return fmt.Errorf("service checking failed")
		}
	}
	return nil
}
