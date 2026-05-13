//go:build linux
// +build linux

package netflow

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/cilium/ebpf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/c"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/cli"
	"golang.org/x/sys/unix"
)

var (
	errCriticalNetflowProbeMissing = errors.New("critical netflow probe missing")
	errCriticalNetflowProbeLoad    = errors.New("critical netflow probe load failed")
	errCriticalNetflowProbeAttach  = errors.New("critical netflow probe attach failed")
	errNetflowKprobeUnavailable    = errors.New("netflow kprobe interface unavailable")
)

type netflowProbeFailure struct {
	Program string
	Symbol  string
	Reason  string
}

func filterUnavailableKernelProbes(probes []*bpfutil.HookSpec) ([]*bpfutil.HookSpec, *netflowProbeFailure) {
	if len(probes) == 0 {
		return probes, nil
	}

	filtered := make([]*bpfutil.HookSpec, 0, len(probes))
	for _, probe := range probes {
		if probe == nil {
			continue
		}

		symbol, ok := bpfutil.KernelProbeSymbol(*probe)
		if !ok {
			filtered = append(filtered, probe)
			continue
		}

		found, err := bpfutil.HasKernelSymbol(symbol)
		if err != nil {
			exporter.RecordKernelFunctionStatus("netflow", probe.ID.Program, symbol, "unknown", err.Error())
			l.Warnf("detect kernel symbol %q for probe %q failed: %v",
				symbol, probe.ID.Program, err)
			filtered = append(filtered, probe)
			continue
		}
		if !found {
			exporter.RecordKernelFunctionStatus("netflow", probe.ID.Program, symbol, "missing", "kernel symbol not found")
			if isCriticalNetflowProgram(probe.ID.Program) {
				return nil, &netflowProbeFailure{
					Program: probe.ID.Program,
					Symbol:  symbol,
					Reason:  "kernel symbol not found",
				}
			}
			l.Warnf("skip netflow probe %q: kernel symbol %q not found",
				probe.ID.Program, symbol)
			continue
		}

		exporter.RecordKernelFunctionStatus("netflow", probe.ID.Program, symbol, "available", "")
		filtered = append(filtered, probe)
	}

	return filtered, nil
}

func filterDisabledNetflowProbes(probes []*bpfutil.HookSpec, disabledPrograms []string) []*bpfutil.HookSpec {
	if len(probes) == 0 || len(disabledPrograms) == 0 {
		return probes
	}

	disabled := make(map[string]struct{}, len(disabledPrograms))
	for _, name := range disabledPrograms {
		disabled[name] = struct{}{}
	}

	filtered := make([]*bpfutil.HookSpec, 0, len(probes))
	for _, probe := range probes {
		if probe == nil {
			continue
		}
		if _, ok := disabled[probe.ID.Program]; ok {
			symbol, _ := bpfutil.KernelProbeSymbol(*probe)
			exporter.RecordKernelFunctionStatus("netflow", probe.ID.Program, symbol, "disabled", "disabled before runtime start")
			continue
		}
		filtered = append(filtered, probe)
	}

	return filtered
}

const (
	NoValue           = "N/A"
	DirectionOutgoing = "outgoing"
	DirectionIncoming = "incoming"
	DirectionUnknown  = "unknown"
)

var ephemeralPortMin int32 = EphemeralPortMin

func SetEphemeralPortMin(val int32) {
	if val <= 0 {
		val = EphemeralPortMin
	}
	l.Debugf("ephemeral port start from %d", val)
	atomic.StoreInt32(&ephemeralPortMin, val)
}

var l = logger.DefaultSLogger("ebpf")

const minIPv6NetflowKernel = uint64(0x0004000700000000) // 4.7.0

const (
	netflowMapMaxEntriesEnv     = "DK_EBPF_NETFLOW_MAP_MAX_ENTRIES"
	defaultNetflowMapMaxEntries = 65536
	minNetflowMapMaxEntries     = 1024
	maxNetflowMapMaxEntries     = 1048576
)

var criticalNetflowPrograms = map[string]struct{}{
	"kprobe__sockfd_lookup_light":    {},
	"kretprobe__sockfd_lookup_light": {},
	"kprobe__do_sendfile":            {},
	"kretprobe__do_sendfile":         {},
	"kprobe__udp_recvmsg":            {},
	"kretprobe__udp_recvmsg":         {},
	"kprobe__inet_bind":              {},
	"kretprobe__inet_bind":           {},
	"kprobe__inet6_bind":             {},
	"kretprobe__inet6_bind":          {},
	"kprobe__udp_destroy_sock":       {},
	"kprobe__tcp_close":              {},
}

func isCriticalNetflowProgram(program string) bool {
	_, ok := criticalNetflowPrograms[program]
	return ok
}

type dnsRecorder interface {
	LookupAddr(ip string) string
}

var dnsRecord dnsRecorder

var k8sNetInfo *cli.K8sInfo

type addrDomainEntry struct {
	domain string
	ts     time.Time
}

type peerDomainKey struct {
	ip        string
	port      uint32
	transport string
	netns     string
}

type addrDomainRecord struct {
	sync.RWMutex
	ipRecord   map[string]addrDomainEntry
	peerRecord map[peerDomainKey]addrDomainEntry
}

const addrDomainTTL = 10 * time.Minute

var sharedAddrDomainRecord = &addrDomainRecord{
	ipRecord:   map[string]addrDomainEntry{},
	peerRecord: map[peerDomainKey]addrDomainEntry{},
}

func SetDNSRecord(r dnsRecorder) {
	dnsRecord = r
}

func (r *addrDomainRecord) RecordAddrDomain(ip, domain string) {
	if ip == "" || domain == "" {
		return
	}

	now := ntp.Now()

	r.Lock()
	defer r.Unlock()

	r.ipRecord[ip] = addrDomainEntry{
		domain: domain,
		ts:     now,
	}

	for k, v := range r.ipRecord {
		if now.Sub(v.ts) > addrDomainTTL {
			delete(r.ipRecord, k)
		}
	}

	for k, v := range r.peerRecord {
		if now.Sub(v.ts) > addrDomainTTL {
			delete(r.peerRecord, k)
		}
	}
}

func (r *addrDomainRecord) RecordPeerDomain(ip string, port uint32, transport, netns, domain string) {
	if ip == "" || domain == "" {
		return
	}

	r.Lock()
	defer r.Unlock()

	r.peerRecord[peerDomainKey{
		ip:        ip,
		port:      port,
		transport: transport,
		netns:     netns,
	}] = addrDomainEntry{
		domain: domain,
		ts:     ntp.Now(),
	}
}

func (r *addrDomainRecord) lookupAddr(ip string, now time.Time) string {
	v, ok := r.ipRecord[ip]
	if !ok {
		return ""
	}

	if now.Sub(v.ts) > addrDomainTTL {
		return ""
	}

	return v.domain
}

func (r *addrDomainRecord) LookupPeerDomain(ip string, port uint32, transport, netns string) string {
	if ip == "" {
		return ""
	}

	now := ntp.Now()

	r.RLock()
	v, ok := r.peerRecord[peerDomainKey{
		ip:        ip,
		port:      port,
		transport: transport,
		netns:     netns,
	}]
	if ok && now.Sub(v.ts) <= addrDomainTTL {
		domain := v.domain
		r.RUnlock()
		return domain
	}
	domain := r.lookupAddr(ip, now)
	r.RUnlock()

	if domain != "" {
		return domain
	}
	if !ok {
		return ""
	}

	if now.Sub(v.ts) > addrDomainTTL {
		key := peerDomainKey{
			ip:        ip,
			port:      port,
			transport: transport,
			netns:     netns,
		}
		r.Lock()
		if cur, ok := r.peerRecord[key]; ok && now.Sub(cur.ts) > addrDomainTTL {
			delete(r.peerRecord, key)
		}
		r.Unlock()
		return ""
	}

	return ""
}

func RecordAddrDomain(ip, domain string) {
	sharedAddrDomainRecord.RecordAddrDomain(ip, domain)
}

func RecordPeerDomain(ip string, port uint32, transport, netns, domain string) {
	sharedAddrDomainRecord.RecordPeerDomain(ip, port, transport, netns, domain)
}

func LookupPeerDomain(ip string, port uint32, transport, netns string) string {
	return sharedAddrDomainRecord.LookupPeerDomain(ip, port, transport, netns)
}

func SetLogger(nl *logger.Logger) {
	l = nl
}

func SetK8sNetInfo(n *cli.K8sInfo) {
	k8sNetInfo = n
}

var SrcIPPortRecorder = func() *srcIPPortRecorder {
	ptr := &srcIPPortRecorder{
		Record: map[[4]uint32]IPPortRecord{},
	}
	go ptr.AutoClean()
	return ptr
}()

type IPPortRecord struct {
	IP [4]uint32
	TS time.Time
}

// Assist httpflow to judge server ip.
type srcIPPortRecorder struct {
	sync.RWMutex
	Record map[[4]uint32]IPPortRecord
}

func (record *srcIPPortRecorder) InsertAndUpdate(ip [4]uint32) {
	record.Lock()
	defer record.Unlock()
	record.Record[ip] = IPPortRecord{
		IP: ip,
		TS: ntp.Now(),
	}
}

func (record *srcIPPortRecorder) Query(ip [4]uint32) (*IPPortRecord, error) {
	record.RLock()
	defer record.RUnlock()
	if v, ok := record.Record[ip]; ok {
		return &v, nil
	} else {
		return nil, fmt.Errorf("not found")
	}
}

const (
	cleanTickerIPPortDur = time.Minute * 3
	cleanIPPortDur       = time.Minute * 5
)

func (record *srcIPPortRecorder) CleanOutdateData() {
	record.Lock()
	defer record.Unlock()
	ts := ntp.Now()
	needDelete := [][4]uint32{}
	for k, v := range record.Record {
		if ts.Sub(v.TS) > cleanIPPortDur {
			needDelete = append(needDelete, k)
		}
	}
	for _, v := range needDelete {
		delete(record.Record, v)
	}
}

func (record *srcIPPortRecorder) AutoClean() {
	ticker := time.NewTicker(cleanTickerIPPortDur)
	for {
		<-ticker.C
		record.CleanOutdateData()
	}
}

func resolveSockfdLookupSymbol() string {
	data, err := os.ReadFile("/proc/kallsyms")
	if err != nil {
		return "sockfd_lookup_light"
	}

	switch {
	case bytes.Contains(data, []byte(" sockfd_lookup_light\n")):
		return "sockfd_lookup_light"
	case bytes.Contains(data, []byte(" sockfd_lookup\n")):
		return "sockfd_lookup"
	default:
		return "sockfd_lookup_light"
	}
}

func disabledNetflowPrograms(kernelVersion uint64, useLegacyConsts bool, ipv6Disabled bool) []string {
	disabled := make([]string, 0, 10)
	appendDisabled := func(programs ...string) {
		for _, program := range programs {
			exists := false
			for _, disabledProgram := range disabled {
				if disabledProgram == program {
					exists = true
					break
				}
			}
			if !exists {
				disabled = append(disabled, program)
			}
		}
	}
	if ipv6Disabled || (useLegacyConsts && kernelVersion < minIPv6NetflowKernel) {
		appendDisabled("kprobe__ip6_make_skb")
	}
	if !enableUDP {
		appendDisabled(
			"kprobe__ip_make_skb",
			"kprobe__ip6_make_skb",
			"kprobe__udp_recvmsg",
			"kretprobe__udp_recvmsg",
			"kprobe__inet_bind",
			"kretprobe__inet_bind",
			"kprobe__inet6_bind",
			"kretprobe__inet6_bind",
			"kprobe__udp_destroy_sock",
		)
	}
	return disabled
}

func netflowMapMaxEntries() uint32 {
	raw := strings.TrimSpace(os.Getenv(netflowMapMaxEntriesEnv))
	if raw == "" {
		return defaultNetflowMapMaxEntries
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		l.Warnf("invalid %s=%q, use default %d", netflowMapMaxEntriesEnv, raw, defaultNetflowMapMaxEntries)
		return defaultNetflowMapMaxEntries
	}
	if n < minNetflowMapMaxEntries {
		return minNetflowMapMaxEntries
	}
	if n > maxNetflowMapMaxEntries {
		return maxNetflowMapMaxEntries
	}
	return uint32(n)
}

func newNetflowRuntimeWithDisabledPrograms(
	patches []bpfutil.ConstantPatch,
	ctMap map[string]*ebpf.Map,
	closedEventHandler bpfutil.PerfHandler,
	ipv6Disabled bool,
	extraDisabledPrograms []string,
) (*bpfutil.Runtime, error) {
	useLegacyConsts, kernelVersion, err := bpfutil.UseLegacyConstObjects()
	if err != nil {
		l.Warnf("detect kernel version for legacy eBPF constants failed: %v", err)
	}
	sockfdLookupSymbol := resolveSockfdLookupSymbol()
	disabledPrograms := append(disabledNetflowPrograms(kernelVersion, useLegacyConsts, ipv6Disabled),
		extraDisabledPrograms...)

	// Some kretprobe type programs need to set maxactive， https://www.kernel.org/doc/Documentation/kprobes.txt.
	probes := filterDisabledNetflowProbes([]*bpfutil.HookSpec{
		{
			ID: bpfutil.HookID{
				Program: "kprobe__sockfd_lookup_light",
			},
			KernelSymbol:    sockfdLookupSymbol,
			KProbeMaxActive: 128,
		}, {
			ID: bpfutil.HookID{
				Program: "kretprobe__sockfd_lookup_light",
			},
			KernelSymbol:    sockfdLookupSymbol,
			KProbeMaxActive: 128,
		}, {
			ID: bpfutil.HookID{
				Program: "kprobe__do_sendfile",
			},
		}, {
			ID: bpfutil.HookID{
				Program: "kretprobe__do_sendfile",
			},
			KProbeMaxActive: 128,
		}, {
			ID: bpfutil.HookID{
				Program: "kprobe__tcp_set_state",
			},
		}, {
			ID: bpfutil.HookID{
				Program: "kretprobe__inet_csk_accept",
			},
			KProbeMaxActive: 128,
		}, {
			ID: bpfutil.HookID{
				Program: "kprobe__inet_csk_listen_stop",
			},
		}, {
			ID: bpfutil.HookID{
				Program: "kprobe__tcp_close",
			},
		}, {
			ID: bpfutil.HookID{
				Program: "kprobe__tcp_retransmit_skb",
			},
		}, {
			ID: bpfutil.HookID{
				Program: "kprobe__tcp_sendmsg",
			},
		}, {
			ID: bpfutil.HookID{
				Program: "kprobe__tcp_cleanup_buf",
			},
			KernelSymbol: "tcp_cleanup_rbuf",
		}, {
			ID: bpfutil.HookID{
				Program: "kprobe__ip_make_skb",
			},
		}, {
			ID: bpfutil.HookID{
				Program: "kprobe__udp_recvmsg",
			},
		}, {
			ID: bpfutil.HookID{
				Program: "kretprobe__udp_recvmsg",
			},
			KProbeMaxActive: 128,
		}, {
			ID: bpfutil.HookID{
				Program: "kprobe__inet_bind",
			},
		}, {
			ID: bpfutil.HookID{
				Program: "kretprobe__inet_bind",
			},
			KProbeMaxActive: 128,
		}, {
			ID: bpfutil.HookID{
				Program: "kprobe__inet6_bind",
			},
		}, {
			ID: bpfutil.HookID{
				Program: "kretprobe__inet6_bind",
			},
			KProbeMaxActive: 128,
		}, {
			ID: bpfutil.HookID{
				Program: "kprobe__udp_destroy_sock",
			},
		},
	}, disabledPrograms)
	probes, probeFailure := filterUnavailableKernelProbes(probes)
	if probeFailure != nil {
		return nil, fmt.Errorf("%w: %s (%s): %s",
			errCriticalNetflowProbeMissing, probeFailure.Program, probeFailure.Symbol, probeFailure.Reason)
	}

	if len(probes) == 0 {
		return nil, fmt.Errorf("no netflow probes available on current kernel")
	}

	perfRingBufferSize := bpfutil.SmallPerfRingBufferSize()
	runtime := &bpfutil.Runtime{
		Probes: probes,
		Streams: []*bpfutil.PerfStream{
			{
				Map: bpfutil.Map{
					Name: "bpfmap_closed_event",
				},
				PerfStreamOptions: bpfutil.PerfStreamOptions{
					PerfRingBufferSize: perfRingBufferSize,
					Watermark:          bpfutil.PerfWatermark(perfRingBufferSize),
					DataHandler:        closedEventHandler,
					LostHandler: func(cpu int, count uint64, stream *bpfutil.PerfStream, runtime *bpfutil.Runtime) {
						exporter.AddPerfLost("netflow", stream.Name, count)
					},
					ErrorHandler: func(err error, stream *bpfutil.PerfStream, runtime *bpfutil.Runtime) {
						exporter.IncPerfReadError("netflow", stream.Name)
						l.Warnf("netflow perf stream stopped: %v", err)
					},
				},
			},
		},
	}
	mapMaxEntries := netflowMapMaxEntries()
	loadSpec := bpfutil.LoadSpec{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
		Constants:        patches,
		LegacyConstants:  useLegacyConsts,
		DisabledPrograms: disabledPrograms,
		MapMaxEntries: map[string]uint32{
			mapConnStats:       mapMaxEntries,
			mapConnTCPStats:    mapMaxEntries,
			mapConnTCPSegments: mapMaxEntries,
			mapPortBind:        mapMaxEntries,
			mapPortBindProc:    mapMaxEntries,
			mapUDPPortBind:     mapMaxEntries,
		},
	}

	if ctMap != nil {
		loadSpec.MapReplacements = ctMap
	}

	binName := "netflow.o"
	binLoader := dkebpf.NetFlowBin
	if useLegacyConsts {
		binName = "netflow_legacy.o"
		binLoader = dkebpf.NetFlowLegacyBin
		l.Infof("kernel %#x lacks read-only map support, loading legacy netflow object", kernelVersion)
	}
	if len(disabledPrograms) > 0 {
		l.Warnf("disable netflow programs on kernel %#x: %v", kernelVersion, disabledPrograms)
	}

	loadRuntime := func(name string, loader func() ([]byte, error), legacyConstants bool) error {
		loadSpec.LegacyConstants = legacyConstants
		buf, err := loader()
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		return runtime.LoadFromReader((bytes.NewReader(buf)), loadSpec)
	}
	if err := loadRuntime(binName, binLoader, useLegacyConsts); err != nil {
		if useLegacyConsts {
			return nil, err
		}
		l.Warnf("load modern netflow object failed, fallback to legacy object without LRU hash maps: %v", err)
		if err := loadRuntime("netflow_legacy.o", dkebpf.NetFlowLegacyBin, true); err != nil {
			return nil, err
		}
	}

	return runtime, nil
}

func NewNetFlowRuntime(
	patches []bpfutil.ConstantPatch,
	ctMap map[string]*ebpf.Map,
	closedEventHandler bpfutil.PerfHandler,
	ipv6Disabled bool,
) (*bpfutil.Runtime, error) {
	return newNetflowRuntimeWithDisabledPrograms(patches, ctMap, closedEventHandler, ipv6Disabled, nil)
}

func StartNetFlowRuntime(
	patches []bpfutil.ConstantPatch,
	ctMap map[string]*ebpf.Map,
	closedEventHandler bpfutil.PerfHandler,
	ipv6Disabled bool,
) (*bpfutil.Runtime, error) {
	kprobeCaps := bpfutil.DetectKprobeCapability()
	recordNetflowKprobeCapability(kprobeCaps)
	if !kprobeCaps.HasAnyInterface() {
		return nil, fmt.Errorf("%w: missing %s",
			errNetflowKprobeUnavailable, strings.Join(kprobeCaps.MissingPaths(), ", "))
	}

	disabled := map[string]struct{}{}

	for attempt := 0; attempt < 16; attempt++ {
		runtime, err := newNetflowRuntimeWithDisabledPrograms(patches, ctMap, closedEventHandler,
			ipv6Disabled, disabledProgramList(disabled))
		if err != nil {
			program, ok := failedNetflowLoadProgram(err)
			if !ok {
				return nil, err
			}
			if isCriticalNetflowProgram(program) {
				exporter.RecordKernelFunctionStatus("netflow", program, kernelFunctionSymbolForProgram(program), "load_failed", err.Error())
				return nil, fmt.Errorf("%w: %s: %v", errCriticalNetflowProbeLoad, program, err)
			}
			if _, exists := disabled[program]; exists {
				return nil, err
			}
			disabled[program] = struct{}{}
			exporter.RecordKernelFunctionStatus("netflow", program, kernelFunctionSymbolForProgram(program), "load_failed", err.Error())
			l.Warnf("disable netflow program after verifier/load failure: %s", program)
			continue
		}

		if err := runtime.StartRuntime(); err != nil {
			program, ok := failedNetflowAttachProgram(err)
			_ = runtime.Shutdown()
			if !ok {
				return nil, err
			}
			if isCriticalNetflowProgram(program) {
				exporter.RecordKernelFunctionStatus("netflow", program, kernelFunctionSymbolForProgram(program), "attach_failed", err.Error())
				return nil, fmt.Errorf("%w: %s: %v", errCriticalNetflowProbeAttach, program, err)
			}
			if _, exists := disabled[program]; exists {
				return nil, err
			}
			disabled[program] = struct{}{}
			exporter.RecordKernelFunctionStatus("netflow", program, kernelFunctionSymbolForProgram(program), "attach_failed", err.Error())
			l.Warnf("disable netflow probe after attach failure: %s", program)
			continue
		}

		return runtime, nil
	}

	return nil, fmt.Errorf("netflow runtime exhausted retry budget")
}

func recordNetflowKprobeCapability(caps bpfutil.KprobeCapability) {
	recordNetflowKprobePathStatus("pmu", "/sys/bus/event_source/devices/kprobe/type", caps.PMUTypePath)
	recordNetflowKprobePathStatus("tracefs", "/sys/kernel/tracing/kprobe_events", caps.TraceFSKprobeEvents)
	recordNetflowKprobePathStatus("debugfs", "/sys/kernel/debug/tracing/kprobe_events", caps.DebugFSKprobeEvents)
}

func recordNetflowKprobePathStatus(program, wantPath, foundPath string) {
	if foundPath != "" {
		exporter.RecordKernelFunctionStatus("netflow", program, wantPath, "available", "")
		return
	}
	exporter.RecordKernelFunctionStatus("netflow", program, wantPath, "unknown", "kprobe interface path not found")
}

func kernelFunctionSymbolForProgram(program string) string {
	symbol, _ := bpfutil.KernelProbeSymbol(bpfutil.HookSpec{
		ID: bpfutil.HookID{
			Program: program,
		},
	})
	return symbol
}

func disabledProgramList(disabled map[string]struct{}) []string {
	if len(disabled) == 0 {
		return nil
	}

	list := make([]string, 0, len(disabled))
	for name := range disabled {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

func failedNetflowLoadProgram(err error) (string, bool) {
	const marker = "load collection: program "

	msg := err.Error()
	idx := strings.Index(msg, marker)
	if idx < 0 {
		return "", false
	}
	rest := msg[idx+len(marker):]
	end := strings.Index(rest, ":")
	if end <= 0 {
		return "", false
	}
	return rest[:end], true
}

func failedNetflowAttachProgram(err error) (string, bool) {
	const marker = `attach probe "`

	msg := err.Error()
	idx := strings.Index(msg, marker)
	if idx < 0 {
		return "", false
	}
	rest := msg[idx+len(marker):]
	end := strings.Index(rest, `"`)
	if end <= 0 {
		return "", false
	}
	return rest[:end], true
}

func AddClientServerInf(mtags map[string]string, mfields map[string]any) (map[string]string, map[string]any) {
	direction, ok := mtags["direction"]
	if !ok {
		return mtags, mfields
	}

	if v, ok := mtags["dst_domain"]; ok {
		mtags["server_domain"] = v
	}

	switch direction {
	case DirectionIncoming:
		mtags["conn_side"] = "server"

		mtags["server_ip"] = mtags["src_ip"]
		mtags["server_ip_type"] = mtags["src_ip_type"]
		mtags["server_port"] = mtags["src_port"]

		mtags["client_ip"] = mtags["dst_ip"]
		mtags["client_ip_type"] = mtags["dst_ip_type"]
		mtags["client_port"] = mtags["dst_port"]

		if v, ok := mfields["bytes_read"]; ok {
			mfields["client_sent"] = v
		}
		if v, ok := mfields["bytes_written"]; ok {
			mfields["server_sent"] = v
		}
	default:
		mtags["conn_side"] = "client"

		mtags["server_ip"] = mtags["dst_ip"]
		mtags["server_ip_type"] = mtags["dst_ip_type"]
		mtags["server_port"] = mtags["dst_port"]

		mtags["client_ip"] = mtags["src_ip"]
		mtags["client_ip_type"] = mtags["src_ip_type"]
		mtags["client_port"] = mtags["src_port"]

		if v, ok := mfields["bytes_written"]; ok {
			mfields["client_sent"] = v
		}
		if v, ok := mfields["bytes_read"]; ok {
			mfields["server_sent"] = v
		}
	}

	return mtags, mfields
}

func IsIncomingFromK8s(k8sNetInfo *cli.K8sInfo, pid int, srcIP string,
	srcPort uint32, transport string,
) bool {
	if k8sNetInfo != nil {
		if t, ok := k8sNetInfo.IsServer(pid,
			transport, srcIP, int(srcPort)); ok {
			return t
		}
	}
	return false
}

func addNoValueK8s(src bool, mTags map[string]string) map[string]string {
	if mTags == nil {
		mTags = map[string]string{}
	}

	client := true
	if v, ok := mTags["direction"]; ok {
		if v == DirectionIncoming {
			client = false
		}
	}

	if src {
		mTags["src_k8s_namespace"] = NoValue
		mTags["src_k8s_pod_name"] = NoValue
		mTags["src_k8s_service_name"] = NoValue
		mTags["src_k8s_deployment_name"] = NoValue
		mTags["src_k8s_workload_name"] = NoValue
		mTags["src_k8s_workload_type"] = NoValue
	} else {
		mTags["dst_k8s_namespace"] = NoValue
		mTags["dst_k8s_pod_name"] = NoValue
		mTags["dst_k8s_service_name"] = NoValue
		mTags["dst_k8s_deployment_name"] = NoValue
		mTags["dst_k8s_workload_name"] = NoValue
		mTags["dst_k8s_workload_type"] = NoValue
	}

	if (client && src) || (!client && !src) {
		mTags["client_k8s_namespace"] = NoValue
		mTags["client_k8s_pod_name"] = NoValue
		mTags["client_k8s_service_name"] = NoValue
		mTags["client_k8s_deployment_name"] = NoValue
		mTags["client_k8s_workload_name"] = NoValue
		mTags["client_k8s_workload_type"] = NoValue
	} else {
		mTags["server_k8s_namespace"] = NoValue
		mTags["server_k8s_pod_name"] = NoValue
		mTags["server_k8s_service_name"] = NoValue
		mTags["server_k8s_deployment_name"] = NoValue
		mTags["server_k8s_workload_name"] = NoValue
		mTags["server_k8s_workload_type"] = NoValue
	}

	return mTags
}

func addPodTag(src bool, k8stag *cli.K8sTag, mTags map[string]string) map[string]string {
	if mTags == nil {
		mTags = map[string]string{}
	}
	client := true
	if v, ok := mTags["direction"]; ok {
		if v == DirectionIncoming {
			client = false
		}
	}

	if src {
		mTags["src_k8s_namespace"] = k8stag.NS
		mTags["src_k8s_pod_name"] = k8stag.PodName
		mTags["src_k8s_service_name"] = k8stag.SvcName
		mTags["src_k8s_deployment_name"] = k8stag.WorkloadName
		mTags["src_k8s_workload_name"] = k8stag.WorkloadName
		mTags["src_k8s_workload_type"] = k8stag.Kind.String()
		for k, v := range k8stag.Labels {
			mTags[k] = v
		}
	} else {
		mTags["dst_k8s_namespace"] = k8stag.NS
		mTags["dst_k8s_pod_name"] = k8stag.PodName
		mTags["dst_k8s_service_name"] = k8stag.SvcName
		mTags["dst_k8s_deployment_name"] = k8stag.WorkloadName
		mTags["dst_k8s_workload_name"] = k8stag.WorkloadName
		mTags["dst_k8s_workload_type"] = k8stag.Kind.String()
	}

	if (client && src) || (!client && !src) {
		mTags["client_k8s_namespace"] = k8stag.NS
		mTags["client_k8s_pod_name"] = k8stag.PodName
		mTags["client_k8s_service_name"] = k8stag.SvcName
		mTags["client_k8s_deployment_name"] = k8stag.WorkloadName
		mTags["client_k8s_workload_name"] = k8stag.WorkloadName
		mTags["client_k8s_workload_type"] = k8stag.Kind.String()
	} else {
		mTags["server_k8s_namespace"] = k8stag.NS
		mTags["server_k8s_pod_name"] = k8stag.PodName
		mTags["server_k8s_service_name"] = k8stag.SvcName
		mTags["server_k8s_deployment_name"] = k8stag.WorkloadName
		mTags["server_k8s_workload_name"] = k8stag.WorkloadName
		mTags["server_k8s_workload_type"] = k8stag.Kind.String()
	}
	return mTags
}

func addSvcTag(src bool, t *cli.PodChainSvc, mTags map[string]string) map[string]string {
	if mTags == nil {
		mTags = map[string]string{}
	}

	client := true
	if v, ok := mTags["direction"]; ok {
		if v == DirectionIncoming {
			client = false
		}
	}

	if src {
		mTags["src_k8s_namespace"] = t.Chain.Tag.NS
		mTags["src_k8s_pod_name"] = NoValue
		mTags["src_k8s_service_name"] = t.Svc.Name
		mTags["src_k8s_deployment_name"] = t.Chain.Tag.WorkloadName
		mTags["src_k8s_workload"] = t.Chain.Tag.WorkloadName
		mTags["src_k8s_workload_type"] = t.Chain.Tag.Kind.String()
	} else {
		mTags["dst_k8s_namespace"] = t.Chain.Tag.NS
		mTags["dst_k8s_pod_name"] = NoValue
		mTags["dst_k8s_service_name"] = t.Svc.Name
		mTags["dst_k8s_deployment_name"] = t.Chain.Tag.WorkloadName
		mTags["dst_k8s_workload"] = t.Chain.Tag.WorkloadName
		mTags["dst_k8s_workload_type"] = t.Chain.Tag.Kind.String()
	}

	if (client && src) || (!client && !src) {
		mTags["client_k8s_namespace"] = t.Chain.Tag.NS
		mTags["client_k8s_pod_name"] = NoValue
		mTags["client_k8s_service_name"] = t.Svc.Name
		mTags["client_k8s_deployment_name"] = t.Chain.Tag.WorkloadName
		mTags["client_k8s_workload"] = t.Chain.Tag.WorkloadName
		mTags["client_k8s_workload_type"] = t.Chain.Tag.Kind.String()
	} else {
		mTags["server_k8s_namespace"] = t.Chain.Tag.NS
		mTags["server_k8s_pod_name"] = NoValue
		mTags["server_k8s_service_name"] = t.Svc.Name
		mTags["server_k8s_deployment_name"] = t.Chain.Tag.WorkloadName
		mTags["server_k8s_workload"] = t.Chain.Tag.WorkloadName
		mTags["server_k8s_workload_type"] = t.Chain.Tag.Kind.String()
	}

	return mTags
}

func AddK8sTags2Map(k8sNetInfo *cli.K8sInfo,
	basekey *BaseKey, mTags map[string]string,
) map[string]string {
	if mTags == nil {
		mTags = map[string]string{}
	}

	if basekey == nil {
		return mTags
	}

	if k8sNetInfo != nil {
		srcK8sFlag := false
		dstK8sFlag := false
		if t, ok := k8sNetInfo.IsServer(basekey.PID, basekey.Transport,
			basekey.SAddr, int(basekey.SPort)); ok && t {
			mTags["direction"] = DirectionIncoming
		}
		if t, ok := k8sNetInfo.QueryPodInfo(
			basekey.PID, basekey.SAddr, int(basekey.SPort), basekey.Transport); ok && t != nil {
			srcK8sFlag = true
			mTags = addPodTag(true, t, mTags)
		} else if t, ok := k8sNetInfo.QuerySvcInfo(
			basekey.Transport, basekey.SAddr, int(basekey.SPort)); ok && t != nil {
			srcK8sFlag = true
			mTags = addSvcTag(true, t, mTags)
		}

		if basekey.DNATAddr != "" && basekey.DNATPort != 0 {
			if t, ok := k8sNetInfo.QueryPodInfo(
				0, basekey.DNATAddr, int(basekey.DNATPort), basekey.Transport); ok && t != nil {
				dstK8sFlag = true
				mTags = addPodTag(false, t, mTags)
				goto skip_dst
			}
		}

		if t, ok := k8sNetInfo.QueryPodInfo(0,
			basekey.DAddr, int(basekey.DPort), basekey.Transport); ok && t != nil {
			// k.dport
			dstK8sFlag = true
			addPodTag(false, t, mTags)
		} else {
			if t, ok := k8sNetInfo.QuerySvcInfo(basekey.Transport,
				basekey.DAddr, int(basekey.DPort)); ok && t != nil {
				dstK8sFlag = true
				mTags = addSvcTag(false, t, mTags)
			}
		}

	skip_dst:
		if srcK8sFlag || dstK8sFlag {
			mTags["sub_source"] = "K8s"
			if !srcK8sFlag {
				addNoValueK8s(true, mTags)
			}
			if !dstK8sFlag {
				addNoValueK8s(false, mTags)
			}
		}
	}
	return mTags
}

func U32BEToIPv4Array(addr uint32) [4]byte {
	ip := [4]byte{}
	binary.LittleEndian.PutUint32(ip[:], addr)
	return ip
}

func U32IP4(addr uint32) net.IP {
	ip4 := U32BEToIPv4Array(addr)
	return net.IP(ip4[:])
}

func U32BEToIPv6Array(addr [4]uint32) [16]byte {
	var ip [16]byte
	for x := 0; x < 4; x++ {
		binary.LittleEndian.PutUint32(ip[x*4:], addr[x])
	}
	return ip
}

func U32IP6(addr [4]uint32) net.IP {
	ip6 := U32BEToIPv6Array(addr)
	return net.IP(ip6[:])
}

func U32BEToIP(addr [4]uint32, isIPv6 bool) net.IP {
	if isIPv6 {
		return U32IP6(addr)
	} else {
		return U32IP4(addr[3])
	}
}

// ConnNotNeedToFilter rules: 1. Filter connections with the same source IP and destination IP;
// 2. Filter the connection of loopback ip;
// 3. Filter connections without data sending and receiving within a collection cycle;
// 4. Filter connections with port 0 or ip address :: or 0.0.0.0;
// Need to filter, the function returns False.
func ConnNotNeedToFilter(conn *ConnectionInfo, connStats *ConnFullStats) bool {
	if !enableUDP && !ConnProtocolIsTCP(conn.Meta) {
		return false
	}
	if (conn.Saddr[0]|conn.Saddr[1]|conn.Saddr[2]|conn.Saddr[3]) == 0 ||
		(conn.Daddr[0]|conn.Daddr[1]|conn.Daddr[2]|conn.Daddr[3]) == 0 ||
		conn.Sport == 0 || conn.Dport == 0 {
		return false
	}
	if ConnAddrIsIPv4(conn.Meta) { // IPv4
		if (conn.Saddr[3]&0xff) == 127 && (conn.Daddr[3]&0xff) == 127 {
			return false
		}
	} else { // IPv6
		if conn.Saddr[2] == 0xffff0000 && conn.Daddr[2] == 0xffff0000 {
			if (conn.Saddr[3]&0xff) == 127 && (conn.Daddr[3]&0xff) == 127 {
				return false
			}
		} else if (conn.Saddr[0]|conn.Saddr[1]|conn.Saddr[2]) == 0 && conn.Saddr[3] == 1 &&
			(conn.Daddr[0]|conn.Daddr[1]|conn.Daddr[2]) == 0 && conn.Daddr[3] == 1 {
			return false
		}
	}

	// Filter connections that have not changed in the previous cycle
	if connStats.Stats.RecvBytes == 0 && connStats.Stats.SentBytes == 0 &&
		connStats.TotalClosed == 0 && connStats.TotalEstablished == 0 {
		return false
	}

	return true
}

func ConnCmpNoSPort(expected, actual ConnectionInfo) bool {
	expected.Sport = 0
	actual.Sport = 0
	return expected == actual
}

func ConnCmpNoPid(expected, actual ConnectionInfo) bool {
	expected.Pid = 0
	actual.Pid = 0
	return expected == actual
}

const (
	EphemeralPortMin = 32768
	EphemeralPortMax = 60999
)

func IsEphemeralPort(port uint32) bool {
	return port >= uint32(ephemeralPortMin)
}

func NormalizeDirectionByPorts(direction string, sport, dport uint32) string {
	srcEphemeral := IsEphemeralPort(sport)
	dstEphemeral := IsEphemeralPort(dport)

	switch {
	case srcEphemeral && !dstEphemeral:
		return DirectionOutgoing
	case !srcEphemeral && dstEphemeral:
		return DirectionIncoming
	default:
		return direction
	}
}

func NormalizeDirectionAndPorts(direction string, sport, dport uint32) (string, uint32, uint32) {
	direction = NormalizeDirectionByPorts(direction, sport, dport)

	switch direction {
	case DirectionOutgoing:
		if IsEphemeralPort(sport) {
			sport = math.MaxUint32
		}
	case DirectionIncoming:
		if IsEphemeralPort(dport) {
			dport = math.MaxUint32
		}
	}

	return direction, sport, dport
}

func IPPortFilterIn(conn *ConnectionInfo) bool {
	if conn.Sport == 0 || conn.Dport == 0 {
		return false
	}

	if ConnAddrIsIPv4(conn.Meta) {
		if (conn.Saddr[3]&0xFF == 0x7F) || (conn.Daddr[3]&0xFF == 0x7F) {
			return false
		}
	} else if (conn.Saddr[0]|conn.Saddr[1]) == 0x00 || (conn.Daddr[0]|conn.Daddr[1]) == 0x00 {
		if (conn.Saddr[2] == 0xffff0000 && conn.Saddr[3]&0xFF == 0x7F) ||
			(conn.Daddr[2] == 0xffff0000 && conn.Daddr[3]&0xFF == 0x7F) {
			return false
		} else if (conn.Saddr[2] == 0x0 && conn.Saddr[3] == 0x01000000) ||
			(conn.Daddr[2] == 0x0 && conn.Daddr[3] == 0x01000000) {
			return false
		}
	}
	return true
}
