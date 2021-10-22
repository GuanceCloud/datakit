package dns

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/afpacket"
	"github.com/google/gopacket/layers"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"

	"golang.org/x/net/bpf"
)

var l = logger.DefaultSLogger("net_ebpf")

func SetLogger(nl *logger.Logger) {
	l = nl
}

type DnsParse struct {
	*gopacket.DecodingLayerParser
	eth  *layers.Ethernet
	ipv4 *layers.IPv4
	ipv6 *layers.IPv6
	udp  *layers.UDP
	tcp  *layers.TCP
	dns  *layers.DNS
}

func NewDNSParse() DnsParse {
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

	return DnsParse{
		gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, l...),
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

type DNSRecord struct {
	sync.Mutex
	record map[string][2]interface{}
}

func (c *DNSRecord) LookupAddr(ip net.IP) string {
	c.Lock()
	defer c.Unlock()

	ipStr := ip.String()
	v, ok := c.record[ipStr]
	if !ok {
		return ""
	}

	if domian, ok := v[0].(string); ok {
		return domian
	} else {
		return ""
	}
}

func (c *DNSRecord) AddRecord(dnsRecord *layers.DNSResourceRecord) {
	c.Lock()
	defer c.Unlock()

	switch dnsRecord.Type {
	case layers.DNSTypeA, layers.DNSTypeAAAA:
		c.record[dnsRecord.IP.String()] = [2]interface{}{
			string(dnsRecord.Name),
			time.Now().Add(time.Second * time.Duration(dnsRecord.TTL)),
		}
	}
}

func (c *DNSRecord) Gather(ctx context.Context, tp *afpacket.TPacket) {
	defer tp.Close()
	dnsParse := NewDNSParse()
	layersDns := []gopacket.LayerType{}

	t := time.NewTicker(time.Millisecond * 20)
	for {
		select {
		case <-t.C:
			d, _, err := tp.ZeroCopyReadPacketData()
			if err != nil {
				l.Error(err)
				continue
			}

			err = dnsParse.DecodeLayers(d, &layersDns)
			if err != nil {
				l.Error(err)
				continue
			}

			for _, answer := range dnsParse.dns.Answers {
				switch answer.Type {
				case layers.DNSTypeA, layers.DNSTypeAAAA:
					c.AddRecord(&answer)
					l.Debugf("dst dns server %s", dnsParse.ipv4.DstIP.String())
					l.Debugf("%s %s %s %s %d\n", string(answer.Name), answer.Type.String(),
						answer.IP.String(), string(answer.CNAME), answer.TTL)
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func NewDNSRecord() *DNSRecord {
	return &DNSRecord{
		record: map[string][2]interface{}{},
	}
}
