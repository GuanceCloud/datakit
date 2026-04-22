//go:build linux
// +build linux

package l4log

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/cilium/ebpf"
	"github.com/google/gopacket/afpacket"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
	dkebpf "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/c"
	"golang.org/x/sys/unix"
)

const (
	hostPeerFilterProgram = "socket__host_peer_filter"
	hostPeerFilterMapName = "bpfmap_host_peer_ifindex"
)

type sharedHostPeerSocketFilter struct {
	runtime    *bpfutil.Runtime
	ifindexMap *ebpf.Map
	ifindexes  []int
	mode       string
}

func sharedHostPeerFilterFingerprint(ifindexes []int) string {
	norm := normalizeIfIndexes(ifindexes)
	if len(norm) == 0 {
		return ""
	}

	parts := make([]string, 0, len(norm))
	for _, ifindex := range norm {
		parts = append(parts, fmt.Sprintf("%d", ifindex))
	}
	return strings.Join(parts, ",")
}

func (f *sharedHostPeerSocketFilter) sync(h *afpacket.TPacket, ifindexes []int) (string, bool, error) {
	norm := normalizeIfIndexes(ifindexes)
	if len(norm) == 0 {
		return "", false, fmt.Errorf("empty ifindex whitelist")
	}

	ebpfFailed := false
	if err := f.syncEBPF(h, norm); err == nil {
		f.mode = "ebpf"
		return f.mode, false, nil
	} else {
		ebpfFailed = true
		log.Warnf("sync shared host-peer ebpf filter failed: %s", err)
		f.close()
	}

	if err := attachSharedHostPeerCBPFFilter(h, norm); err != nil {
		return "", ebpfFailed, err
	}
	f.ifindexes = append(f.ifindexes[:0], norm...)
	f.mode = "cbpf"
	return f.mode, ebpfFailed, nil
}

func (f *sharedHostPeerSocketFilter) syncEBPF(h *afpacket.TPacket, ifindexes []int) error {
	if err := f.ensureEBPF(h); err != nil {
		return err
	}
	if err := f.syncMap(ifindexes); err != nil {
		return err
	}
	return nil
}

func (f *sharedHostPeerSocketFilter) ensureEBPF(h *afpacket.TPacket) error {
	if f.runtime != nil && f.ifindexMap != nil {
		return nil
	}

	fd, err := tpacketFD(h)
	if err != nil {
		return err
	}

	buf, err := dkebpf.HostPeerFilterBin()
	if err != nil {
		return fmt.Errorf("load host peer filter bin: %w", err)
	}

	rt := &bpfutil.Runtime{
		Probes: []*bpfutil.HookSpec{
			{
				ID: bpfutil.HookID{
					EBPFFuncName: hostPeerFilterProgram,
				},
				SocketFD: fd,
			},
		},
	}

	loadSpec := bpfutil.LoadSpec{
		RLimit: &unix.Rlimit{
			Cur: ^uint64(0),
			Max: ^uint64(0),
		},
	}

	if err := rt.LoadFromReader(bytes.NewReader(buf), loadSpec); err != nil {
		return fmt.Errorf("load host peer filter runtime: %w", err)
	}

	if err := rt.StartRuntime(); err != nil {
		if shutdownErr := rt.Shutdown(); shutdownErr != nil {
			log.Debugf("shutdown host peer filter runtime after start failure: %s", shutdownErr)
		}
		return fmt.Errorf("start host peer filter runtime: %w", err)
	}

	mp, err := rt.LookupMap(hostPeerFilterMapName)
	if err != nil {
		if shutdownErr := rt.Shutdown(); shutdownErr != nil {
			log.Debugf("shutdown host peer filter runtime after map lookup failure: %s", shutdownErr)
		}
		return fmt.Errorf("lookup host peer filter map: %w", err)
	}

	f.runtime = rt
	f.ifindexMap = mp
	return nil
}

func (f *sharedHostPeerSocketFilter) syncMap(ifindexes []int) error {
	if f.ifindexMap == nil {
		return fmt.Errorf("host peer ifindex map unavailable")
	}

	prev := make(map[int]struct{}, len(f.ifindexes))
	for _, ifindex := range f.ifindexes {
		prev[ifindex] = struct{}{}
	}

	next := make(map[int]struct{}, len(ifindexes))
	val := uint8(1)
	for _, ifindex := range ifindexes {
		next[ifindex] = struct{}{}
		if _, ok := prev[ifindex]; ok {
			continue
		}
		key := uint32(ifindex)
		if err := f.ifindexMap.Update(&key, &val, ebpf.UpdateAny); err != nil {
			return fmt.Errorf("add host peer ifindex %d: %w", ifindex, err)
		}
	}

	for _, ifindex := range f.ifindexes {
		if _, ok := next[ifindex]; ok {
			continue
		}
		key := uint32(ifindex)
		if err := f.ifindexMap.Delete(&key); err != nil && !errors.Is(err, ebpf.ErrKeyNotExist) {
			return fmt.Errorf("delete host peer ifindex %d: %w", ifindex, err)
		}
	}

	f.ifindexes = append(f.ifindexes[:0], ifindexes...)
	return nil
}

func (f *sharedHostPeerSocketFilter) close() {
	if f == nil {
		return
	}
	if f.runtime != nil {
		if err := f.runtime.Shutdown(); err != nil {
			log.Debugf("shutdown host peer filter runtime: %s", err)
		}
	}
	f.runtime = nil
	f.ifindexMap = nil
	f.ifindexes = nil
	f.mode = ""
}

func attachSharedHostPeerCBPFFilter(h *afpacket.TPacket, ifindexes []int) error {
	raw, err := newSharedHostPeerCBPFFilter(ifindexes)
	if err != nil {
		return err
	}
	return h.SetBPF(raw)
}

func tpacketFD(h *afpacket.TPacket) (int, error) {
	if h == nil {
		return 0, fmt.Errorf("nil tpacket")
	}

	v := reflect.ValueOf(h)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return 0, fmt.Errorf("invalid tpacket pointer")
	}

	f := v.Elem().FieldByName("fd")
	if !f.IsValid() || f.Kind() != reflect.Int {
		return 0, fmt.Errorf("tpacket fd field unavailable")
	}

	fd := int(f.Int())
	if fd <= 0 {
		return 0, fmt.Errorf("invalid tpacket fd %d", fd)
	}
	return fd, nil
}
