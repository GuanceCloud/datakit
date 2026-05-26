//go:build linux
// +build linux

package bpfutil

import (
	"debug/elf"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"golang.org/x/sys/unix"
)

type Constant struct {
	Name  string
	Value interface{}
}

type ConstantPatch = Constant

type ProbeID struct {
	UID          string
	Program      string
	EBPFFuncName string
}

type HookID = ProbeID

type ProbeSpec struct {
	ID              ProbeID
	KProbeMaxActive int
	KernelSymbol    string
	UprobeOffset    uint64
	BinaryPath      string
	SocketFD        int
}

type HookSpec = ProbeSpec

type PerfHandler func(cpu int, data []byte, reader *PerfStream, runtime *Runtime)

type PerfLostHandler func(cpu int, count uint64, reader *PerfStream, runtime *Runtime)

type PerfErrorHandler func(err error, reader *PerfStream, runtime *Runtime)

type PerfReaderSpec struct {
	MapName      string
	BufferSize   int
	Watermark    int
	DataHandler  PerfHandler
	LostHandler  PerfLostHandler
	ErrorHandler PerfErrorHandler
	PerfErrChan  chan error
}

type PerfStreamSpec = PerfReaderSpec

type LoadOptions struct {
	Constants        []Constant
	LegacyConstants  bool
	DisabledPrograms []string
	MapReplacements  map[string]*ebpf.Map
	MapMaxEntries    map[string]uint32
	RLimit           *unix.Rlimit
	VerifierOptions  ebpf.CollectionOptions
}

type LoadSpec = LoadOptions

type attachedProbe struct {
	ProbeSpec
	program     *ebpf.Program
	sectionName string
	link        link.Link
}

type PerfStream struct {
	Map
	PerfStreamOptions

	spec   PerfReaderSpec
	reader *perf.Reader
	wg     sync.WaitGroup
}

type PerfStreamOptions struct {
	PerfRingBufferSize int
	Watermark          int
	PerfErrChan        chan error
	DataHandler        PerfHandler
	LostHandler        PerfLostHandler
	ErrorHandler       PerfErrorHandler
}

const (
	smallPerfDefaultPages    = 32
	smallPerfDefaultMaxBytes = 16 * 1024 * 1024
	smallPerfMinPages        = 4
)

func PerfRingBufferSize(defaultPages, maxBytes, minPages int) int {
	pageSize := os.Getpagesize()
	if defaultPages <= 0 {
		defaultPages = 1
	}
	if minPages <= 0 {
		minPages = 1
	}

	pages := defaultPages
	if maxBytes > 0 && pageSize > 0 {
		if capPages := maxBytes / pageSize / runtime.NumCPU(); capPages > 0 && capPages < pages {
			pages = capPages
		}
	}
	if pages < minPages {
		pages = minPages
	}
	return pages * pageSize
}

func SmallPerfRingBufferSize() int {
	return PerfRingBufferSize(smallPerfDefaultPages, smallPerfDefaultMaxBytes, smallPerfMinPages)
}

func PerfWatermark(bufferSize int) int {
	pageSize := os.Getpagesize()
	if bufferSize <= pageSize || pageSize <= 0 {
		return 0
	}
	return pageSize
}

type Runtime struct {
	collection *ebpf.Collection
	spec       *ebpf.CollectionSpec
	maps       map[string]*ebpf.Map

	staticProbes  []*attachedProbe
	dynamicProbes map[string]*attachedProbe
	perfReaders   []*PerfStream

	started bool
	mu      sync.RWMutex

	Probes  []*HookSpec
	Streams []*PerfStream
}

type Map struct {
	Name string
}

func LoadRuntimeFromReader(reader io.ReaderAt, opts LoadOptions, probes []ProbeSpec, perfSpecs []PerfReaderSpec) (*Runtime, error) {
	if opts.RLimit != nil {
		if err := unix.Setrlimit(unix.RLIMIT_MEMLOCK, opts.RLimit); err != nil {
			return nil, fmt.Errorf("setrlimit memlock: %w", err)
		}
	}

	spec, err := ebpf.LoadCollectionSpecFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("load collection spec: %w", err)
	}

	constants := opts.Constants
	if len(constants) > 0 {
		if opts.LegacyConstants {
			if err := rewriteLegacyConstants(spec, constants); err != nil {
				return nil, fmt.Errorf("rewrite legacy constants: %w", err)
			}
		} else {
			consts := make(map[string]interface{}, len(constants))
			for _, constant := range constants {
				consts[constant.Name] = constant.Value
			}
			if err := spec.RewriteConstants(consts); err != nil {
				return nil, fmt.Errorf("rewrite constants: %w", err)
			}
		}
	}

	if len(opts.DisabledPrograms) > 0 {
		for _, name := range opts.DisabledPrograms {
			delete(spec.Programs, name)
		}
	}
	for name, maxEntries := range opts.MapMaxEntries {
		if maxEntries == 0 {
			continue
		}
		if mp, ok := spec.Maps[name]; ok {
			mp.MaxEntries = maxEntries
		}
	}

	collOpts := opts.VerifierOptions
	mapReplacements := opts.MapReplacements
	if len(mapReplacements) > 0 {
		if collOpts.MapReplacements == nil {
			collOpts.MapReplacements = make(map[string]*ebpf.Map, len(mapReplacements))
		}
		for name, mp := range mapReplacements {
			collOpts.MapReplacements[name] = mp
		}
	}

	coll, err := ebpf.NewCollectionWithOptions(spec, collOpts)
	if err != nil {
		var verifierErr *ebpf.VerifierError
		if errors.As(err, &verifierErr) {
			return nil, fmt.Errorf("load collection: %w\nverifier log:\n%s", err, verifierErr)
		}
		return nil, fmt.Errorf("load collection: %w", err)
	}

	rt := &Runtime{
		collection:    coll,
		spec:          spec,
		maps:          make(map[string]*ebpf.Map, len(coll.Maps)),
		dynamicProbes: make(map[string]*attachedProbe),
	}

	for name, mp := range coll.Maps {
		rt.maps[name] = mp
	}

	for _, probe := range probes {
		ap, err := rt.prepareProbe(probe)
		if err != nil {
			coll.Close()
			return nil, err
		}
		rt.staticProbes = append(rt.staticProbes, ap)
	}

	for _, perfSpec := range perfSpecs {
		rt.perfReaders = append(rt.perfReaders, &PerfStream{spec: perfSpec})
	}

	return rt, nil
}

func (r *Runtime) LoadCollection(reader io.ReaderAt, opts LoadOptions) error {
	probes := make([]ProbeSpec, 0, len(r.Probes))
	for _, probe := range r.Probes {
		if probe == nil {
			continue
		}
		probes = append(probes, *probe)
	}

	perfSpecs := make([]PerfReaderSpec, 0, len(r.Streams))
	for _, stream := range r.Streams {
		if stream == nil {
			continue
		}
		perfSpecs = append(perfSpecs, PerfReaderSpec{
			MapName:      stream.Map.Name,
			BufferSize:   stream.PerfStreamOptions.PerfRingBufferSize,
			Watermark:    stream.PerfStreamOptions.Watermark,
			PerfErrChan:  stream.PerfStreamOptions.PerfErrChan,
			DataHandler:  stream.PerfStreamOptions.DataHandler,
			LostHandler:  stream.PerfStreamOptions.LostHandler,
			ErrorHandler: stream.PerfStreamOptions.ErrorHandler,
		})
	}

	loaded, err := LoadRuntimeFromReader(reader, opts, probes, perfSpecs)
	if err != nil {
		return err
	}

	r.collection = loaded.collection
	r.spec = loaded.spec
	r.maps = loaded.maps
	r.staticProbes = loaded.staticProbes
	r.dynamicProbes = loaded.dynamicProbes
	r.perfReaders = loaded.perfReaders
	r.started = loaded.started
	return nil
}

func (r *Runtime) LoadFromReader(reader io.ReaderAt, spec LoadSpec) error {
	return r.LoadCollection(reader, spec)
}

func (r *Runtime) prepareProbe(spec ProbeSpec) (*attachedProbe, error) {
	spec = normalizeProbeSpec(spec)
	if spec.ID.Program == "" {
		return nil, fmt.Errorf("empty program name")
	}

	prog, ok := r.collection.Programs[spec.ID.Program]
	if !ok {
		return nil, fmt.Errorf("program %q not found", spec.ID.Program)
	}

	ap := &attachedProbe{
		ProbeSpec: spec,
		program:   prog,
	}

	if ps, ok := r.spec.Programs[spec.ID.Program]; ok {
		ap.sectionName = ps.SectionName
	}
	if ap.sectionName == "" {
		ap.sectionName = spec.ID.Program
	}

	return ap, nil
}

func normalizeProbeSpec(spec ProbeSpec) ProbeSpec {
	if spec.ID.Program == "" {
		spec.ID.Program = spec.ID.EBPFFuncName
	}
	return spec
}

func probeKey(id ProbeID) string {
	return id.UID + "\x00" + id.Program
}

func (r *Runtime) Start() error {
	r.mu.Lock()
	if r.started {
		r.mu.Unlock()
		return nil
	}
	staticProbes := append([]*attachedProbe(nil), r.staticProbes...)
	perfReaders := append([]*PerfStream(nil), r.perfReaders...)
	r.mu.Unlock()

	for _, probe := range staticProbes {
		if err := attachProbe(probe); err != nil {
			return fmt.Errorf("attach probe %q (%s): %w", probe.ID.Program, probe.sectionName, err)
		}
	}

	for _, reader := range perfReaders {
		if err := reader.start(r); err != nil {
			return fmt.Errorf("start perf reader %q: %w", reader.spec.MapName, err)
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.started = true
	return nil
}

func (r *Runtime) StartRuntime() error {
	return r.Start()
}

func (r *Runtime) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var firstErr error

	for _, reader := range r.perfReaders {
		if err := reader.close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	for _, probe := range r.dynamicProbes {
		if err := detachProbe(probe); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	for _, probe := range r.staticProbes {
		if err := detachProbe(probe); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if r.collection != nil {
		r.collection.Close()
		r.collection = nil
	}

	r.started = false
	return firstErr
}

func (r *Runtime) Shutdown() error {
	return r.Close()
}

func (r *Runtime) LookupMap(name string) (*ebpf.Map, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	mp, ok := r.maps[name]
	if !ok {
		return nil, fmt.Errorf("map %q not found", name)
	}
	return mp, nil
}

func (r *Runtime) LookupHook(id HookID) (*HookSpec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if probe, ok := r.dynamicProbes[probeKey(id)]; ok {
		return &probe.ProbeSpec, true
	}

	for _, probe := range r.staticProbes {
		if probeKey(probe.ID) == probeKey(id) {
			return &probe.ProbeSpec, true
		}
	}

	return nil, false
}

func (r *Runtime) AttachHook(spec HookSpec) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ap, err := r.prepareProbe(spec)
	if err != nil {
		return err
	}
	if err := attachProbe(ap); err != nil {
		return err
	}
	r.dynamicProbes[probeKey(spec.ID)] = ap
	return nil
}

func (r *Runtime) DetachHookSpec(id HookID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ap, ok := r.dynamicProbes[probeKey(id)]
	if !ok {
		return fmt.Errorf("probe not found: %+v", id)
	}
	if err := detachProbe(ap); err != nil {
		return err
	}
	delete(r.dynamicProbes, probeKey(id))
	return nil
}

func (p *ProbeSpec) Program() *ebpf.Program {
	return nil
}

func attachProbe(probe *attachedProbe) error {
	if probe.link != nil {
		return nil
	}

	section := probe.sectionName
	switch {
	case strings.HasPrefix(section, "kprobe/"):
		symbol := strings.TrimPrefix(section, "kprobe/")
		if probe.KernelSymbol != "" {
			symbol = probe.KernelSymbol
		}
		lnk, err := link.Kprobe(symbol, probe.program, nil)
		if err != nil {
			return err
		}
		probe.link = lnk
		return nil
	case strings.HasPrefix(section, "kretprobe/"):
		symbol := strings.TrimPrefix(section, "kretprobe/")
		if probe.KernelSymbol != "" {
			symbol = probe.KernelSymbol
		}
		var opts *link.KprobeOptions
		if probe.KProbeMaxActive > 0 {
			if supported, _, err := SupportsRetprobeMaxActiveOverride(); err == nil {
				if supported {
					opts = &link.KprobeOptions{RetprobeMaxActive: probe.KProbeMaxActive}
				}
			} else {
				return fmt.Errorf("detect kretprobe maxactive support: %w", err)
			}
		}
		lnk, err := link.Kretprobe(symbol, probe.program, opts)
		if err != nil {
			return err
		}
		probe.link = lnk
		return nil
	case strings.HasPrefix(section, "tracepoint/"):
		rest := strings.TrimPrefix(section, "tracepoint/")
		parts := strings.SplitN(rest, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid tracepoint section %q", section)
		}
		lnk, err := link.Tracepoint(parts[0], parts[1], probe.program, nil)
		if err != nil {
			return err
		}
		probe.link = lnk
		return nil
	case strings.HasPrefix(section, "uprobe/"), strings.HasPrefix(section, "uretprobe/"):
		if probe.BinaryPath == "" {
			return fmt.Errorf("binary path is empty for %q", probe.ID.Program)
		}
		ex, err := link.OpenExecutable(probe.BinaryPath)
		if err != nil {
			return err
		}
		var symbol string
		if strings.HasPrefix(section, "uprobe/") {
			symbol = strings.TrimPrefix(section, "uprobe/")
		} else {
			symbol = strings.TrimPrefix(section, "uretprobe/")
		}
		// Keep a stable non-empty symbol name even for offset-based uprobes.
		// On older kernels, cilium/ebpf falls back from perf_uprobe PMU to
		// tracefs and uses the symbol to derive the trace event name.
		// Address-based attachment still works because opts.Address takes
		// precedence over symbol resolution inside link.Executable.
		opts := &link.UprobeOptions{Address: probe.UprobeOffset}
		if strings.HasPrefix(section, "uretprobe/") {
			lnk, err := ex.Uretprobe(symbol, probe.program, opts)
			if err != nil {
				return err
			}
			probe.link = lnk
			return nil
		}
		lnk, err := ex.Uprobe(symbol, probe.program, opts)
		if err != nil {
			return err
		}
		probe.link = lnk
		return nil
	case strings.HasPrefix(section, "socket"):
		if probe.SocketFD <= 0 {
			return fmt.Errorf("socket fd is empty for %q", probe.ID.Program)
		}
		return unix.SetsockoptInt(probe.SocketFD, unix.SOL_SOCKET, unix.SO_ATTACH_BPF, probe.program.FD())
	default:
		return fmt.Errorf("unsupported section %q for %q", section, probe.ID.Program)
	}
}

func detachProbe(probe *attachedProbe) error {
	if probe == nil {
		return nil
	}
	if probe.link != nil {
		err := probe.link.Close()
		probe.link = nil
		return err
	}
	if probe.SocketFD > 0 {
		return unix.SetsockoptInt(probe.SocketFD, unix.SOL_SOCKET, unix.SO_DETACH_BPF, 0)
	}
	return nil
}

func (s *PerfStream) start(runtime *Runtime) error {
	if s.reader != nil {
		return nil
	}
	mp, err := runtime.LookupMap(s.spec.MapName)
	if err != nil {
		return err
	}
	size := s.spec.BufferSize
	if size == 0 {
		size = os.Getpagesize()
	}
	reader, err := perf.NewReaderWithOptions(mp, size, perf.ReaderOptions{Watermark: s.spec.Watermark})
	if err != nil {
		return err
	}
	s.reader = reader
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		var record perf.Record
		for {
			if err := s.reader.ReadInto(&record); err != nil {
				if errors.Is(err, perf.ErrClosed) {
					return
				}
				if s.spec.ErrorHandler != nil {
					s.spec.ErrorHandler(err, s, runtime)
				}
				if s.spec.PerfErrChan != nil {
					s.spec.PerfErrChan <- err
				}
				return
			}
			if record.LostSamples > 0 {
				if s.spec.LostHandler != nil {
					s.spec.LostHandler(record.CPU, record.LostSamples, s, runtime)
				}
				record.LostSamples = 0
				record.RawSample = record.RawSample[:0]
				continue
			}
			if s.spec.DataHandler != nil {
				s.spec.DataHandler(record.CPU, record.RawSample, s, runtime)
			}
			record.LostSamples = 0
			record.RawSample = record.RawSample[:0]
		}
	}()
	return nil
}

func (s *PerfStream) close() error {
	if s.reader == nil {
		return nil
	}
	err := s.reader.Close()
	s.wg.Wait()
	s.reader = nil
	return err
}

func SanitizeUprobeAddresses(f *elf.File, syms []elf.Symbol) {
	if f == nil {
		return
	}
	for idx := range syms {
		if f.Type != elf.ET_EXEC && f.Type != elf.ET_DYN {
			continue
		}
		for _, prog := range f.Progs {
			if prog.Type != elf.PT_LOAD {
				continue
			}
			if syms[idx].Value >= prog.Vaddr && syms[idx].Value < prog.Vaddr+prog.Memsz {
				syms[idx].Value = syms[idx].Value - prog.Vaddr + prog.Off
				break
			}
		}
	}
}
