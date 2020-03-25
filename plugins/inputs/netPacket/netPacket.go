package netPacket

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// Packet holds all layers information
type NetPacket struct {
	Device string // interface
	Filter string
	Count  int    // count packets (default: 1M)
	Type   string // filter type

	SrcHost string
	DstHost string

	// packet layers data
	Eth    *layers.Ethernet
	ARP    *layers.ARP
	IPv4   *layers.IPv4
	IPv6   *layers.IPv6
	TCP    *layers.TCP
	UDP    *layers.UDP
	ICMPv4 *layers.ICMPv4
	ICMPv6 *layers.ICMPv6

	Payload []byte
}

const sampleConfig = `
	# device  = ""
	# count = 1M
	# filter = tcp
`

const description = `netpack dump`

func (p *NetPacket) Description() string {
	return description
}

func (p *NetPacket) SampleConfig() string {
	return sampleConfig
}

func (p *NetPacket) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("netPacket", func() telegraf.Input { return &NetPacket{} })
}

func (p *NetPacket) Start(acc telegraf.Accumulator) error {
	go p.exec(acc)

	return nil
}

func (p *NetPacket) exec(acc telegraf.Accumulator) {
	var (
		device         string        = p.Device
		snapshotLength int32         = 1024
		promiscuous    bool          = false
		timeout        time.Duration = 30 * time.Second
	)

	// 默认
	if device == "" {
		devices, err := pcap.FindAllDevs()
		if err != nil {
			log.Fatal(err)
		}

		device = devices[0].Name
	}

	handle, err := pcap.OpenLive(device, snapshotLength, promiscuous, timeout)
	if err != nil {
		log.Fatal(err)
	}

	defer handle.Close()

	var filter string = p.Filter
	if filter != "" {
		err = handle.SetBPFFilter(filter)
		if err != nil {
			log.Fatal(err)
		}
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		// 包解析
		p.parsePacketLayers(packet)

		// 构造数据
		p.ethernetType(acc)
	}
}

// 解析包
func (p *NetPacket) parsePacketLayers(packet gopacket.Packet) {
	for _, l := range packet.Layers() {
		switch l.LayerType() {
		case layers.LayerTypeEthernet:
			p.Eth, _ = l.(*layers.Ethernet)
		case layers.LayerTypeARP:
			p.ARP, _ = l.(*layers.ARP)
			p.SrcHost = fmt.Sprintf("%s", p.ARP.SourceProtAddress)
			p.DstHost = fmt.Sprintf("%s", p.ARP.DstProtAddress)
		case layers.LayerTypeIPv4:
			p.IPv4, _ = l.(*layers.IPv4)
			p.SrcHost = fmt.Sprintf("%s", p.IPv4.SrcIP)
			p.DstHost = fmt.Sprintf("%s", p.IPv4.DstIP)
		case layers.LayerTypeIPv6:
			p.IPv6, _ = l.(*layers.IPv6)
			p.SrcHost = fmt.Sprintf("%s", p.IPv6.SrcIP)
			p.DstHost = fmt.Sprintf("%s", p.IPv6.DstIP)
		case layers.LayerTypeICMPv4:
			p.ICMPv4, _ = l.(*layers.ICMPv4)
		case layers.LayerTypeICMPv6:
			p.ICMPv6, _ = l.(*layers.ICMPv6)
		case layers.LayerTypeTCP:
			p.TCP, _ = l.(*layers.TCP)
		case layers.LayerTypeUDP:
			p.UDP, _ = l.(*layers.UDP)
		}
	}

	// Application
	applicationLayer := packet.ApplicationLayer()
	if applicationLayer != nil {
		p.Payload = applicationLayer.Payload()
	}

	// Check for errors
	if err := packet.ErrorLayer(); err != nil {
		//fmt.Println("Error decoding some part of the packet:", err)
		// todo
	}
}

func (p *NetPacket) ethernetType(acc telegraf.Accumulator) {
	switch p.Eth.EthernetType {
	case layers.EthernetTypeIPv4:
		p.ParseData(acc)
	case layers.EthernetTypeIPv6:
		p.ParseData(acc)
	case layers.EthernetTypeARP:
		p.arpData(acc)
	default:
		// todo
	}
}

// 构造数据
func (p *NetPacket) ParseData(acc telegraf.Accumulator) {
	switch {
	case p.IPv4.Protocol == layers.IPProtocolTCP:
		p.tcpData(acc)
	case p.IPv4.Protocol == layers.IPProtocolUDP:
		p.udpData(acc)
	case p.IPv4.Protocol == layers.IPProtocolICMPv4:
		p.icmpData(acc)
	}
}

// tcp data
func (p *NetPacket) tcpData(acc telegraf.Accumulator) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	src := fmt.Sprintf("%s:%s", p.SrcHost, p.TCP.SrcPort)
	dst := fmt.Sprintf("%s:%s", p.DstHost, p.TCP.DstPort)

	tags["src"] = src // 源ip
	tags["dst"] = dst // 目标ip
	tags["protocol"] = "tcp"

	fields["dstMac"] = fmt.Sprintf("%s", p.Eth.DstMAC) // mac地址
	fields["srcMac"] = fmt.Sprintf("%s", p.Eth.SrcMAC) // mac地址

	fields["len"] = len(p.Payload)
	fields["window"] = p.TCP.Window
	fields["seq"] = p.TCP.Seq
	fields["ack"] = p.TCP.Ack
	fields["fin"] = p.TCP.FIN
	fields["syn"] = p.TCP.SYN
	fields["rst"] = p.TCP.RST
	fields["psh"] = p.TCP.PSH
	fields["ugr"] = p.TCP.URG
	fields["ece"] = p.TCP.ECE
	fields["ns"] = p.TCP.NS
	// fields["playload"] = p.Payload

	acc.AddFields("tcpPacket", fields, tags)
}

// upd data
func (p *NetPacket) udpData(acc telegraf.Accumulator) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	src := fmt.Sprintf("%s:%s", p.SrcHost, p.UDP.SrcPort)
	dst := fmt.Sprintf("%s:%s", p.DstHost, p.UDP.DstPort)

	tags["src"] = src // 源ip
	tags["dst"] = dst // 目标ip
	tags["protocol"] = "udp"
	fields["dstMac"] = fmt.Sprintf("%s", p.Eth.DstMAC) // mac地址
	fields["srcMac"] = fmt.Sprintf("%s", p.Eth.SrcMAC) // mac地址

	fields["len"] = len(p.Payload)
	// fields["playload"] = p.Payload

	acc.AddFields("udpPacket", fields, tags)
}

// 构造数据
func (p *NetPacket) arpData(acc telegraf.Accumulator) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	tags["srcAdd"] = fmt.Sprintf("%s", p.ARP.SourceProtAddress) // 源ip
	tags["dstIP"] = fmt.Sprintf("%s", p.ARP.DstProtAddress)     // 目标ip
	tags["protocol"] = fmt.Sprintf("%s", p.ARP.Protocol)

	fields["hardwareAddr"] = string(net.HardwareAddr(p.ARP.SourceHwAddress))

	acc.AddFields("arpPacket", fields, tags)
}

func (p *NetPacket) icmpData(acc telegraf.Accumulator) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	tags["srcIP"] = p.SrcHost // 源ip
	tags["dstIP"] = p.DstHost // 目标ip
	tags["protocol"] = "icmp"

	fields["seq"] = p.ICMPv4.Seq
	fields["type"] = p.ICMPv4.TypeCode.String()

	acc.AddFields("icmpPacket", fields, tags)
}

func (p *NetPacket) httpStat(acc telegraf.Accumulator) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	tags["get"] = p.SrcHost  // 源ip
	tags["post"] = p.DstHost // 目标ip
	tags["protocol"] = "icmp"

	fields["seq"] = p.ICMPv4.Seq
	fields["type"] = p.ICMPv4.TypeCode.String()

	acc.AddFields("icmpPacket", fields, tags)
}

func (p *NetPacket) Stop() {

}
