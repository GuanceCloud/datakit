//go:build linux
// +build linux

package l4log

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type netnsInformation struct {
	hostNS      bool
	nsUID       string
	contianerID string
	nns         *netnsHandle

	pidMu sync.RWMutex
	pid   map[int]struct{}

	tags map[string]string

	nicIPsCache       []string
	nicIPsFingerprint string

	nicInfoCache   []*NICInfo
	nicInfoCacheAt time.Time

	netnsID   int
	netnsIDAt time.Time

	lastTempNetNSError   string
	lastTempNetNSErrorAt time.Time

	bootstrapPending        bool
	lastBootstrapDeferredAt time.Time
	lastSeenAt              time.Time
	missingSinceAt          time.Time
}

type netnsSnapshot struct {
	hostNS      bool
	nsUID       string
	contianerID string
	nns         *netnsHandle
	pid         map[int]struct{}
	tags        map[string]string
}

const (
	containerNICCacheTTL               = 15 * time.Second
	hostNICCacheTTL                    = 15 * time.Second
	unstableContainerNICCacheTTL       = 3 * time.Second
	unstableHostNICCacheTTL            = 3 * time.Second
	containerNetNSErrorBackoff         = 10 * time.Second
	containerBootstrapBudgetPerRound   = 8
	containerBootstrapDeferLogInterval = time.Minute
	namespaceDiscoveryGracePeriod      = time.Minute
	fallbackFuseCooldown               = 10 * time.Minute
	fallbackFuseDropThreshold          = 1024
	fallbackFuseFreezeThreshold        = 32
)

func (nsInf *netnsInformation) Close() {
	if nsInf.nns != nil {
		nsInf.nns.close()
	}
}

func (snap *netnsSnapshot) Close() {
	if snap == nil || snap.nns == nil {
		return
	}
	snap.nns.close()
}

func (nsInf *netnsInformation) snapshotPIDs() map[int]struct{} {
	if nsInf == nil {
		return nil
	}
	nsInf.pidMu.RLock()
	defer nsInf.pidMu.RUnlock()

	return clonePIDSet(nsInf.pid)
}

func (nsInf *netnsInformation) replacePIDs(pids map[int]struct{}) {
	if nsInf == nil {
		return
	}
	nsInf.pidMu.Lock()
	nsInf.pid = clonePIDSet(pids)
	nsInf.pidMu.Unlock()
}

func (nsInf *netnsInformation) syncRuntimeMetadata(fresh *netnsSnapshot) {
	if nsInf == nil || fresh == nil {
		return
	}
	nsInf.replacePIDs(fresh.pid)
	nsInf.contianerID = fresh.contianerID
	nsInf.tags = cloneTags(fresh.tags)
	nsInf.lastSeenAt = time.Now()
	nsInf.missingSinceAt = time.Time{}
}

func (nsInf *netnsInformation) syncRuntimeState(fresh *netnsSnapshot) {
	if nsInf == nil || fresh == nil {
		return
	}

	nsInf.syncRuntimeMetadata(fresh)
	if fresh.nns == nil {
		return
	}
	if nsInf.nns == nil {
		nsInf.nns = fresh.nns
		fresh.nns = nil
		return
	}

	nsInf.nns.hostNet = fresh.nns.hostNet
	nsInf.nns.includeLo = fresh.nns.includeLo
	nsInf.nns.nsStr = fresh.nns.nsStr
	if !nsInf.nns.netns.IsOpen() && fresh.nns.netns.IsOpen() {
		nsInf.nns.netns = fresh.nns.netns
		fresh.nns.netns = netns.NsHandle(-1)
	}
}

func shouldDeferContainerBootstrap(nsInf *netnsInformation, budget *int) bool {
	if nsInf == nil || nsInf.hostNS || !nsInf.bootstrapPending {
		return false
	}
	if budget == nil {
		return true
	}
	if *budget <= 0 {
		return true
	}
	*budget--
	return false
}

func containerNICCacheTTLForNS(nsInf *netnsInformation) time.Duration {
	if nsInf == nil {
		return containerNICCacheTTL
	}
	if nsInf.bootstrapPending {
		return unstableContainerNICCacheTTL
	}
	return containerNICCacheTTL
}

func (m *netlogMonitor) activeFallbackCaptureCount() int {
	if m == nil {
		return 0
	}

	total := 0
	for nsUID := range m.netnsInfo {
		total += m.namespaceFallbackCaptureCount(nsUID)
	}
	return total
}

func (m *netlogMonitor) fallbackFuseActive(now time.Time) bool {
	if m == nil {
		return false
	}
	return now.Before(m.fallbackFuseUntil)
}

func (m *netlogMonitor) collectFallbackProtectionSnapshot() fallbackProtectionSnapshot {
	if m == nil {
		return fallbackProtectionSnapshot{}
	}

	var snapshot fallbackProtectionSnapshot
	for nsUID, nsInf := range m.netnsInfo {
		if !nsInf.hostNS && nsInf.bootstrapPending {
			snapshot.pending++
		}
		m.collectNamespaceFallbackProtection(nsUID, &snapshot)
	}

	return snapshot
}

func (m *netlogMonitor) tripFallbackFuse(now time.Time, reason string) {
	if m == nil {
		return
	}

	m.fallbackFuseUntil = now.Add(fallbackFuseCooldown)
	stopped := 0
	for nsUID, nsInf := range m.netnsInfo {
		if nsInf == nil || nsInf.hostNS {
			continue
		}
		nsInf.bootstrapPending = true
		stopped += m.stopNamespaceFallbackCaptures(nsUID)
	}

	if reason != m.lastFallbackFuseReason || time.Since(m.lastFallbackFuseLoggedAt) >= time.Minute {
		log.Warnf("trip fallback packet capture fuse reason=%s cooldown=%s stopped=%d",
			reason, fallbackFuseCooldown, stopped)
		m.lastFallbackFuseReason = reason
		m.lastFallbackFuseLoggedAt = now
	}
}

func (m *netlogMonitor) CmpAndCleanNetNsNIC(netnsInfo map[string]*netnsSnapshot) {
	now := time.Now()
	for nsUID, infom := range m.netnsInfo {
		if _, ok := netnsInfo[nsUID]; ok {
			if infom != nil {
				infom.lastSeenAt = now
				infom.missingSinceAt = time.Time{}
			}
			continue
		}
		if infom != nil {
			if infom.missingSinceAt.IsZero() {
				infom.missingSinceAt = now
				continue
			}
			if now.Sub(infom.missingSinceAt) < namespaceDiscoveryGracePeriod {
				continue
			}
		}
		if infom == nil {
			continue
		}
		m.stopAllNamespaceCaptures(nsUID)
		infom.Close()
		delete(m.netnsInfo, nsUID)
		delete(m.captures, nsUID)
		m.clearPlansForNS(nsUID)
		m.clearHostPeerOwnersForNS(nsUID)
		if m.hostRuntime != nil {
			m.hostRuntime.clearNamespace(nsUID)
		}
	}
}

func newNetnsInformationFromSnapshot(fresh *netnsSnapshot) *netnsInformation {
	if fresh == nil {
		return nil
	}
	nsInf := &netnsInformation{
		hostNS:      fresh.hostNS,
		nsUID:       fresh.nsUID,
		contianerID: fresh.contianerID,
		nns:         fresh.nns,
		tags:        cloneTags(fresh.tags),
	}
	nsInf.replacePIDs(fresh.pid)
	nsInf.bootstrapPending = !fresh.hostNS
	nsInf.lastSeenAt = time.Now()
	fresh.nns = nil
	return nsInf
}

func (m *netlogMonitor) mergeNamespaceState(nsUID string, fresh *netnsSnapshot) *netnsInformation {
	preNsInf, ok := m.netnsInfo[nsUID]
	if !ok {
		preNsInf = newNetnsInformationFromSnapshot(fresh)
		m.netnsInfo[nsUID] = preNsInf
		if preNsInf != nil && preNsInf.hostNS && m.hostRuntime != nil {
			m.hostRuntime.setNamespace(preNsInf)
		}
		if fresh != nil {
			fresh.Close()
		}
		return preNsInf
	}

	preNsInf.syncRuntimeState(fresh)
	if preNsInf.hostNS && m.hostRuntime != nil {
		m.hostRuntime.setNamespace(preNsInf)
	}
	if fresh != nil {
		fresh.Close()
	}
	return preNsInf
}

func (m *netlogMonitor) clearPlansForNS(nsUID string) {
	if nsUID == "" {
		return
	}
	prefix := nsUID + "|"
	for k := range m.lastPlans {
		if strings.HasPrefix(k, prefix) {
			delete(m.lastPlans, k)
		}
	}
}

func (m *netlogMonitor) clearHostPeerOwnersForNS(nsUID string) {
	if nsUID == "" {
		return
	}
	for iface, ownerNS := range m.hostPeerOwners {
		if ownerNS == nsUID {
			delete(m.hostPeerOwners, iface)
		}
	}
}

func anyPID(pids map[int]struct{}) (int, error) {
	for pid := range pids {
		if pid <= 0 {
			delete(pids, pid)
			continue
		}
		if _, err := os.Stat(fmt.Sprintf("/proc/%d", pid)); err != nil {
			delete(pids, pid)
			continue
		}
		return pid, nil
	}
	return 0, fmt.Errorf("no live pid for namespace")
}

func openContainerNetNS(nsInf *netnsInformation) (int, netns.NsHandle, error) {
	if nsInf == nil {
		return 0, netns.NsHandle(-1), fmt.Errorf("missing namespace info")
	}

	pids := nsInf.snapshotPIDs()
	for {
		pid, err := anyPID(pids)
		if err != nil {
			return 0, netns.NsHandle(-1), err
		}

		ns, err := netns.GetFromPid(pid)
		if err != nil {
			delete(pids, pid)
			continue
		}

		if nsInf.nsUID != "" {
			if got := NSInode(ns); got != nsInf.nsUID {
				_ = ns.Close()
				delete(pids, pid)
				continue
			}
		}

		return pid, ns, nil
	}
}

func checkTempNetNSErrorBackoff(nsInf *netnsInformation) error {
	if nsInf == nil {
		return nil
	}
	if nsInf.lastTempNetNSError == "" {
		return nil
	}
	if time.Since(nsInf.lastTempNetNSErrorAt) >= containerNetNSErrorBackoff {
		return nil
	}
	return fmt.Errorf("temporary netns access backed off: last_error=%s", nsInf.lastTempNetNSError)
}

func recordTempNetNSError(nsInf *netnsInformation, err error) {
	if nsInf == nil || err == nil {
		return
	}
	nsInf.lastTempNetNSError = err.Error()
	nsInf.lastTempNetNSErrorAt = time.Now()
}

func clearTempNetNSError(nsInf *netnsInformation) {
	if nsInf == nil {
		return
	}
	nsInf.lastTempNetNSError = ""
	nsInf.lastTempNetNSErrorAt = time.Time{}
}

func ensureContainerNetNS(nsInf *netnsInformation) error {
	if nsInf == nil {
		return fmt.Errorf("missing namespace info")
	}
	if nsInf.hostNS {
		if nsInf.nns == nil || !nsInf.nns.netns.IsOpen() {
			return fmt.Errorf("missing host netns handle")
		}
		return nil
	}
	if nsInf.nns == nil {
		return fmt.Errorf("missing netns handle")
	}
	if nsInf.nns.netns.IsOpen() {
		return nil
	}
	if err := checkTempNetNSErrorBackoff(nsInf); err != nil {
		return err
	}

	_, ns, err := openContainerNetNS(nsInf)
	if err != nil {
		recordTempNetNSError(nsInf, err)
		return err
	}
	nsInf.nns.netns = ns
	nsInf.nns.nsStr = NSInode(ns)
	clearTempNetNSError(nsInf)
	return nil
}

func lookupNetNSID(nsInf *netnsInformation) int {
	if nsInf == nil || nsInf.hostNS {
		return -1
	}
	if ttl := containerNICCacheTTLForNS(nsInf); !nsInf.netnsIDAt.IsZero() && time.Since(nsInf.netnsIDAt) < ttl {
		return nsInf.netnsID
	}

	if err := ensureContainerNetNS(nsInf); err != nil {
		nsInf.netnsID = -1
		nsInf.netnsIDAt = time.Now()
		return -1
	}

	netnsID, err := netlink.GetNetNsIdByFd(int(nsInf.nns.netns))
	if err != nil {
		log.Debugf("lookup netns id failed: ns=%s err=%s", nsInf.nsUID, err)
		netnsID = -1
	}

	nsInf.netnsID = netnsID
	nsInf.netnsIDAt = time.Now()
	return netnsID
}

func clonePIDSet(src map[int]struct{}) map[int]struct{} {
	if len(src) == 0 {
		return map[int]struct{}{}
	}
	dst := make(map[int]struct{}, len(src))
	for pid := range src {
		dst[pid] = struct{}{}
	}
	return dst
}

func cloneTags(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
