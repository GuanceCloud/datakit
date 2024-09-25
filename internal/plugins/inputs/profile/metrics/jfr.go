// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package metrics

import (
	"fmt"
	"time"

	"github.com/grafana/jfr-parser/common/attributes"
	"github.com/grafana/jfr-parser/common/filters"
	"github.com/grafana/jfr-parser/common/units"
	"github.com/grafana/jfr-parser/parser"
)

const (
	defaultCPUSampleInterval  = 10_000_000 // 10ms
	defaultWallSampleInterval = 10_000_000 // 10ms
)

type jfrChunks []*parser.Chunk

var minValidTime = time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)

func (j jfrChunks) jvmStartTime() (startTime time.Time, err error) {
	var quantity units.IQuantity
	for _, chunk := range j {
		for _, event := range chunk.Apply(filters.VmInfo) {
			if event != nil {
				if quantity, err = attributes.JVMStartTime.GetValue(event); err == nil {
					if startTime, err = units.ToTime(quantity); err == nil && minValidTime.Before(startTime) {
						return
					}
				}
			}
		}
	}
	return startTime, fmt.Errorf("unable to get jvm start time: %w", err)
}

type ddProfilerSetting struct {
	cpuIntervalNanos  int64
	wallIntervalNanos int64
}

func (j jfrChunks) resolveDDProfilerSetting() *ddProfilerSetting {
	cfg := &ddProfilerSetting{
		cpuIntervalNanos:  -1,
		wallIntervalNanos: -1,
	}

	for _, chunk := range j {
		for _, event := range chunk.Apply(filters.DatadogProfilerConfig) {
			if cfg.cpuIntervalNanos <= 0 {
				if cpuInterval, err := attributes.CpuSamplingInterval.GetValue(event); err == nil {
					if nanoInterval, err := cpuInterval.In(units.Nanosecond); err == nil && nanoInterval.IntValue() > 0 {
						cfg.cpuIntervalNanos = nanoInterval.IntValue()
					}
				}
			}
			if cfg.wallIntervalNanos <= 0 {
				if wallInterval, err := attributes.WallSampleInterval.GetValue(event); err == nil {
					if nanoInterval, err := wallInterval.In(units.Nanosecond); err == nil && nanoInterval.IntValue() > 0 {
						cfg.wallIntervalNanos = nanoInterval.IntValue()
					}
				}
			}

			if cfg.cpuIntervalNanos > 0 && cfg.wallIntervalNanos > 0 {
				return cfg
			}
		}
	}
	if cfg.cpuIntervalNanos <= 0 {
		cfg.cpuIntervalNanos = defaultCPUSampleInterval
	}
	if cfg.wallIntervalNanos <= 0 {
		cfg.wallIntervalNanos = defaultWallSampleInterval
	}
	return cfg
}

func (j jfrChunks) cpuTimeDurationNS() int64 {
	cfg := j.resolveDDProfilerSetting()

	totalSamples := int64(0)
	for _, chunk := range j {
		for _, event := range chunk.Apply(filters.DatadogExecutionSample) {
			weight, err := attributes.SampleWeight.GetValue(event)
			if err != nil {
				log.Warnf("unable to get datadog execution sample weight: %v", err)
				continue
			}
			totalSamples += weight
		}
	}
	return totalSamples * cfg.cpuIntervalNanos
}

func (j jfrChunks) allocations() (allocBytes float64, allocCount float64) {
	for _, chunk := range j {
		for _, event := range chunk.Apply(filters.DatadogAllocationSample) {
			size, err := attributes.AllocSize.GetValue(event)
			if err != nil {
				log.Warnf("unable to resolve ddprof allocation size: %v", err)
				continue
			}
			byteSize, err := size.In(units.Byte)
			if err != nil {
				log.Warnf("unable to convert allocation size to bytes: %v", err)
				continue
			}
			weight, err := attributes.AllocWeight.GetValue(event)
			if err != nil {
				log.Warnf("unable to resolve ddprof allocation weight: %v", err)
				continue
			}
			allocBytes += byteSize.FloatValue() * weight
			allocCount += weight
		}
	}
	return
}

func (j jfrChunks) directAllocationBytes() (totalBytes int64) {
	for _, chunk := range j {
		for _, event := range chunk.Apply(filters.DatadogDirectAllocationTotal) {
			value, err := attributes.Allocated.GetValue(event)
			if err != nil {
				log.Warnf("unable to resolve datadog direct allocation allocated bytes: %v", err)
				continue
			}
			if value.Unit() != units.Byte {
				value, err = value.In(units.Byte)
				if err != nil {
					log.Warnf("unable to convert direct allocation allocated to unit byte: %v", err)
					continue
				}
			}
			totalBytes += value.IntValue()
		}
	}
	return
}

func (j jfrChunks) classLoaderCount() (count int64) {
	for _, chunk := range j {
		count += int64(len(chunk.Apply(filters.ClassLoaderStatistics)))
	}
	return
}

func (j jfrChunks) exceptionCount() (count int64) {
	for _, chunk := range j {
		count += int64(len(chunk.Apply(filters.DatadogExceptionSample)))
	}
	return
}

func (j jfrChunks) ioRead(filter parser.EventFilter) (maxReadTimeNS int64, maxBytesRead int64, totalReadTimeNS int64, totalBytesRead int64) {
	for _, chunk := range j {
		for _, event := range chunk.Apply(filter) {
			duration, err := attributes.Duration.GetValue(event)
			if err != nil {
				log.Warnf("unable to resolve file read duration: %v", err)
				continue
			}
			if duration.Unit() != units.Nanosecond {
				duration, err = duration.In(units.Nanosecond)
				if err != nil {
					log.Warnf("unable to convert to file read duration to nanoseconds: %v", err)
					continue
				}
			}
			durationNS := duration.IntValue()
			if maxReadTimeNS < durationNS {
				maxReadTimeNS = durationNS
			}
			totalReadTimeNS += durationNS

			bytesRead, err := attributes.BytesRead.GetValue(event)
			if err != nil {
				log.Warnf("unable to resolve file read bytes: %v", err)
				continue
			}
			if bytesRead.Unit() != units.Byte {
				bytesRead, err = bytesRead.In(units.Byte)
				if err != nil {
					log.Warnf("unable to convert to file bytesread to bytes: %v", err)
					continue
				}
			}

			bytesNum := bytesRead.IntValue()
			if maxBytesRead < bytesNum {
				maxBytesRead = bytesNum
			}
			totalBytesRead += bytesNum
		}
	}
	return // nolint:nakedret
}

func (j jfrChunks) ioWrite(filter parser.EventFilter) (maxWriteTimeNS int64, maxBytesWritten int64, totalWriteTimeNS int64, totalBytesWritten int64) {
	for _, chunk := range j {
		for _, event := range chunk.Apply(filter) {
			duration, err := attributes.Duration.GetValue(event)
			if err != nil {
				log.Warnf("unable to resolve file read duration: %v", err)
				continue
			}
			if duration.Unit() != units.Nanosecond {
				duration, err = duration.In(units.Nanosecond)
				if err != nil {
					log.Warnf("unable to convert to file read duration to nanoseconds: %v", err)
					continue
				}
			}
			durationNS := duration.IntValue()
			if maxWriteTimeNS < durationNS {
				maxWriteTimeNS = durationNS
			}
			totalWriteTimeNS += durationNS

			bytesWritten, err := attributes.BytesWritten.GetValue(event)
			if err != nil {
				log.Warnf("unable to resolve file read bytes: %v", err)
				continue
			}
			if bytesWritten.Unit() != units.Byte {
				bytesWritten, err = bytesWritten.In(units.Byte)
				if err != nil {
					log.Warnf("unable to convert to file bytesread to bytes: %v", err)
					continue
				}
			}

			bytesNum := bytesWritten.IntValue()
			if maxBytesWritten < bytesNum {
				maxBytesWritten = bytesNum
			}
			totalBytesWritten += bytesNum
		}
	}
	return // nolint:nakedret
}

func (j jfrChunks) fileRead() (maxReadTimeNS int64, maxBytesRead int64, totalReadTimeNS int64, totalBytesRead int64) {
	return j.ioRead(filters.FileRead)
}

func (j jfrChunks) fileWrite() (maxWriteTimeNS int64, maxBytesWritten int64, totalWriteTimeNS int64, totalBytesWritten int64) {
	return j.ioWrite(filters.FileWrite)
}

func (j jfrChunks) gcDuration() (durationNS, count int64) {
	for _, chunk := range j {
		for _, event := range chunk.Apply(filters.GarbageCollection) {
			duration, err := attributes.Duration.GetValue(event)
			if err != nil {
				log.Warnf("unable to resolve jfr GC duration: %v", err)
				continue
			}
			if duration.Unit() != units.Nanosecond {
				duration, err = duration.In(units.Nanosecond)
				if err != nil {
					log.Warnf("unable to convert to GC duration to ns: %v", err)
					continue
				}
			}
			durationNS += duration.IntValue()
			count++
		}
	}
	return
}

func (j jfrChunks) gcPauseDuration() (maxPauseNS, totalPauseNS, pauseCount int64) {
	for _, chunk := range j {
		for _, event := range chunk.Apply(filters.GcPause) {
			duration, err := attributes.Duration.GetValue(event)
			if err != nil {
				log.Warnf("unable to resolve to GC pause duration: %v", err)
				continue
			}
			if duration.Unit() != units.Nanosecond {
				duration, err = duration.In(units.Nanosecond)
				if err != nil {
					log.Warnf("unable to convert to GC pause duration to nanoseconds: %v", err)
					continue
				}
			}

			durationNS := duration.IntValue()
			if maxPauseNS < durationNS {
				maxPauseNS = durationNS
			}
			totalPauseNS += durationNS
			pauseCount++
		}
	}
	return
}

func (j jfrChunks) liveHeapSamples() (count int64) {
	for _, chunk := range j {
		count += int64(len(chunk.Apply(filters.DatadogHeapLiveObject)))
	}
	return
}

func (j jfrChunks) jvmHeapUsage() (usageBytes int64) {
	for _, chunk := range j {
		for _, event := range chunk.Apply(filters.DatadogHeapUsage) {
			size, err := attributes.Size.GetValue(event)
			if err != nil {
				log.Warnf("unable to resolve jvm heap usage: %v", err)
				continue
			}
			if size.Unit() != units.Byte {
				size, err = size.In(units.Byte)
				if err != nil {
					log.Warnf("unable to convert to jvm heap usage to byte: %v", err)
					continue
				}
			}
			usageBytes = size.IntValue()
			return
		}
	}
	return
}

func (j jfrChunks) threadContextSwitchRate() float64 {
	var (
		sum   float64
		count int64
	)
	for _, chunk := range j {
		for _, event := range chunk.Apply(filters.ContextSwitchRate) {
			rate, err := attributes.SwitchRate.GetValue(event)
			if err != nil {
				log.Warnf("unable to resolve thread context switch rate: %v", err)
				continue
			}
			sum += rate
			count++
		}
	}
	if count > 0 {
		return sum / float64(count)
	}
	return 0
}

func (j jfrChunks) liveHeap() (totalBytes, objectCount float64) { //nolint: unused
	for _, chunk := range j {
		for _, event := range chunk.Apply(filters.DatadogHeapLiveObject) {
			size, err := attributes.Size.GetValue(event)
			if err != nil {
				log.Warnf("unable to resolve live heap size: %v", err)
				continue
			}
			if size.Unit() != units.Byte {
				size, err = size.In(units.Byte)
				if err != nil {
					log.Warnf("unable to convert to jfr live heap size to unit bytes: %v", err)
					continue
				}
			}

			weight, err := attributes.HeapWeight.GetValue(event)
			if err != nil {
				log.Warnf("unable to resolve live heap weight: %v", err)
				continue
			}

			totalBytes += weight * size.FloatValue()
			objectCount += weight
		}
	}
	return
}

func (j jfrChunks) threadStart() (count int64) {
	for _, chunk := range j {
		count += int64(len(chunk.Apply(filters.ThreadStart)))
	}
	return
}

func (j jfrChunks) deadlockedThread() (count int64) {
	for _, chunk := range j {
		count += int64(len(chunk.Apply(filters.DatadogDeadlockedThread)))
	}
	return
}

func (j jfrChunks) monitorEnter() (maxDurationNS, totalDurationNS float64, count int64) {
	for _, chunk := range j {
		for _, event := range chunk.Apply(filters.MonitorEnter) {
			duration, err := attributes.Duration.GetValue(event)
			if err != nil {
				log.Warnf("unable to resolve monitor blocked duration: %v", err)
				continue
			}
			if duration.Unit() != units.Nanosecond {
				if duration, err = duration.In(units.Nanosecond); err != nil {
					log.Warnf("unable to convert to nanoseconds: %v", err)
					continue
				}
			}
			durationNS := duration.FloatValue()
			if maxDurationNS < durationNS {
				maxDurationNS = durationNS
			}
			totalDurationNS += durationNS
			count++
		}
	}
	return
}

func (j jfrChunks) compilationDuration() (totalDuration int64) {
	for _, chunk := range j {
		for _, event := range chunk.Apply(filters.Compilation) {
			duration, err := attributes.Duration.GetValue(event)
			if err != nil {
				log.Warnf("unable to resolve jvm compilation duration: %v", err)
				continue
			}
			if duration.Unit() != units.Nanosecond {
				if duration, err = duration.In(units.Nanosecond); err != nil {
					log.Warnf("unable to convert to compilation duration to nanoseconds: %v", err)
					continue
				}
			}
			totalDuration += duration.IntValue()
		}
	}
	return
}

func (j jfrChunks) socketIORead() (maxReadTimeNS int64, maxBytesRead int64, totalReadTimeNS int64, totalBytesRead int64) {
	return j.ioRead(filters.SocketRead)
}

func (j jfrChunks) socketIOWrite() (maxWriteTimeNS int64, maxBytesWritten int64, totalWriteTimeNS int64, totalBytesWritten int64) {
	return j.ioWrite(filters.SocketWrite)
}
