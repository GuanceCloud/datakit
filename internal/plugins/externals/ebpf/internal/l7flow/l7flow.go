//go:build linux
// +build linux

// Package l7flow collects http(s) request flow
package l7flow

//go:generate go run ../c/genlayout -target l7flow

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cilium/ebpf"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/c"
	dkct "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/conntrack"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/protodec"
	dknetflow "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/procwatch"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/cli"
	"golang.org/x/sys/unix"
)

const HTTPPayloadMaxsize = 157

const (
	apiflowPerfBufferPagesEnv  = "DK_EBPF_APIFLOW_PERF_PAGES"
	apiflowMinCaptureSizeEnv   = "DK_EBPF_APIFLOW_MIN_CAPTURE_SIZE"
	apiflowPerfLostLogInterval = 10 * time.Second

	apiflowDefaultPerfBufferPages    = 1024
	apiflowDefaultPerfBufferMaxBytes = 128 * 1024 * 1024
	apiflowMinPerfBufferPages        = 64

	apiflowArgMapMaxEntriesEnv            = "DK_EBPF_L7FLOW_ARG_MAP_MAX_ENTRIES"
	apiflowSKMapMaxEntriesEnv             = "DK_EBPF_L7FLOW_SK_MAP_MAX_ENTRIES"
	apiflowProtocolFilterMapMaxEntriesEnv = "DK_EBPF_L7FLOW_PROTOCOL_FILTER_MAP_MAX_ENTRIES"
	defaultApiflowArgMapMaxEntries        = 2048
	defaultApiflowSKMapMaxEntries         = 40960
	defaultApiflowProtocolFilterEntries   = 65536
	minApiflowArgMapMaxEntries            = 128
	minApiflowStateMapMaxEntries          = 1024
	maxApiflowMapMaxEntries               = 1048576
)

const (
	mapAPIFlowSyscallRWArg       = "mp_syscall_rw_arg"
	mapAPIFlowSyscallRWVArg      = "mp_syscall_rw_v_arg"
	mapAPIFlowSKInfo             = "mp_sk_inf"
	mapAPIFlowSSLReadArgs        = "bpfmap_ssl_read_args"
	mapAPIFlowBIONewSocketArgs   = "bpfmap_bio_new_socket_args"
	mapAPIFlowSSLContextSockFD   = "bpfmap_ssl_ctx_sockfd"
	mapAPIFlowSSLBIOFD           = "bpfmap_ssl_bio_fd"
	mapAPIFlowSSLPIDTgidContext  = "bpfmap_ssl_pidtgid_ctx"
	mapAPIFlowSyscallSendfileArg = "bpfmap_syscall_sendfile_arg"
	mapAPIFlowProtocolFilter     = "mp_protocol_filter"
)

// const srcNameM = "httpflow"

const (
	NoValue           = "N/A"
	DirectionOutgoing = "outgoing"
	DirectionIncoming = "incoming"
)

const (
	ConnL3Mask uint32 = dknetflow.ConnL3Mask
	ConnL3IPv4 uint32 = dknetflow.ConnL3IPv4
	ConnL3IPv6 uint32 = dknetflow.ConnL3IPv6

	ConnL4Mask uint32 = dknetflow.ConnL4Mask
	ConnL4TCP  uint32 = dknetflow.ConnL4TCP
	ConnL4UDP  uint32 = dknetflow.ConnL4UDP

	L7BufferShift     = 12
	PayloadBufSize    = 1 << L7BufferShift
	KernelTaskCommLen = 16

	inputHTTPFlow = "ebpf-net/httpflow"
	inputTracing  = "ebpf-net/bpftracing"
)

var (
	// libssl.
	RegexpLibSSL    = regexp.MustCompile(`libssl.so`)
	RegexpLibCrypto = regexp.MustCompile(`libcrypto.so`)

	// TODO: guntls.
)

type (
	HTTPStats struct {
		Direction string

		ReqMethod uint8

		Path     string
		RespCode uint32

		HTTPVersion uint32

		// PidTid uint64

		Recv int
		Send int

		ReqSeq  int64
		RespSeq int64

		ReqTS  uint64
		RespTS uint64
	}

	HTTPReqFinishedInfo struct {
		ConnInfo  comm.ConnectionInfo
		HTTPStats HTTPStats
	}
)

func readMeta(buf *CNetEventComm, dst *comm.ConnectionInfo) {
	conn := buf.meta.sk_inf.conn

	// 暂时屏蔽 uds，其 ip port 为 0； ebpf 暂时不采集此类 socket
	dst.Saddr = (*(*[4]uint32)(unsafe.Pointer(&conn.saddr))) //nolint:gosec
	dst.Daddr = (*(*[4]uint32)(unsafe.Pointer(&conn.daddr))) //nolint:gosec
	dst.Sport = uint32(conn.sport)
	dst.Dport = uint32(conn.dport)
	dst.Pid = conn.pid
	dst.Netns = conn.netns
	dst.Meta = conn.meta
	dst.TaskName = taskCommString((*[KernelTaskCommLen]byte)(unsafe.Pointer(&buf.meta.comm))) //nolint:gosec
	if dst.NATDport == 0 && (dst.NATDaddr[0]|dst.NATDaddr[1]|dst.NATDaddr[2]|dst.NATDaddr[3]) == 0 {
		if natAddr, natPort, ok := dkct.LookupDNATTuple(dst.Saddr, dst.Daddr, dst.Sport, dst.Dport, dst.Netns); ok {
			dst.NATDaddr = natAddr
			dst.NATDport = natPort
		}
	}
}

func taskCommString(comm *[KernelTaskCommLen]byte) string {
	if comm == nil {
		return ""
	}

	start := 0
	for start < len(comm) {
		switch comm[start] {
		case 0, ' ', '\t', '\n', '\r', '\v', '\f':
			start++
		default:
			goto trimRight
		}
	}

	return ""

trimRight:
	end := len(comm)
	for end > start {
		switch comm[end-1] {
		case 0, ' ', '\t', '\n', '\r', '\v', '\f':
			end--
		default:
			return string(comm[start:end])
		}
	}

	return ""
}

var log = logger.DefaultSLogger("ebpf")

func Init(nl *logger.Logger) {
	log = nl
	comm.Init(nl)
	protodec.Init()
}

var (
	libSSLSection = []string{
		"uprobe__SSL_read",
		"uretprobe__SSL_read",
		"uprobe__SSL_write",
		"uprobe__SSL_shutdown",
		"uprobe__SSL_set_fd",
		"uprobe__SSL_set_bio",
	}
	libcryptoSection = []string{
		"uprobe__BIO_new_socket",
		"uretprobe__BIO_new_socket",
	}
)

type perferEventHandle = bpfutil.PerfHandler

type perfLostWarningLimiter struct {
	mu         sync.Mutex
	lastLogAt  time.Time
	suppressed uint64
	now        func() time.Time
}

func newPerfLostWarningLimiter(now func() time.Time) *perfLostWarningLimiter {
	if now == nil {
		now = time.Now
	}
	return &perfLostWarningLimiter{now: now}
}

func (l *perfLostWarningLimiter) format(cpu int, count uint64) string {
	if l == nil || count == 0 {
		return ""
	}

	now := l.now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.lastLogAt.IsZero() || now.Sub(l.lastLogAt) >= apiflowPerfLostLogInterval {
		suppressed := l.suppressed
		l.suppressed = 0
		l.lastLogAt = now
		if suppressed == 0 {
			return fmt.Sprintf("lost %d events on cpu %d", count, cpu)
		}
		return fmt.Sprintf("lost %d events on cpu %d (aggregated %d additional lost events over %s)",
			count, cpu, suppressed, apiflowPerfLostLogInterval)
	}

	l.suppressed += count
	return ""
}

func apiflowPerfRingBufferSize() int {
	return apiflowPerfRingBufferPages(runtime.NumCPU(), os.Getpagesize()) * os.Getpagesize()
}

func apiflowPerfRingBufferPages(numCPU, pageSize int) int {
	pages := defaultApiflowPerfBufferPages(numCPU, pageSize)
	if raw := strings.TrimSpace(os.Getenv(apiflowPerfBufferPagesEnv)); raw != "" {
		switch v, err := strconv.Atoi(raw); {
		case err != nil:
			log.Warnf("invalid %s=%q: %v", apiflowPerfBufferPagesEnv, raw, err)
		case v <= 0:
			log.Warnf("invalid %s=%q: must be > 0", apiflowPerfBufferPagesEnv, raw)
		default:
			pages = v
		}
	}
	if maxPages := maxApiflowPerfBufferPages(numCPU, pageSize); pages > maxPages {
		pages = maxPages
	}
	return pages
}

func defaultApiflowPerfBufferPages(numCPU, pageSize int) int {
	pages := apiflowDefaultPerfBufferPages
	if numCPU > 0 && pageSize > 0 {
		maxPages := maxApiflowPerfBufferPages(numCPU, pageSize)
		if maxPages < pages {
			pages = maxPages
		}
		if pages < apiflowMinPerfBufferPages && maxPages >= apiflowMinPerfBufferPages {
			pages = apiflowMinPerfBufferPages
		}
	}
	if pages < 1 {
		return 1
	}
	if pages < apiflowMinPerfBufferPages &&
		(numCPU <= 0 || pageSize <= 0) {
		return apiflowMinPerfBufferPages
	}
	return pages
}

func maxApiflowPerfBufferPages(numCPU, pageSize int) int {
	if numCPU <= 0 || pageSize <= 0 {
		return apiflowDefaultPerfBufferPages
	}
	pages := apiflowDefaultPerfBufferMaxBytes / pageSize / numCPU
	if pages < 1 {
		return 1
	}
	return pages
}

func apiflowPerfWatermark(bufferSize int) int {
	pageSize := os.Getpagesize()
	if bufferSize <= pageSize || pageSize <= 0 {
		return 0
	}
	return pageSize
}

func apiflowMinCaptureSizePatch() (bpfutil.ConstantPatch, bool) {
	patch := bpfutil.ConstantPatch{
		Name:  "apiflow_min_capture_size",
		Value: uint64(0),
	}

	raw := strings.TrimSpace(os.Getenv(apiflowMinCaptureSizeEnv))
	if raw == "" {
		return patch, true
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		log.Warnf("invalid %s=%q: %v", apiflowMinCaptureSizeEnv, raw, err)
		return patch, true
	}
	if v < 0 {
		log.Warnf("invalid %s=%q: must be >= 0", apiflowMinCaptureSizeEnv, raw)
		return patch, true
	}

	log.Infof("apiflow minimum capture size: %d", v)
	patch.Value = uint64(v)
	return patch, true
}

func apiflowMapMaxEntriesOverride() map[string]uint32 {
	argEntries := apiflowMapMaxEntriesFromEnv(
		apiflowArgMapMaxEntriesEnv,
		defaultApiflowArgMapMaxEntries,
		minApiflowArgMapMaxEntries,
		maxApiflowMapMaxEntries,
	)
	skEntries := apiflowMapMaxEntriesFromEnv(
		apiflowSKMapMaxEntriesEnv,
		defaultApiflowSKMapMaxEntries,
		minApiflowStateMapMaxEntries,
		maxApiflowMapMaxEntries,
	)
	protocolFilterEntries := apiflowMapMaxEntriesFromEnv(
		apiflowProtocolFilterMapMaxEntriesEnv,
		defaultApiflowProtocolFilterEntries,
		minApiflowStateMapMaxEntries,
		maxApiflowMapMaxEntries,
	)

	return map[string]uint32{
		mapAPIFlowSyscallRWArg:       argEntries,
		mapAPIFlowSyscallRWVArg:      argEntries,
		mapAPIFlowSSLReadArgs:        argEntries,
		mapAPIFlowBIONewSocketArgs:   argEntries,
		mapAPIFlowSSLContextSockFD:   argEntries,
		mapAPIFlowSSLBIOFD:           argEntries,
		mapAPIFlowSSLPIDTgidContext:  argEntries,
		mapAPIFlowSyscallSendfileArg: argEntries,
		mapAPIFlowSKInfo:             skEntries,
		mapAPIFlowProtocolFilter:     protocolFilterEntries,
	}
}

func apiflowMapMaxEntriesFromEnv(key string, fallback, min, max uint32) uint32 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || n == 0 {
		log.Warnf("invalid %s=%q, use default %d", key, raw, fallback)
		return fallback
	}
	if n < uint64(min) {
		return min
	}
	if n > uint64(max) {
		return max
	}
	return uint32(n)
}

func pruneLegacyHTTPFlowProbes(probes []*bpfutil.HookSpec) []*bpfutil.HookSpec {
	if len(probes) == 0 {
		return probes
	}

	skip := map[string]struct{}{
		"tracepoint__sys_enter_sendfile64": {},
		"tracepoint__sys_exit_sendfile64":  {},
		"tracepoint__sys_enter_writev":     {},
		"tracepoint__sys_exit_writev":      {},
		"tracepoint__sys_enter_readv":      {},
		"tracepoint__sys_exit_readv":       {},
		"kprobe__tcp_close":                {},
	}

	trimmed := make([]*bpfutil.HookSpec, 0, len(probes))
	for _, probe := range probes {
		if probe == nil {
			continue
		}
		if _, ok := skip[probe.ID.Program]; ok {
			continue
		}
		trimmed = append(trimmed, probe)
	}
	return trimmed
}

func filterUnavailableAPIFlowKernelProbes(probes []*bpfutil.HookSpec) []*bpfutil.HookSpec {
	return filterUnavailableAPIFlowKernelProbesWithLookup(probes, bpfutil.HasKernelSymbol)
}

func filterUnavailableAPIFlowKernelProbesWithLookup(
	probes []*bpfutil.HookSpec,
	hasSymbol func(string) (bool, error),
) []*bpfutil.HookSpec {
	if len(probes) == 0 || hasSymbol == nil {
		return probes
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
		found, err := hasSymbol(symbol)
		if err != nil {
			exporter.RecordKernelFunctionStatus("l7flow", probe.ID.Program, symbol, "unknown", err.Error())
			log.Warnf("detect kernel symbol %q for l7flow probe %q failed: %v",
				symbol, probe.ID.Program, err)
			filtered = append(filtered, probe)
			continue
		}
		if !found {
			exporter.RecordKernelFunctionStatus("l7flow", probe.ID.Program, symbol, "missing", "kernel symbol not found")
			log.Warnf("skip l7flow optional probe %q: kernel symbol %q not found",
				probe.ID.Program, symbol)
			continue
		}
		exporter.RecordKernelFunctionStatus("l7flow", probe.ID.Program, symbol, "available", "")
		filtered = append(filtered, probe)
	}
	return filtered
}

func filterUnavailableAPIFlowTracepoints(probes []*bpfutil.HookSpec) []*bpfutil.HookSpec {
	return filterUnavailableAPIFlowTracepointsWithLookup(probes, tracepointEventExists)
}

func filterUnavailableAPIFlowTracepointsWithLookup(
	probes []*bpfutil.HookSpec,
	hasTracepoint func(group, event string) (bool, error),
) []*bpfutil.HookSpec {
	if len(probes) == 0 || hasTracepoint == nil {
		return probes
	}

	filtered := make([]*bpfutil.HookSpec, 0, len(probes))
	for _, probe := range probes {
		group, event, ok := apiFlowTracepointEvent(probe)
		if !ok {
			filtered = append(filtered, probe)
			continue
		}

		found, err := hasTracepoint(group, event)
		eventName := group + "/" + event
		if err != nil {
			exporter.RecordKernelFunctionStatus("l7flow", probe.ID.Program, eventName, "unknown", err.Error())
			log.Warnf("detect tracepoint %q for l7flow probe %q failed: %v",
				eventName, probe.ID.Program, err)
			filtered = append(filtered, probe)
			continue
		}
		if !found {
			exporter.RecordKernelFunctionStatus("l7flow", probe.ID.Program, eventName, "missing", "tracepoint event not found")
			log.Warnf("skip l7flow optional tracepoint probe %q: event %q not found",
				probe.ID.Program, eventName)
			continue
		}
		exporter.RecordKernelFunctionStatus("l7flow", probe.ID.Program, eventName, "available", "")
		filtered = append(filtered, probe)
	}
	return filtered
}

func apiFlowTracepointEvent(probe *bpfutil.HookSpec) (string, string, bool) {
	if probe == nil {
		return "", "", false
	}
	const tracepointProgramPrefix = "tracepoint__sys_"
	if !strings.HasPrefix(probe.ID.Program, tracepointProgramPrefix) {
		return "", "", false
	}
	return "syscalls", strings.TrimPrefix(probe.ID.Program, "tracepoint__"), true
}

func hasAPIFlowSyscallTracepoint(probes []*bpfutil.HookSpec) bool {
	for _, probe := range probes {
		if _, _, ok := apiFlowTracepointEvent(probe); ok {
			return true
		}
	}
	return false
}

func tracepointEventExists(group, event string) (bool, error) {
	return tracepointEventExistsInRoots([]string{
		"/sys/kernel/tracing/events",
		"/sys/kernel/debug/tracing/events",
	}, group, event)
}

func tracepointEventExistsInRoots(roots []string, group, event string) (bool, error) {
	rootAvailable := false
	for _, root := range roots {
		info, err := os.Stat(root)
		switch {
		case err == nil && info.IsDir():
			rootAvailable = true
		case err == nil:
			continue
		case os.IsNotExist(err):
			continue
		default:
			return false, err
		}

		_, err = os.Stat(filepath.Join(root, group, event))
		switch {
		case err == nil:
			return true, nil
		case os.IsNotExist(err):
			continue
		default:
			return false, err
		}
	}
	if !rootAvailable {
		return false, fmt.Errorf("tracepoint events root not found")
	}
	return false, nil
}

const (
	apiflowSchedGetAffinityProgram = "kprobe__sched_getaffinity"
	apiflowSchedGetAffinityUID     = "kprobe_sched_getaffinity_apiflow"
)

var apiflowKpFlushHookID = bpfutil.HookID{
	Program: apiflowSchedGetAffinityProgram,
	UID:     apiflowSchedGetAffinityUID,
}

func NewHTTPFlowRuntime(patches []bpfutil.ConstantPatch, bmaps map[string]*ebpf.Map,
	bufHandler perferEventHandle, enableTLS bool,
) (*bpfutil.Runtime, *procwatch.LibraryTracker, error) {
	lostWarnLimiter := newPerfLostWarningLimiter(nil)
	if patch, ok := apiflowMinCaptureSizePatch(); ok {
		patches = append(patches, patch)
	}
	useLegacyConsts, kernelVersion, err := bpfutil.UseLegacyConstObjects()
	if err != nil {
		log.Warnf("detect kernel version for legacy eBPF constants failed: %v", err)
	}
	perfRingBufferSize := apiflowPerfRingBufferSize()

	runtime := &bpfutil.Runtime{
		Probes: []*bpfutil.HookSpec{
			{
				ID: bpfutil.HookID{
					Program: "tracepoint__sys_enter_read",
				},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_exit_read"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_enter_write"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_exit_write"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_enter_recvfrom"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_exit_recvfrom"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_enter_sendto"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_exit_sendto"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_enter_writev"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_exit_writev"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_enter_readv"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_exit_readv"},
			},
			{
				ID: bpfutil.HookID{
					Program: "kprobe__tcp_close",
					UID:     "tcp_close_apiflow",
				},
			},
			{
				ID: bpfutil.HookID{
					Program: apiflowSchedGetAffinityProgram,
					UID:     apiflowSchedGetAffinityUID,
				},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_enter_sendfile64"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_exit_sendfile64"},
			},
		},
		Streams: []*bpfutil.PerfStream{
			{
				Map: bpfutil.Map{
					Name: "mp_upload_netwrk_events",
				},
				PerfStreamOptions: bpfutil.PerfStreamOptions{
					PerfRingBufferSize: perfRingBufferSize,
					Watermark:          apiflowPerfWatermark(perfRingBufferSize),
					DataHandler:        bufHandler,
					LostHandler: func(CPU int, count uint64, stream *bpfutil.PerfStream, runtime *bpfutil.Runtime) {
						exporter.AddPerfLost("l7flow", stream.Name, count)
						if msg := lostWarnLimiter.format(CPU, count); msg != "" {
							log.Warn(msg)
						}
					},
					ErrorHandler: func(err error, stream *bpfutil.PerfStream, runtime *bpfutil.Runtime) {
						exporter.IncPerfReadError("l7flow", stream.Name)
						log.Warnf("l7flow perf stream stopped: %v", err)
					},
				},
			},
		},
	}
	loadRuntime := func(legacy bool) error {
		loadSpec := bpfutil.LoadSpec{
			RLimit: &unix.Rlimit{
				Cur: math.MaxUint64,
				Max: math.MaxUint64,
			},
			Constants:       patches,
			LegacyConstants: legacy,
			MapMaxEntries:   apiflowMapMaxEntriesOverride(),
		}
		if bmaps != nil {
			loadSpec.MapReplacements = bmaps
		}

		runtime.Probes = append(runtime.Probes[:0], []*bpfutil.HookSpec{
			{
				ID: bpfutil.HookID{
					Program: "tracepoint__sys_enter_read",
				},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_exit_read"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_enter_write"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_exit_write"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_enter_recvfrom"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_exit_recvfrom"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_enter_sendto"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_exit_sendto"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_enter_writev"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_exit_writev"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_enter_readv"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_exit_readv"},
			},
			{
				ID: bpfutil.HookID{
					Program: "kprobe__tcp_close",
					UID:     "tcp_close_apiflow",
				},
			},
			{
				ID: bpfutil.HookID{
					Program: apiflowSchedGetAffinityProgram,
					UID:     apiflowSchedGetAffinityUID,
				},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_enter_sendfile64"},
			},
			{
				ID: bpfutil.HookID{Program: "tracepoint__sys_exit_sendfile64"},
			},
		}...)

		binName := "apiflow.o"
		binLoader := dkebpf.APIFlowBin
		if legacy {
			log.Warnf("kernel %#x uses degraded legacy apiflow probes: only read/write syscall tracepoints are enabled", kernelVersion)
			runtime.Probes = pruneLegacyHTTPFlowProbes(runtime.Probes)
			binName = "apiflow_legacy.o"
			binLoader = dkebpf.APIFlowLegacyBin
			log.Infof("kernel %#x loading legacy apiflow object", kernelVersion)
		}
		runtime.Probes = filterUnavailableAPIFlowKernelProbes(runtime.Probes)
		runtime.Probes = filterUnavailableAPIFlowTracepoints(runtime.Probes)
		if !hasAPIFlowSyscallTracepoint(runtime.Probes) {
			return fmt.Errorf("no apiflow syscall tracepoints available")
		}

		buf, err := binLoader()
		if err != nil {
			return fmt.Errorf("%s: %w", binName, err)
		}
		return runtime.LoadFromReader(bytes.NewReader(buf), loadSpec)
	}

	if err := loadRuntime(useLegacyConsts); err != nil {
		if useLegacyConsts {
			return nil, nil, err
		}
		log.Warnf("load modern apiflow object failed, fallback to legacy minimal object: %v", err)
		enableTLS = false
		if err := loadRuntime(true); err != nil {
			return nil, nil, err
		}
	}

	var r *procwatch.LibraryTracker
	if enableTLS {
		opensslRules := []procwatch.HookRule{
			{
				Re:     RegexpLibSSL,
				Attach: procwatch.NewAttachFunc(runtime, libSSLSection),
				Detach: procwatch.NewDetachFunc(runtime, libSSLSection),
			},
			{
				Re:     RegexpLibCrypto,
				Attach: procwatch.NewAttachFunc(runtime, libcryptoSection),
				Detach: procwatch.NewDetachFunc(runtime, libcryptoSection),
			},
		}

		var err error
		r, err = procwatch.NewLibraryTracker(opensslRules)
		if err != nil {
			_ = runtime.Shutdown()
			return nil, nil, err
		}
	}

	return runtime, r, nil
}

func hasAPIFlowKpFlushHook(runtime *bpfutil.Runtime) bool {
	if runtime == nil {
		return false
	}
	_, ok := runtime.LookupHook(apiflowKpFlushHookID)
	return ok
}

type APIFlowTracer struct {
	tracer *Tracer
	cancel context.CancelFunc
}

type APITracerOpt func(*apiTracerConfig)

type apiTracerConfig struct {
	tags        map[string]string
	conv2dd     bool
	enableTrace bool
	catalog     *procwatch.Catalog
	protos      map[protodec.L7Protocol]struct{}
	k8sNetInfo  *cli.K8sInfo
	selfPid     int
}

func WithSelfPid(pid int) APITracerOpt {
	return func(cfg *apiTracerConfig) {
		cfg.selfPid = pid
	}
}

func WithTags(tags map[string]string) APITracerOpt {
	return func(cfg *apiTracerConfig) {
		cfg.tags = tags
	}
}

func WithConv2dd(conv2dd bool) APITracerOpt {
	return func(cfg *apiTracerConfig) {
		cfg.conv2dd = conv2dd
	}
}

func WithEnableTrace(enableTrace bool) APITracerOpt {
	return func(cfg *apiTracerConfig) {
		cfg.enableTrace = enableTrace
	}
}

func WithCatalog(catalog *procwatch.Catalog) APITracerOpt {
	return func(cfg *apiTracerConfig) {
		cfg.catalog = catalog
	}
}

func WithProtos(protos map[protodec.L7Protocol]struct{}) APITracerOpt {
	return func(cfg *apiTracerConfig) {
		cfg.protos = protos
	}
}

func WithK8sNetInfo(k8sNetInfo *cli.K8sInfo) APITracerOpt {
	return func(cfg *apiTracerConfig) {
		cfg.k8sNetInfo = k8sNetInfo
	}
}

func NewAPIFlowTracer(ctx context.Context, opts ...APITracerOpt) *APIFlowTracer {
	if ctx == nil {
		ctx = context.Background()
	}

	var cfg apiTracerConfig
	for _, fn := range opts {
		if fn != nil {
			fn(&cfg)
		}
	}

	tracerCtx, cancel := context.WithCancel(ctx)
	return &APIFlowTracer{
		tracer: newTracer(tracerCtx, &cfg),
		cancel: cancel,
	}
}

const bpfMapProtocolFilter = "mp_protocol_filter"

func (tracer *APIFlowTracer) Run(ctx context.Context, patches []bpfutil.ConstantPatch,
	bmaps map[string]*ebpf.Map, enableTLS bool, interval time.Duration,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if tracer == nil || tracer.tracer == nil {
		return fmt.Errorf("api flow tracer is not initialized")
	}

	runtime, r, err := NewHTTPFlowRuntime(patches, bmaps,
		tracer.tracer.PerfEventHandle, enableTLS)
	if err != nil {
		tracer.stop()
		return err
	}

	if err := runtime.StartRuntime(); err != nil {
		log.Error(err)
		_ = runtime.Shutdown()
		tracer.stop()
		return err
	}

	go tracer.tracer.Start(ctx, interval)

	if hasAPIFlowKpFlushHook(runtime) {
		newKpFlushTrigger(ctx)
	} else {
		log.Infof("l7flow kernel flush trigger skipped: %s hook is not loaded",
			apiflowSchedGetAffinityProgram)
	}

	log.Info("api tracer starting ...")

	var fn func(uint64)
	if mp, err := runtime.LookupMap(bpfMapProtocolFilter); err == nil {
		fn = func() func(u uint64) {
			return func(u uint64) {
				val := uint8(1)
				if err := mp.Update(&u, &val, ebpf.UpdateExist); err != nil {
					log.Debug(err)
				}
			}
		}()
	} else {
		fn = func(u uint64) {}
	}

	tracer.tracer.protocolFilter.setFn(fn)

	if r != nil {
		r.Run(ctx, time.Second*30)
	}

	go func() {
		<-ctx.Done()
		_ = runtime.Shutdown()
		tracer.stop()
	}()

	return nil
}

func (tracer *APIFlowTracer) stop() {
	if tracer == nil || tracer.cancel == nil {
		return
	}
	tracer.cancel()
}

func feed(name string, cat point.Category, data []*point.Point) error {
	if len(data) == 0 {
		return nil
	}
	if err := exporter.FeedPoint(name, cat, data); err != nil {
		return err
	}
	return nil
}

func feedEBPFSpan(name string, cat point.Category, data []*point.Point) error {
	if len(data) == 0 {
		return nil
	}
	if err := exporter.FeedEBPFSpan(name, cat, data); err != nil {
		return err
	}
	return nil
}
