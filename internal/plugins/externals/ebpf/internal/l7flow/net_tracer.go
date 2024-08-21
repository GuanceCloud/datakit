//go:build linux
// +build linux

package l7flow

import (
	"context"
	"strconv"
	"sync"
	"time"
	"unsafe"

	manager "github.com/DataDog/ebpf-manager"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/k8sinfo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/protodec"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/tracing"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/spanid"
)

type NetTrace struct {
	// 对于每个进程， ingress 请求抵达时，后续的网络请求都应该继承此生成的 innter trace id
	ESpanLinkDuration time.Duration

	// (kernel) sock ptr and random id ==> network flow pipe
	ConnMap       map[CUniID]*FlowPipe
	ConnMapClosed map[CUniID]*FlowPipe

	ConnAndClosedDelCount [2]int

	threadInnerID comm.ThreadTrace

	ptsPrv []*point.Point
	ptsCur []*point.Point

	allowESPan bool
}

func (netTrace *NetTrace) StreamHandle(tn int64, uniID CUniID, data *comm.NetwrkData,
	aggPool map[protodec.L7Protocol]protodec.AggPool, allowTrace bool,
	protoLi map[protodec.L7Protocol]struct{},
) {
	if netTrace.ConnMap == nil {
		netTrace.ConnMap = make(map[CUniID]*FlowPipe)
	}

	var pipe *FlowPipe
	var inClosedMap bool

	// check in closed map first
	if p, ok := netTrace.ConnMapClosed[uniID]; ok {
		pipe = p
		inClosedMap = true
	}

	if pipe == nil {
		var ok bool
		pipe, ok = netTrace.ConnMap[uniID]
		if !ok {
			pipe = &FlowPipe{
				Conn: data.Conn,
				sort: dataQueue{prvDataPos: 0},
			}
			netTrace.ConnMap[uniID] = pipe
		}
	}

	pipe.lastTime = tn

	// check conn close
	if data.Fn == comm.FnSysClose {
		pipe.connClosed = true

		delete(netTrace.ConnMap, uniID)
		netTrace.ConnMapClosed[uniID] = pipe
		inClosedMap = true
	}

	var dataLi []*comm.NetwrkData
	if pipe.detecTimes < 64 || pipe.Decoder != nil {
		dataLi = pipe.sort.Queue(data)
	} else {
		pipe.sort.li = nil
	}

	defer func(li []*comm.NetwrkData) {
		for _, d := range li {
			putNetwrkData(d)
		}
	}(dataLi)

	var connClose bool
	for _, d := range dataLi {
		if d.Fn == comm.FnSysClose {
			connClose = true
			continue
		}

		txRx := comm.FnInOut(d.Fn)
		if txRx == comm.NICDUnknown {
			continue
		}

		if pipe.Proto == protodec.ProtoUnknown {
			if proto, dec, ok := protodec.ProtoDetector(d.Payload, d.CaptureSize); ok {
				pipe.Proto = proto
				if _, ok := protoLi[pipe.Proto]; !ok && pipe.Proto != protodec.ProtoHTTP {
					continue
				} else {
					pipe.Decoder = dec
				}

				if proto == protodec.ProtoHTTP2 {
					continue
				}
			} else {
				pipe.detecTimes++
				continue
			}
		}

		if pipe.Decoder != nil && d.CaptureSize > 0 {
			pipe.Decoder.Decode(txRx, d, time.Now().UnixNano(), &netTrace.threadInnerID)
		}
	}

	if connClose {
		pipe.connClosed = true
		if pipe.Decoder != nil {
			pipe.Decoder.ConnClose()
		}
		if inClosedMap {
			netTrace.ConnAndClosedDelCount[1]++
			delete(netTrace.ConnMapClosed, uniID)
		} else {
			netTrace.ConnAndClosedDelCount[0]++
			delete(netTrace.ConnMap, uniID)
		}
	}

	if pipe.Decoder != nil {
		if v := pipe.Decoder.Export(connClose); len(v) > 0 {
			if p := aggPool[pipe.Proto]; p != nil {
				for i := 0; i < len(v); i++ {
					p.Obs(&data.Conn, v[i])
				}
			}
			if _, ok := protoLi[pipe.Proto]; ok {
				if netTrace.allowESPan && pipe.Conn.ProcessName != "datakit" &&
					pipe.Conn.ProcessName != "datakit-ebpf" {
					pts := genPts(v, &data.Conn)
					netTrace.ptsCur = append(netTrace.ptsCur, pts...)
				}
			}
		}
	}
}

type FlowPipe struct {
	Conn       comm.ConnectionInfo
	Decoder    protodec.ProtoDecPipe
	Proto      protodec.L7Protocol
	detecTimes int

	sort dataQueue

	lastTime int64

	connClosed bool
}

type ConnWatcher struct {
	netracer *NetTrace
	aggPool  map[protodec.L7Protocol]protodec.AggPool

	procFilter *tracing.ProcessFilter

	tags          map[string]string
	k8sInfo       *k8sinfo.K8sNetInfo
	metricPostURL string
	tracePostURL  string
	enableTrace   bool

	enabledProto map[protodec.L7Protocol]struct{}

	sync.Mutex
}

func (watcher *ConnWatcher) handle(tn int64, uniID CUniID, netdata *comm.NetwrkData) {
	watcher.Lock()
	defer watcher.Unlock()

	pid := int(netdata.Conn.Pid)

	nettracer := watcher.netracer
	if nettracer == nil {
		return
	}

	if watcher.enableTrace && watcher.procFilter != nil {
		nettracer.allowESPan = false
		if v, ok := watcher.procFilter.GetProcInfo(pid); ok {
			nettracer.allowESPan = v.AllowTrace
			netdata.Conn.ProcessName = v.Name
			netdata.Conn.ServiceName = v.Service
		}
	}

	nettracer.StreamHandle(tn, uniID, netdata, watcher.aggPool, watcher.enableTrace, watcher.enabledProto)
}

func (watcher *ConnWatcher) start(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 10)
	tickerClean := time.NewTicker(time.Minute * 5)
	tickerCheck := time.NewTicker(time.Minute * 2)

	for {
		select {
		case <-tickerCheck.C:
			watcher.Lock()
			groupTime := time.Now().UnixNano()
			for uniID, pipe := range watcher.netracer.ConnMap {
				if groupTime-pipe.lastTime > int64(time.Minute)*3 {
					delete(watcher.netracer.ConnMap, uniID)
					watcher.netracer.ConnAndClosedDelCount[0]++
				}
			}

			for uniID, pipe := range watcher.netracer.ConnMapClosed {
				if groupTime-pipe.lastTime > int64(time.Minute) {
					delete(watcher.netracer.ConnMapClosed, uniID)
					watcher.netracer.ConnAndClosedDelCount[1]++
				}
			}

			if watcher.netracer.ConnAndClosedDelCount[0] > 160_000 {
				watcher.netracer.ConnAndClosedDelCount[0] = 0
				connMap := make(map[CUniID]*FlowPipe)
				for k, v := range watcher.netracer.ConnMap {
					connMap[k] = v
				}
			}

			if watcher.netracer.ConnAndClosedDelCount[1] > 160_000 {
				watcher.netracer.ConnAndClosedDelCount[1] = 0
				connMap := make(map[CUniID]*FlowPipe)
				for k, v := range watcher.netracer.ConnMapClosed {
					connMap[k] = v
				}
			}

			watcher.Unlock()
		case <-tickerClean.C:
			watcher.Lock()
			if tracer := watcher.netracer; tracer != nil {
				tracer.threadInnerID.Cleanup()
			}
			watcher.Unlock()
		case <-ticker.C:
			watcher.Lock()
			if tracer := watcher.netracer; tracer != nil {
				for _, pt := range tracer.ptsPrv {
					setInnerID(pt, &tracer.threadInnerID)
				}
				if err := feed(watcher.tracePostURL, tracer.ptsPrv, false); err != nil {
					log.Error(err)
				}
				// feed("http://0.0.0.0:9529/v1/write/logging?input=ebpf-net%2Fespan", v.ptsPrv, false)
				tracer.ptsPrv = tracer.ptsCur
				tracer.ptsCur = nil
			}
			watcher.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func setInnerID(pt *point.Point, threadInnerID *comm.ThreadTrace) {
	d := pt.Get(spanid.Direction).(string)
	if d != comm.DirectionOutgoing {
		return
	}

	var tid [2]int32
	if v1 := pt.Get(comm.FieldUserThread); v1 != nil {
		if v, ok := v1.(int64); ok {
			tid[1] = int32(v)
		}
	}
	if v1 := pt.Get(comm.FieldKernelThread); v1 != nil {
		if v, ok := v1.(int64); ok {
			tid[0] = int32(v)
		}
	}
	var ktime uint64
	if v1 := pt.Get(comm.FieldKernelTime); v1 != nil {
		if v, ok := v1.(int64); ok {
			ktime = uint64(v)
		}
	}
	var pid int32
	if v := pt.Get(comm.FieldPid); v != nil {
		if v, ok := v.(int64); ok {
			pid = int32(v)
		}
	}
	id := threadInnerID.GetInnerID(pid, tid, ktime)
	pt.Add(spanid.ThrTraceID, id)
}

func newConnWatcher(ctx context.Context, cfg *connWatcherConfig) *ConnWatcher {
	p := &ConnWatcher{
		netracer: &NetTrace{
			ConnMap:       make(map[CUniID]*FlowPipe),
			ConnMapClosed: make(map[CUniID]*FlowPipe),
		},
		aggPool:       cfg.aggPool,
		procFilter:    cfg.procFilter,
		tags:          cfg.tags,
		k8sInfo:       cfg.k8sNetInfo,
		metricPostURL: cfg.datakitPostURL,
		tracePostURL:  cfg.tracePostURL,
		enableTrace:   cfg.enableTrace,
		enabledProto:  cfg.protos,
	}
	go p.start(ctx)
	return p
}

type Tracer struct {
	connWatcher *ConnWatcher

	aggPool map[protodec.L7Protocol]protodec.AggPool

	tags    map[string]string
	k8sInfo *k8sinfo.K8sNetInfo

	metricPostURL string
	tracePostURL  string
	enableTrace   bool
}

func (tracer *Tracer) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			for _, p := range tracer.aggPool {
				pts := p.Export(tracer.tags, tracer.k8sInfo)
				if len(pts) > 0 {
					p.Cleanup()
					if err := feed(tracer.metricPostURL, pts, false); err != nil {
						log.Error(err)
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (tracer *Tracer) PerfEventHandle(cpu int, data []byte,
	perfmap *manager.PerfMap, manager *manager.Manager,
) {
	events := (*CNetEvents)(unsafe.Pointer(&data[0])) //nolint:gosec

	eventsNum := int(events.pos.num)
	// eventsLen := int(events.pos.len)

	hdrSize := unsafe.Sizeof(CNetEventComm{}) // nolint:gosec

	pos := int(unsafe.Sizeof(events.pos)) // nolint:gosec

	groupTime := time.Now().UnixNano()

	for i := 0; i < eventsNum; i++ {
		commHdr := *(*CNetEventComm)(unsafe.Pointer(&data[pos])) //nolint:gosec
		pos += int(hdrSize)

		netdata := getNetwrkData(int(commHdr.len))

		readMeta(&commHdr, &netdata.Conn)

		if int(commHdr.len) > 0 {
			v := unsafe.Slice((*byte)(unsafe.Pointer(&data[pos])), int(commHdr.len)) //nolint:gosec
			netdata.Payload = append(netdata.Payload, v...)

			pos += int(commHdr.len)
		}

		pid := int(netdata.Conn.Pid)
		if pid == selfPid {
			return
		}

		netdata.Fn = comm.FnID(commHdr.meta.func_id)

		netdata.CaptureSize = int(commHdr.len)
		netdata.FnCallSize = int(commHdr.meta.original_size)
		netdata.TCPSeq = uint32(commHdr.meta.tcp_seq)
		netdata.Thread = [2]int32{int32(commHdr.meta.tid_utid >> 32), (int32(commHdr.meta.tid_utid))}
		netdata.TS = uint64(commHdr.meta.ts)
		netdata.TSTail = uint64(commHdr.meta.ts_tail)
		netdata.Index = uint64(commHdr.meta.sk_inf.index)

		// log.Info(netdata.String())

		tracer.connWatcher.handle(groupTime, CUniID(commHdr.meta.sk_inf.uni_id), netdata)
	}
}

type connWatcherConfig struct {
	apiTracerConfig
	aggPool map[protodec.L7Protocol]protodec.AggPool
}

func newTracer(ctx context.Context, cfg *apiTracerConfig) *Tracer {
	if cfg == nil {
		return nil
	}

	aggP := protodec.NewProtoAggregators()

	return &Tracer{
		connWatcher: newConnWatcher(ctx, &connWatcherConfig{
			apiTracerConfig: *cfg,
			aggPool:         aggP,
		}),
		aggPool:       aggP,
		tags:          cfg.tags,
		k8sInfo:       cfg.k8sNetInfo,
		metricPostURL: cfg.datakitPostURL,
		tracePostURL:  cfg.tracePostURL,
		enableTrace:   cfg.enableTrace,
	}
}

func genPts(data []*protodec.ProtoData, conn *comm.ConnectionInfo) []*point.Point {
	var pts []*point.Point
	for _, v := range data {
		// comm trace fields
		var spanType string
		switch v.Direction { //nolint:exhaustive
		case comm.DIn:
			spanType = "entry"
		case comm.DOut:
			spanType = "exit"
		default:
			spanType = "unknown"
		}

		// network tracing
		v.KVs = v.KVs.Add(spanid.EBPFSpanType, spanType, false, true)
		v.KVs = v.KVs.Add(spanid.ReqSeq, int64(v.Meta.ReqTCPSeq), false, true)
		v.KVs = v.KVs.Add(spanid.RespSeq, int64(v.Meta.RespTCPSeq), false, true)
		v.KVs = v.KVs.Add(spanid.Direction, v.Direction.String(), false, true)
		// working for process thread inner tracing
		if v.Direction == comm.DIn {
			v.KVs = v.KVs.Add(spanid.ThrTraceID, v.Meta.InnerID, false, true)
		}
		v.KVs = v.KVs.Add(comm.FieldKernelThread, v.Meta.Threads[0][0], false, true)
		if v.Meta.Threads[0][1] != 0 {
			v.KVs = v.KVs.Add(comm.FieldUserThread, v.Meta.Threads[0][1], false, true)
		}
		v.KVs = v.KVs.Add(comm.FieldKernelTime, int64(v.KTime), false, true)

		// app trace info
		if !v.Meta.TraceID.Zero() && !v.Meta.ParentSpanID.Zero() {
			v.KVs = v.KVs.Add(spanid.AppTraceIDL, int64(v.Meta.TraceID.Low), false, true)
			v.KVs = v.KVs.Add(spanid.AppTraceIDH, int64(v.Meta.TraceID.High), false, true)
			v.KVs = v.KVs.Add(spanid.AppParentIDL, int64(v.Meta.ParentSpanID), false, true)
			var aSampled int64
			if v.Meta.SampledSpan {
				aSampled = 1
			} else {
				aSampled = -1
			}
			v.KVs = v.KVs.Add(spanid.AppSpanSampled, aSampled, false, true)
			if v.Meta.SpanHexEnc {
				v.KVs = v.KVs.Add("app_trace_id", v.Meta.TraceID.StringHex(), false, true)
				v.KVs = v.KVs.Add("app_parent_id", v.Meta.ParentSpanID.StringHex(), false, true)
			} else {
				v.KVs = v.KVs.Add("app_trace_id", v.Meta.TraceID.StringDec(), false, true)
				v.KVs = v.KVs.Add("app_parent_id", v.Meta.ParentSpanID.StringDec(), false, true)
			}
		}

		// service info
		v.KVs = v.KVs.Add("source_type", "ebpf", false, true)
		v.KVs = v.KVs.Add("process_name", conn.ProcessName, false, true)
		v.KVs = v.KVs.Add("thread_name", conn.TaskName, false, true)
		if conn.ServiceName == "" {
			if conn.ProcessName != "" {
				v.KVs = v.KVs.Add("service", conn.ProcessName, false, true)
			} else {
				v.KVs = v.KVs.Add("service", conn.TaskName, false, true)
			}
		} else {
			v.KVs = v.KVs.Add("service", conn.ServiceName, false, true)
		}
		v.KVs = v.KVs.Add(comm.FieldPid, strconv.Itoa(int(conn.Pid)), false, true)

		// conn info
		isV6 := !netflow.ConnAddrIsIPv4(conn.Meta)
		ip := netflow.U32BEToIP(conn.Daddr, isV6)
		v.KVs = v.KVs.Add("dst_ip", ip.String(), false, true)
		ip = netflow.U32BEToIP(conn.Saddr, isV6)
		v.KVs = v.KVs.Add("src_ip", ip.String(), false, true)
		v.KVs = v.KVs.Add("src_port", strconv.Itoa(int(conn.Sport)), false, true)
		v.KVs = v.KVs.Add("dst_port", strconv.Itoa(int(conn.Dport)), false, true)

		// span info
		v.KVs = v.KVs.Add("start", v.Time/1000, false, true)        // conv ns to us
		v.KVs = v.KVs.Add("duration", v.Duration/1000, false, true) // conv ns to us
		// v.KVs = v.KVs.Add("cost", v.Cost, false, true)
		v.KVs = v.KVs.Add("span_type", spanType, false, true)

		opt := point.CommonLoggingOptions()
		opt = append(opt, point.WithTimestamp(v.Time))
		pt := point.NewPointV2("dketrace", v.KVs, opt...)
		pts = append(pts, pt)
	}
	return pts
}
