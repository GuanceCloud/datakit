//go:build linux
// +build linux

// Package httpflow collects http(s) request flow
package httpflow

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"regexp"
	"time"
	"unsafe"

	manager "github.com/DataDog/ebpf-manager"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cilium/ebpf"
	"github.com/google/gopacket/afpacket"
	"github.com/shirou/gopsutil/host"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/c"
	dknetflow "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
	dkout "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/output"
	sysmonitor "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/sysmonitor"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/tracing"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/hash"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/spanid"
	"golang.org/x/sys/unix"
)

// #include "../c/apiflow/l7_stats.h"
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

	PayloadBufSize    = 2048
	KernelTaskCommLen = 16
)

var (
	// libssl
	regexpLibSSL    = regexp.MustCompile(`libssl.so`)
	regexpLibCrypto = regexp.MustCompile(`libcrypto.so`)

	// TODO: guntls
)

type (
	CLayer7Http      C.struct_layer7_http
	CHTTPReqFinished C.struct_http_req_finished
	CL7Buffer        C.struct_l7_buffer

	ConnectionInfoC dknetflow.ConnectionInfoC
	ConnectionInfo  dknetflow.ConnectionInfo

	CPayloadID C.struct_payload_id
	HTTPStats  struct {
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
		ConnInfo  ConnectionInfo
		HTTPStats HTTPStats
	}
)

func (conn ConnectionInfo) String() string {
	return fmt.Sprintf("%s:%d -> %s:%d, pid:%d, tcp:%t", dknetflow.U32BEToIP(conn.Saddr,
		!dknetflow.ConnAddrIsIPv4(conn.Meta)), conn.Sport,
		dknetflow.U32BEToIP(conn.Daddr, !dknetflow.ConnAddrIsIPv4(conn.Meta)),
		conn.Dport, conn.Pid, dknetflow.ConnProtocolIsTCP(conn.Meta))
}

func (payloadid CPayloadID) String() string {
	return fmt.Sprintf("cpu%d,ktime%d,pid%d,tid%d,random%d",
		payloadid.cpuid, payloadid.ktime, payloadid.pid_tid>>32, payloadid.pid_tid&0xFFFFFFFF, payloadid.prandom)
}

var l = logger.DefaultSLogger("ebpf")

func SetLogger(nl *logger.Logger) {
	l = nl
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

type perferEventHandle func(cpu int, data []byte, perfmap *manager.PerfMap,
	manager *manager.Manager)

func NewHTTPFlowManger(constEditor []manager.ConstantEditor, bmaps map[string]*ebpf.Map,
	closedEventHandler, bufHandler perferEventHandle, enableTLS bool) (*manager.Manager, *sysmonitor.UprobeRegister, error) {
	m := &manager.Manager{
		Probes: []*manager.Probe{
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sys_enter_read",
				},
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sys_exit_read",
				},
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sys_enter_write",
				},
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sys_exit_write",
				},
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sys_enter_recvfrom",
				},
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sys_exit_recvfrom",
				},
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sys_enter_sendto",
				},
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sys_exit_sendto",
				},
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sys_enter_writev",
				},
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sys_exit_writev",
				},
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sys_enter_readv",
				},
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sys_exit_readv",
				},
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "kprobe__tcp_close",
					UID:          "tcp_close_apiflow",
				},
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sys_enter_sendfile64",
				},
			},
			{
				ProbeIdentificationPair: manager.ProbeIdentificationPair{
					EBPFFuncName: "tracepoint__sys_exit_sendfile64",
				},
			},
		},
		PerfMaps: []*manager.PerfMap{
			{
				Map: manager.Map{
					Name: "bpfmap_httpreq_fin_event",
				},
				PerfMapOptions: manager.PerfMapOptions{
					// pagesize ~= 4k,
					PerfRingBufferSize: 128 * os.Getpagesize(),
					DataHandler:        closedEventHandler,
				},
			},
			{
				Map: manager.Map{
					Name: "bpfmap_l7_buffer_out",
				},
				PerfMapOptions: manager.PerfMapOptions{
					// pagesize ~= 4k,
					PerfRingBufferSize: 512 * os.Getpagesize(),
					DataHandler:        bufHandler,
				},
			},
		},
	}

	var r *sysmonitor.UprobeRegister
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
	}

	if bmaps != nil {
		mOpts.MapEditors = bmaps
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
	datakitPostURL string
	tracePostURL   string
	conv2dd        bool
	enableTrace    bool
	procFilter     *tracing.ProcessFilter
}

func NewHTTPFlowTracer(tags map[string]string, datakitPostURL, tracePostURL string,
	conv2dd, enableTrace bool, filter *tracing.ProcessFilter) *HTTPFlowTracer {
	return &HTTPFlowTracer{
		gTags:          tags,
		datakitPostURL: datakitPostURL,
		tracePostURL:   tracePostURL,
		conv2dd:        conv2dd,
		enableTrace:    enableTrace,
		procFilter:     filter,
	}
}

func (tracer *HTTPFlowTracer) Run(ctx context.Context, constEditor []manager.ConstantEditor,
	bmaps map[string]*ebpf.Map, enableTLS bool, interval time.Duration) error {
	// rawSocket, err := afpacket.NewTPacket()
	// if err != nil {
	// 	return fmt.Errorf("error creating raw socket: %w", err)
	// }

	go tracer.feedHandler(ctx, interval)
	// The underlying socket file descriptor is private, hence the use of reflection
	// socketFD := int(reflect.ValueOf(rawSocket).Elem().FieldByName("fd").Int())

	// tracer.TPacket = rawSocket

	bpfManger, r, err := NewHTTPFlowManger(constEditor, bmaps,
		tracer.reqFinishedEventHandler, tracer.bufHandle, enableTLS)
	if err != nil {
		return err
	}
	if err := bpfManger.Start(); err != nil {
		l.Error(err)
		return err
	}

	httpStatsMap, found, err := bpfManger.GetMap("bpfmap_http_stats")
	if err != nil || !found {
		return err
	}

	go cleanupBPFMapConn(ctx, httpStatsMap)

	if r != nil {
		r.ScanAndUpdate()
		r.Monitor(ctx, time.Second*30)
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
	eventC := (*CHTTPReqFinished)(unsafe.Pointer(&data[0])) //nolint:gosec

	httpStats := &HTTPReqFinishedInfo{
		ConnInfo: ConnectionInfo{
			Saddr:    (*(*[4]uint32)(unsafe.Pointer(&eventC.conn_info.saddr))),
			Sport:    uint32(eventC.conn_info.sport),
			Daddr:    (*(*[4]uint32)(unsafe.Pointer(&eventC.conn_info.daddr))),
			Dport:    uint32(eventC.conn_info.dport),
			Meta:     uint32(eventC.conn_info.meta),
			Pid:      uint32(eventC.conn_info.pid),
			Netns:    uint32(eventC.conn_info.netns),
			NATDaddr: (*(*[4]uint32)(unsafe.Pointer(&eventC.http.nat_daddr))),
			NATDport: uint32(eventC.http.nat_dport),
		},
		HTTPStats: HTTPStats{
			Direction:   dknetflow.ConnDirection2Str(uint8(eventC.http.direction)),
			ReqMethod:   uint8(eventC.http.method),
			HTTPVersion: uint32(eventC.http.http_version),
			RespCode:    uint32(eventC.http.status_code),
			ReqTS:       uint64(eventC.http.req_ts),
			RespTS:      uint64(eventC.http.resp_ts),
			Recv:        int(eventC.http.recv_bytes),
			Send:        int(eventC.http.sent_bytes),
			ReqSeq:      int64(eventC.http.req_seq),
			RespSeq:     int64(eventC.http.resp_seq),
			// PidTid:      uint64(eventC.http.req_payload_id.pid_tid),
		},
	}

	_reqCache.AppendFinReq(CPayloadID(eventC.http.req_payload_id), httpStats)
}

func (tracer *HTTPFlowTracer) bufHandle(cpu int, data []byte,
	perfmap *manager.PerfMap, manager *manager.Manager) {
	ts := time.Now().UnixNano()

	bufferC := *((*CL7Buffer)(unsafe.Pointer(&data[0])))

	bufLen := int(bufferC.len)

	if bufLen <= 0 {
		return
	}
	if bufLen > PayloadBufSize {
		bufLen = PayloadBufSize
	}

	b := *(*[PayloadBufSize]byte)(unsafe.Pointer(&bufferC.payload))

	bufCopy := make([]byte, 0, bufLen)
	bufCopy = append(bufCopy, b[:bufLen]...)

	info, ok := tracing.ParseHTTP1xHeader(bufCopy, ts, tracer.conv2dd)
	if !ok {
		return
	}

	// if err := os.WriteFile(fmt.Sprintf("./tcp_payload_dump/%d", time.Now().UnixNano()),
	// b[:bufLen], 0o777); err != nil {
	// fmt.Println(err)
	// }

	cmdB := *(*[KernelTaskCommLen]byte)(unsafe.Pointer(&bufferC.cmd))
	cmdCpy := make([]byte, 0, len(bufferC.cmd))
	cmdCpy = append(cmdCpy, cmdB[:]...)
	cmdCpy = bytes.Trim(cmdCpy, "\u0000")
	cmdCpy = bytes.TrimSpace(cmdCpy)
	info.TaskComm = string(cmdCpy)

	info.ThrTraceid = spanid.ID64(
		TransPayloadToThrID((*CPayloadID)(&bufferC.thr_trace_id)))
	info.PidTid = uint64(bufferC.id.pid_tid)

	if tracer.procFilter != nil {
		if v, ok := tracer.procFilter.GetProcInfo(int(info.PidTid >> 32)); ok {
			if v.AllowTrace {
				info.AllowTrace = true
			}
			info.ProcessName = v.Name
			info.Service = v.Service
		}
	}

	if bufferC.isentry == 1 {
		info.ESpanType = "entry"
	}

	_reqCache.AppendPayload(CPayloadID(bufferC.id), info)
}

func TransPayloadToThrID(id *CPayloadID) uint64 {
	h := hash.Fnv1aNew()
	b := (*[24]byte)(unsafe.Pointer(id))
	return hash.Fnv1aHashAddByte(h, b[:])
}

func (tracer *HTTPFlowTracer) feedHandler(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	mergeTicker := time.NewTicker(time.Second * 2)
	cleanCacheTicker := time.NewTicker(time.Second * 90)

	var eTrace bool
	if tracer.enableTrace && tracer.tracePostURL != "" {
		eTrace = true
	}

	agg := FlowAgg{}
	for {
		select {
		case <-ticker.C:
			pts := agg.ToPoint(tracer.gTags, k8sNetInfo)
			agg.Clean()
			if err := feed(tracer.datakitPostURL, pts, false); err != nil {
				l.Error(err)
			}

		case <-mergeTicker.C:
			reqs, pts := _reqCache.MergeReq(tracer.gTags, eTrace, tracer.procFilter)
			if err := feed(tracer.tracePostURL, pts, true); err != nil {
				l.Error(err)
			}
			for _, req := range reqs {
				err := agg.Append(req)
				if err != nil {
					l.Debug(err)
				}
			}

		case <-cleanCacheTicker.C:
			_reqCache.CleanPathExpr()
		case <-ctx.Done():
			return
		}
	}
}

func feed(url string, data []*point.Point, gzip bool) error {
	if len(data) == 0 {
		return nil
	}
	if err := dkout.FeedPoint(url, data, gzip); err != nil {
		return err
	}
	return nil
}

func cleanupBPFMapConn(ctx context.Context, m *ebpf.Map) {
	ticker := time.NewTicker(time.Minute * 15)
	for {
		select {
		case <-ticker.C:
			uptime, err := host.Uptime() // seconds since boot
			if err != nil {
				l.Error(err)
				return
			}
			var connStatsC dknetflow.ConnectionStatsC
			var layer7HTTP CLayer7Http

			iter := m.Iterate()

			connNeedDel := []dknetflow.ConnectionStatsC{}

			count := 0
			for iter.Next(unsafe.Pointer(&connStatsC), unsafe.Pointer(&layer7HTTP)) {
				count++
				ts := uint64(0)
				if layer7HTTP.req_ts != 0 {
					ts = uint64(layer7HTTP.req_ts)
				}
				if layer7HTTP.resp_ts != 0 {
					ts = uint64(layer7HTTP.resp_ts)
				}
				if reqExpr(uptime, ts) {
					connNeedDel = append(connNeedDel, connStatsC)
				}
			}

			l.Debugf("Total number of conn resources: %d.", count)
			l.Debugf("The number of connections that need to be cleaned up: %d.", len(connNeedDel))

			if len(connNeedDel) > 0 {
				l.Info("cleannup %d unfinished http conn key", len(connNeedDel))
				for _, v := range connNeedDel {
					if err = m.Delete(unsafe.Pointer(&v)); err != nil {
						l.Warn(err)
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
