//go:build linux
// +build linux

package l4log

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"time"

	"github.com/google/gopacket/afpacket"
	"github.com/vishvananda/netns"
)

type cbRawSocket struct {
	tps map[[2]string]*afpacket.TPacket

	hostNS       bool
	ns           string
	newSocketErr map[[2]string]error
}

func (cb *cbRawSocket) cbNewRawSocket() {
	for nameAndMac, skt := range cb.tps {
		if skt != nil {
			continue
		}

		opt := []any{afpacket.OptInterface(nameAndMac[0])}
		if cb.hostNS {
			opt = append(opt, afpacket.OptNumBlocks(sharedCaptureSocketBlocks))
		}

		if h, err := newRawsocket(newBPFFilter(), opt...); err != nil {
			if cb.newSocketErr == nil {
				cb.newSocketErr = make(map[[2]string]error)
			}
			cb.newSocketErr[nameAndMac] = fmt.Errorf("new raw socket: %w", err)
			continue
		} else {
			log.Infof("new raw socket %s %s %s", nameAndMac[0], nameAndMac[1], cb.ns)
			time.Sleep(time.Millisecond * 100)
			cb.tps[nameAndMac] = h
		}
	}
}

type ifaceInfomation struct {
	conns *TCPConns
	cacel context.CancelFunc
	h     *afpacket.TPacket

	ifaces [2]string

	sharedHostPeer bool
	routeIfIndex   int

	lastTPacketStats tpacketStatsSnapshot
}

type namespaceCaptureRuntime struct {
	ifaces map[[2]string]*ifaceInfomation
}

func newNamespaceCaptureRuntime() *namespaceCaptureRuntime {
	return &namespaceCaptureRuntime{
		ifaces: map[[2]string]*ifaceInfomation{},
	}
}

func (rt *namespaceCaptureRuntime) stopCapture(key [2]string) {
	if rt == nil {
		return
	}
	iface, ok := rt.ifaces[key]
	if !ok || iface == nil {
		return
	}
	if iface.cacel != nil {
		iface.cacel()
	}
	if iface.conns != nil {
		iface.conns.signalStop()
	}
	delete(rt.ifaces, key)
}

func (rt *namespaceCaptureRuntime) stopAllCaptures() {
	if rt == nil {
		return
	}
	for key := range rt.ifaces {
		rt.stopCapture(key)
	}
}

func (rt *namespaceCaptureRuntime) pruneCaptures(diffPlans map[[2]string]*capturePlan) {
	if rt == nil {
		return
	}
	for key := range rt.ifaces {
		if _, ok := diffPlans[key]; ok {
			continue
		}
		rt.stopCapture(key)
	}
}

func (rt *namespaceCaptureRuntime) hasCapture(key [2]string) bool {
	if rt == nil {
		return false
	}
	_, ok := rt.ifaces[key]
	return ok
}

func (rt *namespaceCaptureRuntime) attachCapture(key [2]string, iface *ifaceInfomation) {
	if rt == nil {
		return
	}
	if rt.ifaces == nil {
		rt.ifaces = map[[2]string]*ifaceInfomation{}
	}
	rt.ifaces[key] = iface
}

func (rt *namespaceCaptureRuntime) fallbackCaptureCount() int {
	if rt == nil {
		return 0
	}
	total := 0
	for _, iface := range rt.ifaces {
		if iface == nil || iface.sharedHostPeer {
			continue
		}
		total++
	}
	return total
}

func (rt *namespaceCaptureRuntime) collectFallbackProtection(snapshot *fallbackProtectionSnapshot) {
	if rt == nil || snapshot == nil {
		return
	}
	for _, iface := range rt.ifaces {
		if iface == nil || iface.sharedHostPeer || iface.h == nil {
			continue
		}
		snapshot.active++
		if _, s3, err := iface.h.SocketStats(); err == nil {
			delta := tpacketStatsDelta(&iface.lastTPacketStats,
				uint64(s3.Packets()), uint64(s3.Drops()), uint64(s3.QueueFreezes()))
			snapshot.drops += delta.drops
			snapshot.freezes += delta.freezes
		}
	}
}

func (rt *namespaceCaptureRuntime) stopFallbackCaptures() int {
	if rt == nil {
		return 0
	}
	stopped := 0
	for key, iface := range rt.ifaces {
		if iface == nil || iface.sharedHostPeer {
			continue
		}
		rt.stopCapture(key)
		stopped++
	}
	return stopped
}

func (rt *namespaceCaptureRuntime) hostPeerRoutes() map[int]*TCPConns {
	if rt == nil {
		return nil
	}
	routes := map[int]*TCPConns{}
	for _, iface := range rt.ifaces {
		if iface == nil || !iface.sharedHostPeer || iface.routeIfIndex == 0 || iface.conns == nil {
			continue
		}
		routes[iface.routeIfIndex] = iface.conns
	}
	return routes
}

func (rt *namespaceCaptureRuntime) size() int {
	if rt == nil {
		return 0
	}
	return len(rt.ifaces)
}

func (m *netlogMonitor) captureRuntime(nsUID string) *namespaceCaptureRuntime {
	if m == nil || nsUID == "" {
		return nil
	}
	return m.captures[nsUID]
}

func (m *netlogMonitor) ensureCaptureRuntime(nsUID string) *namespaceCaptureRuntime {
	if m == nil || nsUID == "" {
		return nil
	}
	if m.captures == nil {
		m.captures = map[string]*namespaceCaptureRuntime{}
	}
	rt := m.captures[nsUID]
	if rt == nil {
		rt = newNamespaceCaptureRuntime()
		m.captures[nsUID] = rt
	}
	return rt
}

func (m *netlogMonitor) stopAllNamespaceCaptures(nsUID string) {
	rt := m.captureRuntime(nsUID)
	if rt == nil {
		return
	}
	rt.stopAllCaptures()
}

func (m *netlogMonitor) pruneNamespaceCaptures(nsUID string, diffPlans map[[2]string]*capturePlan) {
	rt := m.captureRuntime(nsUID)
	if rt == nil {
		return
	}
	rt.pruneCaptures(diffPlans)
}

func (m *netlogMonitor) namespaceHasCapture(nsUID string, key [2]string) bool {
	rt := m.captureRuntime(nsUID)
	if rt == nil {
		return false
	}
	return rt.hasCapture(key)
}

func (m *netlogMonitor) attachNamespaceCapture(nsUID string, key [2]string, iface *ifaceInfomation) {
	rt := m.ensureCaptureRuntime(nsUID)
	if rt == nil {
		return
	}
	rt.attachCapture(key, iface)
}

func (m *netlogMonitor) namespaceFallbackCaptureCount(nsUID string) int {
	rt := m.captureRuntime(nsUID)
	if rt == nil {
		return 0
	}
	return rt.fallbackCaptureCount()
}

func (m *netlogMonitor) collectNamespaceFallbackProtection(nsUID string, snapshot *fallbackProtectionSnapshot) {
	rt := m.captureRuntime(nsUID)
	if rt == nil || snapshot == nil {
		return
	}
	rt.collectFallbackProtection(snapshot)
}

func (m *netlogMonitor) stopNamespaceFallbackCaptures(nsUID string) int {
	rt := m.captureRuntime(nsUID)
	if rt == nil {
		return 0
	}
	return rt.stopFallbackCaptures()
}

func (m *netlogMonitor) namespaceHostPeerRoutes(nsUID string) map[int]*TCPConns {
	rt := m.captureRuntime(nsUID)
	if rt == nil {
		return nil
	}
	return rt.hostPeerRoutes()
}

func (m *netlogMonitor) pruneNamespaceCaptureWork(work *namespaceCaptureWork) {
	if work == nil || work.ns == nil {
		return
	}
	m.pruneNamespaceCaptures(work.ns.nsUID, work.diffPlans)
	for k := range work.diffPlans {
		if !m.namespaceHasCapture(work.ns.nsUID, k) {
			continue
		}
		delete(work.diffPlans, k)
		delete(work.containerTPs, k)
		delete(work.hostTPs, k)
	}
	if work.ns.nicIPsFingerprint != work.nicFingerprint {
		work.ns.nicIPsCache = append(work.ns.nicIPsCache[:0], work.nicIPs...)
		work.ns.nicIPsFingerprint = work.nicFingerprint
	}
}

func captureWorkPlanKeys(work *namespaceCaptureWork) [][2]string {
	if work == nil || len(work.diffPlans) == 0 {
		return nil
	}

	keys := make([][2]string, 0, len(work.diffPlans))
	for key := range work.diffPlans {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		if keys[i][0] != keys[j][0] {
			return keys[i][0] < keys[j][0]
		}
		return keys[i][1] < keys[j][1]
	})

	return keys
}

func (m *netlogMonitor) deferFallbackCapture(nsInf *netnsInformation, idx [2]string, now time.Time,
	fallbackFuseActive bool, fallbackSlotsRemaining *int,
) bool {
	if nsInf == nil {
		return false
	}

	if fallbackFuseActive {
		nsInf.bootstrapPending = true
		if time.Since(nsInf.lastBootstrapDeferredAt) >= containerBootstrapDeferLogInterval {
			log.Warnf("defer fallback packet capture ns=%s container_id=%s iface=%s reason=fuse cooldown_until=%s",
				nsInf.nsUID, nsInf.contianerID, idx[0], m.fallbackFuseUntil.Format(time.RFC3339))
			nsInf.lastBootstrapDeferredAt = now
		}
		return true
	}

	if fallbackSlotsRemaining != nil && *fallbackSlotsRemaining <= 0 {
		nsInf.bootstrapPending = true
		if time.Since(nsInf.lastBootstrapDeferredAt) >= containerBootstrapDeferLogInterval {
			log.Warnf("defer fallback packet capture ns=%s container_id=%s iface=%s reason=socket_budget limit=%d",
				nsInf.nsUID, nsInf.contianerID, idx[0], maxFallbackSocketLimit)
			nsInf.lastBootstrapDeferredAt = now
		}
		return true
	}

	if fallbackSlotsRemaining != nil {
		*fallbackSlotsRemaining--
	}

	return false
}

func (m *netlogMonitor) prepareNamespaceCaptureWork(work *namespaceCaptureWork, now time.Time,
	fallbackFuseActive bool, fallbackSlotsRemaining *int,
) {
	if work == nil || work.ns == nil {
		return
	}

	for _, idx := range captureWorkPlanKeys(work) {
		plan := work.diffPlans[idx]
		if !planNeedsFallbackBudget(plan) {
			continue
		}

		if !m.namespaceHasCapture(work.ns.nsUID, idx) &&
			m.deferFallbackCapture(work.ns, idx, now, fallbackFuseActive, fallbackSlotsRemaining) {
			delete(work.containerTPs, idx)
			delete(work.hostTPs, idx)
			work.fallbackDeferred = true
		}
	}
}

func (m *netlogMonitor) openNamespaceCaptureSockets(work *namespaceCaptureWork, hostNSInfo *netnsInformation) error {
	if work == nil || work.ns == nil {
		return nil
	}

	if len(work.containerTPs) > 0 {
		cbRawSkt := &cbRawSocket{
			hostNS: work.ns.hostNS,
			ns:     work.ns.nsUID,
			tps:    work.containerTPs,
		}
		if err := m.callContainerNetNS(work.ns, cbRawSkt.cbNewRawSocket); err != nil {
			return err
		}
		for _, v := range cbRawSkt.newSocketErr {
			log.Error(v)
		}
	}

	if len(work.hostTPs) > 0 && hostNSInfo != nil && hostNSInfo.nns != nil {
		cbRawSkt := &cbRawSocket{
			hostNS: true,
			ns:     NSInode(hostNSInfo.nns.netns),
			tps:    work.hostTPs,
		}
		if err := CallWithNetNS(hostNSInfo.nns.netns, cbRawSkt.cbNewRawSocket); err != nil {
			log.Errorf("call with host netns: %w", err)
		}
		for _, v := range cbRawSkt.newSocketErr {
			log.Error(v)
		}
	}
	return nil
}

func (m *netlogMonitor) applyNamespaceCaptureWork(work *namespaceCaptureWork) {
	if work == nil || work.ns == nil {
		return
	}

	for _, idx := range captureWorkPlanKeys(work) {
		plan := work.diffPlans[idx]
		var h *afpacket.TPacket
		if v, ok := work.containerTPs[idx]; ok {
			h = v
		}
		if v, ok := work.hostTPs[idx]; ok {
			h = v
		}

		if !m.namespaceHasCapture(work.ns.nsUID, idx) && (h != nil || (plan != nil && plan.mode == captureModeHostPeer)) {
			m.startNamespaceCapture(work.ns, idx, plan, h)
		} else if h != nil {
			h.Close()
		}
	}
	work.ns.bootstrapPending = work.fallbackDeferred
}

func (m *netlogMonitor) startNamespaceCapture(nsInf *netnsInformation, idx [2]string, plan *capturePlan,
	h *afpacket.TPacket,
) {
	ctx, cacel := context.WithCancel(context.Background())
	tags := map[string]string{}
	for k, v := range m.gtags {
		tags[k] = v
	}
	for k, v := range nsInf.tags {
		tags[k] = v
	}
	conns := NewTCPConns(tags, nsInf.contianerID, nsInf.nsUID,
		idx, m.portListen, m.transportBlacklist, m.filterRuntime)

	if plan.logicalNIC != nil {
		conns.virtualNIC = plan.logicalNIC.VIface
		conns.hostNetwork = plan.logicalNIC.HostIface
	}
	conns.trustLocal = plan.trustLocal

	m.attachNamespaceCapture(nsInf.nsUID, idx, &ifaceInfomation{
		cacel:          cacel,
		ifaces:         idx,
		conns:          conns,
		h:              h,
		sharedHostPeer: plan.mode == captureModeHostPeer,
		routeIfIndex:   plan.routeIfIndex,
	})
	time.Sleep(time.Millisecond * 100)
	if plan.mode != captureModeHostPeer {
		go conns.CapturePacket(ctx, idx[0], idx[1], nsInf.nsUID, h)
	}
	go conns.Gather(context.Background(), nsInf.nicIPsCache)

	if !nsInf.nns.portListenWatching() {
		go nsInf.nns.tcpPortListenWatcher(ctx, m.portListen, nsInf.snapshotPIDs)
	}
}

func (m *netlogMonitor) callContainerNetNS(nsInf *netnsInformation, fn func()) error {
	if nsInf == nil {
		return fmt.Errorf("missing namespace info")
	}
	if err := ensureContainerNetNS(nsInf); err != nil {
		return err
	}
	if err := CallWithNetNS(nsInf.nns.netns, fn); err != nil {
		recordTempNetNSError(nsInf, err)
		nsInf.nns.close()
		return err
	}
	clearTempNetNSError(nsInf)
	return nil
}

func CallWithNetNS(newNS netns.NsHandle, fn func()) (err error) {
	if fn == nil {
		return fmt.Errorf("missing netns callback")
	}
	if !newNS.IsOpen() {
		return fmt.Errorf("ns fd closed")
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	prevNS, err := netns.Get()
	if err != nil {
		return err
	}
	defer prevNS.Close() //nolint:errcheck

	var recovered any
	callFn := func() {
		defer func() {
			recovered = recover()
		}()
		fn()
	}
	defer func() {
		if recovered != nil {
			panic(recovered)
		}
	}()

	if !newNS.Equal(prevNS) {
		if err = netns.Set(newNS); err != nil {
			return fmt.Errorf("switch netns failed: %w", err)
		}
		defer func() {
			if restoreErr := netns.Set(prevNS); restoreErr != nil {
				if recovered != nil {
					log.Errorf("restore netns after panic failed: %s", restoreErr)
					return
				}
				if err == nil {
					err = fmt.Errorf("restore netns failed: %w", restoreErr)
					return
				}
				err = fmt.Errorf("%v; restore netns failed: %w", err, restoreErr)
			}
		}()
	}

	callFn()
	return err
}
