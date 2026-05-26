//go:build linux
// +build linux

package l4log

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vishvananda/netlink"
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

const (
	containerPortListenInitialDelayBase   = 2 * time.Second
	containerPortListenInitialDelayJitter = 8 * time.Second
)

type portListen struct {
	portListen map[string]*nsPortListenIndex

	sync.RWMutex
}

type nsPortListenIndex struct {
	fingerprint    string
	entries        []*tcpPortInf
	anyPort        map[int]struct{}
	anyPortV6Only  map[int]struct{}
	exactPortToIPs map[int]map[string]struct{}
}

func buildPortListenIndex(entries []*tcpPortInf) *nsPortListenIndex {
	fp := fingerprintTCPPortEntries(entries)
	idx := &nsPortListenIndex{
		fingerprint:    fp,
		entries:        entries,
		anyPort:        map[int]struct{}{},
		anyPortV6Only:  map[int]struct{}{},
		exactPortToIPs: map[int]map[string]struct{}{},
	}

	for _, entry := range entries {
		if entry == nil || entry.Port <= 0 {
			continue
		}
		switch entry.IP {
		case "0.0.0.0":
			idx.anyPort[entry.Port] = struct{}{}
		case "::":
			if entry.V6 {
				idx.anyPortV6Only[entry.Port] = struct{}{}
			} else {
				idx.anyPort[entry.Port] = struct{}{}
			}
		default:
			portIPs := idx.exactPortToIPs[entry.Port]
			if portIPs == nil {
				portIPs = map[string]struct{}{}
				idx.exactPortToIPs[entry.Port] = portIPs
			}
			portIPs[entry.IP] = struct{}{}
		}
	}

	return idx
}

func fingerprintTCPPortEntries(entries []*tcpPortInf) string {
	if len(entries) == 0 {
		return ""
	}

	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s|%d|%t|%s", entry.IP, entry.Port, entry.V6, entry.St))
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func (pr *portListen) Update(nsStr string, p []*tcpPortInf) bool {
	pr.Lock()
	defer pr.Unlock()
	if pr.portListen == nil {
		pr.portListen = make(map[string]*nsPortListenIndex)
	}
	fp := fingerprintTCPPortEntries(p)
	if existing, ok := pr.portListen[nsStr]; ok && existing != nil && existing.fingerprint == fp {
		return false
	}
	pr.portListen[nsStr] = buildPortListenIndex(p)
	return true
}

func (pr *portListen) Query(ns string, k *PMeta, v6 bool, macEQ bool) conndirection {
	pr.RLock()
	defer pr.RUnlock()

	if !macEQ {
		return directionUnknown
	}
	idx, ok := pr.portListen[ns]
	if !ok || idx == nil {
		return directionUnknown
	}

	port := int(k.SrcPort)
	if _, ok := idx.anyPort[port]; ok {
		return directionIncoming
	}
	if v6 {
		if _, ok := idx.anyPortV6Only[port]; ok {
			return directionIncoming
		}
	}
	if portIPs, ok := idx.exactPortToIPs[port]; ok {
		if _, ok := portIPs[k.SrcIP]; ok {
			return directionIncoming
		}
	}
	return directionUnknown
}

type nicInfo struct {
	err              error
	inf              []*NICInfo
	hostNet, allowLo bool
	includeVirtual   bool
}

type linkAttrsSnapshot struct {
	parentIndex int
	netnsID     int
}

func newNicInf(hostNet, allowLo, includeVirtual bool) *nicInfo {
	return &nicInfo{
		hostNet:        hostNet,
		allowLo:        allowLo,
		includeVirtual: includeVirtual,
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

	linkAttrs, errLinkAttrs := currentLinkAttrs()
	if errLinkAttrs != nil {
		log.Debugf("get link attrs info failed: %s", errLinkAttrs)
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

		ifIndex := v.Index
		ifLink := 0
		netnsID := -1
		if attr, ok := linkAttrs[v.Name]; ok {
			ifLink = attr.parentIndex
			netnsID = attr.netnsID
		}
		if ifLink == 0 {
			if fallbackIfLink, err := readIfLink(v.Name); err != nil {
				log.Debugf("read iflink for %s failed: %s", v.Name, err)
			} else {
				ifLink = fallbackIfLink
			}
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
			vIface = true
		} else if vIface && !inf.includeVirtual {
			// host virtual nic only keep lo nic or not for host
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
			Index:      v.Index,
			IfIndex:    ifIndex,
			IfLink:     ifLink,
			NetNsID:    netnsID,
			MAC:        mac,
			Name:       v.Name,
			Addrs:      ipnetLi,
			VIface:     vIface,
			HostIface:  inf.hostNet,
			IsLoopback: lo,
			IsVirtual:  vIface,
		})
	}
}

func currentLinkAttrs() (map[string]linkAttrsSnapshot, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	result := make(map[string]linkAttrsSnapshot, len(links))
	for _, link := range links {
		if link == nil || link.Attrs() == nil || link.Attrs().Name == "" {
			continue
		}
		result[link.Attrs().Name] = linkAttrsSnapshot{
			parentIndex: link.Attrs().ParentIndex,
			netnsID:     link.Attrs().NetNsID,
		}
	}

	return result, nil
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
	Index   int
	IfIndex int
	IfLink  int
	NetNsID int
	Name    string
	MAC     string
	Addrs   []*net.IPNet

	// MulticastAddrs []net.Addr

	HostIface  bool
	VIface     bool
	IsLoopback bool
	IsVirtual  bool
}

func (nns *netnsHandle) nicInfo() ([]*NICInfo, error) {
	return nns.nicInfoWithVirtual(false)
}

func (nns *netnsHandle) nicInfoWithVirtual(includeVirtual bool) ([]*NICInfo, error) {
	var errCall error
	inf := newNicInf(nns.hostNet, nns.includeLo, includeVirtual)

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

func (nns *netnsHandle) portListenWatching() bool {
	return atomic.LoadInt64(&nns.portListenRunner) != 0
}

func portListenInitialDelay(ns string, hostNet bool) time.Duration {
	if hostNet || ns == "" || ns == NSUNKNOWN {
		return 0
	}

	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(ns))
	jitterWindow := uint32(containerPortListenInitialDelayJitter / time.Second)
	if jitterWindow == 0 {
		return containerPortListenInitialDelayBase
	}

	jitter := time.Duration(hasher.Sum32()%jitterWindow) * time.Second
	return containerPortListenInitialDelayBase + jitter
}

func (nns *netnsHandle) tcpPortListenWatcher(ctx context.Context, port *portListen, pidProvider func() map[int]struct{}) {
	if !atomic.CompareAndSwapInt64(&nns.portListenRunner, 0, 1) {
		return
	}
	defer atomic.StoreInt64(&nns.portListenRunner, 0)

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	if delay := portListenInitialDelay(nns.nsStr, nns.hostNet); delay > 0 {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
	}

	for {
		if atomic.LoadInt64(&nns.portListenRunner) == 0 {
			return
		}
		pids := map[int]struct{}{}
		if pidProvider != nil {
			pids = pidProvider()
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
	if nns.netns.IsOpen() {
		_ = nns.netns.Close()
	}
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

func readIfLink(ifName string) (int, error) {
	switch {
	case ifName == "", ifName == ".", ifName == "..":
		return 0, fmt.Errorf("invalid interface name %q", ifName)
	case strings.ContainsRune(ifName, filepath.Separator):
		return 0, fmt.Errorf("invalid interface name %q", ifName)
	}

	//nolint:gosec // ifName is validated above and cannot escape /sys/class/net.
	cnt, err := os.ReadFile(filepath.Join("/sys/class/net", ifName, "iflink"))
	if err != nil {
		return 0, err
	}

	val := strings.TrimSpace(string(cnt))
	if val == "" {
		return 0, fmt.Errorf("empty iflink")
	}

	ifLink, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}

	return ifLink, nil
}

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
