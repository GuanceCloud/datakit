// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mongodb

import (
	"fmt"
	"net"
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
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

// ATTENTION: Docker version should use v20.10.18 in integrate tests. Other versions are not tested.

////////////////////////////////////////////////////////////////////////////////

// /tmp/mongo-testing-initdb.d/file.js
/*
adb = db.getSiblingDB("admin")

adb.auth("root", "example");

adb.createUser({
    "user": "datakit",
    "pwd": "123456",
    "roles": [
      { role: "read", db: "admin" },
      { role: "clusterMonitor", db: "admin" },
      { role: "backup", db: "admin" },
      { role: "read", db: "local" }
    ]
  });
*/

const (
	jsFileName = "file.js"
	jsDataDir  = "mongo-testing-initdb.d"
	jsData     = `adb = db.getSiblingDB("admin")

adb.auth("root", "example");

adb.createUser({
    "user": "datakit",
    "pwd": "123456",
    "roles": [
      { role: "read", db: "admin" },
      { role: "clusterMonitor", db: "admin" },
      { role: "backup", db: "admin" },
      { role: "read", db: "local" }
    ]
  });`
)

func releaseFiles(t *testing.T) {
	t.Helper()

	releaseDir := filepath.Join(os.TempDir(), jsDataDir)
	releaseFullPath := filepath.Join(releaseDir, jsFileName)

	t.Logf("releaseDir = %s", releaseDir)
	t.Logf("releaseFullPath = %s", releaseFullPath)

	// /tmp/mongo-testing-initdb.d
	err := os.MkdirAll(releaseDir, os.ModePerm)
	if err != nil {
		t.Logf("os.MkdirAll() failed: %v", err)
	}

	// /tmp/mongo-testing-initdb.d/file.js
	err = os.WriteFile(releaseFullPath, []byte(jsData), os.ModePerm)
	if err != nil {
		t.Logf("os.WriteFile() failed: %v", err)
	}
}

////////////////////////////////////////////////////////////////////////////////

var (
	mExpect = map[string]struct{}{
		MongoDB:         {},
		MongoDBStats:    {},
		MongoDBColStats: {},
		MongoDBTopStats: {},
	}

	ignoreFieldsOpts = inputs.WithOptionalFields(
		"mapped_megabytes",
		"non-mapped_megabytes",
		"page_faults_per_sec",
		"percent_cache_dirty",
		"percent_cache_used",
		"wtcache_app_threads_page_read_count",
		"wtcache_app_threads_page_read_time",
		"wtcache_app_threads_page_write_count",
		"wtcache_bytes_read_into",
		"wtcache_bytes_written_from",
		"wtcache_current_bytes",
		"wtcache_internal_pages_evicted",
		"wtcache_max_bytes_configured",
		"wtcache_modified_pages_evicted",
		"wtcache_pages_evicted_by_app_thread",
		"wtcache_pages_queued_for_eviction",
		"wtcache_pages_read_into",
		"wtcache_pages_requested_from",
		"wtcache_pages_written_from",
		"wtcache_server_evicting_pages",
		"wtcache_tracked_dirty_bytes",
		"wtcache_unmodified_pages_evicted",
		"wtcache_worker_thread_evictingpages",
	)

	ignoreFieldsOptsDB = inputs.WithOptionalFields(
		"mapped_megabytes",
		"non-mapped_megabytes",
		"page_faults_per_sec",
	)

	extraTagsElection = inputs.WithExtraTags(map[string]string{
		"election": "1",
		"tag1":     "val1",
	})
)

const (
	confElection = `interval = "1s"
	servers = ["mongodb://datakit:123456@"]
	gather_replica_set_stats = false
	gather_cluster_stats = false
	gather_per_db_stats = true
	gather_per_col_stats = true
	col_stats_dbs = []
	gather_top_stat = true
	election = true
[tags]
	tag1 = "val1"`  // set conf URL later.

	confNoElection = `interval = "1s"
	servers = ["mongodb://datakit:123456@"]
	gather_replica_set_stats = false
	gather_cluster_stats = false
	gather_per_db_stats = true
	gather_per_col_stats = true
	col_stats_dbs = []
	gather_top_stat = true
	election = false
[tags]
	tag1 = "val1"`  // set conf URL later.
)

func TestIntegrate(t *testing.T) {
	if !testutils.CheckIntegrationTestingRunning() {
		t.Skip()
	}

	releaseFiles(t)

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

func getConfAccessPointDatakit(host, port string) []string {
	return []string{fmt.Sprintf("mongodb://datakit:123456@%s", net.JoinHostPort(host, port))}
}

func getConfAccessPointRoot(host, port string) []string {
	return []string{fmt.Sprintf("mongodb://root:example@%s", net.JoinHostPort(host, port))}
}

type AccessType int

const (
	AccessDatakit AccessType = 0
	AccessRoot    AccessType = 1
)

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name             string // Also used as build image name:tag.
		conf             string
		dockerFileText   string // Empty if not build image.
		exposedPorts     []string
		cmd              []string
		accessType       AccessType
		optsDB           []inputs.PointCheckOption
		optsDBStats      []inputs.PointCheckOption
		optsDBColStats   []inputs.PointCheckOption
		optsDBShardStats []inputs.PointCheckOption
		optsDBTopStats   []inputs.PointCheckOption
	}{
		////////////////////////////////////////////////////////////////////////
		// Mongo 2.8.0
		////////////////////////////////////////////////////////////////////////
		{
			name:         "mongo:2.8.0",
			conf:         confElection, // set conf URL later.
			exposedPorts: []string{"27017/tcp"},
			optsDB: []inputs.PointCheckOption{
				ignoreFieldsOpts,
				extraTagsElection,
			},
			optsDBStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
				extraTagsElection,
			},
			optsDBColStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
				extraTagsElection,
			},
			optsDBTopStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
				extraTagsElection,
			},
		},
		{
			name:         "mongo:2.8.0",
			conf:         confNoElection, // set conf URL later.
			exposedPorts: []string{"27017/tcp"},
			optsDB: []inputs.PointCheckOption{
				ignoreFieldsOpts,
			},
			optsDBStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
			},
			optsDBColStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
			},
			optsDBTopStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
			},
		},

		////////////////////////////////////////////////////////////////////////
		// Mongo 3.0
		////////////////////////////////////////////////////////////////////////
		{
			name:         "mongo:3.0",
			conf:         confElection, // set conf URL later.
			exposedPorts: []string{"27017/tcp"},
			cmd:          []string{"docker-entrypoint.sh", "mongod", "--smallfiles"},
			optsDB: []inputs.PointCheckOption{
				ignoreFieldsOpts,
				extraTagsElection,
			},
			optsDBStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
				extraTagsElection,
			},
			optsDBColStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
				extraTagsElection,
			},
			optsDBTopStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
				extraTagsElection,
			},
		},
		{
			name:         "mongo:3.0",
			conf:         confNoElection, // set conf URL later.
			exposedPorts: []string{"27017/tcp"},
			cmd:          []string{"docker-entrypoint.sh", "mongod", "--smallfiles"},
			optsDB: []inputs.PointCheckOption{
				ignoreFieldsOpts,
			},
			optsDBStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
			},
			optsDBColStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
			},
			optsDBTopStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
			},
		},

		////////////////////////////////////////////////////////////////////////
		// Mongo 4.0
		////////////////////////////////////////////////////////////////////////
		{
			name:         "mongo:4.0",
			conf:         confElection, // set conf URL later.
			exposedPorts: []string{"27017/tcp"},
			optsDB: []inputs.PointCheckOption{
				ignoreFieldsOptsDB,
				extraTagsElection,
			},
			optsDBStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
				extraTagsElection,
			},
			optsDBColStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
				extraTagsElection,
			},
			optsDBTopStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
				extraTagsElection,
			},
		},
		{
			name:         "mongo:4.0",
			conf:         confNoElection, // set conf URL later.
			exposedPorts: []string{"27017/tcp"},
			optsDB: []inputs.PointCheckOption{
				ignoreFieldsOptsDB,
			},
			optsDBStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
			},
			optsDBColStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
			},
			optsDBTopStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
			},
		},

		////////////////////////////////////////////////////////////////////////
		// Mongo 5.0
		////////////////////////////////////////////////////////////////////////
		{
			name:         "mongo:5.0",
			conf:         confElection, // set conf URL later.
			exposedPorts: []string{"27017/tcp"},
			optsDB: []inputs.PointCheckOption{
				ignoreFieldsOptsDB,
				extraTagsElection,
			},
			optsDBStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
				extraTagsElection,
			},
			optsDBColStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
				extraTagsElection,
			},
			optsDBTopStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
				extraTagsElection,
			},
		},
		{
			name:         "mongo:5.0",
			conf:         confNoElection, // set conf URL later.
			exposedPorts: []string{"27017/tcp"},
			optsDB: []inputs.PointCheckOption{
				ignoreFieldsOptsDB,
			},
			optsDBStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
			},
			optsDBColStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
			},
			optsDBTopStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
			},
		},

		////////////////////////////////////////////////////////////////////////
		// Mongo 6.0
		////////////////////////////////////////////////////////////////////////
		{
			name:         "mongo:6.0",
			conf:         confElection, // set conf URL later.
			exposedPorts: []string{"27017/tcp"},
			optsDB: []inputs.PointCheckOption{
				ignoreFieldsOptsDB,
				extraTagsElection,
			},
			optsDBStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
				extraTagsElection,
			},
			optsDBColStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
				extraTagsElection,
			},
			optsDBTopStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
				extraTagsElection,
			},
		},
		{
			name:         "mongo:6.0",
			conf:         confNoElection, // set conf URL later.
			exposedPorts: []string{"27017/tcp"},
			optsDB: []inputs.PointCheckOption{
				ignoreFieldsOptsDB,
			},
			optsDBStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
			},
			optsDBColStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
			},
			optsDBTopStats: []inputs.PointCheckOption{
				ignoreFieldsOpts,
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
			accessType:     base.accessType,

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
	serverPorts      []string
	accessType       AccessType
	optsDB           []inputs.PointCheckOption
	optsDBStats      []inputs.PointCheckOption
	optsDBColStats   []inputs.PointCheckOption
	optsDBShardStats []inputs.PointCheckOption
	optsDBTopStats   []inputs.PointCheckOption
	cmd              []string
	mCount           map[string]struct{}

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

		measurement := pt.Name()

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

			cs.mCount[MongoDB] = struct{}{}

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

			cs.mCount[MongoDBStats] = struct{}{}

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

			cs.mCount[MongoDBColStats] = struct{}{}

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

			cs.mCount[MongoDBShardStats] = struct{}{}

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

			cs.mCount[MongoDBTopStats] = struct{}{}

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

	mounts := filepath.Join(os.TempDir(), jsDataDir) + ":/docker-entrypoint-initdb.d"

	if len(cs.dockerFileText) == 0 {
		// Just run a container from existing docker image.
		resource, err = p.RunWithOptions(
			&dockertest.RunOptions{
				Name: uniqueContainerName, // ATTENTION: not cs.name.

				Repository: cs.repo,
				Tag:        cs.repoTag,
				Env:        []string{"MONGO_INITDB_ROOT_USERNAME=root", "MONGO_INITDB_ROOT_PASSWORD=example"},
				Cmd:        cs.cmd,

				Mounts: []string{
					mounts,
				},

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
				Env:        []string{"MONGO_INITDB_ROOT_USERNAME=root", "MONGO_INITDB_ROOT_PASSWORD=example"},
				Cmd:        cs.cmd,

				Mounts: []string{
					mounts,
				},

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

	switch cs.accessType {
	case AccessDatakit:
		cs.ipt.Servers = getConfAccessPointDatakit(r.Host, cs.serverPorts[0]) // set conf URL here.
	case AccessRoot:
		cs.ipt.Servers = getConfAccessPointRoot(r.Host, cs.serverPorts[0]) // set conf URL here.
	}

	cs.t.Logf("check service(%s:%v)...", r.Host, cs.serverPorts)

	if err := cs.portsOK(r); err != nil {
		return err
	}

	cs.cr.AddField("container_ready_cost", int64(time.Since(start)))

	if cs.name == "mongo:2.8.0" {
		tick := time.NewTicker(time.Second)
		defer tick.Stop()
		timeout := time.NewTicker(time.Minute)
		defer timeout.Stop()

		out := false

		for {
			if out {
				break
			}

			select {
			case <-tick.C:
				code, err := cs.resource.Exec([]string{
					"mongo",
					"/docker-entrypoint-initdb.d/file.js",
				}, dockertest.ExecOptions{})
				cs.t.Logf("Exec() code = %d, err = %v", code, err)
				if code == 0 && err == nil {
					out = true
				}
			case <-timeout.C:
				cs.t.Logf("ERROR: timeout: %s", cs.name)
				out = true
			default:
				time.Sleep(time.Second)
			}
		}
	}

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

	// for _, v := range pts {
	// 	cs.t.Logf("pt = %s", v.LineProto())
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

	tmpDir, err := os.MkdirTemp("", "dockerfiles_")
	if err != nil {
		cs.t.Logf("os.MkdirTemp() failed: %s", err.Error())
		return "", "", err
	}

	tmpFile, err := os.CreateTemp(tmpDir, "dockerfile_")
	if err != nil {
		cs.t.Logf("os.CreateTemp() failed: %s", err.Error())
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
