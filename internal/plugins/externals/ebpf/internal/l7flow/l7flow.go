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
	pages := defaultApiflowPerfBufferPages(runtime.NumCPU(), os.Getpagesize())
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
	return pages * os.Getpagesize()
}

func defaultApiflowPerfBufferPages(numCPU, pageSize int) int {
	pages := apiflowDefaultPerfBufferPages
	if numCPU > 0 && pageSize > 0 {
		capPages := apiflowDefaultPerfBufferMaxBytes / pageSize / numCPU
		if capPages > 0 && capPages < pages {
			pages = capPages
		}
	}
	if pages < apiflowMinPerfBufferPages {
		return apiflowMinPerfBufferPages
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
					Program: "kprobe__sched_getaffinity",
					UID:     "kprobe_sched_getaffinity_apiflow",
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
					Program: "kprobe__sched_getaffinity",
					UID:     "kprobe_sched_getaffinity_apiflow",
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
			return nil, nil, err
		}
	}

	return runtime, r, nil
}

type APIFlowTracer struct {
	tracer *Tracer
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
	var cfg apiTracerConfig
	for _, fn := range opts {
		if fn != nil {
			fn(&cfg)
		}
	}

	return &APIFlowTracer{
		tracer: newTracer(ctx, &cfg),
	}
}

const bpfMapProtocolFilter = "mp_protocol_filter"

func (tracer *APIFlowTracer) Run(ctx context.Context, patches []bpfutil.ConstantPatch,
	bmaps map[string]*ebpf.Map, enableTLS bool, interval time.Duration,
) error {
	go tracer.tracer.Start(ctx, interval)

	runtime, r, err := NewHTTPFlowRuntime(patches, bmaps,
		tracer.tracer.PerfEventHandle, enableTLS)
	if err != nil {
		return err
	}

	if err := runtime.StartRuntime(); err != nil {
		log.Error(err)
		return err
	}

	newKpFlushTrigger(ctx)

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
	}()

	return nil
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
