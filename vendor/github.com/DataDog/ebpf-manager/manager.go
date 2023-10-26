package manager

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DataDog/gopsutil/process"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/btf"
	"github.com/hashicorp/go-multierror"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"golang.org/x/sys/unix"
)

// ConstantEditor - A constant editor tries to rewrite the value of a constant in a compiled eBPF program.
//
// Constant edition only works before the eBPF programs are loaded in the kernel, and therefore before the
// Manager is started. If no program sections are provided, the manager will try to edit the constant in all eBPF programs.
type ConstantEditor struct {
	// Name - Name of the constant to rewrite
	Name string

	// Value - Value to write in the eBPF bytecode. When using the asm load method, the Value has to be a uint64.
	Value interface{}

	// FailOnMissing - If FailOMissing is set to true, the constant edition process will return an error if the constant
	// was missing in at least one program
	FailOnMissing bool

	// BTFGlobalConstant - Indicates if the constant is a BTF global constant.
	BTFGlobalConstant bool

	// ProbeIdentificationPairs - Identifies the list of programs to edit. If empty, it will apply to all the programs
	// of the manager. Will return an error if at least one edition failed.
	ProbeIdentificationPairs []ProbeIdentificationPair
}

// TailCallRoute - A tail call route defines how tail calls should be routed between eBPF programs.
//
// The provided eBPF program will be inserted in the provided eBPF program array, at the provided key. The eBPF program
// can be provided by its function name or by its *ebpf.Program representation.
type TailCallRoute struct {
	// ProgArrayName - Name of the BPF_MAP_TYPE_PROG_ARRAY map as defined in its section SEC("maps/[ProgArray]")
	ProgArrayName string

	// Key - Key at which the program will be inserted in the ProgArray map
	Key uint32

	// ProbeIdentificationPair - Selector of the program to insert in the ProgArray map
	ProbeIdentificationPair ProbeIdentificationPair

	// Program - Program to insert in the ProgArray map
	Program *ebpf.Program
}

// MapRoute - A map route defines how multiple maps should be routed between eBPF programs.
//
// The provided eBPF map will be inserted in the provided eBPF array of maps (or hash of maps), at the provided key. The
// inserted eBPF map can be provided by its section or by its *ebpf.Map representation.
type MapRoute struct {
	// RoutingMapName - Name of the BPF_MAP_TYPE_ARRAY_OF_MAPS or BPF_MAP_TYPE_HASH_OF_MAPS map, as defined in its
	// section SEC("maps/[RoutingMapName]")
	RoutingMapName string

	// Key - Key at which the program will be inserted in the routing map
	Key interface{}

	// RoutedName - Section of the map that will be inserted
	RoutedName string

	// Map - Map to insert in the routing map
	Map *ebpf.Map
}

// InnerOuterMapSpec - An InnerOuterMapSpec defines the map that should be used as the inner map of the provided outer map.
type InnerOuterMapSpec struct {
	// OuterMapName - Name of the BPF_MAP_TYPE_ARRAY_OF_MAPS or BPF_MAP_TYPE_HASH_OF_MAPS map, as defined in its
	// section SEC("maps/[OuterMapName]")
	OuterMapName string

	// InnerMapName - Name of the inner map of the provided outer map, as defined in its section SEC("maps/[InnerMapName]")
	InnerMapName string
}

// MapSpecEditorFlag - Flag used to specify what a MapSpecEditor should edit.
type MapSpecEditorFlag uint

const (
	EditType       MapSpecEditorFlag = 1 << 1
	EditMaxEntries MapSpecEditorFlag = 1 << 2
	EditFlags      MapSpecEditorFlag = 1 << 3
	EditKeyValue   MapSpecEditorFlag = 1 << 4
)

// MapSpecEditor - A MapSpec editor defines how specific parameters of specific maps should be updated at runtime
//
// For example, this can be used if you need to change the max_entries of a map before it is loaded in the kernel, but
// you don't know what this value should be initially.
type MapSpecEditor struct {
	// Type - Type of the map.
	Type ebpf.MapType
	// MaxEntries - Max Entries of the map.
	MaxEntries uint32
	// Flags - Flags provided to the kernel during the loading process.
	Flags uint32
	// KeySize - Defines the key size of the map
	KeySize uint32
	// Key - Defines the BTF type of the keys of the map
	Key btf.Type
	// ValueSize - Defines the value size of the map
	ValueSize uint32
	// Value - Defines the BTF type of the values of the map
	Value btf.Type
	// EditorFlag - Use this flag to specify what fields should be updated. See MapSpecEditorFlag.
	EditorFlag MapSpecEditorFlag
}

// FunctionExcluder - An interface for types that can be used for `AdditionalExcludedFunctionCollector`
type FunctionExcluder interface {
	// ShouldExcludeFunction - Returns true if the function should be excluded
	ShouldExcludeFunction(name string, prog *ebpf.ProgramSpec) bool
	// CleanCaches - Is called when the manager is done with the excluder (for memory reclaiming for example)
	CleanCaches()
}

// Options - Options of a Manager. These options define how a manager should be initialized.
type Options struct {
	// ActivatedProbes - List of the probes that should be activated, identified by their identification string.
	// If the list is empty, all probes will be activated.
	ActivatedProbes []ProbesSelector

	// KeepUnmappedProgramSpecs - Defines if the manager should keep unmapped ProgramSpec instances in the collection.
	// Enable this feature if you're going to clone one of these ProgramSpec.
	KeepUnmappedProgramSpecs bool

	// ExcludedFunctions - A list of functions that should not even be verified. This list overrides the ActivatedProbes
	// list: since the excluded sections aren't loaded in the kernel, all the probes using those sections will be
	// deactivated.
	ExcludedFunctions []string

	// AdditionalExcludedFunctionCollector - A dynamic function excluder, allowing to exclude functions based on the attach point
	AdditionalExcludedFunctionCollector FunctionExcluder

	// ExcludedMaps - A list of maps that should not be created.
	ExcludedMaps []string

	// ConstantsEditor - Post-compilation constant edition. See ConstantEditor for more.
	ConstantEditors []ConstantEditor

	// MapSpecEditor - Pre-loading MapSpec editors.
	MapSpecEditors map[string]MapSpecEditor

	// VerifierOptions - Defines the log level of the verifier and the size of its log buffer. Set to 0 to disable
	// logging and 1 to get a verbose output of the error. Increase the buffer size if the output is truncated.
	VerifierOptions ebpf.CollectionOptions

	// MapEditors - External map editor. The provided eBPF maps will overwrite the maps of the Manager if their names
	// match.
	// This is particularly useful to share maps across Managers (and therefore across isolated eBPF programs), without
	// having to use the MapRouter indirection. However, this technique only works before the eBPF programs are loaded,
	// and therefore before the Manager is started. The keys of the map are the names of the maps to edit, as defined
	// in their sections SEC("maps/[name]").
	MapEditors map[string]*ebpf.Map

	// MapEditorsIgnoreMissingMaps - If MapEditorsIgnoreMissingMaps is set to true, the map edition process will return an
	// error if a map was missing in at least one program
	MapEditorsIgnoreMissingMaps bool

	// MapRouter - External map routing. See MapRoute for more.
	MapRouter []MapRoute

	// InnerOuterMapSpecs - Defines the mapping between inner and outer maps. See InnerOuterMapSpec for more.
	InnerOuterMapSpecs []InnerOuterMapSpec

	// TailCallRouter - External tail call routing. See TailCallRoute for more.
	TailCallRouter []TailCallRoute

	// SymFile - Kernel symbol file. If not provided, the default `/proc/kallsyms` will be used.
	SymFile string

	// PerfRingBufferSize - Manager-level default value for the perf ring buffers. Defaults to the size of 1 page
	// on the system. See PerfMap.PerfRingBuffer for more.
	DefaultPerfRingBufferSize int

	// RingBufferSize - Manager-level default value for the ring buffers. Defaults to the size of 1 page
	// on the system.
	DefaultRingBufferSize int

	// Watermark - Manager-level default value for the watermarks of the perf ring buffers.
	// See PerfMap.Watermark for more.
	DefaultWatermark int

	// DefaultKProbeMaxActive - Manager-level default value for the kprobe max active parameter.
	// See Probe.MaxActive for more.
	DefaultKProbeMaxActive int

	// DefaultKprobeAttachMethod - Manager-level default value for the Kprobe attach method. Defaults to AttachKprobeWithPerfEventOpen if unset.
	DefaultKprobeAttachMethod KprobeAttachMethod

	// DefaultUprobeAttachMethod - Manager-level default value for the Uprobe attach method. Defaults to AttachWithPerfEventOpen if unset.
	DefaultUprobeAttachMethod AttachMethod

	// ProbeRetry - Defines the number of times that a probe will retry to attach / detach on error.
	DefaultProbeRetry uint

	// ProbeRetryDelay - Defines the delay to wait before a probe should retry to attach / detach on error.
	DefaultProbeRetryDelay time.Duration

	// RLimit - The maps & programs provided to the manager might exceed the maximum allowed memory lock.
	// (RLIMIT_MEMLOCK) If a limit is provided here it will be applied when the manager is initialized.
	RLimit *unix.Rlimit

	// KeepKernelBTF - Defines if the kernel types defined in VerifierOptions.Programs.KernelTypes should be cleaned up
	// once the manager is done using them. By default, the manager will clean them up to save up space. DISCLAIMER: if
	// your program uses "manager.CloneProgram", you might want to enable "KeepKernelBTF". As a workaround, you can also
	// try to strip as much as possible the content of "KernelTypes" to reduce the memory overhead.
	KeepKernelBTF bool

	// SkipPerfMapReaderStartup - Perf maps whose name is set to true with this option will not have their reader gorountine started when calling the manager.Start() function.
	// PerfMap.Start() can then be used to start reading events from the corresponding PerfMap.
	SkipPerfMapReaderStartup map[string]bool

	// SkipRingbufferReaderStartup - Ringbuffer maps whose name is set to true with this option will not have their reader gorountine started when calling the manager.Start() function.
	// RingBuffer.Start() can then be used to start reading events from the corresponding RingBuffer.
	SkipRingbufferReaderStartup map[string]bool
}

// NetlinkSocket - (TC classifier programs and XDP) Netlink socket cache entry holding the netlink socket and the
// TC filter count
type NetlinkSocket struct {
	Sock          *netlink.Handle
	TCFilterCount map[int]int
}

// NewNetlinkSocket - Returns a new NetlinkSocket instance
func NewNetlinkSocket(nsHandle uint64) (*NetlinkSocket, error) {
	var err error
	var netnsHandle netns.NsHandle
	cacheEntry := NetlinkSocket{
		TCFilterCount: make(map[int]int),
	}

	if nsHandle == 0 {
		netnsHandle = netns.None()
	} else {
		netnsHandle = netns.NsHandle(nsHandle)
	}

	// Open a netlink socket for the requested namespace
	cacheEntry.Sock, err = netlink.NewHandleAt(netnsHandle, unix.NETLINK_ROUTE)
	if err != nil {
		return nil, fmt.Errorf("couldn't open a netlink socket: %w", err)
	}
	return &cacheEntry, nil
}

// Manager - Helper structure that manages multiple eBPF programs and maps
type Manager struct {
	collectionSpec     *ebpf.CollectionSpec
	collection         *ebpf.Collection
	options            Options
	netlinkSocketCache *netlinkSocketCache
	state              state
	stateLock          sync.RWMutex

	// Probes - List of probes handled by the manager
	Probes []*Probe

	// Maps - List of maps handled by the manager. PerfMaps should not be defined here, but instead in the PerfMaps
	// section
	Maps []*Map

	// PerfMaps - List of perf ring buffers handled by the manager
	PerfMaps []*PerfMap

	// RingBuffers - List of perf ring buffers handled by the manager
	RingBuffers []*RingBuffer

	// DumpHandler - Callback function called when manager.DumpMaps() is called
	// and dump the current state (human-readable)
	DumpHandler func(manager *Manager, mapName string, currentMap *ebpf.Map) string

	// InstructionPatcher - Callback function called before loading probes, to
	// provide user the ability to perform last minute instruction patching.
	InstructionPatcher func(m *Manager) error
}

// DumpMaps - Return a string containing human-readable info about eBPF maps
// Dumps the set of maps provided, otherwise dumping all maps with a DumpHandler set.
func (m *Manager) DumpMaps(maps ...string) (string, error) {
	m.stateLock.RLock()
	defer m.stateLock.RUnlock()
	if m.collection == nil || m.state < initialized {
		return "", ErrManagerNotInitialized
	}

	if m.DumpHandler == nil {
		return "", nil
	}

	var mapsToDump map[string]struct{}
	if len(maps) > 0 {
		mapsToDump = make(map[string]struct{})
		for _, m := range maps {
			mapsToDump[m] = struct{}{}
		}
	}
	needDump := func(name string) bool {
		if mapsToDump == nil {
			// dump all maps
			return true
		}
		_, found := mapsToDump[name]
		return found
	}

	var output strings.Builder
	// Look in the list of maps
	for mapName, currentMap := range m.collection.Maps {
		if needDump(mapName) {
			output.WriteString(m.DumpHandler(m, mapName, currentMap))
		}
	}
	return output.String(), nil
}

// getMap - Thread unsafe version of GetMap
func (m *Manager) getMap(name string) (*ebpf.Map, bool, error) {
	eBPFMap, ok := m.collection.Maps[name]
	if ok {
		return eBPFMap, true, nil
	}
	// Look in the list of maps
	for _, managerMap := range m.Maps {
		if managerMap.Name == name {
			return managerMap.array, true, nil
		}
	}
	if perfMap, found := m.getPerfMap(name); found {
		return perfMap.array, true, nil
	}
	if ringBuffer, found := m.getRingBuffer(name); found {
		return ringBuffer.array, true, nil
	}
	return nil, false, nil
}

// GetMap - Return a pointer to the requested eBPF map
// name: name of the map, as defined by its section SEC("maps/[name]")
func (m *Manager) GetMap(name string) (*ebpf.Map, bool, error) {
	m.stateLock.RLock()
	defer m.stateLock.RUnlock()
	if m.collection == nil || m.state < initialized {
		return nil, false, ErrManagerNotInitialized
	}
	return m.getMap(name)
}

// GetMaps - Return the list of eBPF maps in the manager
func (m *Manager) GetMaps() (map[string]*ebpf.Map, error) {
	m.stateLock.RLock()
	defer m.stateLock.RUnlock()
	if m.collection == nil || m.state < initialized {
		return nil, ErrManagerNotInitialized
	}

	output := make(map[string]*ebpf.Map, len(m.collection.Maps))
	for section, m := range m.collection.Maps {
		output[section] = m
	}
	return output, nil
}

// getMapSpec - Thread unsafe version of GetMapSpec
func (m *Manager) getMapSpec(name string) (*ebpf.MapSpec, bool, error) {
	eBPFMap, ok := m.collectionSpec.Maps[name]
	if ok {
		return eBPFMap, true, nil
	}
	// Look in the list of maps
	for _, managerMap := range m.Maps {
		if managerMap.Name == name {
			return managerMap.arraySpec, true, nil
		}
	}
	if perfMap, found := m.getPerfMap(name); found {
		return perfMap.arraySpec, true, nil
	}
	if ringBuffer, found := m.getRingBuffer(name); found {
		return ringBuffer.arraySpec, true, nil
	}
	return nil, false, nil
}

// GetMapSpec - Return a pointer to the requested eBPF MapSpec. This is useful when duplicating a map.
func (m *Manager) GetMapSpec(name string) (*ebpf.MapSpec, bool, error) {
	m.stateLock.RLock()
	defer m.stateLock.RUnlock()
	if m.collectionSpec == nil || m.state < initialized {
		return nil, false, ErrManagerNotInitialized
	}
	return m.getMapSpec(name)
}

// getPerfMap - Thread unsafe version of GetPerfMap
func (m *Manager) getPerfMap(name string) (*PerfMap, bool) {
	for _, perfMap := range m.PerfMaps {
		if perfMap.Name == name {
			return perfMap, true
		}
	}
	return nil, false
}

// GetPerfMap - Select a perf map by its name
func (m *Manager) GetPerfMap(name string) (*PerfMap, bool) {
	m.stateLock.RLock()
	defer m.stateLock.RUnlock()
	return m.getPerfMap(name)
}

// getRingBuffer - Thread unsafe version of GetRingBuffer
func (m *Manager) getRingBuffer(name string) (*RingBuffer, bool) {
	for _, ringBuffer := range m.RingBuffers {
		if ringBuffer.Name == name {
			return ringBuffer, true
		}
	}
	return nil, false
}

// GetRingBuffer - Select a ring buffer by its name
func (m *Manager) GetRingBuffer(name string) (*RingBuffer, bool) {
	m.stateLock.RLock()
	defer m.stateLock.RUnlock()
	return m.getRingBuffer(name)
}

// getProgram - Thread unsafe version of GetProgram
func (m *Manager) getProgram(id ProbeIdentificationPair) ([]*ebpf.Program, bool, error) {
	var programs []*ebpf.Program
	if id.UID == "" {
		for _, probe := range m.Probes {
			if probe.EBPFDefinitionMatches(id) {
				programs = append(programs, probe.program)
			}
		}
		if len(programs) > 0 {
			return programs, true, nil
		}
		prog, ok := m.collection.Programs[id.EBPFFuncName]
		return []*ebpf.Program{prog}, ok, nil
	}
	for _, probe := range m.Probes {
		if probe.Matches(id) {
			return []*ebpf.Program{probe.program}, true, nil
		}
	}
	return programs, false, nil
}

// GetProgram - Return a pointer to the requested eBPF program
// section: section of the program, as defined by its section SEC("[section]")
// id: unique identifier given to a probe. If UID is empty, then all the programs matching the provided section are
// returned.
func (m *Manager) GetProgram(id ProbeIdentificationPair) ([]*ebpf.Program, bool, error) {
	m.stateLock.RLock()
	defer m.stateLock.RUnlock()
	if m.collection == nil || m.state < initialized {
		return nil, false, ErrManagerNotInitialized
	}
	return m.getProgram(id)
}

// GetPrograms - Return the list of eBPF programs in the manager
func (m *Manager) GetPrograms() (map[string]*ebpf.Program, error) {
	m.stateLock.RLock()
	defer m.stateLock.RUnlock()
	if m.collection == nil || m.state < initialized {
		return nil, ErrManagerNotInitialized
	}

	return maps.Clone(m.collection.Programs), nil
}

// getProgramSpec - Thread unsafe version of GetProgramSpec
func (m *Manager) getProgramSpec(id ProbeIdentificationPair) ([]*ebpf.ProgramSpec, bool, error) {
	var programs []*ebpf.ProgramSpec
	if id.UID == "" {
		for _, probe := range m.Probes {
			if probe.EBPFDefinitionMatches(id) {
				programs = append(programs, probe.programSpec)
			}
		}
		if len(programs) > 0 {
			return programs, true, nil
		}
		prog, ok := m.collectionSpec.Programs[id.EBPFFuncName]
		return []*ebpf.ProgramSpec{prog}, ok, nil
	}
	for _, probe := range m.Probes {
		if probe.Matches(id) {
			return []*ebpf.ProgramSpec{probe.programSpec}, true, nil
		}
	}
	return programs, false, nil
}

// GetProgramSpec - Return a pointer to the requested eBPF program spec
// section: section of the program, as defined by its section SEC("[section]")
// id: unique identifier given to a probe. If UID is empty, then the original program spec with the right section in the
// collection spec (if found) is return
func (m *Manager) GetProgramSpec(id ProbeIdentificationPair) ([]*ebpf.ProgramSpec, bool, error) {
	m.stateLock.RLock()
	defer m.stateLock.RUnlock()
	if m.collectionSpec == nil || m.state < initialized {
		return nil, false, ErrManagerNotInitialized
	}
	return m.getProgramSpec(id)
}

// getProbe - Thread unsafe version of GetProbe
func (m *Manager) getProbe(id ProbeIdentificationPair) (*Probe, bool) {
	for _, managerProbe := range m.Probes {
		if managerProbe.Matches(id) {
			return managerProbe, true
		}
	}
	return nil, false
}

// GetProbe - Select a probe by its section and UID
func (m *Manager) GetProbe(id ProbeIdentificationPair) (*Probe, bool) {
	m.stateLock.RLock()
	defer m.stateLock.RUnlock()
	return m.getProbe(id)
}

// RenameProbeIdentificationPair - Renames a probe identification pair. This change will propagate to all the features in
// the manager that will try to select the probe by its old ProbeIdentificationPair.
func (m *Manager) RenameProbeIdentificationPair(oldID ProbeIdentificationPair, newID ProbeIdentificationPair) error {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()

	// sanity check: make sure the newID doesn't already exist
	for _, mProbe := range m.Probes {
		if mProbe.Matches(newID) {
			return ErrIdentificationPairInUse
		}
	}

	if oldID.EBPFFuncName != newID.EBPFFuncName {
		// edit the excluded sections
		for i, excludedFuncName := range m.options.ExcludedFunctions {
			if excludedFuncName == oldID.EBPFFuncName {
				m.options.ExcludedFunctions[i] = newID.EBPFFuncName
			}
		}
	}

	// edit the probe selectors
	for _, selector := range m.options.ActivatedProbes {
		selector.EditProbeIdentificationPair(oldID, newID)
	}

	// edit the probe
	p, ok := m.getProbe(oldID)
	if !ok {
		return ErrSymbolNotFound
	}
	return p.RenameProbeIdentificationPair(newID)
}

// Init - Initialize the manager.
// elf: reader containing the eBPF bytecode
func (m *Manager) Init(elf io.ReaderAt) error {
	return m.InitWithOptions(elf, Options{})
}

// InitWithOptions - Initialize the manager.
// elf: reader containing the eBPF bytecode
// options: options provided to the manager to configure its initialization
func (m *Manager) InitWithOptions(elf io.ReaderAt, options Options) error {
	m.stateLock.Lock()
	if m.state > initialized {
		m.stateLock.Unlock()
		return ErrManagerRunning
	}

	m.options = options
	m.netlinkSocketCache = newNetlinkSocketCache()
	if m.options.DefaultPerfRingBufferSize == 0 {
		m.options.DefaultPerfRingBufferSize = os.Getpagesize()
	}
	if m.options.DefaultRingBufferSize == 0 {
		m.options.DefaultRingBufferSize = os.Getpagesize() * 16
	}

	// perform a quick sanity check on the provided probes and maps
	if err := m.sanityCheck(); err != nil {
		m.stateLock.Unlock()
		return err
	}

	// set resource limit if requested
	if m.options.RLimit != nil {
		err := unix.Setrlimit(unix.RLIMIT_MEMLOCK, m.options.RLimit)
		if err != nil {
			return fmt.Errorf("couldn't adjust RLIMIT_MEMLOCK: %w", err)
		}
	}

	// Load the provided elf buffer
	var err error
	m.collectionSpec, err = ebpf.LoadCollectionSpecFromReader(elf)
	if err != nil {
		m.stateLock.Unlock()
		return err
	}

	if m.options.AdditionalExcludedFunctionCollector != nil {
		for key, prog := range m.collectionSpec.Programs {
			if m.options.AdditionalExcludedFunctionCollector.ShouldExcludeFunction(key, prog) {
				m.options.ExcludedFunctions = append(m.options.ExcludedFunctions, key)
			}
		}
		m.options.AdditionalExcludedFunctionCollector.CleanCaches()
	}

	// Remove excluded programs
	for _, excludedFuncName := range m.options.ExcludedFunctions {
		delete(m.collectionSpec.Programs, excludedFuncName)
	}
	for i := 0; i < len(m.Probes); {
		if slices.Contains(m.options.ExcludedFunctions, m.Probes[i].EBPFFuncName) {
			m.Probes = slices.Delete(m.Probes, i, i+1)
		} else {
			i++
		}
	}
	// Remove ConstantEditors that point to only excluded functions
	for i := 0; i < len(m.options.ConstantEditors); {
		ce := m.options.ConstantEditors[i]
		isSpecific := len(ce.ProbeIdentificationPairs) > 0
		// remove excluded ProbeIdentificationPairs
		for j := 0; j < len(ce.ProbeIdentificationPairs); {
			if slices.Contains(m.options.ExcludedFunctions, ce.ProbeIdentificationPairs[j].EBPFFuncName) {
				ce.ProbeIdentificationPairs = slices.Delete(ce.ProbeIdentificationPairs, j, j+1)
			} else {
				j++
			}
		}
		// was a specific constant editor, but all the probes are excluded
		if isSpecific && len(ce.ProbeIdentificationPairs) == 0 {
			m.options.ConstantEditors = slices.Delete(m.options.ConstantEditors, i, i+1)
		} else {
			i++
		}
	}
	// Remove excluded maps
	for _, excludeMapName := range m.options.ExcludedMaps {
		delete(m.collectionSpec.Maps, excludeMapName)
	}
	for i := 0; i < len(m.Maps); {
		if slices.Contains(m.options.ExcludedMaps, m.Maps[i].Name) {
			m.Maps = slices.Delete(m.Maps, i, i+1)
		} else {
			i++
		}
	}

	// Match Maps and program specs
	if err = m.matchSpecs(); err != nil {
		m.stateLock.Unlock()
		return err
	}

	// Configure activated probes
	m.activateProbes()
	m.state = initialized
	m.stateLock.Unlock()
	resetManager := func(m *Manager) {
		m.stateLock.Lock()
		m.state = reset
		m.stateLock.Unlock()
	}

	// newEditor program constants
	if len(options.ConstantEditors) > 0 {
		if err = m.editConstants(); err != nil {
			resetManager(m)
			return err
		}
	}

	// newEditor map spec
	if len(options.MapSpecEditors) > 0 {
		if err = m.editMapSpecs(); err != nil {
			resetManager(m)
			return err
		}
	}

	// Setup map routes
	if len(options.InnerOuterMapSpecs) > 0 {
		for _, ioMapSpec := range options.InnerOuterMapSpecs {
			if err = m.editInnerOuterMapSpec(ioMapSpec); err != nil {
				resetManager(m)
				return err
			}
		}
	}

	// newEditor program maps
	if len(options.MapEditors) > 0 {
		if err = m.editMaps(options.MapEditors); err != nil {
			resetManager(m)
			return err
		}
	}

	// Patch instructions
	if m.InstructionPatcher != nil {
		if err := m.InstructionPatcher(m); err != nil {
			resetManager(m)
			return err
		}
	}

	// Load pinned maps and pinned programs to avoid loading them twice
	if err = m.loadPinnedObjects(); err != nil {
		resetManager(m)
		return err
	}

	// Load eBPF program with the provided verifier options
	if err = m.loadCollection(); err != nil {
		if m.collection != nil {
			_ = m.collection.Close
		}
		resetManager(m)
		return err
	}
	return nil
}

// Start - Attach eBPF programs, start perf ring readers and apply maps and tail calls routing.
func (m *Manager) Start() error {
	m.stateLock.Lock()
	if m.state < initialized {
		m.stateLock.Unlock()
		return ErrManagerNotInitialized
	}
	if m.state >= running {
		m.stateLock.Unlock()
		return nil
	}

	if !m.options.KeepKernelBTF {
		// release kernel BTF. It should no longer be needed
		m.options.VerifierOptions.Programs.KernelTypes = nil
	}

	// clean up tracefs
	if err := m.cleanupTracefs(); err != nil {
		m.stateLock.Unlock()
		return fmt.Errorf("failed to cleanup tracefs: %w", err)
	}

	// Start perf ring readers
	for _, perfRing := range m.PerfMaps {
		if m.options.SkipPerfMapReaderStartup[perfRing.Name] {
			continue
		}
		if err := perfRing.Start(); err != nil {
			// Clean up
			_ = m.stop(CleanInternal)
			m.stateLock.Unlock()
			return err
		}
	}

	// Start ring buffer readers
	for _, ringBuffer := range m.RingBuffers {
		if m.options.SkipRingbufferReaderStartup[ringBuffer.Name] {
			continue
		}
		if err := ringBuffer.Start(); err != nil {
			// Clean up
			_ = m.stop(CleanInternal)
			m.stateLock.Unlock()
			return err
		}
	}

	// Attach eBPF programs
	for _, probe := range m.Probes {
		// ignore the error, they are already collected per probes and will be surfaced by the
		// activation validators if needed.
		_ = probe.Attach()
	}

	m.state = running
	m.stateLock.Unlock()

	// Check probe selectors
	var validationErrs error
	for _, selector := range m.options.ActivatedProbes {
		if err := selector.RunValidator(m); err != nil {
			validationErrs = multierror.Append(validationErrs, err)
		}
	}
	if validationErrs != nil {
		// Clean up
		_ = m.Stop(CleanInternal)
		return fmt.Errorf("probes activation validation failed: %w", validationErrs)
	}

	// Handle Maps router
	if err := m.UpdateMapRoutes(m.options.MapRouter...); err != nil {
		// Clean up
		_ = m.Stop(CleanInternal)
		return err
	}

	// Handle Program router
	if err := m.UpdateTailCallRoutes(m.options.TailCallRouter...); err != nil {
		// Clean up
		_ = m.Stop(CleanInternal)
		return err
	}
	return nil
}

func (m *Manager) Pause() error {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()
	if m.state == paused {
		return nil
	}
	if m.state <= initialized {
		return ErrManagerNotStarted
	}
	for _, probe := range m.Probes {
		if err := probe.Pause(); err != nil {
			return err
		}
	}
	m.state = paused
	return nil
}

func (m *Manager) Resume() error {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()
	if m.state == running {
		return nil
	}
	if m.state <= initialized {
		return ErrManagerNotStarted
	}
	for _, probe := range m.Probes {
		if err := probe.Resume(); err != nil {
			return err
		}
	}
	m.state = running
	return nil
}

// Stop - Detach all eBPF programs and stop perf ring readers. The cleanup parameter defines which maps should be closed.
// See MapCleanupType for mode.
func (m *Manager) Stop(cleanup MapCleanupType) error {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()
	if m.state < initialized {
		return ErrManagerNotInitialized
	}
	return m.stop(cleanup)
}

// StopReaders stop the kernel events readers Perf or Ring buffer
func (m *Manager) StopReaders(cleanup MapCleanupType) error {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()
	return m.stopReaders(cleanup)
}

func (m *Manager) stopReaders(cleanup MapCleanupType) error {
	var err error

	// Stop perf ring readers
	for _, perfRing := range m.PerfMaps {
		if stopErr := perfRing.Stop(cleanup); stopErr != nil {
			err = concatErrors(err, fmt.Errorf("perf ring reader %s couldn't gracefully shut down: %w", perfRing.Name, stopErr))
		}
	}

	// Stop ring buffer readers
	for _, ringBuffer := range m.RingBuffers {
		if stopErr := ringBuffer.Stop(cleanup); stopErr != nil {
			err = concatErrors(err, fmt.Errorf("ring buffer reader %s couldn't gracefully shut down: %w", ringBuffer.Name, stopErr))
		}
	}

	return err
}

func (m *Manager) stop(cleanup MapCleanupType) error {
	err := m.stopReaders(cleanup)

	// Detach eBPF programs
	for _, probe := range m.Probes {
		if stopErr := probe.Stop(); stopErr != nil {
			err = concatErrors(err, fmt.Errorf("program %s couldn't gracefully shut down: %w", probe.ProbeIdentificationPair, stopErr))
		}
	}

	// Close maps
	for _, managerMap := range m.Maps {
		if closeErr := managerMap.Close(cleanup); closeErr != nil {
			err = concatErrors(err, fmt.Errorf("couldn't gracefully close map %s: %w", managerMap.Name, closeErr))
		}
	}

	// Close all netlink sockets
	m.netlinkSocketCache.cleanup()

	// Clean up collection
	// Note: we might end up closing the same programs or maps multiple times but the library gracefully handles those
	// situations. We can't rely only on the collection to close all maps and programs because some pinned objects were
	// removed from the collection.
	m.collection.Close()

	m.state = reset
	return err
}

// NewMap - Create a new map using the provided parameters. The map is added to the list of maps managed by the manager.
// Use a MapRoute to make this map available to the programs of the manager.
func (m *Manager) NewMap(spec ebpf.MapSpec, options MapOptions) (*ebpf.Map, error) {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()
	if m.collection == nil || m.state < initialized {
		return nil, ErrManagerNotInitialized
	}

	// check if the name of the new map is available
	_, exists, _ := m.getMap(spec.Name)
	if exists {
		return nil, ErrMapNameInUse
	}

	// Create the new map
	managerMap, err := loadNewMap(spec, options)
	if err != nil {
		return nil, err
	}

	// init map
	if err := managerMap.init(m); err != nil {
		// Clean up
		_ = managerMap.Close(CleanInternal)
		return nil, err
	}

	// Add map to the list of maps managed by the manager
	m.Maps = append(m.Maps, managerMap)
	return managerMap.array, nil
}

// CloneMap - Duplicates the spec of an existing map, before creating a new one.
// Use a MapRoute to make this map available to the programs of the manager.
func (m *Manager) CloneMap(name string, newName string, options MapOptions) (*ebpf.Map, error) {
	// Select map to clone
	oldSpec, exists, err := m.GetMapSpec(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("failed to clone maps/%s: %w", name, ErrUnknownSection)
	}

	// Duplicate spec and create a new map
	spec := oldSpec.Copy()
	spec.Name = newName
	return m.NewMap(*spec, options)
}

// NewPerfRing - Creates a new perf ring and start listening for events.
// Use a MapRoute to make this map available to the programs of the manager.
func (m *Manager) NewPerfRing(spec ebpf.MapSpec, options MapOptions, perfMapOptions PerfMapOptions) (*ebpf.Map, error) {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()
	if m.state < initialized {
		return nil, ErrManagerNotInitialized
	}

	// check if the name of the new map is available
	_, exists, _ := m.getMap(spec.Name)
	if exists {
		return nil, ErrMapNameInUse
	}

	// Create new map and perf ring buffer reader
	perfMap, err := loadNewPerfMap(spec, options, perfMapOptions)
	if err != nil {
		return nil, err
	}

	// Setup perf buffer reader
	if err := perfMap.init(m); err != nil {
		return nil, err
	}

	// Start perf buffer reader
	if err := perfMap.Start(); err != nil {
		// clean up
		_ = perfMap.Stop(CleanInternal)
		return nil, err
	}

	// Add map to the list of perf ring managed by the manager
	m.PerfMaps = append(m.PerfMaps, perfMap)
	return perfMap.array, nil
}

// ClonePerfRing - Clone an existing perf map and create a new one with the same spec.
// Use a MapRoute to make this map available to the programs of the manager.
func (m *Manager) ClonePerfRing(name string, newName string, options MapOptions, perfMapOptions PerfMapOptions) (*ebpf.Map, error) {
	// Select map to clone
	oldSpec, exists, err := m.GetMapSpec(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("failed to clone maps/%s: couldn't find map: %w", name, ErrUnknownSection)
	}

	// Duplicate spec and create a new map
	spec := oldSpec.Copy()
	spec.Name = newName
	return m.NewPerfRing(*spec, options, perfMapOptions)
}

// NewRingBuffer - Creates a new ring buffer and start listening for events.
// Use a MapRoute to make this map available to the programs of the manager.
func (m *Manager) NewRingBuffer(spec ebpf.MapSpec, options MapOptions, ringBufferOptions RingBufferOptions) (*ebpf.Map, error) {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()
	if m.state < initialized {
		return nil, ErrManagerNotInitialized
	}

	// check if the name of the new map is available
	_, exists, _ := m.getMap(spec.Name)
	if exists {
		return nil, ErrMapNameInUse
	}

	// Create new map and ring buffer reader
	ringBuffer, err := loadNewRingBuffer(spec, options, ringBufferOptions)
	if err != nil {
		return nil, err
	}

	// Setup ring buffer reader
	if err := ringBuffer.init(m); err != nil {
		return nil, err
	}

	// Start perf buffer reader
	if err := ringBuffer.Start(); err != nil {
		// clean up
		_ = ringBuffer.Stop(CleanInternal)
		return nil, err
	}

	// Add map to the list of perf ring managed by the manager
	m.RingBuffers = append(m.RingBuffers, ringBuffer)
	return ringBuffer.array, nil
}

// AddHook - Hook an existing program to a hook point. This is particularly useful when you need to trigger an
// existing program on a hook point that is determined at runtime. For example, you might want to hook an existing
// eBPF TC classifier to the newly created interface of a container. Make sure to specify a unique uid in the new probe,
// you will need it if you want to detach the program later. The original program is selected using the provided UID,
// the section and the eBPF function name provided in the new probe.
func (m *Manager) AddHook(UID string, newProbe *Probe) error {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()
	if m.collection == nil || m.state < initialized {
		return ErrManagerNotInitialized
	}

	oldID := ProbeIdentificationPair{UID: UID, EBPFFuncName: newProbe.EBPFFuncName}
	// Look for the eBPF program
	progs, found, err := m.getProgram(oldID)
	if err != nil {
		return err
	}
	if !found || len(progs) == 0 {
		return fmt.Errorf("couldn't find program %s: %w", oldID, ErrUnknownSectionOrFuncName)
	}
	prog := progs[0]
	progSpecs, found, _ := m.getProgramSpec(oldID)
	if !found || len(progSpecs) == 0 {
		return fmt.Errorf("couldn't find programSpec %s: %w", oldID, ErrUnknownSectionOrFuncName)
	}
	progSpec := progSpecs[0]

	// Ensure that the new probe is enabled
	newProbe.Enabled = true

	// Make sure the provided identification pair is unique
	_, exists, _ := m.getProgramSpec(newProbe.ProbeIdentificationPair)
	if exists {
		return fmt.Errorf("probe %s already exists: %w", newProbe.ProbeIdentificationPair, ErrIdentificationPairInUse)
	}

	// Clone program
	clonedProg, err := prog.Clone()
	if err != nil {
		return fmt.Errorf("couldn't clone %v: %w", oldID, err)
	}
	newProbe.program = clonedProg
	newProbe.programSpec = progSpec

	// init program
	if err = newProbe.init(m); err != nil {
		// clean up
		_ = newProbe.Stop()
		return fmt.Errorf("failed to initialize new probe: %w", err)
	}

	// Pin if needed
	if newProbe.PinPath != "" {
		if err = newProbe.program.Pin(newProbe.PinPath); err != nil {
			// clean up
			_ = newProbe.Stop()
			return fmt.Errorf("couldn't pin new probe: %w", err)
		}
	}

	// Attach program
	if err = newProbe.Attach(); err != nil {
		// clean up
		_ = newProbe.Stop()
		return fmt.Errorf("couldn't attach new probe: %w", err)
	}

	// Add probe to the list of probes
	m.Probes = append(m.Probes, newProbe)
	return nil
}

// DetachHook - Detach an eBPF program from a hook point. If there is only one instance left of this program in the
// kernel, then the probe will be detached but the program will not be closed (so that it can be used later). In that
// case, calling DetachHook has essentially the same effect as calling Detach() on the right Probe instance. However,
// if there are more than one instance in the kernel of the requested program, then the probe selected by the provided
// ProbeIdentificationPair is detached, and its own handle of the program is closed.
func (m *Manager) DetachHook(id ProbeIdentificationPair) error {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()
	if m.collection == nil || m.state < initialized {
		return ErrManagerNotInitialized
	}

	// Check how many instances of the program are left in the kernel
	progs, _, err := m.getProgram(ProbeIdentificationPair{UID: "", EBPFFuncName: id.EBPFFuncName})
	if err != nil {
		return err
	}
	shouldStop := len(progs) > 1

	// Look for the probe
	idToDelete := -1
	for mID, mProbe := range m.Probes {
		if mProbe.Matches(id) {
			// Detach or stop the probe depending on shouldStop
			if shouldStop {
				if err = mProbe.Stop(); err != nil {
					return fmt.Errorf("couldn't stop probe %s: %w", id, err)
				}
			} else {
				if err = mProbe.Detach(); err != nil {
					return fmt.Errorf("couldn't detach probe %s: %w", id, err)
				}
			}
			idToDelete = mID
		}
	}
	if idToDelete >= 0 {
		m.Probes = slices.Delete(m.Probes, idToDelete, idToDelete+1)
	}
	return nil
}

// CloneProgram - Create a clone of a program, load it in the kernel and attach it to its hook point. Since the eBPF
// program instructions are copied before the program is loaded, you can edit them with a ConstantEditor, or remap
// the eBPF maps as you like. This is particularly useful to work around the absence of Array of Maps and Hash of Maps:
// first create the new maps you need, then clone the program you're interested in and rewrite it with the new maps,
// using a MapEditor. The original program is selected using the provided UID and the section provided in the new probe.
// Note that the BTF based constant edition will not work with this method.
func (m *Manager) CloneProgram(UID string, newProbe *Probe, constantsEditors []ConstantEditor, mapEditors map[string]*ebpf.Map) error {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()
	if m.collection == nil || m.state < initialized {
		return ErrManagerNotInitialized
	}

	oldID := ProbeIdentificationPair{UID: UID, EBPFFuncName: newProbe.EBPFFuncName}
	var oldProgramSpec *ebpf.ProgramSpec
	// look for an existing probe
	oldProbe, found := m.getProbe(oldID)
	if found {
		// check if the program spec of this probe was removed
		if !oldProbe.KeepProgramSpec {
			return fmt.Errorf("couldn't clone %s: this probe was cleaned up because KeepProgramSpec was disabled", oldID)
		}
		oldProgramSpec = oldProbe.programSpec
	} else {
		// the cloned program might not have a dedicated Probe, fallback to a collectionPec lookup
		progSpecs, found, err := m.getProgramSpec(oldID)
		if err != nil {
			return err
		}
		if !found || len(progSpecs) == 0 {
			return fmt.Errorf("couldn't find programSpec %v: %w", oldID, ErrUnknownSectionOrFuncName)
		}
		oldProgramSpec = progSpecs[0]

		// the program spec was found
		if !m.options.KeepUnmappedProgramSpecs {
			return fmt.Errorf("couldn't clone %s: this probe was cleaned up because KeepUnmappedProgramSpecs was disabled", oldID)
		}
	}

	// Check if the new probe has a unique identification pair
	_, exists, _ := m.getProgram(newProbe.ProbeIdentificationPair)
	if exists {
		return fmt.Errorf("couldn't add probe %v: %w", newProbe.ProbeIdentificationPair, ErrIdentificationPairInUse)
	}

	// Make sure the new probe is activated
	newProbe.Enabled = true

	// Clone the program
	clonedSpec := oldProgramSpec.Copy()
	newProbe.programSpec = clonedSpec

	// newEditor constants
	for _, editor := range constantsEditors {
		if err := m.editConstant(newProbe.programSpec, editor); err != nil {
			return fmt.Errorf("couldn't edit constant %s: %w", editor.Name, err)
		}
	}

	// Write current maps
	if err := m.rewriteMaps(newProbe.programSpec, m.collection.Maps, false); err != nil {
		return fmt.Errorf("couldn't rewrite maps in %v: %w", newProbe.ProbeIdentificationPair, err)
	}

	// Rewrite with new maps
	if err := m.rewriteMaps(newProbe.programSpec, mapEditors, true); err != nil {
		return fmt.Errorf("couldn't rewrite maps in %v: %w", newProbe.ProbeIdentificationPair, err)
	}

	// init
	if err := newProbe.initWithOptions(m, true, true); err != nil {
		// clean up
		_ = newProbe.Stop()
		return fmt.Errorf("failed to initialize new probe %v: %w", newProbe.ProbeIdentificationPair, err)
	}

	// Attach new program
	if err := newProbe.Attach(); err != nil {
		// clean up
		_ = newProbe.Stop()
		return fmt.Errorf("failed to attach new probe %v: %w", newProbe.ProbeIdentificationPair, err)
	}

	// Add probe to the list of probes
	m.Probes = append(m.Probes, newProbe)
	return nil
}

// UpdateMapRoutes - Update one or multiple map of maps structures so that the provided keys point to the provided maps.
func (m *Manager) UpdateMapRoutes(router ...MapRoute) error {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()
	if m.collection == nil || m.state < initialized {
		return ErrManagerNotInitialized
	}

	for _, route := range router {
		if err := m.updateMapRoute(route); err != nil {
			return err
		}
	}
	return nil
}

// updateMapRoute - Update a map of maps structure so that the provided key points to the provided map
func (m *Manager) updateMapRoute(route MapRoute) error {
	// Select the routing map
	routingMap, found, err := m.getMap(route.RoutingMapName)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("couldn't find routing map %s: %w", route.RoutingMapName, ErrUnknownSection)
	}

	// Get file descriptor of the routed map
	var fd uint32
	if route.Map != nil {
		fd = uint32(route.Map.FD())
	} else {
		var routedMap *ebpf.Map
		routedMap, found, err = m.getMap(route.RoutedName)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("couldn't find routed map %s: %w", route.RoutedName, ErrUnknownSection)
		}
		fd = uint32(routedMap.FD())
	}

	// Insert map
	if err = routingMap.Put(route.Key, fd); err != nil {
		return fmt.Errorf("couldn't update routing map %s: %w", route.RoutingMapName, err)
	}
	return nil
}

// UpdateTailCallRoutes - Update one or multiple program arrays so that the provided keys point to the provided programs.
func (m *Manager) UpdateTailCallRoutes(router ...TailCallRoute) error {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()
	if m.collection == nil || m.state < initialized {
		return ErrManagerNotInitialized
	}

	for _, route := range router {
		if err := m.updateTailCallRoute(route); err != nil {
			return err
		}
	}
	return nil
}

// updateTailCallRoute - Update a program array so that the provided key point to the provided program.
func (m *Manager) updateTailCallRoute(route TailCallRoute) error {
	// Select the routing map
	routingMap, found, err := m.getMap(route.ProgArrayName)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("couldn't find routing map %s: %w", route.ProgArrayName, ErrUnknownSection)
	}

	// Get file descriptor of the routed program
	var fd uint32
	if route.Program != nil {
		fd = uint32(route.Program.FD())
	} else {
		progs, found, err := m.getProgram(route.ProbeIdentificationPair)
		if err != nil {
			return err
		}
		if !found || len(progs) == 0 {
			return fmt.Errorf("couldn't find program %v: %w", route.ProbeIdentificationPair, ErrUnknownSectionOrFuncName)
		}
		if progs[0] == nil {
			return fmt.Errorf("the program that you are trying to route to is empty")
		}
		fd = uint32(progs[0].FD())
	}

	// Insert tail call
	if err = routingMap.Put(route.Key, fd); err != nil {
		return fmt.Errorf("couldn't update routing map %s: %w", route.ProgArrayName, err)
	}
	return nil
}

func (m *Manager) getProbeProgramSpec(funcName string) (*ebpf.ProgramSpec, error) {
	spec, ok := m.collectionSpec.Programs[funcName]
	if !ok {
		return nil, fmt.Errorf("couldn't find program spec for func %s: %w", funcName, ErrUnknownSectionOrFuncName)
	}
	return spec, nil
}

func (m *Manager) getProbeProgram(funcName string) (*ebpf.Program, error) {
	p, ok := m.collection.Programs[funcName]
	if !ok {
		return nil, fmt.Errorf("couldn't find program %s: %w", funcName, ErrUnknownSectionOrFuncName)
	}
	return p, nil
}

// matchSpecs - Match loaded maps and program specs with the maps and programs provided to the manager
func (m *Manager) matchSpecs() error {
	// Match programs
	for _, probe := range m.Probes {
		programSpec, err := m.getProbeProgramSpec(probe.ProbeIdentificationPair.EBPFFuncName)
		if err != nil {
			return err
		}
		if !probe.CopyProgram {
			probe.programSpec = programSpec
		} else {
			probe.programSpec = programSpec.Copy()
			m.collectionSpec.Programs[probe.GetEBPFFuncName()] = probe.programSpec
		}
	}

	// Match maps
	for _, managerMap := range m.Maps {
		spec, ok := m.collectionSpec.Maps[managerMap.Name]
		if !ok {
			return fmt.Errorf("couldn't find map at maps/%s: %w", managerMap.Name, ErrUnknownSection)
		}
		spec.Contents = managerMap.Contents
		spec.Freeze = managerMap.Freeze
		managerMap.arraySpec = spec
	}

	// Match perfmaps
	for _, perfMap := range m.PerfMaps {
		spec, ok := m.collectionSpec.Maps[perfMap.Name]
		if !ok {
			return fmt.Errorf("couldn't find map at maps/%s: %w", perfMap.Name, ErrUnknownSection)
		}
		perfMap.arraySpec = spec
	}

	// Match ring buffer
	for _, ringBuffer := range m.RingBuffers {
		spec, ok := m.collectionSpec.Maps[ringBuffer.Name]
		if !ok {
			return fmt.Errorf("couldn't find map at maps/%s: %w", ringBuffer.Name, ErrUnknownSection)
		}
		ringBuffer.arraySpec = spec
	}

	return nil
}

// matchBPFObjects - Match loaded maps and program specs with the maps and programs provided to the manager
func (m *Manager) matchBPFObjects() error {
	// Match programs
	var mappedProgramSpecNames []string
	for _, probe := range m.Probes {
		program, err := m.getProbeProgram(probe.GetEBPFFuncName())
		if err != nil {
			return err
		}
		probe.program = program
		mappedProgramSpecNames = append(mappedProgramSpecNames, probe.ProbeIdentificationPair.EBPFFuncName)
	}

	// cleanup unmapped ProgramSpec now
	if !m.options.KeepUnmappedProgramSpecs {
	collectionSpec:
		for specName, spec := range m.collectionSpec.Programs {
			for _, name := range mappedProgramSpecNames {
				if specName == name {
					continue collectionSpec
				}
			}
			cleanupProgramSpec(spec)
		}
	}

	// Match maps
	for _, managerMap := range m.Maps {
		if managerMap.externalMap {
			continue
		}
		arr, ok := m.collection.Maps[managerMap.Name]
		if !ok {
			return fmt.Errorf("couldn't find map at maps/%s: %w", managerMap.Name, ErrUnknownSection)
		}
		managerMap.array = arr
	}

	// Match perfmaps
	for _, perfMap := range m.PerfMaps {
		arr, ok := m.collection.Maps[perfMap.Name]
		if !ok {
			return fmt.Errorf("couldn't find map at maps/%s: %w", perfMap.Name, ErrUnknownSection)
		}
		perfMap.array = arr
	}

	// Match ring buffers
	for _, ringBuffer := range m.RingBuffers {
		arr, ok := m.collection.Maps[ringBuffer.Name]
		if !ok {
			return fmt.Errorf("couldn't find map at maps/%s: %w", ringBuffer.Name, ErrUnknownSection)
		}
		ringBuffer.array = arr
	}

	return nil
}

func (m *Manager) activateProbes() {
	shouldPopulateActivatedProbes := len(m.options.ActivatedProbes) == 0
	for _, mProbe := range m.Probes {
		shouldActivate := shouldPopulateActivatedProbes
		for _, selector := range m.options.ActivatedProbes {
			for _, p := range selector.GetProbesIdentificationPairList() {
				if mProbe.Matches(p) {
					shouldActivate = true
				}
			}
		}
		mProbe.Enabled = shouldActivate

		if shouldPopulateActivatedProbes {
			// this will ensure that we check that everything has been activated by default when no selectors are provided
			m.options.ActivatedProbes = append(m.options.ActivatedProbes, &ProbeSelector{
				ProbeIdentificationPair: mProbe.ProbeIdentificationPair,
			})
		}
	}
}

// UpdateActivatedProbes - update the list of activated probes
func (m *Manager) UpdateActivatedProbes(selectors []ProbesSelector) error {
	m.stateLock.Lock()
	if m.state < initialized {
		m.stateLock.Unlock()
		return ErrManagerNotInitialized
	}

	currentProbes := make(map[ProbeIdentificationPair]*Probe)
	for _, p := range m.Probes {
		pip := ProbeIdentificationPair{UID: p.ProbeIdentificationPair.UID, EBPFFuncName: p.ProbeIdentificationPair.EBPFFuncName}
		if p.Enabled {
			currentProbes[pip] = p
		}
	}

	nextProbes := make(map[ProbeIdentificationPair]bool)
	for _, selector := range selectors {
		for _, id := range selector.GetProbesIdentificationPairList() {
			if !slices.Contains(m.options.ExcludedFunctions, id.EBPFFuncName) {
				pip := ProbeIdentificationPair{UID: id.UID, EBPFFuncName: id.EBPFFuncName}
				nextProbes[pip] = true
			}
		}
	}

	for id := range nextProbes {
		var probe *Probe
		if currentProbe, alreadyPresent := currentProbes[id]; alreadyPresent {
			delete(currentProbes, id)
			probe = currentProbe
		} else {
			var found bool
			probe, found = m.getProbe(id)
			if !found {
				m.stateLock.Unlock()
				return fmt.Errorf("couldn't find program %s: %w", id, ErrUnknownSectionOrFuncName)
			}
			probe.Enabled = true
		}
		if !probe.IsRunning() {
			// ignore all errors, they are already collected per probe and will be surfaced by the
			// activation validators if needed.
			_ = probe.init(m)
			_ = probe.Attach()
		}
	}

	for _, probe := range currentProbes {
		if err := probe.Detach(); err != nil {
			m.stateLock.Unlock()
			return err
		}
		probe.Enabled = false
	}

	// update activated probes & check activation
	m.options.ActivatedProbes = selectors

	m.stateLock.Unlock()

	var validationErrs error
	for _, selector := range selectors {
		if err := selector.RunValidator(m); err != nil {
			validationErrs = multierror.Append(validationErrs, err)
		}
	}

	if validationErrs != nil {
		// Clean up
		_ = m.stop(CleanInternal)
		return fmt.Errorf("probes activation validation failed: %w", validationErrs)
	}

	return nil
}

// editConstants - newEditor the programs in the CollectionSpec with the provided constant editors. Tries with the BTF global
// variable first, and fall back to the asm method if BTF is not available.
func (m *Manager) editConstants() error {
	// Start with the BTF based solution
	rodata := m.collectionSpec.Maps[".rodata"]
	if rodata != nil && rodata.Key != nil {
		for _, editor := range m.options.ConstantEditors {
			if !editor.BTFGlobalConstant {
				continue
			}
			constant := map[string]interface{}{
				editor.Name: editor.Value,
			}
			if err := m.collectionSpec.RewriteConstants(constant); err != nil && editor.FailOnMissing {
				return err
			}
		}
	}

	// Fall back to the old school constant edition
	for _, constantEditor := range m.options.ConstantEditors {
		if constantEditor.BTFGlobalConstant {
			continue
		}

		// newEditor the constant of the provided programs
		for _, id := range constantEditor.ProbeIdentificationPairs {
			programs, found, err := m.GetProgramSpec(id)
			if err != nil {
				return err
			}
			if !found || len(programs) == 0 {
				return fmt.Errorf("couldn't find programSpec %v: %w", id, ErrUnknownSectionOrFuncName)
			}
			prog := programs[0]

			// newEditor program
			if err := m.editConstant(prog, constantEditor); err != nil {
				return fmt.Errorf("couldn't edit %s in %v: %w", constantEditor.Name, id, err)
			}
		}

		// Apply to all programs if no section was provided
		if len(constantEditor.ProbeIdentificationPairs) == 0 {
			for section, prog := range m.collectionSpec.Programs {
				if err := m.editConstant(prog, constantEditor); err != nil {
					return fmt.Errorf("couldn't edit %s in %s: %w", constantEditor.Name, section, err)
				}
			}
		}
	}
	return nil
}

// editMapSpecs - Update the MapSpec with the provided MapSpec editors.
func (m *Manager) editMapSpecs() error {
	for name, mapEditor := range m.options.MapSpecEditors {
		// select the map spec
		spec, exists, err := m.GetMapSpec(name)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("failed to edit maps/%s: couldn't find map: %w", name, ErrUnknownSection)
		}
		if mapEditor.EditorFlag == 0 {
			return fmt.Errorf("failed to edit maps/%s: %w", name, ErrMissingEditorFlags)
		}
		if EditType&mapEditor.EditorFlag == EditType {
			spec.Type = mapEditor.Type
		}
		if EditMaxEntries&mapEditor.EditorFlag == EditMaxEntries {
			spec.MaxEntries = mapEditor.MaxEntries
		}
		if EditFlags&mapEditor.EditorFlag == EditFlags {
			spec.Flags = mapEditor.Flags
		}
		if EditKeyValue&mapEditor.EditorFlag == EditKeyValue {
			spec.Key = mapEditor.Key
			spec.KeySize = mapEditor.KeySize
			spec.Value = mapEditor.Value
			spec.ValueSize = mapEditor.ValueSize
		}
	}
	return nil
}

// editConstant - newEditor the provided program with the provided constant using the asm method.
func (m *Manager) editConstant(prog *ebpf.ProgramSpec, editor ConstantEditor) error {
	edit := newEditor(&prog.Instructions)
	data, ok := (editor.Value).(uint64)
	if !ok {
		return fmt.Errorf("with the asm method, the constant value has to be of type uint64")
	}
	if err := edit.RewriteConstant(editor.Name, data); err != nil {
		if isUnreferencedSymbol(err) && editor.FailOnMissing {
			return err
		}
	}
	return nil
}

// rewriteMaps - Rewrite the provided program spec with the provided maps. failOnError controls if an error should be
// returned when a map couldn't be rewritten.
func (m *Manager) rewriteMaps(program *ebpf.ProgramSpec, eBPFMaps map[string]*ebpf.Map, failOnError bool) error {
	for symbol, eBPFMap := range eBPFMaps {
		err := program.Instructions.AssociateMap(symbol, eBPFMap)
		if err != nil && failOnError {
			return fmt.Errorf("couldn't rewrite map %s: %w", symbol, err)
		}
	}
	return nil
}

// editMaps - RewriteMaps replaces all references to specific maps.
func (m *Manager) editMaps(maps map[string]*ebpf.Map) error {
	// Rewrite maps
	for name, toRewrite := range maps {
		// ignore deprecated usage
		//nolint:staticcheck
		if err := m.collectionSpec.RewriteMaps(map[string]*ebpf.Map{name: toRewrite}); err != nil {
			if m.options.MapEditorsIgnoreMissingMaps {
				// make sure the map is removed from the collectionSpec
				delete(m.collectionSpec.Maps, name)
			} else {
				return err
			}
		}
	}

	// The rewrite operation removed the original maps from the CollectionSpec and will therefore not appear in the
	// Collection, make the mapping with the Manager.Maps now
	found := false
	for name, rwMap := range maps {
		for _, managerMap := range m.Maps {
			if managerMap.Name == name {
				managerMap.array = rwMap
				managerMap.externalMap = true
				managerMap.editedMap = true
				found = true
			}
		}
		for _, perfRing := range m.PerfMaps {
			if perfRing.Name == name {
				perfRing.array = rwMap
				perfRing.externalMap = true
				perfRing.editedMap = true
				found = true
			}
		}
		for _, ringBuffer := range m.RingBuffers {
			if ringBuffer.Name == name {
				ringBuffer.array = rwMap
				ringBuffer.externalMap = true
				ringBuffer.editedMap = true
				found = true
			}
		}
		if !found && !m.options.MapEditorsIgnoreMissingMaps {
			// Create a new entry
			m.Maps = append(m.Maps, &Map{
				array:       rwMap,
				manager:     m,
				externalMap: true,
				editedMap:   true,
				Name:        name,
			})
		}
	}
	return nil
}

// editInnerOuterMapSpecs - Update the inner maps of the maps of maps in the collection spec
func (m *Manager) editInnerOuterMapSpec(spec InnerOuterMapSpec) error {
	// find the outer map
	outerSpec, exists, err := m.GetMapSpec(spec.OuterMapName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("failed to set inner map for maps/%s: couldn't find outer map: %w", spec.OuterMapName, ErrUnknownSection)
	}
	// find the inner map
	innerMap, exists, err := m.GetMapSpec(spec.InnerMapName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("failed to set inner map for maps/%s: couldn't find inner map %s: %w", spec.OuterMapName, spec.InnerMapName, ErrUnknownSection)
	}

	// set inner map
	outerSpec.InnerMap = innerMap
	return nil
}

// loadCollection - Load the eBPF maps and programs in the CollectionSpec. Programs and Maps are pinned when requested.
func (m *Manager) loadCollection() error {
	var err error
	// Load collection
	m.collection, err = ebpf.NewCollectionWithOptions(m.collectionSpec, m.options.VerifierOptions)
	if err != nil {
		var ve *ebpf.VerifierError
		if errors.As(err, &ve) {
			// include error twice to preserve context, and still allow unwrapping if desired
			return fmt.Errorf("verifier error loading eBPF programs: %w\n%+v", err, ve)
		}
		return fmt.Errorf("couldn't load eBPF programs: %w", err)
	}

	// match loaded bpf objects
	if err = m.matchBPFObjects(); err != nil {
		return fmt.Errorf("couldn't match bpf objects: %w", err)
	}

	// Initialize Maps
	for _, managerMap := range m.Maps {
		if err := managerMap.init(m); err != nil {
			return err
		}
	}

	// Initialize PerfMaps
	for _, perfMap := range m.PerfMaps {
		if err := perfMap.init(m); err != nil {
			return err
		}
	}

	// Initialize ring buffers
	for _, ringBuffer := range m.RingBuffers {
		if err := ringBuffer.init(m); err != nil {
			return err
		}
	}

	// Initialize Probes
	for _, probe := range m.Probes {
		// Find program
		if err := probe.init(m); err != nil {
			return err
		}
	}
	return nil
}

// loadPinnedObjects - Loads pinned programs and maps from the bpf virtual file system. If a map is found, the
// CollectionSpec will be edited so that references to that map point to the pinned one. If a program is found, it will
// be detached from the CollectionSpec to avoid loading it twice.
func (m *Manager) loadPinnedObjects() error {
	// Look for pinned maps
	for _, managerMap := range m.Maps {
		if managerMap.PinPath == "" {
			continue
		}
		if err := m.loadPinnedMap(managerMap); err != nil {
			if err == ErrPinnedObjectNotFound {
				continue
			}
			return err
		}
	}

	// Look for pinned perf buffer
	for _, perfMap := range m.PerfMaps {
		if perfMap.PinPath == "" {
			continue
		}
		if err := m.loadPinnedMap(&perfMap.Map); err != nil {
			if err == ErrPinnedObjectNotFound {
				continue
			}
			return err
		}
	}

	// Look for pinned perf buffer
	for _, ringBuffer := range m.RingBuffers {
		if ringBuffer.PinPath == "" {
			continue
		}
		if err := m.loadPinnedMap(&ringBuffer.Map); err != nil {
			if err == ErrPinnedObjectNotFound {
				continue
			}
			return err
		}
	}

	// Look for pinned programs
	for _, prog := range m.Probes {
		if prog.PinPath == "" {
			continue
		}
		if err := m.loadPinnedProgram(prog); err != nil {
			if err == ErrPinnedObjectNotFound {
				continue
			}
			return err
		}
	}
	return nil
}

// loadPinnedMap - Loads a pinned map
func (m *Manager) loadPinnedMap(managerMap *Map) error {
	// Check if the pinned object exists
	if _, err := os.Stat(managerMap.PinPath); err != nil {
		return ErrPinnedObjectNotFound
	}

	pinnedMap, err := ebpf.LoadPinnedMap(managerMap.PinPath, nil)
	if err != nil {
		return fmt.Errorf("couldn't load map %s from %s: %w", managerMap.Name, managerMap.PinPath, err)
	}

	// Replace map in CollectionSpec
	if err := m.editMaps(map[string]*ebpf.Map{managerMap.Name: pinnedMap}); err != nil {
		return err
	}
	managerMap.array = pinnedMap
	managerMap.externalMap = true
	return nil
}

// loadPinnedProgram - Loads a pinned program
func (m *Manager) loadPinnedProgram(prog *Probe) error {
	// Check if the pinned object exists
	if _, err := os.Stat(prog.PinPath); err != nil {
		return ErrPinnedObjectNotFound
	}

	pinnedProg, err := ebpf.LoadPinnedProgram(prog.PinPath, nil)
	if err != nil {
		return fmt.Errorf("couldn't load program %v from %s: %w", prog.ProbeIdentificationPair, prog.PinPath, err)
	}
	prog.program = pinnedProg

	// Detach program from CollectionSpec
	delete(m.collectionSpec.Programs, prog.GetEBPFFuncName())
	return nil
}

// sanityCheck - Checks that the probes and maps of the manager were properly defined
func (m *Manager) sanityCheck() error {
	// Check if map names are unique
	cache := map[string]bool{}
	for _, managerMap := range m.Maps {
		_, ok := cache[managerMap.Name]
		if ok {
			return fmt.Errorf("map %s failed the sanity check: %w", managerMap.Name, ErrMapNameInUse)
		}
		cache[managerMap.Name] = true
	}

	for _, perfMap := range m.PerfMaps {
		_, ok := cache[perfMap.Name]
		if ok {
			return fmt.Errorf("map %s failed the sanity check: %w", perfMap.Name, ErrMapNameInUse)
		}
		cache[perfMap.Name] = true
	}

	for _, ringBuffer := range m.RingBuffers {
		_, ok := cache[ringBuffer.Name]
		if ok {
			return fmt.Errorf("map %s failed the sanity check: %w", ringBuffer.Name, ErrMapNameInUse)
		}
		cache[ringBuffer.Name] = true
	}

	// Check if probes identification pairs are unique, request the usage of CloneProbe otherwise
	cache = map[string]bool{}
	for _, managerProbe := range m.Probes {
		_, ok := cache[managerProbe.ProbeIdentificationPair.String()]
		if ok {
			return fmt.Errorf("%v failed the sanity check: %w", managerProbe.ProbeIdentificationPair, ErrCloneProbeRequired)
		}
		cache[managerProbe.ProbeIdentificationPair.String()] = true
	}
	return nil
}

// CleanupNetworkNamespace - Cleans up all references to the provided network namespace within the manager. This means
// that any TC classifier or XDP probe in that network namespace will be stopped and all opened netlink socket in that
// namespace will be closed.
// WARNING: Don't forget to call this method if you've provided a IfIndexNetns and IfIndexNetnsID to one of the probes
// of this manager. Failing to call this cleanup function may lead to leaking the network namespace. Only call this
// function when you're sure that the manager no longer needs to perform anything in the provided network namespace (or
// else call NewNetlinkSocket first).
func (m *Manager) CleanupNetworkNamespace(nsID uint32) error {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()
	if m.state < initialized {
		return ErrManagerNotInitialized
	}

	var err error
	var toDelete []int
	for i, probe := range m.Probes {
		if probe.IfIndexNetnsID != nsID {
			continue
		}

		// disable probe
		probe.Enabled = false

		// stop the probe
		err = concatErrors(err, probe.Stop())

		// append probe to delete (biggest indexes first)
		toDelete = append([]int{i}, toDelete...)
	}

	// delete all netlink sockets, along with netns handles
	m.netlinkSocketCache.remove(nsID)

	// delete probes
	for _, i := range toDelete {
		// we delete the biggest indexes first, so we should be good to go !
		m.Probes = append(m.Probes[:i], m.Probes[i+1:]...)
	}
	return err
}

type procMask uint8

const (
	Running procMask = iota
	Exited
)

// getUIDSet - Returns the list of UIDs used by this manager.
func (m *Manager) getUIDSet() []string {
	var uidSet []string
	for _, p := range m.Probes {
		if len(p.UID) == 0 {
			continue
		}

		var found bool
		for _, uid := range uidSet {
			if uid == p.UID {
				found = true
				break
			}
		}
		if !found {
			uidSet = append(uidSet, p.UID)
		}
	}
	return uidSet
}

func (m *Manager) getTracefsRegex() (*regexp.Regexp, error) {
	uidSet := m.getUIDSet()
	escapedUIDs := make([]string, len(uidSet))
	for i, uid := range uidSet {
		escapedUIDs[i] = regexp.QuoteMeta(uid)
	}
	stringPattern := fmt.Sprintf(`(p|r)[0-9]*:(kprobes|uprobes)\/(.*(%s)*_([0-9]*)) .*`, strings.Join(escapedUIDs, "|"))
	re, err := regexp.Compile(stringPattern)
	if err != nil {
		return nil, fmt.Errorf("event name pattern (%q) generation failed: %w", stringPattern, err)
	}
	return re, nil
}

// cleanupKprobeEvents - Cleans up kprobe_events and uprobe_events by removing entries of known UIDs, that are not used
// anymore.
//
// Previous instances of this manager might have been killed unexpectedly. When this happens,
// kprobe_events is not cleaned up properly and can grow indefinitely until it reaches 65k
// entries (see: https://elixir.bootlin.com/linux/v5.6.1/source/kernel/trace/trace_output.c#L699)
// Once the limit is reached, the kernel refuses to load new probes and throws a "no such device"
// error. To prevent this, start by cleaning up the kprobe_events entries of previous managers that
// are not running anymore.
func (m *Manager) cleanupTracefs() error {
	// build the pattern to look for in kprobe_events and uprobe_events
	pattern, err := m.getTracefsRegex()
	if err != nil {
		return err
	}
	// clean up kprobe_events
	var cleanUpErrors *multierror.Error
	pidMask := map[int]procMask{Getpid(): Running}
	cleanUpErrors = multierror.Append(cleanUpErrors, cleanupKprobeEvents(pattern, pidMask))
	cleanUpErrors = multierror.Append(cleanUpErrors, cleanupUprobeEvents(pattern, pidMask))
	if cleanUpErrors.Len() == 0 {
		return nil
	}
	return cleanUpErrors
}

func cleanupKprobeEvents(pattern *regexp.Regexp, pidMask map[int]procMask) error {
	kprobeEvents, err := readKprobeEvents()
	if err != nil {
		return fmt.Errorf("couldn't read kprobe_events: %w", err)
	}
	var cleanUpErrors error
	for _, match := range pattern.FindAllStringSubmatch(kprobeEvents, -1) {
		if len(match) < 6 {
			continue
		}

		// check if the provided pid still exists
		pid, err := strconv.Atoi(match[5])
		if err != nil {
			continue
		}
		if state, ok := pidMask[pid]; !ok {
			// this short sleep is used to avoid a CPU spike (5s ~ 60k * 80 microseconds)
			time.Sleep(80 * time.Microsecond)

			_, err = process.NewProcess(int32(pid))
			if err == nil {
				// the process is still running, continue
				pidMask[pid] = Running
				continue
			} else {
				pidMask[pid] = Exited
			}
		} else {
			if state == Running {
				// the process is still running, continue
				continue
			}
		}

		// remove the entry
		cleanUpErrors = multierror.Append(cleanUpErrors, unregisterKprobeEventWithEventName(match[3]))
	}
	return cleanUpErrors
}

func cleanupUprobeEvents(pattern *regexp.Regexp, pidMask map[int]procMask) error {
	uprobeEvents, err := readUprobeEvents()
	if err != nil {
		return fmt.Errorf("couldn't read uprobe_events: %w", err)
	}
	var cleanUpErrors error
	for _, match := range pattern.FindAllStringSubmatch(uprobeEvents, -1) {
		if len(match) < 6 {
			continue
		}

		// check if the provided pid still exists
		pid, err := strconv.Atoi(match[5])
		if err != nil {
			continue
		}
		if state, ok := pidMask[pid]; !ok {
			// this short sleep is used to avoid a CPU spike (5s ~ 60k * 80 microseconds)
			time.Sleep(80 * time.Microsecond)

			_, err = process.NewProcess(int32(pid))
			if err == nil {
				// the process is still running, continue
				pidMask[pid] = Running
				continue
			} else {
				pidMask[pid] = Exited
			}
		} else {
			if state == Running {
				// the process is still running, continue
				continue
			}
		}

		// remove the entry
		cleanUpErrors = multierror.Append(cleanUpErrors, unregisterUprobeEventWithEventName(match[3]))
	}
	return cleanUpErrors
}

func (m *Manager) GetNetlinkSocket(nsHandle uint64, nsID uint32) (*NetlinkSocket, error) {
	return m.netlinkSocketCache.getNetlinkSocket(nsHandle, nsID)
}

type netlinkSocketCache struct {
	sync.Mutex
	cache map[uint32]*NetlinkSocket
}

func newNetlinkSocketCache() *netlinkSocketCache {
	return &netlinkSocketCache{
		cache: make(map[uint32]*NetlinkSocket),
	}
}

// getNetlinkSocket - Returns a netlink socket in the requested network namespace from cache or creates a new one.
// TC classifiers are attached by creating a qdisc on the requested interface. A netlink socket
// is required to create a qdisc (or to attach an XDP program to an interface). Since this socket can be re-used for
// multiple probes, instantiate the connection at the manager level and cache the netlink socket. The provided nsID
// should be the ID of the network namespaced returned by a readlink on `/proc/[pid]/ns/net` for a [pid] that lives in
// the network namespace pointed to by the nsHandle.
func (nsc *netlinkSocketCache) getNetlinkSocket(nsHandle uint64, nsID uint32) (*NetlinkSocket, error) {
	nsc.Lock()
	defer nsc.Unlock()

	sock, ok := nsc.cache[nsID]
	if ok {
		return sock, nil
	}

	cacheEntry, err := NewNetlinkSocket(nsHandle)
	if err != nil {
		return nil, fmt.Errorf("namespace %v: %w", nsID, err)
	}

	nsc.cache[nsID] = cacheEntry
	return cacheEntry, nil
}

// cleanup - Cleans up all opened netlink sockets in cache. This function is expected to be called when a
// manager is stopped.
func (nsc *netlinkSocketCache) cleanup() {
	nsc.Lock()
	defer nsc.Unlock()

	for key, s := range nsc.cache {
		delete(nsc.cache, key)
		// close the netlink socket
		s.Sock.Close()
	}
}

func (nsc *netlinkSocketCache) remove(nsID uint32) {
	s, ok := nsc.cache[nsID]
	if ok {
		delete(nsc.cache, nsID)

		// close the netlink socket
		s.Sock.Close()
	}
}
