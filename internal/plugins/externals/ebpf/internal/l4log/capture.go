//go:build linux
// +build linux

// Package l4log capture packets
package l4log

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/google/gopacket"
	"github.com/google/gopacket/afpacket"
	"github.com/google/gopacket/layers"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/k8sinfo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/output"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/spanid"
	"golang.org/x/sys/unix"
)

type NetProtoTyp string

const (
	NetProtoTCP  NetProtoTyp = "tcp"
	NetProtoHTTP NetProtoTyp = "http"
)

var k8sNetInfo *k8sinfo.K8sNetInfo

func SetK8sNetInfo(n *k8sinfo.K8sNetInfo) {
	k8sNetInfo = n
}

type PMeta struct {
	reuseidx uint64

	SrcIP string
	DstIP string

	SrcPort uint16
	DstPort uint16

	VNIID uint32 // vni id
	VXLAN bool   // vxlan
}

type PktTCPHdr struct {
	TXRX string `json:"txrx"`

	// tcp flags
	Flags TCPFlag `json:"tcp_flags"`

	Seq    uint32 `json:"seq"`     // seq
	AckSeq uint32 `json:"ack_seq"` // ack

	TCPPayloadSize uint32 `json:"tcp_payload_size"`

	SrcMAC string `json:"src_mac"`
	DstMAC string `json:"dst_mac"`
	Win    uint32 `json:"win"`

	TS int64 `json:"ts"` // nano second
}

func (f PktTCPHdr) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{f.TXRX, f.SrcMAC, f.DstMAC, f.Flags.String(), f.Seq, f.AckSeq, f.TCPPayloadSize, f.Win, f.TS})
}

type PValue struct {
	connTraceID *spanid.ID128
	reuseByNxt  bool // 被重用或者 rst 时可以考虑不再等待 2MSL
	sMACEQ      bool
	v6          bool

	tcpInfo TCPLog

	httpInfo HTTPLog

	lastGetTS int64
}

type conns struct {
	poolCreateTime int64 // unix timestamp ns
	pool           connsMaps

	twoMSLPool connsMaps // maybe recv/send RST after FIN etc.

	// timeoutDur time.Duration

	sync.RWMutex
}

type TCPConns struct {
	reuseIdx uint64 // 避免 time_wait 中的连接重复

	conns conns

	portListen   *portListen
	ifaceNameMAC [2]string

	tags map[string]string

	url    string
	aggURL string

	started int64 // -1 stop, 0 wait init, 1 started

	ctrID string
	nsUID string

	agg     FlowAggTCP
	aggHTTP FlowAggHTTP

	stop chan struct{}

	runtime   *filterRuntime
	blacklist ast.Stmts
	// upload event
	// ch        chan []*point.Point
	// cleanUpCh chan map[PktMeta]*PktValue
}

func NewTCPConns(gtags map[string]string, url, aggURL, ctrID, nsUID string,
	nameAddr [2]string, pr *portListen, bl ast.Stmts, runtime *filterRuntime,
) *TCPConns {
	tags := map[string]string{}

	for k, v := range tags {
		tags[k] = v
	}

	return &TCPConns{
		runtime:   runtime,
		blacklist: bl,

		aggURL:       aggURL,
		url:          url,
		ifaceNameMAC: nameAddr,
		portListen:   pr,
		tags:         tags,
		ctrID:        ctrID,
		nsUID:        nsUID,
		conns: conns{
			poolCreateTime: time.Now().UnixNano(),

			pool:       *newConnsMaps(defaultTCPKeepAlive / 4),
			twoMSLPool: *newConnsMaps(time.Second * 10),

			// timeoutDur:     timeoutDur,
		},

		stop: make(chan struct{}),
		// ch:        make(chan []*point.Point, 64),
		// cleanUpCh: make(chan map[PktMeta]*PktValue, 2),
	}
}

func (conns *TCPConns) getVal(k *PMeta, ts int64, syncFlagOnly bool) (*PValue, bool) {
	if k == nil {
		return nil, false
	}

	key := *k

	if _, v, ok := conns.conns.pool.getMapAndV(key); ok {
		v.lastGetTS = ts
		return v, true
	}

	var tcpReuse bool

	if _, v, ok := conns.conns.twoMSLPool.getMapAndV(key); ok {
		if syncFlagOnly && v.tcpInfo.GetPktChunk(false).RetransmitsSYN < 3 { // maybe resuse
			v.reuseByNxt = true
			tcpReuse = true
		} else if !v.reuseByNxt {
			v.lastGetTS = ts
			return v, true
		}
	}

	v := &PValue{
		tcpInfo: TCPLog{
			metric: TCPMetrics{
				recEstab: true,
			},
		},
	}
	if id, ok := genID128(); ok {
		// set conn innter traceid
		v.connTraceID = id
	}

	v.lastGetTS = ts
	if tcpReuse {
		v.tcpInfo.reuseConn = tcpReuse
	}

	conns.conns.pool.insert2LastMap(key, v)
	return v, true
}

func (conns *TCPConns) markTCPTimeWait(k *PMeta) {
	if k == nil {
		return
	}

	key := *k
	if mps, val, ok := conns.conns.pool.getMapAndV(key); ok {
		mps.delete(key)

		// 如果上次的仍然存在, 待消费

		if mps2msl, e, ok := conns.conns.twoMSLPool.getMapAndV(key); ok {
			mps2msl.delete(key)

			// 拷贝 tcp meta key， 并设置 reuse id
			nk := key
			conns.reuseIdx++
			if conns.reuseIdx == 0 {
				conns.reuseIdx++
			}
			nk.reuseidx = conns.reuseIdx
			// 写回
			mps2msl.insert(nk, e)
			mps2msl.insert(key, val)
		} else {
			conns.conns.twoMSLPool.insert2LastMap(key, val)
		}

		return
	}
}

func (conns *TCPConns) update(txRx int8, k *PMeta, ln *PktTCPHdr, pktLen,
	tcpPayloadSize int64, payload []byte, scale int, v6 bool,
) {
	if k == nil {
		return
	}

	var smac string
	if txRx == directionRX {
		// ip
		k.SrcIP, k.DstIP = k.DstIP, k.SrcIP
		// port
		k.SrcPort, k.DstPort = k.DstPort, k.SrcPort
		smac = ln.DstMAC
	} else {
		smac = ln.SrcMAC
	}

	if conns.runtime != nil && len(conns.blacklist) > 0 {
		var sPod, dPod string
		if k8sNetInfo != nil {
			sPod = k8sNetInfo.QueryPodName(k.SrcIP, uint32(k.SrcPort), "tcp")
			dPod = k8sNetInfo.QueryPodName(k.DstIP, uint32(k.DstPort), "tcp")
		}
		elem := &netParams{
			tcp:       true,
			k8sSrcPod: sPod,
			k8sDstPod: dPod,
			sPort:     int64(k.SrcPort),
			dPort:     int64(k.DstPort),
		}
		if v6 {
			elem.ipv4 = false
			elem.ip6SAddr = k.SrcIP
			elem.ip6DAddr = k.DstIP
		} else {
			elem.ipv4 = true
			elem.ipSAddr = k.SrcIP
			elem.ipDAddr = k.DstIP
		}

		if conns.runtime.runNetFilterDrop(conns.blacklist, elem) {
			return
		}
	}

	conns.conns.Lock()
	defer conns.conns.Unlock()

	// get conn stautus

	synOnly := ln.Flags.HasFlag(TCPSYN) && !ln.Flags.HasFlag(TCPACK)
	pktVal, _ := conns.getVal(k, ln.TS, synOnly)
	if pktVal == nil {
		return
	}

	if !pktVal.sMACEQ && conns.ifaceNameMAC[1] == smac {
		pktVal.sMACEQ = true
	}

	if v6 && !pktVal.v6 {
		pktVal.v6 = v6
	}

	pktState := pktVal.tcpInfo.Handle(txRx, payload, tcpPayloadSize, ln, k, scale)

	if pktVal.tcpInfo.tcpStatusRec.tcpStatus == TCPTimeWait ||
		pktVal.tcpInfo.tcpStatusRec.tcpStatus == TCPClose {
		conns.markTCPTimeWait(k)
		if !pktVal.tcpInfo.metric.recClose[0] {
			pktVal.tcpInfo.metric.recClose[0] = true
			pktVal.tcpInfo.metric.recClose[1] = true
		}
	} else if pktVal.tcpInfo.GetPktChunk(false).RetransmitsSYN >= 3 {
		conns.markTCPTimeWait(k)
	}

	if pktVal.tcpInfo.tcpStatusRec.tcpStatus == TCPEstablished || tcpPayloadSize > 0 {
		_ = pktVal.httpInfo.Handle(pktVal, txRx, payload, tcpPayloadSize, ln, k,
			pktState, pktVal.tcpInfo.GetPktChunk(false).ChunkID)
	}

	switch pktVal.tcpInfo.direction {
	case directionIncoming:
	case directionOutgoing:
	case directionUnknown:
		d := conns.portListen.Query(conns.nsUID, k, v6, pktVal.sMACEQ)
		if d == directionUnknown {
			if len((pktVal.httpInfo.elems)) > 0 {
				if v := pktVal.httpInfo.elems[0]; v != nil {
					switch v.Direction {
					case DOutging:
						pktVal.tcpInfo.direction = directionOutgoing
					case DIncoming:
						pktVal.tcpInfo.direction = directionIncoming
					}
				}
			}
			if synOnly {
				switch txRx {
				case directionRX:
					pktVal.tcpInfo.direction = directionIncoming
				case directionTX:
					pktVal.tcpInfo.direction = directionOutgoing
				}
			}
		} else {
			pktVal.tcpInfo.direction = d
		}
	default:
	}
}

func (conns *TCPConns) _ForceGather(nicIPList []string) {
	conns.conns.Lock()
	defer conns.conns.Unlock()

	for _, pool := range conns.conns.pool.maps {
		conns.feedNetworkLog(pool,
			false, true,
			conns.ifaceNameMAC, nicIPList)
	}
	for _, map2msl := range conns.conns.twoMSLPool.maps {
		conns.feedNetworkLog(map2msl,
			false, true,
			conns.ifaceNameMAC, nicIPList)
	}
}

func (conns *TCPConns) _Gather(nicIPList []string) {
	conns.conns.Lock()
	defer conns.conns.Unlock()

	{
		connPool := []*connMap{}
		lenMaps := len(conns.conns.pool.maps)
		for i := 0; i < lenMaps; i++ {
			mps := conns.conns.pool.maps[i]
			conns.feedNetworkLog(mps,
				false, false, conns.ifaceNameMAC, nicIPList)

			// keepalive
			if time.Since(mps.tn) >= defaultTCPKeepAlive {
				if lenMaps > i+1 {
					// 如果超时，把其他元素移动到后一个的 map 中
					for k, v := range mps.m {
						conns.conns.pool.maps[i+1].insert(k, v)
					}
				} else { // 后一个 map 不存在则不迁移
					connPool = append(connPool, mps)
				}
			} else {
				connPool = append(connPool, mps)
			}
		}

		conns.conns.pool.maps = connPool
	}

	{
		connPool := []*connMap{}
		lenMaps := len(conns.conns.twoMSLPool.maps)
		for i := 0; i < lenMaps; i++ {
			mps := conns.conns.twoMSLPool.maps[i]
			conns.feedNetworkLog(mps,
				true, false,
				conns.ifaceNameMAC, nicIPList)

			// 2msl
			if time.Since(mps.tn) >= twoMSL {
				if lenMaps > i+1 {
					for k, v := range mps.m {
						conns.conns.twoMSLPool.maps[i+1].insert(k, v)
					}
				} else {
					connPool = append(connPool, mps)
				}
			} else {
				connPool = append(connPool, mps)
			}
		}
		conns.conns.twoMSLPool.maps = connPool
	}
}

func (conns *TCPConns) CapturePacket(ctx context.Context, name, mac, netns string,
	h *afpacket.TPacket,
) {
	if h == nil {
		log.Error("param h is nil")
		return
	}

	if conns == nil {
		log.Error("param conns is nil")
		return
	}

	if !atomic.CompareAndSwapInt64(&conns.started, 0, 1) {
		// or maybe started == -1
		log.Warnf("already started (name: %s ,iface: %s)", name, mac)
		return
	}

	layerLi := make([]gopacket.LayerType, 0, 10)

	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if _, s3, err := h.SocketStats(); err != nil {
				log.Error(err)
			} else {
				log.Infof("name %s, mac %s, ns %s, drops %d, packets %d, freezes %d",
					name, mac, netns, s3.Drops(), s3.Packets(), s3.QueueFreezes())
			}
		case <-ctx.Done():
			h.Close()
			if old := atomic.SwapInt64(&conns.started, -1); old == 1 && conns.stop != nil {
				close(conns.stop)
			}
			return
		default:
		}

		decoder := NewPktDecoder()
		layerLi = layerLi[:0]

		buf, ci, err := h.ZeroCopyReadPacketData()
		if err != nil {
			log.Error(err)
			time.Sleep(time.Millisecond * 300)
		}

		_ = decoder.pktDecode.DecodeLayers(buf, &layerLi)

		if len(layerLi) < 3 || layerLi[0] != layers.LayerTypeEthernet {
			continue
		}

		ipLayerType := layerLi[1]
		var vxlanPkt bool
		var vniID uint32

		switch layerLi[2] {
		case layers.LayerTypeTCP:
		case layers.LayerTypeUDP:
			if !isVxlanLayer(uint16(decoder.udp.SrcPort), uint16(decoder.udp.DstPort)) {
				continue
			}

			layerLi = layerLi[:0]
			_ = decoder.vxlanDecode.DecodeLayers(decoder.udp.Payload, &layerLi)
			if len(layerLi) < 4 || layerLi[0] != layers.LayerTypeVXLAN ||
				layerLi[1] != layers.LayerTypeEthernet ||
				layerLi[3] != layers.LayerTypeTCP {
				continue
			} else {
				vxlanPkt = true
				vniID = decoder.vxlan.VNI
				ipLayerType = layerLi[2]
			}

		default:
			continue
		}

		k := &PMeta{
			// IfIndex:       ci.InterfaceIndex,

			VNIID: vniID,
			VXLAN: vxlanPkt,

			SrcPort: uint16(decoder.tcp.SrcPort),
			DstPort: uint16(decoder.tcp.DstPort),
		}

		ln := &PktTCPHdr{
			SrcMAC: decoder.eth.SrcMAC.String(),
			DstMAC: decoder.eth.DstMAC.String(),
			AckSeq: decoder.tcp.Ack,
			Seq:    decoder.tcp.Seq,

			Win: uint32(decoder.tcp.Window),

			TS: ci.Timestamp.UnixNano(),
		}

		if len(decoder.tcp.Contents) >= 14 {
			ln.Flags = TCPFlag(decoder.tcp.Contents[13])
		}

		var scale int
		if decoder.tcp.SYN {
			for _, opt := range decoder.tcp.Options {
				if opt.OptionType == layers.TCPOptionKindWindowScale {
					scale = int(opt.OptionData[0])
				}
			}
		}

		var isipv6 bool
		// rx ? eth frame min size 60 here (not include FCS 4byte) : no eth padding
		if ipLayerType == layers.LayerTypeIPv4 {
			k.SrcIP = decoder.ipv4.SrcIP.String()
			k.DstIP = decoder.ipv4.DstIP.String()

			if ci.Length > 64 {
				ln.TCPPayloadSize = uint32(int(decoder.ipv4.Length) -
					len(decoder.ipv4.BaseLayer.Contents) -
					len(decoder.tcp.BaseLayer.Contents))
			}
		} else { // ipv6
			isipv6 = true
			k.SrcIP = decoder.ipv6.SrcIP.String()
			k.DstIP = decoder.ipv6.DstIP.String()

			ln.TCPPayloadSize = uint32(int(decoder.ipv6.Length) -
				len(decoder.tcp.BaseLayer.Contents))
		}

		var txRx int8
		for _, v := range ci.AncillaryData {
			if v, ok := v.(afpacket.AncillaryPktType); ok {
				if v.Type == unix.PACKET_OUTGOING {
					txRx = directionTX
					ln.TXRX = "tx"
				} else {
					txRx = directionRX
					ln.TXRX = "rx"
				}
			}
		}

		if txRx == 0 {
			log.Warnf("iface %s, name %s, meta %v value %v", conns.nsUID, conns.ifaceNameMAC, k, *ln)
			continue
		}

		conns.update(txRx, k, ln, int64(ci.Length),
			int64(ln.TCPPayloadSize), decoder.tcp.BaseLayer.Payload, scale, isipv6)
	}
}

func (conns *TCPConns) Gather(ctx context.Context, nicIPList []string) {
	aggTicker := time.NewTicker(time.Second * 60)
	defer aggTicker.Stop()

	aggHTTPTicker := time.NewTicker(time.Second * 60)
	defer aggHTTPTicker.Stop()

	ticker := time.NewTicker(time.Second * 8)
	defer ticker.Stop()

	for {
		select {
		case <-conns.stop:
			log.Info("tcp conns gather stop")
			// 强制清理所有数据进行上报
			conns._ForceGather(nicIPList)

			return
		case <-ticker.C:
			conns._Gather(nicIPList)

		case <-aggTicker.C:
			// netflow data (cat: Network)
			if enabledNetMetric {
				pts := conns.agg.ToPoint(conns.tags, k8sNetInfo)
				if len(pts) > 0 {
					if err := output.FeedPoint(conns.aggURL, pts, false); err != nil {
						log.Errorf("feed point(toatl %d) failed: %w", len(pts), err)
					}
				}
			}

			conns.agg.Clean()

		case <-aggHTTPTicker.C:
			// httpflow
			if enabledNetMetric {
				pts := conns.aggHTTP.ToPoint(conns.tags, k8sNetInfo)
				if len(pts) > 0 {
					if err := output.FeedPoint(conns.aggURL, pts, false); err != nil {
						log.Errorf("feed point(toatl %d) failed: %w", len(pts), err)
					}
				}
			}
			conns.aggHTTP.Clean()

		case <-ctx.Done():
			return
		}
	}
}
