//go:build linux
// +build linux

package l4log

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/vishvananda/netns"
)

func TestCmpAndCleanNetNsNICRemovesNamespaceState(t *testing.T) {
	ns, err := netns.GetFromPid(os.Getpid())
	if err != nil {
		t.Fatalf("get current netns: %v", err)
	}

	info := &netnsInformation{
		nsUID:          "test-ns",
		nns:            newNetNsHandle(true, true, ns),
		missingSinceAt: time.Now().Add(-namespaceDiscoveryGracePeriod - time.Second),
	}

	monitor := &netlogMonitor{
		netnsInfo: map[string]*netnsInformation{
			info.nsUID: info,
		},
		captures: map[string]*namespaceCaptureRuntime{
			info.nsUID: {ifaces: map[[2]string]*ifaceInfomation{
				{"eth0", "aa:bb:cc:dd:ee:ff"}: {},
			}},
		},
		lastPlans: map[string]string{
			"test-ns|eth0|aa:bb:cc:dd:ee:ff|1|2": "plan",
			"other|eth0|aa:bb:cc:dd:ee:ff|1|2":   "plan",
		},
		hostPeerOwners: map[[2]string]string{
			{"veth0", "aa:bb:cc:dd:ee:ff"}: "test-ns",
			{"veth1", "aa:bb:cc:dd:ee:00"}: "other",
		},
	}

	monitor.CmpAndCleanNetNsNIC(map[string]*netnsSnapshot{})

	if info.nns.netns.IsOpen() {
		t.Fatal("expected removed namespace handle to be closed")
	}
	if _, ok := monitor.netnsInfo[info.nsUID]; ok {
		t.Fatal("expected removed namespace entry to be deleted")
	}
	if _, ok := monitor.lastPlans["test-ns|eth0|aa:bb:cc:dd:ee:ff|1|2"]; ok {
		t.Fatal("expected removed namespace plan cache to be deleted")
	}
	if _, ok := monitor.lastPlans["other|eth0|aa:bb:cc:dd:ee:ff|1|2"]; !ok {
		t.Fatal("expected unrelated namespace plan cache to remain")
	}
	if _, ok := monitor.hostPeerOwners[[2]string{"veth0", "aa:bb:cc:dd:ee:ff"}]; ok {
		t.Fatal("expected removed namespace host-peer owner to be deleted")
	}
	if _, ok := monitor.hostPeerOwners[[2]string{"veth1", "aa:bb:cc:dd:ee:00"}]; !ok {
		t.Fatal("expected unrelated host-peer owner to remain")
	}
}

func TestCmpAndCleanNetNsNICDefersPruneWithinGracePeriod(t *testing.T) {
	ns, err := netns.GetFromPid(os.Getpid())
	if err != nil {
		t.Fatalf("get current netns: %v", err)
	}

	info := &netnsInformation{
		nsUID: "test-ns",
		nns:   newNetNsHandle(true, true, ns),
	}

	monitor := &netlogMonitor{
		netnsInfo: map[string]*netnsInformation{
			info.nsUID: info,
		},
		captures:       map[string]*namespaceCaptureRuntime{},
		lastPlans:      map[string]string{},
		hostPeerOwners: map[[2]string]string{},
	}

	monitor.CmpAndCleanNetNsNIC(map[string]*netnsSnapshot{})

	if _, ok := monitor.netnsInfo[info.nsUID]; !ok {
		t.Fatal("expected namespace to remain during discovery grace period")
	}
	if info.missingSinceAt.IsZero() {
		t.Fatal("expected missing namespace to record grace-period start")
	}
	if !info.nns.netns.IsOpen() {
		t.Fatal("expected namespace handle to remain open during grace period")
	}
}

func TestCmpAndCleanNetNsNICClearsMissingStateWhenNamespaceReturns(t *testing.T) {
	ns, err := netns.GetFromPid(os.Getpid())
	if err != nil {
		t.Fatalf("get current netns: %v", err)
	}

	info := &netnsInformation{
		nsUID:          "test-ns",
		nns:            newNetNsHandle(true, true, ns),
		missingSinceAt: time.Now().Add(-10 * time.Second),
	}

	monitor := &netlogMonitor{
		netnsInfo: map[string]*netnsInformation{
			info.nsUID: info,
		},
		captures:       map[string]*namespaceCaptureRuntime{},
		lastPlans:      map[string]string{},
		hostPeerOwners: map[[2]string]string{},
	}

	monitor.CmpAndCleanNetNsNIC(map[string]*netnsSnapshot{
		info.nsUID: {nsUID: info.nsUID},
	})

	if !info.missingSinceAt.IsZero() {
		t.Fatal("expected rediscovered namespace to clear missing grace state")
	}
	if info.lastSeenAt.IsZero() {
		t.Fatal("expected rediscovered namespace to refresh last-seen timestamp")
	}
}

func TestCmpAndCleanNetNsNICClearsHostRuntime(t *testing.T) {
	ns, err := netns.GetFromPid(os.Getpid())
	if err != nil {
		t.Fatalf("get current netns: %v", err)
	}

	info := &netnsInformation{
		hostNS:         true,
		nsUID:          "host-ns",
		nns:            newNetNsHandle(true, true, ns),
		missingSinceAt: time.Now().Add(-namespaceDiscoveryGracePeriod - time.Second),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	monitor := &netlogMonitor{
		netnsInfo: map[string]*netnsInformation{
			info.nsUID: info,
		},
		captures:       map[string]*namespaceCaptureRuntime{},
		lastPlans:      map[string]string{},
		hostPeerOwners: map[[2]string]string{},
		hostRuntime: &hostNamespaceRuntime{
			ns: info,
			inventory: hostNICInventory{
				peers:          map[int]*NICInfo{1: {Name: "eth0"}},
				peerCandidates: map[int][]*NICInfo{2: {{Name: "veth0"}}},
			},
			inventoryAt: time.Now(),
			sharedCapture: &hostPeerSharedCapture{
				ctx:    ctx,
				cancel: cancel,
				ns:     info.nsUID,
				routes: map[int]*TCPConns{},
			},
		},
	}

	monitor.CmpAndCleanNetNsNIC(map[string]*netnsSnapshot{})

	if monitor.hostRuntime.ns != nil {
		t.Fatal("expected host runtime namespace to be cleared")
	}
	if !monitor.hostRuntime.inventoryAt.IsZero() {
		t.Fatal("expected host runtime inventory timestamp to be cleared")
	}
	if len(monitor.hostRuntime.inventory.peers) != 0 {
		t.Fatal("expected host runtime inventory peers to be cleared")
	}
	if len(monitor.hostRuntime.inventory.peerCandidates) != 0 {
		t.Fatal("expected host runtime inventory peer candidates to be cleared")
	}
	if monitor.hostRuntime.sharedCapture != nil {
		t.Fatal("expected host runtime shared capture to be cleared")
	}
}

func TestNetnsInformationSnapshotAndReplacePIDs(t *testing.T) {
	info := &netnsInformation{
		pid: map[int]struct{}{
			10: {},
			20: {},
		},
	}

	got := info.snapshotPIDs()
	if _, ok := got[10]; !ok {
		t.Fatal("expected snapshot to include pid 10")
	}

	delete(got, 10)
	if _, ok := info.pid[10]; !ok {
		t.Fatal("snapshot should be detached from source map")
	}

	info.replacePIDs(map[int]struct{}{30: {}})
	if len(info.pid) != 1 {
		t.Fatalf("expected one pid after replace, got %d", len(info.pid))
	}
	if _, ok := info.pid[30]; !ok {
		t.Fatal("expected pid 30 after replace")
	}
}

func TestHostNetNSMergeKeepsExistingHandle(t *testing.T) {
	existingNS, err := netns.GetFromPid(os.Getpid())
	if err != nil {
		t.Fatalf("get existing netns: %v", err)
	}

	freshNS, err := netns.GetFromPid(os.Getpid())
	if err != nil {
		t.Fatalf("get fresh netns: %v", err)
	}

	existing := &netnsInformation{
		hostNS: true,
		nsUID:  NSInode(existingNS),
		nns:    newNetNsHandle(true, true, existingNS),
	}
	fresh := &netnsSnapshot{
		hostNS: true,
		nsUID:  NSInode(freshNS),
		nns:    newNetNsHandle(true, true, freshNS),
	}
	t.Cleanup(existing.Close)
	t.Cleanup(fresh.Close)

	existing.syncRuntimeState(fresh)
	fresh.Close()

	if existing.nns == nil || !existing.nns.netns.IsOpen() {
		t.Fatal("expected existing host namespace handle to remain open")
	}
	if err := CallWithNetNS(existing.nns.netns, func() {}); err != nil {
		t.Fatalf("expected preserved host namespace handle to stay usable: %v", err)
	}
}

func TestMergeNamespaceStateAdoptsNewHostSnapshot(t *testing.T) {
	hostNS, err := netns.GetFromPid(os.Getpid())
	if err != nil {
		t.Fatalf("get host netns: %v", err)
	}

	snapshot := &netnsSnapshot{
		hostNS: true,
		nsUID:  NSInode(hostNS),
		nns:    newNetNsHandle(true, true, hostNS),
	}
	t.Cleanup(snapshot.Close)

	monitor := &netlogMonitor{
		netnsInfo:      map[string]*netnsInformation{},
		captures:       map[string]*namespaceCaptureRuntime{},
		lastPlans:      map[string]string{},
		hostPeerOwners: map[[2]string]string{},
		hostRuntime:    &hostNamespaceRuntime{},
	}

	merged := monitor.mergeNamespaceState(snapshot.nsUID, snapshot)
	if merged == nil {
		t.Fatal("expected merged namespace state")
	}
	if merged != monitor.netnsInfo[snapshot.nsUID] {
		t.Fatal("expected merged namespace to be registered in monitor")
	}
	if monitor.hostRuntime.ns != merged {
		t.Fatal("expected host runtime to point at merged namespace")
	}
	if snapshot.nns != nil {
		t.Fatal("expected snapshot netns handle ownership to transfer into registry state")
	}
	if merged.nns == nil || !merged.nns.netns.IsOpen() {
		t.Fatal("expected merged host namespace handle to stay open")
	}
	if err := CallWithNetNS(merged.nns.netns, func() {}); err != nil {
		t.Fatalf("expected merged host namespace handle to stay usable: %v", err)
	}
}

func TestContainerNetNSMergeRefreshesClosedHandle(t *testing.T) {
	freshNS, err := netns.GetFromPid(os.Getpid())
	if err != nil {
		t.Fatalf("get fresh netns: %v", err)
	}

	existing := &netnsInformation{
		nsUID: NSInode(freshNS),
		nns:   &netnsHandle{hostNet: false, includeLo: true, nsStr: NSInode(freshNS), netns: netns.NsHandle(-1)},
		pid:   map[int]struct{}{os.Getpid(): {}},
	}
	fresh := &netnsSnapshot{
		nsUID: NSInode(freshNS),
		nns:   newNetNsHandle(false, true, freshNS),
		pid:   map[int]struct{}{os.Getpid(): {}},
	}
	t.Cleanup(existing.Close)
	t.Cleanup(fresh.Close)

	existing.syncRuntimeState(fresh)
	fresh.Close()

	if existing.nns == nil || !existing.nns.netns.IsOpen() {
		t.Fatal("expected existing container namespace handle to be refreshed from fresh state")
	}
	if err := CallWithNetNS(existing.nns.netns, func() {}); err != nil {
		t.Fatalf("expected refreshed container namespace handle to stay usable: %v", err)
	}
}

func TestBuildNICIPsAndFingerprintStableOrder(t *testing.T) {
	nics := []*NICInfo{
		{
			Name:    "eth1",
			MAC:     "bb",
			IfIndex: 2,
			Addrs: []*net.IPNet{
				{IP: net.ParseIP("10.0.0.2"), Mask: net.CIDRMask(24, 32)},
			},
		},
		{
			Name:    "eth0",
			MAC:     "aa",
			IfIndex: 1,
			Addrs: []*net.IPNet{
				{IP: net.ParseIP("10.0.0.1"), Mask: net.CIDRMask(24, 32)},
			},
		},
	}

	ips, fp := buildNICIPsAndFingerprint(nics)
	if len(ips) != 2 || ips[0] != "10.0.0.1" || ips[1] != "10.0.0.2" {
		t.Fatalf("unexpected sorted ips: %v", ips)
	}
	if fp == "" {
		t.Fatal("expected non-empty fingerprint")
	}

	ips2, fp2 := buildNICIPsAndFingerprint([]*NICInfo{nics[1], nics[0]})
	if fp2 != fp {
		t.Fatalf("fingerprint should be stable across order changes: %q != %q", fp2, fp)
	}
	if len(ips2) != len(ips) || ips2[0] != ips[0] || ips2[1] != ips[1] {
		t.Fatalf("sorted ips should be stable across order changes: %v != %v", ips2, ips)
	}
}

func TestClonePIDSetAndTags(t *testing.T) {
	pids := clonePIDSet(map[int]struct{}{1: {}, 2: {}})
	delete(pids, 1)

	srcTags := map[string]string{"a": "1"}
	tags := cloneTags(srcTags)
	tags["a"] = "2"

	if _, ok := pids[1]; ok {
		t.Fatal("cloned pid set should be mutable independently")
	}
	if srcTags["a"] != "1" {
		t.Fatal("cloned tags should not mutate source tags")
	}
}

func TestMaybeLogCapturePlanCachesFingerprint(t *testing.T) {
	monitor := &netlogMonitor{
		stats:     newCapturePlanStats(),
		lastPlans: map[string]string{},
	}
	nsInf := &netnsInformation{nsUID: "ns-a", contianerID: "ctr-a"}
	nic := &NICInfo{Name: "eth0", MAC: "aa", IfIndex: 1, IfLink: 2}

	monitor.maybeLogCapturePlan(nsInf, nic, fallbackCapturePlan(nsInf, nic, reasonHostPeerNotFound))
	monitor.maybeLogCapturePlan(nsInf, nic, fallbackCapturePlan(nsInf, nic, reasonHostPeerNotFound))

	if got := monitor.stats.delta.fallbackInNetNS; got != 1 {
		t.Fatalf("expected duplicate plan fingerprint to increment delta once, got %d", got)
	}

	monitor.maybeLogCapturePlan(nsInf, nic, fallbackCapturePlan(nsInf, nic, reasonHostPeerConflict))
	if got := monitor.stats.delta.fallbackConflict; got != 1 {
		t.Fatalf("expected changed plan fingerprint to increment new delta counter, got %d", got)
	}
}
