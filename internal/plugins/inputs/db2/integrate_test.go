// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package db2

import (
	"context"
	"errors"
	"fmt"
	"io"
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

	r := testutils.GetRemote()
	dockerTCP := r.TCPURL()
	p, res, mounts, err := testutils.RunOraemon(dockerTCP)
	if err != nil {
		panic("RunOraemon failed:" + err.Error())
	}

	testutils.PurgeRemoteByName(inputName) // purge at first.

	defer testutils.RemoveOraemon(p, res)
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
				// t.Parallel() // Should not be parallel, if so, it would dead and timeout due to junk machine.
				caseStart := time.Now()
				tc.mounts = mounts

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
		name               string // Also used as build image name:tag.
		conf               string
		exposedPorts       []string
		dbVolume           string
		optsInstance       []inputs.PointCheckOption
		optsDatabase       []inputs.PointCheckOption
		optsBufferPool     []inputs.PointCheckOption
		optsTableSpace     []inputs.PointCheckOption
		optsTransactionLog []inputs.PointCheckOption
	}{
		{
			name:         "pubrepo.guance.com/image-repo-for-testing/db2:11.5.0.0a-datakit",
			exposedPorts: []string{"50000/tcp"},
			dbVolume:     "/tmp/db2_11.5.0.0a:/database",
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

			exposedPorts:       base.exposedPorts,
			dbVolume:           base.dbVolume,
			optsInstance:       base.optsInstance,
			optsDatabase:       base.optsDatabase,
			optsBufferPool:     base.optsBufferPool,
			optsTableSpace:     base.optsTableSpace,
			optsTransactionLog: base.optsTransactionLog,

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

	name               string
	repo               string
	repoTag            string
	dockerFileText     string
	exposedPorts       []string
	serverPorts        []string
	optsInstance       []inputs.PointCheckOption
	optsDatabase       []inputs.PointCheckOption
	optsBufferPool     []inputs.PointCheckOption
	optsTableSpace     []inputs.PointCheckOption
	optsTransactionLog []inputs.PointCheckOption
	done               chan struct{}
	mCount             map[string]struct{}
	mounts             string
	dbVolume           string

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

	body, err := io.ReadAll(c.Request.Body)
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
			cs.t.Log(pt.LineProto())
		}

		if err := cs.checkPoint(newPts); err != nil {
			cs.t.Logf("%s", err.Error())
			require.NoError(cs.t, err)
			return
		}

	default:
		panic("unknown uri.Path: " + uri.Path)
	}

	if len(cs.mCount) == 5 {
		cs.done <- struct{}{}
	}
}

func (cs *caseSpec) lasterror(c *gin.Context) {
	uri, err := url.ParseRequestURI(c.Request.URL.RequestURI())
	if err != nil {
		cs.t.Logf("%s", err.Error())
		return
	}
	cs.t.Log("uri ==>", uri)

	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	if err != nil {
		cs.t.Logf("%s", err.Error())
		return
	}
	cs.t.Log("lasterror ==>", string(body))
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		var opts []inputs.PointCheckOption
		opts = append(opts, inputs.WithExtraTags(cs.ipt.Tags))

		measurement := pt.Name()

		switch measurement {
		case metricNameInstance:
			opts = append(opts, cs.optsInstance...)
			opts = append(opts, inputs.WithDoc(&instanceMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.t.Logf("db2_instance check completed!")
			cs.mCount[metricNameInstance] = struct{}{}

		case metricNameDatabase:
			opts = append(opts, cs.optsDatabase...)
			opts = append(opts, inputs.WithDoc(&databaseMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.t.Logf("db2_database check completed!")
			cs.mCount[metricNameDatabase] = struct{}{}

		case metricNameBufferPool:
			opts = append(opts, cs.optsBufferPool...)
			opts = append(opts, inputs.WithDoc(&bufferPoolMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.t.Logf("db2_buffer_pool check completed!")
			cs.mCount[metricNameBufferPool] = struct{}{}

		case metricNameTableSpace:
			opts = append(opts, cs.optsTableSpace...)
			opts = append(opts, inputs.WithDoc(&tableSpaceMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.t.Logf("db2_table_space check completed!")
			cs.mCount[metricNameTableSpace] = struct{}{}

		case metricNameTransactionLog:
			opts = append(opts, cs.optsTransactionLog...)
			opts = append(opts, inputs.WithDoc(&transactionLogMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.t.Logf("db2_transaction_log check completed!")
			cs.mCount[metricNameTransactionLog] = struct{}{}

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
	cs.t.Helper()

	r := testutils.GetRemote()
	dockerTCP := r.TCPURL()

	cs.t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	gin.SetMode(gin.DebugMode)
	router := gin.Default()
	router.POST("/v1/write/metric", cs.handler)
	router.POST("/v1/lasterror", cs.lasterror)

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

	p, err := testutils.GetPool(dockerTCP)
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
					"LICENSE=accept",
					"DB2INST1_PASSWORD=123456",
					"DBNAME=testdb",
					"DK_LOG_LEVEL=debug",
					"DK_INTERVAL=5s",
					"DK_USER=db2inst1",
					"DK_DB=testdb",
					"DK_SERVICE=db2_service_name",
					"DK_HOST=" + extIP,
					"DK_PORT=" + randPortStr,
				},
				ExposedPorts: cs.exposedPorts,
				Mounts: []string{
					cs.mounts,
					cs.dbVolume,
				},
				Privileged: true,
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
					"LICENSE=accept",
					"DB2INST1_PASSWORD=123456",
					"DBNAME=testdb",
					"LOG_LEVEL=debug",
					"DATAKIT_INTERVAL=5s",
					"USER_NAME=db2inst1",
					"PASSWORD=123456",
					"DATABASE=testdb",
					"SERVICE_NAME=db2_service_name",
					"DATAKIT_HOST=" + extIP,
					"DATAKIT_PORT=" + randPortStr,
				},
				ExposedPorts: cs.exposedPorts,
				Mounts: []string{
					cs.mounts,
					cs.dbVolume,
				},
				Privileged: true,
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
	cs.t.Logf("checking db2 in %v...", timeout)
	tick := time.NewTicker(timeout)
	out := false
	for {
		if out {
			break
		}

		select {
		case <-tick.C:
			panic("check db2 timeout: " + cs.name)
		case <-cs.done:
			cs.t.Logf("check db2 all done!")
			out = true
		}
	}

	cs.t.Logf("exit...")

	return nil
}

func (cs *caseSpec) getDockerFilePath() (dirName string, fileName string, err error) {
	if len(cs.dockerFileText) == 0 {
		return
	}

	tmpDir, err := os.MkdirTemp("", "dockerfiles_")
	if err != nil {
		cs.t.Logf("os.MkdirTemp.TempDir failed: %s", err.Error())
		return "", "", err
	}

	tmpFile, err := os.CreateTemp(tmpDir, "dockerfile_")
	if err != nil {
		cs.t.Logf("os.CreateTemp failed: %s", err.Error())
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
		if !r.PortOK(docker.Port(v).Port(), 15*time.Minute) {
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

		pt := point.NewPointV2(pt.Name(),
			append(point.NewTags(pt.Tags()), point.NewKVs(fs)...), nil)

		res = append(res, pt)
	}

	return res
}
