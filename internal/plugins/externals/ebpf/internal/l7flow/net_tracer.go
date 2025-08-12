//go:build linux
// +build linux

package l7flow

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	manager "github.com/DataDog/ebpf-manager"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/protodec"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
	sysmonitor "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/sysmonitor"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/cli"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/spanid"
)

type NetTrace struct {
	// 对于每个服务，ingress 请求抵达时，后续的网络请求都应该继承此生成的 innter trace id
	ESpanLinkDuration time.Duration

	// (kernel) sock ptr and random id ==> network flow pipe
	ConnMap       map[CUniID]*FlowPipe
	ConnMapClosed map[CUniID]*FlowPipe

	ConnAndClosedDelCount [2]int

	threadInnerID  comm.ThreadTrace
	protocolFilter *protoKernelFilter
	enabledProto   map[protodec.L7Protocol]struct{}
	protoSet       *protodec.ProtoSet

	ptsPrv []*point.Point
	ptsCur []*point.Point

	allowESPan bool
}

const maxDetec = 64

func (netTrace *NetTrace) StreamHandle(tn int64, uniID CUniID, data *comm.NetwrkData,
	aggPool map[protodec.L7Protocol]protodec.AggPool,
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
	if pipe.detecTimes < maxDetec || pipe.Decoder != nil {
		dataLi = pipe.sort.Queue(data)
	} else {
		if netTrace.protocolFilter != nil {
			netTrace.protocolFilter.filter(data.SockPtr)
		}
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
			if proto, dec, ok := netTrace.protoSet.ProtoDetector(d.Payload, d.CaptureSize); ok {
				pipe.Proto = proto
				if _, ok := netTrace.enabledProto[pipe.Proto]; !ok {
					pipe.detecTimes = maxDetec + 1
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
			pipe.Decoder.Decode(txRx, d, ntp.Now().UnixNano(), &netTrace.threadInnerID)
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
					// Maybe the connection was closed before the response was sent.
					if v[i].Cost <= 0 {
						continue
					}
					p.Obs(&data.Conn, v[i])
				}
			}
			if _, ok := netTrace.enabledProto[pipe.Proto]; ok {
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
	trace   *NetTrace
	aggPool map[protodec.L7Protocol]protodec.AggPool

	tags        map[string]string
	k8sInfo     *cli.K8sInfo
	enableTrace bool

	sync.Mutex
}

func (watcher *ConnWatcher) handle(tn int64, uniID CUniID, netdata *comm.NetwrkData) {
	watcher.Lock()
	defer watcher.Unlock()

	watcher.trace.StreamHandle(tn, uniID, netdata, watcher.aggPool)
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
			for uniID, pipe := range watcher.trace.ConnMap {
				if groupTime-pipe.lastTime > int64(time.Minute)*3 {
					delete(watcher.trace.ConnMap, uniID)
					watcher.trace.ConnAndClosedDelCount[0]++
				}
			}

			for uniID, pipe := range watcher.trace.ConnMapClosed {
				if groupTime-pipe.lastTime > int64(time.Minute) {
					delete(watcher.trace.ConnMapClosed, uniID)
					watcher.trace.ConnAndClosedDelCount[1]++
				}
			}

			if watcher.trace.ConnAndClosedDelCount[0] > 160_000 {
				watcher.trace.ConnAndClosedDelCount[0] = 0
				connMap := make(map[CUniID]*FlowPipe)
				for k, v := range watcher.trace.ConnMap {
					connMap[k] = v
				}
			}

			if watcher.trace.ConnAndClosedDelCount[1] > 160_000 {
				watcher.trace.ConnAndClosedDelCount[1] = 0
				connMap := make(map[CUniID]*FlowPipe)
				for k, v := range watcher.trace.ConnMapClosed {
					connMap[k] = v
				}
			}

			watcher.Unlock()
		case <-tickerClean.C:
			watcher.Lock()
			if tracer := watcher.trace; tracer != nil {
				tracer.threadInnerID.Cleanup()
			}
			watcher.Unlock()
		case <-ticker.C:
			watcher.Lock()
			if tracer := watcher.trace; tracer != nil {
				for _, pt := range tracer.ptsPrv {
					setInnerID(pt, &tracer.threadInnerID)
				}
				if err := feedEBPFSpan(inputTracing, point.Tracing, tracer.ptsPrv); err != nil {
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
		trace: &NetTrace{
			ConnMap:        make(map[CUniID]*FlowPipe),
			ConnMapClosed:  make(map[CUniID]*FlowPipe),
			protocolFilter: cfg.protocolFilter,
			enabledProto:   cfg.protos,
			allowESPan:     cfg.enableTrace,
			protoSet:       cfg.protoSet,
		},
		aggPool:     cfg.aggPool,
		tags:        cfg.tags,
		k8sInfo:     cfg.k8sNetInfo,
		enableTrace: cfg.enableTrace,
	}
	go p.start(ctx)
	return p
}

type Tracer struct {
	connWatcher *ConnWatcher

	aggPool map[protodec.L7Protocol]protodec.AggPool

	tags    map[string]string
	k8sInfo *cli.K8sInfo

	processFilter  *sysmonitor.ProcessFilter
	protocolFilter *protoKernelFilter

	selfPid int
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
					if err := feed(inputHTTPFlow, point.Network, pts); err != nil {
						log.Error(err)
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

const (
	//nolint:gosec
	eventsHdrSize = int(unsafe.Sizeof(CNetEvents{})) -
		int(unsafe.Sizeof(CNetEvents{}.payload))
	//nolint:gosec
	commHdrSize = int(unsafe.Sizeof(CNetEventComm{}))
)

func (tracer *Tracer) PerfEventHandle(cpu int, data []byte,
	perfmap *manager.PerfMap, manager *manager.Manager,
) {
	events := (*CNetEvents)(unsafe.Pointer(&data[0])) //nolint:gosec

	eventsNum := int(events.rec.num)

	pos := eventsHdrSize // nolint:gosec
	groupTime := time.Now().UnixNano()

	for i := 0; i < eventsNum; i++ {
		curHdrPos := pos
		eventHdr := *(*CNetEventComm)(unsafe.Pointer(&data[curHdrPos])) //nolint:gosec
		pos += commHdrSize
		curPayloadPos := pos
		pos += int(eventHdr.rec.bytes)

		netdata := getNetwrkData(int(eventHdr.rec.bytes))
		readMeta(&eventHdr, &netdata.Conn)
		if int(eventHdr.rec.bytes) > 0 {
			v := unsafe.Slice((*byte)(unsafe.Pointer(&data[curPayloadPos])), int(eventHdr.rec.bytes)) //nolint:gosec
			netdata.Payload = append(netdata.Payload, v...)
		}

		// pos must be calculated before the filter is run
		pid := int(netdata.Conn.Pid)
		if pid == tracer.selfPid {
			continue
		}

		if tracer.processFilter != nil {
			if v, ok := tracer.processFilter.GetProcInfo(pid); ok {
				if !v.Filtered() {
					continue
				}
				netdata.Conn.ProcessName = v.Name()
				netdata.Conn.ServiceName = v.ServiceName()
			} else {
				tracer.processFilter.AsyncTryAdd(pid)
			}
		}

		netdata.Fn = comm.FnID(eventHdr.meta.func_id)
		netdata.SockPtr = uint64(eventHdr.meta.sk_inf.skptr)
		netdata.CaptureSize = int(eventHdr.rec.bytes)
		netdata.FnCallSize = int(eventHdr.meta.original_size)
		netdata.TCPSeq = uint32(eventHdr.meta.tcp_seq)
		netdata.Thread = [2]int32{int32(eventHdr.meta.tid_utid >> 32), (int32(eventHdr.meta.tid_utid))}
		netdata.TS = uint64(eventHdr.meta.ts)
		netdata.TSTail = uint64(eventHdr.meta.ts_tail)
		netdata.Index = uint64(eventHdr.meta.sk_inf.index)

		tracer.connWatcher.handle(groupTime, CUniID(eventHdr.meta.sk_inf.uni_id), netdata)
	}
}

type connWatcherConfig struct {
	apiTracerConfig
	protoSet       *protodec.ProtoSet
	aggPool        map[protodec.L7Protocol]protodec.AggPool
	protocolFilter *protoKernelFilter
}

type protoKernelFilter struct {
	fn    chan func(uint64)
	keySk chan uint64

	firstRun int64
}

func (f *protoKernelFilter) filter(key uint64) {
	f.keySk <- key
}

func (f *protoKernelFilter) setFn(fn func(uint64)) {
	f.fn <- fn
}

func (f *protoKernelFilter) run(ctx context.Context) {
	if v := atomic.SwapInt64(&f.firstRun, 1); v != 0 {
		return
	}

	var kFilter func(uint64)
	for {
		select {
		case fn := <-f.fn:
			kFilter = fn
		case k := <-f.keySk:
			if kFilter != nil {
				kFilter(k)
			}
		case <-ctx.Done():
			return
		}
	}
}

func newTracer(ctx context.Context, cfg *apiTracerConfig) *Tracer {
	if cfg == nil {
		return nil
	}

	var protos []protodec.L7Protocol
	for k := range cfg.protos {
		protos = append(protos, k)
	}
	if len(protos) == 0 {
		protos = append(protos, protodec.ProtoHTTP)
	}
	pset := protodec.SubProtoSet(protos...)
	aggP := pset.NewProtoAggregators()

	protoFilter := &protoKernelFilter{
		fn:    make(chan func(uint64)),
		keySk: make(chan uint64, 16),
	}
	go protoFilter.run(ctx)

	return &Tracer{
		connWatcher: newConnWatcher(ctx, &connWatcherConfig{
			apiTracerConfig: *cfg,
			aggPool:         aggP,
			protoSet:        pset,
			protocolFilter:  protoFilter,
		}),
		aggPool:        aggP,
		tags:           cfg.tags,
		k8sInfo:        cfg.k8sNetInfo,
		selfPid:        cfg.selfPid,
		processFilter:  cfg.procFilter,
		protocolFilter: protoFilter,
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
		v.KVs = v.KVs.Set(spanid.EBPFSpanType, spanType)
		v.KVs = v.KVs.Set(spanid.ReqSeq, int64(v.Meta.ReqTCPSeq))
		v.KVs = v.KVs.Set(spanid.RespSeq, int64(v.Meta.RespTCPSeq))
		v.KVs = v.KVs.Set(spanid.Direction, v.Direction.String())
		// working for process thread inner tracing
		if v.Direction == comm.DIn {
			v.KVs = v.KVs.Set(spanid.ThrTraceID, v.Meta.InnerID)
		}
		v.KVs = v.KVs.Set(comm.FieldKernelThread, v.Meta.Threads[0][0])
		if v.Meta.Threads[0][1] != 0 {
			v.KVs = v.KVs.Set(comm.FieldUserThread, v.Meta.Threads[0][1])
		}
		v.KVs = v.KVs.Set(comm.FieldKernelTime, int64(v.KTime))

		// app trace info
		if !v.Meta.TraceID.Zero() && !v.Meta.ParentSpanID.Zero() {
			v.KVs = v.KVs.Set(spanid.AppTraceIDL, int64(v.Meta.TraceID.Low))
			v.KVs = v.KVs.Set(spanid.AppTraceIDH, int64(v.Meta.TraceID.High))
			v.KVs = v.KVs.Set(spanid.AppParentIDL, int64(v.Meta.ParentSpanID))
			var aSampled int64
			if v.Meta.SampledSpan {
				aSampled = 1
			} else {
				aSampled = -1
			}
			v.KVs = v.KVs.Set(spanid.AppSpanSampled, aSampled)
			if v.Meta.SpanHexEnc {
				v.KVs = v.KVs.Set("app_trace_id", v.Meta.TraceID.StringHex())
				v.KVs = v.KVs.Set("app_parent_id", v.Meta.ParentSpanID.StringHex())
			} else {
				v.KVs = v.KVs.Set("app_trace_id", v.Meta.TraceID.StringDec())
				v.KVs = v.KVs.Set("app_parent_id", v.Meta.ParentSpanID.StringDec())
			}
		}

		// service info
		v.KVs = v.KVs.Set("source_type", "ebpf")
		v.KVs = v.KVs.Set("process_name", conn.ProcessName)
		v.KVs = v.KVs.Set("thread_name", conn.TaskName)
		if conn.ServiceName == "" {
			if conn.ProcessName != "" {
				v.KVs = v.KVs.Set("service", conn.ProcessName)
			} else {
				v.KVs = v.KVs.Set("service", conn.TaskName)
			}
		} else {
			v.KVs = v.KVs.Set("service", conn.ServiceName)
		}
		v.KVs = v.KVs.Set(comm.FieldPid, strconv.Itoa(int(conn.Pid)))

		// conn info
		isV6 := !netflow.ConnAddrIsIPv4(conn.Meta)
		ip := netflow.U32BEToIP(conn.Daddr, isV6)
		v.KVs = v.KVs.Set("dst_ip", ip.String())
		ip = netflow.U32BEToIP(conn.Saddr, isV6)
		v.KVs = v.KVs.Set("src_ip", ip.String())
		v.KVs = v.KVs.Set("src_port", strconv.Itoa(int(conn.Sport)))
		v.KVs = v.KVs.Set("dst_port", strconv.Itoa(int(conn.Dport)))

		// span info
		v.KVs = v.KVs.Set("start", v.Time/1000)        // conv ns to us
		v.KVs = v.KVs.Set("duration", v.Duration/1000) // conv ns to us
		// v.KVs = v.KVs.Add("cost", v.Cost, false, true)
		v.KVs = v.KVs.Set("span_type", spanType)

		opt := point.CommonLoggingOptions()
		opt = append(opt, point.WithTimestamp(v.Time))
		pt := point.NewPoint("dketrace", v.KVs, opt...)
		pts = append(pts, pt)
	}
	return pts
}
