// +build linux, ebpf

package offset

// #include "../c/offset_guess/offset.h"
import "C"

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/DataDog/ebpf"
	"github.com/DataDog/ebpf/manager"
	"github.com/shirou/gopsutil/host"
	"golang.org/x/sys/unix"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/c"
)

type OffsetGuessC C.struct_offset_guess

//nolint:structcheck
type OffsetCheck struct {
	skNumOk           uint64
	inetSportOk       uint64
	skFamilyOk        uint64
	skRcvSaddrOk      uint64
	skDaddrOk         uint64
	skV6RcvSaddrOk    uint64
	skV6DaddrOk       uint64
	skDportOk         uint64
	tcpSkSrttUsOk     uint64
	tcpSkMdevUsOk     uint64
	flowi4SaddrOk     uint64
	flowi4DaddrOk     uint64
	flowi4SportOk     uint64
	flowi4DportOk     uint64
	flowi6SaddrOk     uint64
	flowi6DaddrOk     uint64
	flowi6SportOk     uint64
	flowi6DportOk     uint64
	skaddrSinPortOk   uint64
	skaddr6Sin6PortOk uint64
	sknetOk           uint64
	netnsInumOk       uint64
}

const PROCNAMLEN = 16 // Maximum length of process name

//nolint:stylecheck
const (
	GUESS_SK_NUM = iota + 1
	GUESS_INET_SPORT
	GUESS_SK_FAMILY
	GUESS_SK_RCV_SADDR
	GUESS_SK_DADDR
	GUESS_SK_V6_RCV_SADDR
	GUESS_SK_V6_DADDR
	GUESS_SK_DPORT
	GUESS_TCP_SK_SRTT_US
	GUESS_TCP_SK_MDEV_US
	GUESS_FLOWI4_SADDR
	GUESS_FLOWI4_DADDR
	GUESS_FLOWI4_SPORT
	GUESS_FLOWI4_DPORT
	GUESS_FLOWI6_SADDR
	GUESS_FLOWI6_DADDR
	GUESS_FLOWI6_SPORT
	GUESS_FLOWI6_DPORT
	GUESS_SKADDR_SIN_PORT
	GUESS_SKADRR6_SIN6_PORT
	GUESS_SK_NET
	GUESS_NS_COMMON_INUM
)

//nolint:stylecheck
const (
	ERR_G_NOERROR = 0
	ERR_G_SK_NET  = 19
)

var l = logger.DefaultSLogger("net_ebpf")

func SetLogger(nl *logger.Logger) {
	l = nl
}

func NewGuessManger() (*manager.Manager, error) {
	m := &manager.Manager{
		Probes: []*manager.Probe{
			{
				Section: "kprobe/tcp_v6_connect",
			}, {
				Section: "kretprobe/tcp_v6_connect",
			}, {
				Section: "kprobe/tcp_getsockopt",
			}, {
				Section: "kprobe/ip_make_skb",
			},
		},
	}
	mOpts := manager.Options{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
	}
	if buf, err := dkebpf.Asset("offset_guess.o"); err != nil {
		return nil, err
	} else if err := m.InitWithOptions((bytes.NewReader(buf)), mOpts); err != nil {
		return nil, err
	}
	return m, nil
}

func readMapGuessStatus(m *ebpf.Map) (*OffsetGuessC, error) {
	status := OffsetGuessC{}
	zero := uint64(0)
	if err := m.Lookup(&zero, unsafe.Pointer(&status)); err != nil {
		return nil, err
	} else {
		return &status, err
	}
}

func updateMapGuessStatus(m *ebpf.Map, status *OffsetGuessC) error {
	zero := uint64(0)
	status.daddr = [4]C.__u32{}
	status.saddr = [4]C.__u32{}
	status.dport = 0
	status.sport = 0
	status.rtt = 0
	status.rtt_var = 0
	status.netns = 0
	status.err = 0
	status.state = 0

	return m.Update(&zero, unsafe.Pointer(status), ebpf.UpdateAny)
}

func BpfMapGuessInit(m *manager.Manager) (*ebpf.Map, error) {
	bpfmapOffsetGuess, found, err := m.GetMap("bpfmap_offset_guess")
	if err != nil || !found {
		return nil, err
	}
	zero := uint64(0)
	status := newGuessStatus()
	err = bpfmapOffsetGuess.Update(zero, unsafe.Pointer(&status), ebpf.UpdateAny)
	if err != nil {
		return nil, err
	}
	time.Sleep(time.Millisecond * 5)
	return bpfmapOffsetGuess, nil
}

func newGuessStatus() OffsetGuessC {
	procName := filepath.Base(os.Args[0])
	if len(procName) > PROCNAMLEN-1 {
		procName = procName[:PROCNAMLEN-1]
	}

	procNameC := [PROCNAMLEN]C.__u8{}
	for i := 0; i < PROCNAMLEN-1 && i < len(procName); i++ {
		procNameC[i] = C.__u8(procName[i])
	}

	status := OffsetGuessC{
		process_name: procNameC,
		pid_tgid:     C.__u64(uint64(unix.Getpid())<<32 | uint64(unix.Gettid())),
	}

	return status
}

func copyOffset(src *OffsetGuessC, dst *OffsetGuessC) {
	dst.offset_sk_num = src.offset_sk_num
	dst.offset_inet_sport = src.offset_inet_sport
	dst.offset_sk_family = src.offset_sk_family
	dst.offset_sk_rcv_saddr = src.offset_sk_rcv_saddr
	dst.offset_sk_daddr = src.offset_sk_daddr
	dst.offset_sk_v6_rcv_saddr = src.offset_sk_v6_rcv_saddr
	dst.offset_sk_v6_daddr = src.offset_sk_v6_daddr
	dst.offset_sk_dport = src.offset_sk_dport
	dst.offset_tcp_sk_srtt_us = src.offset_tcp_sk_srtt_us
	dst.offset_tcp_sk_mdev_us = src.offset_tcp_sk_mdev_us

	dst.offset_flowi4_saddr = src.offset_flowi4_saddr
	dst.offset_flowi4_daddr = src.offset_flowi4_daddr
	dst.offset_flowi4_sport = src.offset_flowi4_sport
	dst.offset_flowi4_dport = src.offset_flowi4_dport

	dst.offset_flowi6_saddr = src.offset_flowi6_saddr
	dst.offset_flowi6_daddr = src.offset_flowi6_daddr
	dst.offset_flowi6_sport = src.offset_flowi6_sport
	dst.offset_flowi6_dport = src.offset_flowi6_dport

	dst.offset_skaddr_sin_port = src.offset_skaddr_sin_port
	dst.offset_skaddr6_sin6_port = src.offset_skaddr6_sin6_port

	dst.offset_sk_net = src.offset_sk_net
	dst.offset_ns_common_inum = src.offset_ns_common_inum
}

//nolint:gocyclo
func tryGuess(status *OffsetGuessC, check *OffsetCheck, conn *Conninfo, guessWhich int) bool {
	switch guessWhich {
	case GUESS_INET_SPORT:
		if conn.Sport != uint16(status.sport) {
			status.offset_inet_sport++
			check.inetSportOk = 0
			return false
		} else {
			check.inetSportOk++
		}
	case GUESS_SK_FAMILY:
		if uint32(status.meta)&ConnL3Mask != conn.Meta&ConnL3Mask {
			status.offset_sk_family++
			check.skFamilyOk = 0
			return false
		} else {
			check.skFamilyOk++
		}
	case GUESS_SK_DADDR:
		if conn.Daddr != *(*[4]uint32)(unsafe.Pointer(&status.daddr)) {
			status.offset_sk_daddr++
			check.skDaddrOk = 0
			return false
		} else {
			status.offset_sk_rcv_saddr = status.offset_sk_daddr + 4 // +32bit
			check.skDaddrOk++
		}
	case GUESS_SK_DPORT:
		if conn.Dport != uint16(status.dport) {
			status.offset_sk_dport++
			check.skDportOk = 0
			return false
		} else {
			status.offset_sk_num = status.offset_sk_dport + 2 // +sizeof(__be16)
			check.skDportOk++
		}
	case GUESS_SK_V6_DADDR:
		if conn.Daddr != *(*[4]uint32)(unsafe.Pointer(&status.daddr)) {
			status.offset_sk_v6_daddr++
			check.skV6DaddrOk = 0
			return false
		} else {
			status.offset_sk_v6_rcv_saddr = status.offset_sk_v6_daddr + 16 // +128bit
			check.skV6DaddrOk++
		}
	case GUESS_TCP_SK_SRTT_US:
		if conn.Rtt != uint32(status.rtt) {
			status.offset_tcp_sk_srtt_us++
			check.tcpSkSrttUsOk = 0
			return false
		} else {
			check.tcpSkSrttUsOk++
		}
	case GUESS_TCP_SK_MDEV_US:
		if conn.RttVar != uint32(status.rtt_var) {
			status.offset_tcp_sk_mdev_us++
			check.tcpSkMdevUsOk = 0
			return false
		} else {
			check.tcpSkMdevUsOk++
		}
	case GUESS_FLOWI4_SADDR:
		if conn.Saddr != *(*[4]uint32)(unsafe.Pointer(&status.saddr)) {
			status.offset_flowi4_saddr++
			check.flowi4SaddrOk = 0
			return false
		} else {
			check.flowi4SaddrOk++
		}
	case GUESS_FLOWI4_DADDR:
		if conn.Daddr != *(*[4]uint32)(unsafe.Pointer(&status.daddr)) {
			status.offset_flowi4_daddr++
			check.flowi4DaddrOk = 0
			return false
		} else {
			check.flowi4DaddrOk++
		}
	case GUESS_FLOWI4_SPORT:
	case GUESS_FLOWI4_DPORT:
		if conn.Dport != uint16(status.dport) {
			status.offset_flowi4_dport++
			check.flowi4DportOk = 0
			return false
		} else {
			status.offset_flowi4_sport = status.offset_flowi4_dport + 2 // +sizeof(__be16)
			check.flowi4DportOk++
		}
	case GUESS_FLOWI6_SADDR:
	case GUESS_FLOWI6_DADDR:
	case GUESS_FLOWI6_SPORT:
	case GUESS_FLOWI6_DPORT:
	case GUESS_SKADDR_SIN_PORT:
	case GUESS_NS_COMMON_INUM:
		if status.err == ERR_G_SK_NET {
			status.offset_sk_net++
			status.offset_ns_common_inum = 0
			check.sknetOk = 0
			return false
		} else {
			check.sknetOk++

			if conn.NetNS != uint32(status.netns) {
				status.offset_ns_common_inum++
				check.netnsInumOk = 0
				return false
			} else {
				check.netnsInumOk++
			}
		}
	}
	return true
}

// github.com/weaveworks/tcptracer-bpf
func generateRandomIPv6Address() ([4]uint32, net.IP) {
	// multicast (ff00::/8) or link-local (fe80::/10) addresses don't work for
	// our purposes so let's choose a "random number" for the first 32 bits.
	addr := [4]uint32{}
	addr[0] = 0x87586031
	addr[1] = rand.Uint32()
	addr[2] = rand.Uint32()
	addr[3] = rand.Uint32()

	ip := net.IP{}
	for x := 0; x < 4; x++ {
		for y := 0; y < 4; y++ {
			ip = append(ip,
				byte((addr[x]&(0xff<<(8*y)))>>(8*y)),
			)
		}
	}
	return addr, ip
}

func getLinuxKernelVesion() (uint64, error) {
	var err error

	kVersionStr, err := host.KernelVersion()
	if err != nil {
		return 0, err
	}

	kVersionStrArr := strings.Split(strings.Split(kVersionStr, "-")[0], ".")
	var kVersion uint64 = 0 // major(off +0), minor(off +16), patch(off +32), 0(off +48)
	if len(kVersionStrArr) == 3 {
		for i, vStr := range kVersionStrArr {
			if v, err := strconv.Atoi(vStr); err != nil {
				err = fmt.Errorf("linux kernel version parsing failed: %s", kVersionStr)
				return 0, err
			} else {
				kVersion |= uint64(v) << (16 * (3 - i))
			}
		}
	}

	return kVersion, nil
}
