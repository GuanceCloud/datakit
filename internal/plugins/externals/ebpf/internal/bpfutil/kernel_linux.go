//go:build linux
// +build linux

package bpfutil

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/sys/unix"
)

var (
	kernelVersionOnce sync.Once
	kernelVersionVal  uint64
	errKernelVersion  error

	kernelSymbolsOnce sync.Once
	kernelSymbolsVal  map[string]struct{}
	errKernelSymbols  error

	mountProcFSOnce  sync.Once
	errMountProcFS   error
	mountDebugFSOnce sync.Once
	errMountDebugFS  error
)

func CurrentKernelVersion() (uint64, error) {
	kernelVersionOnce.Do(func() {
		var uname unix.Utsname
		if err := unix.Uname(&uname); err != nil {
			errKernelVersion = fmt.Errorf("uname: %w", err)
			return
		}

		var release strings.Builder
		for _, c := range uname.Release {
			if c == 0 {
				break
			}
			release.WriteByte(c)
		}

		kernelVersionVal, errKernelVersion = parseKernelRelease(release.String())
	})

	return kernelVersionVal, errKernelVersion
}

func parseKernelRelease(release string) (uint64, error) {
	versionPart := strings.SplitN(release, "-", 2)[0]
	parts := strings.Split(versionPart, ".")
	if len(parts) < 2 {
		return 0, fmt.Errorf("unexpected kernel release %q", release)
	}

	var numbers [3]uint64
	for idx := 0; idx < len(numbers) && idx < len(parts); idx++ {
		n, err := strconv.Atoi(parts[idx])
		if err != nil {
			return 0, fmt.Errorf("parse kernel release %q: %w", release, err)
		}
		numbers[idx] = uint64(n)
	}

	return numbers[0]<<48 | numbers[1]<<32 | numbers[2]<<16, nil
}

func kernelSupportsRetprobeMaxActiveOverride(kernelVersion uint64) bool {
	const minSupportedKernel = uint64(0x0004000f00000000) // 4.15.0
	return kernelVersion >= minSupportedKernel
}

func SupportsRetprobeMaxActiveOverride() (bool, uint64, error) {
	kernelVersion, err := CurrentKernelVersion()
	if err != nil {
		return false, 0, err
	}
	return kernelSupportsRetprobeMaxActiveOverride(kernelVersion), kernelVersion, nil
}

func KernelProbeSymbol(spec ProbeSpec) (string, bool) {
	if spec.KernelSymbol != "" {
		return spec.KernelSymbol, true
	}

	program := spec.ID.Program
	if program == "" {
		program = spec.ID.EBPFFuncName
	}

	switch {
	case strings.HasPrefix(program, "kprobe__"):
		return strings.TrimPrefix(program, "kprobe__"), true
	case strings.HasPrefix(program, "kretprobe__"):
		return strings.TrimPrefix(program, "kretprobe__"), true
	default:
		return "", false
	}
}

type KprobeCapability struct {
	PMUTypePath         string
	TraceFSKprobeEvents string
	DebugFSKprobeEvents string
}

func (c KprobeCapability) HasAnyInterface() bool {
	return c.PMUTypePath != "" || c.TraceFSKprobeEvents != "" || c.DebugFSKprobeEvents != ""
}

func (c KprobeCapability) MissingPaths() []string {
	paths := make([]string, 0, 3)
	if c.PMUTypePath == "" {
		paths = append(paths, "/sys/bus/event_source/devices/kprobe/type")
	}
	if c.TraceFSKprobeEvents == "" {
		paths = append(paths, "/sys/kernel/tracing/kprobe_events")
	}
	if c.DebugFSKprobeEvents == "" {
		paths = append(paths, "/sys/kernel/debug/tracing/kprobe_events")
	}
	return paths
}

func DetectKprobeCapability() KprobeCapability {
	caps := KprobeCapability{
		PMUTypePath:         firstExistingPath("/sys/bus/event_source/devices/kprobe/type"),
		TraceFSKprobeEvents: firstExistingPath("/sys/kernel/tracing/kprobe_events"),
		DebugFSKprobeEvents: firstExistingPath("/sys/kernel/debug/tracing/kprobe_events"),
	}
	if caps.HasAnyInterface() {
		return caps
	}

	_ = EnsureDebugFSMounted()
	return KprobeCapability{
		PMUTypePath:         firstExistingPath("/sys/bus/event_source/devices/kprobe/type"),
		TraceFSKprobeEvents: firstExistingPath("/sys/kernel/tracing/kprobe_events"),
		DebugFSKprobeEvents: firstExistingPath("/sys/kernel/debug/tracing/kprobe_events"),
	}
}

func firstExistingPath(paths ...string) string {
	for _, path := range paths {
		if path == "" {
			continue
		}
		if _, err := os.Stat(filepath.Clean(path)); err == nil {
			return path
		}
	}
	return ""
}

func HasKernelSymbol(name string) (bool, error) {
	kernelSymbolsOnce.Do(func() {
		f, err := openKernelSymbols()
		if err != nil {
			errKernelSymbols = err
			return
		}
		symbols, scanErr := readKernelSymbols(f)
		closeErr := f.Close()
		if scanErr != nil {
			errKernelSymbols = scanErr
			return
		}
		if closeErr != nil {
			errKernelSymbols = fmt.Errorf("close /proc/kallsyms: %w", closeErr)
			return
		}

		kernelSymbolsVal = symbols
	})

	if errKernelSymbols != nil {
		return false, errKernelSymbols
	}

	_, ok := kernelSymbolsVal[name]
	return ok, nil
}

func readKernelSymbols(f *os.File) (map[string]struct{}, error) {
	symbols := map[string]struct{}{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 {
			symbols[fields[2]] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan /proc/kallsyms: %w", err)
	}
	return symbols, nil
}

func openKernelSymbols() (*os.File, error) {
	f, err := os.Open("/proc/kallsyms")
	if err == nil {
		return f, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("open /proc/kallsyms: %w", err)
	}
	if mountErr := EnsureProcFSMounted(); mountErr != nil {
		return nil, fmt.Errorf("open /proc/kallsyms: %w (mount procfs: %v)", err, mountErr)
	}
	f, err = os.Open("/proc/kallsyms")
	if err != nil {
		return nil, fmt.Errorf("open /proc/kallsyms after mounting procfs: %w", err)
	}
	return f, nil
}

func EnsureProcFSMounted() error {
	mountProcFSOnce.Do(func() {
		if _, err := os.Stat("/proc/kallsyms"); err == nil {
			return
		}
		if err := unix.Mount("proc", "/proc", "proc", 0, ""); err != nil && !errors.Is(err, unix.EBUSY) {
			errMountProcFS = err
		}
	})
	return errMountProcFS
}

func EnsureDebugFSMounted() error {
	mountDebugFSOnce.Do(func() {
		if _, err := os.Stat("/sys/kernel/debug/tracing/kprobe_events"); err == nil {
			return
		}
		if err := unix.Mount("debugfs", "/sys/kernel/debug", "debugfs", 0, ""); err != nil && !errors.Is(err, unix.EBUSY) {
			errMountDebugFS = err
		}
	})
	return errMountDebugFS
}
