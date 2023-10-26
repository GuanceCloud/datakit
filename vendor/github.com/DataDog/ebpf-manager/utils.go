package manager

import (
	"bytes"
	"debug/elf"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/cilium/ebpf"
	"golang.org/x/sys/unix"

	"github.com/DataDog/ebpf-manager/tracefs"
)

type state uint

const (
	reset state = iota
	initialized
	stopped
	paused
	running

	// maxEventNameLen - maximum length for a kprobe (or uprobe) event name
	// MAX_EVENT_NAME_LEN (linux/kernel/trace/trace.h)
	maxEventNameLen    = 64
	minFunctionNameLen = 10

	// maxBPFClassifierNameLen - maximum length for a TC
	// CLS_BPF_NAME_LEN (linux/net/sched/cls_bpf.c)
	maxBPFClassifierNameLen = 256
)

// concatErrors - Concatenate 2 errors into one error.
func concatErrors(err1, err2 error) error {
	if err1 == nil {
		return err2
	}
	if err2 != nil {
		return fmt.Errorf("%s: %w", err2.Error(), err1)
	}
	return err1
}

// availableFilterFunctions - cache of the list of available kernel functions.
var availableFilterFunctions []string

func FindFilterFunction(funcName string) (string, error) {
	// Prepare matching pattern
	searchedName, err := regexp.Compile(funcName)
	if err != nil {
		return "", err
	}

	// Cache available filter functions if necessary
	if len(availableFilterFunctions) == 0 {
		funcs, err := tracefs.ReadFile("available_filter_functions")
		if err != nil {
			return "", err
		}
		availableFilterFunctions = strings.Split(string(funcs), "\n")
		for i, name := range availableFilterFunctions {
			splittedName := strings.Split(name, " ")
			name = splittedName[0]
			splittedName = strings.Split(name, "\t")
			name = splittedName[0]
			availableFilterFunctions[i] = name
		}
		sort.Strings(availableFilterFunctions)
	}

	// Match function name
	var potentialMatches []string
	for _, f := range availableFilterFunctions {
		if searchedName.MatchString(f) {
			potentialMatches = append(potentialMatches, f)
		}
		if f == funcName {
			return f, nil
		}
	}
	if len(potentialMatches) > 0 {
		return potentialMatches[0], nil
	}
	return "", nil
}

// cache of the syscall prefix depending on kernel version
var syscallPrefix string

// GetSyscallFnName - Returns the kernel function of the provided syscall, after reading /proc/kallsyms to retrieve
// the list of symbols of the current kernel.
func GetSyscallFnName(name string) (string, error) {
	return GetSyscallFnNameWithSymFile(name, defaultSymFile)
}

// GetSyscallFnNameWithSymFile - Returns the kernel function of the provided syscall, after reading symFile to retrieve
// the list of symbols of the current kernel.
func GetSyscallFnNameWithSymFile(name string, symFile string) (string, error) {
	if symFile == "" {
		symFile = defaultSymFile
	}
	if syscallPrefix == "" {
		syscallName, err := getSyscallName("open", symFile)
		if err != nil {
			return "", err
		}
		// copy to avoid memory leak due to go subslice
		// see: https://go101.org/article/memory-leaking.html
		var b strings.Builder
		b.WriteString(syscallName)
		syscallName = b.String()

		syscallPrefix = strings.TrimSuffix(syscallName, "open")
	}

	return syscallPrefix + name, nil
}

const defaultSymFile = "/proc/kallsyms"

// Returns the qualified syscall named by going through '/proc/kallsyms' on the
// system on which its executed. It allows bpf programs that may have been compiled
// for older syscall functions to run on newer kernels
func getSyscallName(name string, symFile string) (string, error) {
	// Get kernel symbols
	syms, err := os.ReadFile(symFile)
	if err != nil {
		return "", err
	}
	return getSyscallFnNameWithKallsyms(name, string(syms))
}

func getSyscallFnNameWithKallsyms(name string, kallsymsContent string) (string, error) {
	var arch string
	switch runtime.GOARCH {
	case "386":
		arch = "ia32"
	case "arm64":
		arch = "arm64"
	default:
		arch = "x64"
	}

	// We should search for new syscall function like "__x64__sys_open"
	// Note the start of word boundary. Should return exactly one string
	regexStr := `(\b__` + arch + `_[Ss]y[sS]_` + name + `\b)`
	fnRegex := regexp.MustCompile(regexStr)

	match := fnRegex.FindAllString(kallsymsContent, -1)
	if len(match) > 0 {
		return match[0], nil
	}

	// If nothing found, search for old syscall function to be sure
	regexStr = `(\b[Ss]y[sS]_` + name + `\b)`
	fnRegex = regexp.MustCompile(regexStr)
	match = fnRegex.FindAllString(kallsymsContent, -1)
	// If we get something like 'sys_open' or 'SyS_open', return
	// either (they have same addr) else, just return original string
	if len(match) > 0 {
		return match[0], nil
	}

	// check for '__' prefixed functions, like '__sys_open'
	regexStr = `(\b__[Ss]y[sS]_` + name + `\b)`
	fnRegex = regexp.MustCompile(regexStr)
	match = fnRegex.FindAllString(kallsymsContent, -1)
	// If we get something like '__sys_open' or '__SyS_open', return
	// either (they have same addr) else, just return original string
	if len(match) > 0 {
		return match[0], nil
	}

	return "", fmt.Errorf("could not find a valid syscall name")
}

var safeEventRegexp = regexp.MustCompile("[^a-zA-Z0-9]")

func generateEventName(probeType, funcName, UID string, attachPID int) (string, error) {
	// truncate the function name and UID name to reduce the length of the event
	attachPIDstr := strconv.Itoa(attachPID)
	maxFuncNameLen := maxEventNameLen - 3 /* _ */ - len(probeType) - len(UID) - len(attachPIDstr)
	if maxFuncNameLen < minFunctionNameLen { /* let's guarantee that we have a function name minimum of 10 chars (minFunctionNameLen) or trow an error */
		dbgFullEventString := safeEventRegexp.ReplaceAllString(fmt.Sprintf("%s_%s_%s_%s", probeType, funcName, UID, attachPIDstr), "_")
		return "", fmt.Errorf("event name is too long (kernel limit is %d (MAX_EVENT_NAME_LEN)): minFunctionNameLen %d, len 3, probeType %d, funcName %d, UID %d, attachPIDstr %d ; full event string : '%s'", maxEventNameLen, minFunctionNameLen, len(probeType), len(funcName), len(UID), len(attachPIDstr), dbgFullEventString)
	}
	eventName := safeEventRegexp.ReplaceAllString(fmt.Sprintf("%s_%.*s_%s_%s", probeType, maxFuncNameLen, funcName, UID, attachPIDstr), "_")

	if len(eventName) > maxEventNameLen {
		return "", fmt.Errorf("event name too long (kernel limit MAX_EVENT_NAME_LEN is %d): '%s'", maxEventNameLen, eventName)
	}
	return eventName, nil
}

func generateTCFilterName(UID, sectionName string, attachPID int) (string, error) {
	attachPIDstr := strconv.Itoa(attachPID)
	maxSectionNameLen := maxBPFClassifierNameLen - 3 /* _ */ - len(UID) - len(attachPIDstr)
	if maxSectionNameLen < 0 {
		dbgFullFilterString := safeEventRegexp.ReplaceAllString(fmt.Sprintf("%s_%s_%s", sectionName, UID, attachPIDstr), "_")
		return "", fmt.Errorf("filter name is too long (kernel limit is %d (CLS_BPF_NAME_LEN)): sectionName %d, UID %d, attachPIDstr %d ; full event string : '%s'", maxEventNameLen, len(sectionName), len(UID), len(attachPIDstr), dbgFullFilterString)
	}
	filterName := safeEventRegexp.ReplaceAllString(fmt.Sprintf("%.*s_%s_%s", maxSectionNameLen, sectionName, UID, attachPIDstr), "_")

	if len(filterName) > maxBPFClassifierNameLen {
		return "", fmt.Errorf("filter name too long (kernel limit CLS_BPF_NAME_LEN is %d): '%s'", maxBPFClassifierNameLen, filterName)
	}
	return filterName, nil
}

// getKernelGeneratedEventName returns the pattern used by the kernel when a [k|u]probe is loaded without an event name.
// The library doesn't support loading a [k|u]probe with an address directly, so only one pattern applies here.
func getKernelGeneratedEventName(probeType, funcName string) string {
	return fmt.Sprintf("%s_%s_0", probeType, funcName)
}

// readKprobeEvents - Returns the content of kprobe_events
func readKprobeEvents() (string, error) {
	kprobeEvents, err := tracefs.ReadFile("kprobe_events")
	if err != nil {
		return "", err
	}
	return string(kprobeEvents), nil
}

// registerKprobeEvent - Writes a new kprobe in kprobe_events with the provided parameters. Call DisableKprobeEvent
// to remove the krpobe.
func registerKprobeEvent(probeType, funcName, UID, maxActiveStr string, kprobeAttachPID int) (int, error) {
	// Generate event name
	eventName, err := generateEventName(probeType, funcName, UID, kprobeAttachPID)
	if err != nil {
		return -1, err
	}

	// Write line to kprobe_events
	f, err := tracefs.OpenFile("kprobe_events", os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return -1, fmt.Errorf("cannot open kprobe_events: %w", err)
	}
	defer f.Close()
	cmd := fmt.Sprintf("%s%s:%s %s\n", probeType, maxActiveStr, eventName, funcName)
	if _, err = f.WriteString(cmd); err != nil && !os.IsExist(err) {
		return -1, fmt.Errorf("cannot write %q to kprobe_events: %w", cmd, err)
	}

	// Retrieve kprobe ID
	kprobeIDFile := fmt.Sprintf("events/kprobes/%s/id", eventName)
	kprobeIDBytes, err := tracefs.ReadFile(kprobeIDFile)
	if err != nil {
		if os.IsNotExist(err) {
			return -1, ErrKprobeIDNotExist
		}
		return -1, fmt.Errorf("cannot read kprobe id: %w", err)
	}
	id := strings.TrimSpace(string(kprobeIDBytes))
	kprobeID, err := strconv.Atoi(id)
	if err != nil {
		return -1, fmt.Errorf("invalid kprobe id: '%s': %w", id, err)
	}
	return kprobeID, nil
}

// unregisterKprobeEvent - Removes a kprobe from kprobe_events
func unregisterKprobeEvent(probeType, funcName, UID string, kprobeAttachPID int) error {
	// Generate event name
	eventName, err := generateEventName(probeType, funcName, UID, kprobeAttachPID)
	if err != nil {
		return err
	}
	return unregisterKprobeEventWithEventName(eventName)
}

func unregisterKprobeEventWithEventName(eventName string) error {
	// Write line to kprobe_events
	f, err := tracefs.OpenFile("kprobe_events", os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("cannot open kprobe_events: %w", err)
	}
	defer f.Close()
	cmd := fmt.Sprintf("-:%s\n", eventName)
	if _, err = f.WriteString(cmd); err != nil {
		pathErr, ok := err.(*os.PathError)
		if ok && pathErr.Err == syscall.ENOENT {
			// This can happen when for example two modules
			// use the same elf object and both ratecall `Close()`.
			// The second will encounter the error as the
			// probe already has been cleared by the first.
			return nil
		}
		return fmt.Errorf("cannot write %q to kprobe_events: %w", cmd, err)
	}
	return nil
}

// readUprobeEvents - Returns the content of uprobe_events
func readUprobeEvents() (string, error) {
	uprobeEvents, err := tracefs.ReadFile("uprobe_events")
	if err != nil {
		return "", err
	}
	return string(uprobeEvents), nil
}

// registerUprobeEvent - Writes a new Uprobe in uprobe_events with the provided parameters. Call DisableUprobeEvent
// to remove the krpobe.
func registerUprobeEvent(probeType string, funcName, path, UID string, uprobeAttachPID int, offset uint64) (int, error) {
	// Generate event name
	eventName, err := generateEventName(probeType, funcName, UID, uprobeAttachPID)
	if err != nil {
		return -1, err
	}

	// Write line to uprobe_events, only eventName is tested to max MAX_EVENT_NAME_LEN (linux/kernel/trace/trace.h)

	f, err := tracefs.OpenFile("uprobe_events", os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return -1, fmt.Errorf("cannot open uprobe_events: %w", err)
	}
	defer f.Close()

	cmd := fmt.Sprintf("%s:%s %s:%#x\n", probeType, eventName, path, offset)

	if _, err = f.WriteString(cmd); err != nil && !os.IsExist(err) {
		return -1, fmt.Errorf("cannot write %q to uprobe_events: %w", cmd, err)
	}

	// Retrieve Uprobe ID
	uprobeIDFile := fmt.Sprintf("events/uprobes/%s/id", eventName)
	uprobeIDBytes, err := tracefs.ReadFile(uprobeIDFile)
	if err != nil {
		if os.IsNotExist(err) {
			return -1, ErrUprobeIDNotExist
		}
		return -1, fmt.Errorf("cannot read uprobe id: %w", err)
	}
	uprobeID, err := strconv.Atoi(strings.TrimSpace(string(uprobeIDBytes)))
	if err != nil {
		return -1, fmt.Errorf("invalid uprobe id: %w", err)
	}

	return uprobeID, nil
}

// unregisterUprobeEvent - Removes a uprobe from uprobe_events
func unregisterUprobeEvent(probeType string, funcName string, UID string, uprobeAttachPID int) error {
	// Generate event name
	eventName, err := generateEventName(probeType, funcName, UID, uprobeAttachPID)
	if err != nil {
		return err
	}
	return unregisterUprobeEventWithEventName(eventName)
}

func unregisterUprobeEventWithEventName(eventName string) error {
	// Write uprobe_events line
	f, err := tracefs.OpenFile("uprobe_events", os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("cannot open uprobe_events: %w", err)
	}
	defer f.Close()
	cmd := fmt.Sprintf("-:%s\n", eventName)
	if _, err = f.WriteString(cmd); err != nil {
		return fmt.Errorf("cannot write %q to uprobe_events: %w", cmd, err)
	}
	return nil
}

// OpenAndListSymbols - Opens an elf file and extracts all its symbols
func OpenAndListSymbols(path string) (*elf.File, []elf.Symbol, error) {
	// open elf file
	f, err := elf.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't open elf file %s: %w", path, err)
	}
	defer f.Close()

	// Loop through all symbols
	syms, errSyms := f.Symbols()
	dynSyms, errDynSyms := f.DynamicSymbols()
	syms = append(syms, dynSyms...)

	if len(syms) == 0 {
		var err error
		if errSyms != nil {
			err = fmt.Errorf("failed to list symbols: %w", err)
		}
		if errDynSyms != nil {
			err = fmt.Errorf("failed to list dynamic symbols: %w", err)
		}
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("no symbols found")
	}
	return f, syms, nil
}

// SanitizeUprobeAddresses - sanitizes the addresses of the provided symbols
func SanitizeUprobeAddresses(f *elf.File, syms []elf.Symbol) {
	// If the binary is a non-PIE executable, addr must be a virtual address, otherwise it must be an offset relative to
	// the file load address. For executable (ET_EXEC) binaries and shared objects (ET_DYN), translate the virtual
	// address to physical address in the binary file.
	if f.Type == elf.ET_EXEC || f.Type == elf.ET_DYN {
		for i, sym := range syms {
			for _, prog := range f.Progs {
				if prog.Type == elf.PT_LOAD {
					if sym.Value >= prog.Vaddr && sym.Value < (prog.Vaddr+prog.Memsz) {
						syms[i].Value = sym.Value - prog.Vaddr + prog.Off
					}
				}
			}
		}
	}
}

// findSymbolOffsets - Parses the provided file and returns the offsets of the symbols that match the provided pattern
func findSymbolOffsets(path string, pattern *regexp.Regexp) ([]elf.Symbol, error) {
	f, syms, err := OpenAndListSymbols(path)
	if err != nil {
		return nil, err
	}

	var matches []elf.Symbol
	for _, sym := range syms {
		if elf.ST_TYPE(sym.Info) == elf.STT_FUNC && pattern.MatchString(sym.Name) {
			matches = append(matches, sym)
		}
	}

	if len(matches) == 0 {
		return nil, ErrSymbolNotFound
	}

	SanitizeUprobeAddresses(f, matches)
	return matches, nil
}

// GetTracepointID - Returns a tracepoint ID from its category and name
func GetTracepointID(category, name string) (int, error) {
	tracepointIDFile := fmt.Sprintf("events/%s/%s/id", category, name)
	tracepointIDBytes, err := tracefs.ReadFile(tracepointIDFile)
	if err != nil {
		return -1, fmt.Errorf("cannot read tracepoint id %q: %w", tracepointIDFile, err)
	}
	tracepointID, err := strconv.Atoi(strings.TrimSpace(string(tracepointIDBytes)))
	if err != nil {
		return -1, fmt.Errorf("invalid tracepoint id: %w", err)
	}
	return tracepointID, nil
}

// errClosedFd - Use of closed file descriptor error
var errClosedFd = errors.New("use of closed file descriptor")

// fd - File descriptor
type fd struct {
	raw int64
}

// newFD - returns a new file descriptor
func newFD(value uint32) *fd {
	f := &fd{int64(value)}
	runtime.SetFinalizer(f, func(f *fd) {
		_ = f.Close()
	})
	return f
}

func (fd *fd) String() string {
	return strconv.FormatInt(fd.raw, 10)
}

func (fd *fd) Value() (uint32, error) {
	if fd.raw < 0 {
		return 0, errClosedFd
	}

	return uint32(fd.raw), nil
}

func (fd *fd) Close() error {
	if fd.raw < 0 {
		return nil
	}

	value := int(fd.raw)
	fd.raw = -1

	runtime.SetFinalizer(fd, nil)
	return unix.Close(value)
}

var (
	// kprobePMUType is used to cache the kprobe PMY type value
	kprobePMUType = struct {
		once  sync.Once
		value uint32
		err   error
	}{}
	// uprobePMUType is used to cache the uprobe PMU type value
	uprobePMUType = struct {
		once  sync.Once
		value uint32
		err   error
	}{}
)

func parsePMUEventType(eventType string) (uint32, error) {
	PMUTypeFile := fmt.Sprintf("/sys/bus/event_source/devices/%s/type", eventType)
	f, err := os.Open(PMUTypeFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, fmt.Errorf("pmu type %s: %w", eventType, ErrNotSupported)
		}
		return 0, fmt.Errorf("couldn't open %s: %w", PMUTypeFile, err)
	}

	var t uint32
	_, err = fmt.Fscanf(f, "%d\n", &t)
	if err != nil {
		return 0, fmt.Errorf("couldn't parse type at %s: %v", eventType, err)
	}
	return t, nil
}

// getPMUEventType reads a Performance Monitoring Unit's type (numeric identifier)
// from /sys/bus/event_source/devices/<pmu>/type.
func getPMUEventType(eventType string) (uint32, error) {
	switch eventType {
	case "kprobe":
		kprobePMUType.once.Do(func() {
			kprobePMUType.value, kprobePMUType.err = parsePMUEventType(eventType)
		})
		return kprobePMUType.value, kprobePMUType.err
	case "uprobe":
		uprobePMUType.once.Do(func() {
			uprobePMUType.value, uprobePMUType.err = parsePMUEventType(eventType)
		})
		return uprobePMUType.value, uprobePMUType.err
	default:
		return 0, fmt.Errorf("unknown event type: %s", eventType)
	}
}

var (
	// kprobeRetProbeBit is used to cache the KProbe RetProbe bit value
	kprobeRetProbeBit = struct {
		once  sync.Once
		value uint64
		err   error
	}{}
	// uprobeRetProbeBit is used to cache the UProbe RetProbe bit value
	uprobeRetProbeBit = struct {
		once  sync.Once
		value uint64
		err   error
	}{}
)

// parseRetProbeBit reads a Performance Monitoring Unit's retprobe bit
// from /sys/bus/event_source/devices/<pmu>/format/retprobe.
func parseRetProbeBit(eventType string) (uint64, error) {
	p := filepath.Join("/sys/bus/event_source/devices/", eventType, "/format/retprobe")

	data, err := os.ReadFile(p)
	if err != nil {
		return 0, err
	}

	var rp uint64
	n, err := fmt.Sscanf(string(bytes.TrimSpace(data)), "config:%d", &rp)
	if err != nil {
		return 0, fmt.Errorf("parse retprobe bit: %w", err)
	}
	if n != 1 {
		return 0, fmt.Errorf("parse retprobe bit: expected 1 item, got %d", n)
	}

	return rp, nil
}

func getRetProbeBit(eventType string) (uint64, error) {
	switch eventType {
	case "kprobe":
		kprobeRetProbeBit.once.Do(func() {
			kprobeRetProbeBit.value, kprobeRetProbeBit.err = parseRetProbeBit(eventType)
		})
		return kprobeRetProbeBit.value, kprobeRetProbeBit.err
	case "uprobe":
		uprobeRetProbeBit.once.Do(func() {
			uprobeRetProbeBit.value, uprobeRetProbeBit.err = parseRetProbeBit(eventType)
		})
		return uprobeRetProbeBit.value, uprobeRetProbeBit.err
	default:
		return 0, fmt.Errorf("unknown event type %s", eventType)
	}
}

// getEnv retrieves the environment variable key. If it does not exist it returns the default.
func getEnv(key string, dfault string, combineWith ...string) string {
	value := os.Getenv(key)
	if value == "" {
		value = dfault
	}

	switch len(combineWith) {
	case 0:
		return value
	case 1:
		return filepath.Join(value, combineWith[0])
	default:
		all := make([]string, len(combineWith)+1)
		all[0] = value
		copy(all[1:], combineWith)
		return filepath.Join(all...)
	}
}

// hostProc returns joins the provided path with the host /proc directory
func hostProc(combineWith ...string) string {
	return getEnv("HOST_PROC", "/proc", combineWith...)
}

// Getpid returns the current process ID in the host namespace if $HOST_PROC is defined, the pid in the current namespace
// otherwise
func Getpid() int {
	p, err := os.Readlink(hostProc("/self"))
	if err == nil {
		if pid, err := strconv.ParseInt(p, 10, 32); err == nil {
			return int(pid)
		}
	}
	return os.Getpid()
}

// cleanupProgramSpec removes unused internal fields to free up some memory
func cleanupProgramSpec(spec *ebpf.ProgramSpec) {
	if spec != nil {
		spec.Instructions = nil
	}
}
