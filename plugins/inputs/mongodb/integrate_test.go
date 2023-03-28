// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mongodb

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	dockertest "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestMongoInput(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			caseStart := time.Now()

			t.Logf("testing %s...", tc.name)

			if err := tc.run(); err != nil {
				tc.cr.Status = testutils.TestFailed
				tc.cr.FailedMessage = err.Error()

				assert.NoError(t, err)
			} else {
				tc.cr.Status = testutils.TestPassed
			}

			tc.cr.Cost = time.Since(caseStart)

			assert.NoError(t, testutils.Flush(tc.cr))

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

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name             string // Also used as build image name:tag.
		conf             string
		dockerFileText   string // Empty if not build image.
		exposedPorts     []string
		cmd              []string
		optsDB           []inputs.PointCheckOption
		optsDBStats      []inputs.PointCheckOption
		optsDBColStats   []inputs.PointCheckOption
		optsDBShardStats []inputs.PointCheckOption
		optsDBTopStats   []inputs.PointCheckOption
	}{
		{
			name: "mongo:3.0",
			conf: fmt.Sprintf(`interval = "10s"
			servers = ["mongodb://root:example@%s:27017"]
			gather_replica_set_stats = false
			gather_cluster_stats = false
			gather_per_db_stats = true
			gather_per_col_stats = true
			col_stats_dbs = []
			gather_top_stat = true
			election = true
		[tags]
			tag1 = "val1"`, remote.Host),
			exposedPorts: []string{"27017/tcp"},
			cmd:          []string{"docker-entrypoint.sh", "mongod", "--smallfiles"},
			optsDB: []inputs.PointCheckOption{
				inputs.WithOptionalFields("wtcache_unmodified_pages_evicted", "percent_cache_dirty", "wtcache_max_bytes_configured", "percent_cache_used", "wtcache_internal_pages_evicted", "wtcache_pages_read_into", "wtcache_server_evicting_pages", "wtcache_app_threads_page_write_count", "wtcache_app_threads_page_read_time", "wtcache_pages_written_from", "wtcache_current_bytes", "wtcache_app_threads_page_read_count", "wtcache_pages_requested_from", "wtcache_pages_evicted_by_app_thread", "wtcache_tracked_dirty_bytes", "wtcache_worker_thread_evictingpages", "wtcache_bytes_written_from", "wtcache_pages_queued_for_eviction", "wtcache_bytes_read_into", "wtcache_modified_pages_evicted", "non-mapped_megabytes", "mapped_megabytes", "page_faults_per_sec"), // nolint:lll
			},
		},

		{
			name: "mongo:4.0",
			conf: fmt.Sprintf(`interval = "10s"
			servers = ["mongodb://root:example@%s:27017"]
			gather_replica_set_stats = false
			gather_cluster_stats = false
			gather_per_db_stats = true
			gather_per_col_stats = true
			col_stats_dbs = []
			gather_top_stat = true
			election = true
		[tags]
			tag1 = "val1"`, remote.Host),
			exposedPorts: []string{"27017/tcp"},
			optsDB: []inputs.PointCheckOption{
				inputs.WithOptionalFields("non-mapped_megabytes", "mapped_megabytes", "page_faults_per_sec"), // nolint:lll
			},
			optsDBStats: []inputs.PointCheckOption{
				inputs.WithOptionalFields("wtcache_unmodified_pages_evicted", "percent_cache_dirty", "wtcache_app_threads_page_read_count", "wtcache_max_bytes_configured", "wtcache_pages_evicted_by_app_thread", "wtcache_pages_queued_for_eviction", "wtcache_current_bytes", "wtcache_modified_pages_evicted", "wtcache_app_threads_page_write_count", "wtcache_worker_thread_evictingpages", "wtcache_bytes_read_into", "wtcache_tracked_dirty_bytes", "wtcache_pages_written_from", "wtcache_pages_requested_from", "wtcache_bytes_written_from", "percent_cache_used", "wtcache_app_threads_page_read_time", "wtcache_internal_pages_evicted", "wtcache_server_evicting_pages", "wtcache_pages_read_into"), // nolint:lll
			},
		},

		{
			name: "mongo:5.0",
			conf: fmt.Sprintf(`interval = "10s"
			servers = ["mongodb://root:example@%s:27017"]
			gather_replica_set_stats = false
			gather_cluster_stats = false
			gather_per_db_stats = true
			gather_per_col_stats = true
			col_stats_dbs = []
			gather_top_stat = true
			election = true
		[tags]
			tag1 = "val1"`, remote.Host),
			exposedPorts: []string{"27017/tcp"},
			optsDB: []inputs.PointCheckOption{
				inputs.WithOptionalFields("non-mapped_megabytes", "mapped_megabytes", "page_faults_per_sec"), // nolint:lll
			},
			optsDBStats: []inputs.PointCheckOption{
				inputs.WithOptionalFields("wtcache_unmodified_pages_evicted", "percent_cache_dirty", "wtcache_app_threads_page_read_count", "wtcache_max_bytes_configured", "wtcache_pages_evicted_by_app_thread", "wtcache_pages_queued_for_eviction", "wtcache_current_bytes", "wtcache_modified_pages_evicted", "wtcache_app_threads_page_write_count", "wtcache_worker_thread_evictingpages", "wtcache_bytes_read_into", "wtcache_tracked_dirty_bytes", "wtcache_pages_written_from", "wtcache_pages_requested_from", "wtcache_bytes_written_from", "percent_cache_used", "wtcache_app_threads_page_read_time", "wtcache_internal_pages_evicted", "wtcache_server_evicting_pages", "wtcache_pages_read_into"), // nolint:lll
			},
		},

		{
			name: "mongo:6.0",
			conf: fmt.Sprintf(`interval = "10s"
			servers = ["mongodb://root:example@%s:27017"]
			gather_replica_set_stats = false
			gather_cluster_stats = false
			gather_per_db_stats = true
			gather_per_col_stats = true
			col_stats_dbs = []
			gather_top_stat = true
			election = true
		[tags]
			tag1 = "val1"`, remote.Host),
			exposedPorts: []string{"27017/tcp"},
			optsDB: []inputs.PointCheckOption{
				inputs.WithOptionalFields("non-mapped_megabytes", "mapped_megabytes", "page_faults_per_sec"), // nolint:lll
			},
			optsDBStats: []inputs.PointCheckOption{
				inputs.WithOptionalFields("wtcache_unmodified_pages_evicted", "percent_cache_dirty", "wtcache_app_threads_page_read_count", "wtcache_max_bytes_configured", "wtcache_pages_evicted_by_app_thread", "wtcache_pages_queued_for_eviction", "wtcache_current_bytes", "wtcache_modified_pages_evicted", "wtcache_app_threads_page_write_count", "wtcache_worker_thread_evictingpages", "wtcache_bytes_read_into", "wtcache_tracked_dirty_bytes", "wtcache_pages_written_from", "wtcache_pages_requested_from", "wtcache_bytes_written_from", "percent_cache_used", "wtcache_app_threads_page_read_time", "wtcache_internal_pages_evicted", "wtcache_server_evicting_pages", "wtcache_pages_read_into"), // nolint:lll
			},
		},
	}

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := io.NewMockedFeeder()

		ipt := defaultInput()
		ipt.feeder = feeder

		_, err := toml.Decode(base.conf, ipt)
		assert.NoError(t, err)

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

			optsDB:           base.optsDB,
			optsDBStats:      base.optsDBStats,
			optsDBColStats:   base.optsDBColStats,
			optsDBShardStats: base.optsDBShardStats,
			optsDBTopStats:   base.optsDBTopStats,

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

	name             string
	repo             string
	repoTag          string
	dockerFileText   string
	exposedPorts     []string
	optsDB           []inputs.PointCheckOption
	optsDBStats      []inputs.PointCheckOption
	optsDBColStats   []inputs.PointCheckOption
	optsDBShardStats []inputs.PointCheckOption
	optsDBTopStats   []inputs.PointCheckOption
	cmd              []string

	ipt    *Input
	feeder *io.MockedFeeder

	pool     *dockertest.Pool
	resource *dockertest.Resource

	cr *testutils.CaseResult
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	var opts []inputs.PointCheckOption
	opts = append(opts, inputs.WithExtraTags(cs.ipt.Tags))

	for _, pt := range pts {
		measurement := string(pt.Name())

		switch measurement {
		case MongoDB:
			opts = append(opts, cs.optsDB...)
			opts = append(opts, inputs.WithDoc(&mongodbMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

		case MongoDBStats:
			opts = append(opts, cs.optsDBStats...)
			opts = append(opts, inputs.WithDoc(&mongodbDBMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

		case MongoDBColStats:
			opts = append(opts, cs.optsDBColStats...)
			opts = append(opts, inputs.WithDoc(&mongodbColMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

		case MongoDBShardStats:
			opts = append(opts, cs.optsDBShardStats...)
			opts = append(opts, inputs.WithDoc(&mongodbShardMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

		case MongoDBTopStats:
			opts = append(opts, cs.optsDBTopStats...)
			opts = append(opts, inputs.WithDoc(&mongodbTopMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

		default: // TODO: check other measurement
			panic("not implement")
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

	containerName := cs.getContainterName()

	// Remove the container if exist.
	if err := p.RemoveContainerByName(containerName); err != nil {
		return err
	}

	dockerFileDir, dockerFilePath, err := cs.getDockerFilePath()
	if err != nil {
		return err
	}
	defer os.RemoveAll(dockerFileDir)

	var resource *dockertest.Resource

	if len(cs.dockerFileText) == 0 {
		// Just run a container from existing docker image.
		resource, err = p.RunWithOptions(
			&dockertest.RunOptions{
				Name: containerName, // ATTENTION: not cs.name.

				Repository: cs.repo,
				Tag:        cs.repoTag,
				Env:        []string{"MONGO_INITDB_ROOT_USERNAME=root", "MONGO_INITDB_ROOT_PASSWORD=example"},
				Cmd:        cs.cmd,

				ExposedPorts: cs.exposedPorts,
				PortBindings: cs.getPortBindings(),
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
				Name: cs.name,

				Repository: cs.repo,
				Tag:        cs.repoTag,
				Env:        []string{"MONGO_INITDB_ROOT_USERNAME=root", "MONGO_INITDB_ROOT_PASSWORD=example"},
				Cmd:        cs.cmd,

				ExposedPorts: cs.exposedPorts,
				PortBindings: cs.getPortBindings(),
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
	pts, err := cs.feeder.AnyPoints()
	if err != nil {
		return err
	}

	cs.cr.AddField("point_latency", int64(time.Since(start)))
	cs.cr.AddField("point_count", len(pts))

	for _, v := range pts {
		cs.t.Logf("pt = %s", v.LineProto())
	}

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

func (cs *caseSpec) getContainterName() string {
	nameTag := strings.Split(cs.name, ":")
	name := filepath.Base(nameTag[0])
	return name
}

func (cs *caseSpec) getPortBindings() map[docker.Port][]docker.PortBinding {
	portBindings := make(map[docker.Port][]docker.PortBinding)

	for _, v := range cs.exposedPorts {
		portBindings[docker.Port(v)] = []docker.PortBinding{{HostIP: "0.0.0.0", HostPort: docker.Port(v).Port()}}
	}

	return portBindings
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
