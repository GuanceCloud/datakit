//go:build linux && cgo
// +build linux,cgo

package l7flow

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/GuanceCloud/cliutils/point"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/protodec"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/netflow"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/procwatch"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/cli"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/spanid"
)

type NetTrace struct {
	// 对于每个服务，ingress 请求抵达时，后续的网络请求都应该继承此生成的 innter trace id
	ESpanLinkDuration time.Duration

	// (kernel) sock ptr and random id ==> network flow pipe
	connShards [connMapShardCount]connMapShard

	threadInnerID  comm.ThreadTrace
	protocolFilter *protoKernelFilter
	enabledProto   map[protodec.L7Protocol]struct{}
	protoSet       *protodec.ProtoSet

	ptsPrv []*point.Point
	ptsCur []*point.Point

	allowESPan bool
}

const (
	maxDetec                    = 64
	connMapCompactThreshold     = 160_000
	connMapShardCount           = 32
	connWatcherDefaultQueueSize = 256
	connWatcherDefaultWorkers   = 4
	connWatcherWorkersEnv       = "DK_EBPF_L7FLOW_ASYNC_WORKERS"
	connWatcherQueueSizeEnv     = "DK_EBPF_L7FLOW_ASYNC_QUEUE_SIZE"
)

type connWatcherTask struct {
	tn      int64
	uniID   CUniID
	netdata *comm.NetwrkData
}

func connWatcherWorkerCount() int {
	workers := connWatcherDefaultWorkers
	if maxProcs := runtime.GOMAXPROCS(0); maxProcs > 0 && workers > maxProcs {
		workers = maxProcs
	}
	if raw := os.Getenv(connWatcherWorkersEnv); raw != "" {
		switch v, err := strconv.Atoi(raw); {
		case err != nil:
			log.Warnf("invalid %s=%q: %v", connWatcherWorkersEnv, raw, err)
		case v <= 0:
			log.Warnf("invalid %s=%q: must be > 0", connWatcherWorkersEnv, raw)
		default:
			workers = v
		}
	}
	if workers > connMapShardCount {
		workers = connMapShardCount
	}
	if workers <= 0 {
		workers = 1
	}
	return workers
}

func connWatcherQueueSize() int {
	size := connWatcherDefaultQueueSize
	if raw := os.Getenv(connWatcherQueueSizeEnv); raw != "" {
		switch v, err := strconv.Atoi(raw); {
		case err != nil:
			log.Warnf("invalid %s=%q: %v", connWatcherQueueSizeEnv, raw, err)
		case v <= 0:
			log.Warnf("invalid %s=%q: must be > 0", connWatcherQueueSizeEnv, raw)
		default:
			size = v
		}
	}
	return size
}

type connMapShard struct {
	mu     sync.Mutex
	open   map[CUniID]*FlowPipe
	closed map[CUniID]*FlowPipe

	delCount [2]int
}

func (s *connMapShard) ensureMaps() {
	if s.open == nil {
		s.open = make(map[CUniID]*FlowPipe)
	}
	if s.closed == nil {
		s.closed = make(map[CUniID]*FlowPipe)
	}
}

func (s *connMapShard) maybeCompact() {
	if s.delCount[0] > connMapCompactThreshold {
		s.delCount[0] = 0
		s.open = cloneFlowPipeMap(s.open)
	}

	if s.delCount[1] > connMapCompactThreshold {
		s.delCount[1] = 0
		s.closed = cloneFlowPipeMap(s.closed)
	}
}

func connMapShardIndex(uniID CUniID) int {
	//nolint:gosec
	b := unsafe.Slice((*byte)(unsafe.Pointer(&uniID)), unsafe.Sizeof(uniID))
	var h uint32 = 2166136261
	for _, v := range b {
		h ^= uint32(v)
		h *= 16777619
	}
	return int(h % connMapShardCount)
}

func (netTrace *NetTrace) shardFor(uniID CUniID) *connMapShard {
	return &netTrace.connShards[connMapShardIndex(uniID)]
}

func (netTrace *NetTrace) StreamHandle(tn int64, uniID CUniID, data *comm.NetwrkData,
) *streamHandleResult {
	pipe, inClosedMap := netTrace.getOrCreatePipe(tn, uniID, data)
	if pipe == nil {
		return nil
	}

	result, connClose := netTrace.processPipe(tn, pipe, data)
	if connClose {
		netTrace.finalizePipeClose(uniID, pipe, inClosedMap)
	}
	return result
}

func (netTrace *NetTrace) getOrCreatePipe(tn int64, uniID CUniID, data *comm.NetwrkData) (*FlowPipe, bool) {
	shard := netTrace.shardFor(uniID)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	shard.ensureMaps()

	var pipe *FlowPipe
	inClosedMap := false

	if p, ok := shard.closed[uniID]; ok {
		pipe = p
		inClosedMap = true
	}

	if pipe == nil {
		var ok bool
		pipe, ok = shard.open[uniID]
		if !ok {
			pipe = &FlowPipe{
				Conn: data.Conn,
				sort: dataQueue{prvDataPos: 0},
			}
			shard.open[uniID] = pipe
		}
	}

	atomic.StoreInt64(&pipe.lastTime, tn)
	if data != nil && data.Fn == comm.FnSysClose {
		delete(shard.open, uniID)
		shard.closed[uniID] = pipe
		inClosedMap = true
	}

	return pipe, inClosedMap
}

func (netTrace *NetTrace) finalizePipeClose(uniID CUniID, pipe *FlowPipe, inClosedMap bool) {
	if pipe == nil {
		return
	}
	shard := netTrace.shardFor(uniID)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if inClosedMap {
		if cur, ok := shard.closed[uniID]; ok && cur == pipe {
			shard.delCount[1]++
			delete(shard.closed, uniID)
		}
		return
	}

	if cur, ok := shard.open[uniID]; ok && cur == pipe {
		shard.delCount[0]++
		delete(shard.open, uniID)
	}
}

func (netTrace *NetTrace) processPipe(tn int64, pipe *FlowPipe, data *comm.NetwrkData) (*streamHandleResult, bool) {
	if pipe == nil {
		return nil, false
	}

	pipe.mu.Lock()
	defer pipe.mu.Unlock()

	if data != nil && data.Fn == comm.FnSysClose {
		pipe.connClosed = true
	}

	var dataLi []*comm.NetwrkData
	if pipe.detecTimes < maxDetec || pipe.Decoder != nil {
		dataLi = pipe.sort.Queue(data)
	} else {
		if netTrace.protocolFilter != nil && !pipe.protocolFilterQueued {
			pipe.protocolFilterQueued = netTrace.protocolFilter.tryFilter(data.SockPtr)
		}
		pipe.sort.li = nil
	}

	defer func(li []*comm.NetwrkData) {
		for _, d := range li {
			putNetwrkData(d)
		}
	}(dataLi)

	var connClose bool
	nowNS := tn
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
			pipe.Decoder.Decode(txRx, d, nowNS, &netTrace.threadInnerID)
		}
	}

	if connClose {
		pipe.connClosed = true
		if pipe.Decoder != nil {
			pipe.Decoder.ConnClose()
		}
	}

	if pipe.Decoder != nil {
		if v := pipe.Decoder.Export(connClose); len(v) > 0 {
			_, enabledProto := netTrace.enabledProto[pipe.Proto]
			return &streamHandleResult{
				proto:     pipe.Proto,
				conn:      data.Conn,
				protoData: v,
				appendESpan: enabledProto && netTrace.allowESPan &&
					pipe.Conn.ProcessName != "datakit" &&
					pipe.Conn.ProcessName != "datakit-ebpf",
			}, connClose
		}
	}

	return nil, connClose
}

type streamHandleResult struct {
	proto       protodec.L7Protocol
	conn        comm.ConnectionInfo
	protoData   []*protodec.ProtoData
	appendESpan bool
}

type FlowPipe struct {
	mu sync.Mutex

	Conn       comm.ConnectionInfo
	Decoder    protodec.ProtoDecPipe
	Proto      protodec.L7Protocol
	detecTimes int

	sort dataQueue

	lastTime int64

	connClosed bool

	protocolFilterQueued bool
}

func cloneFlowPipeMap(src map[CUniID]*FlowPipe) map[CUniID]*FlowPipe {
	dst := make(map[CUniID]*FlowPipe, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (netTrace *NetTrace) sweepExpiredConnMaps(groupTime int64) (int, int) {
	if netTrace == nil {
		return 0, 0
	}

	openTotal := 0
	closedTotal := 0
	for i := range netTrace.connShards {
		shard := &netTrace.connShards[i]
		shard.mu.Lock()
		shard.ensureMaps()
		for uniID, pipe := range shard.open {
			if groupTime-atomic.LoadInt64(&pipe.lastTime) > int64(time.Minute)*3 {
				delete(shard.open, uniID)
				shard.delCount[0]++
			}
		}

		for uniID, pipe := range shard.closed {
			if groupTime-atomic.LoadInt64(&pipe.lastTime) > int64(time.Minute) {
				delete(shard.closed, uniID)
				shard.delCount[1]++
			}
		}

		shard.maybeCompact()
		openTotal += len(shard.open)
		closedTotal += len(shard.closed)
		shard.mu.Unlock()
	}
	return openTotal, closedTotal
}

type ConnWatcher struct {
	trace   *NetTrace
	aggPool map[protodec.L7Protocol]protodec.AggPool

	tags        map[string]string
	k8sInfo     *cli.K8sInfo
	enableTrace bool

	eventQueues []chan connWatcherTask
	ptsMu       sync.Mutex
}

func (watcher *ConnWatcher) handle(tn int64, uniID CUniID, netdata *comm.NetwrkData) {
	task := connWatcherTask{
		tn:      tn,
		uniID:   uniID,
		netdata: netdata,
	}
	if len(watcher.eventQueues) > 0 {
		q := watcher.eventQueues[connMapShardIndex(uniID)%len(watcher.eventQueues)]
		waitStart := time.Now()
		q <- task
		exporter.ObserveAsyncQueueWait("l7flow", time.Since(waitStart))
		watcher.observeAsyncQueueDepth()
		return
	}

	watcher.processTask(task)
}

func (watcher *ConnWatcher) processTask(task connWatcherTask) {
	start := time.Now()
	defer exporter.ObserveAsyncProcess("l7flow", time.Since(start))

	result := watcher.trace.StreamHandle(task.tn, task.uniID, task.netdata)
	if result == nil {
		return
	}

	if p := watcher.aggPool[result.proto]; p != nil {
		for i := 0; i < len(result.protoData); i++ {
			// Maybe the connection was closed before the response was sent.
			if result.protoData[i].Cost <= 0 {
				continue
			}
			p.Obs(&result.conn, result.protoData[i])
		}
	}

	if result.appendESpan {
		if pts := genPts(result.protoData, &result.conn); len(pts) > 0 {
			watcher.appendTracePoints(pts)
		}
	}
}

func (watcher *ConnWatcher) startEventWorkers(ctx context.Context) {
	workers := connWatcherWorkerCount()
	queueSize := connWatcherQueueSize()
	watcher.eventQueues = make([]chan connWatcherTask, 0, workers)
	log.Infof("l7flow async workers=%d queue_size=%d conn_shards=%d", workers, queueSize, connMapShardCount)

	for i := 0; i < workers; i++ {
		q := make(chan connWatcherTask, queueSize)
		watcher.eventQueues = append(watcher.eventQueues, q)

		go func(queue <-chan connWatcherTask) {
			for {
				select {
				case task := <-queue:
					watcher.processTask(task)
					watcher.observeAsyncQueueDepth()
				case <-ctx.Done():
					return
				}
			}
		}(q)
	}
}

func (watcher *ConnWatcher) observeAsyncQueueDepth() {
	if len(watcher.eventQueues) == 0 {
		return
	}

	total := 0
	max := 0
	for _, q := range watcher.eventQueues {
		if q == nil {
			continue
		}
		n := len(q)
		total += n
		if n > max {
			max = n
		}
	}
	exporter.ObserveCacheEntries("l7flow", "async_queue_total", total)
	exporter.ObserveCacheEntries("l7flow", "async_queue_max", max)
}

func (watcher *ConnWatcher) appendTracePoints(pts []*point.Point) {
	if len(pts) == 0 {
		return
	}

	watcher.ptsMu.Lock()
	defer watcher.ptsMu.Unlock()

	if watcher.trace != nil {
		watcher.trace.ptsCur = append(watcher.trace.ptsCur, pts...)
	}
}

func (watcher *ConnWatcher) rotateTracePoints() ([]*point.Point, *comm.ThreadTrace, int, int) {
	watcher.ptsMu.Lock()
	defer watcher.ptsMu.Unlock()

	if watcher.trace == nil {
		return nil, nil, 0, 0
	}

	tracer := watcher.trace
	prevLen := len(tracer.ptsPrv)
	currLen := len(tracer.ptsCur)
	pts := tracer.ptsPrv
	tracer.ptsPrv = tracer.ptsCur
	tracer.ptsCur = nil
	return pts, &tracer.threadInnerID, prevLen, currLen
}

func (watcher *ConnWatcher) cleanupThreadTrace() {
	if watcher.trace != nil {
		watcher.trace.threadInnerID.Cleanup()
	}
}

func (watcher *ConnWatcher) start(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 10)
	tickerClean := time.NewTicker(time.Minute * 5)
	tickerCheck := time.NewTicker(time.Minute * 2)
	defer ticker.Stop()
	defer tickerClean.Stop()
	defer tickerCheck.Stop()

	for {
		select {
		case <-tickerCheck.C:
			groupTime := time.Now().UnixNano()
			openLen, closedLen := watcher.trace.sweepExpiredConnMaps(groupTime)
			exporter.ObserveCacheEntries("l7flow", "conn_map", openLen)
			exporter.ObserveCacheEntries("l7flow", "conn_closed_map", closedLen)
		case <-tickerClean.C:
			watcher.cleanupThreadTrace()
		case <-ticker.C:
			pts, threadInnerID, prevLen, currLen := watcher.rotateTracePoints()
			exporter.ObserveCacheEntries("l7flow", "span_prev", prevLen)
			exporter.ObserveCacheEntries("l7flow", "span_curr", currLen)
			if threadInnerID != nil {
				for _, pt := range pts {
					setInnerID(pt, threadInnerID)
				}
			}
			if err := feedEBPFSpan(inputTracing, point.Tracing, pts); err != nil {
				log.Error(err)
			}
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
	p.startEventWorkers(ctx)
	go p.start(ctx)
	return p
}

func appendTraceKV(kvs point.KVs, key string, val any) point.KVs {
	return append(kvs, point.NewKV(key, val))
}

type Tracer struct {
	connWatcher *ConnWatcher

	aggPool map[protodec.L7Protocol]protodec.AggPool

	tags    map[string]string
	k8sInfo *cli.K8sInfo

	catalog        *procwatch.Catalog
	protocolFilter *protoKernelFilter

	selfPid int

	debugPerfEvents int64
}

func (tracer *Tracer) populateProcessInfo(conn *comm.ConnectionInfo) bool {
	if tracer == nil || tracer.catalog == nil || conn == nil {
		return true
	}

	info, ok := tracer.catalog.LookupOrResolve(int(conn.Pid))
	if !ok || info == nil {
		return true
	}
	if !info.Collectable() {
		return false
	}

	conn.ProcessName = info.Name()
	conn.ServiceName = info.ServiceName()
	return true
}

func (tracer *Tracer) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			for _, p := range tracer.aggPool {
				component := "l7flow_" + p.Proto().StringLower()
				if ag, ok := p.(interface{ Len() int }); ok {
					exporter.ObserveAggEntries(component, ag.Len())
				}
				flushStart := time.Now()
				pts := p.Export(tracer.tags, tracer.k8sInfo)
				p.Cleanup()
				exporter.ObserveAggEntries(component, 0)
				if len(pts) > 0 {
					if err := feed(inputHTTPFlow, point.Network, pts); err != nil {
						log.Error(err)
						exporter.ObserveAggFlush(component, len(pts), time.Since(flushStart), "error")
					} else {
						exporter.ObserveAggFlush(component, len(pts), time.Since(flushStart), "ok")
					}
				} else {
					exporter.ObserveAggFlush(component, 0, time.Since(flushStart), "ok")
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
	stream *bpfutil.PerfStream, runtime *bpfutil.Runtime,
) {
	if len(data) < eventsHdrSize {
		log.Debugf("drop short l7 batch event: got %d want >= %d", len(data), eventsHdrSize)
		return
	}

	events := (*CNetEvents)(unsafe.Pointer(&data[0])) //nolint:gosec

	eventsNum := int(events.rec.num)

	pos := eventsHdrSize // nolint:gosec
	groupTime := time.Now().UnixNano()

	for i := 0; i < eventsNum; i++ {
		curHdrPos := pos
		if curHdrPos+commHdrSize > len(data) {
			log.Debugf("drop truncated l7 event header: pos=%d hdr=%d len=%d", curHdrPos, commHdrSize, len(data))
			return
		}
		eventHdr := *(*CNetEventComm)(unsafe.Pointer(&data[curHdrPos])) //nolint:gosec
		pos += commHdrSize
		curPayloadPos := pos
		payloadLen := int(eventHdr.rec.bytes)
		if payloadLen < 0 || curPayloadPos+payloadLen > len(data) {
			log.Debugf("drop truncated l7 payload: pos=%d bytes=%d len=%d", curPayloadPos, payloadLen, len(data))
			return
		}
		pos += payloadLen

		netdata := getNetwrkData(payloadLen)
		readMeta(&eventHdr, &netdata.Conn)
		if payloadLen > 0 {
			v := unsafe.Slice((*byte)(unsafe.Pointer(&data[curPayloadPos])), payloadLen) //nolint:gosec
			netdata.Payload = append(netdata.Payload, v...)
		}

		// pos must be calculated before the filter is run
		pid := int(netdata.Conn.Pid)
		if pid == tracer.selfPid {
			continue
		}

		if !tracer.populateProcessInfo(&netdata.Conn) {
			continue
		}

		netdata.Fn = comm.FnID(eventHdr.meta.func_id)
		netdata.SockPtr = uint64(eventHdr.meta.sk_inf.skptr)
		netdata.CaptureSize = payloadLen
		netdata.FnCallSize = int(eventHdr.meta.original_size)
		netdata.TCPSeq = uint32(eventHdr.meta.tcp_seq)
		netdata.Thread = [2]int32{int32(eventHdr.meta.tid_utid >> 32), (int32(eventHdr.meta.tid_utid))}
		netdata.TS = uint64(eventHdr.meta.ts)
		netdata.TSTail = uint64(eventHdr.meta.ts_tail)
		netdata.Index = uint64(eventHdr.meta.sk_inf.index)

		if n := atomic.AddInt64(&tracer.debugPerfEvents, 1); n <= 16 {
			snippet := netdata.Payload
			if len(snippet) > 64 {
				snippet = snippet[:64]
			}
			log.Debugf("l7 perf event: fn=%s capture=%d pid=%d conn=%s payload=%q",
				netdata.Fn.String(), netdata.CaptureSize, netdata.Conn.Pid,
				netdata.Conn.String(), string(snippet))
		}

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

func (f *protoKernelFilter) tryFilter(key uint64) bool {
	select {
	case f.keySk <- key:
		return true
	default:
		return false
	}
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
		keySk: make(chan uint64, 256),
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
		catalog:        cfg.catalog,
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
		v.KVs = appendTraceKV(v.KVs, spanid.EBPFSpanType, spanType)
		v.KVs = appendTraceKV(v.KVs, spanid.ReqSeq, int64(v.Meta.ReqTCPSeq))
		v.KVs = appendTraceKV(v.KVs, spanid.RespSeq, int64(v.Meta.RespTCPSeq))
		v.KVs = appendTraceKV(v.KVs, spanid.Direction, v.Direction.String())
		// working for process thread inner tracing
		if v.Direction == comm.DIn {
			v.KVs = appendTraceKV(v.KVs, spanid.ThrTraceID, v.Meta.InnerID)
		}
		v.KVs = appendTraceKV(v.KVs, comm.FieldKernelThread, v.Meta.Threads[0][0])
		if v.Meta.Threads[0][1] != 0 {
			v.KVs = appendTraceKV(v.KVs, comm.FieldUserThread, v.Meta.Threads[0][1])
		}
		v.KVs = appendTraceKV(v.KVs, comm.FieldKernelTime, int64(v.KTime))

		// app trace info
		if !v.Meta.TraceID.Zero() && !v.Meta.ParentSpanID.Zero() {
			v.KVs = appendTraceKV(v.KVs, spanid.AppTraceIDL, int64(v.Meta.TraceID.Low))
			v.KVs = appendTraceKV(v.KVs, spanid.AppTraceIDH, int64(v.Meta.TraceID.High))
			v.KVs = appendTraceKV(v.KVs, spanid.AppParentIDL, int64(v.Meta.ParentSpanID))
			var aSampled int64
			if v.Meta.SampledSpan {
				aSampled = 1
			} else {
				aSampled = -1
			}
			v.KVs = appendTraceKV(v.KVs, spanid.AppSpanSampled, aSampled)
			if v.Meta.SpanHexEnc {
				v.KVs = appendTraceKV(v.KVs, "app_trace_id", v.Meta.TraceID.StringHex())
				v.KVs = appendTraceKV(v.KVs, "app_parent_id", v.Meta.ParentSpanID.StringHex())
			} else {
				v.KVs = appendTraceKV(v.KVs, "app_trace_id", v.Meta.TraceID.StringDec())
				v.KVs = appendTraceKV(v.KVs, "app_parent_id", v.Meta.ParentSpanID.StringDec())
			}
		}

		// service info
		v.KVs = appendTraceKV(v.KVs, "source_type", "ebpf")
		v.KVs = appendTraceKV(v.KVs, "process_name", conn.ProcessName)
		v.KVs = appendTraceKV(v.KVs, "thread_name", conn.TaskName)
		if conn.ServiceName == "" {
			if conn.ProcessName != "" {
				v.KVs = appendTraceKV(v.KVs, "service", conn.ProcessName)
			} else {
				v.KVs = appendTraceKV(v.KVs, "service", conn.TaskName)
			}
		} else {
			v.KVs = appendTraceKV(v.KVs, "service", conn.ServiceName)
		}
		v.KVs = appendTraceKV(v.KVs, comm.FieldPid, strconv.Itoa(int(conn.Pid)))

		// conn info
		isV6 := !netflow.ConnAddrIsIPv4(conn.Meta)
		ip := netflow.U32BEToIP(conn.Daddr, isV6)
		v.KVs = appendTraceKV(v.KVs, "dst_ip", ip.String())
		ip = netflow.U32BEToIP(conn.Saddr, isV6)
		v.KVs = appendTraceKV(v.KVs, "src_ip", ip.String())
		v.KVs = appendTraceKV(v.KVs, "src_port", strconv.Itoa(int(conn.Sport)))
		v.KVs = appendTraceKV(v.KVs, "dst_port", strconv.Itoa(int(conn.Dport)))

		// span info
		v.KVs = appendTraceKV(v.KVs, "start", v.Time/1000)        // conv ns to us
		v.KVs = appendTraceKV(v.KVs, "duration", v.Duration/1000) // conv ns to us
		// v.KVs = v.KVs.Add("cost", v.Cost, false, true)
		v.KVs = appendTraceKV(v.KVs, "span_type", spanType)

		opt := point.CommonLoggingOptions()
		opt = append(opt, point.WithTimestamp(v.Time))
		pt := point.NewPoint("dketrace", v.KVs, opt...)
		pts = append(pts, pt)
	}
	return pts
}
