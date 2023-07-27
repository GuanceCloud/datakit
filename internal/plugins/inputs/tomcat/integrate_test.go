// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tomcat

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/statsd"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
)

// ATTENTION: Docker version should use v20.10.18 in integrate tests. Other versions are not tested.

const (
	jvmMetricName = "jvm"
)

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

func buildCases(t *testing.T) ([]*caseSpec, error) {
	t.Helper()

	remote := testutils.GetRemote()

	bases := []struct {
		name           string // Also used as build image name:tag.
		conf           string
		dockerFileText string // Empty if not build image.
		exposedPorts   []string
		mPathCount     map[string]int
		nCountExpect   int
		optsJVM        []inputs.PointCheckOption
		optsTomcat     []inputs.PointCheckOption
	}{
		////////////////////////////////////////////////////////////////////////
		// Tomcat 8
		////////////////////////////////////////////////////////////////////////
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/tomcat:8.5.90-ddtrace",
			conf: `protocol = "udp"
			service_address = ":"
			metric_separator = "_"
			drop_tags = [""]
			metric_mapping = [
			  "jvm_:jvm",
			  "datadog_tracer_:ddtrace",
			  "tomcat_:tomcat",
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
			percentile_limit = 1000`, // set conf address later.
			exposedPorts: []string{"8080/tcp"},
			mPathCount: map[string]int{
				"/": 10,
			},
			nCountExpect: 2,
			optsJVM: []inputs.PointCheckOption{
				inputs.WithOptionalFields(
					"heap_memory",
					"heap_memory_committed",
					"heap_memory_init",
					"heap_memory_max",
					"non_heap_memory",
					"non_heap_memory_committed",
					"non_heap_memory_init",
					"non_heap_memory_max",
					"thread_count",
					"gc_cms_count",
					"gc_major_collection_count",
					"gc_minor_collection_count",
					"gc_parnew_time",
					"gc_major_collection_time",
					"gc_minor_collection_time",
					"os_open_file_descriptors",
					"gc_eden_size",
					"gc_old_gen_size",
					"buffer_pool_direct_used",
					"buffer_pool_direct_capacity",
					"cpu_load_system",
					"buffer_pool_mapped_capacity",
					"buffer_pool_mapped_count",
					"cpu_load_process",
					"gc_survivor_size",
					"buffer_pool_direct_count",
					"gc_metaspace_size",
					"loaded_classes",
					"buffer_pool_mapped_used",
				),
				inputs.WithOptionalTags(
					"env",
					"name",
					"runtime-id",
					"version",
				),
			},
			optsTomcat: []inputs.PointCheckOption{
				inputs.WithOptionalFields(
					"bytes_rcvd",
					"bytes_sent",
					"cache_access_count",
					"cache_hits_count",
					"error_count",
					"jsp_count",
					"jsp_reload_count",
					"max_time",
					"processing_time",
					"request_count",
					"servlet_error_count",
					"servlet_processing_time",
					"servlet_request_count",
					"string_cache_access_count",
					"string_cache_hit_count",
					"threads_busy",
					"threads_count",
					"threads_max",
					"web_cache_hit_count",
					"web_cache_lookup_count",
					"string_cache_access_count",
					"threads_busy",
					"threads_count",
				),
				inputs.WithOptionalTags(
					"name",
					"runtime-id",
				),
			},
		},

		////////////////////////////////////////////////////////////////////////
		// Tomcat 9
		////////////////////////////////////////////////////////////////////////
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/tomcat:9.0.76-ddtrace",
			conf: `protocol = "udp"
			service_address = ":"
			metric_separator = "_"
			drop_tags = [""]
			metric_mapping = [
			  "jvm_:jvm",
			  "datadog_tracer_:ddtrace",
			  "tomcat_:tomcat",
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
			percentile_limit = 1000`, // set conf address later.
			exposedPorts: []string{"8080/tcp"},
			mPathCount: map[string]int{
				"/": 10,
			},
			nCountExpect: 2,
			optsJVM: []inputs.PointCheckOption{
				inputs.WithOptionalFields(
					"heap_memory",
					"heap_memory_committed",
					"heap_memory_init",
					"heap_memory_max",
					"non_heap_memory",
					"non_heap_memory_committed",
					"non_heap_memory_init",
					"non_heap_memory_max",
					"thread_count",
					"gc_cms_count",
					"gc_major_collection_count",
					"gc_minor_collection_count",
					"gc_parnew_time",
					"gc_major_collection_time",
					"gc_minor_collection_time",
					"os_open_file_descriptors",
					"gc_eden_size",
					"gc_old_gen_size",
					"buffer_pool_direct_used",
					"buffer_pool_direct_capacity",
					"cpu_load_system",
					"buffer_pool_mapped_capacity",
					"buffer_pool_mapped_count",
					"cpu_load_process",
					"gc_survivor_size",
					"buffer_pool_direct_count",
					"gc_metaspace_size",
					"loaded_classes",
					"buffer_pool_mapped_used",
				),
				inputs.WithOptionalTags(
					"env",
					"name",
					"runtime-id",
					"version",
				),
			},
			optsTomcat: []inputs.PointCheckOption{
				inputs.WithOptionalFields(
					"bytes_rcvd",
					"bytes_sent",
					"cache_access_count",
					"cache_hits_count",
					"error_count",
					"jsp_count",
					"jsp_reload_count",
					"max_time",
					"processing_time",
					"request_count",
					"servlet_error_count",
					"servlet_processing_time",
					"servlet_request_count",
					"string_cache_access_count",
					"string_cache_hit_count",
					"threads_busy",
					"threads_count",
					"threads_max",
					"web_cache_hit_count",
					"web_cache_lookup_count",
					"string_cache_access_count",
					"threads_busy",
					"threads_count",
				),
				inputs.WithOptionalTags(
					"name",
					"runtime-id",
				),
			},
		},

		////////////////////////////////////////////////////////////////////////
		// Tomcat 10
		////////////////////////////////////////////////////////////////////////
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/tomcat:10.1.10-ddtrace",
			conf: `protocol = "udp"
			service_address = ":"
			metric_separator = "_"
			drop_tags = [""]
			metric_mapping = [
			  "jvm_:jvm",
			  "datadog_tracer_:ddtrace",
			  "tomcat_:tomcat",
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
			percentile_limit = 1000`, // set conf address later.
			exposedPorts: []string{"8080/tcp"},
			mPathCount: map[string]int{
				"/": 10,
			},
			nCountExpect: 2,
			optsJVM: []inputs.PointCheckOption{
				inputs.WithOptionalFields(
					"heap_memory",
					"heap_memory_committed",
					"heap_memory_init",
					"heap_memory_max",
					"non_heap_memory",
					"non_heap_memory_committed",
					"non_heap_memory_init",
					"non_heap_memory_max",
					"thread_count",
					"gc_cms_count",
					"gc_major_collection_count",
					"gc_minor_collection_count",
					"gc_parnew_time",
					"gc_major_collection_time",
					"gc_minor_collection_time",
					"os_open_file_descriptors",
					"gc_eden_size",
					"gc_old_gen_size",
					"buffer_pool_direct_used",
					"buffer_pool_direct_capacity",
					"cpu_load_system",
					"buffer_pool_mapped_capacity",
					"buffer_pool_mapped_count",
					"cpu_load_process",
					"gc_survivor_size",
					"buffer_pool_direct_count",
					"gc_metaspace_size",
					"loaded_classes",
					"buffer_pool_mapped_used",
				),
				inputs.WithOptionalTags(
					"env",
					"name",
					"runtime-id",
					"version",
				),
			},
			optsTomcat: []inputs.PointCheckOption{
				inputs.WithOptionalFields(
					"bytes_rcvd",
					"bytes_sent",
					"cache_access_count",
					"cache_hits_count",
					"error_count",
					"jsp_count",
					"jsp_reload_count",
					"max_time",
					"processing_time",
					"request_count",
					"servlet_error_count",
					"servlet_processing_time",
					"servlet_request_count",
					"string_cache_access_count",
					"string_cache_hit_count",
					"threads_busy",
					"threads_count",
					"threads_max",
					"web_cache_hit_count",
					"web_cache_lookup_count",
					"string_cache_access_count",
					"threads_busy",
					"threads_count",
				),
				inputs.WithOptionalTags(
					"name",
					"runtime-id",
				),
			},
		},

		////////////////////////////////////////////////////////////////////////
		// Tomcat 11
		////////////////////////////////////////////////////////////////////////
		{
			name: "pubrepo.jiagouyun.com/image-repo-for-testing/tomcat:11.0.0-ddtrace",
			conf: `protocol = "udp"
			service_address = ":"
			metric_separator = "_"
			drop_tags = [""]
			metric_mapping = [
			  "jvm_:jvm",
			  "datadog_tracer_:ddtrace",
			  "tomcat_:tomcat",
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
			percentile_limit = 1000`, // set conf address later.
			exposedPorts: []string{"8080/tcp"},
			mPathCount: map[string]int{
				"/": 10,
			},
			nCountExpect: 2,
			optsJVM: []inputs.PointCheckOption{
				inputs.WithOptionalFields(
					"heap_memory",
					"heap_memory_committed",
					"heap_memory_init",
					"heap_memory_max",
					"non_heap_memory",
					"non_heap_memory_committed",
					"non_heap_memory_init",
					"non_heap_memory_max",
					"thread_count",
					"gc_cms_count",
					"gc_major_collection_count",
					"gc_minor_collection_count",
					"gc_parnew_time",
					"gc_major_collection_time",
					"gc_minor_collection_time",
					"os_open_file_descriptors",
					"gc_eden_size",
					"gc_old_gen_size",
					"buffer_pool_direct_used",
					"buffer_pool_direct_capacity",
					"cpu_load_system",
					"buffer_pool_mapped_capacity",
					"buffer_pool_mapped_count",
					"cpu_load_process",
					"gc_survivor_size",
					"buffer_pool_direct_count",
					"gc_metaspace_size",
					"loaded_classes",
					"buffer_pool_mapped_used",
				),
				inputs.WithOptionalTags(
					"env",
					"name",
					"runtime-id",
					"version",
				),
			},
			optsTomcat: []inputs.PointCheckOption{
				inputs.WithOptionalFields(
					"bytes_rcvd",
					"bytes_sent",
					"cache_access_count",
					"cache_hits_count",
					"error_count",
					"jsp_count",
					"jsp_reload_count",
					"max_time",
					"processing_time",
					"request_count",
					"servlet_error_count",
					"servlet_processing_time",
					"servlet_request_count",
					"string_cache_access_count",
					"string_cache_hit_count",
					"threads_busy",
					"threads_count",
					"threads_max",
					"web_cache_hit_count",
					"web_cache_lookup_count",
					"string_cache_access_count",
					"threads_busy",
					"threads_count",
				),
				inputs.WithOptionalTags(
					"name",
					"runtime-id",
				),
			},
		},
	}

	var cases []*caseSpec

	// compose cases
	for _, base := range bases {
		feeder := dkio.NewMockedFeeder()

		ipt := statsd.DefaultInput()
		ipt.Feeder = feeder

		_, err := toml.Decode(base.conf, ipt)
		require.NoError(t, err)

		conn, randPort, err := testutils.RandPortUDP()
		require.NoError(t, err)
		randPortStr := fmt.Sprintf("%d", randPort)
		ipt.ServiceAddress += randPortStr // :8125
		ipt.UDPlistener = conn

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
			serverPorts:    []string{randPortStr},
			mPathCount:     base.mPathCount,
			nCountExpect:   base.nCountExpect,
			optsJVM:        base.optsJVM,
			optsTomcat:     base.optsTomcat,

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
	optsJVM        []inputs.PointCheckOption
	optsTomcat     []inputs.PointCheckOption
	mPathCount     map[string]int
	mCount         map[string]struct{}
	nCountExpect   int

	ipt    *statsd.Input
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
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// Point implement MeasurementV2.
func (m *jvmMeasurement) Point() *point.Point {
	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTime(m.ts))

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
		Name: jvmMetricName,
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
			"runtime-id":  inputs.TagInfo{Desc: ""},
			"service":     inputs.TagInfo{Desc: ""},
			"type":        inputs.TagInfo{Desc: ""},
			"version":     inputs.TagInfo{Desc: ""},
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

func (cs *caseSpec) checkPoint(pts []*point.Point) error {
	for _, pt := range pts {
		var opts []inputs.PointCheckOption
		opts = append(opts, inputs.WithExtraTags(cs.ipt.Tags))

		measurement := string(pt.Name())

		switch measurement {
		case jvmMetricName:
			opts = append(opts, cs.optsJVM...)
			opts = append(opts, inputs.WithDoc(&jvmMeasurement{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[jvmMetricName] = struct{}{}

		case inputName:
			opts = append(opts, cs.optsTomcat...)
			opts = append(opts, inputs.WithDoc(&TomcatM{}))

			msgs := inputs.CheckPoint(pt, opts...)

			for _, msg := range msgs {
				cs.t.Logf("check measurement %s failed: %+#v", measurement, msg)
			}

			// TODO: error here
			if len(msgs) > 0 {
				return fmt.Errorf("check measurement %s failed: %+#v", measurement, msgs)
			}

			cs.mCount[inputName] = struct{}{}

		case "ddtrace": // ignore.
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
	cs.t.Helper()

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

				Repository:   cs.repo,
				Tag:          cs.repoTag,
				Env:          []string{fmt.Sprintf("DATAKIT_HOST=%s", extIP), "UDP_PORT=" + cs.serverPorts[0]},
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

				Repository:   cs.repo,
				Tag:          cs.repoTag,
				Env:          []string{fmt.Sprintf("DATAKIT_HOST=%s", extIP), "UDP_PORT=" + cs.serverPorts[0]},
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

	cs.t.Logf("check service(%s:%v)...", r.Host, cs.exposedPorts)

	if err := cs.portsOK(r); err != nil {
		return err
	}

	cs.t.Logf("listening: %v, remote = %s", cs.serverPorts, r.Host)

	cs.cr.AddField("container_ready_cost", int64(time.Since(start)))

	cs.runHTTPTests(r)

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
	timeout := 10 * time.Minute
	cs.t.Logf("waiting points, %v timeout...", timeout)
	pts, err := cs.feeder.NPoints(100, timeout)
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

	if len(errorMsgs) > 0 {
		return fmt.Errorf("errorMsgs: %#v", errorMsgs)
	}
	errorMsgs = errorMsgs[:0]

	cs.t.Logf("stop input...")
	cs.ipt.Terminate()

	require.Equal(cs.t, cs.nCountExpect, len(cs.mCount))

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

// Launch large amount of HTTP requests to remote web server.
func (cs *caseSpec) runHTTPTests(r *testutils.RemoteInfo) {
	for _, v := range cs.serverPorts {
		for path, count := range cs.mPathCount {
			newURL := fmt.Sprintf("http://%s%s", net.JoinHostPort(r.Host, v), path)
			fmt.Printf("start GET: %s\n", newURL)

			if cs.runHTTPWithTimeout(newURL, count) {
				break
			}
		}
	}
}

// runHTTPWithTimeout returns true if HTTP request succeeded.
func (cs *caseSpec) runHTTPWithTimeout(newURL string, count int) bool {
	done := make(chan struct{})

	iter := time.NewTicker(time.Second)
	defer iter.Stop()

	timeout := time.NewTicker(2 * time.Minute)
	defer timeout.Stop()

	var num int32

	for {
		select {
		case <-iter.C:
			for i := 0; i < count; i++ {
				go func() {
					netTransport := &http.Transport{
						Dial: (&net.Dialer{
							Timeout: 10 * time.Second,
						}).Dial,
						TLSHandshakeTimeout: 10 * time.Second,
					}
					netClient := &http.Client{
						Timeout:   time.Second * 20,
						Transport: netTransport,
					}

					resp, err := netClient.Get(newURL)
					if err != nil {
						fmt.Printf("HTTP GET failed: %v\n", err)
						return
					}
					defer resp.Body.Close()

					// HTTP request succeeded.
					done <- struct{}{}
				}()
			}

		case <-timeout.C:
			return false

		case <-done:
			if val := atomic.AddInt32(&num, 1); val >= int32(count) {
				return true
			}
		}
	}
}
