// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package metrics generates apm metrics from profiling data.
package metrics

import (
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path"
	"strings"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/cliutils/pprofparser/domain/quantity"
	"github.com/grafana/jfr-parser/parser"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

const (
	metricsName = "profiling_metrics"
)

const (
	profJVMCPUCores               = "prof_jvm_cpu_cores"
	profJVMUptimeNS               = "prof_jvm_uptime_nanoseconds"
	profJVMAllocBytesTotal        = "prof_jvm_alloc_bytes_total"
	profJVMAllocBytesPerSec       = "prof_jvm_alloc_bytes_per_sec"
	profJVMAllocsPerSec           = "prof_jvm_allocs_per_sec"
	profJVMDirectAllocBytesPerSec = "prof_jvm_direct_alloc_bytes_per_sec"
	profJVMClassLoadsPerSec       = "prof_jvm_class_loads_per_sec"
	profJVMCompilationTime        = "prof_jvm_compilation_time"
	profJVMContextSwitchesPerSec  = "prof_jvm_context_switches_per_sec"
	profJVMThrowsPerSec           = "prof_jvm_throws_per_sec"
	profJVMThrowsTotal            = "prof_jvm_throws_total"
	profJVMFileIOMaxReadBytes     = "prof_jvm_file_io_max_read_bytes"
	profJVMFileIOReadBytes        = "prof_jvm_file_io_read_bytes"
	profJVMFileIOMaxReadTime      = "prof_jvm_file_io_max_read_time"
	profJVMFileIOReadTime         = "prof_jvm_file_io_read_time"
	profJVMFileIOMaxWriteBytes    = "prof_jvm_file_io_max_write_bytes"
	profJVMFileIOWriteBytes       = "prof_jvm_file_io_write_bytes"
	profJVMFileIOMaxWriteTime     = "prof_jvm_file_io_max_write_time"
	profJVMFileIOWriteTime        = "prof_jvm_file_io_write_time"
	profJVMFileIOTime             = "prof_jvm_file_io_time"
	profJVMAvgGcPauseTime         = "prof_jvm_avg_gc_pause_time"
	profJVMMaxGcPauseTime         = "prof_jvm_max_gc_pause_time"
	profJVMGcPauseTime            = "prof_jvm_gc_pause_time"
	profJVMGcPausesPerSec         = "prof_jvm_gc_pauses_per_sec"
	profJVMLifetimeHeapBytes      = "prof_jvm_lifetime_heap_bytes"
	profJVMLifetimeHeapObjects    = "prof_jvm_lifetime_heap_objects"
	profJVMLocksMaxWaitTime       = "prof_jvm_locks_max_wait_time"
	profJVMLocksPerSec            = "prof_jvm_locks_per_sec"
	profJVMThreadsCreatedPerSec   = "prof_jvm_threads_created_per_sec"
	profJVMThreadsDeadlocked      = "prof_jvm_threads_deadlocked"
	profJVMSocketIOMaxReadTime    = "prof_jvm_socket_io_max_read_time"
	profJVMSocketIOMaxReadBytes   = "prof_jvm_socket_io_max_read_bytes"
	profJVMSocketIOReadTime       = "prof_jvm_socket_io_read_time"
	profJVMSocketIOReadBytes      = "prof_jvm_socket_io_read_bytes"
	profJVMSocketIOMaxWriteTime   = "prof_jvm_socket_io_max_write_time"
	profJVMSocketIOMaxWriteBytes  = "prof_jvm_socket_io_max_write_bytes"
	profJVMSocketIOWriteTime      = "prof_jvm_socket_io_write_time"
	profJVMSocketIOWriteBytes     = "prof_jvm_socket_io_write_bytes"
)

const (
	profGoHeapGrowthBytesPerSec = "prof_go_heap_growth_bytes_per_sec"
	profGoGCsPerSec             = "prof_go_gcs_per_sec"
	profGoGCPauseTime           = "prof_go_gc_pause_time"
	profGoMaxGCPauseTime        = "prof_go_max_gc_pause_time"
	profGoNumGoroutine          = "prof_go_num_goroutine"
	profGoAllocBytesPerSec      = "prof_go_alloc_bytes_per_sec"
	profGoAllocsPerSec          = "prof_go_allocs_per_sec"
	profGoFreesPerSec           = "prof_go_frees_per_sec"
	profGoAllocBytesTotal       = "prof_go_alloc_bytes_total"     // profGoAllocsPerSec * total_seconds
	profGoCPUCoresGcOverhead    = "prof_go_cpu_cores_gc_overhead" // profGoGCPauseTime / total_seconds

	profGoCPUCores            = "prof_go_cpu_cores"
	profGoBlockedTime         = "prof_go_blocked_time"
	profGoMutexDelayTime      = "prof_go_mutex_delay_time"
	profGoLifetimeHeapBytes   = "prof_go_lifetime_heap_bytes"
	profGoLifetimeHeapObjects = "prof_go_lifetime_heap_objects"
)

const (
	profPythonCPUCores               = "prof_python_cpu_cores"
	profPythonAllocBytesPerSec       = "prof_python_alloc_bytes_per_sec"
	profPythonAllocsPerSec           = "prof_python_allocs_per_sec"
	profPythonAllocBytesTotal        = "prof_python_alloc_bytes_total"
	profPythonLockAcquisitionTime    = "prof_python_lock_acquisition_time"
	profPythonLockAcquisitionsPerSec = "prof_python_lock_acquisitions_per_sec"
	profPythonLockHoldTime           = "prof_python_lock_hold_time"
	profPythonExceptionsPerSec       = "prof_python_exceptions_per_sec"
	profPythonExceptionsTotal        = "prof_python_exceptions_total"
	profPythonLifetimeHeapBytes      = "prof_python_lifetime_heap_bytes"
	profPythonWallTime               = "prof_python_wall_time"
)

var goMetricsNameMapping = map[string]string{
	profGoCPUCores:              "go_cpu_cores",
	profGoCPUCoresGcOverhead:    "go_cpu_cores_gc_overhead",
	profGoAllocBytesPerSec:      "go_alloc_bytes_per_sec",
	profGoAllocBytesTotal:       "go_alloc_bytes_total",
	profGoFreesPerSec:           "go_frees_per_sec",
	profGoHeapGrowthBytesPerSec: "go_heap_growth_bytes_per_sec",
	profGoAllocsPerSec:          "go_allocs_per_sec",
	profGoBlockedTime:           "go_blocked_time",
	profGoMutexDelayTime:        "go_mutex_delay_time",
	profGoGCsPerSec:             "go_gcs_per_sec",
	profGoMaxGCPauseTime:        "go_max_gc_pause_time",
	profGoGCPauseTime:           "go_gc_pause_time",
	profGoNumGoroutine:          "go_num_goroutine",
	profGoLifetimeHeapBytes:     "go_lifetime_heap_bytes",
	profGoLifetimeHeapObjects:   "go_lifetime_heap_objects",
}

const (
	goCPUFile      = "cpu"
	goBlockFile    = "block"
	goHeapFile     = "heap"
	goMutexFile    = "mutex"
	goroutinesFile = "goroutines"
)

const (
	cpuTimeMetric          = "cpu-time"
	wallTimeMetric         = "wall-time"
	exceptionSamplesMetric = "exception-samples"
	lockAcquireMetric      = "lock-acquire"
	lockAcquireWaitMetric  = "lock-acquire-wait"
	allocSamplesMetric     = "alloc-samples"
	allocSpaceMetric       = "alloc-space"
	heapSpaceMetric        = "heap-space"
	lockReleaseHoldMetric  = "lock-release-hold"
)

var (
	log           = logger.DefaultSLogger("profilingMetrics")
	metricsFeeder = dkio.DefaultFeeder()
)

func InitLog() {
	log = logger.SLogger("profilingMetrics")
}

func exportMetrics(pts []*point.Point) error {
	if err := metricsFeeder.FeedV2(point.Metric, pts, dkio.WithInputName(metricsName)); err != nil {
		return fmt.Errorf("unable to feed profiling metrics: %w", err)
	}
	return nil
}

type metricKVs point.KVs

func newMetricKVs() *metricKVs {
	return toMetricKVs(nil)
}

func toMetricKVs(kvs point.KVs) *metricKVs {
	return (*metricKVs)(&kvs)
}

func (m *metricKVs) toPointKVs() point.KVs {
	if m == nil {
		return nil
	}
	return point.KVs(*m)
}

func (m *metricKVs) AddTag(k, v string) {
	if m == nil {
		return
	}
	*m = metricKVs(m.toPointKVs().AddTag(k, v))
}

func (m *metricKVs) EasyAdd(k string, v any) {
	m.AddV2(k, v, false)
}

func (m *metricKVs) AddV2(k string, v any, force bool, opts ...point.KVOption) {
	if m == nil {
		return
	}
	*m = metricKVs(m.toPointKVs().AddV2(k, v, force, opts...))
}

func ExportJVMMetrics(files map[string][]*multipart.FileHeader, metadata *ResolvedMetadata, customTags map[string]string) error {
	jfrFile := func() *multipart.FileHeader {
		for field, headers := range files {
			if field == EventFile || field == EventJSONFile {
				continue
			}

			switch field {
			case EventFile, EventJSONFile:
				continue
			case MainFile, MainJFRFile, AutoFile, AutoJFRFile:
				for _, header := range headers {
					if header.Size > 0 {
						return header
					}
				}
			}

			for _, header := range headers {
				if strings.HasSuffix(header.Filename, ".jfr") && header.Size > 0 {
					return header
				}
			}
		}
		return nil
	}()

	if jfrFile == nil {
		return fmt.Errorf("unable to find jfr file")
	}

	f, err := jfrFile.Open()
	if err != nil {
		return fmt.Errorf("unable to open jfr file: %w", err)
	}
	defer f.Close() // nolint:errcheck

	jfrStart, err := ResolveStartTime(metadata.FormValue)
	if err != nil {
		return fmt.Errorf("unable to resolve jfr start time: %w", err)
	}
	jfrEnd, err := ResolveEndTime(metadata.FormValue)
	if err != nil {
		return fmt.Errorf("unable to resolve jfr end time: %w", err)
	}

	jfrDurationNS, jfrDurationSeconds := jfrEnd.Sub(jfrStart).Nanoseconds(), jfrEnd.Sub(jfrStart).Seconds()

	chunks, err := parser.Parse(f)
	if err != nil {
		return fmt.Errorf("unable to parse jfr: %w", err)
	}

	jc := jfrChunks(chunks)

	commonTags := map[string]string{
		"language": Java.String(),
		"host":     metadata.GetTag("host"),
		"service":  metadata.GetTag("service"),
		"env":      metadata.GetTag("env"),
		"version":  metadata.GetTag("version"),
	}

	for k, v := range customTags {
		commonTags[k] = v
	}

	kVs := toMetricKVs(point.NewTags(commonTags))

	if jvmStart, err := jc.jvmStartTime(); err == nil {
		kVs.EasyAdd(profJVMUptimeNS, jfrEnd.Sub(jvmStart).Nanoseconds())
	}

	costCPUCores := float64(jc.cpuTimeDurationNS()) / float64(jfrDurationNS)
	kVs.EasyAdd(profJVMCPUCores, costCPUCores)

	allocBytes, allocCount := jc.allocations()
	kVs.EasyAdd(profJVMAllocBytesTotal, allocBytes)
	kVs.EasyAdd(profJVMAllocBytesPerSec, allocBytes/jfrDurationSeconds)
	kVs.EasyAdd(profJVMAllocsPerSec, allocCount/jfrDurationSeconds)

	directAllocBytes := jc.directAllocationBytes()
	kVs.EasyAdd(profJVMDirectAllocBytesPerSec, float64(directAllocBytes)/jfrDurationSeconds)

	classCount := jc.classLoaderCount()
	kVs.EasyAdd(profJVMClassLoadsPerSec, float64(classCount)/jfrDurationSeconds)

	kVs.EasyAdd(profJVMCompilationTime, jc.compilationDuration())

	kVs.EasyAdd(profJVMContextSwitchesPerSec, jc.threadContextSwitchRate())

	totalExceptions := jc.exceptionCount()
	kVs.EasyAdd(profJVMThrowsTotal, totalExceptions)
	kVs.EasyAdd(profJVMThrowsPerSec, float64(totalExceptions)/jfrDurationSeconds)

	readMaxDurationNS, readMaxBytesRead, totalReadDurationNS, totalBytesRead := jc.fileRead()
	kVs.EasyAdd(profJVMFileIOMaxReadTime, readMaxDurationNS)
	kVs.EasyAdd(profJVMFileIOMaxReadBytes, readMaxBytesRead)
	kVs.EasyAdd(profJVMFileIOReadTime, totalReadDurationNS)
	kVs.EasyAdd(profJVMFileIOReadBytes, totalBytesRead)

	maxWriteDurationNS, maxBytesWritten, totalWriteDurationNS, totalBytesWritten := jc.fileWrite()

	kVs.EasyAdd(profJVMFileIOMaxWriteTime, maxWriteDurationNS)
	kVs.EasyAdd(profJVMFileIOMaxWriteBytes, maxBytesWritten)
	kVs.EasyAdd(profJVMFileIOWriteTime, totalWriteDurationNS)
	kVs.EasyAdd(profJVMFileIOWriteBytes, totalBytesWritten)
	kVs.EasyAdd(profJVMFileIOTime, totalReadDurationNS+totalWriteDurationNS)

	durationNS, count := jc.gcDuration()
	if count == 0 {
		kVs.EasyAdd(profJVMAvgGcPauseTime, 0)
	} else {
		kVs.EasyAdd(profJVMAvgGcPauseTime, float64(durationNS)/float64(count))
	}

	maxPauseNanos, totalPauseNanos, pauseCount := jc.gcPauseDuration()
	kVs.EasyAdd(profJVMMaxGcPauseTime, maxPauseNanos)
	kVs.EasyAdd(profJVMGcPauseTime, totalPauseNanos)
	kVs.EasyAdd(profJVMGcPausesPerSec, float64(pauseCount)/jfrDurationSeconds)

	kVs.EasyAdd(profJVMLifetimeHeapObjects, jc.liveHeapSamples())
	kVs.EasyAdd(profJVMLifetimeHeapBytes, jc.jvmHeapUsage())

	maxLockDurationNS, _, lockCount := jc.monitorEnter()

	kVs.EasyAdd(profJVMLocksMaxWaitTime, maxLockDurationNS)
	kVs.EasyAdd(profJVMLocksPerSec, float64(lockCount)/jfrDurationSeconds)

	kVs.EasyAdd(profJVMThreadsCreatedPerSec, float64(jc.threadStart())/jfrDurationSeconds)
	kVs.EasyAdd(profJVMThreadsDeadlocked, jc.deadlockedThread())

	maxReadTimeNS, maxBytesRead, totalReadTimeNS, totalBytesRead := jc.socketIORead()

	kVs.EasyAdd(profJVMSocketIOMaxReadTime, maxReadTimeNS)
	kVs.EasyAdd(profJVMSocketIOMaxReadBytes, maxBytesRead)
	kVs.EasyAdd(profJVMSocketIOReadTime, totalReadTimeNS)
	kVs.EasyAdd(profJVMSocketIOReadBytes, totalBytesRead)

	maxWriteTimeNS, maxBytesWritten, totalWriteTimeNS, totalBytesWritten := jc.socketIOWrite()

	kVs.EasyAdd(profJVMSocketIOMaxWriteTime, maxWriteTimeNS)
	kVs.EasyAdd(profJVMSocketIOMaxWriteBytes, maxBytesWritten)
	kVs.EasyAdd(profJVMSocketIOWriteTime, totalWriteTimeNS)
	kVs.EasyAdd(profJVMSocketIOWriteBytes, totalBytesWritten)

	pt := point.NewPointV2(metricsName, kVs.toPointKVs(), point.WithPrecision(point.PrecNS), point.WithTime(jfrEnd))
	if err = exportMetrics([]*point.Point{pt}); err != nil {
		return fmt.Errorf("unable to export profiling metrics: %w", err)
	}
	return nil
}

func pickProfileFile(files map[string][]*multipart.FileHeader) *multipart.FileHeader {
	for fieldName, headers := range files {
		if len(headers) > 0 {
			if fieldName == AutoFile || fieldName == MainFile || fieldName == ProfFile {
				return headers[0]
			}
			if path.Ext(fieldName) == PprofExt {
				return headers[0]
			}

			for _, header := range headers {
				if header.Filename == AutoFile || header.Filename == MainFile {
					return header
				}
				if path.Ext(header.Filename) == PprofExt {
					return header
				}
			}
		}
	}
	return nil
}

func ExportPythonMetrics(files map[string][]*multipart.FileHeader, metadata *ResolvedMetadata, customTags map[string]string) error {
	commonTags := map[string]string{
		"language": Python.String(),
		"host":     metadata.GetTag("host"),
		"service":  metadata.GetTag("service"),
		"env":      metadata.GetTag("env"),
		"version":  metadata.GetTag("version"),
	}

	for k, v := range customTags {
		commonTags[k] = v
	}

	pprofStart, err := ResolveStartTime(metadata.FormValue)
	if err != nil {
		return fmt.Errorf("unable to resolve python profiling start time: %w", err)
	}
	pprofEnd, err := ResolveEndTime(metadata.FormValue)
	if err != nil {
		return fmt.Errorf("unable to resolve python profiling end time: %w", err)
	}

	profFile := pickProfileFile(files)
	if profFile == nil {
		return fmt.Errorf("unable to find any pprof file")
	}

	pprofDurationNS, pprofDurationSeconds := pprofEnd.Sub(pprofStart).Nanoseconds(), pprofEnd.Sub(pprofStart).Seconds()

	kVs := toMetricKVs(point.NewTags(commonTags))

	summaries, err := pprofSummaryHeader(profFile)
	if err != nil {
		return fmt.Errorf("unable to resolve summaries from pprof file: %w", err)
	}

	if cpuTime := summaries[cpuTimeMetric]; cpuTime != nil {
		cpuNanos := cpuTime.Value
		cpuUnit, err := quantity.ParseUnit(quantity.Duration, cpuTime.Unit)
		if err != nil {
			log.Warnf("unable to resolve cpu duraiton unit: %v", err)
		} else {
			if q := cpuUnit.Quantity(cpuTime.Value); q.Unit != quantity.NanoSecond {
				cpuNanos, err = q.IntValueIn(quantity.NanoSecond)
				if err != nil {
					log.Warnf("unable to change unit to nanosecond: %v", err)
				}
			}
		}
		kVs.EasyAdd(profPythonCPUCores, float64(cpuNanos)/float64(pprofDurationNS))
	}

	if allocSpace := summaries[allocSpaceMetric]; allocSpace != nil {
		allocBytes := allocSpace.Value
		unit, err := quantity.ParseUnit(quantity.Memory, allocSpace.Unit)
		if err != nil {
			log.Warnf("unable to resolve alloc space unit: %v", err)
		} else {
			if q := unit.Quantity(allocSpace.Value); q.Unit != quantity.Byte {
				allocBytes, err = q.IntValueIn(quantity.Byte)
				if err != nil {
					log.Warnf("unable to change unit to byte: %v", err)
				}
			}
		}
		kVs.EasyAdd(profPythonAllocBytesTotal, allocBytes)
		kVs.EasyAdd(profPythonAllocBytesPerSec, float64(allocBytes)/pprofDurationSeconds)
	}

	if allocSample := summaries[allocSamplesMetric]; allocSample != nil {
		kVs.EasyAdd(profPythonAllocsPerSec, float64(allocSample.Value)/pprofDurationSeconds)
	}

	if lockCount := summaries[lockAcquireMetric]; lockCount != nil {
		kVs.EasyAdd(profPythonLockAcquisitionsPerSec, float64(lockCount.Value)/pprofDurationSeconds)
	}

	if lockWait := summaries[lockAcquireWaitMetric]; lockWait != nil {
		waitDuration := lockWait.Value
		unit, err := quantity.ParseUnit(quantity.Duration, lockWait.Unit)
		if err != nil {
			log.Warnf("unable to resolve lock wait duraiton unit: %v", err)
		} else {
			if q := unit.Quantity(lockWait.Value); q.Unit != quantity.NanoSecond {
				waitDuration, err = q.IntValueIn(quantity.NanoSecond)
				if err != nil {
					log.Warnf("unable to change unit to nanosecond: %v", err)
				}
			}
		}
		kVs.EasyAdd(profPythonLockAcquisitionTime, waitDuration)
	}

	if lockRelease := summaries[lockReleaseHoldMetric]; lockRelease != nil {
		waitDuration := lockRelease.Value
		unit, err := quantity.ParseUnit(quantity.Duration, lockRelease.Unit)
		if err != nil {
			log.Warnf("unable to resolve lock release duraiton unit: %v", err)
		} else {
			if q := unit.Quantity(lockRelease.Value); q.Unit != quantity.NanoSecond {
				waitDuration, err = q.IntValueIn(quantity.NanoSecond)
				if err != nil {
					log.Warnf("unable to change unit to nanosecond: %v", err)
				}
			}
		}
		kVs.EasyAdd(profPythonLockHoldTime, waitDuration)
	}

	if exception := summaries[exceptionSamplesMetric]; exception != nil {
		kVs.EasyAdd(profPythonExceptionsTotal, exception.Value)
		kVs.EasyAdd(profPythonExceptionsPerSec, float64(exception.Value)/pprofDurationSeconds)
	}

	if wallTime := summaries[wallTimeMetric]; wallTime != nil {
		wallDuration := wallTime.Value

		unit, err := quantity.ParseUnit(quantity.Duration, wallTime.Unit)
		if err != nil {
			log.Warnf("unable to resolve wall duraiton unit: %v", err)
		} else {
			if q := unit.Quantity(wallTime.Value); q.Unit != quantity.NanoSecond {
				wallDuration, err = q.IntValueIn(quantity.NanoSecond)
				if err != nil {
					log.Warnf("unable to change unit to nanosecond: %v", err)
				}
			}
		}
		kVs.EasyAdd(profPythonWallTime, wallDuration)
	}

	if heapSpace := summaries[heapSpaceMetric]; heapSpace != nil {
		heapBytes := heapSpace.Value

		unit, err := quantity.ParseUnit(quantity.Memory, heapSpace.Unit)
		if err != nil {
			log.Warnf("unable to resolve alloc space unit: %v", err)
		} else {
			if q := unit.Quantity(heapSpace.Value); q.Unit != quantity.Byte {
				heapBytes, err = q.IntValueIn(quantity.Byte)
				if err != nil {
					log.Warnf("unable to change unit to byte: %v", err)
				}
			}
		}
		kVs.EasyAdd(profPythonLifetimeHeapBytes, heapBytes)
	}

	pt := point.NewPointV2(metricsName, kVs.toPointKVs(), point.WithPrecision(point.PrecNS), point.WithTime(pprofEnd))
	if err = exportMetrics([]*point.Point{pt}); err != nil {
		return fmt.Errorf("unable to export profiling metrics: %w", err)
	}
	return nil
}

func ExportGoMetrics(files map[string][]*multipart.FileHeader, metadata *ResolvedMetadata, customTags map[string]string) error {
	commonTags := map[string]string{
		"language": Golang.String(),
		"host":     metadata.GetTag("host"),
		"service":  metadata.GetTag("service"),
		"env":      metadata.GetTag("env"),
		"version":  metadata.GetTag("version"),
	}

	for k, v := range customTags {
		commonTags[k] = v
	}

	pprofStart, err := ResolveStartTime(metadata.FormValue)
	if err != nil {
		return fmt.Errorf("unable to resolve go profiling start time: %w", err)
	}
	pprofEnd, err := ResolveEndTime(metadata.FormValue)
	if err != nil {
		return fmt.Errorf("unable to resolve go profiling end time: %w", err)
	}

	pprofDurationNS, pprofDurationSeconds := pprofEnd.Sub(pprofStart).Nanoseconds(), pprofEnd.Sub(pprofStart).Seconds()

	kVs := toMetricKVs(point.NewTags(commonTags))

	metricsFile, ok := files[MetricFile]
	if !ok {
		metricsFile = files[MetricJSONFile]
	}

	hasExported := make(map[string]bool)

	if len(metricsFile) > 0 {
		mf, err := metricsFile[0].Open()
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("unable to open metrics.json file: %w", err)
			}
		} else {
			defer mf.Close() // nolint:errcheck

			jsonMetering, err := parseMetricsJSONFile(mf)
			if err != nil {
				return fmt.Errorf("unable to resolve metrics.json: %w", err)
			}

			for metricName, number := range jsonMetering {
				hasExported[metricName] = true
				kVs.EasyAdd(metricName, resolveJSONNumber(number))
			}
			if allocBytesMetric, ok := jsonMetering[profGoAllocBytesPerSec]; ok && !hasExported[profGoAllocBytesTotal] {
				allocPerSec, err := allocBytesMetric.Float64()
				if err == nil && allocPerSec > 0 {
					hasExported[profGoAllocBytesTotal] = true
					kVs.EasyAdd(profGoAllocBytesTotal, allocPerSec*pprofDurationSeconds)
				}
			}

			if gcPauseMetric, ok := jsonMetering[profGoGCPauseTime]; ok && !hasExported[profGoCPUCoresGcOverhead] {
				gcPauseDuration, err := gcPauseMetric.Float64()
				if err == nil && gcPauseDuration > 0 {
					hasExported[profGoCPUCoresGcOverhead] = true
					kVs.EasyAdd(profGoCPUCoresGcOverhead, gcPauseDuration/pprofDurationSeconds)
				}
			}
		}
	}

	pprofFiles := make(map[string]*multipart.FileHeader, 5)

	for field, headers := range files {
		switch {
		case strings.Contains(field, goCPUFile) && len(headers) > 0:
			pprofFiles[goCPUFile] = headers[0]
		case strings.Contains(field, goBlockFile) && len(headers) > 0:
			pprofFiles[goBlockFile] = headers[0]
		case strings.Contains(field, goHeapFile) && len(headers) > 0:
			pprofFiles[goHeapFile] = headers[0]
		case strings.Contains(field, goMutexFile) && len(headers) > 0:
			pprofFiles[goMutexFile] = headers[0]
		case strings.Contains(field, goroutinesFile) && len(headers) > 0:
			pprofFiles[goroutinesFile] = headers[0]
		}
	}

	if !hasExported[profGoCPUCores] {
		if cpuFile, ok := pprofFiles[goCPUFile]; ok {
			cpuNanos, err := pprofCPUDuration(cpuFile)
			if err != nil {
				log.Warnf("unable to resolve pprof cpu duration: %v", err)
			} else {
				hasExported[profGoCPUCores] = true
				kVs.EasyAdd(profGoCPUCores, float64(cpuNanos)/float64(pprofDurationNS))
			}
		}
	}

	if !hasExported[profGoLifetimeHeapObjects] || !hasExported[profGoLifetimeHeapBytes] {
		if heapFile, ok := pprofFiles[goHeapFile]; ok {
			objects, size, err := liveHeapSummary(heapFile)
			if err != nil {
				log.Warnf("unable to resolve go pprof live heap metrics: %v", err)
			} else {
				if !hasExported[profGoLifetimeHeapObjects] {
					hasExported[profGoLifetimeHeapObjects] = true
					kVs.EasyAdd(profGoLifetimeHeapObjects, objects)
				}
				if !hasExported[profGoLifetimeHeapBytes] {
					hasExported[profGoLifetimeHeapBytes] = true
					kVs.EasyAdd(profGoLifetimeHeapBytes, size)
				}
			}
		}
	}

	if !hasExported[profGoBlockedTime] {
		if blockFile, ok := pprofFiles[goBlockFile]; ok {
			delayNS, err := delayDurationNS(blockFile)
			if err != nil {
				log.Warnf("unable to resolve go pprof block delay duration: %v", err)
			} else {
				hasExported[profGoBlockedTime] = true
				kVs.EasyAdd(profGoBlockedTime, delayNS)
			}
		}
	}

	if !hasExported[profGoMutexDelayTime] {
		if mutexFile, ok := pprofFiles[goMutexFile]; ok {
			delayNS, err := delayDurationNS(mutexFile)
			if err != nil {
				log.Warnf("unable to resolve go pprof mutex delay duration: %v", err)
			} else {
				hasExported[profGoMutexDelayTime] = true
				kVs.EasyAdd(profGoMutexDelayTime, delayNS)
			}
		}
	}

	if !hasExported[profGoNumGoroutine] {
		if goroutineFile, ok := pprofFiles[goroutinesFile]; ok {
			gCount, err := goroutinesCount(goroutineFile)
			if err != nil {
				log.Warnf("unable to resolve go pprof goroutines count metric: %w", err)
			} else {
				hasExported[profGoNumGoroutine] = true
				kVs.EasyAdd(profGoNumGoroutine, gCount)
			}
		}
	}

	pt := point.NewPointV2(metricsName, kVs.toPointKVs(), point.WithPrecision(point.PrecNS), point.WithTime(pprofEnd))
	if err = exportMetrics([]*point.Point{pt}); err != nil {
		return fmt.Errorf("unable to export profiling metrics: %w", err)
	}
	return nil
}
