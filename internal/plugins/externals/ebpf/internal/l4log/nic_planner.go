//go:build linux
// +build linux

package l4log

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket/afpacket"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/exporter"
)

type captureMode string

const (
	captureModeInNetNS  captureMode = "in-netns"
	captureModeHostPeer captureMode = "host-peer"
)

type capturePlan struct {
	mode captureMode

	openInHostNS bool
	openNS       string
	openIface    [2]string
	routeIfIndex int

	logicalNIC *NICInfo
	trustLocal bool
	reasonCode string
}

type namespaceCaptureWork struct {
	ns *netnsInformation

	diffPlans    map[[2]string]*capturePlan
	containerTPs map[[2]string]*afpacket.TPacket
	hostTPs      map[[2]string]*afpacket.TPacket

	nicIPs           []string
	nicFingerprint   string
	fallbackDeferred bool
}

const (
	reasonHostNamespaceOrMissingNIC = "host_namespace_or_missing_nic"
	reasonIfLinkUnavailable         = "iflink_unavailable"
	reasonHostPeerNotFound          = "host_peer_not_found"
	reasonHostPeerNotEligible       = "host_peer_not_eligible"
	reasonHostPeerAmbiguous         = "host_peer_ambiguous"
	reasonHostPeerSelected          = "host_peer_selected"
	reasonHostPeerConflict          = "host_peer_conflict"
)

const (
	nicGroupHostPeerShared = "host_peer_shared"
	nicGroupHostDirect     = "host_direct"
	nicGroupFallback       = "fallback"
)

type nicGroupSnapshot struct {
	counts map[string]int
	routes map[string]int
}

func newNICGroupSnapshot() *nicGroupSnapshot {
	return &nicGroupSnapshot{
		counts: map[string]int{},
		routes: map[string]int{},
	}
}

func (s *nicGroupSnapshot) add(plan *capturePlan) {
	if s == nil || plan == nil {
		return
	}
	group := classifyNICGroup(plan)
	virtual := false
	if plan.logicalNIC != nil {
		virtual = plan.logicalNIC.VIface || plan.logicalNIC.IsVirtual
	}
	key := nicGroupKey(group, virtual)
	s.counts[key]++
	if plan.routeIfIndex != 0 {
		s.routes[key]++
	}
}

func (s *nicGroupSnapshot) export() {
	if s == nil {
		return
	}
	for _, group := range []string{nicGroupHostPeerShared, nicGroupHostDirect, nicGroupFallback} {
		for _, virtual := range []bool{false, true} {
			key := nicGroupKey(group, virtual)
			exporter.ObserveNICGroupCount(group, virtual, s.counts[key])
			exporter.ObserveNICGroupRouteCount(group, virtual, s.routes[key])
		}
	}
}

func classifyNICGroup(plan *capturePlan) string {
	if plan == nil {
		return nicGroupFallback
	}
	switch {
	case plan.mode == captureModeHostPeer:
		return nicGroupHostPeerShared
	case plan.logicalNIC != nil && plan.logicalNIC.HostIface:
		return nicGroupHostDirect
	default:
		return nicGroupFallback
	}
}

func nicGroupKey(group string, virtual bool) string {
	return group + "/" + strconv.FormatBool(virtual)
}

func (m *netlogMonitor) buildNamespaceCaptureWork(nsInf *netnsInformation, hostInv hostNICInventory,
	claimedHostPeerOwners map[[2]string]string, currentSnapshot *snapshotPlanCounters,
	nicGroups *nicGroupSnapshot,
) (*namespaceCaptureWork, error) {
	nics, err := m.captureNICsForNS(nsInf, hostInv)
	if err != nil {
		return nil, err
	}

	nicIPs, nicFingerprint := buildNICIPsAndFingerprint(nics)
	work := &namespaceCaptureWork{
		ns:             nsInf,
		diffPlans:      map[[2]string]*capturePlan{},
		containerTPs:   map[[2]string]*afpacket.TPacket{},
		hostTPs:        map[[2]string]*afpacket.TPacket{},
		nicIPs:         nicIPs,
		nicFingerprint: nicFingerprint,
	}

	for _, v := range nics {
		plan := resolveCapturePlan(nsInf, v, hostInv)
		if plan.mode == captureModeHostPeer {
			if ownerNS, ok := claimedHostPeerOwners[plan.openIface]; ok && ownerNS != nsInf.nsUID {
				log.Warnf("capture plan conflict ns=%s container_id=%s nic=%s requested_iface=%s existing_owner_ns=%s action=fallback-to-in-netns",
					nsInf.nsUID, nsInf.contianerID, v.Name, plan.openIface[0], ownerNS)
				plan = fallbackCapturePlan(nsInf, v, reasonHostPeerConflict)
			} else if ownerNS, ok := m.hostPeerOwners[plan.openIface]; ok && ownerNS != nsInf.nsUID {
				log.Warnf("capture plan cached-owner conflict ns=%s container_id=%s nic=%s requested_iface=%s cached_owner_ns=%s action=fallback-to-in-netns",
					nsInf.nsUID, nsInf.contianerID, v.Name, plan.openIface[0], ownerNS)
				plan = fallbackCapturePlan(nsInf, v, reasonHostPeerConflict)
			} else {
				claimedHostPeerOwners[plan.openIface] = nsInf.nsUID
			}
		}
		applyPlanCounter(currentSnapshot, plan)
		nicGroups.add(plan)
		m.maybeLogCapturePlan(nsInf, v, plan)
		work.diffPlans[plan.openIface] = plan
		if plan.mode == captureModeHostPeer {
			continue
		}
		if plan.openInHostNS {
			work.hostTPs[plan.openIface] = nil
		} else {
			work.containerTPs[plan.openIface] = nil
		}
	}

	return work, nil
}

func (m *netlogMonitor) captureNICsForNS(nsInf *netnsInformation, hostInv hostNICInventory) ([]*NICInfo, error) {
	if nsInf == nil || nsInf.nns == nil {
		return nil, fmt.Errorf("missing netns info")
	}
	if nsInf.hostNS {
		return hostInv.captureNICs, nil
	}
	if ttl := containerNICCacheTTLForNS(nsInf); len(nsInf.nicInfoCache) > 0 && time.Since(nsInf.nicInfoCacheAt) < ttl {
		return cloneNICInfos(nsInf.nicInfoCache), nil
	}
	if err := ensureContainerNetNS(nsInf); err != nil {
		return nil, err
	}
	nics, err := nsInf.nns.nicInfoWithVirtual(false)
	if err != nil {
		recordTempNetNSError(nsInf, err)
		nsInf.nns.close()
		return nil, err
	}
	clearTempNetNSError(nsInf)
	nsInf.nicInfoCache = cloneNICInfos(nics)
	nsInf.nicInfoCacheAt = time.Now()
	return nics, nil
}

func (m *netlogMonitor) maybeLogCapturePlan(nsInf *netnsInformation, nic *NICInfo, plan *capturePlan) {
	if nsInf == nil || nic == nil || plan == nil {
		return
	}
	key := capturePlanKey(nsInf, nic)
	fp := capturePlanFingerprint(plan)
	if m.lastPlans[key] == fp {
		return
	}
	if m.stats != nil {
		m.stats.IncDelta(plan)
	}
	m.lastPlans[key] = fp
	logCapturePlan(nsInf, nic, plan)
}

func capturePlanKey(nsInf *netnsInformation, nic *NICInfo) string {
	if nsInf == nil || nic == nil {
		return ""
	}
	var buf []byte
	buf = append(buf, nsInf.nsUID...)
	buf = append(buf, '|')
	buf = append(buf, nic.Name...)
	buf = append(buf, '|')
	buf = append(buf, nic.MAC...)
	buf = append(buf, '|')
	buf = strconv.AppendInt(buf, int64(nic.IfIndex), 10)
	buf = append(buf, '|')
	buf = strconv.AppendInt(buf, int64(nic.IfLink), 10)
	return string(buf)
}

func capturePlanFingerprint(plan *capturePlan) string {
	if plan == nil {
		return ""
	}
	var buf []byte
	buf = append(buf, string(plan.mode)...)
	buf = append(buf, '|')
	buf = strconv.AppendBool(buf, plan.openInHostNS)
	buf = append(buf, '|')
	buf = append(buf, plan.openIface[0]...)
	buf = append(buf, '|')
	buf = strconv.AppendInt(buf, int64(plan.routeIfIndex), 10)
	buf = append(buf, '|')
	buf = strconv.AppendBool(buf, plan.trustLocal)
	buf = append(buf, '|')
	buf = append(buf, plan.reasonCode...)
	return string(buf)
}

func buildNICIPsAndFingerprint(nics []*NICInfo) ([]string, string) {
	ips := make([]string, 0, len(nics)*2)
	parts := make([]string, 0, len(nics)*4)
	for _, nic := range nics {
		if nic == nil {
			continue
		}
		parts = append(parts, nic.Name, nic.MAC, strconv.Itoa(nic.IfIndex))
		for _, addr := range nic.Addrs {
			if addr == nil {
				continue
			}
			ip := addr.IP.String()
			ips = append(ips, ip)
			parts = append(parts, ip)
		}
	}
	sort.Strings(ips)
	sort.Strings(parts)
	return ips, strings.Join(parts, "|")
}

func cloneNICInfos(src []*NICInfo) []*NICInfo {
	dst := make([]*NICInfo, 0, len(src))
	for _, nic := range src {
		if nic == nil {
			continue
		}
		cp := *nic
		if len(nic.Addrs) > 0 {
			cp.Addrs = make([]*net.IPNet, 0, len(nic.Addrs))
			for _, addr := range nic.Addrs {
				if addr == nil {
					cp.Addrs = append(cp.Addrs, nil)
					continue
				}
				addrCopy := *addr
				if addr.IP != nil {
					addrCopy.IP = append(net.IP(nil), addr.IP...)
				}
				if addr.Mask != nil {
					addrCopy.Mask = append(net.IPMask(nil), addr.Mask...)
				}
				cp.Addrs = append(cp.Addrs, &addrCopy)
			}
		}
		dst = append(dst, &cp)
	}
	return dst
}

func cloneNICInfoMap(src map[int]*NICInfo) map[int]*NICInfo {
	if len(src) == 0 {
		return map[int]*NICInfo{}
	}
	dst := make(map[int]*NICInfo, len(src))
	for k, nic := range src {
		if nic == nil {
			dst[k] = nil
			continue
		}
		cp := cloneNICInfos([]*NICInfo{nic})
		if len(cp) == 1 {
			dst[k] = cp[0]
		}
	}
	return dst
}

func cloneNICInfoSliceMap(src map[int][]*NICInfo) map[int][]*NICInfo {
	if len(src) == 0 {
		return map[int][]*NICInfo{}
	}
	dst := make(map[int][]*NICInfo, len(src))
	for k, nics := range src {
		dst[k] = cloneNICInfos(nics)
	}
	return dst
}

func isEligibleHostPeer(nic *NICInfo) bool {
	return nic != nil && nic.IsVirtual && !nic.IsLoopback
}

func findEligibleHostPeers(nics []*NICInfo, netnsID int) []*NICInfo {
	if len(nics) == 0 {
		return nil
	}
	eligible := make([]*NICInfo, 0, len(nics))
	for _, nic := range nics {
		if !isEligibleHostPeer(nic) {
			continue
		}
		if netnsID >= 0 && nic.NetNsID >= 0 && nic.NetNsID != netnsID {
			continue
		}
		eligible = append(eligible, nic)
	}
	return eligible
}

func resolveHostPeer(nsInf *netnsInformation, nic *NICInfo, hostInv hostNICInventory) (*NICInfo, string) {
	if nic == nil {
		return nil, reasonHostPeerNotFound
	}

	netnsID := lookupNetNSID(nsInf)
	if candidates := findEligibleHostPeers(hostInv.peerCandidates[nic.IfIndex], netnsID); len(candidates) > 0 {
		if len(candidates) == 1 {
			return candidates[0], reasonHostPeerSelected
		}
		return nil, reasonHostPeerAmbiguous
	}

	if peer := hostInv.peers[nic.IfLink]; peer != nil {
		switch {
		case !isEligibleHostPeer(peer):
			return nil, reasonHostPeerNotEligible
		case netnsID >= 0 && peer.NetNsID >= 0 && peer.NetNsID != netnsID:
			return nil, reasonHostPeerNotFound
		default:
			return peer, reasonHostPeerSelected
		}
	}

	if candidates := hostInv.peerCandidates[nic.IfIndex]; len(candidates) > 0 {
		for _, candidate := range candidates {
			if candidate != nil {
				return nil, reasonHostPeerNotEligible
			}
		}
	}

	return nil, reasonHostPeerNotFound
}

func resolveCapturePlan(nsInf *netnsInformation, nic *NICInfo, hostInv hostNICInventory) *capturePlan {
	plan := &capturePlan{
		mode:         captureModeInNetNS,
		openInHostNS: nsInf != nil && nsInf.hostNS,
		logicalNIC:   nic,
	}
	if nic != nil {
		plan.openIface = [2]string{nic.Name, nic.MAC}
	}
	if nsInf != nil {
		plan.openNS = nsInf.nsUID
	}
	if nsInf == nil || nic == nil || nsInf.hostNS {
		plan.reasonCode = reasonHostNamespaceOrMissingNIC
		return plan
	}
	if nic.IfIndex == 0 || nic.IfLink == 0 {
		plan.reasonCode = reasonIfLinkUnavailable
		return plan
	}

	peer, reasonCode := resolveHostPeer(nsInf, nic, hostInv)
	if peer == nil {
		plan.reasonCode = reasonCode
		return plan
	}

	plan.mode = captureModeHostPeer
	plan.openInHostNS = true
	plan.openIface = [2]string{peer.Name, peer.MAC}
	plan.routeIfIndex = peer.IfIndex
	plan.trustLocal = true
	plan.reasonCode = reasonCode
	return plan
}

func fallbackCapturePlan(nsInf *netnsInformation, nic *NICInfo, reasonCode string) *capturePlan {
	plan := &capturePlan{
		mode:         captureModeInNetNS,
		openInHostNS: nsInf != nil && nsInf.hostNS,
		logicalNIC:   nic,
		reasonCode:   reasonCode,
	}
	if nic != nil {
		plan.openIface = [2]string{nic.Name, nic.MAC}
	}
	if nsInf != nil {
		plan.openNS = nsInf.nsUID
	}
	return plan
}

func planNeedsFallbackBudget(plan *capturePlan) bool {
	return plan != nil && plan.mode != captureModeHostPeer && !plan.openInHostNS
}

func logCapturePlan(nsInf *netnsInformation, nic *NICInfo, plan *capturePlan) {
	if nsInf == nil || nic == nil || plan == nil {
		return
	}

	log.Infof(
		"capture plan ns=%s container_id=%s nic=%s ifindex=%d iflink=%d "+
			"mode=%s open_iface=%s route_ifindex=%d open_in_host_ns=%t trust_local=%t reason_code=%s",
		nsInf.nsUID,
		nsInf.contianerID,
		nic.Name,
		nic.IfIndex,
		nic.IfLink,
		plan.mode,
		plan.openIface[0],
		plan.routeIfIndex,
		plan.openInHostNS,
		plan.trustLocal,
		plan.reasonCode,
	)
}
