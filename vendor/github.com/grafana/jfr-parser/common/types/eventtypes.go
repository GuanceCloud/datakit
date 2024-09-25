package types

const (
	jdkTypePrefix     = "jdk."
	datadogTypePrefix = "datadog."
)

const (
	CpuLoad                      = jdkTypePrefix + "CPULoad"
	ThreadCpuLoad                = jdkTypePrefix + "ThreadCPULoad"
	ExecutionSample              = jdkTypePrefix + "ExecutionSample"
	ExecutionSamplingInfoEventId = jdkTypePrefix + "ExecutionSampling"
	NativeMethodSample           = jdkTypePrefix + "NativeMethodSample"
	Processes                    = jdkTypePrefix + "SystemProcess"
	OSMemorySummary              = jdkTypePrefix + "PhysicalMemory"
	OSInformation                = jdkTypePrefix + "OSInformation"
	CPUInformation               = jdkTypePrefix + "CPUInformation"
	ThreadAllocationStatistics   = jdkTypePrefix + "ThreadAllocationStatistics"
	HeapConf                     = jdkTypePrefix + "GCHeapConfiguration"
	GcConf                       = jdkTypePrefix + "GCConfiguration"
	HeapSummary                  = jdkTypePrefix + "GCHeapSummary"
	AllocInsideTlab              = jdkTypePrefix + "ObjectAllocationInNewTLAB"
	AllocOutsideTlab             = jdkTypePrefix + "ObjectAllocationOutsideTLAB"
	ObjAllocSample               = jdkTypePrefix + "ObjectAllocationSample"
	VmInfo                       = jdkTypePrefix + "JVMInformation"
	ClassDefine                  = jdkTypePrefix + "ClassDefine"
	ClassLoad                    = jdkTypePrefix + "ClassLoad"
	ClassUnload                  = jdkTypePrefix + "ClassUnload"
	ClassLoadStatistics          = jdkTypePrefix + "ClassLoadingStatistics"
	ClassLoaderStatistics        = jdkTypePrefix + "ClassLoaderStatistics"
	Compilation                  = jdkTypePrefix + "Compilation"
	FileWrite                    = jdkTypePrefix + "FileWrite"
	FileRead                     = jdkTypePrefix + "FileRead"
	SocketWrite                  = jdkTypePrefix + "SocketWrite"
	SocketRead                   = jdkTypePrefix + "SocketRead"
	ThreadPark                   = jdkTypePrefix + "ThreadPark"
	ThreadSleep                  = jdkTypePrefix + "ThreadSleep"
	MonitorEnter                 = jdkTypePrefix + "JavaMonitorEnter"
	MonitorWait                  = jdkTypePrefix + "JavaMonitorWait"
	MetaspaceOom                 = jdkTypePrefix + "MetaspaceOOM"
	CodeCacheFull                = jdkTypePrefix + "CodeCacheFull"
	CodeCacheStatistics          = jdkTypePrefix + "CodeCacheStatistics"
	CodeSweeperStatistics        = jdkTypePrefix + "CodeSweeperStatistics"
	SweepCodeCache               = jdkTypePrefix + "SweepCodeCache"
	EnvironmentVariable          = jdkTypePrefix + "InitialEnvironmentVariable"
	SystemProperties             = jdkTypePrefix + "InitialSystemProperty"
	ObjectCount                  = jdkTypePrefix + "ObjectCount"
	GcReferenceStatistics        = jdkTypePrefix + "GCReferenceStatistics"
	OldObjectSample              = jdkTypePrefix + "OldObjectSample"
	GcPauseL4                    = jdkTypePrefix + "GCPhasePauseLevel4"
	GcPauseL3                    = jdkTypePrefix + "GCPhasePauseLevel3"
	GcPauseL2                    = jdkTypePrefix + "GCPhasePauseLevel2"
	GcPauseL1                    = jdkTypePrefix + "GCPhasePauseLevel1"
	GcPause                      = jdkTypePrefix + "GCPhasePause"
	MetaspaceSummary             = jdkTypePrefix + "MetaspaceSummary"
	GarbageCollection            = jdkTypePrefix + "GarbageCollection"
	ConcurrentModeFailure        = jdkTypePrefix + "ConcurrentModeFailure"
	ThrowableStatistics          = jdkTypePrefix + "ExceptionStatistics"
	ErrorsThrown                 = jdkTypePrefix + "JavaErrorThrow"
	/*
	 * NOTE: The parser filters all JavaExceptionThrow events created from the Error constructor to
	 * avoid duplicates, so this event type represents 'non error throwables' rather than
	 * exceptions. See note in SyntheticAttributeExtension which does the duplicate filtering.
	 */
	ExceptionsThrown                   = jdkTypePrefix + "JavaExceptionThrow"
	CompilerStats                      = jdkTypePrefix + "CompilerStatistics"
	CompilerFailure                    = jdkTypePrefix + "CompilationFailure"
	UlongFlag                          = jdkTypePrefix + "UnsignedLongFlag"
	BooleanFlag                        = jdkTypePrefix + "BooleanFlag"
	StringFlag                         = jdkTypePrefix + "StringFlag"
	DoubleFlag                         = jdkTypePrefix + "DoubleFlag"
	LongFlag                           = jdkTypePrefix + "LongFlag"
	IntFlag                            = jdkTypePrefix + "IntFlag"
	UintFlag                           = jdkTypePrefix + "UnsignedIntFlag"
	UlongFlagChanged                   = jdkTypePrefix + "UnsignedLongFlagChanged"
	BooleanFlagChanged                 = jdkTypePrefix + "BooleanFlagChanged"
	StringFlagChanged                  = jdkTypePrefix + "StringFlagChanged"
	DoubleFlagChanged                  = jdkTypePrefix + "DoubleFlagChanged"
	LongFlagChanged                    = jdkTypePrefix + "LongFlagChanged"
	IntFlagChanged                     = jdkTypePrefix + "IntFlagChanged"
	UintFlagChanged                    = jdkTypePrefix + "UnsignedIntFlagChanged"
	TimeConversion                     = jdkTypePrefix + "CPUTimeStampCounter"
	ThreadDump                         = jdkTypePrefix + "ThreadDump"
	JfrDataLost                        = jdkTypePrefix + "DataLoss"
	DumpReason                         = jdkTypePrefix + "DumpReason"
	GcConfYoungGeneration              = jdkTypePrefix + "YoungGenerationConfiguration"
	GcConfSurvivor                     = jdkTypePrefix + "GCSurvivorConfiguration"
	GcConfTlab                         = jdkTypePrefix + "GCTLABConfiguration"
	JavaThreadStart                    = jdkTypePrefix + "ThreadStart"
	JavaThreadEnd                      = jdkTypePrefix + "ThreadEnd"
	VmOperations                       = jdkTypePrefix + "ExecuteVMOperation"
	VmShutdown                         = jdkTypePrefix + "Shutdown"
	ThreadStatistics                   = jdkTypePrefix + "JavaThreadStatistics"
	ContextSwitchRate                  = jdkTypePrefix + "ThreadContextSwitchRate"
	CompilerConfig                     = jdkTypePrefix + "CompilerConfiguration"
	CodeCacheConfig                    = jdkTypePrefix + "CodeCacheConfiguration"
	CodeSweeperConfig                  = jdkTypePrefix + "CodeSweeperConfiguration"
	CompilerPhase                      = jdkTypePrefix + "CompilerPhase"
	GcCollectorG1GarbageCollection     = jdkTypePrefix + "G1GarbageCollection"
	GcCollectorOldGarbageCollection    = jdkTypePrefix + "OldGarbageCollection"
	GcCollectorParoldGarbageCollection = jdkTypePrefix + "ParallelOldGarbageCollection"
	GcCollectorYoungGarbageCollection  = jdkTypePrefix + "YoungGarbageCollection"
	GcDetailedAllocationRequiringGc    = jdkTypePrefix + "AllocationRequiringGC"
	GcDetailedEvacuationFailed         = jdkTypePrefix + "EvacuationFailed"
	GcDetailedEvacuationInfo           = jdkTypePrefix + "EvacuationInformation"
	GcDetailedObjectCountAfterGc       = jdkTypePrefix + "ObjectCountAfterGC"
	GcDetailedPromotionFailed          = jdkTypePrefix + "PromotionFailed"
	GcHeapPsSummary                    = jdkTypePrefix + "PSHeapSummary"
	GcMetaspaceAllocationFailure       = jdkTypePrefix + "MetaspaceAllocationFailure"
	GcMetaspaceChunkFreeListSummary    = jdkTypePrefix + "MetaspaceChunkFreeListSummary"
	GcMetaspaceGcThreshold             = jdkTypePrefix + "MetaspaceGCThreshold"
	GcG1mmu                            = jdkTypePrefix + "G1MMU"
	GcG1EvacuationYoungStatistics      = jdkTypePrefix + "G1EvacuationYoungStatistics"
	GcG1EvacuationOldStatistics        = jdkTypePrefix + "G1EvacuationOldStatistics"
	GcG1BasicIHOP                      = jdkTypePrefix + "G1BasicIHOP"
	GcG1HeapRegionTypeChange           = jdkTypePrefix + "G1HeapRegionTypeChange"
	GcG1HeapRegionInformation          = jdkTypePrefix + "G1HeapRegionInformation"
	BiasedLockSelfRevocation           = jdkTypePrefix + "BiasedLockSelfRevocation"
	BiasedLockRevocation               = jdkTypePrefix + "BiasedLockRevocation"
	BiasedLockClassRevocation          = jdkTypePrefix + "BiasedLockClassRevocation"
	GcG1AdaptiveIHOP                   = jdkTypePrefix + "G1AdaptiveIHOP"
	RECORDINGS                         = jdkTypePrefix + "ActiveRecording"
	RecordingSetting                   = jdkTypePrefix + "ActiveSetting"

	// SafepointBegin Safepointing begin
	SafepointBegin = jdkTypePrefix + "SafepointBegin"
	// SafepointStateSync Synchronize run state of threads
	SafepointStateSync = jdkTypePrefix + "SafepointStateSynchronization"
	// SafepointWaitBlocked SAFEPOINT_WAIT_BLOCKED Safepointing begin waiting on running threads to block
	SafepointWaitBlocked = jdkTypePrefix + "SafepointWaitBlocked"
	// SafepointCleanup SAFEPOINT_CLEANUP Safepointing begin running cleanup (parent)
	SafepointCleanup = jdkTypePrefix + "SafepointCleanup"
	// SafepointCleanupTask SAFEPOINT_CLEANUP_TASK Safepointing begin running cleanup task, individual subtasks
	SafepointCleanupTask = jdkTypePrefix + "SafepointCleanupTask"
	// SafepointEnd Safepointing end
	SafepointEnd  = jdkTypePrefix + "SafepointEnd"
	ModuleExport  = jdkTypePrefix + "ModuleExport"
	ModuleRequire = jdkTypePrefix + "ModuleRequire"
	NativeLibrary = jdkTypePrefix + "NativeLibrary"
	HeapDump      = jdkTypePrefix + "HeapDump"
	ProcessStart  = jdkTypePrefix + "ProcessStart"

	JavaMonitorInflate = jdkTypePrefix + "JavaMonitorInflate"
)

const (
	ExceptionCount  = datadogTypePrefix + "ExceptionCount"
	ExceptionSample = datadogTypePrefix + "ExceptionSample"
	ProfilerSetting = datadogTypePrefix + "ProfilerSetting"

	// DatadogProfilerConfig 标识符	名称	说明	内容类型
	//datadog.DatadogProfilerConfig	Datadog Profiler Configuration	[datadog.DatadogProfilerConfig]	null
	DatadogProfilerConfig = datadogTypePrefix + "DatadogProfilerConfig"
	SCOPE                 = datadogTypePrefix + "Scope"

	// DatadogExecutionSample EXECUTION_SAMPLE 标识符	名称	说明	内容类型
	//datadog.ExecutionSample	Method CPU Profiling Sample	[datadog.ExecutionSample]	null
	DatadogExecutionSample = datadogTypePrefix + "ExecutionSample"

	// DatadogMethodSample MethodSample METHOD_SAMPLE 标识符	名称	说明	内容类型
	//datadog.MethodSample	Method Wall Profiling Sample	[datadog.MethodSample]	null
	DatadogMethodSample = datadogTypePrefix + "MethodSample"

	// DatadogAllocationSample ALLOCATION_SAMPLE 标识符	名称	说明	内容类型
	//datadog.ObjectSample	Allocation sample	[datadog.ObjectSample]	null
	DatadogAllocationSample = datadogTypePrefix + "ObjectSample"

	// HeapLiveObject 标识符	名称	说明	内容类型
	//datadog.HeapLiveObject	Heap Live Object	[datadog.HeapLiveObject]	null
	HeapLiveObject = datadogTypePrefix + "HeapLiveObject"

	DatadogEndpoint = datadogTypePrefix + "Endpoint"

	// DatadogDirectAllocationTotal 字段	值	详细模式 (Verbose) 值
	//事件类型	Direct Allocation Total Allocated	Datadog direct allocation count event. [datadog.DirectAllocationTotal]
	DatadogDirectAllocationTotal = datadogTypePrefix + "DirectAllocationTotal"

	// DatadogHeapUsage 标识符	名称	说明	内容类型
	//datadog.HeapUsage	JVM Heap Usage	[datadog.HeapUsage]	null
	DatadogHeapUsage = datadogTypePrefix + "HeapUsage"

	// DatadogDeadlockedThread 标识符	名称	说明	内容类型
	//datadog.DeadlockedThread	Deadlocked Thread	Datadog deadlock detection event - thread details. [datadog.DeadlockedThread]	null
	DatadogDeadlockedThread = datadogTypePrefix + "DeadlockedThread"
)
