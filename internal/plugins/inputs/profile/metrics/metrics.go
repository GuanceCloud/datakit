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
)

const (
	// for java.
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

	// for python.
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

	// for golang.
	profGoGCsPerSec             = "prof_go_gcs_per_sec"
	profGoGCPauseTime           = "prof_go_gc_pause_time"
	profGoNumGoroutine          = "prof_go_num_goroutine"
	profGoHeapGrowthBytesPerSec = "prof_go_heap_growth_bytes_per_sec"
	profGoAllocsPerSec          = "prof_go_allocs_per_sec"
	profGoMaxGCPauseTime        = "prof_go_max_gc_pause_time"
	profGoAllocBytesPerSec      = "prof_go_alloc_bytes_per_sec"
	profGoFreesPerSec           = "prof_go_frees_per_sec"
	profGoAllocBytesTotal       = "prof_go_alloc_bytes_total"     // profGoAllocsPerSec * total_seconds
	profGoCPUCoresGcOverhead    = "prof_go_cpu_cores_gc_overhead" // profGoGCPauseTime / total_seconds
	profGoCPUCores              = "prof_go_cpu_cores"
	profGoBlockedTime           = "prof_go_blocked_time"
	profGoMutexDelayTime        = "prof_go_mutex_delay_time"
	profGoLifetimeHeapBytes     = "prof_go_lifetime_heap_bytes"
	profGoLifetimeHeapObjects   = "prof_go_lifetime_heap_objects"

	// golang profile file namess.
	goCPUFile      = "cpu"
	goBlockFile    = "block"
	goHeapFile     = "heap"
	goMutexFile    = "mutex"
	goroutinesFile = "goroutines"

	// summary keys.
	cpuTimeMetric          = "cpu-time"
	wallTimeMetric         = "wall-time"
	exceptionSamplesMetric = "exception-samples"
	lockAcquireMetric      = "lock-acquire"
	lockAcquireWaitMetric  = "lock-acquire-wait"
	allocSamplesMetric     = "alloc-samples"
	allocSpaceMetric       = "alloc-space"
	heapSpaceMetric        = "heap-space"
	lockReleaseHoldMetric  = "lock-release-hold"

	metricsName = "profiling_metrics"
)

var (
	l = logger.DefaultSLogger("profilingMetrics")

	goMetricsNameMapping = map[string]string{
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
)

func InitLog() {
	l = logger.SLogger("profilingMetrics")
}

func ExtractJVMMetrics(files map[string][]*multipart.FileHeader,
	metadata map[string]string,
	customTags map[string]string,
) ([]*point.Point, error) {
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
		return nil, fmt.Errorf("unable to find jfr file")
	}

	f, err := jfrFile.Open()
	if err != nil {
		return nil, fmt.Errorf("unable to open jfr file: %w", err)
	}
	defer f.Close() // nolint:errcheck

	jfrStart, err := ResolveStartTime(metadata)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve jfr start time: %w", err)
	}
	jfrEnd, err := ResolveEndTime(metadata)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve jfr end time: %w", err)
	}

	jfrDurationNS, jfrDurationSeconds := jfrEnd.Sub(jfrStart).Nanoseconds(), jfrEnd.Sub(jfrStart).Seconds()

	chunks, err := parser.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("unable to parse jfr: %w", err)
	}

	jc := jfrChunks(chunks)

	var kvs point.KVs

	kvs = kvs.AddTag("language", Java.String()).
		AddTag("host", metadata["host"]).
		AddTag("service", metadata["service"]).
		AddTag("env", metadata["env"]).
		AddTag("version", metadata["version"])

	for k, v := range customTags {
		kvs = kvs.AddTag(k, v)
	}

	if jvmStart, err := jc.jvmStartTime(); err == nil {
		kvs = kvs.Add(profJVMUptimeNS, jfrEnd.Sub(jvmStart).Nanoseconds())
	}

	if jfrDurationNS > 0 {
		costCPUCores := float64(jc.cpuTimeDurationNS()) / float64(jfrDurationNS)
		kvs = kvs.Add(profJVMCPUCores, costCPUCores)
	}

	allocBytes, allocCount := jc.allocations()
	kvs = kvs.Add(profJVMAllocBytesTotal, allocBytes).
		Add(profJVMAllocBytesPerSec, allocBytes/jfrDurationSeconds).
		Add(profJVMAllocsPerSec, allocCount/jfrDurationSeconds)

	directAllocBytes := jc.directAllocationBytes()
	kvs = kvs.Add(profJVMDirectAllocBytesPerSec, float64(directAllocBytes)/jfrDurationSeconds)

	classCount := jc.classLoaderCount()
	kvs = kvs.Add(profJVMClassLoadsPerSec, float64(classCount)/jfrDurationSeconds)

	kvs = kvs.Add(profJVMCompilationTime, jc.compilationDuration()).
		Add(profJVMContextSwitchesPerSec, jc.threadContextSwitchRate())

	totalExceptions := jc.exceptionCount()
	kvs = kvs.Add(profJVMThrowsTotal, totalExceptions).
		Add(profJVMThrowsPerSec, float64(totalExceptions)/jfrDurationSeconds)

	readMaxDurationNS, readMaxBytesRead, totalReadDurationNS, totalBytesRead := jc.fileRead()
	kvs = kvs.Add(profJVMFileIOMaxReadTime, readMaxDurationNS).
		Add(profJVMFileIOMaxReadBytes, readMaxBytesRead).
		Add(profJVMFileIOReadTime, totalReadDurationNS).
		Add(profJVMFileIOReadBytes, totalBytesRead)

	maxWriteDurationNS, maxBytesWritten, totalWriteDurationNS, totalBytesWritten := jc.fileWrite()

	kvs = kvs.Add(profJVMFileIOMaxWriteTime, maxWriteDurationNS).
		Add(profJVMFileIOMaxWriteBytes, maxBytesWritten).
		Add(profJVMFileIOWriteTime, totalWriteDurationNS).
		Add(profJVMFileIOWriteBytes, totalBytesWritten).
		Add(profJVMFileIOTime, totalReadDurationNS+totalWriteDurationNS)

	durationNS, count := jc.gcDuration()
	if count == 0 {
		kvs = kvs.Add(profJVMAvgGcPauseTime, 0)
	} else {
		kvs = kvs.Add(profJVMAvgGcPauseTime, float64(durationNS)/float64(count))
	}

	maxPauseNanos, totalPauseNanos, pauseCount := jc.gcPauseDuration()
	kvs = kvs.Add(profJVMMaxGcPauseTime, maxPauseNanos).
		Add(profJVMGcPauseTime, totalPauseNanos).
		Add(profJVMGcPausesPerSec, float64(pauseCount)/jfrDurationSeconds)

	kvs = kvs.Add(profJVMLifetimeHeapObjects, jc.liveHeapSamples()).
		Add(profJVMLifetimeHeapBytes, jc.jvmHeapUsage())

	maxLockDurationNS, _, lockCount := jc.monitorEnter()

	kvs = kvs.Add(profJVMLocksMaxWaitTime, maxLockDurationNS).
		Add(profJVMLocksPerSec, float64(lockCount)/jfrDurationSeconds)

	kvs = kvs.Add(profJVMThreadsCreatedPerSec, float64(jc.threadStart())/jfrDurationSeconds).
		Add(profJVMThreadsDeadlocked, jc.deadlockedThread())

	maxReadTimeNS, maxBytesRead, totalReadTimeNS, totalBytesRead := jc.socketIORead()

	kvs = kvs.Add(profJVMSocketIOMaxReadTime, maxReadTimeNS).
		Add(profJVMSocketIOMaxReadBytes, maxBytesRead).
		Add(profJVMSocketIOReadTime, totalReadTimeNS).
		Add(profJVMSocketIOReadBytes, totalBytesRead)

	maxWriteTimeNS, maxBytesWritten, totalWriteTimeNS, totalBytesWritten := jc.socketIOWrite()

	kvs = kvs.Add(profJVMSocketIOMaxWriteTime, maxWriteTimeNS).
		Add(profJVMSocketIOMaxWriteBytes, maxBytesWritten).
		Add(profJVMSocketIOWriteTime, totalWriteTimeNS).
		Add(profJVMSocketIOWriteBytes, totalBytesWritten)

	pt := point.NewPoint(metricsName, kvs, point.WithPrecision(point.PrecNS), point.WithTime(jfrEnd))

	return []*point.Point{pt}, nil
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

func ExtractPythonMetrics(files map[string][]*multipart.FileHeader,
	metadata map[string]string,
	customTags map[string]string,
) ([]*point.Point, error) {
	var kvs point.KVs

	kvs = kvs.AddTag("language", Python.String()).
		AddTag("host", metadata["host"]).
		AddTag("service", metadata["service"]).
		AddTag("env", metadata["env"]).
		AddTag("version", metadata["version"])

	for k, v := range customTags {
		kvs = kvs.AddTag(k, v)
	}

	pprofStart, err := ResolveStartTime(metadata)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve python profiling start time: %w", err)
	}
	pprofEnd, err := ResolveEndTime(metadata)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve python profiling end time: %w", err)
	}

	profFile := pickProfileFile(files)
	if profFile == nil {
		return nil, fmt.Errorf("unable to find any pprof file")
	}

	pprofDurationNS, pprofDurationSeconds := pprofEnd.Sub(pprofStart).Nanoseconds(), pprofEnd.Sub(pprofStart).Seconds()
	summaries, err := pprofSummaryHeader(profFile)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve summaries from pprof file: %w", err)
	}

	if cpuTime := summaries[cpuTimeMetric]; cpuTime != nil {
		cpuNanos := cpuTime.Value
		cpuUnit, err := quantity.ParseUnit(quantity.Duration, cpuTime.Unit)
		if err != nil {
			l.Warnf("unable to resolve cpu duraiton unit: %v", err)
		} else {
			if q := cpuUnit.Quantity(cpuTime.Value); q.Unit != quantity.NanoSecond {
				cpuNanos, err = q.IntValueIn(quantity.NanoSecond)
				if err != nil {
					l.Warnf("unable to change unit to nanosecond: %v", err)
				}
			}
		}
		kvs = kvs.Add(profPythonCPUCores, float64(cpuNanos)/float64(pprofDurationNS))
	}

	if allocSpace := summaries[allocSpaceMetric]; allocSpace != nil {
		allocBytes := allocSpace.Value
		unit, err := quantity.ParseUnit(quantity.Memory, allocSpace.Unit)
		if err != nil {
			l.Warnf("unable to resolve alloc space unit: %v", err)
		} else {
			if q := unit.Quantity(allocSpace.Value); q.Unit != quantity.Byte {
				allocBytes, err = q.IntValueIn(quantity.Byte)
				if err != nil {
					l.Warnf("unable to change unit to byte: %v", err)
				}
			}
		}
		kvs = kvs.Add(profPythonAllocBytesTotal, allocBytes).
			Add(profPythonAllocBytesPerSec, float64(allocBytes)/pprofDurationSeconds)
	}

	if allocSample := summaries[allocSamplesMetric]; allocSample != nil {
		kvs = kvs.Add(profPythonAllocsPerSec, float64(allocSample.Value)/pprofDurationSeconds)
	}

	if lockCount := summaries[lockAcquireMetric]; lockCount != nil {
		kvs = kvs.Add(profPythonLockAcquisitionsPerSec, float64(lockCount.Value)/pprofDurationSeconds)
	}

	if lockWait := summaries[lockAcquireWaitMetric]; lockWait != nil {
		waitDuration := lockWait.Value
		unit, err := quantity.ParseUnit(quantity.Duration, lockWait.Unit)
		if err != nil {
			l.Warnf("unable to resolve lock wait duraiton unit: %v", err)
		} else {
			if q := unit.Quantity(lockWait.Value); q.Unit != quantity.NanoSecond {
				waitDuration, err = q.IntValueIn(quantity.NanoSecond)
				if err != nil {
					l.Warnf("unable to change unit to nanosecond: %v", err)
				}
			}
		}
		kvs = kvs.Add(profPythonLockAcquisitionTime, waitDuration)
	}

	if lockRelease := summaries[lockReleaseHoldMetric]; lockRelease != nil {
		waitDuration := lockRelease.Value
		unit, err := quantity.ParseUnit(quantity.Duration, lockRelease.Unit)
		if err != nil {
			l.Warnf("unable to resolve lock release duraiton unit: %v", err)
		} else {
			if q := unit.Quantity(lockRelease.Value); q.Unit != quantity.NanoSecond {
				waitDuration, err = q.IntValueIn(quantity.NanoSecond)
				if err != nil {
					l.Warnf("unable to change unit to nanosecond: %v", err)
				}
			}
		}
		kvs = kvs.Add(profPythonLockHoldTime, waitDuration)
	}

	if exception := summaries[exceptionSamplesMetric]; exception != nil {
		kvs = kvs.Add(profPythonExceptionsTotal, exception.Value).
			Add(profPythonExceptionsPerSec, float64(exception.Value)/pprofDurationSeconds)
	}

	if wallTime := summaries[wallTimeMetric]; wallTime != nil {
		wallDuration := wallTime.Value

		unit, err := quantity.ParseUnit(quantity.Duration, wallTime.Unit)
		if err != nil {
			l.Warnf("unable to resolve wall duraiton unit: %v", err)
		} else {
			if q := unit.Quantity(wallTime.Value); q.Unit != quantity.NanoSecond {
				wallDuration, err = q.IntValueIn(quantity.NanoSecond)
				if err != nil {
					l.Warnf("unable to change unit to nanosecond: %v", err)
				}
			}
		}
		kvs = kvs.Add(profPythonWallTime, wallDuration)
	}

	if heapSpace := summaries[heapSpaceMetric]; heapSpace != nil {
		heapBytes := heapSpace.Value

		unit, err := quantity.ParseUnit(quantity.Memory, heapSpace.Unit)
		if err != nil {
			l.Warnf("unable to resolve alloc space unit: %v", err)
		} else {
			if q := unit.Quantity(heapSpace.Value); q.Unit != quantity.Byte {
				heapBytes, err = q.IntValueIn(quantity.Byte)
				if err != nil {
					l.Warnf("unable to change unit to byte: %v", err)
				}
			}
		}
		kvs = kvs.Add(profPythonLifetimeHeapBytes, heapBytes)
	}

	pt := point.NewPoint(metricsName, kvs, point.WithPrecision(point.PrecNS), point.WithTime(pprofEnd))

	return []*point.Point{pt}, nil
}

func ExtractGoMetrics(files map[string][]*multipart.FileHeader,
	metadata map[string]string,
	customTags map[string]string,
) ([]*point.Point, error) {
	var kvs point.KVs

	kvs = kvs.AddTag("language", Golang.String()).
		AddTag("host", metadata["host"]).
		AddTag("service", metadata["service"]).
		AddTag("env", metadata["env"]).
		AddTag("version", metadata["version"])

	for k, v := range customTags {
		kvs = kvs.AddTag(k, v)
	}

	pprofStart, err := ResolveStartTime(metadata)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve go profiling start time: %w", err)
	}
	pprofEnd, err := ResolveEndTime(metadata)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve go profiling end time: %w", err)
	}

	pprofDurationNS, pprofDurationSeconds := pprofEnd.Sub(pprofStart).Nanoseconds(), pprofEnd.Sub(pprofStart).Seconds()

	metricsFile, ok := files[MetricFile]
	if !ok {
		metricsFile = files[MetricJSONFile]
	}

	if len(metricsFile) > 0 {
		mf, err := metricsFile[0].Open()
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("unable to open metrics.json file: %w", err)
			}
		} else {
			defer mf.Close() // nolint:errcheck

			jsonMetering, err := parseMetricsJSONFile(mf)
			if err != nil {
				return nil, fmt.Errorf("unable to resolve metrics.json: %w", err)
			}

			l.Debugf("jsonMetering: %+#v", jsonMetering)

			for metricName, number := range jsonMetering {
				kvs = kvs.Add(metricName, resolveJSONNumber(number))
			}

			if allocBytesMetric, ok := jsonMetering[profGoAllocBytesPerSec]; ok && kvs.Get(profGoAllocBytesTotal) == nil {
				allocPerSec, err := allocBytesMetric.Float64()
				if err == nil && allocPerSec > 0 {
					kvs = kvs.Add(profGoAllocBytesTotal, allocPerSec*pprofDurationSeconds)
				}
			}

			if gcPauseMetric, ok := jsonMetering[profGoGCPauseTime]; ok && kvs.Get(profGoCPUCoresGcOverhead) == nil {
				gcPauseDuration, err := gcPauseMetric.Float64()
				if err == nil && gcPauseDuration > 0 {
					kvs = kvs.Add(profGoCPUCoresGcOverhead, gcPauseDuration/pprofDurationSeconds)
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

	if kvs.Get(profGoCPUCores) == nil {
		if cpuFile, ok := pprofFiles[goCPUFile]; ok {
			cpuNanos, err := pprofCPUDuration(cpuFile)
			if err != nil {
				l.Warnf("unable to resolve pprof cpu duration: %v", err)
			} else {
				kvs = kvs.Add(profGoCPUCores, float64(cpuNanos)/float64(pprofDurationNS))
			}
		}
	}

	if kvs.Get(profGoLifetimeHeapObjects) == nil || kvs.Get(profGoLifetimeHeapBytes) == nil {
		if heapFile, ok := pprofFiles[goHeapFile]; ok {
			objects, size, err := liveHeapSummary(heapFile)
			if err != nil {
				l.Warnf("unable to resolve go pprof live heap metrics: %v", err)
			} else {
				if kvs.Get(profGoLifetimeHeapObjects) == nil {
					kvs = kvs.Add(profGoLifetimeHeapObjects, objects)
				}
				if kvs.Get(profGoLifetimeHeapBytes) == nil {
					kvs = kvs.Add(profGoLifetimeHeapBytes, size)
				}
			}
		}
	}

	if kvs.Get(profGoBlockedTime) == nil {
		if blockFile, ok := pprofFiles[goBlockFile]; ok {
			delayNS, err := delayDurationNS(blockFile)
			if err != nil {
				l.Warnf("unable to resolve go pprof block delay duration: %v", err)
			} else {
				kvs = kvs.Add(profGoBlockedTime, delayNS)
			}
		}
	}

	if kvs.Get(profGoMutexDelayTime) == nil {
		if mutexFile, ok := pprofFiles[goMutexFile]; ok {
			delayNS, err := delayDurationNS(mutexFile)
			if err != nil {
				l.Warnf("unable to resolve go pprof mutex delay duration: %v", err)
			} else {
				kvs = kvs.Add(profGoMutexDelayTime, delayNS)
			}
		}
	}

	if kvs.Get(profGoNumGoroutine) == nil {
		if goroutineFile, ok := pprofFiles[goroutinesFile]; ok {
			gCount, err := goroutinesCount(goroutineFile)
			if err != nil {
				l.Warnf("unable to resolve go pprof goroutines count metric: %w", err)
			} else {
				kvs = kvs.Add(profGoNumGoroutine, gCount)
			}
		}
	}

	pt := point.NewPoint(metricsName, kvs, point.WithPrecision(point.PrecNS), point.WithTime(pprofEnd))
	return []*point.Point{pt}, nil
}
