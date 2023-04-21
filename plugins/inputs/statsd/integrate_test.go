// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func TestStatsdInput(t *testing.T) {
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

				panic(err)
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
		name           string // Also used as build image name:tag.
		conf           string
		dockerFileText string // Empty if not build image.
		exposedPorts   []string
		opts           []inputs.PointCheckOption
	}{
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/java:jvm-ddtrace-statsd-8",
			conf: `protocol = "udp"
			service_address = ":58125"
			metric_separator = "_"
			drop_tags = ["runtime-id"]
			metric_mapping = [
			  "jvm_:jvm",
			  "datadog_tracer_:ddtrace",
			]
			delete_gauges = true
			delete_counters = true
			delete_sets = true
			delete_timings = true
			percentiles = [50.0, 90.0, 99.0, 99.9, 99.95, 100.0]
			parse_data_dog_tags = true
			datadog_extensions = true
			datadog_distributions = true
			allowed_pending_messages = 10000
			percentile_limit = 1000`,
			// exposedPorts: []string{"8080/tcp"},
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalFields("heap_memory", "heap_memory_committed", "heap_memory_init", "heap_memory_max", "non_heap_memory", "non_heap_memory_committed", "non_heap_memory_init", "non_heap_memory_max", "thread_count", "gc_cms_count", "gc_major_collection_count", "gc_minor_collection_count", "gc_parnew_time", "gc_major_collection_time", "gc_minor_collection_time", "os_open_file_descriptors", "gc_eden_size", "gc_old_gen_size", "buffer_pool_direct_used", "buffer_pool_direct_capacity", "cpu_load_system", "buffer_pool_mapped_capacity", "buffer_pool_mapped_count", "cpu_load_process", "gc_survivor_size", "buffer_pool_direct_count", "gc_metaspace_size", "loaded_classes", "buffer_pool_mapped_used"), //nolint:lll
				inputs.WithOptionalTags("name"), //nolint:lll
			},
		},

		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/java:jvm-ddtrace-statsd-11",
			conf: `protocol = "udp"
			service_address = ":58125"
			metric_separator = "_"
			drop_tags = ["runtime-id"]
			metric_mapping = [
			  "jvm_:jvm",
			  "datadog_tracer_:ddtrace",
			]
			delete_gauges = true
			delete_counters = true
			delete_sets = true
			delete_timings = true
			percentiles = [50.0, 90.0, 99.0, 99.9, 99.95, 100.0]
			parse_data_dog_tags = true
			datadog_extensions = true
			datadog_distributions = true
			allowed_pending_messages = 10000
			percentile_limit = 1000`,
			// exposedPorts: []string{"8080/tcp"},
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalFields("heap_memory", "heap_memory_committed", "heap_memory_init", "heap_memory_max", "non_heap_memory", "non_heap_memory_committed", "non_heap_memory_init", "non_heap_memory_max", "thread_count", "gc_cms_count", "gc_major_collection_count", "gc_minor_collection_count", "gc_parnew_time", "gc_major_collection_time", "gc_minor_collection_time", "os_open_file_descriptors", "gc_eden_size", "gc_old_gen_size", "buffer_pool_direct_used", "buffer_pool_direct_capacity", "cpu_load_system", "buffer_pool_mapped_capacity", "buffer_pool_mapped_count", "cpu_load_process", "gc_survivor_size", "buffer_pool_direct_count", "gc_metaspace_size", "loaded_classes", "buffer_pool_mapped_used"), //nolint:lll
				inputs.WithOptionalTags("name"), //nolint:lll
			},
		},

		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/java:jvm-ddtrace-statsd-17",
			conf: `protocol = "udp"
			service_address = ":58125"
			metric_separator = "_"
			drop_tags = ["runtime-id"]
			metric_mapping = [
			  "jvm_:jvm",
			  "datadog_tracer_:ddtrace",
			]
			delete_gauges = true
			delete_counters = true
			delete_sets = true
			delete_timings = true
			percentiles = [50.0, 90.0, 99.0, 99.9, 99.95, 100.0]
			parse_data_dog_tags = true
			datadog_extensions = true
			datadog_distributions = true
			allowed_pending_messages = 10000
			percentile_limit = 1000`,
			// exposedPorts: []string{"8080/tcp"},
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalFields("heap_memory", "heap_memory_committed", "heap_memory_init", "heap_memory_max", "non_heap_memory", "non_heap_memory_committed", "non_heap_memory_init", "non_heap_memory_max", "thread_count", "gc_cms_count", "gc_major_collection_count", "gc_minor_collection_count", "gc_parnew_time", "gc_major_collection_time", "gc_minor_collection_time", "os_open_file_descriptors", "gc_eden_size", "gc_old_gen_size", "buffer_pool_direct_used", "buffer_pool_direct_capacity", "cpu_load_system", "buffer_pool_mapped_capacity", "buffer_pool_mapped_count", "cpu_load_process", "gc_survivor_size", "buffer_pool_direct_count", "gc_metaspace_size", "loaded_classes", "buffer_pool_mapped_used"), //nolint:lll
				inputs.WithOptionalTags("name"), //nolint:lll
			},
		},

		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/java:jvm-ddtrace-statsd-20",
			conf: `protocol = "udp"
			service_address = ":58125"
			metric_separator = "_"
			drop_tags = ["runtime-id"]
			metric_mapping = [
			  "jvm_:jvm",
			  "datadog_tracer_:ddtrace",
			]
			delete_gauges = true
			delete_counters = true
			delete_sets = true
			delete_timings = true
			percentiles = [50.0, 90.0, 99.0, 99.9, 99.95, 100.0]
			parse_data_dog_tags = true
			datadog_extensions = true
			datadog_distributions = true
			allowed_pending_messages = 10000
			percentile_limit = 1000`,
			// exposedPorts: []string{"8080/tcp"},
			opts: []inputs.PointCheckOption{
				inputs.WithOptionalFields("heap_memory", "heap_memory_committed", "heap_memory_init", "heap_memory_max", "non_heap_memory", "non_heap_memory_committed", "non_heap_memory_init", "non_heap_memory_max", "thread_count", "gc_cms_count", "gc_major_collection_count", "gc_minor_collection_count", "gc_parnew_time", "gc_major_collection_time", "gc_minor_collection_time", "os_open_file_descriptors", "gc_eden_size", "gc_old_gen_size", "buffer_pool_direct_used", "buffer_pool_direct_capacity", "cpu_load_system", "buffer_pool_mapped_capacity", "buffer_pool_mapped_count", "cpu_load_process", "gc_survivor_size", "buffer_pool_direct_count", "gc_metaspace_size", "loaded_classes", "buffer_pool_mapped_used"), //nolint:lll
				inputs.WithOptionalTags("name"), //nolint:lll
			},
		},
	}

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := dkio.NewMockedFeeder()

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

	ipt    *input
	feeder *dkio.MockedFeeder

	pool     *dockertest.Pool
	resource *dockertest.Resource

	cr *testutils.CaseResult
}

var errorMsgs []string

type FeedMeasurementBody []struct {
	Measurement string                 `json:"measurement"`
	Tags        map[string]string      `json:"tags"`
	Fields      map[string]interface{} `json:"fields"`
}

////////////////////////////////////////////////////////////////////////////////

type jvmMeasurement struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

// Point implement MeasurementV2.
func (m *jvmMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(dkpt.GlobalElectionTags()))
	} else {
		opts = append(opts, point.WithExtraTags(dkpt.GlobalHostTags()))
	}

	return point.NewPointV2([]byte(m.name),
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (j *jvmMeasurement) LineProto() (*dkpt.Point, error) {
	// return point.NewPoint(j.name, j.tags, j.fields, point.MOpt())
	return nil, fmt.Errorf("not implement")
}

// From: https://docs.datadoghq.com/tracing/metrics/runtime_metrics/java/#data-collected
func (j *jvmMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "jvm",
		Fields: map[string]interface{}{
			"heap_memory":                 &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"heap_memory_committed":       &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"heap_memory_init":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"heap_memory_max":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"non_heap_memory":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"non_heap_memory_committed":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"non_heap_memory_init":        &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"non_heap_memory_max":         &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"thread_count":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"gc_cms_count":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Count, Unit: inputs.NCount, Desc: ""},
			"gc_major_collection_count":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"gc_minor_collection_count":   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"gc_parnew_time":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"gc_major_collection_time":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"gc_minor_collection_time":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"os_open_file_descriptors":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"gc_eden_size":                &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"gc_old_gen_size":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"buffer_pool_direct_used":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"buffer_pool_direct_capacity": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"cpu_load_system":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"buffer_pool_mapped_capacity": &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"buffer_pool_mapped_count":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"cpu_load_process":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"gc_survivor_size":            &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"buffer_pool_direct_count":    &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"gc_metaspace_size":           &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"loaded_classes":              &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
			"buffer_pool_mapped_used":     &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.NCount, Desc: ""},
		},
		Tags: map[string]interface{}{
			"env":         inputs.TagInfo{Desc: ""},
			"instance":    inputs.TagInfo{Desc: ""},
			"jmx_domain":  inputs.TagInfo{Desc: ""},
			"metric_type": inputs.TagInfo{Desc: ""},
			"name":        inputs.TagInfo{Desc: ""},
			"service":     inputs.TagInfo{Desc: ""},
			"type":        inputs.TagInfo{Desc: ""},
			"version":     inputs.TagInfo{Desc: ""},
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	var opts []inputs.PointCheckOption
	opts = append(opts, inputs.WithExtraTags(cs.ipt.Tags))
	opts = append(opts, cs.opts...)

	for _, pt := range pts {
		measurement := string(pt.Name())

		switch measurement {
		case "jvm":
			opts = append(opts, inputs.WithDoc(&jvmMeasurement{}))

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
	cs.t.Helper()

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

	extIP, err := externalIP()
	if err != nil {
		return err
	}

	var resource *dockertest.Resource

	if len(cs.dockerFileText) == 0 {
		// Just run a container from existing docker image.
		resource, err = p.RunWithOptions(
			&dockertest.RunOptions{
				Name: containerName, // ATTENTION: not cs.name.

				Repository: cs.repo,
				Tag:        cs.repoTag,
				Env:        []string{fmt.Sprintf("DATAKIT_HOST=%s", extIP), "UDP_PORT=58125"},

				ExposedPorts: cs.exposedPorts,
				PortBindings: cs.getPortBindings(),
			},

			func(c *docker.HostConfig) {
				c.RestartPolicy = docker.RestartPolicy{Name: "no"}
				// c.AutoRemove = true
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
				Env:        []string{fmt.Sprintf("DATAKIT_HOST=%s", extIP), "UDP_PORT=58125"},

				ExposedPorts: cs.exposedPorts,
				PortBindings: cs.getPortBindings(),
			},

			func(c *docker.HostConfig) {
				c.RestartPolicy = docker.RestartPolicy{Name: "no"}
				// c.AutoRemove = true
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
	cs.t.Logf("waiting points, 5 minutes timeout...")
	pts, err := cs.feeder.NPoints(50, 5*time.Minute)
	if err != nil {
		return err
	}

	cs.cr.AddField("point_latency", int64(time.Since(start)))
	cs.cr.AddField("point_count", len(pts))

	cs.t.Logf("get %d points", len(pts))
	if err := cs.checkPoint(pts); err != nil {
		return err
	}

	if len(errorMsgs) > 0 {
		return fmt.Errorf("errorMsgs: %#v", errorMsgs)
	}
	errorMsgs = errorMsgs[:0]

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

////////////////////////////////////////////////////////////////////////////////

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}
