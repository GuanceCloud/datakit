// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flinkv1

import (
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type JobmanagerMeasurement struct{}

type TaskmanagerMeasurement struct{}

// Info flink_jobmanager_.
//
//nolint:lll
func (*JobmanagerMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "flink_jobmanager",
		Cat:    point.Metric,
		Desc:   "Apache Flink JobManager metrics scraped through the Prometheus reporter, including JobManager JVM runtime metrics and cluster-level task manager, job, and task-slot gauges.",
		DescZh: "通过 Prometheus reporter 采集的 Apache Flink JobManager 指标，包括 JobManager JVM 运行时指标以及集群级 TaskManager、作业和任务槽位指标。",
		Fields: map[string]interface{}{
			"Status_JVM_CPU_Load":                                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal, Desc: "Recent JVM process CPU load as a ratio from 0 to 1."},
			"Status_JVM_CPU_Time":                                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.DurationNS, Desc: "Cumulative CPU time used by the JVM process, in nanoseconds."},
			"Status_JVM_ClassLoader_ClassesLoaded":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of classes loaded since the start of the JVM."},
			"Status_JVM_ClassLoader_ClassesUnloaded":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of classes unloaded since the start of the JVM."},
			"Status_JVM_GarbageCollector_Copy_Count":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of collections that have occurred."},
			"Status_JVM_GarbageCollector_Copy_Time":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "Cumulative time spent performing garbage collection, in milliseconds."},
			"Status_JVM_GarbageCollector_MarkSweepCompact_Count":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of collections that have occurred."},
			"Status_JVM_GarbageCollector_MarkSweepCompact_Time":     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "Cumulative time spent performing garbage collection, in milliseconds."},
			"Status_JVM_GarbageCollector_G1_Old_Generation_Count":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of collections that have occurred."},
			"Status_JVM_GarbageCollector_G1_Old_Generation_Time":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "Cumulative time spent performing garbage collection, in milliseconds."},
			"Status_JVM_GarbageCollector_G1_Young_Generation_Count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of collections that have occurred."},
			"Status_JVM_GarbageCollector_G1_Young_Generation_Time":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "Cumulative time spent performing garbage collection, in milliseconds."},
			"Status_JVM_Memory_Direct_Count":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of buffers in the direct buffer pool."},
			"Status_JVM_Memory_Direct_MemoryUsed":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of JVM memory used by the direct buffer pool, in bytes."},
			"Status_JVM_Memory_Direct_TotalCapacity":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total capacity of all buffers in the direct buffer pool, in bytes."},
			"Status_JVM_Memory_Heap_Committed":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of heap memory guaranteed to be available to the JVM, in bytes."},
			"Status_JVM_Memory_Heap_Max":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of heap memory that can be used for memory management, in bytes."},
			"Status_JVM_Memory_Heap_Used":                           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of heap memory currently used, in bytes."},
			"Status_JVM_Memory_Mapped_Count":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of buffers in the mapped buffer pool."},
			"Status_JVM_Memory_Mapped_MemoryUsed":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of JVM memory used by the mapped buffer pool, in bytes."},
			"Status_JVM_Memory_Mapped_TotalCapacity":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total capacity of all buffers in the mapped buffer pool, in bytes."},
			"Status_JVM_Memory_Metaspace_Committed":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory guaranteed to be available to the JVM in the meta-space memory pool (in bytes)."},
			"Status_JVM_Memory_Metaspace_Max":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of memory that can be used in the meta-space memory pool (in bytes)."},
			"Status_JVM_Memory_Metaspace_Used":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "JVM metaspace memory currently used, in bytes."},
			"Status_JVM_Memory_NonHeap_Committed":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of non-heap memory guaranteed to be available to the JVM, in bytes."},
			"Status_JVM_Memory_NonHeap_Max":                         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of non-heap memory that can be used for memory management, in bytes."},
			"Status_JVM_Memory_NonHeap_Used":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of non-heap memory currently used, in bytes."},
			"Status_JVM_Threads_Count":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of live threads."},
			"numRegisteredTaskManagers":                             &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of registered task managers."},
			"numRunningJobs":                                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of running jobs."},
			"taskSlotsAvailable":                                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of available task slots."},
			"taskSlotsTotal":                                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of task slots."},
		},
		Tags: map[string]interface{}{
			"host": inputs.NewTagInfo("Host name."),
		},
	}
}

// Info flink_taskmanager_.
//
//nolint:lll
func (*TaskmanagerMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   "flink_taskmanager",
		Cat:    point.Metric,
		Desc:   "Apache Flink TaskManager metrics scraped through the Prometheus reporter, including TaskManager managed memory, JVM runtime, network memory segment, and Netty shuffle memory metrics.",
		DescZh: "通过 Prometheus reporter 采集的 Apache Flink TaskManager 指标，包括 TaskManager 托管内存、JVM 运行时、网络内存段和 Netty shuffle 内存指标。",
		Fields: map[string]interface{}{
			"Status_Flink_Memory_Managed_Total":                     &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total amount of Flink managed memory, in bytes."},
			"Status_Flink_Memory_Managed_Used":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of Flink managed memory currently used, in bytes."},
			"Status_JVM_CPU_Load":                                   &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.PercentDecimal, Desc: "Recent JVM process CPU load as a ratio from 0 to 1."},
			"Status_JVM_CPU_Time":                                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.DurationNS, Desc: "Cumulative CPU time used by the JVM process, in nanoseconds."},
			"Status_JVM_ClassLoader_ClassesLoaded":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of classes loaded since the start of the JVM."},
			"Status_JVM_ClassLoader_ClassesUnloaded":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of classes unloaded since the start of the JVM."},
			"Status_JVM_GarbageCollector_G1_Old_Generation_Count":   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of collections that have occurred."},
			"Status_JVM_GarbageCollector_G1_Old_Generation_Time":    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "Cumulative time spent performing garbage collection, in milliseconds."},
			"Status_JVM_GarbageCollector_G1_Young_Generation_Count": &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.NCount, Desc: "The total number of collections that have occurred."},
			"Status_JVM_GarbageCollector_G1_Young_Generation_Time":  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Count, Unit: inputs.DurationMS, Desc: "Cumulative time spent performing garbage collection, in milliseconds."},
			"Status_JVM_Memory_Direct_Count":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of buffers in the direct buffer pool."},
			"Status_JVM_Memory_Direct_MemoryUsed":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of JVM memory used by the direct buffer pool, in bytes."},
			"Status_JVM_Memory_Direct_TotalCapacity":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total capacity of all buffers in the direct buffer pool, in bytes."},
			"Status_JVM_Memory_Heap_Committed":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of heap memory guaranteed to be available to the JVM, in bytes."},
			"Status_JVM_Memory_Heap_Max":                            &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of heap memory that can be used for memory management, in bytes."},
			"Status_JVM_Memory_Heap_Used":                           &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of heap memory currently used, in bytes."},
			"Status_JVM_Memory_Mapped_Count":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of buffers in the mapped buffer pool."},
			"Status_JVM_Memory_Mapped_MemoryUsed":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of JVM memory used by the mapped buffer pool, in bytes."},
			"Status_JVM_Memory_Mapped_TotalCapacity":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The total capacity of all buffers in the mapped buffer pool, in bytes."},
			"Status_JVM_Memory_Metaspace_Committed":                 &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of memory guaranteed to be available to the JVM in the meta-space memory pool (in bytes)."},
			"Status_JVM_Memory_Metaspace_Max":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of memory that can be used in the meta-space memory pool (in bytes)."},
			"Status_JVM_Memory_Metaspace_Used":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "JVM metaspace memory currently used, in bytes."},
			"Status_JVM_Memory_NonHeap_Committed":                   &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of non-heap memory guaranteed to be available to the JVM, in bytes."},
			"Status_JVM_Memory_NonHeap_Max":                         &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The maximum amount of non-heap memory that can be used for memory management, in bytes."},
			"Status_JVM_Memory_NonHeap_Used":                        &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of non-heap memory currently used, in bytes."},
			"Status_JVM_Threads_Count":                              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The total number of live threads."},
			"Status_Network_AvailableMemorySegments":                &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of unused memory segments."},
			"Status_Network_TotalMemorySegments":                    &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of allocated memory segments."},
			"Status_Shuffle_Netty_AvailableMemory":                  &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of unused memory in bytes."},
			"Status_Shuffle_Netty_AvailableMemorySegments":          &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of unused memory segments."},
			"Status_Shuffle_Netty_TotalMemory":                      &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of allocated memory in bytes."},
			"Status_Shuffle_Netty_TotalMemorySegments":              &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of allocated memory segments."},
			"Status_Shuffle_Netty_UsedMemory":                       &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.SizeByte, Desc: "The amount of used memory in bytes."},
			"Status_Shuffle_Netty_UsedMemorySegments":               &inputs.FieldInfo{DataType: inputs.Int, Type: inputs.Gauge, Unit: inputs.NCount, Desc: "The number of used memory segments."},
			"Status_Shuffle_Netty_RequestedMemoryUsage":             &inputs.FieldInfo{DataType: inputs.Float, Type: inputs.Gauge, Unit: inputs.Percent, Desc: "Experimental: The usage of the network memory. Shows (as percentage) the total amount of requested memory from all of the subtasks. It can exceed 100% as not all requested memory is required for subtask to make progress. However if usage exceeds 100% throughput can suffer greatly and please consider increasing available network memory, or decreasing configured size of network buffer pools."},
		},
		Tags: map[string]interface{}{
			"host":  inputs.NewTagInfo("Host name."),
			"tm_id": inputs.NewTagInfo("Task manager ID."),
		},
	}
}
