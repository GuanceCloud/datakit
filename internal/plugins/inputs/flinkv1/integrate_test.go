// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flinkv1

import (
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"os"
	"sync"
	T "testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	dt "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/prom"
	tu "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

type caseSpec struct {
	t *T.T

	name        string
	repo        string // docker name
	repoTag     string // docker tag
	tty         bool
	envs        []string
	servicePort string // port (rand)）

	optsJob  []inputs.PointCheckOption
	optsTask []inputs.PointCheckOption

	ipt    *prom.Input // This is real prom
	feeder *io.MockedFeeder

	pool     *dt.Pool
	resource *dt.Resource

	cr *tu.CaseResult // collect `go test -run` metric
}

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		measurement := string(pt.Name())
		var opts []inputs.PointCheckOption
		opts = append(opts, inputs.WithExtraTags(cs.ipt.Tags))

		switch measurement {
		case "flink_jobmanager":
			opts = append(opts, cs.optsJob...)
			opts = append(opts, inputs.WithDoc(&JobmanagerMeasurement{}))
			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

		case "flink_taskmanager":
			opts = append(opts, cs.optsTask...)
			opts = append(opts, inputs.WithDoc(&TaskmanagerMeasurement{}))
			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

		default: // TODO: check other measurement
			return nil
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
	// start remote image server
	r := tu.GetRemote()
	dockerTCP := r.TCPURL() // got "tcp://" + net.JoinHostPort(i.Host, i.Port)

	cs.t.Logf("get remote: %+#v, TCP: %s", r, dockerTCP)

	start := time.Now()

	p, err := dt.NewPool(dockerTCP)
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		cs.t.Logf("get hostname failed: %s, ignored", err)
		hostname = "unknown-hostname"
	}

	containerName := fmt.Sprintf("%s.%s", hostname, cs.name)

	// remove the container if exist.
	if err := p.RemoveContainerByName(containerName); err != nil {
		return err
	}

	resource, err := p.RunWithOptions(&dt.RunOptions{
		// specify container image & tag
		Repository: cs.repo,
		Tag:        cs.repoTag,

		Tty: cs.tty,

		// port binding
		PortBindings: map[docker.Port][]docker.PortBinding{
			"9250/tcp": {{HostIP: "0.0.0.0", HostPort: cs.servicePort}},
		},

		Name: containerName,

		// container run-time envs
		Env: cs.envs,
	}, func(c *docker.HostConfig) {
		c.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return err
	}

	cs.pool = p
	cs.resource = resource

	cs.t.Logf("check service(%s:%s)...", r.Host, cs.servicePort)
	if !r.PortOK(cs.servicePort, time.Minute) {
		return fmt.Errorf("service checking failed")
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
	if err := cs.checkPoint(pts); err != nil {
		return err
	}

	cs.t.Logf("stop input...")
	cs.ipt.Terminate()

	cs.t.Logf("exit...")
	wg.Wait()

	return nil
}

func buildCases(t *T.T) ([]*caseSpec, error) {
	t.Helper()

	remote := tu.GetRemote()

	bases := []struct {
		name     string
		repo     string // docker name
		repoTag  string // docker tag
		tty      bool
		conf     string
		optsJob  []inputs.PointCheckOption
		optsTask []inputs.PointCheckOption
	}{
		{
			name:    "remote",
			repo:    "pubrepo.jiagouyun.com/image-repo-for-testing/flink/flinkprom",
			repoTag: "1.12.0",
			tty:     true,
			conf: fmt.Sprintf(`
source = "flink"
metric_types = ["counter", "gauge"]
interval = "10s"
tls_open = false
url = "http://%s/metrics"

[[measurements]]
prefix = "flink_jobmanager_"
name = "flink_jobmanager"

[[measurements]]
prefix = "flink_taskmanager_"
name = "flink_taskmanager"

[tags]
  tag1 = "some_value"
  tag2 = "some_other_value"`, net.JoinHostPort(remote.Host, fmt.Sprintf("%d", tu.RandPort("tcp")))),
			optsJob: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"instance": "", "tag1": "", "tag2": ""}),
				inputs.WithOptionalTags(),
				inputs.WithOptionalFields("Status_JVM_CPU_Load", "Status_JVM_CPU_Time", "Status_JVM_ClassLoader_ClassesLoaded", "Status_JVM_ClassLoader_ClassesUnloaded", "Status_JVM_GarbageCollector_Copy_Count", "Status_JVM_GarbageCollector_Copy_Time", "Status_JVM_GarbageCollector_MarkSweepCompact_Count", "Status_JVM_GarbageCollector_MarkSweepCompact_Time", "Status_JVM_Memory_Direct_Count", "Status_JVM_Memory_Direct_MemoryUsed", "Status_JVM_Memory_Direct_TotalCapacity", "Status_JVM_Memory_Heap_Committed", "Status_JVM_Memory_Heap_Max", "Status_JVM_Memory_Heap_Used", "Status_JVM_Memory_Mapped_Count", "Status_JVM_Memory_Mapped_MemoryUsed", "Status_JVM_Memory_Mapped_TotalCapacity", "Status_JVM_Memory_Metaspace_Committed", "Status_JVM_Memory_Metaspace_Max", "Status_JVM_Memory_Metaspace_Used", "Status_JVM_Memory_NonHeap_Committed", "Status_JVM_Memory_NonHeap_Max", "Status_JVM_Memory_NonHeap_Used", "Status_JVM_Threads_Count", "numRegisteredTaskManagers", "numRunningJobs", "taskSlotsAvailable", "taskSlotsTotal", "Status_JVM_GarbageCollector_G1_Young_Generation_Time", "Status_JVM_GarbageCollector_G1_Old_Generation_Count", "Status_JVM_GarbageCollector_G1_Old_Generation_Time", "Status_JVM_GarbageCollector_G1_Young_Generation_Count"), // nolint:lll
			},
			optsTask: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"instance": "", "tag1": "", "tag2": ""}),
				inputs.WithOptionalTags("tm_id"),
				inputs.WithOptionalFields("Status_Flink_Memory_Managed_Total", "Status_Flink_Memory_Managed_Used", "Status_JVM_CPU_Load", "Status_JVM_CPU_Time", "Status_JVM_ClassLoader_ClassesLoaded", "Status_JVM_ClassLoader_ClassesUnloaded", "Status_JVM_GarbageCollector_G1_Old_Generation_Count", "Status_JVM_GarbageCollector_G1_Old_Generation_Time", "Status_JVM_GarbageCollector_G1_Young_Generation_Count", "Status_JVM_GarbageCollector_G1_Young_Generation_Time", "Status_JVM_Memory_Direct_Count", "Status_JVM_Memory_Direct_MemoryUsed", "Status_JVM_Memory_Direct_TotalCapacity", "Status_JVM_Memory_Heap_Committed", "Status_JVM_Memory_Heap_Max", "Status_JVM_Memory_Heap_Used", "Status_JVM_Memory_Mapped_Count", "Status_JVM_Memory_Mapped_MemoryUsed", "Status_JVM_Memory_Mapped_TotalCapacity", "Status_JVM_Memory_Metaspace_Committed", "Status_JVM_Memory_Metaspace_Max", "Status_JVM_Memory_Metaspace_Used", "Status_JVM_Memory_NonHeap_Committed", "Status_JVM_Memory_NonHeap_Max", "Status_JVM_Memory_NonHeap_Used", "Status_JVM_Threads_Count", "Status_Network_AvailableMemorySegments", "Status_Network_TotalMemorySegments", "Status_Shuffle_Netty_AvailableMemory", "Status_Shuffle_Netty_AvailableMemorySegments", "Status_Shuffle_Netty_TotalMemory", "Status_Shuffle_Netty_TotalMemorySegments", "Status_Shuffle_Netty_UsedMemory", "Status_Shuffle_Netty_UsedMemorySegments"), // nolint:lll
			},
		}, {
			name:    "remote",
			repo:    "pubrepo.jiagouyun.com/image-repo-for-testing/flink/flinkprom",
			repoTag: "1.17.0",
			tty:     true,
			conf: fmt.Sprintf(`
source = "flink"
metric_types = ["counter", "gauge"]
interval = "10s"
tls_open = false
url = "http://%s/metrics"

[[measurements]]
prefix = "flink_jobmanager_"
name = "flink_jobmanager"

[[measurements]]
prefix = "flink_taskmanager_"
name = "flink_taskmanager"

[tags]
  tag1 = "some_value"
  tag2 = "some_other_value"`, net.JoinHostPort(remote.Host, fmt.Sprintf("%d", tu.RandPort("tcp")))),
			optsJob: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"instance": "", "tag1": "", "tag2": ""}),
				inputs.WithOptionalTags(),
				inputs.WithOptionalFields("Status_JVM_CPU_Load", "Status_JVM_CPU_Time", "Status_JVM_ClassLoader_ClassesLoaded", "Status_JVM_ClassLoader_ClassesUnloaded", "Status_JVM_GarbageCollector_Copy_Count", "Status_JVM_GarbageCollector_Copy_Time", "Status_JVM_GarbageCollector_MarkSweepCompact_Count", "Status_JVM_GarbageCollector_MarkSweepCompact_Time", "Status_JVM_Memory_Direct_Count", "Status_JVM_Memory_Direct_MemoryUsed", "Status_JVM_Memory_Direct_TotalCapacity", "Status_JVM_Memory_Heap_Committed", "Status_JVM_Memory_Heap_Max", "Status_JVM_Memory_Heap_Used", "Status_JVM_Memory_Mapped_Count", "Status_JVM_Memory_Mapped_MemoryUsed", "Status_JVM_Memory_Mapped_TotalCapacity", "Status_JVM_Memory_Metaspace_Committed", "Status_JVM_Memory_Metaspace_Max", "Status_JVM_Memory_Metaspace_Used", "Status_JVM_Memory_NonHeap_Committed", "Status_JVM_Memory_NonHeap_Max", "Status_JVM_Memory_NonHeap_Used", "Status_JVM_Threads_Count", "numRegisteredTaskManagers", "numRunningJobs", "taskSlotsAvailable", "taskSlotsTotal", "Status_JVM_GarbageCollector_G1_Young_Generation_Time", "Status_JVM_GarbageCollector_G1_Old_Generation_Count", "Status_JVM_GarbageCollector_G1_Old_Generation_Time", "Status_JVM_GarbageCollector_G1_Young_Generation_Count"), // nolint:lll
			},
			optsTask: []inputs.PointCheckOption{
				inputs.WithTypeChecking(false),
				inputs.WithExtraTags(map[string]string{"instance": "", "tag1": "", "tag2": ""}),
				inputs.WithOptionalTags("tm_id"),
				inputs.WithOptionalFields("Status_Flink_Memory_Managed_Total", "Status_Flink_Memory_Managed_Used", "Status_JVM_CPU_Load", "Status_JVM_CPU_Time", "Status_JVM_ClassLoader_ClassesLoaded", "Status_JVM_ClassLoader_ClassesUnloaded", "Status_JVM_GarbageCollector_G1_Old_Generation_Count", "Status_JVM_GarbageCollector_G1_Old_Generation_Time", "Status_JVM_GarbageCollector_G1_Young_Generation_Count", "Status_JVM_GarbageCollector_G1_Young_Generation_Time", "Status_JVM_Memory_Direct_Count", "Status_JVM_Memory_Direct_MemoryUsed", "Status_JVM_Memory_Direct_TotalCapacity", "Status_JVM_Memory_Heap_Committed", "Status_JVM_Memory_Heap_Max", "Status_JVM_Memory_Heap_Used", "Status_JVM_Memory_Mapped_Count", "Status_JVM_Memory_Mapped_MemoryUsed", "Status_JVM_Memory_Mapped_TotalCapacity", "Status_JVM_Memory_Metaspace_Committed", "Status_JVM_Memory_Metaspace_Max", "Status_JVM_Memory_Metaspace_Used", "Status_JVM_Memory_NonHeap_Committed", "Status_JVM_Memory_NonHeap_Max", "Status_JVM_Memory_NonHeap_Used", "Status_JVM_Threads_Count", "Status_Network_AvailableMemorySegments", "Status_Network_TotalMemorySegments", "Status_Shuffle_Netty_AvailableMemory", "Status_Shuffle_Netty_AvailableMemorySegments", "Status_Shuffle_Netty_TotalMemory", "Status_Shuffle_Netty_TotalMemorySegments", "Status_Shuffle_Netty_UsedMemory", "Status_Shuffle_Netty_UsedMemorySegments", "Status_Shuffle_Netty_RequestedMemoryUsage"), // nolint:lll
			},
		},
	}

	// TODO: add per-image configs
	perImageCfgs := []interface{}{}
	_ = perImageCfgs

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := io.NewMockedFeeder()

		ipt := prom.NewProm() // This is real prom
		ipt.Feeder = feeder   // Flush metric data to testing_metrics

		// URL from ENV.
		_, err := toml.Decode(base.conf, ipt)
		assert.NoError(t, err)

		url, err := url.Parse(ipt.URL)
		assert.NoError(t, err)

		ipport, err := netip.ParseAddrPort(url.Host)
		assert.NoError(t, err, "parse %s failed: %s", ipt.URL, err)

		cases = append(cases, &caseSpec{
			t:      t,
			ipt:    ipt,
			name:   base.name,
			feeder: feeder,
			tty:    base.tty,

			repo:    base.repo,    // docker name
			repoTag: base.repoTag, // docker tag

			servicePort: fmt.Sprintf("%d", ipport.Port()),

			optsJob:  base.optsJob,
			optsTask: base.optsTask,

			// Test case result.
			cr: &tu.CaseResult{
				Name:        t.Name(),
				Case:        base.name,
				ExtraFields: map[string]any{},
				ExtraTags: map[string]string{
					"image":         base.repo,
					"image_tag":     base.repoTag,
					"remote_server": ipt.URL,
				},
			},
		})
	}
	return cases, nil
}

func TestIntegrate(t *T.T) {
	if !tu.CheckIntegrationTestingRunning() {
		t.Skip()
	}
	start := time.Now()
	cases, err := buildCases(t)
	if err != nil {
		cr := &tu.CaseResult{
			Name:          t.Name(),
			Status:        tu.TestPassed,
			FailedMessage: err.Error(),
			Cost:          time.Since(start),
		}

		_ = tu.Flush(cr)
		return
	}

	t.Logf("testing %d cases...", len(cases))

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			caseStart := time.Now()

			t.Logf("testing %s...", tc.name)

			// Run a test case.
			if err := tc.run(); err != nil {
				tc.cr.Status = tu.TestFailed
				tc.cr.FailedMessage = err.Error()

				assert.NoError(t, err)
			} else {
				tc.cr.Status = tu.TestPassed
			}

			tc.cr.Cost = time.Since(caseStart)

			assert.NoError(t, tu.Flush(tc.cr))

			t.Cleanup(func() {
				// clean remote docker resources
				if tc.resource == nil {
					return
				}

				tc.pool.Purge(tc.resource)
			})
		})
	}
}
