package parser

import (
	"fmt"

	"github.com/pyroscope-io/jfr-parser/reader"
)

var events = map[string]func() Parseable{
	"jdk.ActiveRecording":                      func() Parseable { return new(ActiveRecording) },
	"jdk.ActiveSetting":                        func() Parseable { return new(ActiveSetting) },
	"jdk.BooleanFlag":                          func() Parseable { return new(BooleanFlag) },
	"jdk.CPUInformation":                       func() Parseable { return new(CPUInformation) },
	"jdk.CPULoad":                              func() Parseable { return new(CPULoad) },
	"jdk.CPUTimeStampCounter":                  func() Parseable { return new(CPUTimeStampCounter) },
	"jdk.ClassLoaderStatistics":                func() Parseable { return new(ClassLoaderStatistics) },
	"jdk.ClassLoadingStatistics":               func() Parseable { return new(ClassLoadingStatistics) },
	"jdk.CodeCacheConfiguration":               func() Parseable { return new(CodeCacheConfiguration) },
	"jdk.CodeCacheStatistics":                  func() Parseable { return new(CodeCacheStatistics) },
	"jdk.CodeSweeperConfiguration":             func() Parseable { return new(CodeSweeperConfiguration) },
	"jdk.CodeSweeperStatistics":                func() Parseable { return new(CodeSweeperStatistics) },
	"jdk.CompilerConfiguration":                func() Parseable { return new(CompilerConfiguration) },
	"jdk.CompilerStatistics":                   func() Parseable { return new(CompilerStatistics) },
	"jdk.DoubleFlag":                           func() Parseable { return new(DoubleFlag) },
	"jdk.ExceptionStatistics":                  func() Parseable { return new(ExceptionStatistics) },
	"jdk.ExecutionSample":                      func() Parseable { return new(ExecutionSample) },
	"jdk.GCConfiguration":                      func() Parseable { return new(GCConfiguration) },
	"jdk.GCHeapConfiguration":                  func() Parseable { return new(GCHeapConfiguration) },
	"jdk.GCSurvivorConfiguration":              func() Parseable { return new(GCSurvivorConfiguration) },
	"jdk.GCTLABConfiguration":                  func() Parseable { return new(GCTLABConfiguration) },
	"jdk.InitialEnvironmentVariable":           func() Parseable { return new(InitialEnvironmentVariable) },
	"jdk.InitialSystemProperty":                func() Parseable { return new(InitialSystemProperty) },
	"jdk.IntFlag":                              func() Parseable { return new(IntFlag) },
	"jdk.JavaMonitorEnter":                     func() Parseable { return new(JavaMonitorEnter) },
	"jdk.JavaMonitorWait":                      func() Parseable { return new(JavaMonitorWait) },
	"jdk.JavaThreadStatistics":                 func() Parseable { return new(JavaThreadStatistics) },
	"jdk.JVMInformation":                       func() Parseable { return new(JVMInformation) },
	"jdk.LoaderConstraintsTableStatistics":     func() Parseable { return new(LoaderConstraintsTableStatistics) },
	"jdk.LongFlag":                             func() Parseable { return new(LongFlag) },
	"jdk.ModuleExport":                         func() Parseable { return new(ModuleExport) },
	"jdk.ModuleRequire":                        func() Parseable { return new(ModuleRequire) },
	"jdk.NativeLibrary":                        func() Parseable { return new(NativeLibrary) },
	"jdk.NetworkUtilization":                   func() Parseable { return new(NetworkUtilization) },
	"jdk.ObjectAllocationInNewTLAB":            func() Parseable { return new(ObjectAllocationInNewTLAB) },
	"jdk.ObjectAllocationOutsideTLAB":          func() Parseable { return new(ObjectAllocationOutsideTLAB) },
	"jdk.OSInformation":                        func() Parseable { return new(OSInformation) },
	"jdk.PhysicalMemory":                       func() Parseable { return new(PhysicalMemory) },
	"jdk.PlaceholderTableStatistics":           func() Parseable { return new(PlaceholderTableStatistics) },
	"jdk.ProtectionDomainCacheTableStatistics": func() Parseable { return new(ProtectionDomainCacheTableStatistics) },
	"jdk.StringFlag":                           func() Parseable { return new(StringFlag) },
	"jdk.StringTableStatistics":                func() Parseable { return new(StringTableStatistics) },
	"jdk.SymbolTableStatistics":                func() Parseable { return new(SymbolTableStatistics) },
	"jdk.SystemProcess":                        func() Parseable { return new(SystemProcess) },
	"jdk.ThreadAllocationStatistics":           func() Parseable { return new(ThreadAllocationStatistics) },
	"jdk.ThreadCPULoad":                        func() Parseable { return new(ThreadCPULoad) },
	"jdk.ThreadContextSwitchRate":              func() Parseable { return new(ThreadContextSwitchRate) },
	"jdk.ThreadDump":                           func() Parseable { return new(ThreadDump) },
	"jdk.ThreadPark":                           func() Parseable { return new(ThreadPark) },
	"jdk.ThreadStart":                          func() Parseable { return new(ThreadStart) },
	"jdk.UnsignedIntFlag":                      func() Parseable { return new(UnsignedIntFlag) },
	"jdk.UnsignedLongFlag":                     func() Parseable { return new(UnsignedLongFlag) },
	"jdk.VirtualizationInformation":            func() Parseable { return new(VirtualizationInformation) },
	"jdk.YoungGenerationConfiguration":         func() Parseable { return new(YoungGenerationConfiguration) },
}

func ParseEvent(r reader.Reader, classes ClassMap, cpools PoolMap) (Parseable, error) {
	kind, err := r.VarLong()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve event type: %w", err)
	}
	return parseEvent(r, classes, cpools, int(kind))
}

func parseEvent(r reader.Reader, classes ClassMap, cpools PoolMap, classID int) (Parseable, error) {
	class, ok := classes[classID]
	if !ok {
		return nil, fmt.Errorf("unknown class %d", classID)
	}
	var v Parseable
	if typeFn, ok := events[class.Name]; ok {
		v = typeFn()
	} else {
		v = new(UnsupportedEvent)
	}
	if err := v.Parse(r, classes, cpools, class); err != nil {
		return nil, fmt.Errorf("unable to parse event %s: %w", class.Name, err)
	}
	return v, nil
}

type ActiveRecording struct {
	StartTime         int64
	Duration          int64
	EventThread       *Thread
	ID                int64
	Name              string
	Destination       string
	MaxAge            int64
	MaxSize           int64
	RecordingStart    int64
	RecordingDuration int64
}

func (ar *ActiveRecording) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		ar.StartTime, err = toLong(p)
	case "duration":
		ar.Duration, err = toLong(p)
	case "eventThread":
		ar.EventThread, err = toThread(p)
	case "id":
		ar.ID, err = toLong(p)
	case "name":
		ar.Name, err = toString(p)
	case "destination":
		ar.Destination, err = toString(p)
	case "maxAge":
		ar.MaxAge, err = toLong(p)
	case "maxSize":
		ar.MaxSize, err = toLong(p)
	case "recordingStart":
		ar.RecordingStart, err = toLong(p)
	case "recordingDuration":
		ar.RecordingDuration, err = toLong(p)
	}
	return err
}

func (ar *ActiveRecording) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, ar.parseField)
}

type ActiveSetting struct {
	StartTime   int64
	Duration    int64
	EventThread *Thread
	ID          int64
	Name        string
	Value       string
}

func (as *ActiveSetting) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		as.StartTime, err = toLong(p)
	case "duration":
		as.Duration, err = toLong(p)
	case "eventThread":
		as.EventThread, err = toThread(p)
	case "id":
		as.ID, err = toLong(p)
	case "name":
		as.Name, err = toString(p)
	case "value":
		as.Value, err = toString(p)
	}
	return err
}

func (as *ActiveSetting) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, as.parseField)
}

type BooleanFlag struct {
	StartTime int64
	Name      string
	Value     bool
	Origin    *FlagValueOrigin
}

func (bf *BooleanFlag) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		bf.StartTime, err = toLong(p)
	case "name":
		bf.Name, err = toString(p)
	case "value":
		bf.Value, err = toBoolean(p)
	case "origin":
		bf.Origin, err = toFlagValueOrigin(p)
	}
	return err
}

func (bf *BooleanFlag) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, bf.parseField)
}

type CPUInformation struct {
	StartTime   int64
	CPU         string
	Description string
	Sockets     int32
	Cores       int32
	HWThreads   int32
}

func (ci *CPUInformation) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		ci.StartTime, err = toLong(p)
	case "duration":
		ci.CPU, err = toString(p)
	case "eventThread":
		ci.Description, err = toString(p)
	case "sockets":
		ci.Sockets, err = toInt(p)
	case "cores":
		ci.Cores, err = toInt(p)
	case "hwThreads":
		ci.HWThreads, err = toInt(p)
	}
	return err
}

func (ci *CPUInformation) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, ci.parseField)
}

type CPULoad struct {
	StartTime    int64
	JVMUser      float32
	JVMSystem    float32
	MachineTotal float32
}

func (cl *CPULoad) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		cl.StartTime, err = toLong(p)
	case "jvmUser":
		cl.JVMUser, err = toFloat(p)
	case "jvmSystem":
		cl.JVMSystem, err = toFloat(p)
	case "machineTotal":
		cl.MachineTotal, err = toFloat(p)
	}
	return err
}

func (cl *CPULoad) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, cl.parseField)
}

type CPUTimeStampCounter struct {
	StartTime           int64
	FastTimeEnabled     bool
	FastTimeAutoEnabled bool
	OSFrequency         int64
	FastTimeFrequency   int64
}

func (ctsc *CPUTimeStampCounter) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		ctsc.StartTime, err = toLong(p)
	case "fastTimeEnabled":
		ctsc.FastTimeEnabled, err = toBoolean(p)
	case "fastTimeAutoEnabled":
		ctsc.FastTimeAutoEnabled, err = toBoolean(p)
	case "osFrequency":
		ctsc.OSFrequency, err = toLong(p)
	case "fastTimeFrequency":
		ctsc.FastTimeFrequency, err = toLong(p)
	}
	return err
}

func (ctsc *CPUTimeStampCounter) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, ctsc.parseField)
}

type ClassLoaderStatistics struct {
	StartTime                 int64
	ClassLoader               *ClassLoader
	ParentClassLoader         *ClassLoader
	ClassLoaderData           int64
	ClassCount                int64
	ChunkSize                 int64
	BlockSize                 int64
	AnonymousClassCount       int64
	AnonymousChunkSize        int64
	AnonymousBlockSize        int64
	UnsafeAnonymousClassCount int64
	UnsafeAnonymousChunkSize  int64
	UnsafeAnonymousBlockSize  int64
	HiddenClassCount          int64
	HiddenChunkSize           int64
	HiddenBlockSize           int64
}

func (cls *ClassLoaderStatistics) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		cls.StartTime, err = toLong(p)
	case "classLoader":
		cls.ClassLoader, err = toClassLoader(p)
	case "parentClassLoader":
		cls.ParentClassLoader, err = toClassLoader(p)
	case "classLoaderData":
		cls.ClassLoaderData, err = toLong(p)
	case "classCount":
		cls.ClassCount, err = toLong(p)
	case "chunkSize":
		cls.ChunkSize, err = toLong(p)
	case "blockSize":
		cls.BlockSize, err = toLong(p)
	case "anonymousClassCount":
		cls.AnonymousClassCount, err = toLong(p)
	case "anonymousChunkSize":
		cls.AnonymousChunkSize, err = toLong(p)
	case "anonymousBlockSize":
		cls.AnonymousBlockSize, err = toLong(p)
	case "unsafeAnonymousClassCount":
		cls.UnsafeAnonymousClassCount, err = toLong(p)
	case "unsafeAnonymousChunkSize":
		cls.UnsafeAnonymousChunkSize, err = toLong(p)
	case "unsafeAnonymousBlockSize":
		cls.UnsafeAnonymousBlockSize, err = toLong(p)
	case "hiddenClassCount":
		cls.HiddenClassCount, err = toLong(p)
	case "hiddenChunkSize":
		cls.HiddenChunkSize, err = toLong(p)
	case "hiddenBlockSize":
		cls.HiddenBlockSize, err = toLong(p)
	}
	return err
}

func (cls *ClassLoaderStatistics) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, cls.parseField)
}

type ClassLoadingStatistics struct {
	StartTime          int64
	LoadedClassCount   int64
	UnloadedClassCount int64
}

func (cls *ClassLoadingStatistics) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		cls.StartTime, err = toLong(p)
	case "loadedClassCount":
		cls.LoadedClassCount, err = toLong(p)
	case "unloadedClassCount":
		cls.UnloadedClassCount, err = toLong(p)
	}
	return err
}

func (cls *ClassLoadingStatistics) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, cls.parseField)
}

type CodeCacheConfiguration struct {
	StartTime          int64
	InitialSize        int64
	ReservedSize       int64
	NonNMethodSize     int64
	ProfiledSize       int64
	NonProfiledSize    int64
	ExpansionSize      int64
	MinBlockLength     int64
	StartAddress       int64
	ReservedTopAddress int64
}

func (ccc *CodeCacheConfiguration) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		ccc.StartTime, err = toLong(p)
	case "initialSize":
		ccc.InitialSize, err = toLong(p)
	case "reservedSize":
		ccc.ReservedSize, err = toLong(p)
	case "nonNMethodSize":
		ccc.NonNMethodSize, err = toLong(p)
	case "profiledSize":
		ccc.ProfiledSize, err = toLong(p)
	case "NonProfiledSize":
		ccc.NonProfiledSize, err = toLong(p)
	case "ExpansionSize":
		ccc.ExpansionSize, err = toLong(p)
	case "MinBlockLength":
		ccc.MinBlockLength, err = toLong(p)
	case "StartAddress":
		ccc.StartAddress, err = toLong(p)
	case "ReservedTopAddress":
		ccc.ReservedTopAddress, err = toLong(p)
	}
	return err
}

func (ccc *CodeCacheConfiguration) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, ccc.parseField)
}

type CodeCacheStatistics struct {
	StartTime           int64
	CodeBlobType        *CodeBlobType
	StartAddress        int64
	ReservedTopAddress  int64
	EntryCount          int32
	MethodCount         int32
	AdaptorCount        int32
	UnallocatedCapacity int64
	FullCount           int32
}

func (ccs *CodeCacheStatistics) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		ccs.StartTime, err = toLong(p)
	case "codeBlobType":
		ccs.CodeBlobType, err = toCodeBlobType(p)
	case "startAddress":
		ccs.StartAddress, err = toLong(p)
	case "reservedTopAddress":
		ccs.ReservedTopAddress, err = toLong(p)
	case "entryCount":
		ccs.EntryCount, err = toInt(p)
	case "methodCount":
		ccs.MethodCount, err = toInt(p)
	case "adaptorCount":
		ccs.AdaptorCount, err = toInt(p)
	case "unallocatedCapacity":
		ccs.UnallocatedCapacity, err = toLong(p)
	case "fullCount":
		ccs.FullCount, err = toInt(p)
	}
	return err
}

func (ccs *CodeCacheStatistics) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, ccs.parseField)
}

type CodeSweeperConfiguration struct {
	StartTime       int64
	SweeperEnabled  bool
	FlushingEnabled bool
	SweepThreshold  int64
}

func (csc *CodeSweeperConfiguration) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		csc.StartTime, err = toLong(p)
	case "sweeperEnabled":
		csc.SweeperEnabled, err = toBoolean(p)
	case "flushingEnabled":
		csc.FlushingEnabled, err = toBoolean(p)
	case "sweepThreshold":
		csc.SweepThreshold, err = toLong(p)
	}
	return err
}

func (csc *CodeSweeperConfiguration) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, csc.parseField)
}

type CodeSweeperStatistics struct {
	StartTime            int64
	SweepCount           int32
	MethodReclaimedCount int32
	TotalSweepTime       int64
	PeakFractionTime     int64
	PeakSweepTime        int64
}

func (css *CodeSweeperStatistics) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		css.StartTime, err = toLong(p)
	case "sweepCount":
		css.SweepCount, err = toInt(p)
	case "methodReclaimedCount":
		css.MethodReclaimedCount, err = toInt(p)
	case "totalSweepTime":
		css.TotalSweepTime, err = toLong(p)
	case "peakFractionTime":
		css.PeakFractionTime, err = toLong(p)
	case "peakSweepTime":
		css.PeakSweepTime, err = toLong(p)
	}
	return err
}

func (css *CodeSweeperStatistics) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, css.parseField)
}

type CompilerConfiguration struct {
	StartTime         int64
	ThreadCount       int32
	TieredCompilation bool
}

func (cc *CompilerConfiguration) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		cc.StartTime, err = toLong(p)
	case "threadCount":
		cc.ThreadCount, err = toInt(p)
	case "tieredCompilation":
		cc.TieredCompilation, err = toBoolean(p)
	}
	return err
}

func (cc *CompilerConfiguration) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, cc.parseField)
}

type CompilerStatistics struct {
	StartTime             int64
	CompileCount          int32
	BailoutCount          int32
	InvalidatedCount      int32
	OSRCompileCount       int32
	StandardCompileCount  int32
	OSRBytesCompiled      int64
	StandardBytesCompiled int64
	NMethodsSize          int64
	NMethodCodeSize       int64
	PeakTimeSpent         int64
	TotalTimeSpent        int64
}

func (cs *CompilerStatistics) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		cs.StartTime, err = toLong(p)
	case "compileCount":
		cs.CompileCount, err = toInt(p)
	case "bailoutCount":
		cs.BailoutCount, err = toInt(p)
	case "invalidatedCount":
		cs.InvalidatedCount, err = toInt(p)
	case "osrCompileCount":
		cs.OSRCompileCount, err = toInt(p)
	case "standardCompileCount":
		cs.StandardCompileCount, err = toInt(p)
	case "osrBytesCompiled":
		cs.OSRBytesCompiled, err = toLong(p)
	case "standardBytesCompiled":
		cs.StandardBytesCompiled, err = toLong(p)
	case "nmethodsSize":
		cs.NMethodsSize, err = toLong(p)
	case "nmethodCodeSize":
		cs.NMethodCodeSize, err = toLong(p)
	case "peakTimeSpent":
		cs.PeakTimeSpent, err = toLong(p)
	case "totalTimeSpent":
		cs.TotalTimeSpent, err = toLong(p)
	}
	return err
}

func (cs *CompilerStatistics) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, cs.parseField)
}

type DoubleFlag struct {
	StartTime int64
	Name      string
	Value     float64
	Origin    *FlagValueOrigin
}

func (df *DoubleFlag) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		df.StartTime, err = toLong(p)
	case "name":
		df.Name, err = toString(p)
	case "value":
		df.Value, err = toDouble(p)
	case "origin":
		df.Origin, err = toFlagValueOrigin(p)
	}
	return err
}

func (df *DoubleFlag) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, df.parseField)
}

type ExceptionStatistics struct {
	StartTime   int64
	Duration    int64
	EventThread *Thread
	StackTrace  *StackTrace
	Throwable   int64
}

func (es *ExceptionStatistics) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		es.StartTime, err = toLong(p)
	case "duration":
		es.Duration, err = toLong(p)
	case "eventThread":
		es.EventThread, err = toThread(p)
	case "stackTrace":
		es.StackTrace, err = toStackTrace(p)
	case "throwable":
		es.Throwable, err = toLong(p)
	}
	return err
}

func (es *ExceptionStatistics) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, es.parseField)
}

type ExecutionSample struct {
	StartTime     int64
	SampledThread *Thread
	StackTrace    *StackTrace
	State         *ThreadState
	ContextId     int64
}

func (es *ExecutionSample) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		es.StartTime, err = toLong(p)
	case "sampledThread":
		es.SampledThread, err = toThread(p)
	case "stackTrace":
		es.StackTrace, err = toStackTrace(p)
	case "state":
		es.State, err = toThreadState(p)
	case "contextId":
		es.ContextId, err = toLong(p)
	}
	return err
}

func (es *ExecutionSample) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, es.parseField)
}

type GCConfiguration struct {
	StartTime              int64
	YoungCollector         *GCName
	OldCollector           *GCName
	ParallelGCThreads      int32
	ConcurrentGCThreads    int32
	UsesDynamicGCThreads   bool
	IsExplicitGCConcurrent bool
	IsExplicitGCDisabled   bool
	PauseTarget            int64
	GCTimeRatio            int32
}

func (gc *GCConfiguration) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		gc.StartTime, err = toLong(p)
	case "youngCollector":
		gc.YoungCollector, err = toGCName(p)
	case "oldCollector":
		gc.OldCollector, err = toGCName(p)
	case "parallelGCThreads":
		gc.ParallelGCThreads, err = toInt(p)
	case "concurrentGCThreads":
		gc.ConcurrentGCThreads, err = toInt(p)
	case "usesDynamicGCThreads":
		gc.UsesDynamicGCThreads, err = toBoolean(p)
	case "isExplicitGCConcurrent":
		gc.IsExplicitGCConcurrent, err = toBoolean(p)
	case "isExplicitGCDisabled":
		gc.IsExplicitGCDisabled, err = toBoolean(p)
	case "pauseTarget":
		gc.PauseTarget, err = toLong(p)
	case "gcTimeRatio":
		gc.GCTimeRatio, err = toInt(p)
	}
	return err
}

func (gc *GCConfiguration) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, gc.parseField)
}

type GCHeapConfiguration struct {
	StartTime          int64
	MinSize            int64
	MaxSize            int64
	InitialSize        int64
	UsesCompressedOops bool
	CompressedOopsMode *NarrowOopMode
	ObjectAlignment    int64
	HeapAddressBits    int8
}

func (ghc *GCHeapConfiguration) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		ghc.StartTime, err = toLong(p)
	case "minSize":
		ghc.MinSize, err = toLong(p)
	case "maxSize":
		ghc.MaxSize, err = toLong(p)
	case "initialSize":
		ghc.InitialSize, err = toLong(p)
	case "usesCompressedOops":
		ghc.UsesCompressedOops, err = toBoolean(p)
	case "compressedOopsMode":
		ghc.CompressedOopsMode, err = toNarrowOopMode(p)
	case "objectAlignment":
		ghc.ObjectAlignment, err = toLong(p)
	case "heapAddressBits":
		ghc.HeapAddressBits, err = toByte(p)
	}
	return err
}

func (ghc *GCHeapConfiguration) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, ghc.parseField)
}

type GCSurvivorConfiguration struct {
	StartTime                int64
	MaxTenuringThreshold     int8
	InitialTenuringThreshold int8
}

func (gcs *GCSurvivorConfiguration) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		gcs.StartTime, err = toLong(p)
	case "maxTenuringThreshold":
		gcs.MaxTenuringThreshold, err = toByte(p)
	case "initialTenuringThreshold":
		gcs.InitialTenuringThreshold, err = toByte(p)
	}
	return err
}

func (gsc *GCSurvivorConfiguration) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, gsc.parseField)
}

type GCTLABConfiguration struct {
	StartTime            int64
	UsesTLABs            bool
	MinTLABSize          int64
	TLABRefillWasteLimit int64
}

func (gtc *GCTLABConfiguration) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		gtc.StartTime, err = toLong(p)
	case "usesTLABs":
		gtc.UsesTLABs, err = toBoolean(p)
	case "minTLABSize":
		gtc.MinTLABSize, err = toLong(p)
	case "tlabRefillWasteLimit":
		gtc.TLABRefillWasteLimit, err = toLong(p)
	}
	return err
}

func (gtc *GCTLABConfiguration) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, gtc.parseField)
}

type InitialEnvironmentVariable struct {
	StartTime int64
	Key       string
	Value     string
}

func (iev *InitialEnvironmentVariable) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		iev.StartTime, err = toLong(p)
	case "key":
		iev.Key, err = toString(p)
	case "value":
		iev.Value, err = toString(p)
	}
	return err
}

func (iev *InitialEnvironmentVariable) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, iev.parseField)
}

type InitialSystemProperty struct {
	StartTime int64
	Key       string
	Value     string
}

func (isp *InitialSystemProperty) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		isp.StartTime, err = toLong(p)
	case "key":
		isp.Key, err = toString(p)
	case "value":
		isp.Value, err = toString(p)
	}
	return err
}

func (isp *InitialSystemProperty) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, isp.parseField)
}

type IntFlag struct {
	StartTime int64
	Name      string
	Value     int32
	Origin    *FlagValueOrigin
}

func (f *IntFlag) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		f.StartTime, err = toLong(p)
	case "name":
		f.Name, err = toString(p)
	case "value":
		f.Value, err = toInt(p)
	case "origin":
		f.Origin, err = toFlagValueOrigin(p)
	}
	return err
}

func (f *IntFlag) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, f.parseField)
}

type JavaMonitorEnter struct {
	StartTime     int64
	Duration      int64
	EventThread   *Thread
	StackTrace    *StackTrace
	MonitorClass  *Class
	PreviousOwner *Thread
	Address       int64
	ContextId     int64
}

func (jme *JavaMonitorEnter) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		jme.StartTime, err = toLong(p)
	case "duration":
		jme.Duration, err = toLong(p)
	case "eventThread":
		jme.EventThread, err = toThread(p)
	case "stackTrace":
		jme.StackTrace, err = toStackTrace(p)
	case "monitorClass":
		jme.MonitorClass, err = toClass(p)
	case "previousOwner":
		jme.PreviousOwner, err = toThread(p)
	case "address":
		jme.Address, err = toLong(p)
	case "contextId":
		jme.ContextId, err = toLong(p)
	}
	return err
}

func (jme *JavaMonitorEnter) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, jme.parseField)
}

type JavaMonitorWait struct {
	StartTime    int64
	Duration     int64
	EventThread  *Thread
	StackTrace   *StackTrace
	MonitorClass *Class
	Notifier     *Thread
	Timeout      int64
	TimedOut     bool
	Address      int64
}

func (jmw *JavaMonitorWait) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		jmw.StartTime, err = toLong(p)
	case "duration":
		jmw.Duration, err = toLong(p)
	case "eventThread":
		jmw.EventThread, err = toThread(p)
	case "stackTrace":
		jmw.StackTrace, err = toStackTrace(p)
	case "monitorClass":
		jmw.MonitorClass, err = toClass(p)
	case "notifier":
		jmw.Notifier, err = toThread(p)
	case "timeout":
		jmw.Timeout, err = toLong(p)
	case "timedOut":
		jmw.TimedOut, err = toBoolean(p)
	case "address":
		jmw.Address, err = toLong(p)
	}
	return err
}

func (jmw *JavaMonitorWait) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, jmw.parseField)
}

type JavaThreadStatistics struct {
	StartTime        int64
	ActiveCount      int64
	DaemonCount      int64
	AccumulatedCount int64
	PeakCount        int64
}

func (jts *JavaThreadStatistics) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		jts.StartTime, err = toLong(p)
	case "activeCount":
		jts.ActiveCount, err = toLong(p)
	case "daemonCount":
		jts.DaemonCount, err = toLong(p)
	case "accumulatedCount":
		jts.AccumulatedCount, err = toLong(p)
	case "peakCount":
		jts.PeakCount, err = toLong(p)
	}
	return err
}

func (jts *JavaThreadStatistics) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, jts.parseField)
}

type JVMInformation struct {
	StartTime     int64
	JVMName       string
	JVMVersion    string
	JVMArguments  string
	JVMFlags      string
	JavaArguments string
	JVMStartTime  int64
	PID           int64
}

func (ji *JVMInformation) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		ji.StartTime, err = toLong(p)
	case "jvmName":
		ji.JVMName, err = toString(p)
	case "jvmVersion":
		ji.JVMVersion, err = toString(p)
	case "jvmArguments":
		ji.JVMArguments, err = toString(p)
	case "jvmFlags":
		ji.JVMFlags, err = toString(p)
	case "javaArguments":
		ji.JavaArguments, err = toString(p)
	case "jvmStartTime":
		ji.JVMStartTime, err = toLong(p)
	case "pid":
		ji.PID, err = toLong(p)
	}
	return err
}

func (ji *JVMInformation) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, ji.parseField)
}

type LoaderConstraintsTableStatistics struct {
	StartTime                    int64
	BucketCount                  int64
	EntryCount                   int64
	TotalFootprint               int64
	BucketCountMaximum           int64
	BucketCountAverage           float32
	BucketCountVariance          float32
	BucketCountStandardDeviation float32
	InsertionRate                float32
	RemovalRate                  float32
}

func (lcts *LoaderConstraintsTableStatistics) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		lcts.StartTime, err = toLong(p)
	case "bucketCount":
		lcts.BucketCount, err = toLong(p)
	case "entryCount":
		lcts.EntryCount, err = toLong(p)
	case "totalFootprint":
		lcts.TotalFootprint, err = toLong(p)
	case "bucketCountMaximum":
		lcts.BucketCountMaximum, err = toLong(p)
	case "bucketCountAverage":
		lcts.BucketCountAverage, err = toFloat(p)
	case "bucketCountVariance":
		lcts.BucketCountVariance, err = toFloat(p)
	case "bucketCountStandardDeviation":
		lcts.BucketCountStandardDeviation, err = toFloat(p)
	case "insertionRate":
		lcts.InsertionRate, err = toFloat(p)
	case "removalRate":
		lcts.RemovalRate, err = toFloat(p)
	}
	return err
}

func (lcts *LoaderConstraintsTableStatistics) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, lcts.parseField)
}

type LongFlag struct {
	StartTime int64
	Name      string
	Value     int64
	Origin    *FlagValueOrigin
}

func (lf *LongFlag) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		lf.StartTime, err = toLong(p)
	case "name":
		lf.Name, err = toString(p)
	case "value":
		lf.Value, err = toLong(p)
	case "origin":
		lf.Origin, err = toFlagValueOrigin(p)
	}
	return err
}

func (lf *LongFlag) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, lf.parseField)
}

type ModuleExport struct {
	StartTime       int64
	ExportedPackage *Package
	TargetModule    *Module
}

func (me *ModuleExport) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		me.StartTime, err = toLong(p)
	case "exportedPackage":
		me.ExportedPackage, err = toPackage(p)
	case "targetModule":
		me.TargetModule, err = toModule(p)
	}
	return err
}

func (me *ModuleExport) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, me.parseField)
}

type ModuleRequire struct {
	StartTime      int64
	Source         *Module
	RequiredModule *Module
}

func (mr *ModuleRequire) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		mr.StartTime, err = toLong(p)
	case "sourced":
		mr.Source, err = toModule(p)
	case "requiredModule":
		mr.RequiredModule, err = toModule(p)
	}
	return err
}

func (mr *ModuleRequire) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, mr.parseField)
}

type NativeLibrary struct {
	StartTime   int64
	Name        string
	BaseAddress int64
	TopAddress  int64
}

func (nl *NativeLibrary) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		nl.StartTime, err = toLong(p)
	case "name":
		nl.Name, err = toString(p)
	case "baseAddress":
		nl.BaseAddress, err = toLong(p)
	case "topAddress":
		nl.TopAddress, err = toLong(p)
	}
	return err
}

func (nl *NativeLibrary) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, nl.parseField)
}

type NetworkUtilization struct {
	StartTime        int64
	NetworkInterface *NetworkInterfaceName
	ReadRate         int64
	WriteRate        int64
}

func (nu *NetworkUtilization) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		nu.StartTime, err = toLong(p)
	case "networkInterface":
		nu.NetworkInterface, err = toNetworkInterfaceName(p)
	case "readRate":
		nu.ReadRate, err = toLong(p)
	case "writeRate":
		nu.WriteRate, err = toLong(p)
	}
	return err
}

func (nu *NetworkUtilization) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, nu.parseField)
}

type ObjectAllocationInNewTLAB struct {
	StartTime      int64
	EventThread    *Thread
	StackTrace     *StackTrace
	ObjectClass    *Class
	AllocationSize int64
	TLABSize       int64
	ContextId      int64
}

func (oa *ObjectAllocationInNewTLAB) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		oa.StartTime, err = toLong(p)
	case "sampledThread":
		oa.EventThread, err = toThread(p)
	case "stackTrace":
		oa.StackTrace, err = toStackTrace(p)
	case "objectClass":
		oa.ObjectClass, err = toClass(p)
	case "allocationSize":
		oa.AllocationSize, err = toLong(p)
	case "tlabSize":
		oa.TLABSize, err = toLong(p)
	case "contextId":
		oa.ContextId, err = toLong(p)
	}

	return err
}

func (oa *ObjectAllocationInNewTLAB) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, oa.parseField)
}

type ObjectAllocationOutsideTLAB struct {
	StartTime      int64
	EventThread    *Thread
	StackTrace     *StackTrace
	ObjectClass    *Class
	AllocationSize int64
	ContextId      int64
}

func (oa *ObjectAllocationOutsideTLAB) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		oa.StartTime, err = toLong(p)
	case "sampledThread":
		oa.EventThread, err = toThread(p)
	case "stackTrace":
		oa.StackTrace, err = toStackTrace(p)
	case "objectClass":
		oa.ObjectClass, err = toClass(p)
	case "allocationSize":
		oa.AllocationSize, err = toLong(p)
	case "contextId":
		oa.ContextId, err = toLong(p)
	}
	return err
}

func (oa *ObjectAllocationOutsideTLAB) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, oa.parseField)
}

type OSInformation struct {
	StartTime int64
	OSVersion string
}

func (os *OSInformation) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		os.StartTime, err = toLong(p)
	case "osVersion":
		os.OSVersion, err = toString(p)
	}
	return err
}

func (os *OSInformation) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, os.parseField)
}

type PhysicalMemory struct {
	StartTime int64
	TotalSize int64
	UsedSize  int64
}

func (pm *PhysicalMemory) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		pm.StartTime, err = toLong(p)
	case "totalSize":
		pm.TotalSize, err = toLong(p)
	case "usedSize":
		pm.UsedSize, err = toLong(p)
	}
	return err
}

func (pm *PhysicalMemory) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, pm.parseField)
}

type PlaceholderTableStatistics struct {
	StartTime                    int64
	BucketCount                  int64
	EntryCount                   int64
	TotalFootprint               int64
	BucketCountMaximum           int64
	BucketCountAverage           float32
	BucketCountVariance          float32
	BucketCountStandardDeviation float32
	InsertionRate                float32
	RemovalRate                  float32
}

func (pts *PlaceholderTableStatistics) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		pts.StartTime, err = toLong(p)
	case "bucketCount":
		pts.BucketCount, err = toLong(p)
	case "entryCount":
		pts.EntryCount, err = toLong(p)
	case "totalFootprint":
		pts.TotalFootprint, err = toLong(p)
	case "bucketCountMaximum":
		pts.BucketCountMaximum, err = toLong(p)
	case "bucketCountAverage":
		pts.BucketCountAverage, err = toFloat(p)
	case "bucketCountVariance":
		pts.BucketCountVariance, err = toFloat(p)
	case "bucketCountStandardDeviation":
		pts.BucketCountStandardDeviation, err = toFloat(p)
	case "insertionRate":
		pts.InsertionRate, err = toFloat(p)
	case "removalRate":
		pts.RemovalRate, err = toFloat(p)
	}
	return err
}

func (pts *PlaceholderTableStatistics) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, pts.parseField)
}

type ProtectionDomainCacheTableStatistics struct {
	StartTime                    int64
	BucketCount                  int64
	EntryCount                   int64
	TotalFootprint               int64
	BucketCountMaximum           int64
	BucketCountAverage           float32
	BucketCountVariance          float32
	BucketCountStandardDeviation float32
	InsertionRate                float32
	RemovalRate                  float32
}

func (pdcts *ProtectionDomainCacheTableStatistics) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		pdcts.StartTime, err = toLong(p)
	case "bucketCount":
		pdcts.BucketCount, err = toLong(p)
	case "entryCount":
		pdcts.EntryCount, err = toLong(p)
	case "totalFootprint":
		pdcts.TotalFootprint, err = toLong(p)
	case "bucketCountMaximum":
		pdcts.BucketCountMaximum, err = toLong(p)
	case "bucketCountAverage":
		pdcts.BucketCountAverage, err = toFloat(p)
	case "bucketCountVariance":
		pdcts.BucketCountVariance, err = toFloat(p)
	case "bucketCountStandardDeviation":
		pdcts.BucketCountStandardDeviation, err = toFloat(p)
	case "insertionRate":
		pdcts.InsertionRate, err = toFloat(p)
	case "removalRate":
		pdcts.RemovalRate, err = toFloat(p)
	}
	return err
}

func (pdcts *ProtectionDomainCacheTableStatistics) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, pdcts.parseField)
}

type StringFlag struct {
	StartTime int64
	Name      string
	Value     string
	Origin    *FlagValueOrigin
}

func (sf *StringFlag) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		sf.StartTime, err = toLong(p)
	case "name":
		sf.Name, err = toString(p)
	case "value":
		sf.Value, err = toString(p)
	case "origin":
		sf.Origin, err = toFlagValueOrigin(p)
	}
	return err
}

func (sf *StringFlag) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, sf.parseField)
}

type StringTableStatistics struct {
	StartTime                    int64
	BucketCount                  int64
	EntryCount                   int64
	TotalFootprint               int64
	BucketCountMaximum           int64
	BucketCountAverage           float32
	BucketCountVariance          float32
	BucketCountStandardDeviation float32
	InsertionRate                float32
	RemovalRate                  float32
}

func (sts *StringTableStatistics) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		sts.StartTime, err = toLong(p)
	case "bucketCount":
		sts.BucketCount, err = toLong(p)
	case "entryCount":
		sts.EntryCount, err = toLong(p)
	case "totalFootprint":
		sts.TotalFootprint, err = toLong(p)
	case "bucketCountMaximum":
		sts.BucketCountMaximum, err = toLong(p)
	case "bucketCountAverage":
		sts.BucketCountAverage, err = toFloat(p)
	case "bucketCountVariance":
		sts.BucketCountVariance, err = toFloat(p)
	case "bucketCountStandardDeviation":
		sts.BucketCountStandardDeviation, err = toFloat(p)
	case "insertionRate":
		sts.InsertionRate, err = toFloat(p)
	case "removalRate":
		sts.RemovalRate, err = toFloat(p)
	}
	return err
}

func (sts *StringTableStatistics) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, sts.parseField)
}

type SymbolTableStatistics struct {
	StartTime                    int64
	BucketCount                  int64
	EntryCount                   int64
	TotalFootprint               int64
	BucketCountMaximum           int64
	BucketCountAverage           float32
	BucketCountVariance          float32
	BucketCountStandardDeviation float32
	InsertionRate                float32
	RemovalRate                  float32
}

func (sts *SymbolTableStatistics) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		sts.StartTime, err = toLong(p)
	case "bucketCount":
		sts.BucketCount, err = toLong(p)
	case "entryCount":
		sts.EntryCount, err = toLong(p)
	case "totalFootprint":
		sts.TotalFootprint, err = toLong(p)
	case "bucketCountMaximum":
		sts.BucketCountMaximum, err = toLong(p)
	case "bucketCountAverage":
		sts.BucketCountAverage, err = toFloat(p)
	case "bucketCountVariance":
		sts.BucketCountVariance, err = toFloat(p)
	case "bucketCountStandardDeviation":
		sts.BucketCountStandardDeviation, err = toFloat(p)
	case "insertionRate":
		sts.InsertionRate, err = toFloat(p)
	case "removalRate":
		sts.RemovalRate, err = toFloat(p)
	}
	return err
}

func (sts *SymbolTableStatistics) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, sts.parseField)
}

type SystemProcess struct {
	StartTime   int64
	PID         string
	CommandLine string
}

func (sp *SystemProcess) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		sp.StartTime, err = toLong(p)
	case "pid":
		sp.PID, err = toString(p)
	case "commandLine":
		sp.CommandLine, err = toString(p)
	}
	return err
}

func (sp *SystemProcess) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, sp.parseField)
}

type ThreadAllocationStatistics struct {
	StartTime int64
	Allocated int64
	Thread    *Thread
}

func (tas *ThreadAllocationStatistics) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		tas.StartTime, err = toLong(p)
	case "allocated":
		tas.Allocated, err = toLong(p)
	case "thread":
		tas.Thread, err = toThread(p)
	}
	return err
}

func (tas *ThreadAllocationStatistics) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, tas.parseField)
}

type ThreadCPULoad struct {
	StartTime   int64
	EventThread *Thread
	User        float32
	System      float32
}

func (tcl *ThreadCPULoad) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		tcl.StartTime, err = toLong(p)
	case "eventThread":
		tcl.EventThread, err = toThread(p)
	case "user":
		tcl.User, err = toFloat(p)
	case "system":
		tcl.System, err = toFloat(p)
	}
	return err
}

func (tcl *ThreadCPULoad) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, tcl.parseField)
}

type ThreadContextSwitchRate struct {
	StartTime  int64
	SwitchRate float32
}

func (tcsr *ThreadContextSwitchRate) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		tcsr.StartTime, err = toLong(p)
	case "switchRate":
		tcsr.SwitchRate, err = toFloat(p)
	}
	return err
}

func (tcsr *ThreadContextSwitchRate) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, tcsr.parseField)
}

type ThreadDump struct {
	StartTime int64
	Result    string
}

func (td *ThreadDump) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		td.StartTime, err = toLong(p)
	case "result":
		td.Result, err = toString(p)
	}
	return err
}

func (td *ThreadDump) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, td.parseField)
}

type ThreadPark struct {
	StartTime   int64
	Duration    int64
	EventThread *Thread
	StackTrace  *StackTrace
	ParkedClass *Class
	Timeout     int64
	Until       int64
	Address     int64
	ContextId   int64
}

func (tp *ThreadPark) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		tp.StartTime, err = toLong(p)
	case "duration":
		tp.Duration, err = toLong(p)
	case "eventThread":
		tp.EventThread, err = toThread(p)
	case "stackTrace":
		tp.StackTrace, err = toStackTrace(p)
	case "parkedClass":
		tp.ParkedClass, err = toClass(p)
	case "timeout":
		tp.Timeout, err = toLong(p)
	case "until":
		tp.Until, err = toLong(p)
	case "address":
		tp.Address, err = toLong(p)
	case "contextId": // todo this one seems to be unimplemented in the profiler yet
		tp.ContextId, err = toLong(p)
	}
	return err
}

func (tp *ThreadPark) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, tp.parseField)
}

type ThreadStart struct {
	StartTime    int64
	EventThread  *Thread
	StackTrace   *StackTrace
	Thread       *Thread
	ParentThread *Thread
}

func (ts *ThreadStart) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		ts.StartTime, err = toLong(p)
	case "eventThread":
		ts.EventThread, err = toThread(p)
	case "stackTrace":
		ts.StackTrace, err = toStackTrace(p)
	case "thread":
		ts.Thread, err = toThread(p)
	case "parentThread":
		ts.ParentThread, err = toThread(p)
	}
	return err
}

func (ts *ThreadStart) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, ts.parseField)
}

type UnsignedIntFlag struct {
	StartTime int64
	Name      string
	Value     int32
	Origin    *FlagValueOrigin
}

func (uif *UnsignedIntFlag) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		uif.StartTime, err = toLong(p)
	case "name":
		uif.Name, err = toString(p)
	case "value":
		uif.Value, err = toInt(p)
	case "origin":
		uif.Origin, err = toFlagValueOrigin(p)
	}
	return err
}

func (uif *UnsignedIntFlag) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, uif.parseField)
}

type UnsignedLongFlag struct {
	StartTime int64
	Name      string
	Value     int64
	Origin    *FlagValueOrigin
}

func (ulf *UnsignedLongFlag) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		ulf.StartTime, err = toLong(p)
	case "name":
		ulf.Name, err = toString(p)
	case "value":
		ulf.Value, err = toLong(p)
	case "origin":
		ulf.Origin, err = toFlagValueOrigin(p)
	}
	return err
}

func (ulf *UnsignedLongFlag) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, ulf.parseField)
}

type VirtualizationInformation struct {
	StartTime int64
	Name      string
}

func (vi *VirtualizationInformation) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		vi.StartTime, err = toLong(p)
	case "name":
		vi.Name, err = toString(p)
	}
	return err
}

func (vi *VirtualizationInformation) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, vi.parseField)
}

type YoungGenerationConfiguration struct {
	StartTime int64
	MinSize   int64
	MaxSize   int64
	NewRatio  int32
}

func (ygc *YoungGenerationConfiguration) parseField(name string, p ParseResolvable) (err error) {
	switch name {
	case "startTime":
		ygc.StartTime, err = toLong(p)
	case "minSize":
		ygc.MinSize, err = toLong(p)
	case "maxSize":
		ygc.MaxSize, err = toLong(p)
	case "newRatio":
		ygc.NewRatio, err = toInt(p)
	}
	return err
}

func (ygc *YoungGenerationConfiguration) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, ygc.parseField)
}

type UnsupportedEvent struct{}

func (ue *UnsupportedEvent) parseField(name string, p ParseResolvable) error {
	return nil
}

func (ue *UnsupportedEvent) Parse(r reader.Reader, classes ClassMap, cpools PoolMap, class ClassMetadata) error {
	return parseFields(r, classes, cpools, class, nil, true, ue.parseField)
}
