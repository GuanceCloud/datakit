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
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/sirupsen/logrus"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/c"
	dknetflow "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/netflow"
	dkout "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/output"
	sysmonitor "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/sysmonitor"
	"golang.org/x/sys/unix"
)

// #include "../c/netflow/l7_stats.h"
import "C"

const HTTPPayloadMaxsize = 157

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
	//nolint:stylecheck
	HTTPStatsC struct {
		payload      [HTTPPayloadMaxsize]uint8
		req_method   uint8
		resp_code    uint16
		http_version uint32
		req_ts       uint64
		c_s_pid      uint64
	}
	//nolint:stylecheck
	HTTPReqFinishedInfoC struct {
		conn_info  C.struct_connection_info
		http_stats HTTPStatsC
	}

	ConnectionInfoC dknetflow.ConnectionInfoC
	ConnectionInfo  dknetflow.ConnectionInfo
	HTTPStats       struct {
		// payload    [HTTP_PAYLOAD_MAXSIZE]byte
		Payload     string
		ReqMethod   uint8
		HTTPVersion uint32
		RespCode    uint32
		ReqTS       uint64
		CSPid       uint64
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

func NewHTTPFlowManger(fd int, constEditor []manager.ConstantEditor, bpfMapSockFD *ebpf.Map,
	closedEventHandler func(cpu int, data []byte, perfmap *manager.PerfMap,
		manager *manager.Manager), enableTLS bool) (*manager.Manager, *sysmonitor.UprobeDynamicLibRegister, error) {
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
	if buf, err := dkebpf.HTTPFlowBin(); err != nil {
		return nil, nil, fmt.Errorf("httpflow.o: %w", err)
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

func (tracer *HTTPFlowTracer) Run(ctx context.Context, constEditor []manager.ConstantEditor,
	bpfMapSockFD *ebpf.Map, enableTLS bool, interval time.Duration) error {
	rawSocket, err := afpacket.NewTPacket()
	if err != nil {
		return fmt.Errorf("error creating raw socket: %w", err)
	}

	go tracer.feedHandler(ctx, interval)
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
			ReqMethod:   eventC.http_stats.req_method,
			HTTPVersion: eventC.http_stats.http_version,
			RespCode:    uint32(eventC.http_stats.resp_code),
			ReqTS:       eventC.http_stats.req_ts,
			CSPid:       eventC.http_stats.c_s_pid,
		},
	}
	tracer.finReqCh <- &httpStats
	// l.Warnf("%#v", httpStats)
}

func (tracer *HTTPFlowTracer) feedHandler(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	agg := FlowAgg{}
	for {
		select {
		case <-ticker.C:
			pidMap, _ := sysmonitor.AllProcess()
			pts := agg.ToPoint(tracer.gTags, k8sNetInfo, pidMap)
			agg.Clean()
			if err := feed(tracer.datakitPostURL, pts); err != nil {
				l.Error(err)
			}
		case finReq := <-tracer.finReqCh:
			err := agg.Append(finReq)
			if err != nil {
				l.Debug(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func feed(url string, data []*client.Point) error {
	if len(data) == 0 {
		return nil
	}
	if err := dkout.FeedMeasurement(url, point.WrapPoint(data)); err != nil {
		return err
	}
	return nil
}
