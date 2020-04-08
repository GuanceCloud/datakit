package scanport

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

// scan
type Scan struct {
	MinPort  int
	MaxPort  int
	Target   string
	lport    int
	rport    int
	laddr    net.IP
	raddr    net.IP
	ifName   string
	network  string
	forceV4  bool
	forceV6  bool
	connScan bool
	ip       gopacket.NetworkLayer
}

const sampleConfig = `
# scan port,
# minPort = 1
# maxPort = 1000
# target  = "www.dataflux.com"
 `
const description = `scan network port`

func (s *Scan) Description() string {
	return description
}
func (s *Scan) SampleConfig() string {
	return sampleConfig
}
func (s *Scan) Gather(acc telegraf.Accumulator) error {
	s.exec(acc)
	return nil
}
func (s *Scan) exec(acc telegraf.Accumulator) {
	if s.IsCIDR() {
		fmt.Errorf("it doesn't support CIDR")
		return
	}
	s.setIP()
	openPorts := s.Run()
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	for _, p := range openPorts {
		fields["port"] = p.port
		tags["target"] = s.Target       //域名
		tags["addr"] = s.raddr.String() //ip
		tags["minPort"] = fmt.Sprintf("%d", s.MinPort)
		tags["maxPort"] = fmt.Sprintf("%d", s.MaxPort)
		acc.AddFields("scan_port", fields, tags, p.t)
	}
}
func init() {
	inputs.Add("scanport", func() telegraf.Input { return &Scan{} })
}

// IsCIDR checks the target if it's CIDR
func (s *Scan) IsCIDR() bool {
	_, _, err := net.ParseCIDR(s.Target)
	if err != nil {
		return false
	}
	return true
}
func isIPv4(ip net.IP) bool {
	return len(ip.To4()) == net.IPv4len
}
func isIPv6(ip net.IP) bool {
	if r := strings.Index(ip.String(), ":"); r != -1 {
		return true
	}
	return false
}
func (s *Scan) setIP() error {
	ips, err := net.LookupIP(s.Target)
	if err != nil {
		return err
	}
	for _, ip := range ips {
		if isIPv4(ip) && !s.forceV6 {
			s.raddr = ip
			s.network = "ip4"
			break
		} else if isIPv6(ip) && !s.forceV4 {
			s.raddr = ip
			s.network = "ip6"
			break
		}
	}
	return nil
}

// Run tries to scan wide range ports (TCP)
func (s *Scan) Run() []point {
	var (
		openPorts []point
		err       error
	)
	if s.MinPort != s.MaxPort {
		fmt.Printf("Scan %s (%s) TCP ports %d-%d\n", s.Target, s.raddr, s.MinPort, s.MaxPort)
	} else {
		fmt.Printf("Scan %s (%s) TCP port %d\n", s.Target, s.raddr, s.MinPort)
	}
	tStart := time.Now()
	if s.connScan {
		openPorts = s.tcpConnScan()
	} else {
		openPorts, err = s.tcpSYNScan()
	}
	if err != nil {
		println(err.Error())
		return nil
	}
	if len(openPorts) == 0 {
		println("there isn't any opened port")
	} else {
		println("")
		elapsed := fmt.Sprintf("%.3f seconds", time.Since(tStart).Seconds())
		println("Scan done:", len(openPorts), "opened port(s) found in", elapsed)
	}
	return openPorts
}
func (s *Scan) packetDataTCP(rport int) (error, []byte) {
	tcp := &layers.TCP{
		SrcPort: layers.TCPPort(s.lport),
		DstPort: layers.TCPPort(rport),
		Seq:     1,
		SYN:     true,
		Window:  15000,
	}
	tcp.SetNetworkLayerForChecksum(s.ip)
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	if err := gopacket.SerializeLayers(buf, opts, tcp); err != nil {
		return err, []byte{}
	}
	return nil, buf.Bytes()
}
func (s *Scan) setProto(proto string) error {
	var netProto = fmt.Sprintf("%s:%s", s.network, proto)
	switch netProto {
	case "ip4:tcp":
		s.ip = &layers.IPv4{
			SrcIP:    s.laddr,
			DstIP:    s.raddr,
			Protocol: layers.IPProtocolTCP,
		}
	case "ip4:udp":
		s.ip = &layers.IPv4{
			SrcIP:    s.laddr,
			DstIP:    s.raddr,
			Protocol: layers.IPProtocolUDP,
		}
	case "ip6:tcp":
		s.ip = &layers.IPv6{
			SrcIP:      s.laddr,
			DstIP:      s.raddr,
			NextHeader: layers.IPProtocolTCP,
		}
	case "ip6:udp":
		s.ip = &layers.IPv6{
			SrcIP:      s.laddr,
			DstIP:      s.raddr,
			NextHeader: layers.IPProtocolUDP,
		}
	}
	return nil
}
func (s *Scan) setLocalNet() error {
	conn, err := net.Dial("udp", net.JoinHostPort(s.raddr.String(), "80"))
	if err != nil {
		return err
	}
	defer conn.Close()
	if lAddr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		s.laddr = lAddr.IP
		s.lport = lAddr.Port
	} else {
		return fmt.Errorf("can not find local address/port")
	}
	ifs, _ := net.Interfaces()
	for _, i := range ifs {
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			ip, _, _ := net.ParseCIDR(addr.String())
			if ip.String() == s.laddr.String() {
				s.ifName = i.Name
				break
			}
		}
	}
	return nil
}
func (s *Scan) tcpSYNScan() ([]point, error) {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	if err = s.setLocalNet(); err != nil {
		return nil, fmt.Errorf("source IP address not configured")
	}
	if err = s.setProto("tcp"); err != nil {
		println(err.Error())
	}
	go func() {
		if err := s.sendTCPSYN(); err != nil {
			println(err.Error())
		}
		cancel()
	}()
	openPorts, err := s.pCapture(ctx)
	return openPorts, err
}
func (s *Scan) pCapture(ctx context.Context) ([]point, error) {
	var (
		tcp       *layers.TCP
		timeout   = 100 * time.Nanosecond
		openPorts []point
		ok        bool
	)
	handle, err := pcap.OpenLive(s.ifName, 6*1024, false, timeout)
	if err != nil {
		return nil, err
	}
	defer handle.Close()
	filter := "tcp and src host " + s.raddr.String()
	if err := handle.SetBPFFilter(filter); err != nil {
		return nil, err
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case packet := <-packetSource.Packets():
			tcpLayer := packet.Layer(layers.LayerTypeTCP)
			if tcpLayer != nil {
				if tcp, ok = tcpLayer.(*layers.TCP); !ok {
					continue
				}
				if tcp.SYN && tcp.ACK {
					p, _ := strconv.Atoi(fmt.Sprintf("%d", tcp.SrcPort))
					pt := point{
						port: p,
						t:    time.Now(),
					}
					// ports = append(ports, pt)
					openPorts = append(openPorts, pt)
				}
				continue
			}
		}
	}
	return openPorts, nil
}
func (s *Scan) sendTCPSYN() error {
	var (
		buf  []byte
		err  error
		conn net.PacketConn
	)
	if s.network != "ip6" {
		conn, err = net.ListenPacket(s.network+":tcp", "0.0.0.0")
	} else {
		conn, err = net.ListenPacket(s.network+":tcp", "::")
	}
	if err != nil {
		return err
	}
	for i := s.MinPort; i <= s.MaxPort; i++ {
		if err, buf = s.packetDataTCP(i); err != nil {
			return err
		}
		if _, err := conn.WriteTo(buf, &net.IPAddr{IP: s.raddr}); err != nil {
			println(err.Error())
		}
		time.Sleep(5 * time.Millisecond)
	}
	time.Sleep(1 * time.Second)
	return nil
}

type point struct {
	port int
	t    time.Time
}

// tcpConnScan tries to scan a single host
func (s *Scan) tcpConnScan() []point {
	var wg sync.WaitGroup
	var ports = make([]point, 0)
	for i := s.MinPort; i <= s.MaxPort; i++ {
		wg.Add(1)
		go func(i int) {
			for {
				host := net.JoinHostPort(s.raddr.String(), fmt.Sprintf("%d", i))
				conn, err := net.DialTimeout("tcp", host, 2*time.Second)
				if err != nil {
					if strings.Contains(err.Error(), "too many open files") {
						// random back-off
						time.Sleep(time.Duration(10+rand.Int31n(30)) * time.Millisecond)
						continue
					}
					wg.Done()
					return
				}
				conn.Close()
				break
			}
			p := point{
				port: i,
				t:    time.Now(),
			}
			ports = append(ports, p)
			wg.Done()
		}(i)
	}
	wg.Wait()
	// sort.Ints(ports)
	return ports
}
func uniqSlice(s []int) []int {
	m := map[int]bool{}
	r := []int{}
	for _, v := range s {
		if _, ok := m[v]; !ok {
			m[v] = true
			r = append(r, v)
		}
	}
	return r
}
