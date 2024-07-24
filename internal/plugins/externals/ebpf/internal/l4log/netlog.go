//go:build linux
// +build linux

package l4log

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/google/gopacket"
	"github.com/google/gopacket/afpacket"
	"github.com/google/gopacket/layers"
	cruntime "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
	"golang.org/x/net/bpf"
)

var log = logger.DefaultSLogger("netlog")

func SetLogger(l *logger.Logger) {
	log = l
}

var (
	enableNetlog     = false
	enabledNetMetric = false

	enableL7HTTP = false
)

func ConfigFunc(netlog, netMetric bool, enabledL7Proto []string) {
	log.Info("enable net log: ", netlog)
	log.Info("enable net metric: ", netMetric)

	enableNetlog = netlog
	enabledNetMetric = netMetric

	for _, v := range enabledL7Proto {
		switch strings.ToLower(v) {
		case "http":
			enableL7HTTP = true
			log.Info("enable http protocol")
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
		afpacket.OptNumBlocks(8),
		afpacket.OptAddPktType(true),
	}

	afpktOpt = append(afpktOpt, opts...)

	h, err := afpacket.NewTPacket(afpktOpt...)
	if err != nil {
		return nil, err
	}

	if len(filter) > 0 {
		if err := h.SetBPF(filter); err != nil {
			return nil, err
		}
	}

	return h, nil
}

type netlogCfg struct {
	gTags       map[string]string
	url         string
	aggURL      string
	blacklist   string
	ctrEndpoint []string
}

type CfgFn func(cfg *netlogCfg)

func WithGlobalTags(tags map[string]string) func(cfg *netlogCfg) {
	return func(cfg *netlogCfg) {
		cfg.gTags = tags
	}
}

func WithURL(url string) func(cfg *netlogCfg) {
	return func(cfg *netlogCfg) {
		cfg.url = url
	}
}

func WithAggURL(url string) func(cfg *netlogCfg) {
	return func(cfg *netlogCfg) {
		cfg.aggURL = url
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
	initULID()

	cfg := netlogCfg{}
	for _, fn := range opts {
		if fn != nil {
			fn(&cfg)
		}
	}

	dockerCtr, err := cruntime.NewDockerRuntime("unix:///var/run/docker.sock", "")
	if err != nil {
		log.Warnf("skip connect to docker: %s", err.Error())
	}

	var ctrLi []cruntime.ContainerRuntime

	for _, ep := range cfg.ctrEndpoint {
		if err := checkEndpoint(ep); err != nil {
			log.Warnf("skip connect to %s: %s", ep, err.Error())
			continue
		}
		var r cruntime.ContainerRuntime
		var err error
		if verifyErr := cruntime.VerifyDockerRuntime(ep); verifyErr == nil {
			r, err = cruntime.NewDockerRuntime(ep, "")
		} else {
			r, err = cruntime.NewCRIRuntime(ep, "")
		}
		if err != nil {
			log.Warnf("skip connect to %s: %s", ep, err.Error())
			continue
		} else {
			log.Infof("connect to %s success", ep)
		}
		ctrLi = append(ctrLi, r)
	}

	if dockerCtr == nil && len(ctrLi) == 0 {
		log.Warnf("no container runtime")
	}

	m, err := newNetlogMonitor(cfg.gTags, cfg.url, cfg.aggURL, cfg.blacklist, _fnList)
	if err != nil {
		log.Errorf("create netlog monitor failed: %s", err.Error())
		return
	}

	rCtx, cFn := context.WithCancel(ctx)
	go m.Run(rCtx, ctrLi)
	<-ctx.Done()

	cFn()
}

// checkEndpoint check if endpoint is valid, copy from internal/plugins/inputs/container/impl.go

func checkEndpoint(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint %s, err: %w", endpoint, err)
	}

	switch u.Scheme {
	case "unix":
		// nil
	default:
		return fmt.Errorf("using %s as endpoint is not supported protocol", endpoint)
	}

	info, err := os.Stat(u.Path)
	if os.IsNotExist(err) {
		return fmt.Errorf("endpoint %s does not exist, maybe it is not running", endpoint)
	}
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("endpoint %s cannot be a directory", u.Path)
	}

	return nil
}
