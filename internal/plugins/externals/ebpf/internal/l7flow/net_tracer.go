//go:build linux
// +build linux

package l7flow

import (
	"context"
	"math"
	"strconv"
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

type ProcNetworkTrace struct {
	Pid int

	// 对于该进程， ingress 请求抵达时，后续的网络请求都应该继承此生成的 innter trace id
	ESpanLinkDuration time.Duration

	// (kernel) sock ptr and random id ==> network flow pipe
	ConnMap         map[comm.ConnUniID]*FlowPipe
	delConnMapCount int

	threadInnerID comm.ThreadTrace

	ptsPrv []*point.Point
	ptsCur []*point.Point

	allowESPan bool
}

func (netTrace *ProcNetworkTrace) Handle(data *comm.NetwrkData,
	aggPool map[protodec.L7Protocol]protodec.AggPool,
) {
	if netTrace.ConnMap == nil {
		netTrace.ConnMap = make(map[comm.ConnUniID]*FlowPipe)
	}

	ptr := data.ConnUniID

	pipe, ok := netTrace.ConnMap[ptr]
	if !ok {
		pipe = &FlowPipe{
			Conn: data.Conn,
			sort: netdata{prvDataPos: 0},
		}
		netTrace.ConnMap[ptr] = pipe
	}

	dataLi := pipe.sort.Queue(data)
	defer func(datli []*comm.NetwrkData) {
		for _, v := range dataLi {
			putNetwrkData(v)
		}
	}(dataLi)

	var connClose bool
	for _, d := range dataLi {
		txRx := comm.FnInOut(d.Fn)
		if txRx == comm.NICDUnknown {
			if d.Fn == comm.FnSysClose {
				connClose = true
			}
			continue
		}

		if pipe.Decoder == nil {
			if proto, dec, ok := protodec.ProtoDetector(d.Payload); ok {
				pipe.Decoder = dec
				pipe.Proto = proto
				if proto == protodec.ProtoHTTP2 {
					continue
				}
			} else {
				pipe.detecTimes++
				continue
			}
		}

		if d.ActSize > 0 {
			pipe.Decoder.Decode(txRx, d, time.Now().UnixNano(), &netTrace.threadInnerID)
		}
	}

	if connClose {
		pipe.connClosed = true
		if pipe.Decoder != nil {
			pipe.Decoder.ConnClose()
		}
		delete(netTrace.ConnMap, ptr)
		netTrace.delConnMapCount++
		if netTrace.delConnMapCount >= 1e3 {
			mp := make(map[comm.ConnUniID]*FlowPipe)
			for k, v := range netTrace.ConnMap {
				mp[k] = v
			}
			netTrace.ConnMap = mp
			netTrace.delConnMapCount = 0
		}
	}

	if pipe.Decoder != nil {
		if v := pipe.Decoder.Export(connClose); len(v) > 0 {
			if p := aggPool[pipe.Proto]; p != nil {
				for i := 0; i < len(v); i++ {
					p.Obs(&data.Conn, v[i])
				}
			}
			if netTrace.allowESPan && pipe.Conn.ProcessName != "datakit" &&
				pipe.Conn.ProcessName != "datakit-ebpf" {
				pts := genPts(v, &data.Conn)
				netTrace.ptsCur = append(netTrace.ptsCur, pts...)
			}
		}
	}
}

type FlowPipe struct {
	Conn       comm.ConnectionInfo
	Decoder    protodec.ProtoDecPipe
	Proto      protodec.L7Protocol
	detecTimes int

	sort netdata

	connClosed bool
}

type PidMap struct {
	pidMap  map[int]*ProcNetworkTrace
	ch      chan *comm.NetwrkData
	aggPool map[protodec.L7Protocol]protodec.AggPool

	procFilter *tracing.ProcessFilter

	tags          map[string]string
	k8sInfo       *k8sinfo.K8sNetInfo
	metricPostURL string
	tracePostURL  string
	enableTrace   bool
}

func (p *PidMap) start(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	tickerClean := time.NewTicker(time.Minute * 5)

	for {
		select {
		case netdata := <-p.ch:
			pid := int(netdata.Conn.Pid)
			var nettracer *ProcNetworkTrace
			if v, ok := p.pidMap[pid]; !ok {
				if p.pidMap == nil {
					p.pidMap = make(map[int]*ProcNetworkTrace)
				}
				nettracer = &ProcNetworkTrace{
					Pid:     pid,
					ConnMap: make(map[comm.ConnUniID]*FlowPipe),
				}
				p.pidMap[pid] = nettracer
			} else {
				nettracer = v
			}
			if p.enableTrace && p.procFilter != nil {
				nettracer.allowESPan = false
				if v, ok := p.procFilter.GetProcInfo(pid); ok {
					nettracer.allowESPan = v.AllowTrace
					netdata.Conn.ProcessName = v.Name
					netdata.Conn.ServiceName = v.Service
				}
			}

			nettracer.Handle(netdata, p.aggPool)

		case <-tickerClean.C:
			for _, v := range p.pidMap {
				v.threadInnerID.Cleanup()
			}
		case <-ticker.C:
			for _, v := range p.pidMap {
				for _, pt := range v.ptsPrv {
					setInnerID(pt, &v.threadInnerID)
				}
				if err := feed(p.tracePostURL, v.ptsPrv, false); err != nil {
					log.Error(err)
				}
				// feed("http://0.0.0.0:9529/v1/write/logging?input=ebpf-net%2Fespan", v.ptsPrv, false)
				v.ptsPrv = v.ptsCur
				v.ptsCur = nil
			}

		case <-ctx.Done():
			return
		}
	}
}

func (p *PidMap) Handle(v *comm.NetwrkData) {
	p.ch <- v
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
	id := threadInnerID.GetInnerID(tid, ktime)
	pt.Add(spanid.ThrTraceID, id)
}

func newPidMap(ctx context.Context, aggPool map[protodec.L7Protocol]protodec.AggPool, procFilter *tracing.ProcessFilter,
	tags map[string]string, k8sInfo *k8sinfo.K8sNetInfo, metricPostURL string, tracePostURL string, enableTrace bool,
) *PidMap {
	p := &PidMap{
		pidMap:        make(map[int]*ProcNetworkTrace),
		ch:            make(chan *comm.NetwrkData, 32),
		aggPool:       aggPool,
		procFilter:    procFilter,
		tags:          tags,
		k8sInfo:       k8sInfo,
		metricPostURL: metricPostURL,
		tracePostURL:  tracePostURL,
		enableTrace:   enableTrace,
	}
	go p.start(ctx)
	return p
}

type Tracer struct {
	pidMap [5]*PidMap

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
	bufferC := (*CL7Buffer)(unsafe.Pointer(&data[0])) //nolint:gosec

	netdata := getNetwrkData()

	readMeta(bufferC, &netdata.Conn)

	actLen := int(bufferC.meta.act_size)
	bufLen := int(bufferC.meta.buf_len)

	if bufLen > 0 {
		if bufLen > PayloadBufSize {
			bufLen = PayloadBufSize
		}
		if actLen > 0 {
			b := *(*[PayloadBufSize]byte)(unsafe.Pointer(&bufferC.payload)) //nolint:gosec
			netdata.Payload = append(netdata.Payload, b[:bufLen]...)
		}
	}

	pid := int(netdata.Conn.Pid)
	if pid == selfPid {
		return
	}

	netdata.ConnUniID.Ktime = uint32(bufferC.meta.uni_id.ktime)
	netdata.ConnUniID.SkPtr = uint64(bufferC.meta.uni_id.sk)
	netdata.ConnUniID.Rand = uint32(bufferC.meta.uni_id.prandom)

	netdata.Fn = comm.FnID(bufferC.meta.func_id)

	netdata.ActSize = actLen
	netdata.TCPSeq = uint32(bufferC.meta.tcp_seq)
	netdata.Thread = [2]int32{int32(bufferC.meta.tid_utid >> 32), (int32(bufferC.meta.tid_utid))}
	netdata.TS = uint64(bufferC.meta.ts)
	netdata.TSTail = uint64(bufferC.meta.ts_tail)
	netdata.Index = uint32(bufferC.meta.index)

	// log.Info(netdata.String())

	tracer.pidMap[pid%5].Handle(netdata)
}

func newTracer(ctx context.Context, procFilter *tracing.ProcessFilter, tags map[string]string,
	k8sInfo *k8sinfo.K8sNetInfo, metricPostURL string, tracePostURL string, enableTrace bool,
) *Tracer {
	mps := [5]*PidMap{}

	aggP := protodec.NewProtoAggregators()
	for i := 0; i < 5; i++ {
		mps[i] = newPidMap(ctx, aggP, procFilter, tags, k8sInfo, metricPostURL, tracePostURL, enableTrace)
	}

	return &Tracer{
		pidMap:        mps,
		aggPool:       aggP,
		tags:          tags,
		k8sInfo:       k8sInfo,
		metricPostURL: metricPostURL,
		tracePostURL:  tracePostURL,
		enableTrace:   enableTrace,
	}
}

type netdata struct {
	li []*comm.NetwrkData
	// 从 1 开始索引，如果值为 0，视为发生翻转
	prvDataPos uint32
}

func (n *netdata) Queue(data *comm.NetwrkData) []*comm.NetwrkData {
	var val []*comm.NetwrkData
	if data == nil {
		return val
	}

	lenQ := len(n.li)
	switch lenQ {
	case 0:
		if n.prvDataPos+1 == data.Index {
			n.prvDataPos = data.Index
			return []*comm.NetwrkData{data}
		} else {
			n.li = append(n.li, data)
		}
	default:
		for i := 0; i < lenQ; i++ {
			if idxLess(data.Index, n.li[i].Index) {
				n.li = append(n.li, nil)
				copy(n.li[i+1:], n.li[i:])
				n.li[i] = data
				break
			}
			if i+1 == lenQ {
				n.li = append(n.li, data)
				break
			}
		}
	}

	// try flush cache
	i := 0
	for ; i < len(n.li); i++ {
		cur := n.li[i].Index
		if cur == n.prvDataPos+1 {
			val = append(val, n.li[i])
			n.prvDataPos = cur
		} else {
			break
		}
	}

	// clean cache
	if i >= len(n.li) {
		n.li = n.li[:0]
	} else if i > 0 {
		copy(n.li, n.li[i:])
		n.li = n.li[:len(n.li)-i]
	}

	// 可能存在数据丢失情况
	if len(n.li) >= 1024 && len(val) == 0 {
		x := 128
		val = append(val, n.li[:x]...)
		n.li = n.li[x:]
		n.prvDataPos = val[x-1].Index
	}

	return val
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
		v.KVs = v.KVs.Add("pid", strconv.Itoa(int(conn.Pid)), false, true)

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
		v.KVs = v.KVs.Add("cost", v.Cost, false, true)
		v.KVs = v.KVs.Add("span_type", spanType, false, true)

		opt := point.CommonLoggingOptions()
		opt = append(opt, point.WithTimestamp(v.Time))
		pt := point.NewPointV2("dketrace", v.KVs, opt...)
		pts = append(pts, pt)
	}
	return pts
}

func idxLess(l, r uint32) bool {
	// 可能发生回绕现象，预留窗口应与 buffer 长度相近
	if l > math.MaxUint32-1025 && r <= 1025 {
		return true
	}

	return l < r
}
