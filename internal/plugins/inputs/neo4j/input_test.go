// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package neo4j scrape neo4j exporter metrics.
package neo4j

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
)

func Test_basic(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		body := `# HELP neo4j_dbms_page_cache_usage_ratio Generated from Dropwizard metric import (metric=neo4j.dbms.page_cache.usage_ratio, type=com.neo4j.metrics.source.db.PageCacheMetrics$$Lambda$2430/0x00007f39d8b2ba48)
# TYPE neo4j_dbms_page_cache_usage_ratio gauge
neo4j_dbms_page_cache_usage_ratio 0.0
`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		t.Log(srv.URL)
		defer srv.Close()

		t.Logf("url: %s", srv.URL)

		inp := NewInput()
		inp.URLs = []string{srv.URL}

		inp.tagger = &taggerMock{
			hostTags: map[string]string{
				"host":  "foo",
				"hello": "world",
			},

			electionTags: map[string]string{
				"project": "foo",
				"cluster": "bar",
			},
		}

		inp.setup()

		pts, err := inp.getPts()
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			assert.True(t, pt.Tags().Has("instance"))
			assert.True(t, pt.Tags().Has("project"))
			assert.True(t, pt.Tags().Has("cluster"))

			assert.Equal(t, float64(0.0), pt.Get("dbms_page_cache_usage_ratio").(float64))

			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run("disable_instance_tag = true", func(t *testing.T) {
		body := `# HELP neo4j_dbms_page_cache_usage_ratio Generated from Dropwizard metric import (metric=neo4j.dbms.page_cache.usage_ratio, type=com.neo4j.metrics.source.db.PageCacheMetrics$$Lambda$2430/0x00007f39d8b2ba48)
# TYPE neo4j_dbms_page_cache_usage_ratio gauge
neo4j_dbms_page_cache_usage_ratio 0.0
`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		t.Log(srv.URL)
		defer srv.Close()

		t.Logf("url: %s", srv.URL)

		inp := NewInput()
		inp.URLs = []string{srv.URL}
		inp.DisableInstanceTag = true

		inp.setup()

		pts, err := inp.getPts()
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			assert.False(t, pt.Tags().Has("instance"))

			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run("one tag", func(t *testing.T) {
		body := `# HELP neo4j_dbms_pool_bolt_total_size Generated from Dropwizard metric import (metric=neo4j.dbms.pool.bolt.total_size, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$2496/0x00007f39d8b386a8)
# TYPE neo4j_dbms_pool_bolt_total_size gauge
neo4j_dbms_pool_bolt_total_size 0
`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		t.Log(srv.URL)
		defer srv.Close()

		t.Logf("url: %s", srv.URL)

		inp := NewInput()
		inp.URLs = []string{srv.URL}

		inp.setup()

		pts, err := inp.getPts()
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			assert.Equal(t, pt.GetTag("pool"), "bolt")

			assert.Equal(t, float64(0.0), pt.Get("dbms_pool_total_size").(float64))

			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run("some tag", func(t *testing.T) {
		body := `# HELP neo4j_database_system_pool_transaction_system_used_heap Generated from Dropwizard metric import (metric=neo4j.database.system.pool.transaction.system.used_heap, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$2100/0x00007f39d8a63428)
# TYPE neo4j_database_system_pool_transaction_system_used_heap gauge
neo4j_database_system_pool_transaction_system_used_heap 0.0
`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		t.Log(srv.URL)
		defer srv.Close()

		t.Logf("url: %s", srv.URL)

		inp := NewInput()
		inp.URLs = []string{srv.URL}

		inp.setup()

		pts, err := inp.getPts()
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			assert.Equal(t, pt.GetTag("db"), "system")
			assert.Equal(t, pt.GetTag("pool"), "transaction")
			assert.Equal(t, pt.GetTag("database"), "system")

			assert.Equal(t, float64(0.0), pt.Get("database_pool_used_heap").(float64))

			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run("tag have _", func(t *testing.T) {
		body := `# HELP neo4j_dbms_vm_gc_time_g1_young_generation_total Generated from Dropwizard metric import (metric=neo4j.dbms.vm.gc.time.g1_young_generation, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_dbms_vm_gc_time_g1_young_generation_total counter
neo4j_dbms_vm_gc_time_g1_young_generation_total 71.0
`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		t.Log(srv.URL)
		defer srv.Close()

		t.Logf("url: %s", srv.URL)

		inp := NewInput()
		inp.URLs = []string{srv.URL}

		inp.setup()

		pts, err := inp.getPts()
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			assert.Equal(t, pt.GetTag("gc"), "g1_young_generation")

			assert.Equal(t, float64(71.0), pt.Get("dbms_vm_gc_time_total").(float64))

			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run("tag in head", func(t *testing.T) {
		body := `# HELP neo4j_neo4j_check_point_total_time_total Generated from Dropwizard metric import (metric=neo4j.neo4j.check_point.total_time, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_neo4j_check_point_total_time_total counter
neo4j_neo4j_check_point_total_time_total 0.0
`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		t.Log(srv.URL)
		defer srv.Close()

		t.Logf("url: %s", srv.URL)

		inp := NewInput()
		inp.URLs = []string{srv.URL}

		inp.setup()

		pts, err := inp.getPts()
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			assert.Equal(t, pt.GetTag("db"), "neo4j")

			assert.Equal(t, float64(0.0), pt.Get("check_point_total_time_total").(float64))

			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run("tag in tail", func(t *testing.T) {
		body := `# HELP neo4j_vm_memory_pool_g1_eden_space Generated from Dropwizard metric import (metric=neo4j.vm.memory.pool.g1_eden_space, type=com.neo4j.metrics.source.jvm.JVMMemoryPoolMetrics$$Lambda$1424/0x000000084099b840)
# TYPE neo4j_vm_memory_pool_g1_eden_space gauge
neo4j_vm_memory_pool_g1_eden_space 0.0
`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		t.Log(srv.URL)
		defer srv.Close()

		t.Logf("url: %s", srv.URL)

		inp := NewInput()
		inp.URLs = []string{srv.URL}

		inp.setup()

		pts, err := inp.getPts()
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			assert.Equal(t, pt.GetTag("pool"), "g1_eden_space")

			assert.Equal(t, float64(0.0), pt.Get("vm_memory_pool").(float64))

			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run("metric name not be neo4j", func(t *testing.T) {
		body := `# HELP vm_memory_buffer_direct_capacity Generated from Dropwizard metric import (metric=vm.memory.buffer.direct.capacity, type=org.neo4j.metrics.source.jvm.MemoryBuffersMetrics$$Lambda$424/1242427797)
# TYPE vm_memory_buffer_direct_capacity gauge
vm_memory_buffer_direct_capacity 221184.0
`
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, body)
		}))
		t.Log(srv.URL)
		defer srv.Close()

		t.Logf("url: %s", srv.URL)

		inp := NewInput()
		inp.URLs = []string{srv.URL}

		inp.setup()

		pts, err := inp.getPts()
		assert.NoError(t, err)

		if len(pts) == 0 {
			t.Errorf("got nil pts error.")
		}

		for _, pt := range pts {
			assert.Equal(t, pt.GetTag("bufferpool"), "direct")
			assert.Equal(t, float64(221184.0), pt.Get("vm_memory_buffer_capacity").(float64))

			t.Logf("%s", pt.Pretty())
		}
	})
}

func Test_collect(t *testing.T) {
	type fields struct {
		Interval               time.Duration
		Timeout                time.Duration
		ConnectKeepAlive       time.Duration
		URLs                   []string
		IgnoreReqErr           bool
		MetricTypes            []string
		MetricNameFilter       []string
		MetricNameFilterIgnore []string
		MeasurementPrefix      string
		MeasurementName        string
		Measurements           []iprom.Rule
		Output                 string
		MaxFileSize            int64
		TLSOpen                bool
		UDSPath                string
		CacertFile             string
		CertFile               string
		KeyFile                string
		TagsIgnore             []string
		TagsRename             *iprom.RenameTags
		AsLogging              *iprom.AsLogging
		IgnoreTagKV            map[string][]string
		HTTPHeaders            map[string]string
		Tags                   map[string]string
		DisableHostTag         bool
		DisableInstanceTag     bool
		DisableInfoTag         bool
		Auth                   map[string]string
		pm                     *iprom.Prom
		Feeder                 io.Feeder
		Election               bool
		pauseCh                chan bool
		pause                  bool
		Tagger                 datakit.GlobalTagger
		urls                   []*url.URL
		semStop                *cliutils.Sem
		mergedTags             map[string]urlTags
		l                      *logger.Logger
	}
	tests := []struct {
		name     string
		fields   fields
		mockData string
		want     []string
		wantErr  bool
	}{
		{
			name:     "neo4j v 5.11.0",
			mockData: mock5_11_0,
			want:     want5_11_0,
		},
		{
			name:     "neo4j v 4.4.0",
			mockData: mock4_4_0,
			want:     want4_4_0,
		},
		{
			name:     "neo4j v 3.4.0",
			mockData: mock3_4_0,
			want:     want3_4_0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields = fields{
				Election:           true,
				pauseCh:            make(chan bool, maxPauseCh),
				Tags:               make(map[string]string),
				DisableInstanceTag: true,
				mergedTags:         map[string]urlTags{},
				Tagger:             datakit.DefaultGlobalTagger(),
				l:                  logger.SLogger(inputName),
			}

			inp := &Input{
				Interval:               tt.fields.Interval,
				Timeout:                tt.fields.Timeout,
				ConnectKeepAlive:       tt.fields.ConnectKeepAlive,
				URLs:                   tt.fields.URLs,
				IgnoreReqErr:           tt.fields.IgnoreReqErr,
				MetricTypes:            tt.fields.MetricTypes,
				MetricNameFilter:       tt.fields.MetricNameFilter,
				MetricNameFilterIgnore: tt.fields.MetricNameFilterIgnore,
				MeasurementPrefix:      tt.fields.MeasurementPrefix,
				MeasurementName:        tt.fields.MeasurementName,
				Measurements:           tt.fields.Measurements,
				Output:                 tt.fields.Output,
				MaxFileSize:            tt.fields.MaxFileSize,
				TLSOpen:                tt.fields.TLSOpen,
				UDSPath:                tt.fields.UDSPath,
				CacertFile:             tt.fields.CacertFile,
				CertFile:               tt.fields.CertFile,
				KeyFile:                tt.fields.KeyFile,
				TagsIgnore:             tt.fields.TagsIgnore,
				TagsRename:             tt.fields.TagsRename,
				IgnoreTagKV:            tt.fields.IgnoreTagKV,
				HTTPHeaders:            tt.fields.HTTPHeaders,
				Tags:                   tt.fields.Tags,
				DisableHostTag:         tt.fields.DisableHostTag,
				DisableInstanceTag:     tt.fields.DisableInstanceTag,
				DisableInfoTag:         tt.fields.DisableInfoTag,
				Auth:                   tt.fields.Auth,
				pm:                     tt.fields.pm,
				feeder:                 tt.fields.Feeder,
				Election:               tt.fields.Election,
				pauseCh:                tt.fields.pauseCh,
				pause:                  tt.fields.pause,
				tagger:                 tt.fields.Tagger,
				urls:                   tt.fields.urls,
				semStop:                tt.fields.semStop,
				mergedTags:             tt.fields.mergedTags,
				l:                      tt.fields.l,
			}

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, tt.mockData)
			}))
			t.Log(srv.URL)
			defer srv.Close()

			inp.URLs = []string{srv.URL}

			err := inp.setup()
			assert.NoError(t, err)

			pts, err := inp.getPts()
			if (err != nil) != tt.wantErr {
				t.Errorf("Input.collect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got := []string{}
			for _, pt := range pts {
				s := pt.LineProto()
				s = s[:strings.LastIndex(s, " ")]
				got = append(got, s)
			}

			sort.Strings(got)
			sort.Strings(tt.want)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Input.collect() = %v, want %v", got, tt.want)
			}
		})
	}
}

var mock5_11_0 string = `# HELP neo4j_database_neo4j_check_point_flushed_bytes Generated from Dropwizard metric import (metric=neo4j.database.neo4j.check_point.flushed_bytes, type=com.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$2064/0x00007f39d8a5bc08)
# TYPE neo4j_database_neo4j_check_point_flushed_bytes gauge
neo4j_database_neo4j_check_point_flushed_bytes 0.0
# HELP neo4j_database_system_pool_transaction_system_used_heap Generated from Dropwizard metric import (metric=neo4j.database.system.pool.transaction.system.used_heap, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$2100/0x00007f39d8a63428)
# TYPE neo4j_database_system_pool_transaction_system_used_heap gauge
neo4j_database_system_pool_transaction_system_used_heap 0.0
# HELP neo4j_database_neo4j_transaction_committed_read_total Generated from Dropwizard metric import (metric=neo4j.database.neo4j.transaction.committed_read, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_neo4j_transaction_committed_read_total counter
neo4j_database_neo4j_transaction_committed_read_total 0.0
# HELP neo4j_database_system_check_point_pages_flushed Generated from Dropwizard metric import (metric=neo4j.database.system.check_point.pages_flushed, type=com.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$2054/0x00007f39d8a5a678)
# TYPE neo4j_database_system_check_point_pages_flushed gauge
neo4j_database_system_check_point_pages_flushed 0.0
# HELP neo4j_database_system_db_query_execution_latency_millis Generated from Dropwizard metric import (metric=neo4j.database.system.db.query.execution.latency.millis, type=com.codahale.metrics.Histogram)
# TYPE neo4j_database_system_db_query_execution_latency_millis summary
neo4j_database_system_db_query_execution_latency_millis{quantile="0.5",} 0.0
neo4j_database_system_db_query_execution_latency_millis{quantile="0.75",} 0.0
neo4j_database_system_db_query_execution_latency_millis{quantile="0.95",} 0.0
neo4j_database_system_db_query_execution_latency_millis{quantile="0.98",} 0.0
neo4j_database_system_db_query_execution_latency_millis{quantile="0.99",} 0.0
neo4j_database_system_db_query_execution_latency_millis{quantile="0.999",} 0.0
neo4j_database_system_db_query_execution_latency_millis_count 0.0
# HELP neo4j_dbms_vm_memory_buffer_direct_used Generated from Dropwizard metric import (metric=neo4j.dbms.vm.memory.buffer.direct.used, type=com.neo4j.metrics.source.jvm.JVMMemoryBuffersMetrics$$Lambda$2452/0x00007f39d8b32460)
# TYPE neo4j_dbms_vm_memory_buffer_direct_used gauge
neo4j_dbms_vm_memory_buffer_direct_used 7864451.0
# HELP neo4j_database_neo4j_ids_in_use_relationship Generated from Dropwizard metric import (metric=neo4j.database.neo4j.ids_in_use.relationship, type=com.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$2074/0x00007f39d8a58c50)
# TYPE neo4j_database_neo4j_ids_in_use_relationship gauge
neo4j_database_neo4j_ids_in_use_relationship 0.0
# HELP neo4j_dbms_vm_file_descriptors_count Generated from Dropwizard metric import (metric=neo4j.dbms.vm.file.descriptors.count, type=com.neo4j.metrics.source.jvm.FileDescriptorMetrics$$Lambda$2454/0x00007f39d8b328a8)
# TYPE neo4j_dbms_vm_file_descriptors_count gauge
neo4j_dbms_vm_file_descriptors_count 868.0
# HELP neo4j_database_neo4j_transaction_active_write Generated from Dropwizard metric import (metric=neo4j.database.neo4j.transaction.active_write, type=com.neo4j.metrics.source.db.TransactionMetrics$$Lambda$2013/0x00007f39d8a56970)
# TYPE neo4j_database_neo4j_transaction_active_write gauge
neo4j_database_neo4j_transaction_active_write 0.0
# HELP neo4j_database_system_transaction_rollbacks_write_total Generated from Dropwizard metric import (metric=neo4j.database.system.transaction.rollbacks_write, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_system_transaction_rollbacks_write_total counter
neo4j_database_system_transaction_rollbacks_write_total 0.0
# HELP neo4j_database_system_transaction_peak_concurrent_total Generated from Dropwizard metric import (metric=neo4j.database.system.transaction.peak_concurrent, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_system_transaction_peak_concurrent_total counter
neo4j_database_system_transaction_peak_concurrent_total 1.0
# HELP neo4j_database_system_check_point_io_limit Generated from Dropwizard metric import (metric=neo4j.database.system.check_point.io_limit, type=com.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$2058/0x00007f39d8a5af18)
# TYPE neo4j_database_system_check_point_io_limit gauge
neo4j_database_system_check_point_io_limit 0.0
# HELP neo4j_database_neo4j_transaction_rollbacks_read_total Generated from Dropwizard metric import (metric=neo4j.database.neo4j.transaction.rollbacks_read, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_neo4j_transaction_rollbacks_read_total counter
neo4j_database_neo4j_transaction_rollbacks_read_total 0.0
# HELP neo4j_database_neo4j_check_point_limit_times Generated from Dropwizard metric import (metric=neo4j.database.neo4j.check_point.limit_times, type=com.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$2060/0x00007f39d8a5b368)
# TYPE neo4j_database_neo4j_check_point_limit_times gauge
neo4j_database_neo4j_check_point_limit_times 0.0
# HELP neo4j_database_neo4j_db_query_execution_success_total Generated from Dropwizard metric import (metric=neo4j.database.neo4j.db.query.execution.success, type=com.codahale.metrics.Meter)
# TYPE neo4j_database_neo4j_db_query_execution_success_total counter
neo4j_database_neo4j_db_query_execution_success_total 0.0
# HELP neo4j_database_neo4j_db_query_execution_latency_millis Generated from Dropwizard metric import (metric=neo4j.database.neo4j.db.query.execution.latency.millis, type=com.codahale.metrics.Histogram)
# TYPE neo4j_database_neo4j_db_query_execution_latency_millis summary
neo4j_database_neo4j_db_query_execution_latency_millis{quantile="0.5",} 0.0
neo4j_database_neo4j_db_query_execution_latency_millis{quantile="0.75",} 0.0
neo4j_database_neo4j_db_query_execution_latency_millis{quantile="0.95",} 0.0
neo4j_database_neo4j_db_query_execution_latency_millis{quantile="0.98",} 0.0
neo4j_database_neo4j_db_query_execution_latency_millis{quantile="0.99",} 0.0
neo4j_database_neo4j_db_query_execution_latency_millis{quantile="0.999",} 0.0
neo4j_database_neo4j_db_query_execution_latency_millis_count 0.0
# HELP neo4j_dbms_bolt_messages_started_total Generated from Dropwizard metric import (metric=neo4j.dbms.bolt.messages_started, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_dbms_bolt_messages_started_total counter
neo4j_dbms_bolt_messages_started_total 0.0
# HELP neo4j_database_neo4j_ids_in_use_property Generated from Dropwizard metric import (metric=neo4j.database.neo4j.ids_in_use.property, type=com.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$2076/0x00007f39d8a60228)
# TYPE neo4j_database_neo4j_ids_in_use_property gauge
neo4j_database_neo4j_ids_in_use_property 8.0
# HELP neo4j_database_neo4j_transaction_active_read Generated from Dropwizard metric import (metric=neo4j.database.neo4j.transaction.active_read, type=com.neo4j.metrics.source.db.TransactionMetrics$$Lambda$2011/0x00007f39d8a56520)
# TYPE neo4j_database_neo4j_transaction_active_read gauge
neo4j_database_neo4j_transaction_active_read 0.0
# HELP neo4j_dbms_bolt_connections_closed_total Generated from Dropwizard metric import (metric=neo4j.dbms.bolt.connections_closed, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_dbms_bolt_connections_closed_total counter
neo4j_dbms_bolt_connections_closed_total 0.0
# HELP neo4j_dbms_bolt_connections_idle Generated from Dropwizard metric import (metric=neo4j.dbms.bolt.connections_idle, type=com.neo4j.metrics.source.db.BoltMetrics$$Lambda$2473/0x00007f39d8b35190)
# TYPE neo4j_dbms_bolt_connections_idle gauge
neo4j_dbms_bolt_connections_idle 0.0
# HELP neo4j_database_neo4j_transaction_committed_total Generated from Dropwizard metric import (metric=neo4j.database.neo4j.transaction.committed, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_neo4j_transaction_committed_total counter
neo4j_database_neo4j_transaction_committed_total 0.0
# HELP neo4j_database_system_pool_transaction_system_total_used Generated from Dropwizard metric import (metric=neo4j.database.system.pool.transaction.system.total_used, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$2102/0x00007f39d8a63878)
# TYPE neo4j_database_system_pool_transaction_system_total_used gauge
neo4j_database_system_pool_transaction_system_total_used 0.0
# HELP neo4j_database_neo4j_store_size_total Generated from Dropwizard metric import (metric=neo4j.database.neo4j.store.size.total, type=com.neo4j.metrics.source.db.StoreSizeMetrics$$Lambda$2083/0x00007f39d8a61140)
# TYPE neo4j_database_neo4j_store_size_total gauge
neo4j_database_neo4j_store_size_total 2.7036229E8
# HELP neo4j_database_system_check_point_total_time_total Generated from Dropwizard metric import (metric=neo4j.database.system.check_point.total_time, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_system_check_point_total_time_total counter
neo4j_database_system_check_point_total_time_total 0.0
# HELP neo4j_dbms_vm_gc_time_g1_young_generation_total Generated from Dropwizard metric import (metric=neo4j.dbms.vm.gc.time.g1_young_generation, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_dbms_vm_gc_time_g1_young_generation_total counter
neo4j_dbms_vm_gc_time_g1_young_generation_total 71.0
# HELP neo4j_database_neo4j_ids_in_use_relationship_type Generated from Dropwizard metric import (metric=neo4j.database.neo4j.ids_in_use.relationship_type, type=com.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$2078/0x00007f39d8a60678)
# TYPE neo4j_database_neo4j_ids_in_use_relationship_type gauge
neo4j_database_neo4j_ids_in_use_relationship_type 0.0
# HELP neo4j_database_neo4j_db_query_execution_failure_total Generated from Dropwizard metric import (metric=neo4j.database.neo4j.db.query.execution.failure, type=com.codahale.metrics.Meter)
# TYPE neo4j_database_neo4j_db_query_execution_failure_total counter
neo4j_database_neo4j_db_query_execution_failure_total 0.0
# HELP neo4j_database_neo4j_cypher_replan_events_total Generated from Dropwizard metric import (metric=neo4j.database.neo4j.cypher.replan_events, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_neo4j_cypher_replan_events_total counter
neo4j_database_neo4j_cypher_replan_events_total 0.0
# HELP neo4j_database_system_store_size_database Generated from Dropwizard metric import (metric=neo4j.database.system.store.size.database, type=com.neo4j.metrics.source.db.StoreSizeMetrics$$Lambda$2086/0x00007f39d8a617b8)
# TYPE neo4j_database_system_store_size_database gauge
neo4j_database_system_store_size_database 1344370.0
# HELP neo4j_dbms_vm_pause_time_total Generated from Dropwizard metric import (metric=neo4j.dbms.vm.pause_time, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_dbms_vm_pause_time_total counter
neo4j_dbms_vm_pause_time_total 0.0
# HELP neo4j_database_neo4j_check_point_duration Generated from Dropwizard metric import (metric=neo4j.database.neo4j.check_point.duration, type=com.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$2052/0x00007f39d8a5a228)
# TYPE neo4j_database_neo4j_check_point_duration gauge
neo4j_database_neo4j_check_point_duration 0.0
# HELP neo4j_dbms_pool_bolt_total_size Generated from Dropwizard metric import (metric=neo4j.dbms.pool.bolt.total_size, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$2496/0x00007f39d8b386a8)
# TYPE neo4j_dbms_pool_bolt_total_size gauge
neo4j_dbms_pool_bolt_total_size 9.223372036854776E18
# HELP neo4j_database_neo4j_check_point_events_total Generated from Dropwizard metric import (metric=neo4j.database.neo4j.check_point.events, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_neo4j_check_point_events_total counter
neo4j_database_neo4j_check_point_events_total 0.0
# HELP neo4j_database_neo4j_check_point_io_limit Generated from Dropwizard metric import (metric=neo4j.database.neo4j.check_point.io_limit, type=com.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$2058/0x00007f39d8a5af18)
# TYPE neo4j_database_neo4j_check_point_io_limit gauge
neo4j_database_neo4j_check_point_io_limit 0.0
# HELP neo4j_database_system_ids_in_use_relationship_type Generated from Dropwizard metric import (metric=neo4j.database.system.ids_in_use.relationship_type, type=com.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$2078/0x00007f39d8a60678)
# TYPE neo4j_database_system_ids_in_use_relationship_type gauge
neo4j_database_system_ids_in_use_relationship_type 9.0
# HELP neo4j_database_system_transaction_last_committed_tx_id_total Generated from Dropwizard metric import (metric=neo4j.database.system.transaction.last_committed_tx_id, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_system_transaction_last_committed_tx_id_total counter
neo4j_database_system_transaction_last_committed_tx_id_total 95.0
# HELP neo4j_database_neo4j_transaction_peak_concurrent_total Generated from Dropwizard metric import (metric=neo4j.database.neo4j.transaction.peak_concurrent, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_neo4j_transaction_peak_concurrent_total counter
neo4j_database_neo4j_transaction_peak_concurrent_total 0.0
# HELP neo4j_database_system_ids_in_use_relationship Generated from Dropwizard metric import (metric=neo4j.database.system.ids_in_use.relationship, type=com.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$2074/0x00007f39d8a58c50)
# TYPE neo4j_database_system_ids_in_use_relationship gauge
neo4j_database_system_ids_in_use_relationship 111.0
# HELP neo4j_database_system_cypher_replan_events_total Generated from Dropwizard metric import (metric=neo4j.database.system.cypher.replan_events, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_system_cypher_replan_events_total counter
neo4j_database_system_cypher_replan_events_total 0.0
# HELP neo4j_database_neo4j_transaction_rollbacks_write_total Generated from Dropwizard metric import (metric=neo4j.database.neo4j.transaction.rollbacks_write, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_neo4j_transaction_rollbacks_write_total counter
neo4j_database_neo4j_transaction_rollbacks_write_total 0.0
# HELP neo4j_database_system_ids_in_use_property Generated from Dropwizard metric import (metric=neo4j.database.system.ids_in_use.property, type=com.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$2076/0x00007f39d8a60228)
# TYPE neo4j_database_system_ids_in_use_property gauge
neo4j_database_system_ids_in_use_property 135.0
# HELP neo4j_dbms_vm_threads Generated from Dropwizard metric import (metric=neo4j.dbms.vm.threads, type=com.neo4j.metrics.source.jvm.ThreadMetrics$$Lambda$2446/0x00007f39d8b30ad8)
# TYPE neo4j_dbms_vm_threads gauge
neo4j_dbms_vm_threads 126.0
# HELP neo4j_dbms_bolt_connections_running Generated from Dropwizard metric import (metric=neo4j.dbms.bolt.connections_running, type=com.neo4j.metrics.source.db.BoltMetrics$$Lambda$2471/0x00007f39d8b34d40)
# TYPE neo4j_dbms_bolt_connections_running gauge
neo4j_dbms_bolt_connections_running 0.0
# HELP neo4j_dbms_vm_memory_pool_g1_eden_space Generated from Dropwizard metric import (metric=neo4j.dbms.vm.memory.pool.g1_eden_space, type=com.neo4j.metrics.source.jvm.JVMMemoryPoolMetrics$$Lambda$2448/0x00007f39d8b319b8)
# TYPE neo4j_dbms_vm_memory_pool_g1_eden_space gauge
neo4j_dbms_vm_memory_pool_g1_eden_space 0.0
# HELP neo4j_database_system_pool_transaction_system_used_native Generated from Dropwizard metric import (metric=neo4j.database.system.pool.transaction.system.used_native, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$2101/0x00007f39d8a63650)
# TYPE neo4j_database_system_pool_transaction_system_used_native gauge
neo4j_database_system_pool_transaction_system_used_native 0.0
# HELP neo4j_database_system_transaction_committed_read_total Generated from Dropwizard metric import (metric=neo4j.database.system.transaction.committed_read, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_system_transaction_committed_read_total counter
neo4j_database_system_transaction_committed_read_total 3.0
# HELP neo4j_database_system_transaction_rollbacks_total Generated from Dropwizard metric import (metric=neo4j.database.system.transaction.rollbacks, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_system_transaction_rollbacks_total counter
neo4j_database_system_transaction_rollbacks_total 3.0
# HELP neo4j_database_system_transaction_rollbacks_read_total Generated from Dropwizard metric import (metric=neo4j.database.system.transaction.rollbacks_read, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_system_transaction_rollbacks_read_total counter
neo4j_database_system_transaction_rollbacks_read_total 3.0
# HELP neo4j_database_neo4j_check_point_total_time_total Generated from Dropwizard metric import (metric=neo4j.database.neo4j.check_point.total_time, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_neo4j_check_point_total_time_total counter
neo4j_database_neo4j_check_point_total_time_total 0.0
# HELP neo4j_database_neo4j_check_point_io_performed Generated from Dropwizard metric import (metric=neo4j.database.neo4j.check_point.io_performed, type=com.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$2056/0x00007f39d8a5aac8)
# TYPE neo4j_database_neo4j_check_point_io_performed gauge
neo4j_database_neo4j_check_point_io_performed 0.0
# HELP neo4j_database_neo4j_transaction_rollbacks_total Generated from Dropwizard metric import (metric=neo4j.database.neo4j.transaction.rollbacks, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_neo4j_transaction_rollbacks_total counter
neo4j_database_neo4j_transaction_rollbacks_total 0.0
# HELP neo4j_database_neo4j_store_size_database Generated from Dropwizard metric import (metric=neo4j.database.neo4j.store.size.database, type=com.neo4j.metrics.source.db.StoreSizeMetrics$$Lambda$2086/0x00007f39d8a617b8)
# TYPE neo4j_database_neo4j_store_size_database gauge
neo4j_database_neo4j_store_size_database 878258.0
# HELP neo4j_database_system_check_point_duration Generated from Dropwizard metric import (metric=neo4j.database.system.check_point.duration, type=com.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$2052/0x00007f39d8a5a228)
# TYPE neo4j_database_system_check_point_duration gauge
neo4j_database_system_check_point_duration 0.0
# HELP neo4j_dbms_vm_gc_time_g1_old_generation_total Generated from Dropwizard metric import (metric=neo4j.dbms.vm.gc.time.g1_old_generation, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_dbms_vm_gc_time_g1_old_generation_total counter
neo4j_dbms_vm_gc_time_g1_old_generation_total 0.0
# HELP neo4j_database_neo4j_pool_transaction_neo4j_used_heap Generated from Dropwizard metric import (metric=neo4j.database.neo4j.pool.transaction.neo4j.used_heap, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$2100/0x00007f39d8a63428)
# TYPE neo4j_database_neo4j_pool_transaction_neo4j_used_heap gauge
neo4j_database_neo4j_pool_transaction_neo4j_used_heap 0.0
# HELP neo4j_database_neo4j_pool_transaction_neo4j_used_native Generated from Dropwizard metric import (metric=neo4j.database.neo4j.pool.transaction.neo4j.used_native, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$2101/0x00007f39d8a63650)
# TYPE neo4j_database_neo4j_pool_transaction_neo4j_used_native gauge
neo4j_database_neo4j_pool_transaction_neo4j_used_native 0.0
# HELP neo4j_database_system_db_query_execution_success_total Generated from Dropwizard metric import (metric=neo4j.database.system.db.query.execution.success, type=com.codahale.metrics.Meter)
# TYPE neo4j_database_system_db_query_execution_success_total counter
neo4j_database_system_db_query_execution_success_total 0.0
# HELP neo4j_dbms_page_cache_usage_ratio Generated from Dropwizard metric import (metric=neo4j.dbms.page_cache.usage_ratio, type=com.neo4j.metrics.source.db.PageCacheMetrics$$Lambda$2430/0x00007f39d8b2ba48)
# TYPE neo4j_dbms_page_cache_usage_ratio gauge
neo4j_dbms_page_cache_usage_ratio 0.004258578431372549
# HELP neo4j_database_system_check_point_events_total Generated from Dropwizard metric import (metric=neo4j.database.system.check_point.events, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_system_check_point_events_total counter
neo4j_database_system_check_point_events_total 0.0
# HELP neo4j_database_system_store_size_total Generated from Dropwizard metric import (metric=neo4j.database.system.store.size.total, type=com.neo4j.metrics.source.db.StoreSizeMetrics$$Lambda$2083/0x00007f39d8a61140)
# TYPE neo4j_database_system_store_size_total gauge
neo4j_database_system_store_size_total 2.70828402E8
# HELP neo4j_database_system_check_point_limit_millis Generated from Dropwizard metric import (metric=neo4j.database.system.check_point.limit_millis, type=com.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$2062/0x00007f39d8a5b7b8)
# TYPE neo4j_database_system_check_point_limit_millis gauge
neo4j_database_system_check_point_limit_millis 0.0
# HELP neo4j_database_neo4j_ids_in_use_node Generated from Dropwizard metric import (metric=neo4j.database.neo4j.ids_in_use.node, type=com.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$2072/0x00007f39d8a58800)
# TYPE neo4j_database_neo4j_ids_in_use_node gauge
neo4j_database_neo4j_ids_in_use_node 0.0
# HELP neo4j_database_system_transaction_active_write Generated from Dropwizard metric import (metric=neo4j.database.system.transaction.active_write, type=com.neo4j.metrics.source.db.TransactionMetrics$$Lambda$2013/0x00007f39d8a56970)
# TYPE neo4j_database_system_transaction_active_write gauge
neo4j_database_system_transaction_active_write 0.0
# HELP neo4j_dbms_page_cache_hit_ratio Generated from Dropwizard metric import (metric=neo4j.dbms.page_cache.hit_ratio, type=com.neo4j.metrics.source.db.PageCacheHitRatioGauge)
# TYPE neo4j_dbms_page_cache_hit_ratio gauge
neo4j_dbms_page_cache_hit_ratio 0.0
# HELP neo4j_database_neo4j_transaction_committed_write_total Generated from Dropwizard metric import (metric=neo4j.database.neo4j.transaction.committed_write, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_neo4j_transaction_committed_write_total counter
neo4j_database_neo4j_transaction_committed_write_total 0.0
# HELP neo4j_database_system_db_query_execution_failure_total Generated from Dropwizard metric import (metric=neo4j.database.system.db.query.execution.failure, type=com.codahale.metrics.Meter)
# TYPE neo4j_database_system_db_query_execution_failure_total counter
neo4j_database_system_db_query_execution_failure_total 0.0
# HELP neo4j_database_system_transaction_committed_write_total Generated from Dropwizard metric import (metric=neo4j.database.system.transaction.committed_write, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_system_transaction_committed_write_total counter
neo4j_database_system_transaction_committed_write_total 0.0
# HELP neo4j_database_neo4j_check_point_limit_millis Generated from Dropwizard metric import (metric=neo4j.database.neo4j.check_point.limit_millis, type=com.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$2062/0x00007f39d8a5b7b8)
# TYPE neo4j_database_neo4j_check_point_limit_millis gauge
neo4j_database_neo4j_check_point_limit_millis 0.0
# HELP neo4j_dbms_page_cache_page_faults_total Generated from Dropwizard metric import (metric=neo4j.dbms.page_cache.page_faults, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_dbms_page_cache_page_faults_total counter
neo4j_dbms_page_cache_page_faults_total 314.0
# HELP neo4j_database_system_check_point_flushed_bytes Generated from Dropwizard metric import (metric=neo4j.database.system.check_point.flushed_bytes, type=com.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$2064/0x00007f39d8a5bc08)
# TYPE neo4j_database_system_check_point_flushed_bytes gauge
neo4j_database_system_check_point_flushed_bytes 0.0
# HELP neo4j_database_neo4j_transaction_last_committed_tx_id_total Generated from Dropwizard metric import (metric=neo4j.database.neo4j.transaction.last_committed_tx_id, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_neo4j_transaction_last_committed_tx_id_total counter
neo4j_database_neo4j_transaction_last_committed_tx_id_total 3.0
# HELP neo4j_dbms_pool_bolt_used_heap Generated from Dropwizard metric import (metric=neo4j.dbms.pool.bolt.used_heap, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$2100/0x00007f39d8a63428)
# TYPE neo4j_dbms_pool_bolt_used_heap gauge
neo4j_dbms_pool_bolt_used_heap 0.0
# HELP neo4j_dbms_pool_bolt_free Generated from Dropwizard metric import (metric=neo4j.dbms.pool.bolt.free, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$2498/0x00007f39d8b38b18)
# TYPE neo4j_dbms_pool_bolt_free gauge
neo4j_dbms_pool_bolt_free 9.223372036854776E18
# HELP neo4j_dbms_pool_bolt_total_used Generated from Dropwizard metric import (metric=neo4j.dbms.pool.bolt.total_used, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$2102/0x00007f39d8a63878)
# TYPE neo4j_dbms_pool_bolt_total_used gauge
neo4j_dbms_pool_bolt_total_used 0.0
# HELP neo4j_dbms_vm_heap_used Generated from Dropwizard metric import (metric=neo4j.dbms.vm.heap.used, type=com.neo4j.metrics.source.jvm.HeapMetrics$$Lambda$2441/0x00007f39d8b2c400)
# TYPE neo4j_dbms_vm_heap_used gauge
neo4j_dbms_vm_heap_used 7.5342848E7
# HELP neo4j_database_neo4j_pool_transaction_neo4j_total_used Generated from Dropwizard metric import (metric=neo4j.database.neo4j.pool.transaction.neo4j.total_used, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$2102/0x00007f39d8a63878)
# TYPE neo4j_database_neo4j_pool_transaction_neo4j_total_used gauge
neo4j_database_neo4j_pool_transaction_neo4j_total_used 0.0
# HELP neo4j_database_system_ids_in_use_node Generated from Dropwizard metric import (metric=neo4j.database.system.ids_in_use.node, type=com.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$2072/0x00007f39d8a58800)
# TYPE neo4j_database_system_ids_in_use_node gauge
neo4j_database_system_ids_in_use_node 55.0
# HELP neo4j_dbms_vm_memory_pool_g1_old_gen Generated from Dropwizard metric import (metric=neo4j.dbms.vm.memory.pool.g1_old_gen, type=com.neo4j.metrics.source.jvm.JVMMemoryPoolMetrics$$Lambda$2448/0x00007f39d8b319b8)
# TYPE neo4j_dbms_vm_memory_pool_g1_old_gen gauge
neo4j_dbms_vm_memory_pool_g1_old_gen 6.695424E7
# HELP neo4j_database_neo4j_check_point_pages_flushed Generated from Dropwizard metric import (metric=neo4j.database.neo4j.check_point.pages_flushed, type=com.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$2054/0x00007f39d8a5a678)
# TYPE neo4j_database_neo4j_check_point_pages_flushed gauge
neo4j_database_neo4j_check_point_pages_flushed 0.0
# HELP neo4j_dbms_bolt_connections_opened_total Generated from Dropwizard metric import (metric=neo4j.dbms.bolt.connections_opened, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_dbms_bolt_connections_opened_total counter
neo4j_dbms_bolt_connections_opened_total 0.0
# HELP neo4j_database_system_transaction_committed_total Generated from Dropwizard metric import (metric=neo4j.database.system.transaction.committed, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_database_system_transaction_committed_total counter
neo4j_database_system_transaction_committed_total 3.0
# HELP neo4j_dbms_page_cache_hits_total Generated from Dropwizard metric import (metric=neo4j.dbms.page_cache.hits, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_dbms_page_cache_hits_total counter
neo4j_dbms_page_cache_hits_total 883.0
# HELP neo4j_database_system_check_point_io_performed Generated from Dropwizard metric import (metric=neo4j.database.system.check_point.io_performed, type=com.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$2056/0x00007f39d8a5aac8)
# TYPE neo4j_database_system_check_point_io_performed gauge
neo4j_database_system_check_point_io_performed 0.0
# HELP neo4j_database_system_transaction_active_read Generated from Dropwizard metric import (metric=neo4j.database.system.transaction.active_read, type=com.neo4j.metrics.source.db.TransactionMetrics$$Lambda$2011/0x00007f39d8a56520)
# TYPE neo4j_database_system_transaction_active_read gauge
neo4j_database_system_transaction_active_read 0.0
# HELP neo4j_dbms_bolt_messages_received_total Generated from Dropwizard metric import (metric=neo4j.dbms.bolt.messages_received, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_dbms_bolt_messages_received_total counter
neo4j_dbms_bolt_messages_received_total 0.0
# HELP neo4j_database_system_check_point_limit_times Generated from Dropwizard metric import (metric=neo4j.database.system.check_point.limit_times, type=com.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$2060/0x00007f39d8a5b368)
# TYPE neo4j_database_system_check_point_limit_times gauge
neo4j_database_system_check_point_limit_times 0.0
`

var want5_11_0 []string = []string{
	"neo4j,db=system database_store_size_total=270828402",
	"neo4j,db=system database_ids_in_use_relationship=111",
	"neo4j,db=system database_ids_in_use_property=135",
	"neo4j,db=system database_db_query_execution_success_total=0",
	"neo4j dbms_vm_threads=126",
	"neo4j,db=neo4j database_ids_in_use_node=0",
	"neo4j,db=system database_db_query_execution_failure_total=0",
	"neo4j,db=system database_check_point_limit_times=0",
	"neo4j,db=neo4j database_db_query_execution_success_total=0",
	"neo4j dbms_bolt_connections_closed_total=0",
	"neo4j dbms_bolt_connections_idle=0",
	"neo4j,db=system database_ids_in_use_relationship_type=9",
	"neo4j dbms_bolt_connections_running=0",
	"neo4j,db=system database_transaction_rollbacks_read_total=3",
	"neo4j,db=neo4j database_store_size_database=878258",
	"neo4j dbms_bolt_messages_received_total=0",
	"neo4j,db=neo4j database_check_point_flushed_bytes=0",
	"neo4j,db=neo4j database_transaction_rollbacks_read_total=0",
	"neo4j,db=neo4j database_check_point_io_limit=0",
	"neo4j,db=neo4j database_transaction_active_write=0",
	"neo4j,db=system database_check_point_limit_millis=0",
	"neo4j,db=neo4j database_check_point_pages_flushed=0",
	"neo4j,database=system,db=system,pool=transaction database_pool_used_heap=0",
	"neo4j,db=neo4j database_store_size_total=270362290",
	"neo4j,db=system database_check_point_total_time_total=0",
	"neo4j,db=neo4j database_transaction_rollbacks_total=0",
	"neo4j,database=neo4j,db=neo4j,pool=transaction database_pool_used_native=0",
	"neo4j,db=system database_check_point_flushed_bytes=0",
	"neo4j dbms_vm_file_descriptors_count=868",
	"neo4j,db=system database_transaction_peak_concurrent_total=1",
	"neo4j,db=neo4j database_cypher_replan_events_total=0",
	"neo4j dbms_page_cache_hits_total=883",
	"neo4j,db=neo4j database_db_query_execution_latency_millis_count=0,database_db_query_execution_latency_millis_sum=0",
	"neo4j,db=neo4j,quantile=0.5 database_db_query_execution_latency_millis=0",
	"neo4j,db=neo4j,quantile=0.75 database_db_query_execution_latency_millis=0",
	"neo4j,db=neo4j,quantile=0.95 database_db_query_execution_latency_millis=0",
	"neo4j,db=neo4j,quantile=0.98 database_db_query_execution_latency_millis=0",
	"neo4j,db=neo4j,quantile=0.99 database_db_query_execution_latency_millis=0",
	"neo4j,db=neo4j,quantile=0.999 database_db_query_execution_latency_millis=0",
	"neo4j,db=system database_transaction_committed_read_total=3",
	"neo4j,db=system database_transaction_active_write=0",
	"neo4j,pool=bolt dbms_pool_total_size=9223372036854776000",
	"neo4j dbms_page_cache_hit_ratio=0",
	"neo4j,db=system database_transaction_active_read=0",
	"neo4j,gc=g1_old_generation dbms_vm_gc_time_total=0",
	"neo4j,db=system database_check_point_events_total=0",
	"neo4j,db=neo4j database_transaction_committed_read_total=0",
	"neo4j,db=system database_cypher_replan_events_total=0",
	"neo4j,db=neo4j database_check_point_io_performed=0",
	"neo4j,db=neo4j database_transaction_last_committed_tx_id_total=3",
	"neo4j dbms_bolt_messages_started_total=0",
	"neo4j,pool=g1_eden_space dbms_vm_memory_pool=0",
	"neo4j dbms_page_cache_page_faults_total=314",
	"neo4j,db=neo4j database_transaction_rollbacks_write_total=0",
	"neo4j,db=system database_check_point_io_limit=0",
	"neo4j,db=neo4j database_db_query_execution_failure_total=0",
	"neo4j,db=neo4j database_check_point_duration=0",
	"neo4j,db=neo4j database_check_point_events_total=0",
	"neo4j dbms_bolt_connections_opened_total=0",
	"neo4j,db=system database_db_query_execution_latency_millis_count=0,database_db_query_execution_latency_millis_sum=0",
	"neo4j,db=system,quantile=0.5 database_db_query_execution_latency_millis=0",
	"neo4j,db=system,quantile=0.75 database_db_query_execution_latency_millis=0",
	"neo4j,db=system,quantile=0.95 database_db_query_execution_latency_millis=0",
	"neo4j,db=system,quantile=0.98 database_db_query_execution_latency_millis=0",
	"neo4j,db=system,quantile=0.99 database_db_query_execution_latency_millis=0",
	"neo4j,db=system,quantile=0.999 database_db_query_execution_latency_millis=0",
	"neo4j,db=neo4j database_ids_in_use_property=8",
	"neo4j,db=neo4j database_transaction_active_read=0",
	"neo4j,db=system database_transaction_rollbacks_total=3",
	"neo4j,db=system database_check_point_duration=0",
	"neo4j,db=system database_transaction_committed_write_total=0",
	"neo4j,db=neo4j database_check_point_limit_millis=0",
	"neo4j,pool=bolt dbms_pool_used_heap=0",
	"neo4j,db=neo4j database_check_point_limit_times=0",
	"neo4j,db=neo4j database_ids_in_use_relationship_type=0",
	"neo4j,db=neo4j database_transaction_peak_concurrent_total=0",
	"neo4j dbms_vm_heap_used=75342848",
	"neo4j,db=system database_transaction_last_committed_tx_id_total=95",
	"neo4j,database=neo4j,db=neo4j,pool=transaction database_pool_used_heap=0",
	"neo4j dbms_page_cache_usage_ratio=0.004258578431372549",
	"neo4j,db=neo4j database_transaction_committed_write_total=0",
	"neo4j,database=neo4j,db=neo4j,pool=transaction database_pool_total_used=0",
	"neo4j,db=system database_check_point_pages_flushed=0",
	"neo4j,bufferpool=direct dbms_vm_memory_buffer_used=7864451",
	"neo4j,database=system,db=system,pool=transaction database_pool_total_used=0",
	"neo4j,db=system database_ids_in_use_node=55",
	"neo4j,pool=g1_old_gen dbms_vm_memory_pool=66954240",
	"neo4j,db=system database_check_point_io_performed=0",
	"neo4j,db=neo4j database_check_point_total_time_total=0",
	"neo4j,pool=bolt dbms_pool_free=9223372036854776000",
	"neo4j,pool=bolt dbms_pool_total_used=0",
	"neo4j,db=neo4j database_ids_in_use_relationship=0",
	"neo4j,db=system database_store_size_database=1344370",
	"neo4j,database=system,db=system,pool=transaction database_pool_used_native=0",
	"neo4j dbms_vm_pause_time_total=0",
	"neo4j,db=system database_transaction_committed_total=3",
	"neo4j,db=system database_transaction_rollbacks_write_total=0",
	"neo4j,db=neo4j database_transaction_committed_total=0",
	"neo4j,gc=g1_young_generation dbms_vm_gc_time_total=71",
}

var mock4_4_0 string = `# HELP neo4j_system_cypher_replan_events_total Generated from Dropwizard metric import (metric=neo4j.system.cypher.replan_events, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_system_cypher_replan_events_total counter
neo4j_system_cypher_replan_events_total 0.0
# HELP neo4j_system_db_query_execution_failure_total Generated from Dropwizard metric import (metric=neo4j.system.db.query.execution.failure, type=com.codahale.metrics.Meter)
# TYPE neo4j_system_db_query_execution_failure_total counter
neo4j_system_db_query_execution_failure_total 0.0
# HELP neo4j_system_transaction_active_write Generated from Dropwizard metric import (metric=neo4j.system.transaction.active_write, type=com.neo4j.metrics.source.db.TransactionMetrics$$Lambda$1110/0x00000008408c0040)
# TYPE neo4j_system_transaction_active_write gauge
neo4j_system_transaction_active_write 0.0
# HELP neo4j_vm_pause_time Generated from Dropwizard metric import (metric=neo4j.vm.pause_time, type=com.neo4j.metrics.source.jvm.PauseMetrics$$Lambda$1433/0x000000084099dc40)
# TYPE neo4j_vm_pause_time gauge
neo4j_vm_pause_time 0.0
# HELP neo4j_system_db_query_execution_success_total Generated from Dropwizard metric import (metric=neo4j.system.db.query.execution.success, type=com.codahale.metrics.Meter)
# TYPE neo4j_system_db_query_execution_success_total counter
neo4j_system_db_query_execution_success_total 0.0
# HELP neo4j_neo4j_ids_in_use_property Generated from Dropwizard metric import (metric=neo4j.neo4j.ids_in_use.property, type=com.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$1147/0x00000008408c9440)
# TYPE neo4j_neo4j_ids_in_use_property gauge
neo4j_neo4j_ids_in_use_property 8.0
# HELP neo4j_system_db_query_execution_latency_millis Generated from Dropwizard metric import (metric=neo4j.system.db.query.execution.latency.millis, type=com.codahale.metrics.Histogram)
# TYPE neo4j_system_db_query_execution_latency_millis summary
neo4j_system_db_query_execution_latency_millis{quantile="0.5",} 0.0
neo4j_system_db_query_execution_latency_millis{quantile="0.75",} 0.0
neo4j_system_db_query_execution_latency_millis{quantile="0.95",} 0.0
neo4j_system_db_query_execution_latency_millis{quantile="0.98",} 0.0
neo4j_system_db_query_execution_latency_millis{quantile="0.99",} 0.0
neo4j_system_db_query_execution_latency_millis{quantile="0.999",} 0.0
neo4j_system_db_query_execution_latency_millis_count 0.0
# HELP neo4j_system_transaction_rollbacks_total Generated from Dropwizard metric import (metric=neo4j.system.transaction.rollbacks, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_system_transaction_rollbacks_total counter
neo4j_system_transaction_rollbacks_total 46.0
# HELP neo4j_system_pool_transaction_system_used_heap Generated from Dropwizard metric import (metric=neo4j.system.pool.transaction.system.used_heap, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$1170/0x00000008408cf040)
# TYPE neo4j_system_pool_transaction_system_used_heap gauge
neo4j_system_pool_transaction_system_used_heap 0.0
# HELP neo4j_bolt_connections_running Generated from Dropwizard metric import (metric=neo4j.bolt.connections_running, type=com.neo4j.metrics.source.db.BoltMetrics$$Lambda$1440/0x000000084099f840)
# TYPE neo4j_bolt_connections_running gauge
neo4j_bolt_connections_running 0.0
# HELP neo4j_vm_file_descriptors_count Generated from Dropwizard metric import (metric=neo4j.vm.file.descriptors.count, type=com.neo4j.metrics.source.jvm.FileDescriptorMetrics$$Lambda$1430/0x000000084099d040)
# TYPE neo4j_vm_file_descriptors_count gauge
neo4j_vm_file_descriptors_count 526.0
# HELP neo4j_vm_heap_used Generated from Dropwizard metric import (metric=neo4j.vm.heap.used, type=com.neo4j.metrics.source.jvm.HeapMetrics$$Lambda$1417/0x0000000840999c40)
# TYPE neo4j_vm_heap_used gauge
neo4j_vm_heap_used 1.13157592E8
# HELP neo4j_neo4j_check_point_total_time_total Generated from Dropwizard metric import (metric=neo4j.neo4j.check_point.total_time, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_neo4j_check_point_total_time_total counter
neo4j_neo4j_check_point_total_time_total 0.0
# HELP neo4j_system_check_point_duration Generated from Dropwizard metric import (metric=neo4j.system.check_point.duration, type=com.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$1135/0x00000008408c6440)
# TYPE neo4j_system_check_point_duration gauge
neo4j_system_check_point_duration 0.0
# HELP neo4j_neo4j_transaction_rollbacks_total Generated from Dropwizard metric import (metric=neo4j.neo4j.transaction.rollbacks, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_neo4j_transaction_rollbacks_total counter
neo4j_neo4j_transaction_rollbacks_total 0.0
# HELP neo4j_system_transaction_committed_read_total Generated from Dropwizard metric import (metric=neo4j.system.transaction.committed_read, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_system_transaction_committed_read_total counter
neo4j_system_transaction_committed_read_total 9.0
# HELP neo4j_system_pool_transaction_system_used_native Generated from Dropwizard metric import (metric=neo4j.system.pool.transaction.system.used_native, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$1171/0x00000008408cf440)
# TYPE neo4j_system_pool_transaction_system_used_native gauge
neo4j_system_pool_transaction_system_used_native 0.0
# HELP neo4j_system_transaction_active_read Generated from Dropwizard metric import (metric=neo4j.system.transaction.active_read, type=com.neo4j.metrics.source.db.TransactionMetrics$$Lambda$1108/0x00000008408a5840)
# TYPE neo4j_system_transaction_active_read gauge
neo4j_system_transaction_active_read 0.0
# HELP neo4j_vm_memory_buffer_direct_used Generated from Dropwizard metric import (metric=neo4j.vm.memory.buffer.direct.used, type=com.neo4j.metrics.source.jvm.JVMMemoryBuffersMetrics$$Lambda$1428/0x000000084099c840)
# TYPE neo4j_vm_memory_buffer_direct_used gauge
neo4j_vm_memory_buffer_direct_used 1728516.0
# HELP neo4j_system_store_size_database Generated from Dropwizard metric import (metric=neo4j.system.store.size.database, type=com.neo4j.metrics.source.db.StoreSizeMetrics$$Lambda$1156/0x00000008408cb840)
# TYPE neo4j_system_store_size_database gauge
neo4j_system_store_size_database 1155870.0
# HELP neo4j_system_pool_transaction_system_total_used Generated from Dropwizard metric import (metric=neo4j.system.pool.transaction.system.total_used, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$1172/0x00000008408cf840)
# TYPE neo4j_system_pool_transaction_system_total_used gauge
neo4j_system_pool_transaction_system_total_used 0.0
# HELP neo4j_neo4j_pool_transaction_neo4j_total_used Generated from Dropwizard metric import (metric=neo4j.neo4j.pool.transaction.neo4j.total_used, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$1172/0x00000008408cf840)
# TYPE neo4j_neo4j_pool_transaction_neo4j_total_used gauge
neo4j_neo4j_pool_transaction_neo4j_total_used 0.0
# HELP neo4j_vm_memory_pool_g1_eden_space Generated from Dropwizard metric import (metric=neo4j.vm.memory.pool.g1_eden_space, type=com.neo4j.metrics.source.jvm.JVMMemoryPoolMetrics$$Lambda$1424/0x000000084099b840)
# TYPE neo4j_vm_memory_pool_g1_eden_space gauge
neo4j_vm_memory_pool_g1_eden_space 3.07232768E8
# HELP neo4j_neo4j_transaction_active_write Generated from Dropwizard metric import (metric=neo4j.neo4j.transaction.active_write, type=com.neo4j.metrics.source.db.TransactionMetrics$$Lambda$1110/0x00000008408c0040)
# TYPE neo4j_neo4j_transaction_active_write gauge
neo4j_neo4j_transaction_active_write 0.0
# HELP neo4j_neo4j_check_point_duration Generated from Dropwizard metric import (metric=neo4j.neo4j.check_point.duration, type=com.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$1135/0x00000008408c6440)
# TYPE neo4j_neo4j_check_point_duration gauge
neo4j_neo4j_check_point_duration 0.0
# HELP neo4j_system_store_size_total Generated from Dropwizard metric import (metric=neo4j.system.store.size.total, type=com.neo4j.metrics.source.db.StoreSizeMetrics$$Lambda$1154/0x00000008408cb040)
# TYPE neo4j_system_store_size_total gauge
neo4j_system_store_size_total 2.63300318E8
# HELP neo4j_bolt_connections_closed_total Generated from Dropwizard metric import (metric=neo4j.bolt.connections_closed, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_bolt_connections_closed_total counter
neo4j_bolt_connections_closed_total 0.0
# HELP neo4j_bolt_messages_started_total Generated from Dropwizard metric import (metric=neo4j.bolt.messages_started, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_bolt_messages_started_total counter
neo4j_bolt_messages_started_total 0.0
# HELP neo4j_vm_thread_total Generated from Dropwizard metric import (metric=neo4j.vm.thread.total, type=com.neo4j.metrics.source.jvm.ThreadMetrics$$Lambda$1422/0x000000084099b040)
# TYPE neo4j_vm_thread_total gauge
neo4j_vm_thread_total 56.0
# HELP neo4j_system_transaction_rollbacks_write_total Generated from Dropwizard metric import (metric=neo4j.system.transaction.rollbacks_write, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_system_transaction_rollbacks_write_total counter
neo4j_system_transaction_rollbacks_write_total 0.0
# HELP neo4j_neo4j_transaction_last_committed_tx_id_total Generated from Dropwizard metric import (metric=neo4j.neo4j.transaction.last_committed_tx_id, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_neo4j_transaction_last_committed_tx_id_total counter
neo4j_neo4j_transaction_last_committed_tx_id_total 3.0
# HELP neo4j_dbms_pool_bolt_total_size Generated from Dropwizard metric import (metric=neo4j.dbms.pool.bolt.total_size, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$1457/0x00000008409a3c40)
# TYPE neo4j_dbms_pool_bolt_total_size gauge
neo4j_dbms_pool_bolt_total_size 9.223372036854776E18
# HELP neo4j_system_ids_in_use_relationship_type Generated from Dropwizard metric import (metric=neo4j.system.ids_in_use.relationship_type, type=com.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$1149/0x00000008408c9c40)
# TYPE neo4j_system_ids_in_use_relationship_type gauge
neo4j_system_ids_in_use_relationship_type 8.0
# HELP neo4j_vm_gc_time_g1_young_generation_total Generated from Dropwizard metric import (metric=neo4j.vm.gc.time.g1_young_generation, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_vm_gc_time_g1_young_generation_total counter
neo4j_vm_gc_time_g1_young_generation_total 266.0
# HELP neo4j_system_transaction_committed_total Generated from Dropwizard metric import (metric=neo4j.system.transaction.committed, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_system_transaction_committed_total counter
neo4j_system_transaction_committed_total 9.0
# HELP neo4j_vm_memory_pool_g1_old_gen Generated from Dropwizard metric import (metric=neo4j.vm.memory.pool.g1_old_gen, type=com.neo4j.metrics.source.jvm.JVMMemoryPoolMetrics$$Lambda$1424/0x000000084099b840)
# TYPE neo4j_vm_memory_pool_g1_old_gen gauge
neo4j_vm_memory_pool_g1_old_gen 2.0882904E7
# HELP neo4j_neo4j_store_size_database Generated from Dropwizard metric import (metric=neo4j.neo4j.store.size.database, type=com.neo4j.metrics.source.db.StoreSizeMetrics$$Lambda$1156/0x00000008408cb840)
# TYPE neo4j_neo4j_store_size_database gauge
neo4j_neo4j_store_size_database 869066.0
# HELP neo4j_neo4j_store_size_total Generated from Dropwizard metric import (metric=neo4j.neo4j.store.size.total, type=com.neo4j.metrics.source.db.StoreSizeMetrics$$Lambda$1154/0x00000008408cb040)
# TYPE neo4j_neo4j_store_size_total gauge
neo4j_neo4j_store_size_total 2.63013514E8
# HELP neo4j_vm_gc_time_g1_old_generation_total Generated from Dropwizard metric import (metric=neo4j.vm.gc.time.g1_old_generation, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_vm_gc_time_g1_old_generation_total counter
neo4j_vm_gc_time_g1_old_generation_total 0.0
# HELP neo4j_system_transaction_rollbacks_read_total Generated from Dropwizard metric import (metric=neo4j.system.transaction.rollbacks_read, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_system_transaction_rollbacks_read_total counter
neo4j_system_transaction_rollbacks_read_total 46.0
# HELP neo4j_neo4j_transaction_committed_write_total Generated from Dropwizard metric import (metric=neo4j.neo4j.transaction.committed_write, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_neo4j_transaction_committed_write_total counter
neo4j_neo4j_transaction_committed_write_total 0.0
# HELP neo4j_neo4j_transaction_rollbacks_write_total Generated from Dropwizard metric import (metric=neo4j.neo4j.transaction.rollbacks_write, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_neo4j_transaction_rollbacks_write_total counter
neo4j_neo4j_transaction_rollbacks_write_total 0.0
# HELP neo4j_neo4j_ids_in_use_node Generated from Dropwizard metric import (metric=neo4j.neo4j.ids_in_use.node, type=com.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$1143/0x00000008408c8440)
# TYPE neo4j_neo4j_ids_in_use_node gauge
neo4j_neo4j_ids_in_use_node 0.0
# HELP neo4j_neo4j_db_query_execution_latency_millis Generated from Dropwizard metric import (metric=neo4j.neo4j.db.query.execution.latency.millis, type=com.codahale.metrics.Histogram)
# TYPE neo4j_neo4j_db_query_execution_latency_millis summary
neo4j_neo4j_db_query_execution_latency_millis{quantile="0.5",} 0.0
neo4j_neo4j_db_query_execution_latency_millis{quantile="0.75",} 0.0
neo4j_neo4j_db_query_execution_latency_millis{quantile="0.95",} 0.0
neo4j_neo4j_db_query_execution_latency_millis{quantile="0.98",} 0.0
neo4j_neo4j_db_query_execution_latency_millis{quantile="0.99",} 0.0
neo4j_neo4j_db_query_execution_latency_millis{quantile="0.999",} 0.0
neo4j_neo4j_db_query_execution_latency_millis_count 0.0
# HELP neo4j_system_transaction_committed_write_total Generated from Dropwizard metric import (metric=neo4j.system.transaction.committed_write, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_system_transaction_committed_write_total counter
neo4j_system_transaction_committed_write_total 0.0
# HELP neo4j_page_cache_hit_ratio Generated from Dropwizard metric import (metric=neo4j.page_cache.hit_ratio, type=com.neo4j.metrics.source.db.PageCacheHitRatioGauge)
# TYPE neo4j_page_cache_hit_ratio gauge
neo4j_page_cache_hit_ratio 1.0
# HELP neo4j_bolt_connections_idle Generated from Dropwizard metric import (metric=neo4j.bolt.connections_idle, type=com.neo4j.metrics.source.db.BoltMetrics$$Lambda$1442/0x00000008409a0040)
# TYPE neo4j_bolt_connections_idle gauge
neo4j_bolt_connections_idle 0.0
# HELP neo4j_page_cache_page_faults_total Generated from Dropwizard metric import (metric=neo4j.page_cache.page_faults, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_page_cache_page_faults_total counter
neo4j_page_cache_page_faults_total 283.0
# HELP neo4j_neo4j_ids_in_use_relationship_type Generated from Dropwizard metric import (metric=neo4j.neo4j.ids_in_use.relationship_type, type=com.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$1149/0x00000008408c9c40)
# TYPE neo4j_neo4j_ids_in_use_relationship_type gauge
neo4j_neo4j_ids_in_use_relationship_type 0.0
# HELP neo4j_system_ids_in_use_property Generated from Dropwizard metric import (metric=neo4j.system.ids_in_use.property, type=com.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$1147/0x00000008408c9440)
# TYPE neo4j_system_ids_in_use_property gauge
neo4j_system_ids_in_use_property 126.0
# HELP neo4j_neo4j_transaction_active_read Generated from Dropwizard metric import (metric=neo4j.neo4j.transaction.active_read, type=com.neo4j.metrics.source.db.TransactionMetrics$$Lambda$1108/0x00000008408a5840)
# TYPE neo4j_neo4j_transaction_active_read gauge
neo4j_neo4j_transaction_active_read 0.0
# HELP neo4j_neo4j_transaction_rollbacks_read_total Generated from Dropwizard metric import (metric=neo4j.neo4j.transaction.rollbacks_read, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_neo4j_transaction_rollbacks_read_total counter
neo4j_neo4j_transaction_rollbacks_read_total 0.0
# HELP neo4j_neo4j_pool_transaction_neo4j_used_native Generated from Dropwizard metric import (metric=neo4j.neo4j.pool.transaction.neo4j.used_native, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$1171/0x00000008408cf440)
# TYPE neo4j_neo4j_pool_transaction_neo4j_used_native gauge
neo4j_neo4j_pool_transaction_neo4j_used_native 0.0
# HELP neo4j_bolt_connections_opened_total Generated from Dropwizard metric import (metric=neo4j.bolt.connections_opened, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_bolt_connections_opened_total counter
neo4j_bolt_connections_opened_total 0.0
# HELP neo4j_neo4j_cypher_replan_events_total Generated from Dropwizard metric import (metric=neo4j.neo4j.cypher.replan_events, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_neo4j_cypher_replan_events_total counter
neo4j_neo4j_cypher_replan_events_total 0.0
# HELP neo4j_page_cache_usage_ratio Generated from Dropwizard metric import (metric=neo4j.page_cache.usage_ratio, type=com.neo4j.metrics.source.db.PageCacheMetrics$$Lambda$1406/0x000000084098c040)
# TYPE neo4j_page_cache_usage_ratio gauge
neo4j_page_cache_usage_ratio 0.0029871323529411763
# HELP neo4j_neo4j_db_query_execution_failure_total Generated from Dropwizard metric import (metric=neo4j.neo4j.db.query.execution.failure, type=com.codahale.metrics.Meter)
# TYPE neo4j_neo4j_db_query_execution_failure_total counter
neo4j_neo4j_db_query_execution_failure_total 0.0
# HELP neo4j_neo4j_transaction_committed_total Generated from Dropwizard metric import (metric=neo4j.neo4j.transaction.committed, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_neo4j_transaction_committed_total counter
neo4j_neo4j_transaction_committed_total 0.0
# HELP neo4j_neo4j_db_query_execution_success_total Generated from Dropwizard metric import (metric=neo4j.neo4j.db.query.execution.success, type=com.codahale.metrics.Meter)
# TYPE neo4j_neo4j_db_query_execution_success_total counter
neo4j_neo4j_db_query_execution_success_total 0.0
# HELP neo4j_neo4j_pool_transaction_neo4j_used_heap Generated from Dropwizard metric import (metric=neo4j.neo4j.pool.transaction.neo4j.used_heap, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$1170/0x00000008408cf040)
# TYPE neo4j_neo4j_pool_transaction_neo4j_used_heap gauge
neo4j_neo4j_pool_transaction_neo4j_used_heap 0.0
# HELP neo4j_system_ids_in_use_relationship Generated from Dropwizard metric import (metric=neo4j.system.ids_in_use.relationship, type=com.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$1145/0x00000008408c8c40)
# TYPE neo4j_system_ids_in_use_relationship gauge
neo4j_system_ids_in_use_relationship 101.0
# HELP neo4j_page_cache_hits_total Generated from Dropwizard metric import (metric=neo4j.page_cache.hits, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_page_cache_hits_total counter
neo4j_page_cache_hits_total 1648.0
# HELP neo4j_system_transaction_last_committed_tx_id_total Generated from Dropwizard metric import (metric=neo4j.system.transaction.last_committed_tx_id, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_system_transaction_last_committed_tx_id_total counter
neo4j_system_transaction_last_committed_tx_id_total 73.0
# HELP neo4j_neo4j_transaction_peak_concurrent_total Generated from Dropwizard metric import (metric=neo4j.neo4j.transaction.peak_concurrent, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_neo4j_transaction_peak_concurrent_total counter
neo4j_neo4j_transaction_peak_concurrent_total 0.0
# HELP neo4j_vm_thread_count Generated from Dropwizard metric import (metric=neo4j.vm.thread.count, type=com.neo4j.metrics.source.jvm.ThreadMetrics$$Lambda$1420/0x000000084099a840)
# TYPE neo4j_vm_thread_count gauge
neo4j_vm_thread_count 51.0
# HELP neo4j_dbms_pool_bolt_used_heap Generated from Dropwizard metric import (metric=neo4j.dbms.pool.bolt.used_heap, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$1170/0x00000008408cf040)
# TYPE neo4j_dbms_pool_bolt_used_heap gauge
neo4j_dbms_pool_bolt_used_heap 0.0
# HELP neo4j_dbms_pool_bolt_free Generated from Dropwizard metric import (metric=neo4j.dbms.pool.bolt.free, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$1458/0x00000008409a4040)
# TYPE neo4j_dbms_pool_bolt_free gauge
neo4j_dbms_pool_bolt_free 9.223372036854776E18
# HELP neo4j_dbms_pool_bolt_total_used Generated from Dropwizard metric import (metric=neo4j.dbms.pool.bolt.total_used, type=com.neo4j.metrics.source.db.AbstractMemoryPoolMetrics$$Lambda$1172/0x00000008408cf840)
# TYPE neo4j_dbms_pool_bolt_total_used gauge
neo4j_dbms_pool_bolt_total_used 0.0
# HELP neo4j_neo4j_ids_in_use_relationship Generated from Dropwizard metric import (metric=neo4j.neo4j.ids_in_use.relationship, type=com.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$1145/0x00000008408c8c40)
# TYPE neo4j_neo4j_ids_in_use_relationship gauge
neo4j_neo4j_ids_in_use_relationship 0.0
# HELP neo4j_system_ids_in_use_node Generated from Dropwizard metric import (metric=neo4j.system.ids_in_use.node, type=com.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$1143/0x00000008408c8440)
# TYPE neo4j_system_ids_in_use_node gauge
neo4j_system_ids_in_use_node 55.0
# HELP neo4j_system_transaction_peak_concurrent_total Generated from Dropwizard metric import (metric=neo4j.system.transaction.peak_concurrent, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_system_transaction_peak_concurrent_total counter
neo4j_system_transaction_peak_concurrent_total 1.0
# HELP neo4j_bolt_messages_received_total Generated from Dropwizard metric import (metric=neo4j.bolt.messages_received, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_bolt_messages_received_total counter
neo4j_bolt_messages_received_total 0.0
# HELP neo4j_neo4j_transaction_committed_read_total Generated from Dropwizard metric import (metric=neo4j.neo4j.transaction.committed_read, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_neo4j_transaction_committed_read_total counter
neo4j_neo4j_transaction_committed_read_total 0.0
# HELP neo4j_system_check_point_total_time_total Generated from Dropwizard metric import (metric=neo4j.system.check_point.total_time, type=com.neo4j.metrics.metric.MetricsCounter)
# TYPE neo4j_system_check_point_total_time_total counter
neo4j_system_check_point_total_time_total 0.0
`

var want4_4_0 []string = []string{
	"neo4j bolt_connections_closed_total=0",
	"neo4j,db=neo4j store_size_database=869066",
	"neo4j,db=neo4j ids_in_use_relationship_type=0",
	"neo4j bolt_connections_opened_total=0",
	"neo4j,db=neo4j transaction_peak_concurrent_total=0",
	"neo4j,db=neo4j ids_in_use_property=8",
	"neo4j,db=neo4j transaction_committed_read_total=0",
	"neo4j,database=neo4j,db=neo4j,pool=transaction pool_used_heap=0",
	"neo4j,database=system,db=system,pool=transaction pool_used_native=0",
	"neo4j page_cache_usage_ratio=0.0029871323529411763",
	"neo4j,db=system ids_in_use_relationship=101",
	"neo4j,pool=bolt dbms_pool_free=9223372036854776000",
	"neo4j bolt_connections_running=0",
	"neo4j,pool=bolt dbms_pool_total_size=9223372036854776000",
	"neo4j bolt_connections_idle=0",
	"neo4j,db=neo4j ids_in_use_relationship=0",
	"neo4j,db=system db_query_execution_latency_millis_count=0,db_query_execution_latency_millis_sum=0",
	"neo4j,db=system,quantile=0.5 db_query_execution_latency_millis=0",
	"neo4j,db=system,quantile=0.75 db_query_execution_latency_millis=0",
	"neo4j,db=system,quantile=0.95 db_query_execution_latency_millis=0",
	"neo4j,db=system,quantile=0.98 db_query_execution_latency_millis=0",
	"neo4j,db=system,quantile=0.99 db_query_execution_latency_millis=0",
	"neo4j,db=system,quantile=0.999 db_query_execution_latency_millis=0",
	"neo4j,db=neo4j transaction_rollbacks_write_total=0",
	"neo4j,db=system ids_in_use_property=126",
	"neo4j,db=neo4j cypher_replan_events_total=0",
	"neo4j,pool=bolt dbms_pool_used_heap=0",
	"neo4j,db=neo4j transaction_committed_write_total=0",
	"neo4j vm_heap_used=113157592",
	"neo4j,db=system store_size_database=1155870",
	"neo4j,gc=g1_old_generation vm_gc_time_total=0",
	"neo4j page_cache_hit_ratio=1",
	"neo4j,db=system ids_in_use_node=55",
	"neo4j,db=system db_query_execution_success_total=0",
	"neo4j,db=neo4j transaction_rollbacks_total=0",
	"neo4j,bufferpool=direct vm_memory_buffer_used=1728516",
	"neo4j,db=neo4j transaction_last_committed_tx_id_total=3",
	"neo4j,gc=g1_young_generation vm_gc_time_total=266",
	"neo4j,database=system,db=system,pool=transaction pool_used_heap=0",
	"neo4j,db=system ids_in_use_relationship_type=8",
	"neo4j,db=neo4j ids_in_use_node=0",
	"neo4j,database=neo4j,db=neo4j,pool=transaction pool_used_native=0",
	"neo4j,db=neo4j transaction_committed_total=0",
	"neo4j page_cache_hits_total=1648",
	"neo4j vm_thread_count=51",
	"neo4j,db=system check_point_total_time_total=0",
	"neo4j vm_thread_total=56",
	"neo4j page_cache_page_faults_total=283",
	"neo4j vm_pause_time=0",
	"neo4j vm_file_descriptors_count=526",
	"neo4j,db=neo4j check_point_total_time_total=0",
	"neo4j,db=system transaction_committed_read_total=9",
	"neo4j bolt_messages_received_total=0",
	"neo4j,db=system cypher_replan_events_total=0",
	"neo4j,pool=g1_old_gen vm_memory_pool=20882904",
	"neo4j,db=system transaction_rollbacks_total=46",
	"neo4j,db=system transaction_active_write=0",
	"neo4j,db=system check_point_duration=0",
	"neo4j,database=neo4j,db=neo4j,pool=transaction pool_total_used=0",
	"neo4j,pool=g1_eden_space vm_memory_pool=307232768",
	"neo4j,db=neo4j transaction_active_write=0",
	"neo4j,db=neo4j check_point_duration=0",
	"neo4j,db=system store_size_total=263300318",
	"neo4j,db=system db_query_execution_failure_total=0",
	"neo4j,db=neo4j store_size_total=263013514",
	"neo4j,db=system transaction_committed_write_total=0",
	"neo4j,db=system transaction_committed_total=9",
	"neo4j,db=system transaction_peak_concurrent_total=1",
	"neo4j,db=system transaction_rollbacks_write_total=0",
	"neo4j,db=neo4j transaction_active_read=0",
	"neo4j,db=neo4j transaction_rollbacks_read_total=0",
	"neo4j,db=neo4j db_query_execution_success_total=0",
	"neo4j bolt_messages_started_total=0",
	"neo4j,database=system,db=system,pool=transaction pool_total_used=0",
	"neo4j,db=neo4j db_query_execution_failure_total=0",
	"neo4j,db=system transaction_last_committed_tx_id_total=73",
	"neo4j,db=system transaction_active_read=0",
	"neo4j,db=neo4j db_query_execution_latency_millis_count=0,db_query_execution_latency_millis_sum=0",
	"neo4j,db=neo4j,quantile=0.5 db_query_execution_latency_millis=0",
	"neo4j,db=neo4j,quantile=0.75 db_query_execution_latency_millis=0",
	"neo4j,db=neo4j,quantile=0.95 db_query_execution_latency_millis=0",
	"neo4j,db=neo4j,quantile=0.98 db_query_execution_latency_millis=0",
	"neo4j,db=neo4j,quantile=0.99 db_query_execution_latency_millis=0",
	"neo4j,db=neo4j,quantile=0.999 db_query_execution_latency_millis=0",
	"neo4j,pool=bolt dbms_pool_total_used=0",
	"neo4j,db=system transaction_rollbacks_read_total=46",
}

var mock3_4_0 string = `# HELP neo4j_bolt_accumulated_processing_time Generated from Dropwizard metric import (metric=neo4j.bolt.accumulated_processing_time, type=org.neo4j.metrics.source.db.BoltMetrics$$Lambda$420/30563356)
# TYPE neo4j_bolt_accumulated_processing_time gauge
neo4j_bolt_accumulated_processing_time 0.0
# HELP neo4j_bolt_accumulated_queue_time Generated from Dropwizard metric import (metric=neo4j.bolt.accumulated_queue_time, type=org.neo4j.metrics.source.db.BoltMetrics$$Lambda$419/103394942)
# TYPE neo4j_bolt_accumulated_queue_time gauge
neo4j_bolt_accumulated_queue_time 0.0
# HELP neo4j_bolt_connections_closed Generated from Dropwizard metric import (metric=neo4j.bolt.connections_closed, type=org.neo4j.metrics.source.db.BoltMetrics$$Lambda$412/732118572)
# TYPE neo4j_bolt_connections_closed gauge
neo4j_bolt_connections_closed 0.0
# HELP neo4j_bolt_connections_idle Generated from Dropwizard metric import (metric=neo4j.bolt.connections_idle, type=org.neo4j.metrics.source.db.BoltMetrics$$Lambda$414/837233852)
# TYPE neo4j_bolt_connections_idle gauge
neo4j_bolt_connections_idle 0.0
# HELP neo4j_bolt_connections_opened Generated from Dropwizard metric import (metric=neo4j.bolt.connections_opened, type=org.neo4j.metrics.source.db.BoltMetrics$$Lambda$411/345986913)
# TYPE neo4j_bolt_connections_opened gauge
neo4j_bolt_connections_opened 0.0
# HELP neo4j_bolt_connections_running Generated from Dropwizard metric import (metric=neo4j.bolt.connections_running, type=org.neo4j.metrics.source.db.BoltMetrics$$Lambda$413/521746054)
# TYPE neo4j_bolt_connections_running gauge
neo4j_bolt_connections_running 0.0
# HELP neo4j_bolt_messages_done Generated from Dropwizard metric import (metric=neo4j.bolt.messages_done, type=org.neo4j.metrics.source.db.BoltMetrics$$Lambda$417/993452032)
# TYPE neo4j_bolt_messages_done gauge
neo4j_bolt_messages_done 0.0
# HELP neo4j_bolt_messages_failed Generated from Dropwizard metric import (metric=neo4j.bolt.messages_failed, type=org.neo4j.metrics.source.db.BoltMetrics$$Lambda$418/859617558)
# TYPE neo4j_bolt_messages_failed gauge
neo4j_bolt_messages_failed 0.0
# HELP neo4j_bolt_messages_received Generated from Dropwizard metric import (metric=neo4j.bolt.messages_received, type=org.neo4j.metrics.source.db.BoltMetrics$$Lambda$415/1605190078)
# TYPE neo4j_bolt_messages_received gauge
neo4j_bolt_messages_received 0.0
# HELP neo4j_bolt_messages_started Generated from Dropwizard metric import (metric=neo4j.bolt.messages_started, type=org.neo4j.metrics.source.db.BoltMetrics$$Lambda$416/1842173497)
# TYPE neo4j_bolt_messages_started gauge
neo4j_bolt_messages_started 0.0
# HELP neo4j_bolt_sessions_started Generated from Dropwizard metric import (metric=neo4j.bolt.sessions_started, type=org.neo4j.metrics.source.db.BoltMetrics$$Lambda$410/1746458880)
# TYPE neo4j_bolt_sessions_started gauge
neo4j_bolt_sessions_started 0.0
# HELP neo4j_check_point_events Generated from Dropwizard metric import (metric=neo4j.check_point.events, type=org.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$393/1035357140)
# TYPE neo4j_check_point_events gauge
neo4j_check_point_events 0.0
# HELP neo4j_check_point_total_time Generated from Dropwizard metric import (metric=neo4j.check_point.total_time, type=org.neo4j.metrics.source.db.CheckPointingMetrics$$Lambda$394/1846568576)
# TYPE neo4j_check_point_total_time gauge
neo4j_check_point_total_time 0.0
# HELP neo4j_cypher_replan_events Generated from Dropwizard metric import (metric=neo4j.cypher.replan_events, type=org.neo4j.metrics.source.db.CypherMetrics$$Lambda$404/582702662)
# TYPE neo4j_cypher_replan_events gauge
neo4j_cypher_replan_events 0.0
# HELP neo4j_cypher_replan_wait_time Generated from Dropwizard metric import (metric=neo4j.cypher.replan_wait_time, type=org.neo4j.metrics.source.db.CypherMetrics$$Lambda$405/468033320)
# TYPE neo4j_cypher_replan_wait_time gauge
neo4j_cypher_replan_wait_time 0.0
# HELP neo4j_ids_in_use_node Generated from Dropwizard metric import (metric=neo4j.ids_in_use.node, type=org.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$397/935520971)
# TYPE neo4j_ids_in_use_node gauge
neo4j_ids_in_use_node 0.0
# HELP neo4j_ids_in_use_property Generated from Dropwizard metric import (metric=neo4j.ids_in_use.property, type=org.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$399/851033362)
# TYPE neo4j_ids_in_use_property gauge
neo4j_ids_in_use_property 0.0
# HELP neo4j_ids_in_use_relationship Generated from Dropwizard metric import (metric=neo4j.ids_in_use.relationship, type=org.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$398/36883680)
# TYPE neo4j_ids_in_use_relationship gauge
neo4j_ids_in_use_relationship 0.0
# HELP neo4j_ids_in_use_relationship_type Generated from Dropwizard metric import (metric=neo4j.ids_in_use.relationship_type, type=org.neo4j.metrics.source.db.EntityCountMetrics$$Lambda$400/892237946)
# TYPE neo4j_ids_in_use_relationship_type gauge
neo4j_ids_in_use_relationship_type 0.0
# HELP neo4j_log_rotation_events Generated from Dropwizard metric import (metric=neo4j.log_rotation.events, type=org.neo4j.metrics.source.db.LogRotationMetrics$$Lambda$395/1441577726)
# TYPE neo4j_log_rotation_events gauge
neo4j_log_rotation_events 0.0
# HELP neo4j_log_rotation_total_time Generated from Dropwizard metric import (metric=neo4j.log_rotation.total_time, type=org.neo4j.metrics.source.db.LogRotationMetrics$$Lambda$396/1519100796)
# TYPE neo4j_log_rotation_total_time gauge
neo4j_log_rotation_total_time 0.0
# HELP neo4j_network_master_network_store_writes Generated from Dropwizard metric import (metric=neo4j.network.master_network_store_writes, type=org.neo4j.metrics.source.cluster.NetworkMetrics$$Lambda$402/2075093711)
# TYPE neo4j_network_master_network_store_writes gauge
neo4j_network_master_network_store_writes 0.0
# HELP neo4j_network_master_network_tx_writes Generated from Dropwizard metric import (metric=neo4j.network.master_network_tx_writes, type=org.neo4j.metrics.source.cluster.NetworkMetrics$$Lambda$401/757779849)
# TYPE neo4j_network_master_network_tx_writes gauge
neo4j_network_master_network_tx_writes 0.0
# HELP neo4j_network_slave_network_tx_writes Generated from Dropwizard metric import (metric=neo4j.network.slave_network_tx_writes, type=org.neo4j.metrics.source.cluster.NetworkMetrics$$Lambda$403/31906520)
# TYPE neo4j_network_slave_network_tx_writes gauge
neo4j_network_slave_network_tx_writes 0.0
# HELP neo4j_page_cache_eviction_exceptions Generated from Dropwizard metric import (metric=neo4j.page_cache.eviction_exceptions, type=org.neo4j.metrics.source.db.PageCacheMetrics$$Lambda$390/1198158701)
# TYPE neo4j_page_cache_eviction_exceptions gauge
neo4j_page_cache_eviction_exceptions 0.0
# HELP neo4j_page_cache_evictions Generated from Dropwizard metric import (metric=neo4j.page_cache.evictions, type=org.neo4j.metrics.source.db.PageCacheMetrics$$Lambda$385/4015102)
# TYPE neo4j_page_cache_evictions gauge
neo4j_page_cache_evictions 0.0
# HELP neo4j_page_cache_flushes Generated from Dropwizard metric import (metric=neo4j.page_cache.flushes, type=org.neo4j.metrics.source.db.PageCacheMetrics$$Lambda$389/1605834811)
# TYPE neo4j_page_cache_flushes gauge
neo4j_page_cache_flushes 1.0
# HELP neo4j_page_cache_hit_ratio Generated from Dropwizard metric import (metric=neo4j.page_cache.hit_ratio, type=org.neo4j.metrics.source.db.PageCacheMetrics$$Lambda$391/382252989)
# TYPE neo4j_page_cache_hit_ratio gauge
neo4j_page_cache_hit_ratio 0.6724137931034483
# HELP neo4j_page_cache_hits Generated from Dropwizard metric import (metric=neo4j.page_cache.hits, type=org.neo4j.metrics.source.db.PageCacheMetrics$$Lambda$388/905940937)
# TYPE neo4j_page_cache_hits gauge
neo4j_page_cache_hits 39.0
# HELP neo4j_page_cache_page_faults Generated from Dropwizard metric import (metric=neo4j.page_cache.page_faults, type=org.neo4j.metrics.source.db.PageCacheMetrics$$Lambda$384/1322484262)
# TYPE neo4j_page_cache_page_faults gauge
neo4j_page_cache_page_faults 19.0
# HELP neo4j_page_cache_pins Generated from Dropwizard metric import (metric=neo4j.page_cache.pins, type=org.neo4j.metrics.source.db.PageCacheMetrics$$Lambda$386/1957530885)
# TYPE neo4j_page_cache_pins gauge
neo4j_page_cache_pins 110.0
# HELP neo4j_page_cache_unpins Generated from Dropwizard metric import (metric=neo4j.page_cache.unpins, type=org.neo4j.metrics.source.db.PageCacheMetrics$$Lambda$387/1735390128)
# TYPE neo4j_page_cache_unpins gauge
neo4j_page_cache_unpins 57.0
# HELP neo4j_page_cache_usage_ratio Generated from Dropwizard metric import (metric=neo4j.page_cache.usage_ratio, type=org.neo4j.metrics.source.db.PageCacheMetrics$$Lambda$392/1901663135)
# TYPE neo4j_page_cache_usage_ratio gauge
neo4j_page_cache_usage_ratio 2.9105392156862745E-4
# HELP neo4j_server_threads_jetty_all Generated from Dropwizard metric import (metric=neo4j.server.threads.jetty.all, type=org.neo4j.metrics.source.server.ServerMetrics$$Lambda$426/153443333)
# TYPE neo4j_server_threads_jetty_all gauge
neo4j_server_threads_jetty_all 12.0
# HELP neo4j_server_threads_jetty_idle Generated from Dropwizard metric import (metric=neo4j.server.threads.jetty.idle, type=org.neo4j.metrics.source.server.ServerMetrics$$Lambda$425/1123862502)
# TYPE neo4j_server_threads_jetty_idle gauge
neo4j_server_threads_jetty_idle 2.0
# HELP neo4j_transaction_active Generated from Dropwizard metric import (metric=neo4j.transaction.active, type=org.neo4j.metrics.source.db.TransactionMetrics$$Lambda$370/1585841343)
# TYPE neo4j_transaction_active gauge
neo4j_transaction_active 0.0
# HELP neo4j_transaction_active_read Generated from Dropwizard metric import (metric=neo4j.transaction.active_read, type=org.neo4j.metrics.source.db.TransactionMetrics$$Lambda$371/537483956)
# TYPE neo4j_transaction_active_read gauge
neo4j_transaction_active_read 0.0
# HELP neo4j_transaction_active_write Generated from Dropwizard metric import (metric=neo4j.transaction.active_write, type=org.neo4j.metrics.source.db.TransactionMetrics$$Lambda$372/1311315651)
# TYPE neo4j_transaction_active_write gauge
neo4j_transaction_active_write 0.0
# HELP neo4j_transaction_committed Generated from Dropwizard metric import (metric=neo4j.transaction.committed, type=org.neo4j.metrics.source.db.TransactionMetrics$$Lambda$373/1688917723)
# TYPE neo4j_transaction_committed gauge
neo4j_transaction_committed 2.0
# HELP neo4j_transaction_committed_read Generated from Dropwizard metric import (metric=neo4j.transaction.committed_read, type=org.neo4j.metrics.source.db.TransactionMetrics$$Lambda$374/182949133)
# TYPE neo4j_transaction_committed_read gauge
neo4j_transaction_committed_read 2.0
# HELP neo4j_transaction_committed_write Generated from Dropwizard metric import (metric=neo4j.transaction.committed_write, type=org.neo4j.metrics.source.db.TransactionMetrics$$Lambda$375/1624355359)
# TYPE neo4j_transaction_committed_write gauge
neo4j_transaction_committed_write 0.0
# HELP neo4j_transaction_last_closed_tx_id Generated from Dropwizard metric import (metric=neo4j.transaction.last_closed_tx_id, type=org.neo4j.metrics.source.db.TransactionMetrics$$Lambda$383/2124261761)
# TYPE neo4j_transaction_last_closed_tx_id gauge
neo4j_transaction_last_closed_tx_id 1.0
# HELP neo4j_transaction_last_committed_tx_id Generated from Dropwizard metric import (metric=neo4j.transaction.last_committed_tx_id, type=org.neo4j.metrics.source.db.TransactionMetrics$$Lambda$382/1077316166)
# TYPE neo4j_transaction_last_committed_tx_id gauge
neo4j_transaction_last_committed_tx_id 1.0
# HELP neo4j_transaction_peak_concurrent Generated from Dropwizard metric import (metric=neo4j.transaction.peak_concurrent, type=org.neo4j.metrics.source.db.TransactionMetrics$$Lambda$369/1248598189)
# TYPE neo4j_transaction_peak_concurrent gauge
neo4j_transaction_peak_concurrent 1.0
# HELP neo4j_transaction_rollbacks Generated from Dropwizard metric import (metric=neo4j.transaction.rollbacks, type=org.neo4j.metrics.source.db.TransactionMetrics$$Lambda$376/1724399560)
# TYPE neo4j_transaction_rollbacks gauge
neo4j_transaction_rollbacks 0.0
# HELP neo4j_transaction_rollbacks_read Generated from Dropwizard metric import (metric=neo4j.transaction.rollbacks_read, type=org.neo4j.metrics.source.db.TransactionMetrics$$Lambda$377/1415979460)
# TYPE neo4j_transaction_rollbacks_read gauge
neo4j_transaction_rollbacks_read 0.0
# HELP neo4j_transaction_rollbacks_write Generated from Dropwizard metric import (metric=neo4j.transaction.rollbacks_write, type=org.neo4j.metrics.source.db.TransactionMetrics$$Lambda$378/1646234040)
# TYPE neo4j_transaction_rollbacks_write gauge
neo4j_transaction_rollbacks_write 0.0
# HELP neo4j_transaction_started Generated from Dropwizard metric import (metric=neo4j.transaction.started, type=org.neo4j.metrics.source.db.TransactionMetrics$$Lambda$368/117249632)
# TYPE neo4j_transaction_started gauge
neo4j_transaction_started 2.0
# HELP neo4j_transaction_terminated Generated from Dropwizard metric import (metric=neo4j.transaction.terminated, type=org.neo4j.metrics.source.db.TransactionMetrics$$Lambda$379/255041198)
# TYPE neo4j_transaction_terminated gauge
neo4j_transaction_terminated 0.0
# HELP neo4j_transaction_terminated_read Generated from Dropwizard metric import (metric=neo4j.transaction.terminated_read, type=org.neo4j.metrics.source.db.TransactionMetrics$$Lambda$380/673367807)
# TYPE neo4j_transaction_terminated_read gauge
neo4j_transaction_terminated_read 0.0
# HELP neo4j_transaction_terminated_write Generated from Dropwizard metric import (metric=neo4j.transaction.terminated_write, type=org.neo4j.metrics.source.db.TransactionMetrics$$Lambda$381/1303362110)
# TYPE neo4j_transaction_terminated_write gauge
neo4j_transaction_terminated_write 0.0
# HELP vm_gc_count_g1_old_generation Generated from Dropwizard metric import (metric=vm.gc.count.g1_old_generation, type=org.neo4j.metrics.source.jvm.GCMetrics$$Lambda$407/356338363)
# TYPE vm_gc_count_g1_old_generation gauge
vm_gc_count_g1_old_generation 0.0
# HELP vm_gc_count_g1_young_generation Generated from Dropwizard metric import (metric=vm.gc.count.g1_young_generation, type=org.neo4j.metrics.source.jvm.GCMetrics$$Lambda$407/356338363)
# TYPE vm_gc_count_g1_young_generation gauge
vm_gc_count_g1_young_generation 7.0
# HELP vm_gc_time_g1_old_generation Generated from Dropwizard metric import (metric=vm.gc.time.g1_old_generation, type=org.neo4j.metrics.source.jvm.GCMetrics$$Lambda$406/753162875)
# TYPE vm_gc_time_g1_old_generation gauge
vm_gc_time_g1_old_generation 0.0
# HELP vm_gc_time_g1_young_generation Generated from Dropwizard metric import (metric=vm.gc.time.g1_young_generation, type=org.neo4j.metrics.source.jvm.GCMetrics$$Lambda$406/753162875)
# TYPE vm_gc_time_g1_young_generation gauge
vm_gc_time_g1_young_generation 187.0
# HELP vm_memory_buffer_direct_capacity Generated from Dropwizard metric import (metric=vm.memory.buffer.direct.capacity, type=org.neo4j.metrics.source.jvm.MemoryBuffersMetrics$$Lambda$424/1242427797)
# TYPE vm_memory_buffer_direct_capacity gauge
vm_memory_buffer_direct_capacity 221184.0
# HELP vm_memory_buffer_direct_count Generated from Dropwizard metric import (metric=vm.memory.buffer.direct.count, type=org.neo4j.metrics.source.jvm.MemoryBuffersMetrics$$Lambda$422/274426173)
# TYPE vm_memory_buffer_direct_count gauge
vm_memory_buffer_direct_count 7.0
# HELP vm_memory_buffer_direct_used Generated from Dropwizard metric import (metric=vm.memory.buffer.direct.used, type=org.neo4j.metrics.source.jvm.MemoryBuffersMetrics$$Lambda$423/66774422)
# TYPE vm_memory_buffer_direct_used gauge
vm_memory_buffer_direct_used 221185.0
# HELP vm_memory_buffer_mapped_capacity Generated from Dropwizard metric import (metric=vm.memory.buffer.mapped.capacity, type=org.neo4j.metrics.source.jvm.MemoryBuffersMetrics$$Lambda$424/1242427797)
# TYPE vm_memory_buffer_mapped_capacity gauge
vm_memory_buffer_mapped_capacity 0.0
# HELP vm_memory_buffer_mapped_count Generated from Dropwizard metric import (metric=vm.memory.buffer.mapped.count, type=org.neo4j.metrics.source.jvm.MemoryBuffersMetrics$$Lambda$422/274426173)
# TYPE vm_memory_buffer_mapped_count gauge
vm_memory_buffer_mapped_count 0.0
# HELP vm_memory_buffer_mapped_used Generated from Dropwizard metric import (metric=vm.memory.buffer.mapped.used, type=org.neo4j.metrics.source.jvm.MemoryBuffersMetrics$$Lambda$423/66774422)
# TYPE vm_memory_buffer_mapped_used gauge
vm_memory_buffer_mapped_used 0.0
# HELP vm_memory_pool_code_cache Generated from Dropwizard metric import (metric=vm.memory.pool.code_cache, type=org.neo4j.metrics.source.jvm.MemoryPoolMetrics$$Lambda$421/1539995236)
# TYPE vm_memory_pool_code_cache gauge
vm_memory_pool_code_cache 1.2113088E7
# HELP vm_memory_pool_compressed_class_space Generated from Dropwizard metric import (metric=vm.memory.pool.compressed_class_space, type=org.neo4j.metrics.source.jvm.MemoryPoolMetrics$$Lambda$421/1539995236)
# TYPE vm_memory_pool_compressed_class_space gauge
vm_memory_pool_compressed_class_space 8315352.0
# HELP vm_memory_pool_g1_eden_space Generated from Dropwizard metric import (metric=vm.memory.pool.g1_eden_space, type=org.neo4j.metrics.source.jvm.MemoryPoolMetrics$$Lambda$421/1539995236)
# TYPE vm_memory_pool_g1_eden_space gauge
vm_memory_pool_g1_eden_space 1.59383552E8
# HELP vm_memory_pool_g1_old_gen Generated from Dropwizard metric import (metric=vm.memory.pool.g1_old_gen, type=org.neo4j.metrics.source.jvm.MemoryPoolMetrics$$Lambda$421/1539995236)
# TYPE vm_memory_pool_g1_old_gen gauge
vm_memory_pool_g1_old_gen 1.4218288E7
# HELP vm_memory_pool_g1_survivor_space Generated from Dropwizard metric import (metric=vm.memory.pool.g1_survivor_space, type=org.neo4j.metrics.source.jvm.MemoryPoolMetrics$$Lambda$421/1539995236)
# TYPE vm_memory_pool_g1_survivor_space gauge
vm_memory_pool_g1_survivor_space 3.3554432E7
# HELP vm_memory_pool_metaspace Generated from Dropwizard metric import (metric=vm.memory.pool.metaspace, type=org.neo4j.metrics.source.jvm.MemoryPoolMetrics$$Lambda$421/1539995236)
# TYPE vm_memory_pool_metaspace gauge
vm_memory_pool_metaspace 5.711568E7
# HELP vm_thread_count Generated from Dropwizard metric import (metric=vm.thread.count, type=org.neo4j.metrics.source.jvm.ThreadMetrics$$Lambda$408/1516759394)
# TYPE vm_thread_count gauge
vm_thread_count 44.0
# HELP vm_thread_total Generated from Dropwizard metric import (metric=vm.thread.total, type=org.neo4j.metrics.source.jvm.ThreadMetrics$$Lambda$409/1415469015)
# TYPE vm_thread_total gauge
vm_thread_total 47.0
`

var want3_4_0 []string = []string{
	"neo4j transaction_committed_read=2",
	"neo4j vm_thread_total=47",
	"neo4j log_rotation_total_time=0",
	"neo4j page_cache_evictions=0",
	"neo4j network_master_network_tx_writes=0",
	"neo4j transaction_rollbacks_write=0",
	"neo4j transaction_terminated_read=0",
	"neo4j,gc=g1_young_generation vm_gc_count=7",
	"neo4j bolt_messages_received=0",
	"neo4j bolt_messages_started=0",
	"neo4j transaction_committed=2",
	"neo4j,bufferpool=direct vm_memory_buffer_count=7",
	"neo4j page_cache_flushes=1",
	"neo4j transaction_started=2",
	"neo4j,bufferpool=mapped vm_memory_buffer_count=0",
	"neo4j,pool=g1_survivor_space vm_memory_pool=33554432",
	"neo4j bolt_connections_idle=0",
	"neo4j cypher_replan_wait_time=0",
	"neo4j transaction_last_closed_tx_id=1",
	"neo4j cypher_replan_events=0",
	"neo4j transaction_active=0",
	"neo4j transaction_active_read=0",
	"neo4j transaction_last_committed_tx_id=1",
	"neo4j bolt_connections_closed=0",
	"neo4j bolt_messages_done=0",
	"neo4j page_cache_usage_ratio=0.00029105392156862745",
	"neo4j transaction_terminated=0",
	"neo4j bolt_accumulated_queue_time=0",
	"neo4j check_point_events=0",
	"neo4j,gc=g1_young_generation vm_gc_time=187",
	"neo4j transaction_rollbacks_read=0",
	"neo4j,pool=g1_old_gen vm_memory_pool=14218288",
	"neo4j bolt_messages_failed=0",
	"neo4j transaction_peak_concurrent=1",
	"neo4j bolt_connections_running=0",
	"neo4j ids_in_use_property=0",
	"neo4j page_cache_eviction_exceptions=0",
	"neo4j page_cache_pins=110",
	"neo4j server_threads_jetty_all=12",
	"neo4j,gc=g1_old_generation vm_gc_count=0",
	"neo4j bolt_accumulated_processing_time=0",
	"neo4j bolt_connections_opened=0",
	"neo4j,pool=compressed_class_space vm_memory_pool=8315352",
	"neo4j,gc=g1_old_generation vm_gc_time=0",
	"neo4j,bufferpool=direct vm_memory_buffer_used=221185",
	"neo4j transaction_active_write=0",
	"neo4j transaction_rollbacks=0",
	"neo4j ids_in_use_relationship_type=0",
	"neo4j log_rotation_events=0",
	"neo4j,bufferpool=mapped vm_memory_buffer_used=0",
	"neo4j,pool=code_cache vm_memory_pool=12113088",
	"neo4j vm_thread_count=44",
	"neo4j page_cache_page_faults=19",
	"neo4j transaction_committed_write=0",
	"neo4j network_master_network_store_writes=0",
	"neo4j page_cache_hit_ratio=0.6724137931034483",
	"neo4j page_cache_unpins=57",
	"neo4j transaction_terminated_write=0",
	"neo4j,bufferpool=direct vm_memory_buffer_capacity=221184",
	"neo4j check_point_total_time=0",
	"neo4j ids_in_use_node=0",
	"neo4j server_threads_jetty_idle=2",
	"neo4j,pool=metaspace vm_memory_pool=57115680",
	"neo4j,pool=g1_eden_space vm_memory_pool=159383552",
	"neo4j bolt_sessions_started=0",
	"neo4j network_slave_network_tx_writes=0",
	"neo4j,bufferpool=mapped vm_memory_buffer_capacity=0",
	"neo4j ids_in_use_relationship=0",
	"neo4j page_cache_hits=39",
}

type taggerMock struct {
	hostTags, electionTags map[string]string
}

func (m *taggerMock) HostTags() map[string]string {
	return m.hostTags
}

func (m *taggerMock) ElectionTags() map[string]string {
	return m.electionTags
}
