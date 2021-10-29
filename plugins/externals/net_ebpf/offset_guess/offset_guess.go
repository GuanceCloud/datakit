// +build linux

package offset_guess

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/DataDog/ebpf"
	"github.com/DataDog/ebpf/manager"
	"golang.org/x/sys/unix"
)

const (
	CONN_L3_MASK = 0xFF // 0xFF
	CONN_L3_IPv4 = 0x00 // 0x00
	CONN_L3_IPv6 = 0x01 // 0x01

	CONN_L4_MASK = 0xFF00 // 0xFF00
	CONN_L4_TCP  = 0x0000 // 0x00 << 8
	CONN_L4_UDP  = 0x0100 // 0x01 << 8
	MAXOFFSET    = 2048
	MINSUCCESS   = 5
)

var (
	listen_ipv4     = "127.0.0.2"
	listen_ipv4_arr = [4]uint32{0, 0, 0, 0x0200007F}
)

type Conninfo struct {
	Saddr [4]uint32
	Daddr [4]uint32

	Sport uint16
	Dport uint16

	Meta uint32

	Rtt     uint32
	Rtt_var uint32
}

// GuessTCP guess the offset of the structure field, such as tcp_sock.srtt_us.
func GuessTCP(ebpfMapGuess *ebpf.Map, guessed *OffsetGuessC) (*OffsetGuessC, error) {
	ctx := context.Background()
	defer ctx.Done()

	tcp4ServerPort, err := runTCPServer(ctx, "tcp4", listen_ipv4)
	if err != nil {
		return nil, err
	}
	serverAddr := fmt.Sprintf("%s:%d", listen_ipv4, tcp4ServerPort)

	conninfo := Conninfo{
		Dport: tcp4ServerPort,
		Daddr: listen_ipv4_arr,
		Meta:  CONN_L4_TCP | CONN_L3_IPv4,
	}

	newStatus := newGuessStatus()

	newStatus.meta = _Ctype_uint(conninfo.Meta)

	if guessed != nil {
		copyOffset(guessed, &newStatus)
	}

	rtt_ok := 0
	rttvar_ok := 0
	inet_sport_ok := 0
	sk_dport_ok := 0
	for {
		conn, err := net.Dial("tcp4", serverAddr)
		if err != nil {
			if err := conn.Close(); err != nil {
				return nil, err
			}
			return nil, err
		}
		time.Sleep(time.Millisecond * 5)
		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			return nil, fmt.Errorf("conv conn to tcp conn")
		}
		if err := tcpConn.SetLinger(0); err != nil {
			return nil, err
		}

		sport, err := strconv.Atoi(strings.Split(tcpConn.LocalAddr().String(), ":")[1])
		if err != nil {
			return nil, err
		}
		conninfo.Sport = uint16(sport)
		connFile, err := tcpConn.File()
		if err != nil {
			return nil, err
		}
		tcpInfo, err := unix.GetsockoptTCPInfo(int(connFile.Fd()), syscall.SOL_TCP, syscall.TCP_INFO)
		if err != nil {
			return nil, err
		}

		conninfo.Rtt = tcpInfo.Rtt
		conninfo.Rtt_var = tcpInfo.Rttvar

		statusAct, err := readMapGuessStatus(ebpfMapGuess)
		if err != nil {
			return nil, err
		}
		if statusAct.state == 0 { // lost
			continue
		}

		if try_guess(statusAct, &conninfo, GUESS_INET_SPORT) {
			inet_sport_ok++
		} else {
			inet_sport_ok = 0
		}

		if try_guess(statusAct, &conninfo, GUESS_SK_DPORT) {
			sk_dport_ok++
		} else {
			sk_dport_ok = 0
		}

		if try_guess(statusAct, &conninfo, GUESS_TCP_SK_SRTT_US) {
			rtt_ok++
		} else {
			rtt_ok = 0
		}

		if try_guess(statusAct, &conninfo, GUESS_TCP_SK_MDEV_US) {
			rttvar_ok++
		} else {
			rttvar_ok = 0
		}

		if rtt_ok > MINSUCCESS && rttvar_ok > MINSUCCESS && inet_sport_ok > MINSUCCESS && sk_dport_ok > MINSUCCESS {
			newStatus = newGuessStatus()
			copyOffset(statusAct, &newStatus)
			return &newStatus, nil
		}

		if statusAct.offset_tcp_sk_srtt_us > MAXOFFSET ||
			statusAct.offset_tcp_sk_mdev_us > MAXOFFSET ||
			statusAct.offset_inet_sport > MAXOFFSET ||
			statusAct.offset_sk_dport > MAXOFFSET {
			break
		}

		newStatus = newGuessStatus()
		copyOffset(statusAct, &newStatus)
		if err = connFile.Close(); err != nil {
			return nil, err
		}
		if err = conn.Close(); err != nil {
			return nil, err
		}
		time.Sleep(time.Millisecond * 5)
		if err = updateMapGuessStatus(ebpfMapGuess, &newStatus); err != nil {
			return nil, err
		}
	}
	return nil, fmt.Errorf("failed")
}

func NewConstEditor(offsetGuess *OffsetGuessC) []manager.ConstantEditor {
	return []manager.ConstantEditor{
		{
			Name:  "offset_tcp_sk_srtt_us",
			Value: uint64(offsetGuess.offset_tcp_sk_srtt_us),
		},
		{
			Name:  "offset_tcp_sk_mdev_us",
			Value: uint64(offsetGuess.offset_tcp_sk_mdev_us),
		},
		{
			Name:  "offset_inet_sport",
			Value: uint64(offsetGuess.offset_inet_sport),
		},
		{
			Name:  "offset_sk_dport",
			Value: uint64(offsetGuess.offset_sk_dport),
		},
		{
			Name:  "offset_sk_num",
			Value: uint64(offsetGuess.offset_sk_num),
		},
	}
}

func runTCPServer(ctx context.Context, network, address string) (uint16, error) {
	netListen, err := net.Listen(network, address+":0")
	if err != nil {
		return 0, err
	}
	serverPort, err := strconv.Atoi(strings.Split(netListen.Addr().String(), ":")[1])
	if err != nil {
		return 0, err
	}

	go func() {
		<-ctx.Done()
		err := netListen.Close()
		l.Error(err)
	}()

	go func() {
		for {
			conn, err := netListen.Accept()
			if err != nil {
				return
			}
			if err = conn.Close(); err != nil {
				l.Error(err)
			}
		}
	}()

	return uint16(serverPort), nil
}
