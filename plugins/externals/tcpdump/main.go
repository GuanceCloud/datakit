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
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"golang.org/x/net/context/ctxhttp"
)

var (
	flagCfgFile  = flag.String("cfg-file", "", "json config file")
	flagLog      = flag.String("log", filepath.Join(datakit.InstallDir, "externals", "tcpdump.log"), "log file")
	flagLogLevel = flag.String("log-level", "info", "log file")
	flagDesc     = flag.String("desc", "", "description of the process, for debugging")

	l      *logger.Logger
	config *Config = new(Config)
)

func main() {
	// 配置参数
	flag.Parse()

	if *flagCfgFile == "" {
		panic("config missing")
	}

	if *flagDesc != "" {
		l = logger.SLogger("tcpdump-" + *flagDesc)
	} else {
		l = logger.SLogger("tcpdump")
	}

	l.Infof("log level: %s", *flagLogLevel)

	cfg, err := ParseConfig(*flagCfgFile)
	if err != nil {
		panic("config missing")
	}

	logger.SetGlobalRootLogger(*flagLog,
		*flagLogLevel,
		logger.OPT_ENC_CONSOLE|logger.OPT_SHORT_CALLER)

	cfg.NetPacket.run()
}

// ParseConfig 解析配置文件
func ParseConfig(fpath string) (*Config, error) {
	_, err := toml.DecodeFile(fpath, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

type Config struct {
	NetPacket *NetPacket `toml:"tcpdump"`
}

type NetPacket struct {
	Metric string
	// tcpdump.Tcpdump
	Device   []string `toml:"device"`
	Protocol []string `toml:"protocol"`
	WriteUrl string   `toml:"writeUrl"`
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

	p.Metric = "tcpdump"

	if p.WriteUrl == "" {
		p.WriteUrl = "http://0.0.0.0:9529/telegraf"
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
			p.ethernetType(dc)
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

func (p *NetPacket) ethernetType(dc string) {
	switch p.Eth.EthernetType {
	case layers.EthernetTypeIPv4:
		p.ParseData(dc)
	default:
		// todo
	}
}

// 构造数据
func (p *NetPacket) ParseData(dc string) {
	switch {
	case p.IPv4.Protocol == layers.IPProtocolTCP:
		p.tcpData(dc)
	case p.IPv4.Protocol == layers.IPProtocolUDP:
		p.udpData(dc)
	}
}

// tcp data
func (p *NetPacket) tcpData(dc string) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	src := fmt.Sprintf("%s:%s", p.SrcHost, p.TCP.SrcPort)
	dst := fmt.Sprintf("%s:%s", p.DstHost, p.TCP.DstPort)

	fields["src"] = src // 源ip
	fields["dst"] = dst // 目标ip
	tags["protocol"] = "tcp"
	tags["device"] = dc

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

	if err := WriteData(ptline, p.WriteUrl); err != nil {
		l.Errorf("err msg", err)
	}
}

// upd data
func (p *NetPacket) udpData(dc string) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	src := fmt.Sprintf("%s:%s", p.SrcHost, p.UDP.SrcPort)
	dst := fmt.Sprintf("%s:%s", p.DstHost, p.UDP.DstPort)

	fields["src"] = src // 源ip
	fields["dst"] = dst // 目标ip
	tags["protocol"] = "udp"
	tags["device"] = dc

	fields["dstMac"] = fmt.Sprintf("%s", p.Eth.DstMAC) // mac地址
	fields["srcMac"] = fmt.Sprintf("%s", p.Eth.SrcMAC) // mac地址

	fields["len"] = len(p.Payload)

	ptline, err := io.MakeMetric(p.Metric, tags, fields, time.Now())
	if err != nil {
		l.Errorf("new point failed: %s", err.Error())
	}

	if err := WriteData(ptline, p.WriteUrl); err != nil {
		l.Errorf("err msg", err)
	}
}

func (p *NetPacket) run() {
	l.Info("start monit tcpdump...")
	p.handle()
}

// dataway写入数据
func WriteData(data []byte, urlPath string) error {
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
	l.Debugf("[debug] %s %d", string(body), httpResp.StatusCode)
	return err
}
