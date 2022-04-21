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
	"time"
	"unsafe"

	"github.com/DataDog/ebpf/manager"
	"github.com/google/gopacket/afpacket"
	"github.com/sirupsen/logrus"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/c"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/feed"
	dknetflow "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/ebpf/netflow"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"golang.org/x/sys/unix"
)

// #include "../c/netflow/l7_stats.h"
import "C"

const HTTPPayloadMaxsize = 69

const srcNameM = "httpflow"

const (
	ConnL3Mask uint32 = dknetflow.ConnL3Mask
	ConnL3IPv4 uint32 = dknetflow.ConnL3IPv4
	ConnL3IPv6 uint32 = dknetflow.ConnL3IPv6

	ConnL4Mask uint32 = dknetflow.ConnL4Mask
	ConnL4TCP  uint32 = dknetflow.ConnL4TCP
	ConnL4UDP  uint32 = dknetflow.ConnL4UDP
)

type (
	HTTPStatsC           C.struct_http_stats
	HTTPReqFinishedInfoC C.struct_http_req_finished_info

	ConnectionInfoC dknetflow.ConnectionInfoC
	ConnectionInfo  dknetflow.ConnectionInfo
	HTTPStats       struct {
		// payload    [HTTP_PAYLOAD_MAXSIZE]byte
		payload      string
		req_method   uint8
		http_version uint32
		resp_code    uint32
		req_ts       uint64
		resp_ts      uint64
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

func NewHTTPFlowManger(fd int, closedEventHandler func(cpu int, data []byte,
	perfmap *manager.PerfMap, manager *manager.Manager)) (*manager.Manager, error) {
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
	mOpts := manager.Options{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
		// ConstantEditors: constEditor,
	}
	if buf, err := dkebpf.Asset("httpflow.o"); err != nil {
		return nil, err
	} else if err := m.InitWithOptions((bytes.NewReader(buf)), mOpts); err != nil {
		return nil, err
	}
	return m, nil
}

type HTTPFlowTracer struct {
	TPacket        *afpacket.TPacket
	gTags          map[string]string
	finReqCh       chan *HTTPReqFinishedInfo
	datakitPostURL string
}

func NewHTTPFlowTracer(tags map[string]string, datakitPostURL string) *HTTPFlowTracer {
	return &HTTPFlowTracer{
		finReqCh:       make(chan *HTTPReqFinishedInfo, 128),
		gTags:          tags,
		datakitPostURL: datakitPostURL,
	}
}

func (tracer *HTTPFlowTracer) Run(ctx context.Context) error {
	rawSocket, err := afpacket.NewTPacket()
	if err != nil {
		return fmt.Errorf("error creating raw socket: %s", err)
	}

	go tracer.feedHandler(ctx)
	// The underlying socket file descriptor is private, hence the use of reflection
	socketFD := int(reflect.ValueOf(rawSocket).Elem().FieldByName("fd").Int())

	logrus.Error(socketFD)
	tracer.TPacket = rawSocket

	bpfManger, err := NewHTTPFlowManger(socketFD, tracer.reqFinishedEventHandler)
	if err != nil {
		return err
	}
	if err := bpfManger.Start(); err != nil {
		l.Error(err)
		return err
	}

	go func() {
		<-ctx.Done()
		tracer.TPacket.Close()
		bpfManger.Stop(manager.CleanAll)
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
			payload:      unix.ByteSliceToString((*(*[HTTPPayloadMaxsize]byte)(unsafe.Pointer(&eventC.http_stats.payload)))[:]),
			req_method:   uint8(eventC.http_stats.req_method),
			http_version: uint32(eventC.http_stats.http_version),
			resp_code:    uint32(eventC.http_stats.resp_code),
			req_ts:       uint64(eventC.http_stats.req_ts),
			resp_ts:      uint64(eventC.http_stats.resp_ts),
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
			ms := make([]inputs.Measurement, 0)
			for _, httpFinReq := range cache {
				if !ConnNotNeedToFilter(httpFinReq.ConnInfo) {
					continue
				}
				m := conv2M(httpFinReq, tracer.gTags)
				if m == nil {
					continue
				}
				ms = append(ms, m)
			}
			cache = []*HTTPReqFinishedInfo{}
			if len(ms) == 0 {
				continue
			}
			if err := feed.FeedMeasurement(ms, tracer.datakitPostURL); err != nil {
				l.Error(err)
			}
		case finReq := <-tracer.finReqCh:
			cache = append(cache, finReq)
		case <-ctx.Done():
			return
		}
	}
}

type measurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *measurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *measurement) Info() *inputs.MeasurementInfo {
	return nil
}

func conv2M(httpFinReq *HTTPReqFinishedInfo, tags map[string]string) *measurement {
	m := measurement{
		name:   srcNameM,
		tags:   map[string]string{},
		fields: map[string]interface{}{},
		ts:     time.Now(),
	}
	direction := "outgoing"
	if _, err := dknetflow.SrcIPPortRecorder.Query(httpFinReq.ConnInfo.Daddr); err == nil {
		httpFinReq.ConnInfo.Saddr, httpFinReq.ConnInfo.Daddr = httpFinReq.ConnInfo.Daddr, httpFinReq.ConnInfo.Saddr
		httpFinReq.ConnInfo.Sport, httpFinReq.ConnInfo.Dport = httpFinReq.ConnInfo.Dport, httpFinReq.ConnInfo.Sport
		direction = "incoming"
	}
	path := FindHTTPURI(httpFinReq.HTTPStats.payload)
	if path == "" {
		return nil
	}
	for k, v := range tags {
		m.tags[k] = v
	}
	m.tags["direction"] = direction

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
		m.tags["src_ip_type"] = dknetflow.ConnIPv6Type(httpFinReq.ConnInfo.Saddr)
		m.tags["dst_ip_type"] = dknetflow.ConnIPv6Type(httpFinReq.ConnInfo.Daddr)
		m.tags["family"] = "IPv6"
	} else {
		m.tags["src_ip_type"] = dknetflow.ConnIPv4Type(httpFinReq.ConnInfo.Saddr[3])
		m.tags["dst_ip_type"] = dknetflow.ConnIPv4Type(httpFinReq.ConnInfo.Daddr[3])
		m.tags["family"] = "IPv4"
	}
	m.tags["src_ip"] = dknetflow.U32BEToIP(httpFinReq.ConnInfo.Saddr, isV6).String()
	m.tags["src_port"] = fmt.Sprintf("%d", httpFinReq.ConnInfo.Sport)
	m.tags["dst_ip"] = dknetflow.U32BEToIP(httpFinReq.ConnInfo.Daddr, isV6).String()
	m.tags["dst_port"] = fmt.Sprintf("%d", httpFinReq.ConnInfo.Dport)

	if dknetflow.ConnProtocolIsTCP(httpFinReq.ConnInfo.Meta) {
		m.tags["transport"] = "tcp"
	} else {
		m.tags["transport"] = "udp"
	}

	m.fields = map[string]interface{}{
		"path":         path,
		"status_code":  int(httpFinReq.HTTPStats.resp_code),
		"latency":      int64(httpFinReq.HTTPStats.resp_ts - httpFinReq.HTTPStats.req_ts),
		"method":       HTTPMethodInt(int(httpFinReq.HTTPStats.req_method)),
		"http_version": ParseHTTPVersion(httpFinReq.HTTPStats.http_version),
	}

	if k8sNetInfo != nil {
		srcK8sFlag := false
		dstK8sFlag := false
		_, srcPodName, srcSvcName, ns, svcP, err := k8sNetInfo.QueryPodInfo(m.tags["src_ip"], httpFinReq.ConnInfo.Sport, m.tags["transport"])
		if err == nil {
			srcK8sFlag = true
			m.tags["src_k8s_namespace"] = ns
			m.tags["src_k8s_pod_name"] = srcPodName
			m.tags["src_k8s_service_name"] = srcSvcName
			if svcP == httpFinReq.ConnInfo.Sport {
				m.tags["direction"] = "incoming"
			}
		}

		_, dstPoName, dstSvcName, ns, svcP, err := k8sNetInfo.QueryPodInfo(m.tags["dst_ip"], httpFinReq.ConnInfo.Dport, m.tags["transport"])
		if err == nil {
			dstK8sFlag = true
			m.tags["dst_k8s_namespace"] = ns
			m.tags["dst_k8s_pod_name"] = dstPoName
			m.tags["dst_k8s_service_name"] = dstSvcName
			if svcP == httpFinReq.ConnInfo.Dport {
				m.tags["direction"] = "outgoing"
			}
		} else {
			dstSvcName, ns, err := k8sNetInfo.QuerySvcInfo(m.tags["dst_ip"])
			if err == nil {
				dstK8sFlag = true
				m.tags["dst_k8s_namespace"] = ns
				m.tags["dst_k8s_pod_name"] = "N/A"
				m.tags["dst_k8s_service_name"] = dstSvcName
				m.tags["direction"] = "outgoing"
			}
		}

		if srcK8sFlag || dstK8sFlag {
			m.tags["sub_source"] = "K8s"
			if !srcK8sFlag {
				m.tags["src_k8s_namespace"] = "N/A"
				m.tags["src_k8s_pod_name"] = "N/A"
				m.tags["src_k8s_service_name"] = "N/A"
			}
			if !dstK8sFlag {
				m.tags["dst_k8s_namespace"] = "N/A"
				m.tags["dst_k8s_pod_name"] = "N/A"
				m.tags["dst_k8s_service_name"] = "N/A"
			}
		}
	}
	return &m
}
