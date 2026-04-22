//go:build linux
// +build linux

package l4log

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/GuanceCloud/platypus/pkg/ast"
	cruntime "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/runtime"
)

type netlogMonitor struct {
	netnsInfo map[string]*netnsInformation
	captures  map[string]*namespaceCaptureRuntime

	gtags map[string]string

	transportBlacklist ast.Stmts
	filterRuntime      *filterRuntime

	portListen     *portListen
	stats          *capturePlanStats
	lastPlans      map[string]string
	hostPeerOwners map[[2]string]string
	hostRuntime    *hostNamespaceRuntime

	fallbackFuseUntil        time.Time
	lastFallbackFuseReason   string
	lastFallbackFuseLoggedAt time.Time
}

type fallbackProtectionSnapshot struct {
	active  int
	pending int
	drops   uint64
	freezes uint64
}

func shouldTripFallbackFuse(snapshot fallbackProtectionSnapshot) (bool, string) {
	switch {
	case snapshot.active == 0:
		return false, ""
	case snapshot.pending > 0 && snapshot.active >= maxFallbackSocketLimit:
		return true, fmt.Sprintf("socket_saturation active=%d pending=%d limit=%d",
			snapshot.active, snapshot.pending, maxFallbackSocketLimit)
	case snapshot.drops >= fallbackFuseDropThreshold:
		return true, fmt.Sprintf("drop_burst drops=%d threshold=%d",
			snapshot.drops, fallbackFuseDropThreshold)
	case snapshot.freezes >= fallbackFuseFreezeThreshold:
		return true, fmt.Sprintf("queue_freeze freezes=%d threshold=%d",
			snapshot.freezes, fallbackFuseFreezeThreshold)
	default:
		return false, ""
	}
}

func newNetlogMonitor(gtags map[string]string, blacklist string, fnG *fnGroup,
) (*netlogMonitor, error) {
	m := &netlogMonitor{
		netnsInfo:      map[string]*netnsInformation{},
		captures:       map[string]*namespaceCaptureRuntime{},
		gtags:          gtags,
		portListen:     &portListen{},
		filterRuntime:  &filterRuntime{fnG: fnG},
		stats:          newCapturePlanStats(),
		lastPlans:      map[string]string{},
		hostPeerOwners: map[[2]string]string{},
		hostRuntime:    &hostNamespaceRuntime{},
	}

	if blacklist != "" {
		stmts, err := parseFilter(blacklist)
		if err != nil {
			return nil, fmt.Errorf("parse filter: %w", err)
		}
		err = m.filterRuntime.checkStmts(stmts, &netParams{})
		if err != nil {
			return nil, fmt.Errorf("check filter: %w", err)
		}

		m.transportBlacklist = stmts
		log.Infof("transport blacklist: \n\n%s\n", blacklist)
	}

	return m, nil
}

func (m *netlogMonitor) Run(ctx context.Context, containerCtr []cruntime.ContainerRuntime,
) {
	ticker := time.NewTicker(time.Second * 20)
	statsTicker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	defer statsTicker.Stop()
	var allowLo bool

	for {
		netnsInfo := ListContainersAndHostNetNS(containerCtr, allowLo)
		m.CmpAndCleanNetNsNIC(netnsInfo)
		m.CmpAndAddNIC(netnsInfo)
		select {
		case <-ctx.Done():
			for nsUID, infom := range m.netnsInfo {
				m.stopAllNamespaceCaptures(nsUID)
				infom.Close()
			}
			if m.hostRuntime != nil {
				m.hostRuntime.stopSharedCapture()
			}
			m.netnsInfo = make(map[string]*netnsInformation)
			m.captures = make(map[string]*namespaceCaptureRuntime)
			return
		case <-statsTicker.C:
			if snapshot := m.collectFallbackProtectionSnapshot(); true {
				if trip, reason := shouldTripFallbackFuse(snapshot); trip && !m.fallbackFuseActive(time.Now()) {
					m.tripFallbackFuse(time.Now(), reason)
				}
			}
			if m.stats != nil {
				log.Info(m.stats.SummaryAndResetDelta())
			}
			if m.hostRuntime != nil && m.hostRuntime.sharedCapture != nil {
				log.Info(m.hostRuntime.sharedCapture.summary())
			}
		case <-ticker.C:
		}
	}
}

type capturePlanStats struct {
	mu sync.Mutex

	snapshot snapshotPlanCounters
	delta    snapshotPlanCounters
}

type snapshotPlanCounters struct {
	hostPeerSelected int64
	fallbackInNetNS  int64
	fallbackConflict int64
	hostNSDirect     int64
	reasons          map[string]int64
}

func newSnapshotPlanCounters() snapshotPlanCounters {
	return snapshotPlanCounters{
		reasons: map[string]int64{},
	}
}

func newCapturePlanStats() *capturePlanStats {
	return &capturePlanStats{
		snapshot: newSnapshotPlanCounters(),
		delta:    newSnapshotPlanCounters(),
	}
}

func (s *capturePlanStats) ReplaceSnapshot(counters snapshotPlanCounters) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.snapshot = counters
	s.mu.Unlock()
}

func (s *capturePlanStats) IncDelta(plan *capturePlan) {
	if s == nil || plan == nil {
		return
	}
	s.mu.Lock()
	applyPlanCounter(&s.delta, plan)
	s.mu.Unlock()
}

func (s *capturePlanStats) SummaryAndResetDelta() string {
	if s == nil {
		return "capture plan stats unavailable"
	}
	s.mu.Lock()
	snapshot := cloneSnapshotPlanCounters(s.snapshot)
	delta := cloneSnapshotPlanCounters(s.delta)
	s.delta = newSnapshotPlanCounters()
	defer s.mu.Unlock()

	return fmt.Sprintf(
		"capture plan stats snapshot_host_peer_selected=%d snapshot_fallback_in_netns=%d "+
			"snapshot_fallback_conflict=%d snapshot_host_namespace_direct=%d "+
			"delta_host_peer_selected=%d delta_fallback_in_netns=%d "+
			"delta_fallback_conflict=%d delta_host_namespace_direct=%d "+
			"snapshot_reasons=[%s] delta_reasons=[%s]",
		snapshot.hostPeerSelected,
		snapshot.fallbackInNetNS,
		snapshot.fallbackConflict,
		snapshot.hostNSDirect,
		delta.hostPeerSelected,
		delta.fallbackInNetNS,
		delta.fallbackConflict,
		delta.hostNSDirect,
		formatReasonCounts(snapshot.reasons),
		formatReasonCounts(delta.reasons),
	)
}

func applyPlanCounter(counters *snapshotPlanCounters, plan *capturePlan) {
	if counters == nil || plan == nil {
		return
	}

	switch plan.reasonCode {
	case reasonHostPeerSelected:
		counters.hostPeerSelected++
	case reasonHostPeerConflict:
		counters.fallbackConflict++
	case reasonHostNamespaceOrMissingNIC:
		counters.hostNSDirect++
	default:
		counters.fallbackInNetNS++
	}

	if counters.reasons == nil {
		counters.reasons = map[string]int64{}
	}
	if plan.reasonCode != "" {
		counters.reasons[plan.reasonCode]++
	}
}

func cloneSnapshotPlanCounters(src snapshotPlanCounters) snapshotPlanCounters {
	dst := snapshotPlanCounters{
		hostPeerSelected: src.hostPeerSelected,
		fallbackInNetNS:  src.fallbackInNetNS,
		fallbackConflict: src.fallbackConflict,
		hostNSDirect:     src.hostNSDirect,
		reasons:          map[string]int64{},
	}
	for k, v := range src.reasons {
		dst.reasons[k] = v
	}
	return dst
}

func formatReasonCounts(reasons map[string]int64) string {
	reasonKeys := make([]string, 0, len(reasons))
	for k := range reasons {
		reasonKeys = append(reasonKeys, k)
	}
	sort.Strings(reasonKeys)

	reasonParts := make([]string, 0, len(reasonKeys))
	for _, k := range reasonKeys {
		reasonParts = append(reasonParts, fmt.Sprintf("%s=%d", k, reasons[k]))
	}

	return strings.Join(reasonParts, ",")
}

func (m *netlogMonitor) CmpAndAddNIC(netnsInfo map[string]*netnsSnapshot) {
	claimedHostPeerOwners := map[[2]string]string{}
	currentSnapshot := newSnapshotPlanCounters()
	nicGroups := newNICGroupSnapshot()
	bootstrapBudget := containerBootstrapBudgetPerRound
	now := time.Now()
	fallbackFuseActive := m.fallbackFuseActive(now)
	fallbackSlotsRemaining := maxFallbackSocketLimit - m.activeFallbackCaptureCount()
	if fallbackSlotsRemaining < 0 {
		fallbackSlotsRemaining = 0
	}

	nsUIDs := make([]string, 0, len(netnsInfo))
	for nsUID := range netnsInfo {
		nsUIDs = append(nsUIDs, nsUID)
	}
	sort.Strings(nsUIDs)

	for _, nsUID := range nsUIDs {
		m.mergeNamespaceState(nsUID, netnsInfo[nsUID])
	}

	hostInv := m.buildHostNICInventory()
	hostNSInfo := hostInv.nsInfo

	for _, nsUID := range nsUIDs {
		preNsInf := m.netnsInfo[nsUID]

		if shouldDeferContainerBootstrap(preNsInf, &bootstrapBudget) {
			if time.Since(preNsInf.lastBootstrapDeferredAt) >= containerBootstrapDeferLogInterval {
				log.Infof("defer container netns bootstrap ns=%s container_id=%s remaining_budget=0", preNsInf.nsUID, preNsInf.contianerID)
				preNsInf.lastBootstrapDeferredAt = time.Now()
			}
			continue
		}

		work, err := m.buildNamespaceCaptureWork(preNsInf, hostInv, claimedHostPeerOwners, &currentSnapshot, nicGroups)
		if err != nil {
			log.Errorf("get network interface info: %w, ns: %s", err, nsUID)
			continue
		}

		m.pruneNamespaceCaptureWork(work)
		m.prepareNamespaceCaptureWork(work, now, fallbackFuseActive, &fallbackSlotsRemaining)

		if err := m.openNamespaceCaptureSockets(work, hostNSInfo); err != nil {
			log.Errorf("call with netns: %w", err)
			preNsInf.Close()
			continue
		}

		m.applyNamespaceCaptureWork(work)
	}
	m.syncHostPeerSharedCapture(hostNSInfo)
	m.hostPeerOwners = claimedHostPeerOwners
	if m.stats != nil {
		m.stats.ReplaceSnapshot(currentSnapshot)
	}
	nicGroups.export()
}
