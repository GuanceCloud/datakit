// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package promtail

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
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

// grafana/promtail:2.8.2-datakit --> 2.8.2
func getVersion(name string) string {
	preDefined1 := "promtail:"
	preDefined2 := "-"

	ndx1 := strings.Index(name, preDefined1)
	if ndx1 == -1 {
		return ""
	}

	name2 := name[ndx1+len(preDefined1):]
	ndx2 := strings.Index(name2, preDefined2)
	if ndx2 == -1 {
		return name2
	}

	return name2[:ndx2]
}

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name    string // Also used as build image name:tag.
		verConf BUILD_CONFIG_VER
		cmd     []string
		conf    string
		opts    []inputs.PointCheckOption
	}{
		{
			name:    "grafana/promtail:2.8.2-datakit",
			verConf: VER_3,
			cmd: []string{
				"-config.file=/etc/promtail/config.yml",
				"-config.expand-env=true",
			},
		},

		{
			name:    "pubrepo.guance.com/image-repo-for-testing/promtail:2.0.0-datakit",
			verConf: VER_2,
		},

		{
			name:    "pubrepo.guance.com/image-repo-for-testing/promtail:1.5.0-datakit",
			verConf: VER_2,
		},

		{
			name:    "pubrepo.guance.com/image-repo-for-testing/promtail:1.0.0-datakit",
			verConf: VER_2,
		},

		{
			name:    "pubrepo.guance.com/image-repo-for-testing/promtail:0.1.0-datakit",
			conf:    `legacy = true`,
			verConf: VER_1,
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

		version := getVersion(base.name)

		cases = append(cases, &caseSpec{
			t:       t,
			ipt:     ipt,
			name:    base.name,
			feeder:  feeder,
			repo:    repoTag[0],
			repoTag: repoTag[1],

			cmd:            base.cmd,
			dockerFileText: getDockerfile(version, base.verConf),
			opts:           base.opts,

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
		case measurementName:
			opts = append(opts, inputs.WithDoc(&promtailSampleMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[measurementName] = struct{}{}

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

func (cs *caseSpec) handler(c *gin.Context) {
	cs.ipt.ServeHTTP(c.Writer, c.Request)
}

func (cs *caseSpec) run() error {
	r := testutils.GetRemote()
	dockerTCP := r.TCPURL()

	cs.t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	////////////////////////////////////////////////////////////////////////////

	gin.SetMode(gin.DebugMode)
	router := gin.Default()
	router.POST("v1/write/promtail", cs.handler)

	randPort := testutils.RandPort("tcp")
	randPortStr := fmt.Sprintf("%d", randPort)
	cs.t.Logf("listening port " + randPortStr + "...")

	srv := &http.Server{
		Addr:    ":" + randPortStr,
		Handler: router,
	}

	go func() {
		cs.done = make(chan struct{})
	startListen:
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			if strings.Contains(err.Error(), "address already in use") {
				cs.t.Log(err.Error())
				goto startListen
			}
			panic(err)
		}
	}()

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil && errors.Is(err, http.ErrServerClosed) {
			cs.t.Logf("Shutdown failed: %v", err)
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

	extIP, err := testutils.ExternalIP()
	if err != nil {
		return err
	}

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

	cs.t.Logf("check service(%s:%v)...", r.Host, cs.exposedPorts)

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

////////////////////////////////////////////////////////////////////////////////

// selfBuild indicates the image was customized built.
func getDockerfile(version string, buildConfVer BUILD_CONFIG_VER) string {
	if len(version) == 0 {
		panic("version is empty")
	}

	replacePair := map[string]string{
		"VERSION":      version,
		"DATAKIT_HOST": "${DATAKIT_HOST}",
		"DATAKIT_PORT": "${DATAKIT_PORT}",
	}

	switch buildConfVer {
	case VER_1:
		return os.Expand(dockerFileV1, func(k string) string { return replacePair[k] })
	case VER_2:
		return os.Expand(dockerFileV2, func(k string) string { return replacePair[k] })
	case VER_3:
		return os.Expand(dockerFileV3, func(k string) string { return replacePair[k] })
	}

	panic("should not been here.")
}

////////////////////////////////////////////////////////////////////////////////

// Dockerfiles.

type BUILD_CONFIG_VER int

const (
	VER_1 BUILD_CONFIG_VER = iota + 1
	VER_2
	VER_3
)

// v0.1.0
/*
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://${DATAKIT_HOST}:${DATAKIT_PORT}/v1/write/promtail

scrape_configs:
- job_name: system
  static_configs:
  - targets:
      - localhost
    labels:
      job: varlogs
      __path__: /var/log/*log
*/
//nolint:lll
const dockerFileV1 = `FROM pubrepo.guance.com/image-repo-for-testing/promtail:${VERSION}

RUN    touch /var/log/1.log /var/log/2.log \
    && echo "123" >> /var/log/1.log \
    && echo "456" >> /var/log/2.log \
    && sed -i 's/http:\/\/loki:3100\/api\/prom\/push/http:\/\/\${DATAKIT_HOST}:\${DATAKIT_PORT}\/v1\/write\/promtail/' /etc/promtail/config.yml`

// v1.0.0 ~ v2.0.0
/*
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://${DATAKIT_HOST}:${DATAKIT_PORT}/v1/write/promtail

scrape_configs:
- job_name: system
  static_configs:
  - targets:
      - localhost
    labels:
      job: varlogs
      __path__: /var/log/*log
*/
//nolint:lll
const dockerFileV2 = `FROM pubrepo.guance.com/image-repo-for-testing/promtail:${VERSION}

RUN    touch /var/log/1.log /var/log/2.log \
    && echo "123" >> /var/log/1.log \
    && echo "456" >> /var/log/2.log \
    && sed -i 's/http:\/\/loki:3100\/loki\/api\/v1\/push/http:\/\/\${DATAKIT_HOST}:\${DATAKIT_PORT}\/v1\/write\/promtail/' /etc/promtail/config.yml`

// v2.8.2
/*
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://${DATAKIT_HOST}:${DATAKIT_PORT}/v1/write/promtail

scrape_configs:
- job_name: system
  static_configs:
  - targets:
      - localhost
    labels:
      job: varlogs
      __path__: /var/log/*log
*/
//nolint:lll
const dockerFileV3 = `FROM grafana/promtail:${VERSION}

RUN sed -i 's/http:\/\/loki:3100\/loki\/api\/v1\/push/http:\/\/\${DATAKIT_HOST}:\${DATAKIT_PORT}\/v1\/write\/promtail/' /etc/promtail/config.yml`
