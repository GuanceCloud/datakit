//go:build linux
// +build linux

package l4log

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/google/gopacket"
	"github.com/google/gopacket/afpacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/bpf"
)

var log = logger.DefaultSLogger("netlog")

func SetLogger(l *logger.Logger) {
	log = l
}

var (
	enableNetlog     = false
	enabledNetMetric = false

	enableL7HTTP  = false
	enableL7HTTP2 = false

	fallbackCaptureSocketBlocks = defaultFallbackCaptureSocketBlocks
	sharedCaptureSocketBlocks   = defaultSharedCaptureSocketBlocks
	maxFallbackSocketLimit      = defaultMaxFallbackSocketLimit
)

const (
	defaultFallbackCaptureSocketBlocks = 8
	defaultSharedCaptureSocketBlocks   = 128
	defaultMaxFallbackSocketLimit      = 16
	defaultCapturePollTimeout          = time.Second

	// K8s nodes already pay extra per-pod netns bookkeeping; default to shared-only capture.
	k8sFallbackCaptureSocketBlocks = 4
	k8sSharedCaptureSocketBlocks   = 64
	k8sMaxFallbackSocketLimit      = 0
	netlogMonitorStopTimeout       = 10 * time.Second
)

func ConfigFunc(netlog, netMetric bool, enabledL7Proto []string, l7LogHeaders ...[]string) {
	log.Info("enable net log: ", netlog)
	log.Info("enable net metric: ", netMetric)

	enableNetlog = netlog
	enabledNetMetric = netMetric
	enableL7HTTP = false
	enableL7HTTP2 = false
	if len(l7LogHeaders) > 0 {
		configureL7LogHeaders(l7LogHeaders[0])
	} else {
		configureL7LogHeaders(nil)
	}

	for _, v := range enabledL7Proto {
		switch strings.ToLower(v) {
		case "http":
			enableL7HTTP = true
			enableL7HTTP2 = true
			log.Info("enable http protocol")
		case "http1":
			enableL7HTTP = true
			log.Info("enable http1 protocol")
		case "http2", "grpc":
			enableL7HTTP2 = true
			log.Infof("enable %s protocol", strings.ToLower(v))
		default:
		}
	}
}

type L7Proto uint16

const (
	L7ProtoUnknown L7Proto = iota
	L7ProtoHTTP
	L7ProtoHTTP2
	L7ProtoGRPC

	L7ProtoMySQL
	L7ProtoRedis
)

func (p L7Proto) String() string {
	switch p {
	case L7ProtoHTTP:
		return "http"
	case L7ProtoHTTP2:
		return "http2"
	case L7ProtoGRPC:
		return "grpc"
	case L7ProtoMySQL:
		return "mysql"
	case L7ProtoRedis:
		return "redis"
	case L7ProtoUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

type pktDecoder struct {
	pktDecode   *gopacket.DecodingLayerParser
	vxlanDecode *gopacket.DecodingLayerParser

	eth  *layers.Ethernet
	ipv4 *layers.IPv4
	ipv6 *layers.IPv6
	tcp  *layers.TCP
	udp  *layers.UDP

	// vxlan
	vxlan *layers.VXLAN
}

const (
	// directionUnknown int8 = iota.
	directionTX int8 = iota + 1
	directionRX
)

func isVxlanLayer(port ...uint16) bool {
	for _, p := range port {
		switch p {
		case 8472, 4789:
			return true
		}
	}

	return false
}

func NewPktDecoder() *pktDecoder {
	var eth layers.Ethernet
	var ipv4 layers.IPv4
	var ipv6 layers.IPv6
	var tcp layers.TCP
	var udp layers.UDP

	var vxlan layers.VXLAN

	l := []gopacket.DecodingLayer{
		&eth,
		&ipv4, &ipv6,
		&udp,
		&tcp,
	}

	vxlanLi := []gopacket.DecodingLayer{
		&vxlan,

		&eth,
		&ipv4, &ipv6,
		&udp,
		&tcp,
	}

	return &pktDecoder{
		gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, l...),
		gopacket.NewDecodingLayerParser(layers.LayerTypeVXLAN, vxlanLi...),
		&eth,
		&ipv4, &ipv6,
		&tcp,
		&udp,

		&vxlan,
	}
}

func newRawsocket(filter []bpf.RawInstruction, opts ...any) (*afpacket.TPacket, error) {
	afpktOpt := []any{
		afpacket.OptNumBlocks(fallbackCaptureSocketBlocks),
		afpacket.OptAddPktType(true),
		afpacket.OptPollTimeout(defaultCapturePollTimeout),
	}

	afpktOpt = append(afpktOpt, opts...)

	h, err := afpacket.NewTPacket(afpktOpt...)
	if err != nil {
		return nil, err
	}

	if len(filter) > 0 {
		if err := h.SetBPF(filter); err != nil {
			h.Close()
			return nil, err
		}
	}

	return h, nil
}

type netlogCfg struct {
	gTags           map[string]string
	blacklist       string
	ctrEndpoint     []string
	fallbackSockets int
	fallbackBlocks  int
	sharedBlocks    int
}

type CfgFn func(cfg *netlogCfg)

func WithGlobalTags(tags map[string]string) func(cfg *netlogCfg) {
	return func(cfg *netlogCfg) {
		cfg.gTags = tags
	}
}

func WithBlacklist(blacklist string) func(cfg *netlogCfg) {
	return func(cfg *netlogCfg) {
		cfg.blacklist = blacklist
	}
}

func WithCtrEndpointOverride(endpoint []string) func(cfg *netlogCfg) {
	return func(cfg *netlogCfg) {
		cfg.ctrEndpoint = endpoint
	}
}

func WithCaptureLimits(fallbackSockets, fallbackBlocks, sharedBlocks int) func(cfg *netlogCfg) {
	return func(cfg *netlogCfg) {
		cfg.fallbackSockets = fallbackSockets
		cfg.fallbackBlocks = fallbackBlocks
		cfg.sharedBlocks = sharedBlocks
	}
}

func applyCaptureLimits(cfg *netlogCfg) {
	fallbackCaptureSocketBlocks = defaultFallbackCaptureSocketBlocks
	sharedCaptureSocketBlocks = defaultSharedCaptureSocketBlocks
	maxFallbackSocketLimit = defaultMaxFallbackSocketLimit

	if k8sNetInfo != nil {
		fallbackCaptureSocketBlocks = k8sFallbackCaptureSocketBlocks
		sharedCaptureSocketBlocks = k8sSharedCaptureSocketBlocks
		maxFallbackSocketLimit = k8sMaxFallbackSocketLimit
	}

	if cfg == nil {
		return
	}
	if cfg.fallbackSockets > 0 {
		maxFallbackSocketLimit = cfg.fallbackSockets
	}
	if cfg.fallbackBlocks > 0 {
		fallbackCaptureSocketBlocks = cfg.fallbackBlocks
	}
	if cfg.sharedBlocks > 0 {
		sharedCaptureSocketBlocks = cfg.sharedBlocks
	}
}

func DefaultEndpoint(rootPath string) []string {
	basePath := []string{
		"/var/run/docker.sock",
		"/var/run/containerd/containerd.sock",
		"/var/run/k3s/containerd/containerd.sock",
		"/var/run/crio/crio.sock",
	}
	if rootPath != "" {
		for i := range basePath {
			basePath[i] = filepath.Join(rootPath, basePath[i])
			if v, err := filepath.Abs(basePath[i]); err == nil {
				basePath[i] = v
			}
		}
	}
	for i := range basePath {
		basePath[i] = "unix://" + basePath[i]
	}

	return basePath
}

func NetLog(ctx context.Context, opts ...CfgFn) {
	if ctx == nil {
		ctx = context.Background()
	}
	initULID()

	cfg := netlogCfg{}
	for _, fn := range opts {
		if fn != nil {
			fn(&cfg)
		}
	}

	applyCaptureLimits(&cfg)
	log.Infof("netlog capture limits: fallback_sockets=%d fallback_blocks=%d shared_blocks=%d",
		maxFallbackSocketLimit, fallbackCaptureSocketBlocks, sharedCaptureSocketBlocks)
	if k8sNetInfo != nil && maxFallbackSocketLimit == 0 {
		log.Infof("netlog k8s fallback capture disabled by default; shared capture only")
	}

	ctrLi := newContainerRuntimes(cfg.ctrEndpoint)

	if len(ctrLi) == 0 {
		log.Warnf("no container runtime")
	}

	m, err := newNetlogMonitor(cfg.gTags, cfg.blacklist, _fnList)
	if err != nil {
		log.Errorf("create netlog monitor failed: %s", err.Error())
		closeContainerRuntimes(ctrLi)
		return
	}

	rCtx, cFn := context.WithCancel(ctx)
	monitorDone := make(chan struct{})
	go func() {
		defer close(monitorDone)
		defer closeContainerRuntimes(ctrLi)
		m.Run(rCtx, ctrLi)
	}()
	<-ctx.Done()

	cFn()
	select {
	case <-monitorDone:
	case <-time.After(netlogMonitorStopTimeout):
		log.Warnf("netlog monitor did not stop within %s", netlogMonitorStopTimeout)
	}
}
