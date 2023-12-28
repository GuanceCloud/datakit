//go:build linux
// +build linux

package l4log

import (
	"context"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/kubernetes/pkg/kubelet/cri/remote"
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
)

func ConfigFunc(netlog, netMetric bool) {
	enableNetlog = netlog
	enabledNetMetric = netMetric
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

func NetLog(ctx context.Context, gtags map[string]string, url, aggURL, blacklist string) {
	initULID()

	dockerCtr, err := cruntime.NewDockerRuntime("unix:///var/run/docker.sock", "")
	if err != nil {
		log.Warnf("skip connect to docker: %w", err)
	}

	containerdCtr, err := remote.NewRemoteRuntimeService("unix:///var/run/containerd/containerd.sock", time.Second*5)
	if err != nil {
		log.Warnf("skip connect to containerd: %w", err)
	}

	if dockerCtr == nil && containerdCtr == nil {
		log.Error("no container runtime")
	}

	m, err := newNetlogMonitor(gtags, url, aggURL, blacklist, _fnList)
	if err != nil {
		log.Errorf("create netlog monitor failed: %s", err.Error())
		return
	}
	rCtx, cFn := context.WithCancel(ctx)
	go m.Run(rCtx, containerdCtr, dockerCtr)
	<-ctx.Done()

	cFn()
}
