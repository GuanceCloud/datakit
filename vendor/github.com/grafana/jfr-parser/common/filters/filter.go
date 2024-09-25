package filters

import (
	"github.com/grafana/jfr-parser/common/attributes"
	"github.com/grafana/jfr-parser/common/types"
	"github.com/grafana/jfr-parser/parser"
	"reflect"
)

const (
	Parked    = "PARKED"
	Runnable  = "RUNNABLE"
	Waiting   = "WAITING"
	Contended = "CONTENDED"
)

var (
	JavaMonitorInflate          = Types(types.JavaMonitorInflate)
	DatadogProfilerSetting      = Types(types.ProfilerSetting)
	DatadogEndpoint             = Types(types.DatadogEndpoint)
	DatadogScope                = Types(types.SCOPE)
	DatadogExceptionCount       = Types(types.ExceptionCount)
	DatadogExceptionSample      = Types(types.ExceptionSample)
	DatadogMethodSample         = Types(types.DatadogMethodSample)
	ThreadParked                = AttributeEqual(attributes.ThreadStat, Parked)
	ThreadWaiting               = AttributeEqual(attributes.ThreadStat, Waiting)
	DdMethodSampleThreadParked  = AndFilters(DatadogMethodSample, ThreadParked)
	DdMethodSampleThreadWaiting = AndFilters(DatadogMethodSample, ThreadWaiting)
	FileIo                      = Types(types.FileRead, types.FileWrite)
	SocketIO                    = Types(types.SocketRead, types.SocketWrite)
	StacktraceNotNull           = NotNull(attributes.EventStacktrace)
	ALLOCATION                  = AndFilters(AllocAll, StacktraceNotNull)
	ObjAllocation               = AndFilters(ObjAlloc, StacktraceNotNull)
	MonitorEnterStacktrace      = AndFilters(MonitorEnter, StacktraceNotNull)
	MonitorWait                 = Types(types.MonitorWait)
	ThreadPark                  = Types(types.ThreadPark)
	ThreadSleep                 = Types(types.ThreadSleep)
	AsyncProfilerLock           = Types(types.MonitorEnter, types.ThreadPark)
	GarbageCollectionStacktrace = AndFilters(GarbageCollection, StacktraceNotNull)
	SYNCHRONIZATION             = Types(types.MonitorWait, types.ThreadPark)
	OldObjectSample             = Types(types.OldObjectSample)
	DatadogProfilerConfig       = Types(types.DatadogProfilerConfig)
	DatadogAllocationSample     = Types(types.DatadogAllocationSample)
	DatadogHeapLiveObject       = Types(types.HeapLiveObject)
)

var (
	SocketRead        = Types(types.SocketRead)
	SocketWrite       = Types(types.SocketWrite)
	SocketReadOrWrite = OrFilters(SocketRead, SocketWrite)
	NoRmiSocketRead   = AndFilters(SocketRead, NotFilter(MethodFilter("sun.rmi.transport.tcp.TCPTransport", "handleMessages")),
		NotFilter(MethodFilter("javax.management.remote.rmi.RMIConnector$RMINotifClient", "fetchNotifs")))
	NoRmiSocketWrite = AndFilters(SocketWrite,
		NotFilter(MethodFilter("sun.rmi.transport.tcp.TCPTransport$ConnectionHandler", "run")),
		NotFilter(MethodFilter("sun.rmi.transport.tcp.TCPTransport$ConnectionHandler", "run0")))
	EnvironmentVariable     = Types(types.EnvironmentVariable)
	FileRead                = Types(types.FileRead)
	FileWrite               = Types(types.FileWrite)
	CodeCacheFull           = Types(types.CodeCacheFull)
	CodeCacheStatistics     = Types(types.CodeCacheStatistics)
	CodeCacheConfig         = Types(types.CodeCacheConfig)
	SweepCodeCache          = Types(types.SweepCodeCache)
	CodeCache               = OrFilters(CodeCacheFull, CodeCacheStatistics, SweepCodeCache, CodeCacheConfig)
	CpuInformation          = Types(types.CPUInformation)
	GcConfig                = Types(types.GcConf)
	HeapConfig              = Types(types.HeapConf)
	BeforeGc                = AttributeEqual(attributes.GcWhen, "Before GC") //$NON-NLS-1$
	AfterGc                 = AttributeEqual(attributes.GcWhen, "After GC")  //$NON-NLS-1$
	AllocOutsideTlab        = Types(types.AllocOutsideTlab)
	AllocInsideTlab         = Types(types.AllocInsideTlab)
	AllocAll                = Types(types.AllocInsideTlab, types.AllocOutsideTlab)
	ObjAlloc                = Types(types.ObjAllocSample)
	ReferenceStatistics     = Types(types.GcReferenceStatistics)
	GarbageCollection       = Types(types.GarbageCollection)
	OldGarbageCollection    = Types(types.GcCollectorOldGarbageCollection)
	YoungGarbageCollection  = Types(types.GcCollectorYoungGarbageCollection)
	ConcurrentModeFailure   = Types(types.ConcurrentModeFailure)
	ERRORS                  = Types(types.ErrorsThrown)
	EXCEPTIONS              = Types(types.ExceptionsThrown)
	Throwables              = OrFilters(EXCEPTIONS, ERRORS)
	ThrowablesStatistics    = Types(types.ThrowableStatistics)
	ClassUnload             = Types(types.ClassUnload)
	ClassLoadStatistics     = Types(types.ClassLoadStatistics)
	ClassLoaderStatistics   = Types(types.ClassLoaderStatistics)
	ClassLoad               = Types(types.ClassLoad)
	ClassLoadOrUnload       = OrFilters(ClassLoad, ClassUnload)
	ClassDefine             = Types(types.ClassDefine)
	ClassLoaderEvents       = OrFilters(ClassLoad, ClassUnload, ClassDefine, ClassLoaderStatistics)
	MonitorEnter            = Types(types.MonitorEnter)
	FileOrSocketIo          = Types(types.SocketRead, types.SocketWrite, types.FileRead, types.FileWrite) // NOTE: Are there more types to add (i.e. relevant types with duration)?
	ThreadLatencies         = Types(types.MonitorEnter, types.MonitorWait, types.ThreadSleep, types.ThreadPark, types.SocketRead, types.SocketWrite, types.FileRead, types.FileWrite, types.ClassLoad, types.Compilation, types.ExecutionSamplingInfoEventId)
	FilterExecutionSample   = Types(types.ExecutionSample)
	DatadogExecutionSample  = Types(types.DatadogExecutionSample)
	ContextSwitchRate       = Types(types.ContextSwitchRate)
	CpuLoad                 = Types(types.CpuLoad)
	GcPause                 = Types(types.GcPause)
	GcPausePhase            = Types(types.GcPauseL1, types.GcPauseL2, types.GcPauseL3, types.GcPauseL4)
	TimeConversion          = Types(types.TimeConversion)
	VmInfo                  = Types(types.VmInfo)
	ThreadDump              = Types(types.ThreadDump)
	SystemProperties        = Types(types.SystemProperties)
	JfrDataLost             = Types(types.JfrDataLost)
	PROCESSES               = Types(types.Processes)
	ObjectCount             = Types(types.ObjectCount)
	MetaspaceOOM            = Types(types.MetaspaceOom)
	Compilation             = Types(types.Compilation)
	CompilerFailure         = Types(types.CompilerFailure)
	CompilerStats           = Types(types.CompilerStats)
	OsMemorySummary         = Types(types.OSMemorySummary)
	HeapSummary             = Types(types.HeapSummary)
	HeapSummaryBeforeGc     = AndFilters(HeapSummary, BeforeGc)
	HeapSummaryAfterGc      = AndFilters(HeapSummary, AfterGc)
	MetaspaceSummary        = Types(types.MetaspaceSummary)
	MetaspaceSummaryAfterGc = AndFilters(MetaspaceSummary, AfterGc)
	RECORDINGS              = Types(types.RECORDINGS)
	RecordingSetting        = Types(types.RecordingSetting)
	SafePoints              = Types(types.SafepointBegin,
		types.SafepointCleanup, types.SafepointCleanupTask, types.SafepointStateSync,
		types.SafepointWaitBlocked, types.SafepointEnd)
	VmOperations                    = Types(types.VmOperations) // NOTE: Not sure if there are any VM events that are neither blocking nor safepoint, but just in case.
	VmOperationsBlockingOrSafepoint = AndFilters(
		Types(types.VmOperations), OrFilters(AttributeEqual(attributes.Blocking, true), AttributeEqual(attributes.Safepoint, true)))
	// VmOperationsSafepoint NOTE: Are there any VM operations that are blocking, but not safepoints. Should we include those in the VM Thread??
	VmOperationsSafepoint        = AndFilters(Types(types.VmOperations), AttributeEqual(attributes.Safepoint, true))
	ApplicationPauses            = OrFilters(GcPause, SafePoints, VmOperationsSafepoint)
	BiasedLockingRevocations     = Types(types.BiasedLockClassRevocation, types.BiasedLockRevocation, types.BiasedLockSelfRevocation)
	ThreadCpuLoad                = Types(types.ThreadCpuLoad)
	NativeMethodSample           = Types(types.NativeMethodSample)
	ThreadStart                  = Types(types.JavaThreadStart)
	ThreadEnd                    = Types(types.JavaThreadEnd)
	DatadogDirectAllocationTotal = Types(types.DatadogDirectAllocationTotal)
	DatadogHeapUsage             = Types(types.DatadogHeapUsage)
	DatadogDeadlockedThread      = Types(types.DatadogDeadlockedThread)
)

type EventFilterFunc func(metadata *parser.ClassMetadata) parser.Predicate[parser.Event]

func (e EventFilterFunc) GetPredicate(metadata *parser.ClassMetadata) parser.Predicate[parser.Event] {
	return e(metadata)
}

type AndPredicate []parser.PredicateFunc

func (a AndPredicate) Test(t parser.Event) bool {
	for _, fn := range a {
		if !fn(t) {
			return false
		}
	}
	return true
}

type OrPredicate []parser.PredicateFunc

func (o OrPredicate) Test(t parser.Event) bool {
	for _, fn := range o {
		if fn(t) {
			return true
		}
	}
	return false
}

type NotPredicate parser.PredicateFunc

func (n NotPredicate) Test(t parser.Event) bool {
	return !n(t)
}

func And(p ...parser.Predicate[parser.Event]) parser.Predicate[parser.Event] {
	ap := make(AndPredicate, 0, len(p))
	for _, pp := range p {
		ap = append(ap, pp.Test)
	}
	return ap
}

func Or(p ...parser.Predicate[parser.Event]) parser.Predicate[parser.Event] {
	op := make(OrPredicate, 0, len(p))
	for _, pp := range p {
		op = append(op, pp.Test)
	}
	return op
}

func Not(p parser.Predicate[parser.Event]) parser.Predicate[parser.Event] {
	return NotPredicate(p.Test)
}

func AndAlways(p ...parser.Predicate[parser.Event]) parser.Predicate[parser.Event] {
	switch len(p) {
	case 0:
		return parser.AlwaysTrue
	case 1:
		return p[0]
	}

	notAlwaysPred := make([]parser.Predicate[parser.Event], 0)
	for _, pp := range p {
		if parser.IsAlwaysFalse(pp) {
			return parser.AlwaysFalse
		}
		if parser.IsAlwaysTrue(pp) {
			continue
		}
		notAlwaysPred = append(notAlwaysPred, pp)
	}
	if len(notAlwaysPred) == 0 {
		return parser.AlwaysTrue
	}
	return And(notAlwaysPred...)
}

func OrAlways(p ...parser.Predicate[parser.Event]) parser.Predicate[parser.Event] {
	switch len(p) {
	case 0:
		return parser.AlwaysFalse
	case 1:
		return p[0]
	}
	notAlwaysPred := make([]parser.Predicate[parser.Event], 0)
	for _, pp := range p {
		if parser.IsAlwaysTrue(pp) {
			return parser.AlwaysTrue
		}
		if parser.IsAlwaysFalse(pp) {
			continue
		}
		notAlwaysPred = append(notAlwaysPred, pp)
	}
	if len(notAlwaysPred) == 0 {
		return parser.AlwaysFalse
	}
	return Or(notAlwaysPred...)
}

func NotAlways(p parser.Predicate[parser.Event]) parser.Predicate[parser.Event] {
	switch {
	case parser.IsAlwaysTrue(p):
		return parser.AlwaysFalse
	case parser.IsAlwaysFalse(p):
		return parser.AlwaysTrue
	}
	return Not(p)
}

func Types(classNames ...string) parser.EventFilter {
	if len(classNames) == 1 {
		return EventFilterFunc(func(metadata *parser.ClassMetadata) parser.Predicate[parser.Event] {
			if classNames[0] == metadata.Name {
				return parser.AlwaysTrue
			}
			return parser.AlwaysFalse
		})
	} else {
		et := make(map[string]struct{}, len(classNames))
		for _, className := range classNames {
			et[className] = struct{}{}
		}
		return EventFilterFunc(func(metadata *parser.ClassMetadata) parser.Predicate[parser.Event] {
			if _, ok := et[metadata.Name]; ok {
				return parser.AlwaysTrue
			}
			return parser.AlwaysFalse
		})
	}
}

func AttributeEqual[T comparable](attr *attributes.Attribute[T], target T) parser.EventFilter {
	return EventFilterFunc(func(metadata *parser.ClassMetadata) parser.Predicate[parser.Event] {

		if parser.IsAlwaysFalse(HasAttribute(attr).GetPredicate(metadata)) {
			return parser.AlwaysFalse
		}

		return parser.PredicateFunc(func(e parser.Event) bool {
			value, err := attr.GetValue(e.(*parser.GenericEvent))
			if err != nil {
				return false
			}
			return value == target
		})
	})
}

func HasAttribute[T any](attr *attributes.Attribute[T]) parser.EventFilter {
	return EventFilterFunc(func(metadata *parser.ClassMetadata) parser.Predicate[parser.Event] {
		f := metadata.GetField(attr.Name)
		if f == nil {
			return parser.AlwaysFalse
		}

		if metadata.ClassMap[f.ClassID].Name != string(attr.ClassName) {
			return parser.AlwaysFalse
		}

		return parser.AlwaysTrue
	})
}

func MethodFilter(typeName, method string) parser.EventFilter {
	return EventFilterFunc(func(metadata *parser.ClassMetadata) parser.Predicate[parser.Event] {
		methodFilter := EventFilterFunc(func(metadata *parser.ClassMetadata) parser.Predicate[parser.Event] {

			return parser.PredicateFunc(func(e parser.Event) bool {
				stacktrace, err := attributes.EventStacktrace.GetValue(e.(*parser.GenericEvent))
				if err != nil {
					return false
				}
				for _, frame := range stacktrace.Frames {
					// todo check type full name
					if frame.Method.Type.Name.String == typeName && frame.Method.Name.String == method {
						return true
					}
				}
				return false
			})
		})

		return AndFilters(HasAttribute[*parser.StackTrace](attributes.EventStacktrace), methodFilter).GetPredicate(metadata)
	})
}

func AndFilters(filters ...parser.EventFilter) parser.EventFilter {
	return EventFilterFunc(func(metadata *parser.ClassMetadata) parser.Predicate[parser.Event] {
		predicates := make([]parser.Predicate[parser.Event], 0, len(filters))

		for _, filter := range filters {
			predicates = append(predicates, filter.GetPredicate(metadata))
		}

		return AndAlways(predicates...)
	})
}

func OrFilters(filters ...parser.EventFilter) parser.EventFilter {
	return EventFilterFunc(func(metadata *parser.ClassMetadata) parser.Predicate[parser.Event] {

		predicates := make([]parser.Predicate[parser.Event], 0, len(filters))

		for _, filter := range filters {
			predicates = append(predicates, filter.GetPredicate(metadata))
		}

		return OrAlways(predicates...)
	})
}

func NotFilter(filter parser.EventFilter) parser.EventFilter {
	return EventFilterFunc(func(metadata *parser.ClassMetadata) parser.Predicate[parser.Event] {
		return NotAlways(filter.GetPredicate(metadata))
	})
}

func NotNull[T any](attr *attributes.Attribute[T]) parser.EventFilter {
	return EventFilterFunc(func(metadata *parser.ClassMetadata) parser.Predicate[parser.Event] {
		if parser.IsAlwaysFalse(HasAttribute[T](attr).GetPredicate(metadata)) {
			return parser.AlwaysFalse
		}

		return parser.PredicateFunc(func(e parser.Event) bool {
			value, err := attr.GetValue(e.(*parser.GenericEvent))
			if err != nil {
				return false
			}
			rv := reflect.ValueOf(value)
			switch rv.Kind() {
			case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
				return rv.IsNil()
			default:
				return false
			}
		})
	})
}
