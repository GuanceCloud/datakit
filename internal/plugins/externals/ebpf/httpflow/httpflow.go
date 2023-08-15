//go:build (linux && amd64 && ebpf) || (linux && arm64 && ebpf)
// +build linux,amd64,ebpf linux,arm64,ebpf

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

	"github.com/DataDog/ebpf"
	"github.com/DataDog/ebpf/manager"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/google/gopacket/afpacket"
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/shirou/gopsutil/host"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/c"
	dknetflow "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/netflow"
	dkout "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/output"
	sysmonitor "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/sysmonitor"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/tracing"
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

	PayloadBufSize = 2048
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

		Pid uint64

		Recv int
		Send int

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

type perferEventHandle func(cpu int, data []byte, perfmap *manager.PerfMap,
	manager *manager.Manager)

func NewHTTPFlowManger(constEditor []manager.ConstantEditor, ctMap *ebpf.Map,
	closedEventHandler, bufHandler perferEventHandle, enableTLS bool) (*manager.Manager, *sysmonitor.UprobeDynamicLibRegister, error) {
	m := &manager.Manager{
		Probes: []*manager.Probe{
			{
				Section: "tracepoint/syscalls/sys_enter_read",
			},
			{
				Section: "tracepoint/syscalls/sys_exit_read",
			},
			{
				Section: "tracepoint/syscalls/sys_enter_write",
			},
			{
				Section: "tracepoint/syscalls/sys_exit_write",
			},

			{
				Section: "tracepoint/syscalls/sys_enter_recvfrom",
			},
			{
				Section: "tracepoint/syscalls/sys_exit_recvfrom",
			},
			{
				Section: "tracepoint/syscalls/sys_enter_sendto",
			},
			{
				Section: "tracepoint/syscalls/sys_exit_sendto",
			},
			{
				Section: "tracepoint/syscalls/sys_enter_writev",
			},
			{
				Section: "tracepoint/syscalls/sys_exit_writev",
			},
			{
				Section: "tracepoint/syscalls/sys_enter_readv",
			},
			{
				Section: "tracepoint/syscalls/sys_exit_readv",
			},
			{
				Section: "tracepoint/syscalls/sys_enter_close",
			},
			{
				Section: "tracepoint/syscalls/sys_enter_sendfile64",
			},
			{
				Section: "tracepoint/syscalls/sys_exit_sendfile64",
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
	}

	if ctMap != nil {
		mOpts.MapEditors = map[string]*ebpf.Map{
			"bpfmap_conntrack_tuple": ctMap,
		}
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
}

func NewHTTPFlowTracer(tags map[string]string, datakitPostURL string) *HTTPFlowTracer {
	return &HTTPFlowTracer{
		gTags:          tags,
		datakitPostURL: datakitPostURL,
	}
}

func (tracer *HTTPFlowTracer) Run(ctx context.Context, constEditor []manager.ConstantEditor,
	ctMap *ebpf.Map, enableTLS bool, interval time.Duration) error {
	// rawSocket, err := afpacket.NewTPacket()
	// if err != nil {
	// 	return fmt.Errorf("error creating raw socket: %w", err)
	// }

	go tracer.feedHandler(ctx, interval)
	// The underlying socket file descriptor is private, hence the use of reflection
	// socketFD := int(reflect.ValueOf(rawSocket).Elem().FieldByName("fd").Int())

	// tracer.TPacket = rawSocket

	bpfManger, r, err := NewHTTPFlowManger(constEditor, ctMap,
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
	eventC := (*CHTTPReqFinished)(unsafe.Pointer(&data[0])) //nolint:gosec

	httpStats := &HTTPReqFinishedInfo{
		ConnInfo: ConnectionInfo{
			Saddr:    (*(*[4]uint32)(unsafe.Pointer(&eventC.conn_info.saddr))),
			Sport:    uint32(eventC.conn_info.sport),
			Daddr:    (*(*[4]uint32)(unsafe.Pointer(&eventC.conn_info.daddr))),
			Dport:    uint32(eventC.conn_info.dport),
			Meta:     uint32(eventC.conn_info.meta),
			Pid:      uint32(eventC.conn_info.pid),
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
			Pid:         uint64(eventC.conn_info.pid),
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

	// if err := os.WriteFile(fmt.Sprintf("./tcp_payload_dump/%d", time.Now().UnixNano()),
	// 	b[:bufLen], 0o777); err != nil {
	// 	fmt.Println(err)
	// }

	bufCopy := make([]byte, 0, bufLen)
	bufCopy = append(bufCopy, b[:bufLen]...)

	info, ok := tracing.ParseHTTP1xHeader(bufCopy, ts)
	if !ok {
		return
	}

	_reqCache.AppendPayload(CPayloadID(bufferC.id), info)
}

func (tracer *HTTPFlowTracer) feedHandler(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	mergeTicker := time.NewTicker(time.Second * 15)
	cleanCacheTicker := time.NewTicker(time.Second * 90)

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
		case <-mergeTicker.C:
			reqs := _reqCache.MergeReq()

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

func feed(url string, data []*client.Point) error {
	if len(data) == 0 {
		return nil
	}
	if err := dkout.FeedMeasurement(url, point.WrapPoint(data)); err != nil {
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

			iter := m.IterateFrom(connStatsC)

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
