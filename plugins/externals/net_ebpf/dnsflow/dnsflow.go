//go:build (linux && ignore) || ebpf
// +build linux,ignore ebpf

package dnsflow

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/afpacket"
	"github.com/google/gopacket/layers"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	dkfeed "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/feed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"golang.org/x/net/bpf"
)

const inputName = "dnsflow"

var l = logger.DefaultSLogger("net_ebpf")

func SetLogger(nl *logger.Logger) {
	l = nl
}

type DNSParser struct {
	*gopacket.DecodingLayerParser
	layers []gopacket.LayerType
	eth    *layers.Ethernet
	ipv4   *layers.IPv4
	ipv6   *layers.IPv6
	udp    *layers.UDP
	tcp    *layers.TCP
	dns    *layers.DNS
}

func NewDNSParse() DNSParser {
	var eth layers.Ethernet
	var ipv4 layers.IPv4
	var ipv6 layers.IPv6
	var udp layers.UDP
	var tcp layers.TCP
	var dns layers.DNS
	l := []gopacket.DecodingLayer{
		&eth,
		&ipv4, &ipv6,
		&udp, &tcp,
		&dns,
	}

	return DNSParser{
		gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, l...),
		[]gopacket.LayerType{},
		&eth,
		&ipv4, &ipv6,
		&udp, &tcp,
		&dns,
	}
}

func NewTPacketDNS() (*afpacket.TPacket, error) {
	h, err := afpacket.NewTPacket()
	if err != nil {
		return nil, err
	}
	// 执行 sudo tcpdump tcp port 53 or udp port 53 -dd 生成socket匹配过滤指令序列(cBPF)
	dnsFilterRawInst := []bpf.RawInstruction{
		{Op: 0x28, Jt: 0, Jf: 0, K: 0x0000000c},
		{Op: 0x15, Jt: 0, Jf: 7, K: 0x000086dd},
		{Op: 0x30, Jt: 0, Jf: 0, K: 0x00000014},
		{Op: 0x15, Jt: 1, Jf: 0, K: 0x00000006},
		{Op: 0x15, Jt: 0, Jf: 16, K: 0x00000011},
		{Op: 0x28, Jt: 0, Jf: 0, K: 0x00000036},
		{Op: 0x15, Jt: 13, Jf: 0, K: 0x00000035},
		{Op: 0x28, Jt: 0, Jf: 0, K: 0x00000038},
		{Op: 0x15, Jt: 11, Jf: 12, K: 0x00000035},
		{Op: 0x15, Jt: 0, Jf: 11, K: 0x00000800},
		{Op: 0x30, Jt: 0, Jf: 0, K: 0x00000017},
		{Op: 0x15, Jt: 1, Jf: 0, K: 0x00000006},
		{Op: 0x15, Jt: 0, Jf: 8, K: 0x00000011},
		{Op: 0x28, Jt: 0, Jf: 0, K: 0x00000014},
		{Op: 0x45, Jt: 6, Jf: 0, K: 0x00001fff},
		{Op: 0xb1, Jt: 0, Jf: 0, K: 0x0000000e},
		{Op: 0x48, Jt: 0, Jf: 0, K: 0x0000000e},
		{Op: 0x15, Jt: 2, Jf: 0, K: 0x00000035},
		{Op: 0x48, Jt: 0, Jf: 0, K: 0x00000010},
		{Op: 0x15, Jt: 0, Jf: 1, K: 0x00000035},
		{Op: 0x6, Jt: 0, Jf: 0, K: 0x00040000},
		{Op: 0x6, Jt: 0, Jf: 0, K: 0x00000000},
	}
	err = h.SetBPF(dnsFilterRawInst)
	if err != nil {
		return nil, err
	}
	return h, nil
}

type DNSQAKey struct {
	TransactionID uint16 // DNS transaction ID
	IsUDP         bool   // UDP(true) TCP(false)
	IsV4          bool   // IPv4(true) IPv6(false)
	ClientPort    uint16 // Client Port
	ServerPort    uint16 // Server Port
	ClientIP      string
	ServerIP      string
}

type DNSPacketInfo struct {
	Key     DNSQAKey
	QR      bool // query(false) response(true)
	RCODE   uint8
	TS      time.Time
	Answers []layers.DNSResourceRecord
}

func ReadPacketInfoFromDNSParser(ts time.Time, dnsParser *DNSParser) (*DNSPacketInfo, error) {
	pinfo := DNSPacketInfo{
		TS: ts,
	}
	if dnsParser == nil {
		return nil, fmt.Errorf("*DnsParser: nil")
	}
	if dnsParser.dns != nil {
		pinfo.Key.TransactionID = dnsParser.dns.ID
		pinfo.QR = dnsParser.dns.QR
		pinfo.RCODE = uint8(dnsParser.dns.ResponseCode)
		pinfo.Answers = dnsParser.dns.Answers
	} else {
		return nil, fmt.Errorf("*DnsParser.dns: nil")
	}

	switch {
	case dnsParser.udp != nil:
		pinfo.Key.IsUDP = true
		pinfo.Key.ClientPort = uint16(dnsParser.udp.SrcPort)
		pinfo.Key.ServerPort = uint16(dnsParser.udp.DstPort)
	case dnsParser.tcp != nil:
		pinfo.Key.IsUDP = false
		pinfo.Key.ClientPort = uint16(dnsParser.udp.SrcPort)
		pinfo.Key.ServerPort = uint16(dnsParser.udp.DstPort)
	default:
		return nil, fmt.Errorf("*DnsParser.udp and *DnsParser.tcpS: nil")
	}

	if dnsParser.ipv4 != nil {
		pinfo.Key.IsV4 = true
		pinfo.Key.ClientIP = dnsParser.ipv4.SrcIP.String()
		pinfo.Key.ServerIP = dnsParser.ipv4.DstIP.String()
	} else if dnsParser.ipv6 != nil {
		pinfo.Key.IsV4 = false
		pinfo.Key.ClientIP = dnsParser.ipv6.SrcIP.String()
		pinfo.Key.ServerIP = dnsParser.ipv6.DstIP.String()
	}

	if pinfo.QR {
		pinfo.Key.ClientPort, pinfo.Key.ServerPort = pinfo.Key.ServerPort, pinfo.Key.ClientPort
		pinfo.Key.ClientIP, pinfo.Key.ServerIP = pinfo.Key.ServerIP, pinfo.Key.ClientIP
	}

	return &pinfo, nil
}

type DNSStats struct {
	TS        time.Time
	RCODE     uint8
	RespTime  time.Duration
	Timeout   bool
	Responded bool
}

type DNSStatsRecord struct {
	sync.Mutex
	statsMap          map[DNSQAKey]DNSStats
	gTag              map[string]string
	finishedStatsList [][2]interface{}
}

func (s *DNSStatsRecord) addDNSStats(packetInfo *DNSPacketInfo, dnsRecord *DNSRecord) {
	s.Lock()
	defer s.Unlock()
	if s.finishedStatsList == nil {
		s.finishedStatsList = make([][2]interface{}, 0)
	}

	stats, ok := s.statsMap[packetInfo.Key]

	if !ok {
		if !packetInfo.QR { // query
			s.statsMap[packetInfo.Key] = DNSStats{
				TS:        packetInfo.TS,
				Timeout:   false,
				Responded: false,
			}
		}
		return
	} else {
		if packetInfo.QR { // answer
			stats.Responded = true
			stats.RespTime = packetInfo.TS.Sub(stats.TS)
			stats.RCODE = packetInfo.RCODE
			stats.Timeout = false
			delete(s.statsMap, packetInfo.Key)
			if dnsRecord != nil {
				dnsRecord.addRecord(packetInfo)
			}
		} else {
			return
		}
	}
	s.finishedStatsList = append(s.finishedStatsList, [2]interface{}{packetInfo.Key, stats})
}

func (s *DNSStatsRecord) Dump() []inputs.Measurement {
	s.Lock()
	defer s.Unlock()
	m := []inputs.Measurement{}
	for _, kv := range s.finishedStatsList {
		k, ok := kv[0].(DNSQAKey)
		if !ok {
			l.Warn("type %s", reflect.TypeOf(k))
			continue
		}
		v, ok := kv[1].(DNSStats)
		if !ok {
			l.Warn("type %s", reflect.TypeOf(v))
			continue
		}
		m = append(m, s.Conv2M(k, v))
	}
	s.finishedStatsList = make([][2]interface{}, 0)
	return m
}

func (s *DNSStatsRecord) Conv2M(key DNSQAKey, stats DNSStats) *measurement {
	m := measurement{
		ts:     stats.TS,
		name:   inputName,
		tags:   map[string]string{},
		fields: map[string]interface{}{},
	}
	for k, v := range s.gTag {
		m.tags[k] = v
	}
	m.tags["src_ip"] = key.ClientIP
	m.tags["src_port"] = fmt.Sprintf("%d", key.ClientPort)
	m.tags["dst_ip"] = key.ServerIP
	m.tags["dst_port"] = fmt.Sprintf("%d", key.ServerPort)
	if key.IsUDP {
		m.tags["transport"] = "udp"
	} else {
		m.tags["transport"] = "tcp"
	}
	if key.IsV4 {
		m.tags["family"] = "IPv4"
	} else {
		m.tags["family"] = "IPv6"
	}
	m.fields["timeout"] = stats.Timeout
	m.fields["rcode"] = stats.RCODE
	m.fields["resp_time"] = stats.RespTime.Nanoseconds()
	l.Debug(m.fields, m.tags)
	return &m
}

const DNSTIMEOUT = time.Second * 5

func (s *DNSStatsRecord) CheckTimeout(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 15)
	for {
		select {
		case <-ticker.C:
			s.Lock()
			for k, v := range s.statsMap {
				if !v.Responded && time.Since(v.TS) > DNSTIMEOUT {
					v.Responded = true
					v.Timeout = true
					s.finishedStatsList = append(s.finishedStatsList, [2]interface{}{k, v})
				}
			}
			s.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

type DNSFlowTracer struct {
	dnsStats DNSStatsRecord
}

func (tracer *DNSFlowTracer) readPacket(ctx context.Context, tp *afpacket.TPacket, ch chan *DNSPacketInfo) {
	dnsParser := NewDNSParse()
	for {
		d, _, err := tp.ZeroCopyReadPacketData()
		if err != nil {
			l.Error(err)
			continue
		}

		err = dnsParser.DecodeLayers(d, &dnsParser.layers)
		if err != nil {
			l.Error(err)
			continue
		}
		pinfo, err := ReadPacketInfoFromDNSParser(time.Now(), &dnsParser)
		if err != nil {
			l.Error(err)
			continue
		}
		select {
		case <-ctx.Done():
			return
		case ch <- pinfo:
		default:
			l.Error("pinfoCh full")
		}
	}
}

func (tracer *DNSFlowTracer) Run(ctx context.Context, tp *afpacket.TPacket, gTag map[string]string,
	dnsRecord *DNSRecord, feedAddr string) {
	tracer.dnsStats.gTag = gTag
	mCh := make(chan []inputs.Measurement, 64)
	pInfoCh := make(chan *DNSPacketInfo, 64)

	go tracer.readPacket(ctx, tp, pInfoCh)
	go tracer.dnsStats.CheckTimeout(ctx)
	go func() {
		t := time.NewTicker(time.Second * 60)
		for {
			select {
			case <-t.C:
				mCh <- tracer.dnsStats.Dump()
			case pinfo := <-pInfoCh:
				serverIP := net.ParseIP(pinfo.Key.ServerIP)
				if serverIP == nil {
					continue
				} else if serverIP.IsLoopback() {
					continue
				}

				tracer.dnsStats.addDNSStats(pinfo, dnsRecord)
			case <-ctx.Done():
				return
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-mCh:
			if len(m) == 0 {
				l.Warn("dnsflow: no data")
			} else if err := dkfeed.FeedMeasurement(m, feedAddr); err != nil {
				l.Error(err)
			}
		}
	}
}

func NewDNSFlowTracer() DNSFlowTracer {
	return DNSFlowTracer{
		dnsStats: DNSStatsRecord{
			statsMap:          map[DNSQAKey]DNSStats{},
			finishedStatsList: [][2]interface{}{},
		},
	}
}
