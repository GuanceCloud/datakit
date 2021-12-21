package flinkv1

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type JobmanagerMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

type TaskmanagerMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (mm *JobmanagerMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(mm.name, mm.tags, mm.fields, mm.ts)
}

func (mm *TaskmanagerMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(mm.name, mm.tags, mm.fields, mm.ts)
}

//nolint:lll
// Info flink_jobmanager_.
func (*JobmanagerMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "Jobmanager Metrics",
		Fields: map[string]interface{}{
			"Status_JVM_CPU_Load":                                newCountFieldInfo("The recent CPU usage of the JVM."),
			"Status_JVM_CPU_Time":                                newCountFieldInfo("The CPU time used by the JVM."),
			"Status_JVM_ClassLoader_ClassesLoaded":               newCountFieldInfo("The total number of classes loaded since the start of the JVM."),
			"Status_JVM_ClassLoader_ClassesUnloaded":             newCountFieldInfo("The total number of classes unloaded since the start of the JVM."),
			"Status_JVM_GarbageCollector_Copy_Count":             newCountFieldInfo("The total number of collections that have occurred."),
			"Status_JVM_GarbageCollector_Copy_Time":              newCountFieldInfo("The total time spent performing garbage collection."),
			"Status_JVM_GarbageCollector_MarkSweepCompact_Count": newCountFieldInfo("The total number of collections that have occurred."),
			"Status_JVM_GarbageCollector_MarkSweepCompact_Time":  newCountFieldInfo("The total time spent performing garbage collection."),
			"Status_JVM_Memory_Direct_Count":                     newCountFieldInfo("The number of buffers in the direct buffer pool."),
			"Status_JVM_Memory_Direct_MemoryUsed":                newCountFieldInfo("The amount of memory used by the JVM for the direct buffer pool."),
			"Status_JVM_Memory_Direct_TotalCapacity":             newCountFieldInfo("The total capacity of all buffers in the direct buffer pool."),
			"Status_JVM_Memory_Heap_Committed":                   newCountFieldInfo("The amount of heap memory guaranteed to be available to the JVM."),
			"Status_JVM_Memory_Heap_Max":                         newCountFieldInfo("The maximum amount of heap memory that can be used for memory management."),
			"Status_JVM_Memory_Heap_Used":                        newCountFieldInfo("The amount of heap memory currently used."),
			"Status_JVM_Memory_Mapped_Count":                     newCountFieldInfo("The number of buffers in the mapped buffer pool."),
			"Status_JVM_Memory_Mapped_MemoryUsed":                newCountFieldInfo("The amount of memory used by the JVM for the mapped buffer pool."),
			"Status_JVM_Memory_Mapped_TotalCapacity":             newCountFieldInfo("The number of buffers in the mapped buffer pool."),
			"Status_JVM_Memory_Metaspace_Committed":              newCountFieldInfo("The amount of memory guaranteed to be available to the JVM in the Metaspace memory pool (in bytes)."),
			"Status_JVM_Memory_Metaspace_Max":                    newCountFieldInfo("The maximum amount of memory that can be used in the Metaspace memory pool (in bytes)."),
			"Status_JVM_Memory_Metaspace_Used":                   newCountFieldInfo("Used bytes of a given JVM memory area"),
			"Status_JVM_Memory_NonHeap_Committed":                newCountFieldInfo("The amount of non-heap memory guaranteed to be available to the JVM."),
			"Status_JVM_Memory_NonHeap_Max":                      newCountFieldInfo("The maximum amount of non-heap memory that can be used for memory management"),
			"Status_JVM_Memory_NonHeap_Used":                     newCountFieldInfo("The amount of non-heap memory currently used."),
			"Status_JVM_Threads_Count":                           newCountFieldInfo("The total number of live threads."),
			"numRegisteredTaskManagers":                          newCountFieldInfo("The number of registered taskmanagers."),
			"numRunningJobs":                                     newCountFieldInfo("The number of running jobs."),
			"taskSlotsAvailable":                                 newCountFieldInfo("The number of available task slots."),
			"taskSlotsTotal":                                     newCountFieldInfo("The total number of task slots."),
		},
		Tags: map[string]interface{}{
			"host": inputs.NewTagInfo("host name"),
		},
	}
}

//nolint:lll
// Info flink_taskmanager_.
func (*TaskmanagerMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "Taskmanager Metrics",
		Fields: map[string]interface{}{
			"Status_Flink_Memory_Managed_Total":                     newCountFieldInfo("The total amount of managed memory."),
			"Status_Flink_Memory_Managed_Used":                      newCountFieldInfo("The amount of managed memory currently used."),
			"Status_JVM_CPU_Load":                                   newCountFieldInfo("The recent CPU usage of the JVM."),
			"Status_JVM_CPU_Time":                                   newCountFieldInfo("The CPU time used by the JVM."),
			"Status_JVM_ClassLoader_ClassesLoaded":                  newCountFieldInfo("The total number of classes loaded since the start of the JVM."),
			"Status_JVM_ClassLoader_ClassesUnloaded":                newCountFieldInfo("The total number of classes unloaded since the start of the JVM."),
			"Status_JVM_GarbageCollector_G1_Old_Generation_Count":   newCountFieldInfo("The total number of collections that have occurred."),
			"Status_JVM_GarbageCollector_G1_Old_Generation_Time":    newCountFieldInfo("The total time spent performing garbage collection."),
			"Status_JVM_GarbageCollector_G1_Young_Generation_Count": newCountFieldInfo("The total number of collections that have occurred."),
			"Status_JVM_GarbageCollector_G1_Young_Generation_Time":  newCountFieldInfo("The total time spent performing garbage collection."),
			"Status_JVM_Memory_Direct_Count":                        newCountFieldInfo("The number of buffers in the direct buffer pool."),
			"Status_JVM_Memory_Direct_MemoryUsed":                   newCountFieldInfo("The amount of memory used by the JVM for the direct buffer pool."),
			"Status_JVM_Memory_Direct_TotalCapacity":                newCountFieldInfo("The total capacity of all buffers in the direct buffer pool."),
			"Status_JVM_Memory_Heap_Committed":                      newCountFieldInfo("The amount of heap memory guaranteed to be available to the JVM."),
			"Status_JVM_Memory_Heap_Max":                            newCountFieldInfo("The maximum amount of heap memory that can be used for memory management."),
			"Status_JVM_Memory_Heap_Used":                           newCountFieldInfo("The amount of heap memory currently used."),
			"Status_JVM_Memory_Mapped_Count":                        newCountFieldInfo("The number of buffers in the mapped buffer pool."),
			"Status_JVM_Memory_Mapped_MemoryUsed":                   newCountFieldInfo("The amount of memory used by the JVM for the mapped buffer pool."),
			"Status_JVM_Memory_Mapped_TotalCapacity":                newCountFieldInfo("The number of buffers in the mapped buffer pool."),
			"Status_JVM_Memory_Metaspace_Committed":                 newCountFieldInfo("The amount of memory guaranteed to be available to the JVM in the Metaspace memory pool (in bytes)."),
			"Status_JVM_Memory_Metaspace_Max":                       newCountFieldInfo("The maximum amount of memory that can be used in the Metaspace memory pool (in bytes)."),
			"Status_JVM_Memory_Metaspace_Used":                      newCountFieldInfo("Used bytes of a given JVM memory area"),
			"Status_JVM_Memory_NonHeap_Committed":                   newCountFieldInfo("The amount of non-heap memory guaranteed to be available to the JVM."),
			"Status_JVM_Memory_NonHeap_Max":                         newCountFieldInfo("The maximum amount of non-heap memory that can be used for memory management"),
			"Status_JVM_Memory_NonHeap_Used":                        newCountFieldInfo("The amount of non-heap memory currently used."),
			"Status_JVM_Threads_Count":                              newCountFieldInfo("The total number of live threads."),
			"Status_Network_AvailableMemorySegments":                newCountFieldInfo("The number of unused memory segments."),
			"Status_Network_TotalMemorySegments":                    newCountFieldInfo("The number of allocated memory segments."),
			"Status_Shuffle_Netty_AvailableMemory":                  newCountFieldInfo("The amount of unused memory in bytes."),
			"Status_Shuffle_Netty_AvailableMemorySegments":          newCountFieldInfo("The number of unused memory segments."),
			"Status_Shuffle_Netty_TotalMemory":                      newCountFieldInfo("The amount of allocated memory in bytes."),
			"Status_Shuffle_Netty_TotalMemorySegments":              newCountFieldInfo("The number of allocated memory segments."),
			"Status_Shuffle_Netty_UsedMemory":                       newCountFieldInfo("The amount of used memory in bytes."),
			"Status_Shuffle_Netty_UsedMemorySegments":               newCountFieldInfo("The number of used memory segments."),
		},
		Tags: map[string]interface{}{
			"host": inputs.NewTagInfo("host name"),
		},
	}
}

func newCountFieldInfo(desc string) *inputs.FieldInfo {
	return &inputs.FieldInfo{
		DataType: inputs.Int,
		Type:     inputs.Count,
		Unit:     inputs.NCount,
		Desc:     desc,
	}
}
