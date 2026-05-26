//go:build linux
// +build linux

// Package conntrack place probes on kernel functions
// `__nf_conntrack_hash_insert` and `nf_ct_delete`
package conntrack

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/cilium/ebpf"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/c"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"
	"golang.org/x/sys/unix"
)

const (
	componentID                  = "conntrack"
	conntrackTupleMapName        = "bpfmap_conntrack_tuple"
	conntrackUpdateFailMapName   = "bpfmap_conntrack_update_fail"
	conntrackDefaultObserveEvery = time.Minute
)

var conntrackInsertSymbols = []string{
	"nf_conntrack_hash_check_insert",
	"__nf_conntrack_hash_insert",
}

var log = logger.DefaultSLogger(componentID)

var conntrackConfirmSymbols = []string{
	"__nf_conntrack_confirm",
}

var conntrackDeleteSymbols = []string{
	"nf_ct_delete",
}

type HookSelection struct {
	InsertSymbol  string
	InsertSymbols []string
	DeleteSymbol  string
}

type originTupleKey struct {
	SrcIP   [4]uint32
	DstIP   [4]uint32
	SrcPort uint16
	DstPort uint16
	NetNS   uint32
}

type replyTupleValue struct {
	TS      uint64
	SrcIP   [4]uint32
	DstIP   [4]uint32
	SrcPort uint16
	DstPort uint16
	NetNS   uint32
}

var (
	tupleMapMu sync.RWMutex
	tupleMap   *ebpf.Map
)

func StartMapObserver(ctx context.Context, runtime *bpfutil.Runtime, interval time.Duration) {
	if runtime == nil {
		return
	}
	tupleMap, err := runtime.LookupMap(conntrackTupleMapName)
	if err != nil {
		return
	}
	failMap, _ := runtime.LookupMap(conntrackUpdateFailMapName)
	if interval <= 0 {
		interval = conntrackDefaultObserveEvery
	}

	go func() {
		observeConntrackMaps(tupleMap, failMap)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				observeConntrackMaps(tupleMap, failMap)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func observeConntrackMaps(tupleMap, failMap *ebpf.Map) {
	if tupleMap != nil {
		entries, err := countTupleMapEntries(tupleMap)
		if err != nil {
			exporter.IncBPFMapObserveError(componentID, conntrackTupleMapName, "iterate")
		} else {
			exporter.ObserveBPFMap(componentID, conntrackTupleMapName, entries, tupleMap.MaxEntries())
		}
	}

	if failMap == nil {
		return
	}
	var (
		key   uint32
		count uint64
	)
	if err := failMap.Lookup(&key, &count); err != nil {
		exporter.IncBPFMapObserveError(componentID, conntrackUpdateFailMapName, "lookup")
		return
	}
	exporter.ObserveCacheEntries(componentID, "tuple_update_fail_total", uint64ToInt(count))
}

func countTupleMapEntries(mp *ebpf.Map) (uint32, error) {
	var (
		key   originTupleKey
		reply replyTupleValue
		count uint32
	)

	iter := mp.Iterate()
	for iter.Next(&key, &reply) {
		count++
	}
	if err := iter.Err(); err != nil {
		return 0, err
	}
	return count, nil
}

func uint64ToInt(v uint64) int {
	if v > uint64(math.MaxInt) {
		return math.MaxInt
	}
	return int(v)
}

func SetTupleMap(mp *ebpf.Map) {
	tupleMapMu.Lock()
	defer tupleMapMu.Unlock()
	tupleMap = mp
}

func LookupDNATTuple(saddr, daddr [4]uint32, sport, dport, netns uint32) ([4]uint32, uint32, bool) {
	if sport == 0 || sport > math.MaxUint16 || dport == 0 || dport > math.MaxUint16 {
		return [4]uint32{}, 0, false
	}

	key := originTupleKey{
		SrcIP:   saddr,
		DstIP:   daddr,
		SrcPort: uint16(sport),
		DstPort: uint16(dport),
		NetNS:   netns,
	}

	reply, ok := lookupReplyTuple(key)
	if !ok && key.NetNS != 0 {
		key.NetNS = 0
		reply, ok = lookupReplyTuple(key)
	}
	if !ok || reply.SrcPort == 0 || isZeroIPv46(reply.SrcIP) {
		return [4]uint32{}, 0, false
	}

	return reply.SrcIP, uint32(reply.SrcPort), true
}

func lookupReplyTuple(key originTupleKey) (replyTupleValue, bool) {
	tupleMapMu.RLock()
	mp := tupleMap
	tupleMapMu.RUnlock()
	if mp == nil {
		return replyTupleValue{}, false
	}

	var reply replyTupleValue
	if err := mp.Lookup(&key, &reply); err != nil {
		return replyTupleValue{}, false
	}
	return reply, true
}

func isZeroIPv46(addr [4]uint32) bool {
	return addr[0]|addr[1]|addr[2]|addr[3] == 0
}

func DumpTupleMap(limit int) []string {
	tupleMapMu.RLock()
	mp := tupleMap
	tupleMapMu.RUnlock()
	if mp == nil || limit <= 0 {
		return nil
	}

	lines := make([]string, 0, limit)
	var (
		key   originTupleKey
		reply replyTupleValue
	)

	iter := mp.Iterate()
	for iter.Next(&key, &reply) {
		lines = append(lines, fmt.Sprintf(
			"origin=%s:%d->%s:%d netns=%d reply=%s:%d->%s:%d rnetns=%d",
			formatTupleIP(key.SrcIP), key.SrcPort,
			formatTupleIP(key.DstIP), key.DstPort,
			key.NetNS,
			formatTupleIP(reply.SrcIP), reply.SrcPort,
			formatTupleIP(reply.DstIP), reply.DstPort,
			reply.NetNS,
		))
		if len(lines) >= limit {
			break
		}
	}
	if err := iter.Err(); err != nil {
		lines = append(lines, fmt.Sprintf("iterate_error=%v", err))
	}
	return lines
}

func formatTupleIP(addr [4]uint32) string {
	if addr[0] == 0 && addr[1] == 0 && addr[2] == 0 {
		return formatIPv4(addr[3])
	}
	parts := make([]string, 0, len(addr))
	for _, v := range addr {
		parts = append(parts, strconv.FormatUint(uint64(v), 16))
	}
	return strings.Join(parts, ":")
}

func formatIPv4(v uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(v),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24),
	)
}

func conntrackKprobeInterfaceAvailableAt(paths []string) bool {
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

func conntrackKprobeInterfaceAvailable() bool {
	return conntrackKprobeInterfaceAvailableAt([]string{
		"/sys/bus/event_source/devices/kprobe/type",
		"/sys/kernel/debug/tracing/kprobe_events",
		"/sys/kernel/tracing/kprobe_events",
	})
}

func kernelSymbolsAvailable(symbolsText string, names []string) bool {
	return firstAvailableKernelSymbol(symbolsText, names) != ""
}

func firstAvailableKernelSymbol(symbolsText string, names []string) string {
	symbols := availableKernelSymbolsOrdered(symbolsText, names)
	if len(symbols) == 0 {
		return ""
	}
	return symbols[0]
}

func availableKernelSymbolsOrdered(symbolsText string, names []string) []string {
	want := make(map[string]struct{}, len(names))
	for _, name := range names {
		want[name] = struct{}{}
	}
	found := make(map[string]struct{}, len(names))

	for _, line := range strings.Split(symbolsText, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		if _, ok := want[fields[2]]; ok {
			found[fields[2]] = struct{}{}
		}
	}

	ordered := make([]string, 0, len(names))
	for _, name := range names {
		if _, ok := found[name]; ok {
			ordered = append(ordered, name)
		}
	}
	return ordered
}

func resolveInsertSymbols(symbolsText string, kernelVersion uint64) []string {
	insertNames := conntrackInsertSymbols
	if kernelSymbolsAvailable(symbolsText, conntrackConfirmSymbols) {
		insertNames = append(append([]string{}, conntrackConfirmSymbols...), insertNames...)
	}

	return availableKernelSymbolsOrdered(symbolsText, insertNames)
}

func resolveHookSelectionWithKernelVersion(symbolsText string, kernelVersion uint64) HookSelection {
	insertSymbols := resolveInsertSymbols(symbolsText, kernelVersion)
	insertSymbol := ""
	if len(insertSymbols) > 0 {
		insertSymbol = insertSymbols[0]
	}
	recordConntrackKernelFunctions(insertSymbols, firstAvailableKernelSymbol(symbolsText, conntrackDeleteSymbols))

	return HookSelection{
		InsertSymbol:  insertSymbol,
		InsertSymbols: insertSymbols,
		DeleteSymbol:  firstAvailableKernelSymbol(symbolsText, conntrackDeleteSymbols),
	}
}

func recordConntrackKernelFunctions(insertSymbols []string, deleteSymbol string) {
	found := make(map[string]struct{}, len(insertSymbols)+1)
	for _, symbol := range insertSymbols {
		found[symbol] = struct{}{}
	}
	if deleteSymbol != "" {
		found[deleteSymbol] = struct{}{}
	}

	candidates := append(append(append([]string{}, conntrackConfirmSymbols...), conntrackInsertSymbols...), conntrackDeleteSymbols...)
	for _, symbol := range candidates {
		status := "missing"
		reason := "kernel symbol not found"
		if _, ok := found[symbol]; ok {
			status = "available"
			reason = ""
		}
		exporter.RecordKernelFunctionStatus("conntrack", conntrackProgramName(symbol), symbol, status, reason)
	}
}

func resolveHookSelection(symbolsText string) HookSelection {
	kernelVersion, err := bpfutil.CurrentKernelVersion()
	if err != nil {
		kernelVersion = 0
	}

	return resolveHookSelectionWithKernelVersion(symbolsText, kernelVersion)
}

func ResolveConntrackHookSelection() (HookSelection, error) {
	if !conntrackKprobeInterfaceAvailable() {
		return HookSelection{}, nil
	}

	symbols, err := os.ReadFile("/proc/kallsyms")
	if err != nil {
		for _, symbol := range append(append([]string{}, conntrackInsertSymbols...), conntrackDeleteSymbols...) {
			exporter.RecordKernelFunctionStatus("conntrack", conntrackProgramName(symbol), symbol, "detect_error", err.Error())
		}
		return HookSelection{}, fmt.Errorf("read /proc/kallsyms: %w", err)
	}

	return resolveHookSelection(string(symbols)), nil
}

const (
	conntrackConfirmSymbol         = "__nf_conntrack_confirm"
	conntrackHashCheckInsertSymbol = "nf_conntrack_hash_check_insert"
	conntrackHashInsertSymbol      = "__nf_conntrack_hash_insert"
)

func insertProgramName(symbol string) string {
	switch symbol {
	case conntrackConfirmSymbol:
		return "kprobe___nf_conntrack_confirm"
	case conntrackHashCheckInsertSymbol:
		return "kprobe__nf_conntrack_hash_check_insert"
	case conntrackHashInsertSymbol:
		return "kprobe___nf_conntrack_hash_insert"
	default:
		return ""
	}
}

func conntrackProgramName(symbol string) string {
	if program := insertProgramName(symbol); program != "" {
		return program
	}
	switch symbol {
	case "nf_ct_delete":
		return "kprobe__nf_ct_delete"
	default:
		return ""
	}
}

func ConntrackInsertProgramNames(symbols []string) []string {
	programs := make([]string, 0, len(symbols))
	seen := make(map[string]struct{}, len(symbols))

	for _, symbol := range symbols {
		program := insertProgramName(symbol)
		if program == "" {
			continue
		}
		if _, ok := seen[program]; ok {
			continue
		}
		seen[program] = struct{}{}
		programs = append(programs, program)
	}

	return programs
}

func ConntrackHooksAvailable() (bool, string, error) {
	if !conntrackKprobeInterfaceAvailable() {
		return false, "kprobe interface unavailable", nil
	}

	symbols, err := ResolveConntrackHookSelection()
	if err != nil {
		return false, "", err
	}
	if symbols.InsertSymbol == "" {
		return false, "conntrack insert kernel symbols unavailable", nil
	}
	return true, "", nil
}

func NewConntrackRuntime(patches []bpfutil.ConstantPatch) (*bpfutil.Runtime, error) {
	hooks, err := ResolveConntrackHookSelection()
	if err != nil {
		return nil, err
	}

	insertPrograms := ConntrackInsertProgramNames(hooks.InsertSymbols)
	if len(insertPrograms) == 0 {
		return nil, fmt.Errorf("conntrack insert kernel symbols unavailable")
	}

	useLegacyConsts, _, err := bpfutil.UseLegacyConstObjects()
	if err != nil {
		return nil, err
	}

	probes := make([]*bpfutil.HookSpec, 0, len(insertPrograms)+1)
	for _, insertProgram := range insertPrograms {
		probes = append(probes, &bpfutil.HookSpec{
			ID: bpfutil.HookID{
				Program: insertProgram,
			},
		})
	}
	if hooks.DeleteSymbol != "" {
		probes = append(probes, &bpfutil.HookSpec{
			ID: bpfutil.HookID{
				Program: "kprobe__nf_ct_delete",
			},
		})
	}

	runtime := &bpfutil.Runtime{
		Probes: probes,
	}
	loadSpec := bpfutil.LoadSpec{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
		Constants:       patches,
		LegacyConstants: useLegacyConsts,
	}

	bufLoader := dkebpf.ConntrackBin
	binName := "conntrack.o"
	if useLegacyConsts {
		bufLoader = dkebpf.ConntrackLegacyBin
		binName = "conntrack_legacy.o"
	}

	loadRuntime := func(name string, loader func() ([]byte, error), legacyConstants bool) error {
		loadSpec.LegacyConstants = legacyConstants
		buf, err := loader()
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		if err := runtime.LoadFromReader((bytes.NewReader(buf)), loadSpec); err != nil {
			return fmt.Errorf("init conntrack tracer: %w", err)
		}
		return nil
	}
	if err := loadRuntime(binName, bufLoader, useLegacyConsts); err != nil {
		if useLegacyConsts {
			return nil, err
		}
		log.Warnf("load modern conntrack object failed, fallback to legacy object without LRU hash maps: %v", err)
		if err := loadRuntime("conntrack_legacy.o", dkebpf.ConntrackLegacyBin, true); err != nil {
			return nil, err
		}
	}
	return runtime, nil
}
