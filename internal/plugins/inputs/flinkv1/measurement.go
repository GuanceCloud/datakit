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
		Name: "flink_jobmanager",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"Status_JVM_CPU_Load":                                   newGaugeFieldInfo("The recent CPU usage of the JVM."),
			"Status_JVM_CPU_Time":                                   newGaugeFieldInfo("The CPU time used by the JVM."),
			"Status_JVM_ClassLoader_ClassesLoaded":                  newGaugeFieldInfo("The total number of classes loaded since the start of the JVM."),
			"Status_JVM_ClassLoader_ClassesUnloaded":                newGaugeFieldInfo("The total number of classes unloaded since the start of the JVM."),
			"Status_JVM_GarbageCollector_Copy_Count":                newGaugeFieldInfo("The total number of collections that have occurred."),
			"Status_JVM_GarbageCollector_Copy_Time":                 newGaugeFieldInfo("The total time spent performing garbage collection."),
			"Status_JVM_GarbageCollector_MarkSweepCompact_Count":    newGaugeFieldInfo("The total number of collections that have occurred."),
			"Status_JVM_GarbageCollector_MarkSweepCompact_Time":     newGaugeFieldInfo("The total time spent performing garbage collection."),
			"Status_JVM_GarbageCollector_G1_Old_Generation_Count":   newGaugeFieldInfo("The total number of collections that have occurred."),
			"Status_JVM_GarbageCollector_G1_Old_Generation_Time":    newGaugeFieldInfo("The total time spent performing garbage collection."),
			"Status_JVM_GarbageCollector_G1_Young_Generation_Count": newGaugeFieldInfo("The total number of collections that have occurred."),
			"Status_JVM_GarbageCollector_G1_Young_Generation_Time":  newGaugeFieldInfo("The total time spent performing garbage collection."),
			"Status_JVM_Memory_Direct_Count":                        newGaugeFieldInfo("The number of buffers in the direct buffer pool."),
			"Status_JVM_Memory_Direct_MemoryUsed":                   newGaugeFieldInfo("The amount of memory used by the JVM for the direct buffer pool."),
			"Status_JVM_Memory_Direct_TotalCapacity":                newGaugeFieldInfo("The total capacity of all buffers in the direct buffer pool."),
			"Status_JVM_Memory_Heap_Committed":                      newGaugeFieldInfo("The amount of heap memory guaranteed to be available to the JVM."),
			"Status_JVM_Memory_Heap_Max":                            newGaugeFieldInfo("The maximum amount of heap memory that can be used for memory management."),
			"Status_JVM_Memory_Heap_Used":                           newGaugeFieldInfo("The amount of heap memory currently used."),
			"Status_JVM_Memory_Mapped_Count":                        newGaugeFieldInfo("The number of buffers in the mapped buffer pool."),
			"Status_JVM_Memory_Mapped_MemoryUsed":                   newGaugeFieldInfo("The amount of memory used by the JVM for the mapped buffer pool."),
			"Status_JVM_Memory_Mapped_TotalCapacity":                newGaugeFieldInfo("The number of buffers in the mapped buffer pool."),
			"Status_JVM_Memory_Metaspace_Committed":                 newGaugeFieldInfo("The amount of memory guaranteed to be available to the JVM in the meta-space memory pool (in bytes)."),
			"Status_JVM_Memory_Metaspace_Max":                       newGaugeFieldInfo("The maximum amount of memory that can be used in the meta-space memory pool (in bytes)."),
			"Status_JVM_Memory_Metaspace_Used":                      newGaugeFieldInfo("Used bytes of a given JVM memory area"),
			"Status_JVM_Memory_NonHeap_Committed":                   newGaugeFieldInfo("The amount of non-heap memory guaranteed to be available to the JVM."),
			"Status_JVM_Memory_NonHeap_Max":                         newGaugeFieldInfo("The maximum amount of non-heap memory that can be used for memory management"),
			"Status_JVM_Memory_NonHeap_Used":                        newGaugeFieldInfo("The amount of non-heap memory currently used."),
			"Status_JVM_Threads_Count":                              newGaugeFieldInfo("The total number of live threads."),
			"numRegisteredTaskManagers":                             newGaugeFieldInfo("The number of registered task managers."),
			"numRunningJobs":                                        newGaugeFieldInfo("The number of running jobs."),
			"taskSlotsAvailable":                                    newGaugeFieldInfo("The number of available task slots."),
			"taskSlotsTotal":                                        newGaugeFieldInfo("The total number of task slots."),
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
		Name: "flink_taskmanager",
		Cat:  point.Metric,
		Fields: map[string]interface{}{
			"Status_Flink_Memory_Managed_Total":                     newGaugeFieldInfo("The total amount of managed memory."),
			"Status_Flink_Memory_Managed_Used":                      newGaugeFieldInfo("The amount of managed memory currently used."),
			"Status_JVM_CPU_Load":                                   newGaugeFieldInfo("The recent CPU usage of the JVM."),
			"Status_JVM_CPU_Time":                                   newGaugeFieldInfo("The CPU time used by the JVM."),
			"Status_JVM_ClassLoader_ClassesLoaded":                  newGaugeFieldInfo("The total number of classes loaded since the start of the JVM."),
			"Status_JVM_ClassLoader_ClassesUnloaded":                newGaugeFieldInfo("The total number of classes unloaded since the start of the JVM."),
			"Status_JVM_GarbageCollector_G1_Old_Generation_Count":   newGaugeFieldInfo("The total number of collections that have occurred."),
			"Status_JVM_GarbageCollector_G1_Old_Generation_Time":    newGaugeFieldInfo("The total time spent performing garbage collection."),
			"Status_JVM_GarbageCollector_G1_Young_Generation_Count": newGaugeFieldInfo("The total number of collections that have occurred."),
			"Status_JVM_GarbageCollector_G1_Young_Generation_Time":  newGaugeFieldInfo("The total time spent performing garbage collection."),
			"Status_JVM_Memory_Direct_Count":                        newGaugeFieldInfo("The number of buffers in the direct buffer pool."),
			"Status_JVM_Memory_Direct_MemoryUsed":                   newGaugeFieldInfo("The amount of memory used by the JVM for the direct buffer pool."),
			"Status_JVM_Memory_Direct_TotalCapacity":                newGaugeFieldInfo("The total capacity of all buffers in the direct buffer pool."),
			"Status_JVM_Memory_Heap_Committed":                      newGaugeFieldInfo("The amount of heap memory guaranteed to be available to the JVM."),
			"Status_JVM_Memory_Heap_Max":                            newGaugeFieldInfo("The maximum amount of heap memory that can be used for memory management."),
			"Status_JVM_Memory_Heap_Used":                           newGaugeFieldInfo("The amount of heap memory currently used."),
			"Status_JVM_Memory_Mapped_Count":                        newGaugeFieldInfo("The number of buffers in the mapped buffer pool."),
			"Status_JVM_Memory_Mapped_MemoryUsed":                   newGaugeFieldInfo("The amount of memory used by the JVM for the mapped buffer pool."),
			"Status_JVM_Memory_Mapped_TotalCapacity":                newGaugeFieldInfo("The number of buffers in the mapped buffer pool."),
			"Status_JVM_Memory_Metaspace_Committed":                 newGaugeFieldInfo("The amount of memory guaranteed to be available to the JVM in the meta-space memory pool (in bytes)."),
			"Status_JVM_Memory_Metaspace_Max":                       newGaugeFieldInfo("The maximum amount of memory that can be used in the meta-space memory pool (in bytes)."),
			"Status_JVM_Memory_Metaspace_Used":                      newGaugeFieldInfo("Used bytes of a given JVM memory area"),
			"Status_JVM_Memory_NonHeap_Committed":                   newGaugeFieldInfo("The amount of non-heap memory guaranteed to be available to the JVM."),
			"Status_JVM_Memory_NonHeap_Max":                         newGaugeFieldInfo("The maximum amount of non-heap memory that can be used for memory management"),
			"Status_JVM_Memory_NonHeap_Used":                        newGaugeFieldInfo("The amount of non-heap memory currently used."),
			"Status_JVM_Threads_Count":                              newGaugeFieldInfo("The total number of live threads."),
			"Status_Network_AvailableMemorySegments":                newGaugeFieldInfo("The number of unused memory segments."),
			"Status_Network_TotalMemorySegments":                    newGaugeFieldInfo("The number of allocated memory segments."),
			"Status_Shuffle_Netty_AvailableMemory":                  newGaugeFieldInfo("The amount of unused memory in bytes."),
			"Status_Shuffle_Netty_AvailableMemorySegments":          newGaugeFieldInfo("The number of unused memory segments."),
			"Status_Shuffle_Netty_TotalMemory":                      newGaugeFieldInfo("The amount of allocated memory in bytes."),
			"Status_Shuffle_Netty_TotalMemorySegments":              newGaugeFieldInfo("The number of allocated memory segments."),
			"Status_Shuffle_Netty_UsedMemory":                       newGaugeFieldInfo("The amount of used memory in bytes."),
			"Status_Shuffle_Netty_UsedMemorySegments":               newGaugeFieldInfo("The number of used memory segments."),
			"Status_Shuffle_Netty_RequestedMemoryUsage":             newGaugeFieldInfo("Experimental: The usage of the network memory. Shows (as percentage) the total amount of requested memory from all of the subtasks. It can exceed 100% as not all requested memory is required for subtask to make progress. However if usage exceeds 100% throughput can suffer greatly and please consider increasing available network memory, or decreasing configured size of network buffer pools."),
		},
		Tags: map[string]interface{}{
			"host":  inputs.NewTagInfo("Host name."),
			"tm_id": inputs.NewTagInfo("Task manager ID."),
		},
	}
}

func newGaugeFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Gauge,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}
