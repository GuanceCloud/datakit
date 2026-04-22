//go:build linux
// +build linux

// Package l4log capture packets
package l4log

import (
	"context"
	"encoding/binary"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/google/gopacket"
	"github.com/google/gopacket/afpacket"
	"github.com/google/gopacket/layers"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/cli"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/spanid"
	"golang.org/x/sys/unix"
)

func ancillaryDirection(data []interface{}) (int8, string) {
	for _, v := range data {
		if v, ok := v.(afpacket.AncillaryPktType); ok {
			if v.Type == unix.PACKET_OUTGOING {
				return directionTX, "tx"
			}

			return directionRX, "rx"
		}
	}

	return 0, ""
}

type packetStringCache struct {
	mac  map[[6]byte]string
	ipv4 map[[4]byte]string
	ipv6 map[[16]byte]string
}

const (
	packetStringCacheMaxMAC  = 256
	packetStringCacheMaxIPv4 = 4096
	packetStringCacheMaxIPv6 = 2048
)

func newPacketStringCache() *packetStringCache {
	return &packetStringCache{
		mac:  make(map[[6]byte]string, 32),
		ipv4: make(map[[4]byte]string, 256),
		ipv6: make(map[[16]byte]string, 64),
	}
}

func (c *packetStringCache) macString(addr net.HardwareAddr) string {
	if len(addr) != 6 {
		return addr.String()
	}

	var key [6]byte
	copy(key[:], addr)
	if v, ok := c.mac[key]; ok {
		return v
	}

	if len(c.mac) >= packetStringCacheMaxMAC {
		c.mac = make(map[[6]byte]string, 32)
	}

	v := addr.String()
	c.mac[key] = v
	return v
}

func (c *packetStringCache) ipString(ip net.IP) string {
	if v4 := ip.To4(); len(v4) == 4 {
		var key [4]byte
		copy(key[:], v4)
		if v, ok := c.ipv4[key]; ok {
			return v
		}

		if len(c.ipv4) >= packetStringCacheMaxIPv4 {
			c.ipv4 = make(map[[4]byte]string, 256)
		}

		v := v4.String()
		c.ipv4[key] = v
		return v
	}

	if len(ip) == net.IPv6len {
		var key [16]byte
		copy(key[:], ip)
		if v, ok := c.ipv6[key]; ok {
			return v
		}

		if len(c.ipv6) >= packetStringCacheMaxIPv6 {
			c.ipv6 = make(map[[16]byte]string, 64)
		}

		v := ip.String()
		c.ipv6[key] = v
		return v
	}

	return ip.String()
}

type fastPacketInfo struct {
	key        PMeta
	tcpHdr     PktTCPHdr
	payload    []byte
	scale      int
	ipv6       bool
	payloadLen int64
}

const (
	etherTypeIPv4 = 0x0800
	etherTypeIPv6 = 0x86DD
	etherTypeVLAN = 0x8100
	etherTypeQinQ = 0x88a8

	minEthernetHeader = 14
	minIPv4Header     = 20
	minIPv6Header     = 40
	minTCPHeader      = 20
)

func parseFastTCPPacket(buf []byte, ts int64, cache *packetStringCache) (fastPacketInfo, bool) {
	var info fastPacketInfo

	if len(buf) < minEthernetHeader+minIPv4Header+minTCPHeader || cache == nil {
		return info, false
	}

	etherType := binary.BigEndian.Uint16(buf[12:14])
	switch etherType {
	case etherTypeVLAN, etherTypeQinQ:
		return info, false
	}

	info.tcpHdr = PktTCPHdr{
		SrcMAC: cache.macString(net.HardwareAddr(buf[6:12])),
		DstMAC: cache.macString(net.HardwareAddr(buf[0:6])),
		TS:     ts,
	}

	switch etherType {
	case etherTypeIPv4:
		if !parseFastIPv4Packet(buf, cache, &info) {
			return fastPacketInfo{}, false
		}
	case etherTypeIPv6:
		if !parseFastIPv6Packet(buf, cache, &info) {
			return fastPacketInfo{}, false
		}
	default:
		return info, false
	}

	return info, true
}

func parseFastIPv4Packet(buf []byte, cache *packetStringCache, info *fastPacketInfo) bool {
	if len(buf) < minEthernetHeader+minIPv4Header+minTCPHeader {
		return false
	}

	ipStart := minEthernetHeader
	if version := buf[ipStart] >> 4; version != 4 {
		return false
	}

	ihl := int(buf[ipStart]&0x0f) * 4
	if ihl < minIPv4Header || len(buf) < ipStart+ihl+minTCPHeader {
		return false
	}

	totalLen := int(binary.BigEndian.Uint16(buf[ipStart+2 : ipStart+4]))
	if totalLen < ihl+minTCPHeader || len(buf) < ipStart+totalLen {
		return false
	}

	frag := binary.BigEndian.Uint16(buf[ipStart+6 : ipStart+8])
	if frag&0x3fff != 0 {
		return false
	}

	if buf[ipStart+9] != byte(layers.IPProtocolTCP) {
		return false
	}

	tcpStart := ipStart + ihl
	return parseFastTCPHeader(buf[tcpStart:ipStart+totalLen], cache.ipString(net.IP(buf[ipStart+12:ipStart+16])),
		cache.ipString(net.IP(buf[ipStart+16:ipStart+20])), false, info)
}

func parseFastIPv6Packet(buf []byte, cache *packetStringCache, info *fastPacketInfo) bool {
	if len(buf) < minEthernetHeader+minIPv6Header+minTCPHeader {
		return false
	}

	ipStart := minEthernetHeader
	if version := buf[ipStart] >> 4; version != 6 {
		return false
	}

	payloadLen := int(binary.BigEndian.Uint16(buf[ipStart+4 : ipStart+6]))
	if payloadLen < minTCPHeader || len(buf) < ipStart+minIPv6Header+payloadLen {
		return false
	}

	if buf[ipStart+6] != byte(layers.IPProtocolTCP) {
		return false
	}

	tcpStart := ipStart + minIPv6Header
	return parseFastTCPHeader(buf[tcpStart:ipStart+minIPv6Header+payloadLen],
		cache.ipString(net.IP(buf[ipStart+8:ipStart+24])),
		cache.ipString(net.IP(buf[ipStart+24:ipStart+40])), true, info)
}

func parseFastTCPHeader(seg []byte, srcIP, dstIP string, ipv6 bool, info *fastPacketInfo) bool {
	if len(seg) < minTCPHeader {
		return false
	}

	dataOffset := int(seg[12]>>4) * 4
	if dataOffset < minTCPHeader || len(seg) < dataOffset {
		return false
	}

	info.key.SrcIP = srcIP
	info.key.DstIP = dstIP
	info.key.SrcPort = binary.BigEndian.Uint16(seg[0:2])
	info.key.DstPort = binary.BigEndian.Uint16(seg[2:4])
	info.ipv6 = ipv6

	info.tcpHdr.Seq = binary.BigEndian.Uint32(seg[4:8])
	info.tcpHdr.AckSeq = binary.BigEndian.Uint32(seg[8:12])
	info.tcpHdr.Flags = TCPFlag(seg[13])
	info.tcpHdr.Win = uint32(binary.BigEndian.Uint16(seg[14:16]))

	info.payload = seg[dataOffset:]
	info.payloadLen = int64(len(info.payload))
	info.tcpHdr.TCPPayloadSize = len(info.payload)

	if info.tcpHdr.Flags.HasFlag(TCPSYN) {
		info.scale = parseTCPWindowScale(seg[minTCPHeader:dataOffset])
	}

	return true
}

func parseTCPWindowScale(opts []byte) int {
	for idx := 0; idx < len(opts); {
		kind := opts[idx]
		switch kind {
		case 0:
			return 0
		case 1:
			idx++
			continue
		}

		if idx+1 >= len(opts) {
			return 0
		}

		ln := int(opts[idx+1])
		if ln < 2 || idx+ln > len(opts) {
			return 0
		}

		if kind == uint8(layers.TCPOptionKindWindowScale) && ln >= 3 {
			return int(opts[idx+2])
		}

		idx += ln
	}

	return 0
}

type NetProtoTyp string

const (
	NetProtoTCP  NetProtoTyp = "tcp"
	NetProtoHTTP NetProtoTyp = "http"
)

var k8sNetInfo *cli.K8sInfo

func SetK8sNetInfo(n *cli.K8sInfo) {
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

	TCPPayloadSize int `json:"tcp_payload_size"`

	SrcMAC string `json:"src_mac"`
	DstMAC string `json:"dst_mac"`
	Win    uint32 `json:"win"`

	TS int64 `json:"ts"` // nano second
}

func (f PktTCPHdr) MarshalJSON() ([]byte, error) {
	buf := make([]byte, 0, 96)
	buf = appendPktTCPHdrJSON(buf, f)
	return buf, nil
}

func appendPktTCPHdrJSON(buf []byte, f PktTCPHdr) []byte {
	buf = append(buf, '[')
	buf = strconv.AppendQuote(buf, f.TXRX)
	buf = append(buf, ',')
	buf = strconv.AppendQuote(buf, f.SrcMAC)
	buf = append(buf, ',')
	buf = strconv.AppendQuote(buf, f.DstMAC)
	buf = append(buf, ',')
	buf = strconv.AppendQuote(buf, f.Flags.String())
	buf = append(buf, ',')
	buf = strconv.AppendUint(buf, uint64(f.Seq), 10)
	buf = append(buf, ',')
	buf = strconv.AppendUint(buf, uint64(f.AckSeq), 10)
	buf = append(buf, ',')
	buf = strconv.AppendInt(buf, int64(f.TCPPayloadSize), 10)
	buf = append(buf, ',')
	buf = strconv.AppendUint(buf, uint64(f.Win), 10)
	buf = append(buf, ',')
	buf = strconv.AppendInt(buf, f.TS, 10)
	buf = append(buf, ']')
	return buf
}

type PValue struct {
	connTraceID *spanid.ID128
	reuseByNxt  bool // 被重用或者 rst 时可以考虑不再等待 2MSL
	sMACEQ      bool
	v6          bool

	tcpInfo TCPLog

	httpInfo HTTPLog

	tlsSNI tlsClientHelloState

	lastGetTS            int64
	directionLastProbeTS int64
	directionProbeMisses uint8

	baseKVsCache     point.KVs
	baseTagsCacheTS  int64
	baseTagsCacheDir string
}

type blacklistCacheEntry struct {
	drop   bool
	lastTS int64
}

type blacklistCacheKey struct {
	meta   PMeta
	srcPod string
	dstPod string
}

const (
	blacklistCacheTTL             = defaultTCPKeepAlive
	blacklistCacheCleanupInterval = 30 * time.Second
	blacklistCacheCleanupMinSize  = 512

	directionProbeBaseInterval = 200 * time.Millisecond
	directionProbeMaxInterval  = 5 * time.Second
)

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

	blacklistMu               sync.Mutex
	blacklistCache            map[blacklistCacheKey]blacklistCacheEntry
	blacklistCacheLastCleanup int64
	blacklistCacheRuleFirst   any
	blacklistCacheRuleLen     int

	portListen   *portListen
	ifaceNameMAC [2]string

	hostNetwork bool
	trustLocal  bool
	virtualNIC  bool

	tagsMu sync.RWMutex
	tags   map[string]string

	started int64 // -1 stop, 0 wait init, 1 started

	ctrID string
	nsUID string

	agg     FlowAggTCP
	aggHTTP FlowAggHTTP

	lastTPacketStats tpacketStatsSnapshot

	stop     chan struct{}
	stopOnce sync.Once

	runtime   *filterRuntime
	blacklist ast.Stmts
	// upload event
	// ch        chan []*point.Point
	// cleanUpCh chan map[PktMeta]*PktValue
}

func NewTCPConns(gtags map[string]string, ctrID, nsUID string,
	nameAddr [2]string, pr *portListen, bl ast.Stmts, runtime *filterRuntime,
) *TCPConns {
	return &TCPConns{
		runtime:   runtime,
		blacklist: bl,

		ifaceNameMAC: nameAddr,
		portListen:   pr,
		tags:         cloneStringMap(gtags),
		ctrID:        ctrID,
		nsUID:        nsUID,
		conns: conns{
			poolCreateTime: ntp.Now().UnixNano(),

			pool:       *newConnsMaps(defaultTCPKeepAlive / 4),
			twoMSLPool: *newConnsMaps(time.Second * 10),
		},

		stop: make(chan struct{}),
	}
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}

	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (conns *TCPConns) runtimeTags() map[string]string {
	if conns == nil {
		return nil
	}

	conns.tagsMu.RLock()
	tags := conns.tags
	conns.tagsMu.RUnlock()
	return tags
}

func (conns *TCPConns) UpdateTags(tags map[string]string) {
	if conns == nil {
		return
	}

	next := cloneStringMap(tags)

	conns.tagsMu.Lock()
	conns.tags = next
	conns.tagsMu.Unlock()

	conns.conns.Lock()
	defer conns.conns.Unlock()

	reset := func(pool *connsMaps) {
		if pool == nil {
			return
		}
		for _, mps := range pool.maps {
			if mps == nil {
				continue
			}
			for _, v := range mps.m {
				if v == nil {
					continue
				}
				v.baseKVsCache = nil
				v.baseTagsCacheTS = 0
			}
		}
	}

	reset(&conns.conns.pool)
	reset(&conns.conns.twoMSLPool)
}

func (conns *TCPConns) signalStop() {
	if conns == nil {
		return
	}

	conns.stopOnce.Do(func() {
		if conns.stop != nil {
			close(conns.stop)
		}
	})
}

func (conns *TCPConns) getVal(k *PMeta, ts int64, syncFlagOnly bool) (*PValue, *connMap, bool) {
	if k == nil {
		return nil, nil, false
	}

	key := *k

	if mps, v, ok := conns.conns.pool.getMapAndV(key); ok {
		v.lastGetTS = ts
		return v, mps, true
	}

	var tcpReuse bool

	if mps, v, ok := conns.conns.twoMSLPool.getMapAndV(key); ok {
		if syncFlagOnly && v.tcpInfo.RetransmitsSYN < 3 { // maybe resuse
			v.reuseByNxt = true
			tcpReuse = true
			v.tcpInfo.RetransmitsSYN = 0
		} else if !v.reuseByNxt {
			v.lastGetTS = ts
			return v, mps, true
		}
	}

	v := &PValue{
		sMACEQ: conns.trustLocal,
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

	mps := conns.conns.pool.insert2LastMap(key, v)
	return v, mps, true
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
			mps2msl.markDirty(key)
		} else {
			mps2msl := conns.conns.twoMSLPool.insert2LastMap(key, val)
			mps2msl.markDirty(key)
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

	if conns.shouldDropByBlacklist(k, v6, ln.TS) {
		return
	}

	conns.conns.Lock()
	defer conns.conns.Unlock()

	// get conn stautus

	synOnly := ln.Flags.HasFlag(TCPSYN) && !ln.Flags.HasFlag(TCPACK)
	pktVal, pktMap, _ := conns.getVal(k, ln.TS, synOnly)
	if pktVal == nil {
		return
	}
	if pktMap != nil {
		pktMap.markDirty(*k)
	}

	if !pktVal.sMACEQ && conns.ifaceNameMAC[1] == smac {
		pktVal.sMACEQ = true
	}

	if v6 && !pktVal.v6 {
		pktVal.v6 = v6
	}

	pktState := pktVal.tcpInfo.Handle(txRx, payload, tcpPayloadSize, ln, k, scale)

	if pktVal.tcpInfo.tcpState == TCPTimeWait ||
		pktVal.tcpInfo.tcpState == TCPClose {
		conns.markTCPTimeWait(k)
		if !pktVal.tcpInfo.metric.recClose[0] {
			pktVal.tcpInfo.metric.recClose[0] = true
			pktVal.tcpInfo.metric.recClose[1] = true
		}
	} else if pktVal.tcpInfo.RetransmitsSYN >= 3 {
		conns.markTCPTimeWait(k)
	}

	if enableL7HTTP && pktVal.httpInfo.ShouldHandle(payload) {
		_ = pktVal.httpInfo.Handle(pktVal, txRx, payload, tcpPayloadSize, ln, k,
			pktState, pktVal.tcpInfo.GetPktChunk(false, false).ChunkID)
	}

	pktVal.observeTLSSNI(txRx, payload, k, conns.nsUID)

	// maybe proto will change
	if pktVal.httpInfo.isHTTP {
		pktVal.tcpInfo.l7proto = L7ProtoHTTP
	}

	switch pktVal.tcpInfo.direction {
	case directionIncoming:
	case directionOutgoing:
	case directionUnknown:
		if pktVal.shouldProbeDirection(ln.TS, synOnly) {
			d := conns.portListen.Query(conns.nsUID, k, v6, pktVal.sMACEQ)
			pktVal.recordDirectionProbe(ln.TS, d)
			if d != directionUnknown {
				pktVal.tcpInfo.direction = d
				break
			}
		}

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
	default:
	}
}

func (v *PValue) shouldProbeDirection(ts int64, force bool) bool {
	if force || ts <= 0 || v.directionLastProbeTS == 0 {
		return true
	}

	interval := directionProbeBaseInterval
	if misses := v.directionProbeMisses; misses > 1 {
		for i := uint8(1); i < misses; i++ {
			interval *= 2
			if interval >= directionProbeMaxInterval {
				interval = directionProbeMaxInterval
				break
			}
		}
	}

	return ts-v.directionLastProbeTS >= interval.Nanoseconds()
}

func (v *PValue) recordDirectionProbe(ts int64, d conndirection) {
	if ts > 0 {
		v.directionLastProbeTS = ts
	}

	if d == directionUnknown {
		if v.directionProbeMisses < ^uint8(0) {
			v.directionProbeMisses++
		}
		return
	}

	v.directionProbeMisses = 0
}

func (conns *TCPConns) shouldDropByBlacklist(k *PMeta, v6 bool, ts int64) bool {
	if k == nil || conns.runtime == nil || len(conns.blacklist) == 0 {
		return false
	}

	now := ts
	if now <= 0 {
		now = ntp.Now().UnixNano()
	}

	var sPod, dPod string
	if k8sNetInfo != nil {
		sPod = k8sNetInfo.QueryPodName(0, k.SrcIP)
		dPod = k8sNetInfo.QueryPodName(0, k.DstIP)
	}

	key := blacklistCacheKey{
		meta:   *k,
		srcPod: sPod,
		dstPod: dPod,
	}
	if cached, ok := conns.getBlacklistCache(key, now); ok {
		return cached
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

	drop := conns.runtime.runNetFilterDrop(conns.blacklist, elem)
	return conns.storeBlacklistCache(key, drop, now)
}

func (conns *TCPConns) getBlacklistCache(key blacklistCacheKey, now int64) (bool, bool) {
	conns.blacklistMu.Lock()
	defer conns.blacklistMu.Unlock()

	conns.refreshBlacklistCacheScopeLocked()
	if len(conns.blacklistCache) == 0 {
		return false, false
	}

	entry, ok := conns.blacklistCache[key]
	if !ok {
		conns.cleanupBlacklistCacheLocked(now)
		return false, false
	}

	entry.lastTS = now
	conns.blacklistCache[key] = entry
	conns.cleanupBlacklistCacheLocked(now)
	return entry.drop, true
}

func (conns *TCPConns) storeBlacklistCache(key blacklistCacheKey, drop bool, now int64) bool {
	conns.blacklistMu.Lock()
	defer conns.blacklistMu.Unlock()

	conns.refreshBlacklistCacheScopeLocked()
	if conns.blacklistCache == nil {
		conns.blacklistCache = make(map[blacklistCacheKey]blacklistCacheEntry)
	}

	if entry, ok := conns.blacklistCache[key]; ok {
		entry.lastTS = now
		conns.blacklistCache[key] = entry
		conns.cleanupBlacklistCacheLocked(now)
		return entry.drop
	}

	conns.blacklistCache[key] = blacklistCacheEntry{
		drop:   drop,
		lastTS: now,
	}
	conns.cleanupBlacklistCacheLocked(now)
	return drop
}

func (conns *TCPConns) refreshBlacklistCacheScopeLocked() {
	var first any
	if len(conns.blacklist) > 0 {
		first = conns.blacklist[0]
	}

	if conns.blacklistCacheRuleLen == len(conns.blacklist) &&
		conns.blacklistCacheRuleFirst == first {
		return
	}

	conns.blacklistCacheRuleLen = len(conns.blacklist)
	conns.blacklistCacheRuleFirst = first
	conns.blacklistCache = nil
	conns.blacklistCacheLastCleanup = 0
}

func (conns *TCPConns) cleanupBlacklistCacheLocked(now int64) {
	if len(conns.blacklistCache) < blacklistCacheCleanupMinSize {
		return
	}
	if now-conns.blacklistCacheLastCleanup < blacklistCacheCleanupInterval.Nanoseconds() {
		return
	}

	expireBefore := now - blacklistCacheTTL.Nanoseconds()
	for key, entry := range conns.blacklistCache {
		if entry.lastTS < expireBefore {
			delete(conns.blacklistCache, key)
		}
	}

	conns.blacklistCacheLastCleanup = now
}

func (conns *TCPConns) _ForceGather(nicIPList []string) {
	conns.conns.Lock()
	defer conns.conns.Unlock()

	for _, pool := range conns.conns.pool.maps {
		fullScan := needConnMapFullScan(pool, defaultTCPKeepAlive, true)
		conns.feedNetworkLog(pool,
			false, true, fullScan, nicIPList)
	}
	for _, map2msl := range conns.conns.twoMSLPool.maps {
		fullScan := needConnMapFullScan(map2msl, twoMSL, true)
		conns.feedNetworkLog(map2msl,
			false, true, fullScan, nicIPList)
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
			fullScan := needConnMapFullScan(mps, defaultTCPKeepAlive, false)
			conns.feedNetworkLog(mps,
				false, false, fullScan, nicIPList)

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
			fullScan := needConnMapFullScan(mps, twoMSL, false)
			conns.feedNetworkLog(mps,
				true, false, fullScan, nicIPList)

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
	decoder := NewPktDecoder()
	stringCache := newPacketStringCache()

	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if _, s3, err := h.SocketStats(); err != nil {
				log.Error(err)
			} else {
				observeTPacketStatsDelta("l4log", &conns.lastTPacketStats,
					uint64(s3.Packets()), uint64(s3.Drops()), uint64(s3.QueueFreezes()))
				log.Infof("name %s, mac %s, ns %s, drops %d, packets %d, freezes %d",
					name, mac, netns, s3.Drops(), s3.Packets(), s3.QueueFreezes())
			}
		case <-ctx.Done():
			h.Close()
			if old := atomic.SwapInt64(&conns.started, -1); old == 1 {
				conns.signalStop()
			}
			return
		default:
		}

		buf, ci, err := h.ZeroCopyReadPacketData()
		if err != nil {
			log.Error(err)
			continue
		}

		conns.handleCapturedPacket(decoder, layerLi, stringCache, buf, ci)
	}
}

func (conns *TCPConns) handleCapturedPacket(decoder *pktDecoder, layerLi []gopacket.LayerType,
	stringCache *packetStringCache, buf []byte, ci gopacket.CaptureInfo,
) {
	if conns == nil || decoder == nil {
		return
	}
	if stringCache == nil {
		stringCache = newPacketStringCache()
	}

	layerLi = layerLi[:0]

	txRx, txrxStr := ancillaryDirection(ci.AncillaryData)
	if txRx == 0 {
		log.Warnf("iface %s, name %s, packet direction unknown", conns.nsUID, conns.ifaceNameMAC)
		return
	}

	if fastPkt, ok := parseFastTCPPacket(buf, ci.Timestamp.UnixNano(), stringCache); ok {
		fastPkt.tcpHdr.TXRX = txrxStr
		conns.update(txRx, &fastPkt.key, &fastPkt.tcpHdr, int64(ci.Length),
			fastPkt.payloadLen, fastPkt.payload, fastPkt.scale, fastPkt.ipv6)
		return
	}

	_ = decoder.pktDecode.DecodeLayers(buf, &layerLi)

	if len(layerLi) < 3 || layerLi[0] != layers.LayerTypeEthernet {
		return
	}

	ipLayerType := layerLi[1]
	var vxlanPkt bool
	var vniID uint32

	switch layerLi[2] {
	case layers.LayerTypeTCP:
	case layers.LayerTypeUDP:
		if !isVxlanLayer(uint16(decoder.udp.SrcPort), uint16(decoder.udp.DstPort)) {
			return
		}

		layerLi = layerLi[:0]
		_ = decoder.vxlanDecode.DecodeLayers(decoder.udp.Payload, &layerLi)
		if len(layerLi) < 4 || layerLi[0] != layers.LayerTypeVXLAN ||
			layerLi[1] != layers.LayerTypeEthernet ||
			layerLi[3] != layers.LayerTypeTCP {
			return
		}

		vxlanPkt = true
		vniID = decoder.vxlan.VNI
		ipLayerType = layerLi[2]
	default:
		return
	}

	k := PMeta{
		VNIID:   vniID,
		VXLAN:   vxlanPkt,
		SrcPort: uint16(decoder.tcp.SrcPort),
		DstPort: uint16(decoder.tcp.DstPort),
	}

	ln := PktTCPHdr{
		SrcMAC: stringCache.macString(decoder.eth.SrcMAC),
		DstMAC: stringCache.macString(decoder.eth.DstMAC),
		AckSeq: decoder.tcp.Ack,
		Seq:    decoder.tcp.Seq,
		Win:    uint32(decoder.tcp.Window),
		TS:     ci.Timestamp.UnixNano(),
		TXRX:   txrxStr,
	}

	if len(decoder.tcp.Contents) >= 14 {
		ln.Flags = TCPFlag(decoder.tcp.Contents[13])
	}

	var scale int
	if decoder.tcp.SYN {
		for _, opt := range decoder.tcp.Options {
			if opt.OptionType == layers.TCPOptionKindWindowScale && len(opt.OptionData) > 0 {
				scale = int(opt.OptionData[0])
			}
		}
	}

	var isipv6 bool
	if ipLayerType == layers.LayerTypeIPv4 {
		k.SrcIP = stringCache.ipString(decoder.ipv4.SrcIP)
		k.DstIP = stringCache.ipString(decoder.ipv4.DstIP)

		if ci.Length > 64 {
			ln.TCPPayloadSize = int(decoder.ipv4.Length) -
				len(decoder.ipv4.BaseLayer.Contents) -
				len(decoder.tcp.BaseLayer.Contents)
		}
	} else {
		isipv6 = true
		k.SrcIP = stringCache.ipString(decoder.ipv6.SrcIP)
		k.DstIP = stringCache.ipString(decoder.ipv6.DstIP)

		ln.TCPPayloadSize = int(decoder.ipv6.Length) -
			len(decoder.tcp.BaseLayer.Contents)
	}

	conns.update(txRx, &k, &ln, int64(ci.Length),
		int64(ln.TCPPayloadSize), decoder.tcp.BaseLayer.Payload, scale, isipv6)
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
			log.Infof("close raw socket %s %s %s",
				conns.ifaceNameMAC[0], conns.ifaceNameMAC[1], conns.nsUID)

			// 强制清理所有数据进行上报
			conns._ForceGather(nicIPList)

			return
		case <-ticker.C:
			exporter.ObserveCacheEntries("l4log", "conn_pool", conns.conns.pool.entries())
			exporter.ObserveCacheEntries("l4log", "two_msl_pool", conns.conns.twoMSLPool.entries())
			exporter.ObserveCacheEntries("l4log", "blacklist_cache", len(conns.blacklistCache))
			conns._Gather(nicIPList)

		case <-aggTicker.C:
			// netflow data (cat: Network)
			if enabledNetMetric {
				exporter.ObserveAggEntries("l4log_netflow", conns.agg.Len())
				flushStart := time.Now()
				pts := conns.agg.ToPoint(conns.tags, k8sNetInfo)
				if len(pts) > 0 {
					if err := exporter.FeedPoint("bpf-netlog/netflow",
						point.Network, pts); err != nil {
						log.Errorf("feed point(toatl %d) failed: %w", len(pts), err)
						exporter.ObserveAggFlush("l4log_netflow", len(pts), time.Since(flushStart), "error")
					} else {
						exporter.ObserveAggFlush("l4log_netflow", len(pts), time.Since(flushStart), "ok")
					}
				} else {
					exporter.ObserveAggFlush("l4log_netflow", 0, time.Since(flushStart), "ok")
				}
			}

			conns.agg.Clean()
			exporter.ObserveAggEntries("l4log_netflow", 0)

		case <-aggHTTPTicker.C:
			// httpflow
			if enabledNetMetric {
				exporter.ObserveAggEntries("l4log_httpflow", conns.aggHTTP.Len())
				flushStart := time.Now()
				pts := conns.aggHTTP.ToPoint(conns.tags, k8sNetInfo)
				if len(pts) > 0 {
					if err := exporter.FeedPoint("bpf-netlog/httpflow",
						point.Network, pts); err != nil {
						log.Errorf("feed point(toatl %d) failed: %w", len(pts), err)
						exporter.ObserveAggFlush("l4log_httpflow", len(pts), time.Since(flushStart), "error")
					} else {
						exporter.ObserveAggFlush("l4log_httpflow", len(pts), time.Since(flushStart), "ok")
					}
				} else {
					exporter.ObserveAggFlush("l4log_httpflow", 0, time.Since(flushStart), "ok")
				}
			}
			conns.aggHTTP.Clean()
			exporter.ObserveAggEntries("l4log_httpflow", 0)

		case <-ctx.Done():
			return
		}
	}
}
