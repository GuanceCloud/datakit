package offset_guess

// #include "../c/offset_guess/offset.h"
import "C"

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"time"
	"unsafe"

	"github.com/DataDog/ebpf"
	"github.com/DataDog/ebpf/manager"
	"golang.org/x/sys/unix"

	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/externals/net_ebpf/c"
)

type OffsetGuessC C.struct_offset_guess

func NewGuessManger() (*manager.Manager, error) {
	m := &manager.Manager{
		Probes: []*manager.Probe{
			{
				Section:         "kprobe/tcp_rcv_established",
				KProbeMaxActive: 128,
			},
		},
	}
	m_opts := manager.Options{
		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
	}
	if buf, err := dkebpf.Asset("offset_guess.o"); err != nil {
		return nil, err
	} else {
		if err := m.InitWithOptions((bytes.NewReader(buf)), m_opts); err != nil {
			return nil, err
		}
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
	return m.Update(&zero, unsafe.Pointer(status), ebpf.UpdateAny)
}

func BpfMapGuessInit(m *manager.Manager) (*ebpf.Map, error) {
	bpfmap_offset_guess, found, err := m.GetMap("bpfmap_offset_guess")
	if err != nil || !found {
		return nil, err
	}
	zero := uint64(0)
	status := newGuessStatus()
	err = bpfmap_offset_guess.Update(zero, unsafe.Pointer(&status), ebpf.UpdateAny)
	if err != nil {
		return nil, err
	}
	time.Sleep(time.Millisecond * 5)
	return bpfmap_offset_guess, nil
}

func newGuessStatus() OffsetGuessC {
	proc_name := filepath.Base(os.Args[0])
	if len(proc_name) > PROCNAMLEN-1 {
		proc_name = proc_name[:PROCNAMLEN-1]
	}

	proc_name_c := [PROCNAMLEN]C.__u8{}
	for i := 0; i < PROCNAMLEN-1 && i < len(proc_name); i++ {
		proc_name_c[i] = C.__u8(proc_name[i])
	}
	status := OffsetGuessC{
		process_name: proc_name_c,
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
	dst.offest_flowi4_daddr = src.offest_flowi4_daddr
	dst.offset_flowi4_sport = src.offset_flowi4_sport
	dst.offset_flowi4_dport = src.offset_flowi4_dport

	dst.offset_flowi6_saddr = src.offset_flowi6_saddr
	dst.offset_flowi6_daddr = src.offset_flowi6_daddr
	dst.offset_flowi6_sport = src.offset_flowi6_sport
	dst.offset_flowi6_dport = src.offset_flowi6_dport

	dst.offset_skaddr_sin_port = src.offset_skaddr_sin_port
	dst.offset_skaddr6_sin6_port = src.offset_skaddr6_sin6_port
}

func setConnType(status *OffsetGuessC, meta uint32) {
	status.conn_type = C.__u32(meta)
}
