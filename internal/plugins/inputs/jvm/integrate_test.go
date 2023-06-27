// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jvm

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
// Reference: https://jolokia.org/reference/html/agents.html#jvm-agent

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

					require.NoError(t, tc.pool.Purge(tc.resource))
				})
			})
		}(tc)
	}
}

func getConfAccessPoint(host, port string) []string {
	return []string{fmt.Sprintf("http://%s/jolokia", net.JoinHostPort(host, port))}
}

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name                 string // Also used as build image name:tag.
		conf                 string
		dockerFileText       string // Empty if not build image.
		exposedPorts         []string
		optsJavaRuntime      []inputs.PointCheckOption
		optsJavaMemory       []inputs.PointCheckOption
		optsGarbageCollector []inputs.PointCheckOption
		optsThreading        []inputs.PointCheckOption
		optsClassLoading     []inputs.PointCheckOption
		optsMemoryPool       []inputs.PointCheckOption
	}{
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/java:jvm-jolokia-8",
			conf: `urls = [""]
			interval   = "1s"
			[[metric]]
			  name  = "java_runtime"
			  mbean = "java.lang:type=Runtime"
			  paths = ["Uptime"]
			[[metric]]
			  name  = "java_memory"
			  mbean = "java.lang:type=Memory"
			  paths = ["HeapMemoryUsage", "NonHeapMemoryUsage", "ObjectPendingFinalizationCount"]
			[[metric]]
			  name     = "java_garbage_collector"
			  mbean    = "java.lang:name=*,type=GarbageCollector"
			  paths    = ["CollectionTime", "CollectionCount"]
			  tag_keys = ["name"]
			[[metric]]
			  name  = "java_threading"
			  mbean = "java.lang:type=Threading"
			  paths = ["TotalStartedThreadCount", "ThreadCount", "DaemonThreadCount", "PeakThreadCount"]
			[[metric]]
			  name  = "java_class_loading"
			  mbean = "java.lang:type=ClassLoading"
			  paths = ["LoadedClassCount", "UnloadedClassCount", "TotalLoadedClassCount"]
			[[metric]]
			  name     = "java_memory_pool"
			  mbean    = "java.lang:name=*,type=MemoryPool"
			  paths    = ["Usage", "PeakUsage", "CollectionUsage"]
			  tag_keys = ["name"]`, // set conf URL later.
			exposedPorts: []string{"59090/tcp"},
			optsJavaRuntime: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsJavaMemory: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsGarbageCollector: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsThreading: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsClassLoading: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsMemoryPool: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
		},

		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/java:jvm-jolokia-11",
			conf: `urls = [""]
			interval   = "1s"
			[[metric]]
			  name  = "java_runtime"
			  mbean = "java.lang:type=Runtime"
			  paths = ["Uptime"]
			[[metric]]
			  name  = "java_memory"
			  mbean = "java.lang:type=Memory"
			  paths = ["HeapMemoryUsage", "NonHeapMemoryUsage", "ObjectPendingFinalizationCount"]
			[[metric]]
			  name     = "java_garbage_collector"
			  mbean    = "java.lang:name=*,type=GarbageCollector"
			  paths    = ["CollectionTime", "CollectionCount"]
			  tag_keys = ["name"]
			[[metric]]
			  name  = "java_threading"
			  mbean = "java.lang:type=Threading"
			  paths = ["TotalStartedThreadCount", "ThreadCount", "DaemonThreadCount", "PeakThreadCount"]
			[[metric]]
			  name  = "java_class_loading"
			  mbean = "java.lang:type=ClassLoading"
			  paths = ["LoadedClassCount", "UnloadedClassCount", "TotalLoadedClassCount"]
			[[metric]]
			  name     = "java_memory_pool"
			  mbean    = "java.lang:name=*,type=MemoryPool"
			  paths    = ["Usage", "PeakUsage", "CollectionUsage"]
			  tag_keys = ["name"]`, // set conf URL later.
			exposedPorts: []string{"59090/tcp"},
			optsJavaRuntime: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsJavaMemory: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsGarbageCollector: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsThreading: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsClassLoading: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsMemoryPool: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
		},

		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/java:jvm-jolokia-17",
			conf: `urls = [""]
			interval   = "1s"
			[[metric]]
			  name  = "java_runtime"
			  mbean = "java.lang:type=Runtime"
			  paths = ["Uptime"]
			[[metric]]
			  name  = "java_memory"
			  mbean = "java.lang:type=Memory"
			  paths = ["HeapMemoryUsage", "NonHeapMemoryUsage", "ObjectPendingFinalizationCount"]
			[[metric]]
			  name     = "java_garbage_collector"
			  mbean    = "java.lang:name=*,type=GarbageCollector"
			  paths    = ["CollectionTime", "CollectionCount"]
			  tag_keys = ["name"]
			[[metric]]
			  name  = "java_threading"
			  mbean = "java.lang:type=Threading"
			  paths = ["TotalStartedThreadCount", "ThreadCount", "DaemonThreadCount", "PeakThreadCount"]
			[[metric]]
			  name  = "java_class_loading"
			  mbean = "java.lang:type=ClassLoading"
			  paths = ["LoadedClassCount", "UnloadedClassCount", "TotalLoadedClassCount"]
			[[metric]]
			  name     = "java_memory_pool"
			  mbean    = "java.lang:name=*,type=MemoryPool"
			  paths    = ["Usage", "PeakUsage", "CollectionUsage"]
			  tag_keys = ["name"]`, // set conf URL later.
			exposedPorts: []string{"59090/tcp"},
			optsJavaRuntime: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsJavaMemory: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsGarbageCollector: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsThreading: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsClassLoading: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsMemoryPool: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
		},

		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/java:jvm-jolokia-20",
			conf: `urls = [""]
			interval   = "1s"
			[[metric]]
			  name  = "java_runtime"
			  mbean = "java.lang:type=Runtime"
			  paths = ["Uptime"]
			[[metric]]
			  name  = "java_memory"
			  mbean = "java.lang:type=Memory"
			  paths = ["HeapMemoryUsage", "NonHeapMemoryUsage", "ObjectPendingFinalizationCount"]
			[[metric]]
			  name     = "java_garbage_collector"
			  mbean    = "java.lang:name=*,type=GarbageCollector"
			  paths    = ["CollectionTime", "CollectionCount"]
			  tag_keys = ["name"]
			[[metric]]
			  name  = "java_threading"
			  mbean = "java.lang:type=Threading"
			  paths = ["TotalStartedThreadCount", "ThreadCount", "DaemonThreadCount", "PeakThreadCount"]
			[[metric]]
			  name  = "java_class_loading"
			  mbean = "java.lang:type=ClassLoading"
			  paths = ["LoadedClassCount", "UnloadedClassCount", "TotalLoadedClassCount"]
			[[metric]]
			  name     = "java_memory_pool"
			  mbean    = "java.lang:name=*,type=MemoryPool"
			  paths    = ["Usage", "PeakUsage", "CollectionUsage"]
			  tag_keys = ["name"]`, // set conf URL later.
			exposedPorts: []string{"59090/tcp"},
			optsJavaRuntime: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsJavaMemory: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsGarbageCollector: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsThreading: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsClassLoading: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
			optsMemoryPool: []inputs.PointCheckOption{
				inputs.WithOptionalFields("CollectionUsageinit", "CollectionUsagecommitted", "CollectionUsagemax", "CollectionUsageused"), // nolint:lll
			},
		},
	}

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := io.NewMockedFeeder()

		ipt := defaultInput()
		ipt.Feeder = feeder

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

			dockerFileText: base.dockerFileText,
			exposedPorts:   base.exposedPorts,

			optsJavaRuntime:      base.optsJavaRuntime,
			optsJavaMemory:       base.optsJavaMemory,
			optsGarbageCollector: base.optsGarbageCollector,
			optsThreading:        base.optsThreading,
			optsClassLoading:     base.optsClassLoading,
			optsMemoryPool:       base.optsMemoryPool,

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

	name                 string
	repo                 string
	repoTag              string
	dockerFileText       string
	exposedPorts         []string
	serverPorts          []string
	optsJavaRuntime      []inputs.PointCheckOption
	optsJavaMemory       []inputs.PointCheckOption
	optsGarbageCollector []inputs.PointCheckOption
	optsThreading        []inputs.PointCheckOption
	optsClassLoading     []inputs.PointCheckOption
	optsMemoryPool       []inputs.PointCheckOption
	mCount               map[string]struct{}

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
		case javaRuntime:
			opts = append(opts, cs.optsJavaRuntime...)
			opts = append(opts, inputs.WithDoc(&JavaRuntimeMemt{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[javaRuntime] = struct{}{}

		case javaMemory:
			opts = append(opts, cs.optsJavaMemory...)
			opts = append(opts, inputs.WithDoc(&JavaMemoryMemt{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[javaMemory] = struct{}{}

		case javaGarbageCollector:
			opts = append(opts, cs.optsGarbageCollector...)
			opts = append(opts, inputs.WithDoc(&JavaGcMemt{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[javaGarbageCollector] = struct{}{}

		case javaThreading:
			opts = append(opts, cs.optsThreading...)
			opts = append(opts, inputs.WithDoc(&JavaThreadMemt{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[javaThreading] = struct{}{}

		case javaClassLoading:
			opts = append(opts, cs.optsClassLoading...)
			opts = append(opts, inputs.WithDoc(&JavaClassLoadMemt{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[javaClassLoading] = struct{}{}

		case javaMemoryPool:
			opts = append(opts, cs.optsMemoryPool...)
			opts = append(opts, inputs.WithDoc(&JavaMemoryPoolMemt{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[javaMemoryPool] = struct{}{}

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
				Env:        []string{"JOLOKIA_PORT=59090"},

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
				Env:        []string{"JOLOKIA_PORT=59090"},

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
	cs.ipt.URLs = getConfAccessPoint(r.Host, cs.serverPorts[0]) // set conf URL here.

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
	pts, err := cs.feeder.NPoints(50, 5*time.Minute)
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

	require.Equal(cs.t, 6, len(cs.mCount))

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
