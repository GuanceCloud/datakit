// +build ignore

// +build linux,amd64

package netPacket

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/tcpdump"
	"google.golang.org/grpc"
)

func (p *NetPacket) exec() {
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
		case layers.LayerTypeTCP:
			p.TCP, _ = l.(*layers.TCP)
		case layers.LayerTypeUDP:
			p.UDP, _ = l.(*layers.UDP)
		}
	}

	// Check for errors
	if err := packet.ErrorLayer(); err != nil {
		//fmt.Println("Error decoding some part of the packet:", err)
		// todo
	}
}

func (p *NetPacket) ethernetType() {
	switch p.Eth.EthernetType {
	case layers.EthernetTypeIPv4:
		p.ParseData()
	default:
		// todo
	}
}

// 构造数据
func (p *NetPacket) ParseData() {
	switch {
	case p.IPv4.Protocol == layers.IPProtocolTCP:
		p.tcpData()
	case p.IPv4.Protocol == layers.IPProtocolUDP:
		p.udpData()
	}
}

// tcp data
func (p *NetPacket) tcpData() {
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
func (p *NetPacket) udpData() {
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

var (
	flagCfgStr = flag.String("cfg", "", "json config string")
	flagDesc   = flag.String("desc", "", "description of the process, for debugging")

	flagRPCServer = flag.String("rpc-server", "unix://"+datakit.GRPCDomainSock, "gRPC server")
	flagLog       = flag.String("log", filepath.Join(datakit.InstallDir, "externals", "oraclemonitor.log"), "log file")
	flagLogLevel  = flag.String("log-level", "info", "log file")

	l      *logger.Logger
	rpcCli io.DataKitClient
)

type NetPacket struct {
	tcpdump.Tcpdump
	SrcHost string
	DstHost string

	// packet layers data
	Eth *layers.Ethernet
	// ARP    *layers.ARP
	// IPv4   *layers.IPv4
	// IPv6   *layers.IPv6
	TCP *layers.TCP
	UDP *layers.UDP
	// ICMPv4 *layers.ICMPv4
	// ICMPv6 *layers.ICMPv6
}

func main() {
	flag.Parse()

	if *flagCfgStr == "" {
		panic("config(json string) missing")
	}

	logger.SetGlobalRootLogger(*flagLog,
		*flagLogLevel,
		logger.OPT_ENC_CONSOLE|logger.OPT_SHORT_CALLER)

	if *flagDesc != "" {
		l = logger.SLogger("oraclemonitor-" + *flagDesc)
	} else {
		l = logger.SLogger("oraclemonitor")
	}

	l.Infof("log level: %s", *flagLogLevel)

	var dump NetPacket
	cfg, err := base64.StdEncoding.DecodeString(*flagCfgStr)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(cfg, &dump); err != nil {
		l.Errorf("failed to parse json `%s': %s", *flagCfgStr, err)
		return
	}

	l.Infof("gRPC dial %s...", *flagRPCServer)
	conn, err := grpc.Dial(*flagRPCServer, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second*5))
	if err != nil {
		l.Fatalf("connect RCP failed: %s", err)
	}

	l.Infof("gRPC connect %s ok", *flagRPCServer)
	defer conn.Close()

	rpcCli = io.NewDataKitClient(conn)

	dump.run()
}

func (dump *Netdump) run() {

	l.Info("start monit oracle...")

}

func handleResponse(m *monitor, k string, response []map[string]interface{}) error {
	lines := [][]byte{}

	for _, item := range response {

		tags := map[string]string{}

		tags["oracle_server"] = m.Server
		tags["oracle_port"] = m.Port
		tags["instance_id"] = m.InstanceId
		tags["instance_desc"] = m.Desc
		tags["product"] = "oracle"
		tags["host"] = m.Host
		tags["type"] = k

		if tagKeys, ok := tagsMap[k]; ok {
			for _, tagKey := range tagKeys {
				tags[tagKey] = String(item[tagKey])
				delete(item, tagKey)
			}
		}

		// add user-added tags
		// XXX: this may overwrite tags within @tags
		for k, v := range m.Tags {
			tags[k] = v
		}

		ptline, err := io.MakeMetric(m.Metric, tags, item, time.Now())
		if err != nil {
			l.Errorf("new point failed: %s", err.Error())
			return err
		}

		lines = append(lines, ptline)
		l.Debugf("add point %+#v", string(ptline))
	}

	if len(lines) == 0 {
		l.Debugf("no metric collected on %s", k)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := rpcCli.Send(ctx, &io.Request{
		Lines:     bytes.Join(lines, []byte("\n")),
		Precision: "ns",
		Name:      "oraclemonitor",
	})
	if err != nil {
		l.Errorf("feed error: %s", err.Error())
		return err
	}

	l.Debugf("feed %d points, error: %s", r.GetPoints(), r.GetErr())

	return nil
}
