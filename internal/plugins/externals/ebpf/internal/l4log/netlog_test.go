//go:build linux
// +build linux

package l4log

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/gopacket/afpacket"

	"github.com/stretchr/testify/assert"
	"github.com/vishvananda/netns"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/pkg/cli"
)

func TestPortListen(t *testing.T) {
	nns, err := netns.Get()
	if err != nil {
		t.Error(err)
	}

	h := newNetNsHandle(true, true, nns)

	if v, err := h.tcpPortListen(nil); err != nil {
		t.Error(err)
	} else {
		for _, v := range v {
			t.Log(v.IP, " ", v.Port, " ", v.St, " ", v.V6)
		}
	}

	if v, err := h.nicInfo(); err != nil {
		t.Error(err)
	} else {
		for _, v := range v {
			s := strings.Builder{}
			s.WriteString(fmt.Sprintf("%s %s %d %v\n", v.Name, v.MAC, v.Index, v.VIface))
			for _, v := range v.Addrs {
				s.WriteString(fmt.Sprintf("%s\n", v.IP))
			}
			t.Log(s.String())
		}
	}
}

func TestSpanid(t *testing.T) {
	spanID := "tpIhb+V93r4="

	r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(spanID))
	buf := make([]byte, 16)
	if _, err := r.Read(buf); err != nil {
		t.Error(err)
	}

	buf = buf[:8]

	d := binary.BigEndian.Uint64(buf)

	s := strconv.FormatUint(d, 16)
	t.Log(s)
}

func TestMatch(t *testing.T) {
	name := "veth"
	assert.True(t, strings.HasPrefix(name, "veth"))

	name = "veth1"
	assert.True(t, strings.HasPrefix(name, "veth"))
}

func TestPortListenIndexQuery(t *testing.T) {
	pr := &portListen{}
	pr.Update("ns-1", []*tcpPortInf{
		{IP: "0.0.0.0", Port: 80, V6: false},
		{IP: "::", Port: 443, V6: true},
		{IP: "10.0.0.8", Port: 8080, V6: false},
	})

	assert.Equal(t, directionIncoming, pr.Query("ns-1", &PMeta{SrcIP: "1.2.3.4", SrcPort: 80}, false, true))
	assert.Equal(t, directionIncoming, pr.Query("ns-1", &PMeta{SrcIP: "2001::1", SrcPort: 443}, true, true))
	assert.Equal(t, directionIncoming, pr.Query("ns-1", &PMeta{SrcIP: "10.0.0.8", SrcPort: 8080}, false, true))
	assert.Equal(t, directionUnknown, pr.Query("ns-1", &PMeta{SrcIP: "10.0.0.9", SrcPort: 8080}, false, true))
	assert.Equal(t, directionUnknown, pr.Query("ns-1", &PMeta{SrcIP: "1.2.3.4", SrcPort: 80}, false, false))
}

func TestPortListenUpdateDedupByFingerprint(t *testing.T) {
	pr := &portListen{}

	changed := pr.Update("ns-1", []*tcpPortInf{
		{IP: "10.0.0.8", Port: 8080, V6: false, St: string(TCPListen)},
		{IP: "0.0.0.0", Port: 80, V6: false, St: string(TCPListen)},
	})
	assert.True(t, changed)

	changed = pr.Update("ns-1", []*tcpPortInf{
		{IP: "0.0.0.0", Port: 80, V6: false, St: string(TCPListen)},
		{IP: "10.0.0.8", Port: 8080, V6: false, St: string(TCPListen)},
	})
	assert.False(t, changed)

	changed = pr.Update("ns-1", []*tcpPortInf{
		{IP: "0.0.0.0", Port: 81, V6: false, St: string(TCPListen)},
	})
	assert.True(t, changed)
}

func TestDetectProto(t *testing.T) {
	h := HTTPLog{}
	if !h.DetectProto([]byte("GET / HTTP1.1\r\n")) {
		t.Fatalf("should be http proto")
	}
}

func TestResolveCapturePlanFallsBackToNetNS(t *testing.T) {
	nsInf := &netnsInformation{nsUID: "42", netnsID: -1, netnsIDAt: time.Now()}
	nic := &NICInfo{
		Name:    "eth0",
		MAC:     "00:11:22:33:44:55",
		IfIndex: 3,
		IfLink:  3,
		Addrs:   []*net.IPNet{},
	}

	plan := resolveCapturePlan(nsInf, nic, hostNICInventory{})
	assert.Equal(t, captureModeInNetNS, plan.mode)
	assert.False(t, plan.openInHostNS)
	assert.False(t, plan.trustLocal)
	assert.Equal(t, [2]string{"eth0", "00:11:22:33:44:55"}, plan.openIface)
	assert.Equal(t, reasonHostPeerNotFound, plan.reasonCode)
}

func TestResolveCapturePlanUsesHostPeerForVirtualPeer(t *testing.T) {
	nsInf := &netnsInformation{nsUID: "42", netnsID: -1, netnsIDAt: time.Now()}
	nic := &NICInfo{
		Name:    "eth0",
		MAC:     "00:11:22:33:44:55",
		IfIndex: 7,
		IfLink:  11,
		Addrs:   []*net.IPNet{},
	}
	hostInv := hostNICInventory{
		peers: map[int]*NICInfo{
			11: {
				Name:      "vethabcd",
				MAC:       "66:77:88:99:aa:bb",
				IfIndex:   11,
				IsVirtual: true,
			},
		},
	}

	plan := resolveCapturePlan(nsInf, nic, hostInv)
	assert.Equal(t, captureModeHostPeer, plan.mode)
	assert.True(t, plan.openInHostNS)
	assert.True(t, plan.trustLocal)
	assert.Equal(t, 11, plan.routeIfIndex)
	assert.Equal(t, [2]string{"vethabcd", "66:77:88:99:aa:bb"}, plan.openIface)
	assert.Equal(t, nic, plan.logicalNIC)
}

func TestResolveCapturePlanUsesHostPeerWhenIfLinkEqualsIfIndex(t *testing.T) {
	nsInf := &netnsInformation{nsUID: "42", netnsID: -1, netnsIDAt: time.Now()}
	nic := &NICInfo{
		Name:    "eth0",
		MAC:     "00:11:22:33:44:55",
		IfIndex: 2,
		IfLink:  2,
		Addrs:   []*net.IPNet{},
	}
	hostInv := hostNICInventory{
		peers: map[int]*NICInfo{
			2: {
				Name:      "veth-peer",
				MAC:       "66:77:88:99:aa:bb",
				IfIndex:   2,
				IsVirtual: true,
			},
		},
	}

	plan := resolveCapturePlan(nsInf, nic, hostInv)
	assert.Equal(t, captureModeHostPeer, plan.mode)
	assert.True(t, plan.openInHostNS)
	assert.True(t, plan.trustLocal)
	assert.Equal(t, 2, plan.routeIfIndex)
	assert.Equal(t, [2]string{"veth-peer", "66:77:88:99:aa:bb"}, plan.openIface)
	assert.Equal(t, reasonHostPeerSelected, plan.reasonCode)
}

func TestResolveCapturePlanUsesNetNSScopedCandidate(t *testing.T) {
	nsInf := &netnsInformation{nsUID: "42", netnsID: 23, netnsIDAt: time.Now()}
	nic := &NICInfo{
		Name:    "eth0",
		MAC:     "00:11:22:33:44:55",
		IfIndex: 2,
		IfLink:  2,
		Addrs:   []*net.IPNet{},
	}
	hostInv := hostNICInventory{
		peers: map[int]*NICInfo{
			2: {
				Name:    "eth0",
				IfIndex: 2,
			},
		},
		peerCandidates: map[int][]*NICInfo{
			2: {
				{
					Name:      "veth-a",
					MAC:       "66:77:88:99:aa:01",
					IfIndex:   101,
					IfLink:    2,
					NetNsID:   17,
					IsVirtual: true,
				},
				{
					Name:      "veth-b",
					MAC:       "66:77:88:99:aa:02",
					IfIndex:   102,
					IfLink:    2,
					NetNsID:   23,
					IsVirtual: true,
				},
			},
		},
	}

	plan := resolveCapturePlan(nsInf, nic, hostInv)
	assert.Equal(t, captureModeHostPeer, plan.mode)
	assert.Equal(t, 102, plan.routeIfIndex)
	assert.Equal(t, [2]string{"veth-b", "66:77:88:99:aa:02"}, plan.openIface)
	assert.Equal(t, reasonHostPeerSelected, plan.reasonCode)
}

func TestResolveCapturePlanFallsBackWhenHostPeerAmbiguous(t *testing.T) {
	nsInf := &netnsInformation{nsUID: "42", netnsID: -1, netnsIDAt: time.Now()}
	nic := &NICInfo{
		Name:    "eth0",
		MAC:     "00:11:22:33:44:55",
		IfIndex: 2,
		IfLink:  2,
		Addrs:   []*net.IPNet{},
	}
	hostInv := hostNICInventory{
		peerCandidates: map[int][]*NICInfo{
			2: {
				{
					Name:      "veth-a",
					MAC:       "66:77:88:99:aa:01",
					IfIndex:   101,
					IfLink:    2,
					NetNsID:   17,
					IsVirtual: true,
				},
				{
					Name:      "veth-b",
					MAC:       "66:77:88:99:aa:02",
					IfIndex:   102,
					IfLink:    2,
					NetNsID:   23,
					IsVirtual: true,
				},
			},
		},
	}

	plan := resolveCapturePlan(nsInf, nic, hostInv)
	assert.Equal(t, captureModeInNetNS, plan.mode)
	assert.Equal(t, reasonHostPeerAmbiguous, plan.reasonCode)
	assert.Equal(t, [2]string{"eth0", "00:11:22:33:44:55"}, plan.openIface)
}

func TestFallbackCapturePlanUsesOriginalInterface(t *testing.T) {
	nsInf := &netnsInformation{nsUID: "42"}
	nic := &NICInfo{
		Name: "eth0",
		MAC:  "00:11:22:33:44:55",
	}

	plan := fallbackCapturePlan(nsInf, nic, reasonHostPeerConflict)
	assert.Equal(t, captureModeInNetNS, plan.mode)
	assert.False(t, plan.openInHostNS)
	assert.False(t, plan.trustLocal)
	assert.Equal(t, reasonHostPeerConflict, plan.reasonCode)
	assert.Equal(t, [2]string{"eth0", "00:11:22:33:44:55"}, plan.openIface)
	assert.Equal(t, nic, plan.logicalNIC)
}

func TestCapturePlanStatsSummary(t *testing.T) {
	stats := newCapturePlanStats()
	snapshot := newSnapshotPlanCounters()
	applyPlanCounter(&snapshot, &capturePlan{reasonCode: reasonHostPeerSelected})
	applyPlanCounter(&snapshot, &capturePlan{reasonCode: reasonHostPeerConflict})
	applyPlanCounter(&snapshot, &capturePlan{reasonCode: reasonIfLinkUnavailable})
	applyPlanCounter(&snapshot, &capturePlan{reasonCode: reasonHostNamespaceOrMissingNIC})
	stats.ReplaceSnapshot(snapshot)

	stats.IncDelta(&capturePlan{reasonCode: reasonHostPeerSelected})
	stats.IncDelta(&capturePlan{reasonCode: reasonHostPeerConflict})

	got := stats.SummaryAndResetDelta()
	assert.Contains(t, got, "snapshot_host_peer_selected=1")
	assert.Contains(t, got, "snapshot_fallback_in_netns=1")
	assert.Contains(t, got, "snapshot_fallback_conflict=1")
	assert.Contains(t, got, "snapshot_host_namespace_direct=1")
	assert.Contains(t, got, "delta_host_peer_selected=1")
	assert.Contains(t, got, "delta_fallback_conflict=1")
	assert.Contains(t, got, "snapshot_reasons=[")
	assert.Contains(t, got, "delta_reasons=[")
	assert.Contains(t, got, reasonHostPeerSelected+"=1")
	assert.Contains(t, got, reasonHostPeerConflict+"=1")

	got2 := stats.SummaryAndResetDelta()
	assert.Contains(t, got2, "delta_host_peer_selected=0")
	assert.Contains(t, got2, "delta_fallback_conflict=0")
}

func TestBuildNICIPsAndFingerprint(t *testing.T) {
	nics := []*NICInfo{
		{
			Name:    "eth0",
			MAC:     "00:11:22:33:44:55",
			IfIndex: 7,
			Addrs: []*net.IPNet{
				{IP: net.ParseIP("10.0.0.2")},
				{IP: net.ParseIP("fd00::2")},
			},
		},
		{
			Name:    "eth1",
			MAC:     "00:11:22:33:44:66",
			IfIndex: 8,
			Addrs: []*net.IPNet{
				{IP: net.ParseIP("10.0.0.3")},
			},
		},
	}

	ips, fp := buildNICIPsAndFingerprint(nics)
	assert.Equal(t, []string{"10.0.0.2", "10.0.0.3", "fd00::2"}, ips)
	assert.Contains(t, fp, "eth0")
	assert.Contains(t, fp, "eth1")
	assert.Contains(t, fp, "10.0.0.2")
	assert.Contains(t, fp, "fd00::2")
}

func TestAnyPID(t *testing.T) {
	cur := os.Getpid()
	pid, err := anyPID(map[int]struct{}{cur: {}})
	assert.NoError(t, err)
	assert.Equal(t, cur, pid)

	_, err = anyPID(map[int]struct{}{})
	assert.Error(t, err)
}

func TestAnyPIDPrunesInvalidEntries(t *testing.T) {
	pids := map[int]struct{}{
		-1: {},
		0:  {},
	}
	_, err := anyPID(pids)
	assert.Error(t, err)
	assert.Empty(t, pids)
}

func TestOpenContainerNetNSMatchesExpectedNamespace(t *testing.T) {
	cur := os.Getpid()
	ns, err := netns.GetFromPid(cur)
	if !assert.NoError(t, err) {
		return
	}
	defer ns.Close() //nolint:errcheck

	nsInf := &netnsInformation{
		nsUID: NSInode(ns),
		pid:   map[int]struct{}{cur: {}},
	}

	pid, opened, err := openContainerNetNS(nsInf)
	if !assert.NoError(t, err) {
		return
	}
	defer opened.Close() //nolint:errcheck

	assert.Equal(t, cur, pid)
	assert.Equal(t, nsInf.nsUID, NSInode(opened))
}

func TestOpenContainerNetNSRejectsNamespaceMismatch(t *testing.T) {
	cur := os.Getpid()
	nsInf := &netnsInformation{
		nsUID: "definitely-not-current-netns",
		pid:   map[int]struct{}{cur: {}},
	}

	_, opened, err := openContainerNetNS(nsInf)
	assert.Error(t, err)
	assert.Equal(t, netns.NsHandle(-1), opened)
}

func TestCallWithNetNSRejectsNilCallback(t *testing.T) {
	ns, err := netns.Get()
	if !assert.NoError(t, err) {
		return
	}
	defer ns.Close() //nolint:errcheck

	err = CallWithNetNS(ns, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing netns callback")
}

func TestCallWithNetNSPropagatesPanic(t *testing.T) {
	ns, err := netns.Get()
	if !assert.NoError(t, err) {
		return
	}
	defer ns.Close() //nolint:errcheck

	assert.PanicsWithValue(t, "boom", func() {
		_ = CallWithNetNS(ns, func() {
			panic("boom")
		})
	})
}

func TestCloneNICInfos(t *testing.T) {
	src := []*NICInfo{
		{
			Name:    "eth0",
			MAC:     "00:11:22:33:44:55",
			IfIndex: 7,
			Addrs: []*net.IPNet{
				{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(24, 32)},
			},
		},
	}

	dst := cloneNICInfos(src)
	if assert.Len(t, dst, 1) {
		assert.Equal(t, src[0].Name, dst[0].Name)
		assert.NotSame(t, src[0], dst[0])
		assert.NotSame(t, src[0].Addrs[0], dst[0].Addrs[0])
	}

	src[0].Name = "changed"
	src[0].Addrs[0].IP[0] = 42
	assert.Equal(t, "eth0", dst[0].Name)
	assert.NotEqual(t, src[0].Addrs[0].IP.String(), dst[0].Addrs[0].IP.String())
}

func TestContainerNICCacheTTLConstant(t *testing.T) {
	assert.Equal(t, 15*time.Second, containerNICCacheTTL)
}

func TestHostNICCacheTTLConstant(t *testing.T) {
	assert.Equal(t, 15*time.Second, hostNICCacheTTL)
}

func TestUnstableNICCacheTTLConstants(t *testing.T) {
	assert.Equal(t, 3*time.Second, unstableContainerNICCacheTTL)
	assert.Equal(t, 3*time.Second, unstableHostNICCacheTTL)
}

func TestContainerNetNSErrorBackoffConstant(t *testing.T) {
	assert.Equal(t, 10*time.Second, containerNetNSErrorBackoff)
}

func TestContainerNICCacheTTLForNS(t *testing.T) {
	assert.Equal(t, containerNICCacheTTL, containerNICCacheTTLForNS(nil))
	assert.Equal(t, containerNICCacheTTL, containerNICCacheTTLForNS(&netnsInformation{}))
	assert.Equal(t, unstableContainerNICCacheTTL, containerNICCacheTTLForNS(&netnsInformation{
		bootstrapPending: true,
	}))
}

func TestHostNICInventoryTTL(t *testing.T) {
	m := &netlogMonitor{
		netnsInfo:   map[string]*netnsInformation{},
		hostRuntime: &hostNamespaceRuntime{},
	}
	assert.Equal(t, hostNICCacheTTL, m.hostNICInventoryTTL())

	m.netnsInfo["pod-a"] = &netnsInformation{bootstrapPending: true}
	assert.Equal(t, unstableHostNICCacheTTL, m.hostNICInventoryTTL())

	m.netnsInfo["pod-a"].bootstrapPending = false
	m.hostRuntime.sharedCapture = &hostPeerSharedCapture{}
	assert.Equal(t, unstableHostNICCacheTTL, m.hostNICInventoryTTL())
}

func TestHostNamespaceRuntimeCachedInventoryHonorsTTL(t *testing.T) {
	rt := &hostNamespaceRuntime{
		ns:          &netnsInformation{nsUID: "host"},
		inventory:   hostNICInventory{peers: map[int]*NICInfo{1: {Name: "eth0"}}},
		inventoryAt: time.Now().Add(-5 * time.Second),
	}

	_, ok := rt.cachedInventory(10 * time.Second)
	assert.True(t, ok)

	_, ok = rt.cachedInventory(2 * time.Second)
	assert.False(t, ok)
}

func TestNetnsInformationSyncRuntimeMetadata(t *testing.T) {
	current := &netnsInformation{
		contianerID: "old",
		pid:         map[int]struct{}{11: {}},
		tags:        map[string]string{"k8s_pod_name": "old-pod"},
	}
	fresh := &netnsSnapshot{
		contianerID: "new",
		pid:         map[int]struct{}{22: {}, 33: {}},
		tags:        map[string]string{"k8s_pod_name": "new-pod"},
	}

	current.syncRuntimeMetadata(fresh)

	assert.Equal(t, "new", current.contianerID)
	assert.Equal(t, map[int]struct{}{22: {}, 33: {}}, current.snapshotPIDs())
	assert.Equal(t, map[string]string{"k8s_pod_name": "new-pod"}, current.tags)
}

func TestBuildHostPeerSharedRoutes(t *testing.T) {
	m := &netlogMonitor{
		netnsInfo: map[string]*netnsInformation{
			"ns-a": {},
			"ns-b": {},
		},
		captures: map[string]*namespaceCaptureRuntime{
			"ns-a": {ifaces: map[[2]string]*ifaceInfomation{
				{"veth0", "aa"}: {
					conns:          &TCPConns{nsUID: "ns-a"},
					sharedHostPeer: true,
					routeIfIndex:   11,
				},
				{"eth0", "bb"}: {
					conns:        &TCPConns{nsUID: "ns-a"},
					routeIfIndex: 12,
				},
			}},
			"ns-b": {ifaces: map[[2]string]*ifaceInfomation{
				{"veth1", "cc"}: {
					conns:          &TCPConns{nsUID: "ns-b"},
					sharedHostPeer: true,
					routeIfIndex:   21,
				},
			}},
		},
	}

	routes := m.buildHostPeerSharedRoutes()
	if assert.Len(t, routes, 2) {
		assert.Equal(t, "ns-a", routes[11].nsUID)
		assert.Equal(t, "ns-b", routes[21].nsUID)
	}
}

func TestNormalizeIfIndexes(t *testing.T) {
	got := normalizeIfIndexes([]int{7, 0, 3, 7, -1, 5})
	assert.Equal(t, []int{3, 5, 7}, got)
}

func TestNewSharedHostPeerCBPFFilter(t *testing.T) {
	raw, err := newSharedHostPeerCBPFFilter([]int{11, 7, 11})
	if assert.NoError(t, err) {
		assert.Len(t, raw, len(newBPFFilter())+4)
		assert.Equal(t, uint16(0x20), raw[0].Op)
		assert.Equal(t, uint32(0xfffff008), raw[0].K)
		assert.Equal(t, uint16(0x15), raw[1].Op)
		assert.Equal(t, uint32(7), raw[1].K)
		assert.Equal(t, uint32(11), raw[2].K)
		assert.Equal(t, uint32(0), raw[3].K)
		base := newBPFFilter()
		assert.Equal(t, base[0], raw[4])
		assert.Equal(t, base[len(base)-1], raw[len(raw)-1])
	}
}

func TestSharedHostPeerFilterFingerprint(t *testing.T) {
	assert.Equal(t, "13,21", sharedHostPeerFilterFingerprint([]int{21, 13, 21}))
	assert.Empty(t, sharedHostPeerFilterFingerprint([]int{0, -1}))
}

func TestAttachSharedHostPeerSocketFilterRuntime(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root")
	}

	lo, err := net.InterfaceByName("lo")
	if err != nil {
		t.Fatal(err)
	}

	h, err := newRawsocket(nil, afpacket.OptInterface("lo"), afpacket.OptNumBlocks(1))
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	filter := &sharedHostPeerSocketFilter{}
	mode, ebpfFailed, err := filter.sync(h, []int{lo.Index})
	if err != nil {
		t.Fatal(err)
	}
	defer filter.close()

	t.Logf("runtime attach result: mode=%s ebpf_failed=%t", mode, ebpfFailed)
	assert.Contains(t, []string{"ebpf", "cbpf"}, mode)
	if mode == "ebpf" {
		assert.False(t, ebpfFailed)
	}
}

func TestHostPeerSharedSummary(t *testing.T) {
	c := &hostPeerSharedCapture{
		ns:                "ns-1",
		mode:              "ebpf",
		filterFingerprint: "13,21",
		attachStats: sharedFilterAttachStats{
			syncCount:   3,
			ebpfSuccess: 2,
			cbpfSuccess: 1,
			ebpfFailure: 1,
		},
		routes: map[int]*TCPConns{
			13: {nsUID: "a"},
			21: {nsUID: "b"},
		},
	}

	got := c.summary()
	assert.Contains(t, got, "ns=ns-1")
	assert.Contains(t, got, "mode=ebpf")
	assert.Contains(t, got, "routes=2")
	assert.Contains(t, got, "ifindexes=13,21")
	assert.Contains(t, got, "sync_count=3")
	assert.Contains(t, got, "ebpf_success=2")
	assert.Contains(t, got, "cbpf_success=1")
	assert.Contains(t, got, "ebpf_failure=1")
}

func TestTempNetNSErrorBackoffHelpers(t *testing.T) {
	nsInf := &netnsInformation{}
	assert.NoError(t, checkTempNetNSErrorBackoff(nsInf))

	recordTempNetNSError(nsInf, fmt.Errorf("boom"))
	assert.Error(t, checkTempNetNSErrorBackoff(nsInf))

	nsInf.lastTempNetNSErrorAt = time.Now().Add(-containerNetNSErrorBackoff - time.Second)
	assert.NoError(t, checkTempNetNSErrorBackoff(nsInf))

	clearTempNetNSError(nsInf)
	assert.NoError(t, checkTempNetNSErrorBackoff(nsInf))
}

func TestNetnsHandleCloseWithClosedFD(t *testing.T) {
	nns := &netnsHandle{
		netns: netns.NsHandle(-1),
	}
	nns.close()
}

func TestPortListenInitialDelayForHostNet(t *testing.T) {
	assert.Zero(t, portListenInitialDelay("anything", true))
	assert.Zero(t, portListenInitialDelay("", false))
	assert.Zero(t, portListenInitialDelay(NSUNKNOWN, false))
}

func TestPortListenInitialDelayForContainerNetns(t *testing.T) {
	delayA := portListenInitialDelay("4026532999", false)
	delayB := portListenInitialDelay("4026532999", false)
	assert.Equal(t, delayA, delayB)
	assert.GreaterOrEqual(t, delayA, containerPortListenInitialDelayBase)
	assert.Less(t, delayA, containerPortListenInitialDelayBase+containerPortListenInitialDelayJitter)
}

func TestShouldDeferContainerBootstrap(t *testing.T) {
	budget := 1
	nsInf := &netnsInformation{bootstrapPending: true}

	assert.False(t, shouldDeferContainerBootstrap(nsInf, &budget))
	assert.Equal(t, 0, budget)
	assert.True(t, shouldDeferContainerBootstrap(nsInf, &budget))
}

func TestShouldDeferContainerBootstrapSkipsHostAndCompleted(t *testing.T) {
	budget := 0
	assert.False(t, shouldDeferContainerBootstrap(&netnsInformation{hostNS: true, bootstrapPending: true}, &budget))
	assert.False(t, shouldDeferContainerBootstrap(&netnsInformation{bootstrapPending: false}, &budget))
}

func TestPlanNeedsFallbackBudget(t *testing.T) {
	assert.False(t, planNeedsFallbackBudget(nil))
	assert.False(t, planNeedsFallbackBudget(&capturePlan{mode: captureModeHostPeer}))
	assert.False(t, planNeedsFallbackBudget(&capturePlan{mode: captureModeInNetNS, openInHostNS: true}))
	assert.True(t, planNeedsFallbackBudget(&capturePlan{mode: captureModeInNetNS}))
}

func TestPrepareNamespaceCaptureWorkKeepsHostDirectWhenFallbackBudgetZero(t *testing.T) {
	key := [2]string{"eth0", "aa:bb:cc:dd:ee:ff"}
	work := &namespaceCaptureWork{
		ns: &netnsInformation{
			hostNS: true,
			nsUID:  "host",
		},
		diffPlans: map[[2]string]*capturePlan{
			key: {
				mode:         captureModeInNetNS,
				openInHostNS: true,
				openIface:    key,
				reasonCode:   reasonHostNamespaceOrMissingNIC,
			},
		},
		hostTPs: map[[2]string]*afpacket.TPacket{
			key: nil,
		},
	}

	slots := 0
	monitor := &netlogMonitor{}
	monitor.prepareNamespaceCaptureWork(work, time.Now(), false, &slots)

	_, ok := work.hostTPs[key]
	assert.True(t, ok)
	assert.False(t, work.fallbackDeferred)
	assert.False(t, work.ns.bootstrapPending)
	assert.Equal(t, 0, slots)
}

func TestPrepareNamespaceCaptureWorkSkipsFallbackOpenWhenBudgetZero(t *testing.T) {
	key := [2]string{"eth0", "00:11:22:33:44:55"}
	work := &namespaceCaptureWork{
		ns: &netnsInformation{
			nsUID:       "pod-a",
			contianerID: "ctr-a",
		},
		diffPlans: map[[2]string]*capturePlan{
			key: {
				mode:       captureModeInNetNS,
				openIface:  key,
				reasonCode: reasonHostPeerNotFound,
			},
		},
		containerTPs: map[[2]string]*afpacket.TPacket{
			key: nil,
		},
	}

	slots := 0
	monitor := &netlogMonitor{}
	monitor.prepareNamespaceCaptureWork(work, time.Now(), false, &slots)

	_, ok := work.containerTPs[key]
	assert.False(t, ok)
	assert.True(t, work.fallbackDeferred)
	assert.True(t, work.ns.bootstrapPending)
	assert.Equal(t, 0, slots)
}

func TestPrepareNamespaceCaptureWorkReservesFallbackSlotsDeterministically(t *testing.T) {
	keyA := [2]string{"eth1", "bb"}
	keyB := [2]string{"eth0", "aa"}
	work := &namespaceCaptureWork{
		ns: &netnsInformation{
			nsUID:       "pod-a",
			contianerID: "ctr-a",
		},
		diffPlans: map[[2]string]*capturePlan{
			keyA: {mode: captureModeInNetNS, openIface: keyA, reasonCode: reasonHostPeerNotFound},
			keyB: {mode: captureModeInNetNS, openIface: keyB, reasonCode: reasonHostPeerNotFound},
		},
		containerTPs: map[[2]string]*afpacket.TPacket{
			keyA: nil,
			keyB: nil,
		},
	}

	slots := 1
	monitor := &netlogMonitor{}
	monitor.prepareNamespaceCaptureWork(work, time.Now(), false, &slots)

	_, keepB := work.containerTPs[keyB]
	_, keepA := work.containerTPs[keyA]
	assert.True(t, keepB)
	assert.False(t, keepA)
	assert.True(t, work.fallbackDeferred)
	assert.True(t, work.ns.bootstrapPending)
	assert.Equal(t, 0, slots)
}

func TestActiveFallbackCaptureCount(t *testing.T) {
	m := &netlogMonitor{
		netnsInfo: map[string]*netnsInformation{
			"ns-a": {},
			"ns-b": {},
		},
		captures: map[string]*namespaceCaptureRuntime{
			"ns-a": {ifaces: map[[2]string]*ifaceInfomation{
				{"eth0", "aa"}:  {sharedHostPeer: false},
				{"veth0", "bb"}: {sharedHostPeer: true},
			}},
			"ns-b": {ifaces: map[[2]string]*ifaceInfomation{
				{"eth0", "cc"}: {sharedHostPeer: false},
			}},
		},
	}

	assert.Equal(t, 2, m.activeFallbackCaptureCount())
}

func TestApplyCaptureLimits(t *testing.T) {
	origFallbackSockets := maxFallbackSocketLimit
	origFallbackBlocks := fallbackCaptureSocketBlocks
	origSharedBlocks := sharedCaptureSocketBlocks
	origK8sNetInfo := k8sNetInfo
	defer func() {
		maxFallbackSocketLimit = origFallbackSockets
		fallbackCaptureSocketBlocks = origFallbackBlocks
		sharedCaptureSocketBlocks = origSharedBlocks
		k8sNetInfo = origK8sNetInfo
	}()

	k8sNetInfo = nil
	applyCaptureLimits(&netlogCfg{
		fallbackSockets: 9,
		fallbackBlocks:  4,
		sharedBlocks:    96,
	})

	assert.Equal(t, 9, maxFallbackSocketLimit)
	assert.Equal(t, 4, fallbackCaptureSocketBlocks)
	assert.Equal(t, 96, sharedCaptureSocketBlocks)

	applyCaptureLimits(&netlogCfg{})
	assert.Equal(t, defaultMaxFallbackSocketLimit, maxFallbackSocketLimit)
	assert.Equal(t, defaultFallbackCaptureSocketBlocks, fallbackCaptureSocketBlocks)
	assert.Equal(t, defaultSharedCaptureSocketBlocks, sharedCaptureSocketBlocks)

	k8sNetInfo = &cli.K8sInfo{}
	applyCaptureLimits(&netlogCfg{})
	assert.Equal(t, k8sMaxFallbackSocketLimit, maxFallbackSocketLimit)
	assert.Equal(t, k8sFallbackCaptureSocketBlocks, fallbackCaptureSocketBlocks)
	assert.Equal(t, defaultSharedCaptureSocketBlocks, sharedCaptureSocketBlocks)

	applyCaptureLimits(&netlogCfg{
		fallbackSockets: 12,
		fallbackBlocks:  6,
		sharedBlocks:    192,
	})
	assert.Equal(t, 12, maxFallbackSocketLimit)
	assert.Equal(t, 6, fallbackCaptureSocketBlocks)
	assert.Equal(t, 192, sharedCaptureSocketBlocks)
}

func TestApplyCaptureLimitsResetFromPreviousRun(t *testing.T) {
	origFallbackSockets := maxFallbackSocketLimit
	origFallbackBlocks := fallbackCaptureSocketBlocks
	origSharedBlocks := sharedCaptureSocketBlocks
	origK8sNetInfo := k8sNetInfo
	defer func() {
		maxFallbackSocketLimit = origFallbackSockets
		fallbackCaptureSocketBlocks = origFallbackBlocks
		sharedCaptureSocketBlocks = origSharedBlocks
		k8sNetInfo = origK8sNetInfo
	}()

	k8sNetInfo = nil
	maxFallbackSocketLimit = 99
	fallbackCaptureSocketBlocks = 77
	sharedCaptureSocketBlocks = 55
	applyCaptureLimits(nil)
	assert.Equal(t, defaultMaxFallbackSocketLimit, maxFallbackSocketLimit)
	assert.Equal(t, defaultFallbackCaptureSocketBlocks, fallbackCaptureSocketBlocks)
	assert.Equal(t, defaultSharedCaptureSocketBlocks, sharedCaptureSocketBlocks)
}

func TestShouldTripFallbackFuse(t *testing.T) {
	origLimit := maxFallbackSocketLimit
	maxFallbackSocketLimit = 8
	defer func() { maxFallbackSocketLimit = origLimit }()

	trip, reason := shouldTripFallbackFuse(fallbackProtectionSnapshot{})
	assert.False(t, trip)
	assert.Empty(t, reason)

	trip, reason = shouldTripFallbackFuse(fallbackProtectionSnapshot{
		active:  8,
		pending: 2,
	})
	assert.True(t, trip)
	assert.Contains(t, reason, "socket_saturation")

	trip, reason = shouldTripFallbackFuse(fallbackProtectionSnapshot{
		active: 1,
		drops:  fallbackFuseDropThreshold,
	})
	assert.True(t, trip)
	assert.Contains(t, reason, "drop_burst")

	trip, reason = shouldTripFallbackFuse(fallbackProtectionSnapshot{
		active:  1,
		freezes: fallbackFuseFreezeThreshold,
	})
	assert.True(t, trip)
	assert.Contains(t, reason, "queue_freeze")
}

func TestTripFallbackFuseStopsOnlyFallback(t *testing.T) {
	m := &netlogMonitor{
		netnsInfo: map[string]*netnsInformation{
			"host":  {hostNS: true},
			"pod-a": {hostNS: false},
		},
		captures: map[string]*namespaceCaptureRuntime{
			"host": {ifaces: map[[2]string]*ifaceInfomation{
				{"eth0", "aa"}: {sharedHostPeer: false},
			}},
			"pod-a": {ifaces: map[[2]string]*ifaceInfomation{
				{"eth0", "bb"}:  {sharedHostPeer: false},
				{"veth0", "cc"}: {sharedHostPeer: true},
			}},
		},
	}

	now := time.Now()
	m.tripFallbackFuse(now, "test")

	assert.True(t, m.fallbackFuseActive(now.Add(time.Second)))
	assert.Equal(t, "test", m.lastFallbackFuseReason)
	assert.Equal(t, 1, m.captures["host"].size())
	assert.Equal(t, 1, m.captures["pod-a"].size())
	_, keepShared := m.captures["pod-a"].ifaces[[2]string{"veth0", "cc"}]
	assert.True(t, keepShared)
	assert.True(t, m.netnsInfo["pod-a"].bootstrapPending)
}

func TestPt(t *testing.T) {
	msg := map[string]any{
		"tcp": map[string]any{
			"tcp_series_col_name": []string{
				"txrx", "flags", "seq", "ack_seq", "payload_size", "win", "ts",
			},
			"tcp_series": []*PktTCPHdr{
				{
					TXRX:           "tx",
					Flags:          TCPSYN | TCPACK,
					Seq:            1,
					AckSeq:         2,
					TCPPayloadSize: 3,
					Win:            4,
					TS:             5,
				},
			},
		},
	}

	v, _ := json.Marshal(msg)
	t.Log(string(v))
}

func TestRetrans(t *testing.T) {
	rec := tcpRetransAndReorder{}

	elems := []*tcpSortElem{
		{
			idx:    0,
			txRx:   directionTX,
			seq:    100,
			len:    10,
			ackSeq: 20,
		},
		{
			idx:    1,
			txRx:   directionRX,
			seq:    20,
			len:    10,
			ackSeq: 120,
		},
		{
			idx:    2,
			txRx:   directionTX,
			seq:    130,
			len:    10,
			ackSeq: 30,
		},
		{
			idx:    3,
			txRx:   directionTX,
			seq:    120,
			len:    10,
			ackSeq: 30,
		},
		{
			idx:    4,
			txRx:   directionTX,
			seq:    120,
			len:    10,
			ackSeq: 30,
		},
		{
			idx:    5,
			txRx:   directionRX,
			seq:    30,
			len:    10,
			ackSeq: 140,
		},
		{
			idx:    6,
			txRx:   directionTX,
			seq:    140,
			len:    0,
			ackSeq: 40,
		},

		{
			idx:    7,
			txRx:   directionTX,
			seq:    140,
			len:    0,
			ackSeq: 39,
		},
		{
			idx:    8,
			txRx:   directionTX,
			seq:    140,
			len:    0,
			ackSeq: 39,
		},
	}

	var counter int
	for _, e := range elems {
		if r := rec.insert(e); r == 1 {
			counter++
		}
	}

	t.Log(elems)

	assert.Equal(t, []tcpSortElem{*elems[0], *elems[3], *elems[4], *elems[2], *elems[7], *elems[8], *elems[6]}, rec.txPkts) // tx
	assert.Equal(t, []tcpSortElem{*elems[1], *elems[5]}, rec.rxPkts)
	assert.Equal(t, 2, counter)
}

// func TestM(t *testing.T) {
// 	k8sinfo, err := k8sinfo.NewK8sInfoFromENV()
// 	if err != nil {
// 		t.Error(err)
// 	} else {
// 		k8sinfo.AutoUpdate(context.Background())
// 	}

// 	enableNetlog = true
// 	enabledNetMetric = true
// 	initULID()

// 	SetK8sNetInfo(k8sinfo)

// 	exporter.Init(log)
// 	rt, err := cruntime.NewDockerRuntime("unix:///var/run/docker.sock", "")
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	svc, err := remote.NewRemoteRuntimeService("unix:///var/run/containerd/containerd.sock", time.Second*5)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	m, err := newNetlogMonitor(map[string]string{}, fmt.Sprintf("http://%s%s?input=",
// 		"0.0.0.0:9529", point.Logging.URL())+url.QueryEscape("netlog"),
// 		fmt.Sprintf("http://%s%s?input=",
// 			"0.0.0.0:9529", point.Network.URL())+url.QueryEscape("netlog"), "udp", nil)

// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	m.Run(context.Background(), svc, rt)
// }

func TestStructHTTPLog(t *testing.T) {
	elem := HTTPLogElem{}

	s, err := json.Marshal(elem)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(s))
}

var req = "\x47\x45\x54\x20\x2f\x61\x70\x69\x2f\x64\x61\x74\x61\x2f\x3f\x63" +
	"\x6f\x6e\x74\x65\x78\x74\x4b\x65\x79\x73\x3d\x62\x69\x6e\x6f\x63" +
	"\x75\x6c\x61\x72\x73\x20\x48\x54\x54\x50\x2f\x31\x2e\x31\x0d\x0a" +
	"\x68\x6f\x73\x74\x3a\x20\x6f\x70\x65\x6e\x74\x65\x6c\x65\x6d\x65" +
	"\x74\x72\x79\x2d\x64\x65\x6d\x6f\x2d\x66\x72\x6f\x6e\x74\x65\x6e" +
	"\x64\x70\x72\x6f\x78\x79\x3a\x38\x30\x38\x30\x0d\x0a\x75\x73\x65" +
	"\x72\x2d\x61\x67\x65\x6e\x74\x3a\x20\x70\x79\x74\x68\x6f\x6e\x2d" +
	"\x72\x65\x71\x75\x65\x73\x74\x73\x2f\x32\x2e\x33\x31\x2e\x30\x0d" +
	"\x0a\x61\x63\x63\x65\x70\x74\x2d\x65\x6e\x63\x6f\x64\x69\x6e\x67" +
	"\x3a\x20\x67\x7a\x69\x70\x2c\x20\x64\x65\x66\x6c\x61\x74\x65\x2c" +
	"\x20\x62\x72\x0d\x0a\x61\x63\x63\x65\x70\x74\x3a\x20\x2a\x2f\x2a" +
	"\x0d\x0a\x62\x61\x67\x67\x61\x67\x65\x3a\x20\x73\x79\x6e\x74\x68" +
	"\x65\x74\x69\x63\x5f\x72\x65\x71\x75\x65\x73\x74\x3d\x74\x72\x75" +
	"\x65\x0d\x0a\x78\x2d\x66\x6f\x72\x77\x61\x72\x64\x65\x64\x2d\x70" +
	"\x72\x6f\x74\x6f\x3a\x20\x68\x74\x74\x70\x0d\x0a\x78\x2d\x72\x65" +
	"\x71\x75\x65\x73\x74\x2d\x69\x64\x3a\x20\x64\x39\x33\x33\x36\x36" +
	"\x66\x62\x2d\x31\x63\x62\x34\x2d\x39\x38\x38\x38\x2d\x62\x32\x31" +
	"\x65\x2d\x31\x65\x36\x38\x35\x63\x32\x32\x63\x66\x37\x64\x0d\x0a" +
	"\x78\x2d\x65\x6e\x76\x6f\x79\x2d\x65\x78\x70\x65\x63\x74\x65\x64" +
	"\x2d\x72\x71\x2d\x74\x69\x6d\x65\x6f\x75\x74\x2d\x6d\x73\x3a\x20" +
	"\x31\x35\x30\x30\x30\x0d\x0a\x74\x72\x61\x63\x65\x70\x61\x72\x65" +
	"\x6e\x74\x3a\x20\x30\x30\x2d\x64\x32\x64\x32\x33\x34\x65\x37\x38" +
	"\x38\x65\x31\x37\x31\x66\x37\x63\x35\x64\x61\x38\x61\x39\x31\x31" +
	"\x34\x31\x34\x34\x33\x64\x31\x2d\x31\x64\x61\x61\x66\x33\x34\x32" +
	"\x65\x37\x35\x64\x66\x62\x30\x35\x2d\x30\x31\x0d\x0a\x74\x72\x61" +
	"\x63\x65\x73\x74\x61\x74\x65\x3a\x20\x0d\x0a\x0d\x0a"

var resp = "\x48\x54\x54\x50\x2f\x31\x2e\x31\x20\x33\x30\x38\x20\x50\x65\x72" +
	"\x6d\x61\x6e\x65\x6e\x74\x20\x52\x65\x64\x69\x72\x65\x63\x74\x0d" +
	"\x0a\x4c\x6f\x63\x61\x74\x69\x6f\x6e\x3a\x20\x2f\x61\x70\x69\x2f" +
	"\x64\x61\x74\x61\x3f\x63\x6f\x6e\x74\x65\x78\x74\x4b\x65\x79\x73" +
	"\x3d\x62\x69\x6e\x6f\x63\x75\x6c\x61\x72\x73\x0d\x0a\x52\x65\x66" +
	"\x72\x65\x73\x68\x3a\x20\x30\x3b\x75\x72\x6c\x3d\x2f\x61\x70\x69" +
	"\x2f\x64\x61\x74\x61\x3f\x63\x6f\x6e\x74\x65\x78\x74\x4b\x65\x79" +
	"\x73\x3d\x62\x69\x6e\x6f\x63\x75\x6c\x61\x72\x73\x0d\x0a\x44\x61" +
	"\x74\x65\x3a\x20\x54\x68\x75\x2c\x20\x31\x36\x20\x4e\x6f\x76\x20" +
	"\x32\x30\x32\x33\x20\x30\x36\x3a\x30\x33\x3a\x30\x35\x20\x47\x4d" +
	"\x54\x0d\x0a\x43\x6f\x6e\x6e\x65\x63\x74\x69\x6f\x6e\x3a\x20\x6b" +
	"\x65\x65\x70\x2d\x61\x6c\x69\x76\x65\x0d\x0a\x4b\x65\x65\x70\x2d" +
	"\x41\x6c\x69\x76\x65\x3a\x20\x74\x69\x6d\x65\x6f\x75\x74\x3d\x35" +
	"\x0d\x0a\x54\x72\x61\x6e\x73\x66\x65\x72\x2d\x45\x6e\x63\x6f\x64" +
	"\x69\x6e\x67\x3a\x20\x63\x68\x75\x6e\x6b\x65\x64\x0d\x0a\x0d\x0a" +
	"\x32\x30\x0d\x0a\x2f\x61\x70\x69\x2f\x64\x61\x74\x61\x3f\x63\x6f" +
	"\x6e\x74\x65\x78\x74\x4b\x65\x79\x73\x3d\x62\x69\x6e\x6f\x63\x75" +
	"\x6c\x61\x72\x73\x0d\x0a\x30\x0d\x0a\x0d\x0a"

func TestHTTPLog(t *testing.T) {
	httplog := HTTPLog{}
	_ = resp

	httplog.Handle(nil, 0, []byte(req), 10, &PktTCPHdr{
		TXRX:           "tx",
		Flags:          TCPSYN | TCPACK,
		Seq:            1,
		AckSeq:         2,
		TCPPayloadSize: 3,
		Win:            4,
		TS:             5,
	}, &PMeta{
		SrcIP:   "sd",
		DstIP:   "dd",
		SrcPort: 100,
		DstPort: 200,
	}, 0, 1)

	httplog.Handle(nil, 0, []byte(req), 10, &PktTCPHdr{
		TXRX:           "tx",
		Flags:          TCPSYN | TCPACK,
		Seq:            1,
		AckSeq:         2,
		TCPPayloadSize: 3,
		Win:            4,
		TS:             5,
	}, &PMeta{
		SrcIP:   "sd",
		DstIP:   "dd",
		SrcPort: 100,
		DstPort: 200,
	}, 0, 1)

	t.Log(httplog)
}

func TestParseHTTPRequestMetaHost(t *testing.T) {
	method, path, host, traceID, parentID, ok := parseHTTPRequestMeta([]byte(req))
	if !ok {
		t.Fatal("expected parsed request")
	}

	assert.Equal(t, "GET", method)
	assert.Equal(t, "/api/data/", path)
	assert.Equal(t, "opentelemetry-demo-frontendproxy", host)
	assert.Equal(t, "d2d234e788e171f7c5da8a91141443d1", traceID)
	assert.Equal(t, "1daaf342e75dfb05", parentID)
}
