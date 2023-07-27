// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	lp "github.com/GuanceCloud/cliutils/lineproto"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/gin-gonic/gin"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
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
				// t.Parallel() // Oracle should not be parallel, if so, it would dead and timeout due to junk machine.
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

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name           string // Also used as build image name:tag.
		conf           string
		exposedPorts   []string
		sid            string
		optsProcess    []inputs.PointCheckOption
		optsTablespace []inputs.PointCheckOption
		optsSystem     []inputs.PointCheckOption
	}{
		{
			name:         "pubrepo.jiagouyun.com/image-repo-for-testing/oracle:11g-xe-datakit-v4",
			exposedPorts: []string{"1521/tcp"},
			sid:          "XE",
			optsProcess: []inputs.PointCheckOption{
				inputs.WithOptionalTags(
					"pdb_name",
				),
				inputs.WithOptionalFields(
					"pid",
				),
			},
			optsTablespace: []inputs.PointCheckOption{
				inputs.WithOptionalTags(
					"pdb_name",
				),
			},
			optsSystem: []inputs.PointCheckOption{
				inputs.WithOptionalTags(
					"pdb_name",
				),
				inputs.WithOptionalFields(
					"cache_blocks_corrupt",
					"cache_blocks_lost",
					"cursor_cachehit_ratio",
					"database_wait_time_ratio",
					"disk_sorts",
					"enqueue_timeouts",
					"gc_cr_block_received",
					"memory_sorts_ratio",
					"rows_per_sort",
					"service_response_time",
					"session_count",
					"session_limit_usage",
					"sorts_per_user_call",
					"temp_space_used",
					"user_rollbacks",
				),
			},
		},

		{
			name:         "pubrepo.jiagouyun.com/image-repo-for-testing/oracle:12c-se-datakit-v4",
			exposedPorts: []string{"1521/tcp"},
			sid:          "xe",
			optsTablespace: []inputs.PointCheckOption{
				inputs.WithOptionalTags(
					"pdb_name",
				),
			},
			optsSystem: []inputs.PointCheckOption{
				inputs.WithOptionalTags(
					"pdb_name",
				),
				inputs.WithOptionalFields(
					"cache_blocks_corrupt",
					"cache_blocks_lost",
					"cursor_cachehit_ratio",
					"database_wait_time_ratio",
					"disk_sorts",
					"enqueue_timeouts",
					"gc_cr_block_received",
					"memory_sorts_ratio",
					"rows_per_sort",
					"service_response_time",
					"session_count",
					"session_limit_usage",
					"sorts_per_user_call",
					"temp_space_used",
					"user_rollbacks",
				),
			},
		},

		{
			name:         "pubrepo.jiagouyun.com/image-repo-for-testing/oracle:19c-ee-datakit-v4",
			exposedPorts: []string{"1521/tcp"},
			sid:          "XE",
			optsSystem: []inputs.PointCheckOption{
				inputs.WithOptionalFields(
					"cache_blocks_corrupt",
					"cache_blocks_lost",
					"cursor_cachehit_ratio",
					"database_wait_time_ratio",
					"disk_sorts",
					"enqueue_timeouts",
					"gc_cr_block_received",
					"memory_sorts_ratio",
					"rows_per_sort",
					"service_response_time",
					"session_count",
					"session_limit_usage",
					"sorts_per_user_call",
					"temp_space_used",
					"user_rollbacks",
				),
			},
		},
	}

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := dkio.NewMockedFeeder()

		ipt := defaultInput()
		// ipt.feeder = feeder // no need.

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

			exposedPorts:   base.exposedPorts,
			sid:            base.sid,
			optsProcess:    base.optsProcess,
			optsTablespace: base.optsTablespace,
			optsSystem:     base.optsSystem,

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
	serverPorts    []string
	sid            string
	optsProcess    []inputs.PointCheckOption
	optsTablespace []inputs.PointCheckOption
	optsSystem     []inputs.PointCheckOption
	done           chan struct{}
	mCount         map[string]struct{}

	ipt    *Input
	feeder *dkio.MockedFeeder

	pool     *dockertest.Pool
	resource *dockertest.Resource

	cr *testutils.CaseResult
}

type FeedMeasurementBody []struct {
	Measurement string                 `json:"measurement"`
	Tags        map[string]string      `json:"tags"`
	Fields      map[string]interface{} `json:"fields"`
}

func (cs *caseSpec) handler(c *gin.Context) {
	uri, err := url.ParseRequestURI(c.Request.URL.RequestURI())
	if err != nil {
		cs.t.Logf("%s", err.Error())
		return
	}

	body, err := ioutil.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	if err != nil {
		cs.t.Logf("%s", err.Error())
		return
	}

	switch uri.Path {
	case "/v1/write/metric":
		pts, err := lp.ParsePoints(body, nil)
		if err != nil {
			cs.t.Logf("ParsePoints failed: %s", err.Error())
			return
		}

		newPts := dkpt2point(pts...)

		for _, pt := range newPts {
			fmt.Println(pt.LineProto())
		}

		if err := cs.checkPoint(newPts); err != nil {
			cs.t.Logf("%s", err.Error())
			require.NoError(cs.t, err)
			return
		}

	default:
		panic("unknown measurement")
	}

	if len(cs.mCount) == 3 {
		cs.done <- struct{}{}
	}
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		var opts []inputs.PointCheckOption
		opts = append(opts, inputs.WithExtraTags(cs.ipt.Tags))

		measurement := string(pt.Name())

		switch measurement {
		case oracleProcess:
			_, ok := cs.mCount[oracleProcess]
			if ok {
				continue
			}

			opts = append(opts, cs.optsProcess...)
			opts = append(opts, inputs.WithDoc(&processMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.t.Logf("oracle_process check completed!")
			cs.mCount[oracleProcess] = struct{}{}

		case oracleTablespace:
			_, ok := cs.mCount[oracleTablespace]
			if ok {
				continue
			}

			opts = append(opts, cs.optsTablespace...)
			opts = append(opts, inputs.WithDoc(&tablespaceMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.t.Logf("oracle_tablespace check completed!")
			cs.mCount[oracleTablespace] = struct{}{}

		case oracleSystem:
			_, ok := cs.mCount[oracleSystem]
			if ok {
				continue
			}

			opts = append(opts, cs.optsSystem...)
			opts = append(opts, inputs.WithDoc(&systemMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.t.Logf("oracle_system check completed!")
			cs.mCount[oracleSystem] = struct{}{}

		default: // TODO: check other measurement
			panic("unknown measurement")
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
	cs.t.Helper()

	r := testutils.GetRemote()
	dockerTCP := r.TCPURL()

	cs.t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	gin.SetMode(gin.DebugMode)
	router := gin.Default()
	router.POST("/v1/write/metric", cs.handler)

	var (
		listener    net.Listener
		randPortStr string
		err         error
	)

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

	srv := &http.Server{Handler: router}

	go func() {
		cs.done = make(chan struct{})
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
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
				Env:        []string{fmt.Sprintf("DATAKIT_HOST=%s", extIP), "DATAKIT_PORT=" + randPortStr, "ORACLE_PASSWORD=123456", "ORACLE_SID=" + cs.sid, "DATAKIT_INTERVAL=5s", "IMPORT_FROM_VOLUME=true"},

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
				Env:        []string{fmt.Sprintf("DATAKIT_HOST=%s", extIP), "DATAKIT_PORT=" + randPortStr, "ORACLE_PASSWORD=123456", "ORACLE_SID=" + cs.sid, "DATAKIT_INTERVAL=5s", "IMPORT_FROM_VOLUME=true"},

				ExposedPorts: cs.exposedPorts,
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

	if err := cs.getMappingPorts(); err != nil {
		return err
	}

	cs.t.Logf("check service(%s:%v)...", r.Host, cs.serverPorts)

	if err := cs.portsOK(r); err != nil {
		return err
	}

	cs.cr.AddField("container_ready_cost", int64(time.Since(start)))

	cs.mCount = map[string]struct{}{}

	timeout := 10 * time.Minute
	cs.t.Logf("checking oracle in %v...", timeout)
	tick := time.NewTicker(timeout)
	out := false
	for {
		if out {
			break
		}

		select {
		case <-tick.C:
			panic("check oracle timeout: " + cs.name)
		case <-cs.done:
			cs.t.Logf("check oracle all done!")
			out = true
		}
	}

	cs.t.Logf("exit...")

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
		if !r.PortOK(docker.Port(v).Port(), 2*time.Minute) {
			return fmt.Errorf("service checking failed")
		}
	}
	return nil
}

// nolint: deadcode,unused
// dkpt2point convert old io/point.Point to point.Point.
func dkpt2point(pts ...*influxdb.Point) (res []*point.Point) {
	for _, pt := range pts {
		fs, err := pt.Fields()
		if err != nil {
			continue
		}

		pt := point.NewPointV2([]byte(pt.Name()),
			append(point.NewTags(pt.Tags()), point.NewKVs(fs)...), nil)

		res = append(res, pt)
	}

	return res
}
