package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	sysIO "io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"golang.org/x/net/context/ctxhttp"
)

var input chan []byte

func init() {
	input = make(chan []byte, 128)
}

// ParseConfig 解析配置文件
func ParseConfig(fpath string) (*NetPacket, error) {
	var c NetPacket
	_, err := toml.DecodeFile(fpath, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

type NetPacket struct {
	Metric string
	// tcpdump.Tcpdump
	Device   []string `toml:"device"`
	Protocol []string `toml:"protocol"`
	DataWay  string   `toml:"dataway"`
	SrcHost  string
	DstHost  string
	Eth      *layers.Ethernet
	ARP      *layers.ARP
	IPv4     *layers.IPv4
	IPv6     *layers.IPv6
	TCP      *layers.TCP
	UDP      *layers.UDP

	Payload []byte
}

func (p *NetPacket) handle() {
	var (
		device         []string      = p.Device
		snapshotLength int32         = 1024
		promiscuous    bool          = false
		timeout        time.Duration = 30 * time.Second
	)

	// 默认
	if len(device) == 0 {
		devices, err := pcap.FindAllDevs()
		if err != nil {
			log.Fatal(err)
		}

		device = make([]string, 1)

		device[0] = devices[0].Name
	}

	if p.Metric == "" {
		p.Metric = "tcpdump"
	}

	for _, dc := range device {
		handle, err := pcap.OpenLive(dc, snapshotLength, promiscuous, timeout)
		if err != nil {
			log.Fatal(err)
		}

		defer handle.Close()

		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			// 包解析
			p.parsePacketLayers(packet)
			// 构造数据
			p.ethernetType()
		}
	}
}

// 解析包
func (p *NetPacket) parsePacketLayers(packet gopacket.Packet) {
	for _, l := range packet.Layers() {
		switch l.LayerType() {
		case layers.LayerTypeEthernet:
			p.Eth, _ = l.(*layers.Ethernet)
		case layers.LayerTypeIPv4:
			p.IPv4, _ = l.(*layers.IPv4)
			p.SrcHost = fmt.Sprintf("%s", p.IPv4.SrcIP)
			p.DstHost = fmt.Sprintf("%s", p.IPv4.DstIP)
		case layers.LayerTypeTCP:
			p.TCP, _ = l.(*layers.TCP)
		case layers.LayerTypeUDP:
			p.UDP, _ = l.(*layers.UDP)
		}
	}

	// Check for errors
	if err := packet.ErrorLayer(); err != nil {
		l.Errorf("Error decoding some part of the packet: %v", err)
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

	ptline, err := io.MakeMetric(p.Metric, tags, fields, time.Now())
	if err != nil {
		l.Errorf("new point failed: %s", err.Error())
	}

	if err := WriteDataWay(ptline, p.DataWay); err != nil {
		l.Errorf("err msg", err)
	}
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

	ptline, err := io.MakeMetric(p.Metric, tags, fields, time.Now())
	if err != nil {
		l.Errorf("new point failed: %s", err.Error())
	}

	if err := WriteDataWay(ptline, p.DataWay); err != nil {
		l.Errorf("err msg", err)
	}
}

var (
	flagCfgStr = flag.String("cfg", "", "json config string")

	l *logger.Logger
)

func main() {
	dump, err := ParseConfig("./config.conf")
	if err != nil {
		panic("config(json string) missing")
	}

	logger.SetGlobalRootLogger("./tcpdump.log",
		"info",
		logger.OPT_ENC_CONSOLE|logger.OPT_SHORT_CALLER)

	l = logger.SLogger("tcpdump")

	l.Infof("log level: %s", "info")
	dump.run()
}

func (p *NetPacket) run() {
	l.Info("start monit oracle...")
	p.handle()

	go p.pushData()
}

// dataway写入数据
func WriteDataWay(data []byte, urlPath string) error {
	// dataway path
	ctx, _ := context.WithCancel(context.Background())
	httpReq, err := http.NewRequest("POST", urlPath, bytes.NewBuffer(data))

	if err != nil {
		l.Errorf("[error] %s", err.Error())
		return err
	}

	httpReq = httpReq.WithContext(ctx)
	httpReq.Header.Set("X-Datakit-UUID", "mockdata_test")

	tmctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	httpResp, err := ctxhttp.Do(tmctx, http.DefaultClient, httpReq)
	if err != nil {
		l.Errorf("[error] %s", err.Error())
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode/100 != 2 {
		scanner := bufio.NewScanner(sysIO.LimitReader(httpResp.Body, 1024*1024))
		line := ""
		if scanner.Scan() {
			line = scanner.Text()
		}
		err = fmt.Errorf("server returned HTTP status %s: %s", httpResp.Status, line)
	}
	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		l.Errorf("[error] read error %s", err.Error())
		return err
	}
	l.Errorf("[debug] %s %d", string(body), httpResp.StatusCode)
	return err
}

func (p *NetPacket) pushData() {
	for ch := range input {
		if err := WriteDataWay(ch, "http://172.16.0.12:32758/v1/write/metrics?token=tkn_caba81680c8c4fb6b773e95b162623fe"); err != nil {
			l.Errorf("err msg", err)
		}
	}
}
