//go:build (linux && amd64 && ebpf) || (linux && arm64 && ebpf)
// +build linux,amd64,ebpf linux,arm64,ebpf

package httpflow

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"reflect"
	"regexp"
	"time"
	"unsafe"

	"github.com/DataDog/ebpf"
	"github.com/DataDog/ebpf/manager"
	"github.com/google/gopacket/afpacket"
	"github.com/sirupsen/logrus"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/c"
	dknetflow "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/netflow"
	dkout "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/output"
	sysmonitor "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/sysmonitor"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"golang.org/x/sys/unix"
)

// #include "../c/netflow/l7_stats.h"
import "C"

const HTTPPayloadMaxsize = 69

const srcNameM = "httpflow"

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
)

var (
	// libssl
	regexpLibSSL    = regexp.MustCompile(`libssl.so`)
	regexpLibCrypto = regexp.MustCompile(`libcrypto.so`)

	// TODO: guntls
)

type (
	HTTPStatsC           C.struct_http_stats
	HTTPReqFinishedInfoC C.struct_http_req_finished_info

	ConnectionInfoC dknetflow.ConnectionInfoC
	ConnectionInfo  dknetflow.ConnectionInfo
	HTTPStats       struct {
		// payload    [HTTP_PAYLOAD_MAXSIZE]byte
		Payload     string
		ReqMethod   uint8
		HTTPVersion uint32
		RespCode    uint32
		ReqTS       uint64
		RespTS      uint64
	}
	HTTPReqFinishedInfo struct {
		ConnInfo  ConnectionInfo
		HTTPStats HTTPStats
	}
)

var l = logger.DefaultSLogger("ebpf")

func SetLogger(nl *logger.Logger) {
	l = nl
}

var (
	libSSLSection = []string{
		"uprobe/SSL_read",
		"uretprobe/SSL_read",
		"uprobe/SSL_write",
		"uprobe/SSL_shutdown",
		"uprobe/SSL_set_fd",
		"uprobe/SSL_set_bio",
	}
	libcryptoSection = []string{
		"uprobe/BIO_new_socket",
		"uretprobe/BIO_new_socket",
	}
)

func NewHTTPFlowManger(fd int, constEditor []manager.ConstantEditor, bpfMapSockFD *ebpf.Map, closedEventHandler func(cpu int, data []byte,
	perfmap *manager.PerfMap, manager *manager.Manager), enableTLS bool) (*manager.Manager, *sysmonitor.UprobeDynamicLibRegister, error) {
	m := &manager.Manager{
		Probes: []*manager.Probe{
			{
				Section:  "socket/http_filter",
				SocketFD: fd,
			},
			{
				Section: "kretprobe/tcp_sendmsg",
			},
		},
		PerfMaps: []*manager.PerfMap{
			{
				Map: manager.Map{
					Name: "bpfmap_httpreq_fin_event",
				},
				PerfMapOptions: manager.PerfMapOptions{
					// pagesize ~= 4k,
					PerfRingBufferSize: 32 * os.Getpagesize(),
					DataHandler:        closedEventHandler,
				},
			},
		},
	}

	var r *sysmonitor.UprobeDynamicLibRegister
	if enableTLS {
		opensslRules := []sysmonitor.UprobeRegRule{
			{
				Re:         regexpLibSSL,
				Register:   sysmonitor.NewRegisterFunc(m, libSSLSection),
				UnRegister: sysmonitor.NewUnRegisterFunc(m, libSSLSection),
			},
			{
				Re:         regexpLibCrypto,
				Register:   sysmonitor.NewRegisterFunc(m, libcryptoSection),
				UnRegister: sysmonitor.NewUnRegisterFunc(m, libcryptoSection),
			},
		}

		var err error
		r, err = sysmonitor.NewUprobeDyncLibRegister(opensslRules)
		if err != nil {
			return nil, nil, err
		}
	}

	mOpts := manager.Options{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
		ConstantEditors: constEditor,
		MapEditors: map[string]*ebpf.Map{
			"bpfmap_sockfd": bpfMapSockFD,
		},
	}
	if buf, err := dkebpf.Asset("httpflow.o"); err != nil {
		return nil, nil, err
	} else if err := m.InitWithOptions((bytes.NewReader(buf)), mOpts); err != nil {
		return nil, nil, err
	}

	return m, r, nil
}

type HTTPFlowTracer struct {
	TPacket        *afpacket.TPacket
	gTags          map[string]string
	finReqCh       chan *HTTPReqFinishedInfo
	datakitPostURL string
}

func NewHTTPFlowTracer(tags map[string]string, datakitPostURL string) *HTTPFlowTracer {
	return &HTTPFlowTracer{
		finReqCh:       make(chan *HTTPReqFinishedInfo, 64),
		gTags:          tags,
		datakitPostURL: datakitPostURL,
	}
}

func (tracer *HTTPFlowTracer) Run(ctx context.Context, constEditor []manager.ConstantEditor, bpfMapSockFD *ebpf.Map, enableTLS bool) error {
	rawSocket, err := afpacket.NewTPacket()
	if err != nil {
		return fmt.Errorf("error creating raw socket: %w", err)
	}

	go tracer.feedHandler(ctx)
	// The underlying socket file descriptor is private, hence the use of reflection
	socketFD := int(reflect.ValueOf(rawSocket).Elem().FieldByName("fd").Int())

	logrus.Error(socketFD)
	tracer.TPacket = rawSocket

	bpfManger, r, err := NewHTTPFlowManger(socketFD, constEditor, bpfMapSockFD, tracer.reqFinishedEventHandler, enableTLS)
	if err != nil {
		return err
	}
	if err := bpfManger.Start(); err != nil {
		l.Error(err)
		return err
	}

	if enableTLS && r != nil {
		r.ScanAndUpdate()
		r.Monitor(ctx, time.Minute*5)
	}

	go func() {
		<-ctx.Done()
		tracer.TPacket.Close()
		_ = bpfManger.Stop(manager.CleanAll)
	}()

	return nil
}

func (tracer *HTTPFlowTracer) reqFinishedEventHandler(cpu int, data []byte,
	perfmap *manager.PerfMap, manager *manager.Manager) {
	eventC := (*HTTPReqFinishedInfoC)(unsafe.Pointer(&data[0])) //nolint:gosec

	httpStats := HTTPReqFinishedInfo{
		ConnInfo: ConnectionInfo{
			Saddr: (*(*[4]uint32)(unsafe.Pointer(&eventC.conn_info.saddr))),
			Sport: uint32(eventC.conn_info.sport),
			Daddr: (*(*[4]uint32)(unsafe.Pointer(&eventC.conn_info.daddr))),
			Dport: uint32(eventC.conn_info.dport),
			Meta:  uint32(eventC.conn_info.meta),
		},
		HTTPStats: HTTPStats{
			Payload:     unix.ByteSliceToString((*(*[HTTPPayloadMaxsize]byte)(unsafe.Pointer(&eventC.http_stats.payload)))[:]),
			ReqMethod:   uint8(eventC.http_stats.req_method),
			HTTPVersion: uint32(eventC.http_stats.http_version),
			RespCode:    uint32(eventC.http_stats.resp_code),
			ReqTS:       uint64(eventC.http_stats.req_ts),
			RespTS:      uint64(eventC.http_stats.resp_ts),
		},
	}
	tracer.finReqCh <- &httpStats
	// l.Warnf("%#v", httpStats)
}

func (tracer *HTTPFlowTracer) feedHandler(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 10)
	cache := []*HTTPReqFinishedInfo{}
	for {
		select {
		case <-ticker.C:
			if len(cache) == 0 {
				continue
			}
			if err := feed(tracer.datakitPostURL, cache, tracer.gTags); err != nil {
				l.Error(err)
			}
			cache = make([]*HTTPReqFinishedInfo, 0)
		case finReq := <-tracer.finReqCh:
			cache = append(cache, finReq)
			if len(cache) > 256 {
				if err := feed(tracer.datakitPostURL, cache, tracer.gTags); err != nil {
					l.Error(err)
				}
				cache = make([]*HTTPReqFinishedInfo, 0)
			}
		case <-ctx.Done():
			return
		}
	}
}

func conv2M(httpFinReq *HTTPReqFinishedInfo, tags map[string]string) (*point.Point, error) {
	// name:   srcNameM,
	mTags := map[string]string{}

	direction := DirectionOutgoing
	if _, err := dknetflow.SrcIPPortRecorder.Query(httpFinReq.ConnInfo.Daddr); err == nil {
		httpFinReq.ConnInfo.Saddr, httpFinReq.ConnInfo.Daddr = httpFinReq.ConnInfo.Daddr, httpFinReq.ConnInfo.Saddr
		httpFinReq.ConnInfo.Sport, httpFinReq.ConnInfo.Dport = httpFinReq.ConnInfo.Dport, httpFinReq.ConnInfo.Sport
		direction = DirectionIncoming
	}
	path := FindHTTPURI(httpFinReq.HTTPStats.Payload)
	if path == "" {
		return nil, fmt.Errorf("path == \"\"")
	}
	for k, v := range tags {
		mTags[k] = v
	}
	mTags["direction"] = direction

	isV6 := !dknetflow.ConnAddrIsIPv4(httpFinReq.ConnInfo.Meta)

	if httpFinReq.ConnInfo.Saddr[0] == 0 && httpFinReq.ConnInfo.Saddr[1] == 0 &&
		httpFinReq.ConnInfo.Daddr[0] == 0 && httpFinReq.ConnInfo.Daddr[1] == 0 {
		if httpFinReq.ConnInfo.Saddr[2] == 0xffff0000 && httpFinReq.ConnInfo.Daddr[2] == 0xffff0000 {
			isV6 = false
		} else if httpFinReq.ConnInfo.Saddr[2] == 0 && httpFinReq.ConnInfo.Daddr[2] == 0 &&
			httpFinReq.ConnInfo.Saddr[3] > 1 && httpFinReq.ConnInfo.Daddr[3] > 1 {
			isV6 = false
		}
	}
	if isV6 {
		mTags["src_ip_type"] = dknetflow.ConnIPv6Type(httpFinReq.ConnInfo.Saddr)
		mTags["dst_ip_type"] = dknetflow.ConnIPv6Type(httpFinReq.ConnInfo.Daddr)
		mTags["family"] = "IPv6"
	} else {
		mTags["src_ip_type"] = dknetflow.ConnIPv4Type(httpFinReq.ConnInfo.Saddr[3])
		mTags["dst_ip_type"] = dknetflow.ConnIPv4Type(httpFinReq.ConnInfo.Daddr[3])
		mTags["family"] = "IPv4"
	}
	srcIP := dknetflow.U32BEToIP(httpFinReq.ConnInfo.Saddr, isV6).String()
	dstIP := dknetflow.U32BEToIP(httpFinReq.ConnInfo.Daddr, isV6).String()
	mTags["src_ip"] = srcIP
	mTags["src_port"] = fmt.Sprintf("%d", httpFinReq.ConnInfo.Sport)
	mTags["dst_ip"] = dstIP
	mTags["dst_port"] = fmt.Sprintf("%d", httpFinReq.ConnInfo.Dport)

	var l4proto string
	if dknetflow.ConnProtocolIsTCP(httpFinReq.ConnInfo.Meta) {
		l4proto = "tcp"
	} else {
		l4proto = "udp"
	}
	mTags["transport"] = l4proto

	mFields := map[string]interface{}{
		"path":         path,
		"status_code":  int(httpFinReq.HTTPStats.RespCode),
		"latency":      int64(httpFinReq.HTTPStats.RespTS - httpFinReq.HTTPStats.ReqTS),
		"method":       HTTPMethodInt(int(httpFinReq.HTTPStats.ReqMethod)),
		"http_version": ParseHTTPVersion(httpFinReq.HTTPStats.HTTPVersion),
	}

	mTags = dknetflow.AddK8sTags2Map(k8sNetInfo, srcIP, dstIP,
		httpFinReq.ConnInfo.Sport, httpFinReq.ConnInfo.Dport, l4proto, mTags)
	return point.NewPoint(srcNameM, mTags, mFields, inputs.OptNetwork)
}

func feed(url string, data []*HTTPReqFinishedInfo, tags map[string]string) error {
	if len(data) == 0 {
		return nil
	}
	ms := make([]*point.Point, 0)
	for _, httpFinReq := range data {
		if !ConnNotNeedToFilter(httpFinReq.ConnInfo) {
			continue
		}
		if m, err := conv2M(httpFinReq, tags); err != nil {
			l.Error(err)
		} else {
			ms = append(ms, m)
		}
	}
	if len(ms) == 0 {
		return nil
	}
	if err := dkout.FeedMeasurement(url, ms); err != nil {
		return err
	}
	return nil
}
