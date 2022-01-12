//go:build (linux && ignore) || ebpf
// +build linux,ignore ebpf

package offset

import (
	"context"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/DataDog/ebpf"
	"github.com/DataDog/ebpf/manager"
	"golang.org/x/sys/unix"
)

const (
	ConnL3Mask = 0xFF // 0xFF
	ConnL3IPv4 = 0x00 // 0x00
	ConnL3IPv6 = 0x01 // 0x01

	ConnL4Mask = 0xFF00 // 0xFF00
	ConnL4TCP  = 0x0000 // 0x00 << 8
	ConnL4UDP  = 0x0100 // 0x01 << 8
	MAXOFFSET  = 2048
	MINSUCCESS = 5
)

var (
	listenIPv4    = "127.0.0.2"
	listenIPv4Arr = [4]uint32{0, 0, 0, 0x0200007F}
)

type Conninfo struct {
	Saddr [4]uint32
	Daddr [4]uint32

	Sport uint16
	Dport uint16

	Meta uint32

	NetNS uint32

	Rtt    uint32
	RttVar uint32
}

const minKernelVersionB16 = 0x0004000400000000

func NewConstEditor(offsetGuess *OffsetGuessC) []manager.ConstantEditor {
	kernelVersion, err := getLinuxKernelVesion()
	if err != nil {
		l.Error(err)
		kernelVersion = minKernelVersionB16
	}
	return []manager.ConstantEditor{
		{
			Name:  "kernel_version",
			Value: kernelVersion,
		},
		{
			Name:  "offset_sk_num",
			Value: uint64(offsetGuess.offset_sk_num),
		},
		{
			Name:  "offset_inet_sport",
			Value: uint64(offsetGuess.offset_inet_sport),
		},
		{
			Name:  "offset_sk_family",
			Value: uint64(offsetGuess.offset_sk_family),
		},
		{
			Name:  "offset_sk_rcv_saddr",
			Value: uint64(offsetGuess.offset_sk_rcv_saddr),
		},
		{
			Name:  "offset_sk_daddr",
			Value: uint64(offsetGuess.offset_sk_daddr),
		},
		{
			Name:  "offset_sk_v6_rcv_saddr",
			Value: uint64(offsetGuess.offset_sk_v6_rcv_saddr),
		},
		{
			Name:  "offset_sk_v6_daddr",
			Value: uint64(offsetGuess.offset_sk_v6_daddr),
		},
		{
			Name:  "offset_sk_dport",
			Value: uint64(offsetGuess.offset_sk_dport),
		},
		{
			Name:  "offset_tcp_sk_srtt_us",
			Value: uint64(offsetGuess.offset_tcp_sk_srtt_us),
		},
		{
			Name:  "offset_tcp_sk_mdev_us",
			Value: uint64(offsetGuess.offset_tcp_sk_mdev_us),
		},
		{
			Name:  "offset_flowi4_saddr",
			Value: uint64(offsetGuess.offset_flowi4_saddr),
		},
		{
			Name:  "offset_flowi4_daddr",
			Value: uint64(offsetGuess.offset_flowi4_daddr),
		},
		{
			Name:  "offset_flowi4_sport",
			Value: uint64(offsetGuess.offset_flowi4_sport),
		},
		{
			Name:  "offset_flowi4_dport",
			Value: uint64(offsetGuess.offset_flowi4_dport),
		},
		{
			Name:  "offset_flowi6_saddr",
			Value: uint64(offsetGuess.offset_flowi6_saddr),
		},
		{
			Name:  "offset_flowi6_daddr",
			Value: uint64(offsetGuess.offset_flowi6_daddr),
		},
		{
			Name:  "offset_flowi6_sport",
			Value: uint64(offsetGuess.offset_flowi6_sport),
		},
		{
			Name:  "offset_flowi6_dport",
			Value: uint64(offsetGuess.offset_flowi6_dport),
		},
		{
			Name:  "offset_sk_net",
			Value: uint64(offsetGuess.offset_sk_net),
		},
		{
			Name:  "offset_ns_common_inum",
			Value: uint64(offsetGuess.offset_ns_common_inum),
		},
		{
			Name:  "offset_socket_sk",
			Value: uint64(offsetGuess.offset_socket_sk),
		},
	}
}

// GuessOffset guess the offset of the structure field, such as tcp_sock.srtt_us.
func GuessOffset(ebpfMapGuess *ebpf.Map, guessed *OffsetGuessC) (*OffsetGuessC, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tcp4ServerPort, err := runTCPServer(ctx, "tcp4", listenIPv4)
	if err != nil {
		return nil, err
	}

	udp4ServerPort, err := runUDPServer(ctx, "udp4", listenIPv4)
	if err != nil {
		return nil, err
	}

	serverAddr := fmt.Sprintf("%s:%d", listenIPv4, tcp4ServerPort)

	serverAddrUDP := fmt.Sprintf("%s:%d", listenIPv4, udp4ServerPort)

	var s syscall.Stat_t
	var netns uint32 = 0
	if err := syscall.Stat("/proc/self/ns/net", &s); err != nil {
		l.Error(err)
	} else {
		netns = uint32(s.Ino)
	}
	conninfo := Conninfo{
		Dport: tcp4ServerPort,
		Daddr: listenIPv4Arr,
		Meta:  ConnL4TCP | ConnL3IPv4,
		NetNS: netns,
	}

	conninfoUDP := Conninfo{
		Dport: udp4ServerPort,
		Daddr: listenIPv4Arr,
		Meta:  ConnL4UDP | ConnL3IPv4,
	}

	daddr6, ip6 := generateRandomIPv6Address()
	conninfo6 := Conninfo{
		Dport: 57391,
		Daddr: daddr6,
		Meta:  ConnL3IPv6 | ConnL4TCP,
	}
	serverAddr6 := fmt.Sprintf("[%s]:%d", ip6.String(), conninfo6.Dport)

	status := newGuessStatus()

	status.meta = _Ctype_uint(conninfo.Meta)

	if guessed != nil {
		copyOffset(guessed, &status)
	}
	status.pid_tgid = _Ctype_ulonglong(uint64(unix.Getpid())<<32 | uint64(unix.Gettid()))
	if err := updateMapGuessStatus(ebpfMapGuess, &status); err != nil {
		return nil, err
	}

	offsetCheck := OffsetCheck{}
	for {
		err := guessTCP4(serverAddr, conninfo, ebpfMapGuess, &offsetCheck, &status)
		if err != nil {
			return nil, err
		}
		err = guessTCP6(serverAddr6, conninfo6, ebpfMapGuess, &offsetCheck, &status)
		if err != nil {
			return nil, err
		}
		err = guessUDP4(serverAddrUDP, conninfoUDP, ebpfMapGuess, &offsetCheck, &status)
		if err != nil {
			return nil, err
		}

		if offsetCheck.tcpSkSrttUsOk > MINSUCCESS && offsetCheck.tcpSkMdevUsOk > MINSUCCESS &&
			offsetCheck.inetSportOk > MINSUCCESS && offsetCheck.skDportOk > MINSUCCESS &&
			offsetCheck.skDaddrOk > MINSUCCESS && offsetCheck.skV6DaddrOk > MINSUCCESS &&
			offsetCheck.skFamilyOk > MINSUCCESS && offsetCheck.flowi4DaddrOk > MINSUCCESS &&
			offsetCheck.flowi4DportOk > MINSUCCESS && offsetCheck.flowi4SaddrOk > MINSUCCESS &&
			offsetCheck.netnsInumOk > MINSUCCESS && offsetCheck.sknetOk > MINSUCCESS &&
			offsetCheck.socketSkOK > MINSUCCESS {
			newStatus := newGuessStatus()
			copyOffset(&status, &newStatus)
			if newStatus.offset_flowi4_daddr > newStatus.offset_flowi4_saddr {
				// + sizeof(flowi_common)
				newStatus.offset_flowi6_daddr = newStatus.offset_flowi4_saddr
			} else {
				newStatus.offset_flowi6_daddr = newStatus.offset_flowi4_daddr
			}
			newStatus.offset_flowi6_saddr = newStatus.offset_flowi6_daddr + 16 // +128bit
			newStatus.offset_flowi6_dport = newStatus.offset_flowi6_daddr + 36 // +256bit + 32bit
			newStatus.offset_flowi6_sport = newStatus.offset_flowi6_daddr + 38 // +256bit + 32bit +16bit
			return &newStatus, nil
		}
	}
}

func guessTCP4(serverAddr string, conninfo Conninfo, ebpfMapGuess *ebpf.Map,
	offsetCheck *OffsetCheck, status *OffsetGuessC) error {
	status.conn_type = ConnL3IPv4 | ConnL4TCP
	if err := updateMapGuessStatus(ebpfMapGuess, status); err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 17)
	conn, err := net.Dial("tcp4", serverAddr)
	if err != nil {
		return err
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("conv conn to tcp conn")
	}
	if err := tcpConn.SetLinger(0); err != nil {
		return err
	}

	sport, err := strconv.Atoi(strings.Split(tcpConn.LocalAddr().String(), ":")[1])
	if err != nil {
		return err
	}
	conninfo.Sport = uint16(sport)
	connFile, err := tcpConn.File()
	if err != nil {
		return err
	}

	tcpInfo, err := unix.GetsockoptTCPInfo(int(connFile.Fd()), syscall.SOL_TCP, syscall.TCP_INFO)
	if err != nil {
		return err
	}
	conninfo.Rtt = tcpInfo.Rtt
	conninfo.RttVar = tcpInfo.Rttvar

	if err = connFile.Close(); err != nil {
		return err
	}
	if err = conn.Close(); err != nil {
		return err
	}

	statusAct, err := readMapGuessStatus(ebpfMapGuess)
	if err != nil {
		return err
	}
	if statusAct.state == 0 { // lost
		l.Error(statusAct.pid_tgid)
		time.Sleep(time.Millisecond * 20)
		return nil
	}

	tryGuess(statusAct, offsetCheck, &conninfo, GUESS_INET_SPORT)
	tryGuess(statusAct, offsetCheck, &conninfo, GUESS_SK_DPORT)
	tryGuess(statusAct, offsetCheck, &conninfo, GUESS_TCP_SK_SRTT_US)
	tryGuess(statusAct, offsetCheck, &conninfo, GUESS_TCP_SK_MDEV_US)
	tryGuess(statusAct, offsetCheck, &conninfo, GUESS_SK_DADDR)
	tryGuess(statusAct, offsetCheck, &conninfo, GUESS_SK_FAMILY)
	tryGuess(statusAct, offsetCheck, &conninfo, GUESS_NS_COMMON_INUM)
	tryGuess(statusAct, offsetCheck, &conninfo, GUESS_SOCKET_SK)

	copyOffset(statusAct, status)
	if status.offset_tcp_sk_srtt_us > MAXOFFSET ||
		status.offset_tcp_sk_mdev_us > MAXOFFSET ||
		status.offset_inet_sport > MAXOFFSET ||
		status.offset_sk_dport > MAXOFFSET ||
		status.offset_socket_sk > MAXOFFSET ||
		status.offset_sk_daddr > MAXOFFSET ||
		status.offset_sk_family > MAXOFFSET ||
		status.offset_sk_net > MAXOFFSET ||
		status.offset_ns_common_inum > MAXOFFSET {
		l.Error(status)
		return fmt.Errorf("guess tcp4: offset > MAXOFFSET")
	}

	if err = updateMapGuessStatus(ebpfMapGuess, status); err != nil {
		return err
	}

	return nil
}

func guessTCP6(serverAddr6 string, conninfo6 Conninfo, ebpfMapGuess *ebpf.Map,
	offsetCheck *OffsetCheck, status *OffsetGuessC) error {
	if offsetCheck.skV6DaddrOk > MINSUCCESS && offsetCheck.skFamilyOk > MINSUCCESS {
		return nil
	}
	status.conn_type = ConnL3IPv6 | ConnL4TCP
	if err := updateMapGuessStatus(ebpfMapGuess, status); err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 10)

	if conn, err := net.DialTimeout("tcp6", serverAddr6, time.Millisecond*10); err == nil {
		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			return fmt.Errorf("conv conn to tcp conn")
		}
		if err := tcpConn.SetLinger(0); err != nil {
			return err
		}
		if err := conn.Close(); err != nil {
			return err
		}
	}

	var err error
	statusAct, err := readMapGuessStatus(ebpfMapGuess)
	if err != nil {
		return err
	}

	if statusAct.state == 0 { // lost
		l.Error(status.pid_tgid)
		time.Sleep(time.Millisecond * 20)
		return nil
	}
	tryGuess(statusAct, offsetCheck, &conninfo6, GUESS_SK_V6_DADDR)
	tryGuess(statusAct, offsetCheck, &conninfo6, GUESS_SK_FAMILY)
	copyOffset(statusAct, status)
	if status.offset_sk_v6_daddr > MAXOFFSET ||
		status.offset_sk_family > MAXOFFSET {
		l.Error(status)
		return fmt.Errorf("guesss tcp6: offset > MAXOFFSET")
	}
	if err = updateMapGuessStatus(ebpfMapGuess, status); err != nil {
		return err
	}

	return nil
}

func guessUDP4(serverAddrUDP string, conninfoUDP Conninfo, ebpfMapGuess *ebpf.Map,
	offsetCheck *OffsetCheck, status *OffsetGuessC) error {
	if offsetCheck.flowi4DaddrOk > MINSUCCESS &&
		offsetCheck.flowi4SaddrOk > MINSUCCESS &&
		offsetCheck.flowi4DportOk > MINSUCCESS {
		return nil
	}
	status.conn_type = ConnL3IPv4 | ConnL4UDP
	if err := updateMapGuessStatus(ebpfMapGuess, status); err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 10)
	conn, err := net.Dial("udp4", serverAddrUDP)
	if err != nil {
		return err
	}

	srcIP := net.ParseIP(strings.Split(conn.LocalAddr().String(), ":")[0])
	if srcIP != nil {
		srcIP = srcIP.To16()
	}
	if srcIP == nil {
		return fmt.Errorf("src ip: %s", conn.LocalAddr().String())
	}

	ip4arr := *(*[4]uint32)(unsafe.Pointer(&srcIP[0])) //nolint:gosec
	conninfoUDP.Saddr = [4]uint32{0, 0, 0, ip4arr[3]}
	_, err = conn.Write([]byte("guess flowi4"))
	if err != nil {
		return err
	}
	if err = conn.Close(); err != nil {
		return err
	}
	statusAct, err := readMapGuessStatus(ebpfMapGuess)
	if err != nil {
		return err
	}
	if statusAct.state == 0 { // lost
		time.Sleep(time.Millisecond * 20)
		return nil
	}

	tryGuess(statusAct, offsetCheck, &conninfoUDP, GUESS_FLOWI4_DADDR)
	tryGuess(statusAct, offsetCheck, &conninfoUDP, GUESS_FLOWI4_SADDR)
	tryGuess(statusAct, offsetCheck, &conninfoUDP, GUESS_FLOWI4_DPORT)
	copyOffset(statusAct, status)
	if status.offset_flowi4_daddr > MAXOFFSET ||
		status.offset_flowi4_saddr > MAXOFFSET ||
		status.offset_flowi4_dport > MAXOFFSET {
		l.Error(status)
		return fmt.Errorf("guess upd4: offset > MAXOFFSET")
	}
	if err = updateMapGuessStatus(ebpfMapGuess, status); err != nil {
		return err
	}
	return nil
}

func runTCPServer(ctx context.Context, network, address string) (uint16, error) {
	netListen, err := net.Listen(network, address+":0")
	if err != nil {
		return 0, err
	}

	addr := netListen.Addr().String()
	var serverPort int
	if addr[:1] == "[" {
		serverPort, err = strconv.Atoi(strings.Split(addr, "]")[1][1:])
	} else {
		serverPort, err = strconv.Atoi(strings.Split(addr, ":")[1])
	}
	if err != nil {
		return 0, err
	}

	go func() {
		<-ctx.Done()
		if err := netListen.Close(); err != nil {
			l.Error(err)
		}
	}()

	go func() {
		for {
			conn, err := netListen.Accept()
			if err != nil {
				return
			}
			if err = conn.Close(); err != nil {
				l.Error(err)
				return
			}
		}
	}()

	return uint16(serverPort), nil
}

func runUDPServer(ctx context.Context, network, addr string) (uint16, error) {
	netListen, err := net.ListenPacket(network, addr+":0")
	if err != nil {
		return 0, err
	}
	localAddr := netListen.LocalAddr().String()
	var serverPort int
	if localAddr[:1] == "[" {
		serverPort, err = strconv.Atoi(strings.Split(localAddr, "]")[1][1:])
	} else {
		serverPort, err = strconv.Atoi(strings.Split(localAddr, ":")[1])
	}
	if err != nil {
		return 0, err
	}
	go func() {
		<-ctx.Done()
		if err := netListen.Close(); err != nil {
			l.Error(err)
		}
	}()

	go func() {
		for {
			p := []byte{}
			err := netListen.SetReadDeadline(time.Now().Add(time.Microsecond * 20))
			if err != nil {
				l.Error(err)
			}
			_, _, err = netListen.ReadFrom(p)
			if err != nil && !os.IsTimeout(err) {
				return
			}
		}
	}()

	return uint16(serverPort), nil
}
