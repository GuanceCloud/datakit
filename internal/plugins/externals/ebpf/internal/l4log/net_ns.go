//go:build linux
// +build linux

package l4log

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vishvananda/netns"
	"golang.org/x/exp/slices"
	"golang.org/x/sys/unix"
)

type TCPSt string

const (
	TCPUnkown TCPSt  = ""
	TCPListen TCPSt  = "0A"
	NSUNKNOWN string = "unknown"
)

type portListen struct {
	portListen map[string][]*tcpPortInf

	sync.RWMutex
}

func (pr *portListen) Update(nsStr string, p []*tcpPortInf) {
	pr.Lock()
	defer pr.Unlock()
	if pr.portListen == nil {
		pr.portListen = make(map[string][]*tcpPortInf)
	}
	pr.portListen[nsStr] = p
}

func (pr *portListen) Query(ns string, k *PMeta, v6 bool, macEQ bool) conndirection {
	pr.RLock()
	defer pr.RUnlock()

	if !macEQ {
		return directionUnknown
	}
	if v, ok := pr.portListen[ns]; ok {
		for _, v := range v {
			switch v.IP {
			case "0.0.0.0", "::":
				if v6 && !v.V6 { // "::"(v4 and v6)
					continue
				}
				if k.SrcPort == uint16(v.Port) {
					return directionIncoming
				}
			default:
				if k.SrcIP == v.IP && k.SrcPort == uint16(v.Port) {
					return directionIncoming
				}
			}
		}
	}
	return directionUnknown
}

type nicInfo struct {
	err              error
	inf              []*NICInfo
	hostNet, allowLo bool
}

func newNicInf(hostNet, allowLo bool) *nicInfo {
	return &nicInfo{
		hostNet: hostNet,
		allowLo: allowLo,
	}
}

func (inf *nicInfo) _nicInfo() {
	var vifaces map[string]struct{}
	var errVi error
	if inf.hostNet {
		vifaces, errVi = virtualInterfaces()
		if errVi != nil {
			log.Errorf("get virtual interface info failed %s", errVi.Error())
		}
	}

	var netifaces []net.Interface
	netifaces, inf.err = net.Interfaces()
	if inf.err != nil {
		inf.err = fmt.Errorf("get net interfaces: %w", inf.err)
		return
	}

	for _, v := range netifaces {
		if v.Flags&0b1 != net.FlagUp {
			continue
		}
		mac := v.HardwareAddr.String()
		var lo bool
		if v.Flags&0b100 == net.FlagLoopback {
			lo = true
		}

		var vIface bool
		if _, ok := vifaces[v.Name]; ok {
			vIface = true
		}

		if !inf.hostNet {
			// only filter lo nic for containers
			if lo && !inf.allowLo {
				continue
			}
		} else if vIface {
			// (virtual nic) keep lo nic or not for host
			if !lo || !inf.allowLo {
				continue
			}
		}

		if errVi != nil {
			if strings.HasPrefix(v.Name, "veth") || strings.HasPrefix(v.Name, "cali") {
				// 容器的 veth 网卡不记录
				continue
			}
			if lo && !inf.allowLo {
				continue
			}
		}

		var ipnetLi []*net.IPNet
		addrs, _ := v.Addrs()
		for _, v := range addrs {
			if v, ok := v.(*net.IPNet); ok {
				ipnetLi = append(ipnetLi, v)
			}
		}
		// multiCAddr, _ := v.MulticastAddrs()
		inf.inf = append(inf.inf, &NICInfo{
			Index:  v.Index,
			MAC:    mac,
			Name:   v.Name,
			Addrs:  ipnetLi,
			VIface: vIface,
		})
	}
}

type nsInfo struct {
	nsstr   string
	err     [2]error
	portInf []*tcpPortInf

	listen map[struct {
		IP   string
		Port int
	}]struct{}
}

func newNsInf(nsstr string) *nsInfo {
	return &nsInfo{
		nsstr: nsstr,
		listen: make(map[struct {
			IP   string
			Port int
		}]struct{}),
	}
}

func (inf *nsInfo) _portListen(pid int) {
	var v4, v6 string
	if pid > 0 {
		v4 = fmt.Sprintf("/proc/%d/net/tcp", pid)
		v6 = fmt.Sprintf("/proc/%d/net/tcp6", pid)
	} else {
		v4 = "/proc/net/tcp"
		v6 = "/proc/net/tcp6"
	}

	inf.portInf = make([]*tcpPortInf, 0)
	if v, err := parseTCPStFromFile(v4, false, true); err != nil {
		inf.err[0] = err
	} else {
		for _, v := range v {
			k := struct {
				IP   string
				Port int
			}{v.IP, v.Port}
			if _, ok := inf.listen[k]; !ok {
				inf.listen[k] = struct{}{}
				inf.portInf = append(inf.portInf, v)
			}
		}
	}

	if v, err := parseTCPStFromFile(v6, true, true); err != nil {
		inf.err[1] = err
	} else {
		for _, v := range v {
			k := struct {
				IP   string
				Port int
			}{v.IP, v.Port}
			if _, ok := inf.listen[k]; !ok {
				inf.listen[k] = struct{}{}
				inf.portInf = append(inf.portInf, v)
			}
		}
	}
}

type netnsHandle struct {
	hostNet   bool
	includeLo bool

	nsStr string
	netns netns.NsHandle

	portListenRunner int64
}

func newNetNsHandle(hostnet, includeLo bool, ns netns.NsHandle) *netnsHandle {
	return &netnsHandle{
		hostNet:   hostnet,
		includeLo: includeLo,
		nsStr:     NSInode(ns),
		netns:     ns,
	}
}

type tcpPortInf struct {
	V6   bool
	IP   string
	Port int
	St   string
}

type NICInfo struct {
	Index int
	Name  string
	MAC   string
	Addrs []*net.IPNet

	// MulticastAddrs []net.Addr

	VIface bool
}

func (nns *netnsHandle) nicInfo() ([]*NICInfo, error) {
	var errCall error
	inf := newNicInf(nns.hostNet, nns.includeLo)

	errCall = CallWithNetNS(nns.netns, inf._nicInfo)

	switch {
	case errCall != nil:
		return nil, errCall
	case inf.err != nil:
		return nil, inf.err
	default:
		return inf.inf, nil
	}
}

func (nns *netnsHandle) tcpPortListen(pids map[int]struct{}) ([]*tcpPortInf, error) {
	inf := newNsInf(nns.nsStr)

	if nns.hostNet {
		inf._portListen(0)
	} else {
		for k := range pids {
			inf._portListen(k)
		}
	}
	if len(inf.portInf) == 0 {
		switch {
		case inf.err[0] != nil:
			return nil, inf.err[0]
		case inf.err[1] != nil:
			return nil, inf.err[1]
		}
	}

	return inf.portInf, nil
}

func (nns *netnsHandle) tcpPortListenWatcher(ctx context.Context, port *portListen, pids map[int]struct{}) {
	if !atomic.CompareAndSwapInt64(&nns.portListenRunner, 0, 1) {
		return
	}

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		if atomic.LoadInt64(&nns.portListenRunner) == 0 {
			return
		}
		if p, err := nns.tcpPortListen(pids); err != nil {
			log.Errorf("get port info failed: %s", err.Error())
		} else {
			port.Update(nns.nsStr, p)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (nns *netnsHandle) close() {
	atomic.StoreInt64(&nns.portListenRunner, 0)
	_ = nns.netns.Close()
}

func parseTCPStFromFile(fp string, v6 bool, listenOnly bool) ([]*tcpPortInf, error) {
	f, err := os.Open(fp) //nolint:gosec
	if err != nil {
		return nil, err
	}

	var portInf []*tcpPortInf

	scanner := bufio.NewScanner(f)

	ln := 0
	for scanner.Scan() {
		ln++
		if ln == 1 {
			continue
		}
		cnt := scanner.Text()
		if v, ok := parseTCPSt(cnt, v6, listenOnly); ok {
			portInf = append(portInf, v)
		} else {
			break
		}
	}

	if err := f.Close(); err != nil {
		log.Errorf("close file: %w", err)
	}

	return portInf, nil
}

func parseTCPSt(s string, v6 bool, listenOnly bool) (*tcpPortInf, bool) {
	v := strings.Split(s, " ")
	val := [4]string{}
	count := 0
	for _, v := range v {
		if v != "" && count < 4 {
			val[count] = v
			count++
		}
	}
	if count != 4 {
		return nil, false
	}

	tp := &tcpPortInf{}

	if listenOnly && val[3] != "0A" {
		return nil, false
	} else {
		tp.St = val[3]
	}

	if v := strings.Split(val[1], ":"); len(v) == 2 {
		if v, err := hex.DecodeString(v[0]); err != nil {
			return nil, false
		} else {
			if len(v) == 16 {
				for i := 0; i < 4; i++ {
					slices.Reverse(v[i*4 : i*4+4])
				}
			} else {
				slices.Reverse(v)
			}
			tp.IP = net.IP(v).String()
			tp.V6 = v6
		}
		if v, err := hex.DecodeString(v[1]); err != nil {
			return nil, false
		} else {
			tp.Port = int(binary.BigEndian.Uint16(v))
		}
	} else {
		return nil, false
	}

	return tp, true
}

func NSInode(ns netns.NsHandle) string {
	if ns == -1 {
		return NSUNKNOWN
	}

	var s unix.Stat_t
	if err := unix.Fstat(int(ns), &s); err != nil {
		return NSUNKNOWN
	}

	return fmt.Sprintf("%d", s.Ino)
}

const vnicDevPath = "/sys/devices/virtual/net/"

func virtualInterfaces() (map[string]struct{}, error) {
	v, err := os.ReadDir(vnicDevPath)
	if err != nil {
		return nil, fmt.Errorf("read dir %s` failed: %w",
			vnicDevPath, err)
	}

	cardVirtual := make(map[string]struct{})
	for _, v := range v {
		if v.IsDir() {
			cardVirtual[v.Name()] = struct{}{}
		}
	}
	return cardVirtual, nil
}
