package manager

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/cilium/ebpf"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	"github.com/DataDog/gopsutil/process"
)

// XdpAttachMode selects a way how XDP program will be attached to interface
type XdpAttachMode int

const (
	// XdpAttachModeNone stands for "best effort" - the kernel automatically
	// selects the best mode (would try Drv first, then fallback to Generic).
	// NOTE: Kernel will not fall back to Generic XDP if NIC driver failed
	//       to install XDP program.
	XdpAttachModeNone XdpAttachMode = 0
	// XdpAttachModeSkb is "generic", kernel mode, less performant comparing to native,
	// but does not requires driver support.
	XdpAttachModeSkb XdpAttachMode = 1 << 1
	// XdpAttachModeDrv is native, driver mode (support from driver side required)
	XdpAttachModeDrv XdpAttachMode = 1 << 2
	// XdpAttachModeHw suitable for NICs with hardware XDP support
	XdpAttachModeHw XdpAttachMode = 1 << 3
	// DefaultTCFilterPriority is the default TC filter priority if none were given
	DefaultTCFilterPriority = 50
)

type TrafficType uint16

func (tt TrafficType) String() string {
	switch tt {
	case Ingress:
		return "ingress"
	case Egress:
		return "egress"
	default:
		return fmt.Sprintf("TrafficType(%d)", tt)
	}
}

const (
	Ingress          = TrafficType(uint16(netlink.HANDLE_MIN_INGRESS & 0x0000FFFF))
	Egress           = TrafficType(uint16(netlink.HANDLE_MIN_EGRESS & 0x0000FFFF))
	clsactQdisc      = uint16(netlink.HANDLE_INGRESS >> 16)
	UnknownProbeType = ""
	ProbeType        = "p"
	RetProbeType     = "r"
)

const (
	BpfFlagActDirect = uint32(1) // see TCA_BPF_FLAG_ACT_DIRECT
)

type ProbeIdentificationPair struct {
	kprobeType string

	// UID - (optional) this field can be used to identify your probes when the same eBPF program is used on multiple
	// hook points. Keep in mind that the pair (probe section, probe UID) needs to be unique
	// system-wide for the kprobes and uprobes registration to work.
	UID string

	// EBPFFuncName - Name of the main eBPF function of your eBPF program.
	EBPFFuncName string

	// EBPFSection - Section in which EBPFFuncName lives.
	//
	// Deprecated: Only EBPFFuncName is necessary
	EBPFSection string
}

func (pip ProbeIdentificationPair) String() string {
	return fmt.Sprintf("{UID:%s EBPFFuncName:%s EBPFSection:%s}", pip.UID, pip.EBPFFuncName, pip.EBPFSection)
}

// Matches - Returns true if the identification pair (probe uid, probe section, probe func name) matches.
func (pip ProbeIdentificationPair) Matches(id ProbeIdentificationPair) bool {
	return pip.UID == id.UID && pip.EBPFDefinitionMatches(id)
}

// EBPFDefinitionMatches - Returns true if the eBPF definition matches.
func (pip ProbeIdentificationPair) EBPFDefinitionMatches(id ProbeIdentificationPair) bool {
	return pip.EBPFFuncName == id.EBPFFuncName
}

// GetKprobeType - Identifies the probe type of the provided KProbe section
func (p *Probe) GetKprobeType() string {
	if len(p.kprobeType) == 0 {
		if strings.HasPrefix(p.programSpec.SectionName, "kretprobe/") {
			p.kprobeType = RetProbeType
		} else if strings.HasPrefix(p.programSpec.SectionName, "kprobe/") {
			p.kprobeType = ProbeType
		} else {
			p.kprobeType = UnknownProbeType
		}
	}
	return p.kprobeType
}

// GetUprobeType - Identifies the probe type of the provided Uprobe section
func (p *Probe) GetUprobeType() string {
	if len(p.kprobeType) == 0 {
		if strings.HasPrefix(p.programSpec.SectionName, "uretprobe/") {
			p.kprobeType = RetProbeType
		} else if strings.HasPrefix(p.programSpec.SectionName, "uprobe/") {
			p.kprobeType = ProbeType
		} else {
			p.kprobeType = UnknownProbeType
		}
	}
	return p.kprobeType
}

type AttachMethod uint32

const (
	AttachMethodNotSet AttachMethod = iota
	AttachWithPerfEventOpen
	AttachWithProbeEvents
)

type KprobeAttachMethod = AttachMethod

const (
	AttachKprobeMethodNotSet      = AttachMethodNotSet
	AttachKprobeWithPerfEventOpen = AttachWithPerfEventOpen
	AttachKprobeWithKprobeEvents  = AttachWithProbeEvents
)

// Probe - Main eBPF probe wrapper. This structure is used to store the required data to attach a loaded eBPF
// program to its hook point.
type Probe struct {
	netlinkSocketCache      *netlinkSocketCache
	program                 *ebpf.Program
	programSpec             *ebpf.ProgramSpec
	perfEventFD             *fd
	rawTracepointFD         *fd
	state                   state
	stateLock               sync.RWMutex
	manualLoadNeeded        bool
	checkPin                bool
	attachPID               int
	attachRetryAttempt      uint
	attachedWithDebugFS     bool
	kprobeHookPointNotExist bool
	systemWideID            int
	programTag              string
	link                    netlink.Link
	tcFilter                netlink.BpfFilter
	tcClsActQdisc           netlink.Qdisc

	// lastError - stores the last error that the probe encountered, it is used to surface a more useful error message
	// when one of the validators (see Options.ActivatedProbes) fails.
	lastError error

	// ProbeIdentificationPair is used to identify the current probe
	ProbeIdentificationPair

	// CopyProgram - When enabled, this option will make a unique copy of the program section for the current program
	CopyProgram bool

	// KeepProgramSpec - Defines if the internal *ProgramSpec should be cleaned up after the probe has been successfully
	// attached to free up memory. If you intend to make a copy of this Probe later, you should explicitly set this
	// option to true.
	KeepProgramSpec bool

	// SyscallFuncName - Name of the syscall on which the program should be hooked. As the exact kernel symbol may
	// differ from one kernel version to the other, the right prefix will be computed automatically at runtime.
	// If a syscall name is not provided, the section name (without its probe type prefix) is assumed to be the
	// hook point.
	SyscallFuncName string

	// MatchFuncName - Pattern used to find the function(s) to attach to
	// FOR KPROBES: When this field is used, the provided pattern is matched against the list of available symbols
	// in /sys/kernel/debug/tracing/available_filter_functions. If the exact function does not exist, then the first
	// symbol matching the provided pattern will be used. This option requires debugfs.
	//
	// FOR UPROBES: When this field is used, the provided pattern is matched against the list of symbols in the symbol
	// table of the provided elf binary. If the exact function does not exist, then the first symbol matching the
	// provided pattern will be used.
	MatchFuncName string

	// HookFuncName - Exact name of the symbol to hook onto. When this field is set, MatchFuncName and SyscallFuncName
	// are ignored.
	HookFuncName string

	// TracepointCategory - (Tracepoint) The manager expects the tracepoint category to be parsed from the eBPF section
	// in which the eBPF function of this Probe lives (SEC("tracepoint/[category]/[name])). However you can use this
	// field to override it.
	TracepointCategory string

	// TracepointName - (Tracepoint) The manager expects the tracepoint name to be parsed from the eBPF section
	// in which the eBPF function of this Probe lives (SEC("tracepoint/[category]/[name])). However you can use this
	// field to override it.
	TracepointName string

	// Enabled - Indicates if a probe should be enabled or not. This parameter can be set at runtime using the
	// Manager options (see ActivatedProbes)
	Enabled bool

	// PinPath - Once loaded, the eBPF program will be pinned to this path. If the eBPF program has already been pinned
	// and is already running in the kernel, then it will be loaded from this path.
	PinPath string

	// KProbeMaxActive - (kretprobes) With kretprobes, you can configure the maximum number of instances of the function that can be
	// probed simultaneously with maxactive. If maxactive is 0 it will be set to the default value: if CONFIG_PREEMPT is
	// enabled, this is max(10, 2*NR_CPUS); otherwise, it is NR_CPUS. For kprobes, maxactive is ignored.
	KProbeMaxActive int

	// KprobeAttachMethod - Method to use for attaching the kprobe. Either use perfEventOpen ABI or kprobe events
	KprobeAttachMethod KprobeAttachMethod

	// UprobeAttachMethod - Method to use for attaching the uprobe. Either use perfEventOpen ABI or uprobe events
	UprobeAttachMethod AttachMethod

	// UprobeOffset - If UprobeOffset is provided, the uprobe will be attached to it directly without looking for the
	// symbol in the elf binary. If the file is a non-PIE executable, the provided address must be a virtual address,
	// otherwise it must be an offset relative to the file load address.
	UprobeOffset uint64

	// ProbeRetry - Defines the number of times that the probe will retry to attach / detach on error.
	ProbeRetry uint

	// ProbeRetryDelay - Defines the delay to wait before the probe should retry to attach / detach on error.
	ProbeRetryDelay time.Duration

	// BinaryPath - (uprobes) A Uprobe is attached to a specific symbol in a user space binary. The offset is
	// automatically computed for the symbol name provided in the uprobe section ( SEC("uprobe/[symbol_name]") ).
	BinaryPath string

	// CGroupPath - (cgroup family programs) All CGroup programs are attached to a CGroup (v2). This field provides the
	// path to the CGroup to which the probe should be attached. The attach type is determined by the section.
	CGroupPath string

	// SocketFD - (socket filter) Socket filter programs are bound to a socket and filter the packets they receive
	// before they reach user space. The probe will be bound to the provided file descriptor
	SocketFD int

	// IfIndex - (TC classifier & XDP) Interface index used to identify the interface on which the probe will be
	// attached. If not set, fall back to Ifname.
	IfIndex int

	// IfName - (TC Classifier & XDP) Interface name on which the probe will be attached.
	IfName string

	// IfIndexNetns - (TC Classifier & XDP) Network namespace in which the network interface lives. If this value is
	// provided, then IfIndexNetnsID is required too.
	// WARNING: it is up to the caller of "Probe.Start()" to close this netns handle. Failing to close this handle may
	// lead to leaking the network namespace. This handle can be safely closed once "Probe.Start()" returns.
	IfIndexNetns uint64

	// IfIndexNetnsID - (TC Classifier & XDP) Network namespace ID associated of the IfIndexNetns handle. If this value
	// is provided, then IfIndexNetns is required too.
	// WARNING: it is up to the caller of "Probe.Start()" to call "manager.CleanupNetworkNamespace()" once the provided
	// IfIndexNetnsID is no longer needed. Failing to call this cleanup function may lead to leaking the network
	// namespace. Remember that "manager.CleanupNetworkNamespace()" will close the netlink socket opened with the netns
	// handle provided above. If you want to start the probe again, you'll need to provide a new valid netns handle so
	// that a new netlink socket can be created in that namespace.
	IfIndexNetnsID uint32

	// XDPAttachMode - (XDP) XDP attach mode. If not provided the kernel will automatically select the best available
	// mode.
	XDPAttachMode XdpAttachMode

	// NetworkDirection - (TC classifier) Network traffic direction of the classifier. Can be either Ingress or Egress. Keep
	// in mind that if you are hooking on the host side of a virtuel ethernet pair, Ingress and Egress are inverted.
	NetworkDirection TrafficType

	// TCFilterHandle - (TC classifier) defines the handle to use when loading the classifier. Leave unset to let the kernel decide which handle to use.
	TCFilterHandle uint32

	// TCFilterPrio - (TC classifier) defines the priority of the classifier added to the clsact qdisc. Defaults to DefaultTCFilterPriority.
	TCFilterPrio uint16

	// TCCleanupQDisc - (TC classifier) defines if the manager should clean up the clsact qdisc when a probe is unloaded
	TCCleanupQDisc bool

	// TCFilterProtocol - (TC classifier) defines the protocol to match in order to trigger the classifier. Defaults to
	// ETH_P_ALL.
	TCFilterProtocol uint16

	// SamplePeriod - (Perf event) This parameter defines when the perf_event eBPF program is triggered. When SamplePeriod > 0
	// the program will be triggered every SamplePeriod events.
	SamplePeriod int

	// SampleFrequency - (Perf event) This parameter defines when the perf_event eBPF program is triggered. When
	// SampleFrequency > 0, SamplePeriod is ignored and the eBPF program is triggered at the requested frequency.
	SampleFrequency int

	// PerfEventType - (Perf event) This parameter defines the type of the perf_event program. Allowed values are
	// unix.PERF_TYPE_HARDWARE and unix.PERF_TYPE_SOFTWARE
	PerfEventType int

	// PerfEventPID - (Perf event, uprobes) This parameter defines the PID for which the program should be triggered.
	// Do not set this value to monitor the whole host.
	PerfEventPID int

	// PerfEventConfig - (Perf event) This parameter defines which software or hardware event is being monitored. See the
	// PERF_COUNT_SW_* and PERF_COUNT_HW_* constants in the unix package.
	PerfEventConfig int

	// PerfEventCPUCount - (Perf event) This parameter defines the number of CPUs to monitor. If not set, defaults to
	// runtime.NumCPU(). Disclaimer: in containerized environment and depending on the CPU affinity of the program
	// holding the manager, runtime.NumCPU might not return the real CPU count of the host.
	PerfEventCPUCount int

	// perfEventCPUFDs - (Perf event) holds the fd of the perf_event program per CPU
	perfEventCPUFDs []*fd
}

// GetEBPFFuncName - Returns EBPFFuncName with the UID as a postfix if the Probe was copied
func (p *Probe) GetEBPFFuncName() string {
	if p.CopyProgram {
		return fmt.Sprintf("%s_%s", p.EBPFFuncName, p.UID)
	}
	return p.EBPFFuncName
}

// Copy - Returns a copy of the current probe instance. Only the exported fields are copied.
func (p *Probe) Copy() *Probe {
	return &Probe{
		ProbeIdentificationPair: ProbeIdentificationPair{
			UID:          p.UID,
			EBPFFuncName: p.EBPFFuncName,
		},
		SyscallFuncName:    p.SyscallFuncName,
		CopyProgram:        p.CopyProgram,
		KeepProgramSpec:    p.KeepProgramSpec,
		SamplePeriod:       p.SamplePeriod,
		SampleFrequency:    p.SampleFrequency,
		PerfEventType:      p.PerfEventType,
		PerfEventPID:       p.PerfEventPID,
		PerfEventConfig:    p.PerfEventConfig,
		MatchFuncName:      p.MatchFuncName,
		TracepointCategory: p.TracepointCategory,
		TracepointName:     p.TracepointName,
		Enabled:            p.Enabled,
		PinPath:            p.PinPath,
		KProbeMaxActive:    p.KProbeMaxActive,
		BinaryPath:         p.BinaryPath,
		CGroupPath:         p.CGroupPath,
		SocketFD:           p.SocketFD,
		IfIndex:            p.IfIndex,
		IfName:             p.IfName,
		IfIndexNetns:       p.IfIndexNetns,
		IfIndexNetnsID:     p.IfIndexNetnsID,
		XDPAttachMode:      p.XDPAttachMode,
		NetworkDirection:   p.NetworkDirection,
		ProbeRetry:         p.ProbeRetry,
		ProbeRetryDelay:    p.ProbeRetryDelay,
		TCFilterProtocol:   p.TCFilterProtocol,
		TCFilterPrio:       p.TCFilterPrio,
	}
}

// GetLastError - Returns the last error that the probe encountered
func (p *Probe) GetLastError() error {
	return p.lastError
}

// ID returns the system-wide unique ID for this program
func (p *Probe) ID() uint32 {
	return uint32(p.systemWideID)
}

// IsRunning - Returns true if the probe was successfully initialized, started and is currently running or paused.
func (p *Probe) IsRunning() bool {
	p.stateLock.RLock()
	defer p.stateLock.RUnlock()
	return p.state == running || p.state == paused
}

// IsInitialized - Returns true if the probe was successfully initialized.
func (p *Probe) IsInitialized() bool {
	p.stateLock.RLock()
	defer p.stateLock.RUnlock()
	return p.state >= initialized
}

// RenameProbeIdentificationPair - Renames the probe identification pair of a probe
func (p *Probe) RenameProbeIdentificationPair(newID ProbeIdentificationPair) error {
	p.stateLock.Lock()
	defer p.stateLock.Unlock()
	if p.state >= paused {
		return fmt.Errorf("couldn't rename ProbeIdentificationPair of %s with %s: %w", p.ProbeIdentificationPair, newID, ErrProbeRunning)
	}
	p.UID = newID.UID
	return nil
}

// Test - Triggers the probe with the provided test data. Returns the length of the output, the raw output or an error.
func (p *Probe) Test(in []byte) (uint32, []byte, error) {
	return p.program.Test(in)
}

// Benchmark - Benchmark runs the Program with the given input for a number of times and returns the time taken per
// iteration.
//
// Returns the result of the last execution of the program and the time per run or an error. reset is called whenever
// the benchmark syscall is interrupted, and should be set to testing.B.ResetTimer or similar.
func (p *Probe) Benchmark(in []byte, repeat int, reset func()) (uint32, time.Duration, error) {
	return p.program.Benchmark(in, repeat, reset)
}

// initWithOptions - Initializes a probe with options
func (p *Probe) initWithOptions(manager *Manager, manualLoadNeeded bool, checkPin bool) error {
	if !p.Enabled {
		return nil
	}

	p.stateLock.Lock()
	defer p.stateLock.Unlock()
	p.manualLoadNeeded = manualLoadNeeded
	p.checkPin = checkPin
	return p.internalInit(manager)
}

// init - Initialize a probe
func (p *Probe) init(manager *Manager) error {
	if !p.Enabled {
		return nil
	}
	p.stateLock.Lock()
	defer p.stateLock.Unlock()
	return p.internalInit(manager)
}

func (p *Probe) Program() *ebpf.Program {
	return p.program
}

func (p *Probe) internalInit(manager *Manager) error {
	if p.state >= initialized {
		return nil
	}

	p.netlinkSocketCache = manager.netlinkSocketCache
	p.state = reset
	// Load spec if necessary
	if p.manualLoadNeeded {
		prog, err := ebpf.NewProgramWithOptions(p.programSpec, manager.options.VerifierOptions.Programs)
		if err != nil {
			p.lastError = err
			var ve *ebpf.VerifierError
			if errors.As(err, &ve) {
				// include error twice to preserve context, and still allow unwrapping if desired
				return fmt.Errorf("verifier error loading new probe %v: %w\n%+v", p.ProbeIdentificationPair, err, ve)
			}
			return fmt.Errorf("couldn't load new probe %v: %w", p.ProbeIdentificationPair, err)
		}
		p.program = prog
	}

	// Retrieve eBPF program if one isn't already set
	if p.program == nil {
		if p.program, p.lastError = manager.getProbeProgram(p.GetEBPFFuncName()); p.lastError != nil {
			return fmt.Errorf("couldn't find program %s: %w", p.GetEBPFFuncName(), ErrUnknownSectionOrFuncName)
		}
		p.checkPin = true
	}

	if p.programSpec == nil {
		if p.programSpec, p.lastError = manager.getProbeProgramSpec(p.GetEBPFFuncName()); p.lastError != nil {
			return fmt.Errorf("couldn't find program spec %s: %w", p.GetEBPFFuncName(), ErrUnknownSectionOrFuncName)
		}
	}

	if p.programSpec.Type == ebpf.SchedCLS {
		// sanity check
		if p.NetworkDirection != Egress && p.NetworkDirection != Ingress {
			return fmt.Errorf("%s has an invalid configuration: %w", p.ProbeIdentificationPair, ErrNoNetworkDirection)
		}
	}

	if p.checkPin {
		// Pin program if needed
		if p.PinPath != "" {
			if err := p.program.Pin(p.PinPath); err != nil {
				p.lastError = err
				return fmt.Errorf("couldn't pin program %s at %s: %w", p.GetEBPFFuncName(), p.PinPath, err)
			}
		}
		p.checkPin = false
	}

	// Update syscall function name with the correct arch prefix
	if p.SyscallFuncName != "" && len(p.HookFuncName) == 0 {
		var err error
		p.HookFuncName, err = GetSyscallFnNameWithSymFile(p.SyscallFuncName, manager.options.SymFile)
		if err != nil {
			p.lastError = err
			return err
		}
	}

	// Find function name match if required
	if p.MatchFuncName != "" && len(p.HookFuncName) == 0 {
		// if this is a kprobe or a kretprobe, look for the symbol now
		if p.GetKprobeType() != UnknownProbeType {
			var err error
			p.HookFuncName, err = FindFilterFunction(p.MatchFuncName)
			if err != nil {
				p.lastError = err
				return err
			}
		}
	}

	if len(p.HookFuncName) == 0 {
		// default back to the AttachTo field in Program, as parsed by Cilium
		p.HookFuncName = p.programSpec.AttachTo
	}

	// resolve netns ID from netns handle
	if p.IfIndexNetns == 0 && p.IfIndexNetnsID != 0 || p.IfIndexNetns != 0 && p.IfIndexNetnsID == 0 {
		return fmt.Errorf("both IfIndexNetns and IfIndexNetnsID are required if one is provided (IfIndexNetns: %d IfIndexNetnsID: %d)", p.IfIndexNetns, p.IfIndexNetnsID)
	}

	// set default TC classifier priority
	if p.TCFilterPrio == 0 {
		p.TCFilterPrio = DefaultTCFilterPriority
	}

	// set default TC classifier protocol
	if p.TCFilterProtocol == 0 {
		p.TCFilterProtocol = unix.ETH_P_ALL
	}

	// Default max active value
	if p.KProbeMaxActive == 0 {
		p.KProbeMaxActive = manager.options.DefaultKProbeMaxActive
	}

	// Default retry
	if p.ProbeRetry == 0 {
		p.ProbeRetry = manager.options.DefaultProbeRetry
	}

	// Default retry delay
	if p.ProbeRetryDelay == 0 {
		p.ProbeRetryDelay = manager.options.DefaultProbeRetryDelay
	}

	// fetch system-wide program ID, if the feature is available
	if p.program != nil {
		programInfo, err := p.program.Info()
		if err == nil {
			p.programTag = programInfo.Tag
			id, available := programInfo.ID()
			if available {
				p.systemWideID = int(id)
			}
		}
	}

	// set default kprobe attach method
	if p.KprobeAttachMethod == AttachKprobeMethodNotSet {
		if manager != nil {
			p.KprobeAttachMethod = manager.options.DefaultKprobeAttachMethod
		}
		if p.KprobeAttachMethod == AttachKprobeMethodNotSet {
			p.KprobeAttachMethod = AttachKprobeWithPerfEventOpen
		}
	}

	// set default uprobe attach method
	if p.UprobeAttachMethod == AttachMethodNotSet {
		if manager != nil {
			p.UprobeAttachMethod = manager.options.DefaultUprobeAttachMethod
		}
		if p.UprobeAttachMethod == AttachMethodNotSet {
			p.UprobeAttachMethod = AttachWithPerfEventOpen
		}
	}

	// update probe state
	p.state = initialized
	return nil
}

// ResolveLink - Resolves the Probe's network interface
func (p *Probe) ResolveLink() (netlink.Link, error) {
	return p.resolveLink()
}

func (p *Probe) resolveLink() (netlink.Link, error) {
	if p.link != nil {
		return p.link, nil
	}

	// get a netlink socket in the probe network namespace
	ntl, err := p.getNetlinkSocket()
	if err != nil {
		return nil, err
	}

	if p.IfIndex > 0 {
		p.link, err = ntl.Sock.LinkByIndex(p.IfIndex)
		if err != nil {
			return nil, fmt.Errorf("couldn't resolve interface with IfIndex %d in namespace %d: %w", p.IfIndex, p.IfIndexNetnsID, err)
		}
	} else if len(p.IfName) > 0 {
		p.link, err = ntl.Sock.LinkByName(p.IfName)
		if err != nil {
			return nil, fmt.Errorf("couldn't resolve interface with IfName %s in namespace %d: %w", p.IfName, p.IfIndexNetnsID, err)
		}
	} else {
		return nil, ErrInterfaceNotSet
	}

	attrs := p.link.Attrs()
	if attrs != nil {
		p.IfIndex = attrs.Index
		p.IfName = attrs.Name
	}

	return p.link, nil
}

// Attach - Attaches the probe to the right hook point in the kernel depending on the program type and the provided
// parameters.
func (p *Probe) Attach() error {
	return retry.Do(func() error {
		p.attachRetryAttempt++
		err := p.attach()
		if err == nil {
			return nil
		}

		// not available, not a temporary error
		if errors.Is(err, syscall.ENOENT) || errors.Is(err, syscall.EINVAL) {
			return nil
		}

		return err
	}, retry.Attempts(p.getRetryAttemptCount()), retry.Delay(p.ProbeRetryDelay), retry.LastErrorOnly(true))
}

func (p *Probe) Pause() error {
	return p.pause()
}

func (p *Probe) Resume() error {
	return p.resume()
}

// attach - Thread unsafe version of attach
func (p *Probe) attach() error {
	p.stateLock.Lock()
	defer p.stateLock.Unlock()
	if p.state >= paused || !p.Enabled {
		return nil
	}
	if p.state < initialized {
		if p.lastError == nil {
			p.lastError = ErrProbeNotInitialized
		}
		return ErrProbeNotInitialized
	}

	p.attachPID = Getpid()

	// Per program type start
	var err error
	switch p.programSpec.Type {
	case ebpf.UnspecifiedProgram:
		err = fmt.Errorf("invalid program type, make sure to use the right section prefix: %w", ErrSectionFormat)
	case ebpf.Kprobe:
		err = p.attachKprobe()
	case ebpf.TracePoint:
		err = p.attachTracepoint()
	case ebpf.RawTracepoint, ebpf.RawTracepointWritable:
		err = p.attachRawTracepoint()
	case ebpf.CGroupDevice, ebpf.CGroupSKB, ebpf.CGroupSock, ebpf.CGroupSockAddr, ebpf.CGroupSockopt, ebpf.CGroupSysctl:
		err = p.attachCGroup()
	case ebpf.SocketFilter:
		err = p.attachSocket()
	case ebpf.SchedCLS:
		err = p.attachTCCLS()
	case ebpf.XDP:
		err = p.attachXDP()
	case ebpf.LSM:
		err = p.attachLSM()
	case ebpf.PerfEvent:
		err = p.attachPerfEvent()
	case ebpf.Tracing:
		err = p.attachTracing()
	default:
		err = fmt.Errorf("program type %s not implemented yet", p.programSpec.Type)
	}
	if err != nil {
		p.lastError = err
		// Clean up any progress made in the attach attempt
		_ = p.stop(false)
		return fmt.Errorf("couldn't start probe %s: %w", p.ProbeIdentificationPair, err)
	}

	// update probe state
	p.state = running
	p.attachRetryAttempt = p.getRetryAttemptCount()

	// cleanup ProgramSpec to free up some memory
	p.cleanupProgramSpec()
	return nil
}

// cleanupProgramSpec - Cleans up the internal ProgramSpec attribute to free up some memory
func (p *Probe) cleanupProgramSpec() {
	if p.KeepProgramSpec {
		return
	}
	cleanupProgramSpec(p.programSpec)
}

func (p *Probe) pause() error {
	p.stateLock.Lock()
	defer p.stateLock.Unlock()
	if p.state <= paused || !p.Enabled {
		return nil
	}

	var err error
	switch p.programSpec.Type {
	case ebpf.Kprobe:
		err = p.pauseKprobe()
	case ebpf.SocketFilter:
		err = p.detachSocket()
	case ebpf.TracePoint:
		err = p.pauseTracepoint()
	default:
		return fmt.Errorf("pause not supported for program type %s", p.programSpec.Type)
	}
	if err != nil {
		p.lastError = err
		return fmt.Errorf("error pausing probe %s: %w", p.ProbeIdentificationPair, err)
	}

	p.state = paused
	return nil
}

func (p *Probe) resume() error {
	p.stateLock.Lock()
	defer p.stateLock.Unlock()
	if p.state != paused || !p.Enabled {
		return nil
	}

	var err error
	switch p.programSpec.Type {
	case ebpf.Kprobe:
		err = p.resumeKprobe()
	case ebpf.SocketFilter:
		err = p.attachSocket()
	case ebpf.TracePoint:
		err = p.resumeTracepoint()
	default:
		return fmt.Errorf("resume not supported for program type %s", p.programSpec.Type)
	}
	if err != nil {
		p.lastError = err
		return fmt.Errorf("error resuming probe %s: %w", p.ProbeIdentificationPair, err)
	}

	p.state = running
	return nil
}

// Detach - Detaches the probe from its hook point depending on the program type and the provided parameters. This
// method does not close the underlying eBPF program, which means that Attach can be called again later.
func (p *Probe) Detach() error {
	p.stateLock.Lock()
	defer p.stateLock.Unlock()
	if p.state < paused || !p.Enabled {
		return nil
	}

	// detach from hook point
	err := p.detachRetry()

	// update state of the probe
	if err != nil {
		p.lastError = err
	} else {
		p.state = initialized
	}

	return err
}

// detachRetry - Thread unsafe version of Detach with retry
func (p *Probe) detachRetry() error {
	return retry.Do(p.detach, retry.Attempts(p.getRetryAttemptCount()), retry.Delay(p.ProbeRetryDelay), retry.LastErrorOnly(true))
}

// detach - Thread unsafe version of Detach.
func (p *Probe) detach() error {
	var err error
	// Remove pin if needed
	if p.PinPath != "" {
		err = concatErrors(err, os.Remove(p.PinPath))
	}

	// Shared with all probes: close the perf event file descriptor
	if p.perfEventFD != nil {
		err = p.perfEventFD.Close()
	}

	// Per program type cleanup
	switch p.programSpec.Type {
	case ebpf.UnspecifiedProgram:
		// nothing to do
		break
	case ebpf.Kprobe:
		err = concatErrors(err, p.detachKprobe())
	case ebpf.RawTracepoint, ebpf.RawTracepointWritable:
		err = concatErrors(err, p.detachRawTracepoint())
	case ebpf.CGroupDevice, ebpf.CGroupSKB, ebpf.CGroupSock, ebpf.CGroupSockAddr, ebpf.CGroupSockopt, ebpf.CGroupSysctl:
		err = concatErrors(err, p.detachCgroup())
	case ebpf.SocketFilter:
		err = concatErrors(err, p.detachSocket())
	case ebpf.SchedCLS:
		err = concatErrors(err, p.detachTCCLS())
	case ebpf.XDP:
		err = concatErrors(err, p.detachXDP())
	case ebpf.LSM:
		err = concatErrors(err, p.detachLSM())
	case ebpf.Tracing:
		err = concatErrors(err, p.detachTracing())
	case ebpf.PerfEvent:
		err = concatErrors(err, p.detachPerfEvent())
	default:
		// unsupported section, nothing to do either
		break
	}
	return err
}

// Stop - Detaches the probe from its hook point and close the underlying eBPF program.
func (p *Probe) Stop() error {
	p.stateLock.Lock()
	defer p.stateLock.Unlock()
	if p.state < paused || !p.Enabled {
		p.reset()
		return nil
	}
	return p.stop(true)
}

func (p *Probe) stop(saveStopError bool) error {
	// detach from hook point
	err := p.detachRetry()

	// close the loaded program
	if p.attachRetryAttempt >= p.getRetryAttemptCount() {
		err = concatErrors(err, p.program.Close())
	}

	// update state of the probe
	if saveStopError {
		p.lastError = concatErrors(p.lastError, err)
	}

	// Cleanup probe if stop was successful
	if err == nil && p.attachRetryAttempt >= p.getRetryAttemptCount() {
		p.reset()
	}
	if err != nil {
		return fmt.Errorf("couldn't stop probe %s: %w", p.ProbeIdentificationPair, err)
	}
	return nil
}

// reset - Cleans up the internal fields of the probe
func (p *Probe) reset() {
	p.kprobeType = ""
	p.netlinkSocketCache = nil
	p.program = nil
	p.programSpec = nil
	p.perfEventFD = nil
	p.rawTracepointFD = nil
	p.state = reset
	p.manualLoadNeeded = false
	p.checkPin = false
	p.attachPID = 0
	p.attachRetryAttempt = 0
	p.attachedWithDebugFS = false
	p.kprobeHookPointNotExist = false
	p.systemWideID = 0
	p.programTag = ""
	p.tcFilter = netlink.BpfFilter{}
	p.tcClsActQdisc = nil
}

// getNetlinkSocket returns a netlink socket in the probe network namespace
func (p *Probe) getNetlinkSocket() (*NetlinkSocket, error) {
	return p.netlinkSocketCache.getNetlinkSocket(p.IfIndexNetns, p.IfIndexNetnsID)
}

// attachWithKprobeEvents attaches the kprobe using the kprobes_events ABI
func (p *Probe) attachWithKprobeEvents() error {
	if p.kprobeHookPointNotExist {
		return ErrKProbeHookPointNotExist
	}

	// Prepare kprobe_events line parameters
	var maxActiveStr string
	if p.GetKprobeType() == RetProbeType {
		if p.KProbeMaxActive > 0 {
			maxActiveStr = fmt.Sprintf("%d", p.KProbeMaxActive)
		}
	}

	// Fallback to debugfs, write kprobe_events line to register kprobe
	var kprobeID int
	kprobeID, err := registerKprobeEvent(p.GetKprobeType(), p.HookFuncName, p.UID, maxActiveStr, p.attachPID)
	if err == ErrKprobeIDNotExist {
		// The probe might have been loaded under a kernel generated event name. Clean up just in case.
		_ = unregisterKprobeEventWithEventName(getKernelGeneratedEventName(p.GetKprobeType(), p.HookFuncName))
		// fallback without KProbeMaxActive
		kprobeID, err = registerKprobeEvent(p.GetKprobeType(), p.HookFuncName, p.UID, "", p.attachPID)
	}

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			p.kprobeHookPointNotExist = true
		}
		return fmt.Errorf("couldn't enable kprobe %s: %w", p.ProbeIdentificationPair, err)
	}

	// create perf event fd
	p.perfEventFD, err = perfEventOpenTracingEvent(kprobeID, -1)
	if err != nil {
		return fmt.Errorf("couldn't open perf event fd for %s: %w", p.ProbeIdentificationPair, err)
	}
	p.attachedWithDebugFS = true

	return nil
}

// attachKprobe - Attaches the probe to its kprobe
func (p *Probe) attachKprobe() error {
	var err error

	if len(p.HookFuncName) == 0 {
		return errors.New("HookFuncName, MatchFuncName or SyscallFuncName is required")
	}

	if p.GetKprobeType() == UnknownProbeType {
		// this might actually be a UProbe
		return p.attachUprobe()
	}

	isKRetProbe := p.GetKprobeType() == RetProbeType

	// currently the perf event open ABI doesn't allow to specify the max active parameter
	if p.KProbeMaxActive > 0 && isKRetProbe {
		if err = p.attachWithKprobeEvents(); err != nil {
			if p.perfEventFD, err = perfEventOpenPMU(p.HookFuncName, 0, -1, "kprobe", isKRetProbe, 0); err != nil {
				return err
			}
		}
	} else if p.KprobeAttachMethod == AttachKprobeWithPerfEventOpen {
		if p.perfEventFD, err = perfEventOpenPMU(p.HookFuncName, 0, -1, "kprobe", isKRetProbe, 0); err != nil {
			if err = p.attachWithKprobeEvents(); err != nil {
				return err
			}
		}
	} else if p.KprobeAttachMethod == AttachKprobeWithKprobeEvents {
		if err = p.attachWithKprobeEvents(); err != nil {
			if p.perfEventFD, err = perfEventOpenPMU(p.HookFuncName, 0, -1, "kprobe", isKRetProbe, 0); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("Invalid kprobe attach method: %d\n", p.KprobeAttachMethod)
	}

	// enable perf event
	if err = ioctlPerfEventSetBPF(p.perfEventFD, p.program.FD()); err != nil {
		return fmt.Errorf("couldn't set perf event bpf %s: %w", p.ProbeIdentificationPair, err)
	}
	if err = ioctlPerfEventEnable(p.perfEventFD); err != nil {
		return fmt.Errorf("couldn't enable perf event %s: %w", p.ProbeIdentificationPair, err)
	}
	return nil
}

// detachKprobe - Detaches the probe from its kprobe
func (p *Probe) detachKprobe() error {
	// Prepare kprobe_events line parameters
	if p.GetKprobeType() == UnknownProbeType {
		// this might be a Uprobe
		return p.detachUprobe()
	}

	if !p.attachedWithDebugFS {
		// nothing to do
		return nil
	}

	// Write kprobe_events line to remove hook point
	return unregisterKprobeEvent(p.GetKprobeType(), p.HookFuncName, p.UID, p.attachPID)
}

func (p *Probe) pauseKprobe() error {
	return ioctlPerfEventDisable(p.perfEventFD)
}

func (p *Probe) resumeKprobe() error {
	return ioctlPerfEventEnable(p.perfEventFD)
}

// attachTracepoint - Attaches the probe to its tracepoint
func (p *Probe) attachTracepoint() error {
	// Parse section
	if len(p.TracepointCategory) == 0 || len(p.TracepointName) == 0 {
		traceGroup := strings.SplitN(p.programSpec.SectionName, "/", 3)
		if len(traceGroup) != 3 {
			return fmt.Errorf("expected SEC(\"tracepoint/[category]/[name]\") got %s: %w", p.programSpec.SectionName, ErrSectionFormat)
		}
		p.TracepointCategory = traceGroup[1]
		p.TracepointName = traceGroup[2]
	}

	// Get the ID of the tracepoint to activate
	tracepointID, err := GetTracepointID(p.TracepointCategory, p.TracepointName)
	if err != nil {
		return fmt.Errorf("couldn't activate tracepoint %s: %w", p.ProbeIdentificationPair, err)
	}

	// Hook the eBPF program to the tracepoint
	p.perfEventFD, err = perfEventOpenTracingEvent(tracepointID, -1)
	if err != nil {
		return fmt.Errorf("couldn't enable tracepoint %s: %w", p.ProbeIdentificationPair, err)
	}
	if err = ioctlPerfEventSetBPF(p.perfEventFD, p.program.FD()); err != nil {
		return fmt.Errorf("couldn't set perf event bpf %s: %w", p.ProbeIdentificationPair, err)
	}
	if err = ioctlPerfEventEnable(p.perfEventFD); err != nil {
		return fmt.Errorf("couldn't enable perf event %s: %w", p.ProbeIdentificationPair, err)
	}
	return nil
}

func (p *Probe) pauseTracepoint() error {
	return ioctlPerfEventDisable(p.perfEventFD)
}

func (p *Probe) resumeTracepoint() error {
	return ioctlPerfEventEnable(p.perfEventFD)
}

// attachWithUprobeEvents attaches the uprobe using the uprobes_events ABI
func (p *Probe) attachWithUprobeEvents() error {
	// fallback to debugfs
	var uprobeID int
	uprobeID, err := registerUprobeEvent(p.GetUprobeType(), p.HookFuncName, p.BinaryPath, p.UID, p.attachPID, p.UprobeOffset)
	if err != nil {
		return fmt.Errorf("couldn't enable uprobe %s: %w", p.ProbeIdentificationPair, err)
	}

	// Activate perf event
	p.perfEventFD, err = perfEventOpenTracingEvent(uprobeID, p.PerfEventPID)
	if err != nil {
		return fmt.Errorf("couldn't open perf event fd for %s: %w", p.ProbeIdentificationPair, err)
	}
	p.attachedWithDebugFS = true
	return nil
}

// attachUprobe - Attaches the probe to its Uprobe
func (p *Probe) attachUprobe() error {
	var err error

	// Prepare uprobe_events line parameters
	if p.GetUprobeType() == UnknownProbeType {
		// unknown type
		return fmt.Errorf("program type unrecognized in %s: %w", p.ProbeIdentificationPair, ErrSectionFormat)
	}

	// compute the offset if it was not provided
	if p.UprobeOffset == 0 {
		var funcPattern string

		// find the offset of the first symbol matching the provided pattern
		if len(p.MatchFuncName) > 0 {
			funcPattern = p.MatchFuncName
		} else {
			funcPattern = fmt.Sprintf("^%s$", p.HookFuncName)
		}
		pattern, err := regexp.Compile(funcPattern)
		if err != nil {
			return fmt.Errorf("failed to compile pattern %s: %w", funcPattern, err)
		}

		// Retrieve dynamic symbol offset
		offsets, err := findSymbolOffsets(p.BinaryPath, pattern)
		if err != nil {
			return fmt.Errorf("couldn't find symbol matching %s in %s: %w", pattern.String(), p.BinaryPath, err)
		}
		p.UprobeOffset = offsets[0].Value
		p.HookFuncName = offsets[0].Name
	}

	isURetProbe := p.GetUprobeType() == "r"
	if p.UprobeAttachMethod == AttachWithPerfEventOpen {
		if p.perfEventFD, err = perfEventOpenPMU(p.BinaryPath, int(p.UprobeOffset), p.PerfEventPID, "uprobe", isURetProbe, 0); err != nil {
			if err = p.attachWithUprobeEvents(); err != nil {
				return err
			}
		}
	} else if p.UprobeAttachMethod == AttachWithProbeEvents {
		if err = p.attachWithUprobeEvents(); err != nil {
			if p.perfEventFD, err = perfEventOpenPMU(p.BinaryPath, int(p.UprobeOffset), p.PerfEventPID, "uprobe", isURetProbe, 0); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("Invalid uprobe attach method: %d\n", p.UprobeAttachMethod)
	}

	// enable perf event
	if err = ioctlPerfEventSetBPF(p.perfEventFD, p.program.FD()); err != nil {
		return fmt.Errorf("couldn't set perf event bpf %s: %w", p.ProbeIdentificationPair, err)
	}
	if err = ioctlPerfEventEnable(p.perfEventFD); err != nil {
		return fmt.Errorf("couldn't enable perf event %s: %w", p.ProbeIdentificationPair, err)
	}
	return nil
}

// detachUprobe - Detaches the probe from its Uprobe
func (p *Probe) detachUprobe() error {
	if !p.attachedWithDebugFS {
		// nothing to do
		return nil
	}

	// Prepare uprobe_events line parameters
	if p.GetUprobeType() == UnknownProbeType {
		// unknown type
		return fmt.Errorf("program type unrecognized in section %v: %w", p.ProbeIdentificationPair, ErrSectionFormat)
	}

	// Write uprobe_events line to remove hook point
	return unregisterUprobeEvent(p.GetUprobeType(), p.HookFuncName, p.UID, p.attachPID)
}

// attachCGroup - Attaches the probe to a cgroup hook point
func (p *Probe) attachCGroup() error {
	// open CGroupPath
	f, err := os.Open(p.CGroupPath)
	if err != nil {
		return fmt.Errorf("error opening cgroup %s from probe %s: %w", p.CGroupPath, p.ProbeIdentificationPair, err)
	}
	defer f.Close()

	// Attach CGroup
	ret, err := bpfProgAttach(p.program.FD(), int(f.Fd()), p.programSpec.AttachType)
	if ret < 0 {
		return fmt.Errorf("failed to attach probe %v to cgroup %s: %w", p.ProbeIdentificationPair, p.CGroupPath, err)
	}
	return nil
}

// detachCGroup - Detaches the probe from its cgroup hook point
func (p *Probe) detachCgroup() error {
	// open CGroupPath
	f, err := os.Open(p.CGroupPath)
	if err != nil {
		return fmt.Errorf("error opening cgroup %s from probe %s: %w", p.CGroupPath, p.ProbeIdentificationPair, err)
	}

	// Detach CGroup
	ret, err := bpfProgDetach(p.program.FD(), int(f.Fd()), p.programSpec.AttachType)
	if ret < 0 {
		return fmt.Errorf("failed to detach probe %v from cgroup %s: %w", p.ProbeIdentificationPair, p.CGroupPath, err)
	}
	return nil
}

// attachSocket - Attaches the probe to the provided socket
func (p *Probe) attachSocket() error {
	return sockAttach(p.SocketFD, p.program.FD())
}

// detachSocket - Detaches the probe from its socket
func (p *Probe) detachSocket() error {
	return sockDetach(p.SocketFD, p.program.FD())
}

func (p *Probe) buildTCClsActQdisc() netlink.Qdisc {
	if p.tcClsActQdisc == nil {
		p.tcClsActQdisc = &netlink.GenericQdisc{
			QdiscType: "clsact",
			QdiscAttrs: netlink.QdiscAttrs{
				LinkIndex: p.IfIndex,
				Handle:    netlink.MakeHandle(0xffff, 0),
				Parent:    netlink.HANDLE_INGRESS,
			},
		}
	}
	return p.tcClsActQdisc
}

func (p *Probe) getTCFilterParentHandle() uint32 {
	return netlink.MakeHandle(clsactQdisc, uint16(p.NetworkDirection))
}

func (p *Probe) buildTCFilter() (netlink.BpfFilter, error) {
	if p.tcFilter.FilterAttrs.LinkIndex == 0 {
		var filterName string
		filterName, err := generateTCFilterName(p.UID, p.programSpec.SectionName, p.attachPID)
		if err != nil {
			return p.tcFilter, fmt.Errorf("couldn't create TC filter for %v: %w", p.ProbeIdentificationPair, err)
		}

		p.tcFilter = netlink.BpfFilter{
			FilterAttrs: netlink.FilterAttrs{
				LinkIndex: p.IfIndex,
				Parent:    p.getTCFilterParentHandle(),
				Handle:    p.TCFilterHandle,
				Priority:  p.TCFilterPrio,
				Protocol:  p.TCFilterProtocol,
			},
			Fd:           p.program.FD(),
			Name:         filterName,
			DirectAction: true,
		}
	}
	return p.tcFilter, nil
}

// attachTCCLS - Attaches the probe to its TC classifier hook point
func (p *Probe) attachTCCLS() error {
	var err error
	// Resolve Probe's interface
	if _, err = p.resolveLink(); err != nil {
		return err
	}

	// Recover the netlink socket of the interface from the manager
	ntl, err := p.getNetlinkSocket()
	if err != nil {
		return err
	}

	// Create a Qdisc for the provided interface
	err = ntl.Sock.QdiscAdd(p.buildTCClsActQdisc())
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			// cleanup previous TC filters if necessary
			if err = p.cleanupTCFilters(ntl); err != nil {
				return fmt.Errorf("couldn't clean up existing \"clsact\" qdisc filters for %s[%d]: %w", p.IfName, p.IfIndex, err)
			}
		} else {
			return fmt.Errorf("couldn't add a \"clsact\" qdisc to interface %s[%d]: %w", p.IfName, p.IfIndex, err)
		}
	}

	// Create qdisc filter
	_, err = p.buildTCFilter()
	if err != nil {
		return err
	}
	if err = ntl.Sock.FilterAdd(&p.tcFilter); err != nil {
		return fmt.Errorf("couldn't add a %v filter to interface %s[%d]: %v", p.NetworkDirection, p.IfName, p.IfIndex, err)
	}

	// retrieve filter handle
	resp, err := ntl.Sock.FilterList(p.link, p.tcFilter.Parent)
	if err != nil {
		return fmt.Errorf("couldn't list filters of interface %s[%d]: %v", p.IfName, p.IfIndex, err)
	}

	var found bool
	bpfType := (&netlink.BpfFilter{}).Type()
	for _, elem := range resp {
		if elem.Type() != bpfType {
			continue
		}

		bpfFilter, ok := elem.(*netlink.BpfFilter)
		if !ok {
			continue
		}

		// we can't test the equality of the program tag, there is a bug in the netlink library.
		// See https://github.com/vishvananda/netlink/issues/722
		if bpfFilter.Id == p.systemWideID && strings.Contains(p.programTag, bpfFilter.Tag) { //
			found = true
			p.tcFilter.Handle = bpfFilter.Handle
		}
	}
	if !found {
		return fmt.Errorf("couldn't create TC filter for %v: filter not found", p.ProbeIdentificationPair)
	}
	ntl.TCFilterCount[p.IfIndex]++

	return nil
}

func (p *Probe) IsTCFilterActive() bool {
	p.stateLock.Lock()
	defer p.stateLock.Unlock()
	if p.state < paused || !p.Enabled {
		return false
	}
	if p.programSpec.Type != ebpf.SchedCLS {
		return false
	}

	// Recover the netlink socket of the interface from the manager
	ntl, err := p.getNetlinkSocket()
	if err != nil {
		return false
	}

	resp, err := ntl.Sock.FilterList(p.link, p.tcFilter.Parent)
	if err != nil {
		return false
	}

	bpfType := (&netlink.BpfFilter{}).Type()
	for _, elem := range resp {
		if elem.Type() != bpfType {
			continue
		}

		bpfFilter, ok := elem.(*netlink.BpfFilter)
		if !ok {
			continue
		}

		// we can't test the equality of the program tag, there is a bug in the netlink library.
		// See https://github.com/vishvananda/netlink/issues/722
		if bpfFilter.Id == p.systemWideID && strings.Contains(p.programTag, bpfFilter.Tag) {
			return true
		}
	}

	// This TC filter is no longer active, the interface has been deleted or the filter was replaced by a third party.
	// Regardless of the reason, we do not hold the current Handle on this filter, remove it, so we make sure we won't
	// delete something that we do not own.
	p.tcFilter.Handle = 0
	return false
}

// detachTCCLS - Detaches the probe from its TC classifier hook point
func (p *Probe) detachTCCLS() error {
	// Recover the netlink socket of the interface from the manager
	ntl, err := p.getNetlinkSocket()
	if err != nil {
		return err
	}

	if p.tcFilter.Handle > 0 {
		// delete the current filter
		if err = ntl.Sock.FilterDel(&p.tcFilter); err != nil {
			// the device or the filter might already be gone, ignore the error if that's the case
			if !errors.Is(err, syscall.ENODEV) && !errors.Is(err, syscall.ENOENT) {
				return fmt.Errorf("couldn't remove TC classifier %v: %w", p.ProbeIdentificationPair, err)
			}
		}
	}
	ntl.TCFilterCount[p.IfIndex]--

	// check if the qdisc should be deleted
	if ntl.TCFilterCount[p.IfIndex] >= 1 {
		// at list one of our classifiers is still using it
		return nil
	}
	delete(ntl.TCFilterCount, p.IfIndex)

	if p.TCCleanupQDisc {

		// check if someone else is using the clsact qdisc on ingress
		resp, err := ntl.Sock.FilterList(p.link, netlink.HANDLE_MIN_INGRESS)
		if err != nil || err == nil && len(resp) > 0 {
			// someone is still using it
			return nil
		}

		// check on egress
		resp, err = ntl.Sock.FilterList(p.link, netlink.HANDLE_MIN_EGRESS)
		if err != nil || err == nil && len(resp) > 0 {
			// someone is still using it
			return nil
		}

		// delete qdisc
		if err = ntl.Sock.QdiscDel(p.buildTCClsActQdisc()); err != nil {
			// the device might already be gone, ignore the error if that's the case
			if !errors.Is(err, syscall.ENODEV) {
				return fmt.Errorf("couldn't remove clsact qdisc: %w", err)
			}
		}
	}
	return nil
}

// cleanupTCFilters - Cleans up existing TC Filters by removing entries of known UIDs, that they're not used anymore.
//
// Previous instances of this manager might have been killed unexpectedly. When this happens, TC filters are not cleaned
// up properly and can grow indefinitely. To prevent this, start by cleaning up the TC filters of previous managers that
// are not running anymore.
func (p *Probe) cleanupTCFilters(ntl *NetlinkSocket) error {
	// build the pattern to look for in the TC filters name
	pattern, err := regexp.Compile(fmt.Sprintf(`.*(%s)_([0-9]*)`, p.UID))
	if err != nil {
		return fmt.Errorf("filter name pattern generation failed: %w", err)
	}

	resp, err := ntl.Sock.FilterList(p.link, p.getTCFilterParentHandle())
	if err != nil {
		return err
	}

	var deleteErr error
	bpfType := (&netlink.BpfFilter{}).Type()
	for _, elem := range resp {
		if elem.Type() != bpfType {
			continue
		}

		bpfFilter, ok := elem.(*netlink.BpfFilter)
		if !ok {
			continue
		}

		match := pattern.FindStringSubmatch(bpfFilter.Name)
		if len(match) < 3 {
			continue
		}

		// check if the manager that loaded this TC filter is still up
		var pid int
		pid, err = strconv.Atoi(match[2])
		if err != nil {
			continue
		}

		// this short sleep is used to avoid a CPU spike (5s ~ 60k * 80 microseconds)
		time.Sleep(80 * time.Microsecond)

		_, err = process.NewProcess(int32(pid))
		if err == nil {
			continue
		}

		// remove this filter
		deleteErr = concatErrors(deleteErr, ntl.Sock.FilterDel(elem))
	}
	return deleteErr
}

// attachXDP - Attaches the probe to an interface with an XDP hook point
func (p *Probe) attachXDP() error {
	var err error
	// Resolve Probe's interface
	if _, err = p.resolveLink(); err != nil {
		return err
	}

	// Attach program
	err = netlink.LinkSetXdpFdWithFlags(p.link, p.program.FD(), int(p.XDPAttachMode))
	if err != nil {
		return fmt.Errorf("couldn't attach XDP program %v to interface %v: %w", p.ProbeIdentificationPair, p.IfIndex, err)
	}
	return nil
}

// detachXDP - Detaches the probe from its XDP hook point
func (p *Probe) detachXDP() error {
	var err error
	// Resolve Probe's interface
	if _, err = p.resolveLink(); err != nil {
		return err
	}

	// Detach program
	err = netlink.LinkSetXdpFdWithFlags(p.link, -1, int(p.XDPAttachMode))
	if err != nil {
		return fmt.Errorf("couldn't detach XDP program %v from interface %v: %w", p.ProbeIdentificationPair, p.IfIndex, err)
	}
	return nil
}

// attachLSM - Attaches the probe to its LSM hook point
func (p *Probe) attachLSM() error {
	var err error
	p.rawTracepointFD, err = rawTracepointOpen("", p.program.FD())
	if err != nil {
		return fmt.Errorf("failed to attach LSM hook point: %w", err)
	}
	return nil
}

// detachLSM - Detaches the probe from its LSM hook point
func (p *Probe) detachLSM() error {
	if p.rawTracepointFD != nil {
		if closeErr := p.rawTracepointFD.Close(); closeErr != nil {
			return fmt.Errorf("failed to detach LSM hook point: %w", closeErr)
		}
	}
	return nil
}

func (p *Probe) attachTracing() error {
	var err error
	p.rawTracepointFD, err = rawTracepointOpen("", p.program.FD())
	if err != nil {
		return fmt.Errorf("failed to attach tracing hook point: %w", err)
	}
	return nil
}

func (p *Probe) detachTracing() error {
	if p.rawTracepointFD != nil {
		if closeErr := p.rawTracepointFD.Close(); closeErr != nil {
			return fmt.Errorf("failed to detach tracing hook point: %w", closeErr)
		}
	}
	return nil
}

// attachPerfEvent - Attaches the perf_event program
func (p *Probe) attachPerfEvent() error {
	if p.PerfEventType != unix.PERF_TYPE_HARDWARE && p.PerfEventType != unix.PERF_TYPE_SOFTWARE {
		return fmt.Errorf("unknown PerfEventType parameter: %v (expected unix.PERF_TYPE_HARDWARE or unix.PERF_TYPE_SOFTWARE)", p.PerfEventType)
	}

	attr := unix.PerfEventAttr{
		Type:   uint32(p.PerfEventType),
		Sample: uint64(p.SamplePeriod),
		Config: uint64(p.PerfEventConfig),
	}

	if p.SampleFrequency > 0 {
		attr.Sample = uint64(p.SampleFrequency)
		attr.Bits |= unix.PerfBitFreq
	}

	if p.PerfEventCPUCount == 0 {
		p.PerfEventCPUCount = runtime.NumCPU()
	}

	pid := p.PerfEventPID
	if pid == 0 {
		pid = -1
	}

	for cpu := 0; cpu < p.PerfEventCPUCount; cpu++ {
		fd, err := perfEventOpenRaw(&attr, pid, cpu, -1, 0)
		if err != nil {
			return fmt.Errorf("couldn't attach perf_event program %s on pid %d and CPU %d: %v", p.ProbeIdentificationPair, pid, cpu, err)
		}
		p.perfEventCPUFDs = append(p.perfEventCPUFDs, fd)

		if err = ioctlPerfEventSetBPF(fd, p.program.FD()); err != nil {
			return fmt.Errorf("couldn't set perf event bpf %s: %w", p.ProbeIdentificationPair, err)
		}
		if err = ioctlPerfEventEnable(fd); err != nil {
			return fmt.Errorf("couldn't enable perf event %s for pid %d and CPU %d: %w", p.ProbeIdentificationPair, pid, cpu, err)
		}
	}
	return nil
}

// detachPerfEvent - Detaches the perf_event program
func (p *Probe) detachPerfEvent() error {
	var err error
	for _, fd := range p.perfEventCPUFDs {
		err = concatErrors(err, fd.Close())
	}
	p.perfEventCPUFDs = []*fd{}
	return err
}

// attachRawTracepoint - Attaches the probe to its raw_tracepoint
func (p *Probe) attachRawTracepoint() error {
	var err error
	p.rawTracepointFD, err = rawTracepointOpen(p.TracepointName, p.program.FD())
	if err != nil {
		return fmt.Errorf("failed to attach raw_tracepoint %s: %w", p.ProbeIdentificationPair, err)
	}
	return nil
}

// detachRawTracepoint - Detaches the probe from its raw_tracepoint
func (p *Probe) detachRawTracepoint() error {
	if p.rawTracepointFD != nil {
		if closeErr := p.rawTracepointFD.Close(); closeErr != nil {
			return fmt.Errorf("failed to detech raw_tracepoint %s: %w", p.ProbeIdentificationPair, closeErr)
		}
	}
	return nil
}

func (p *Probe) getRetryAttemptCount() uint {
	return p.ProbeRetry + 1
}
