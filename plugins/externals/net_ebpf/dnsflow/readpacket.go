//go:build (linux && ignore) || ebpf
// +build linux,ignore ebpf

package dnsflow

import (
	"fmt"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/afpacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/bpf"
)

type DNSParser struct {
	*gopacket.DecodingLayerParser
	layers     []gopacket.LayerType
	eth        *layers.Ethernet
	ipv4       *layers.IPv4
	ipv6       *layers.IPv6
	udp        *layers.UDP
	tcpSupport *tcpWithDNSSupport
	dns        *layers.DNS
}

func NewDNSParse() DNSParser {
	var eth layers.Ethernet
	var ipv4 layers.IPv4
	var ipv6 layers.IPv6
	var udp layers.UDP
	var tcpSupport tcpWithDNSSupport
	var dns layers.DNS

	l := []gopacket.DecodingLayer{
		&eth,
		&ipv4, &ipv6,
		&udp, &tcpSupport,
		&dns,
	}

	return DNSParser{
		gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, l...),
		[]gopacket.LayerType{},
		&eth,
		&ipv4, &ipv6,
		&udp, &tcpSupport,
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

type DNSStats struct {
	TS        time.Time
	RCODE     uint8
	RespTime  time.Duration
	Timeout   bool
	Responded bool
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
		return nil, fmt.Errorf("DnsParser: nil")
	}

	haveDNSLayer := false
	for _, layer := range dnsParser.layers {
		switch layer {
		case layers.LayerTypeUDP:
			pinfo.Key.IsUDP = true
			pinfo.Key.ClientPort = uint16(dnsParser.udp.SrcPort)
			pinfo.Key.ServerPort = uint16(dnsParser.udp.DstPort)
		case layers.LayerTypeTCP:
			pinfo.Key.IsUDP = false
			pinfo.Key.ClientPort = uint16(dnsParser.tcpSupport.tcp.SrcPort)
			pinfo.Key.ServerPort = uint16(dnsParser.tcpSupport.tcp.DstPort)
		case layers.LayerTypeIPv4:
			pinfo.Key.IsV4 = true
			pinfo.Key.ClientIP = dnsParser.ipv4.SrcIP.String()
			pinfo.Key.ServerIP = dnsParser.ipv4.DstIP.String()
		case layers.LayerTypeIPv6:
			pinfo.Key.IsV4 = false
			pinfo.Key.ClientIP = dnsParser.ipv6.SrcIP.String()
			pinfo.Key.ServerIP = dnsParser.ipv6.DstIP.String()
		case layers.LayerTypeDNS:
			pinfo.Key.TransactionID = dnsParser.dns.ID
			pinfo.QR = dnsParser.dns.QR
			pinfo.RCODE = uint8(dnsParser.dns.ResponseCode)
			pinfo.Answers = dnsParser.dns.Answers
			haveDNSLayer = true
		case gopacket.LayerTypeDecodeFailure, gopacket.LayerTypeFragment,
			gopacket.LayerTypePayload, gopacket.LayerTypeZero:
		default:
		}
	}

	if !haveDNSLayer {
		return nil, fmt.Errorf("no dns layer")
	}
	if pinfo.QR {
		pinfo.Key.ClientPort, pinfo.Key.ServerPort = pinfo.Key.ServerPort, pinfo.Key.ClientPort
		pinfo.Key.ClientIP, pinfo.Key.ServerIP = pinfo.Key.ServerIP, pinfo.Key.ClientIP
	}

	return &pinfo, nil
}
