//go:build linux
// +build linux

package l4log

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/afpacket"
)

type hostNICInventory struct {
	nsInfo         *netnsInformation
	peers          map[int]*NICInfo
	peerCandidates map[int][]*NICInfo
	captureNICs    []*NICInfo
}

type hostNamespaceRuntime struct {
	ns *netnsInformation

	inventory   hostNICInventory
	inventoryAt time.Time

	sharedCapture *hostPeerSharedCapture
}

type hostPeerSharedCapture struct {
	ctx    context.Context
	cancel context.CancelFunc
	h      *afpacket.TPacket
	ns     string
	mode   string

	mu     sync.RWMutex
	routes map[int]*TCPConns

	filterFingerprint string
	attachStats       sharedFilterAttachStats
	filter            *sharedHostPeerSocketFilter

	lastTPacketStats tpacketStatsSnapshot
}

type sharedFilterAttachStats struct {
	ebpfSuccess int64
	cbpfSuccess int64
	ebpfFailure int64
	syncCount   int64
}

func (r *hostNamespaceRuntime) setNamespace(ns *netnsInformation) {
	if r == nil {
		return
	}
	r.ns = ns
}

func (r *hostNamespaceRuntime) clearNamespace(nsUID string) {
	if r == nil || r.ns == nil || r.ns.nsUID != nsUID {
		return
	}
	r.ns = nil
	r.inventory = hostNICInventory{
		peers:          map[int]*NICInfo{},
		peerCandidates: map[int][]*NICInfo{},
	}
	r.inventoryAt = time.Time{}
	r.stopSharedCapture()
}

func (r *hostNamespaceRuntime) cacheInventory(inv hostNICInventory) {
	if r == nil {
		return
	}
	r.inventory = hostNICInventory{
		nsInfo:         r.ns,
		peers:          cloneNICInfoMap(inv.peers),
		peerCandidates: cloneNICInfoSliceMap(inv.peerCandidates),
		captureNICs:    cloneNICInfos(inv.captureNICs),
	}
	r.inventoryAt = time.Now()
}

func (r *hostNamespaceRuntime) cachedInventory(ttl time.Duration) (hostNICInventory, bool) {
	if ttl <= 0 {
		ttl = hostNICCacheTTL
	}
	if r == nil || r.ns == nil || time.Since(r.inventoryAt) >= ttl {
		return hostNICInventory{}, false
	}
	return hostNICInventory{
		nsInfo:         r.ns,
		peers:          cloneNICInfoMap(r.inventory.peers),
		peerCandidates: cloneNICInfoSliceMap(r.inventory.peerCandidates),
		captureNICs:    cloneNICInfos(r.inventory.captureNICs),
	}, true
}

func (m *netlogMonitor) hostNICInventoryTTL() time.Duration {
	if m == nil {
		return hostNICCacheTTL
	}
	if m.hostRuntime != nil && m.hostRuntime.sharedCapture != nil {
		return unstableHostNICCacheTTL
	}
	for _, nsInf := range m.netnsInfo {
		if nsInf != nil && !nsInf.hostNS && nsInf.bootstrapPending {
			return unstableHostNICCacheTTL
		}
	}
	return hostNICCacheTTL
}

func (r *hostNamespaceRuntime) stopSharedCapture() {
	if r == nil || r.sharedCapture == nil {
		return
	}
	log.Infof("stop host peer shared socket ns=%s", r.sharedCapture.ns)
	r.sharedCapture.stop()
	r.sharedCapture = nil
}

func (m *netlogMonitor) buildHostNICInventory() hostNICInventory {
	result := hostNICInventory{
		peers:          map[int]*NICInfo{},
		peerCandidates: map[int][]*NICInfo{},
	}
	if m.hostRuntime == nil || m.hostRuntime.ns == nil || m.hostRuntime.ns.nns == nil {
		return result
	}

	result.nsInfo = m.hostRuntime.ns
	if cached, ok := m.hostRuntime.cachedInventory(m.hostNICInventoryTTL()); ok {
		return cached
	}

	nics, err := m.hostRuntime.ns.nns.nicInfoWithVirtual(true)
	if err != nil {
		log.Errorf("get host network interface info: %w, ns: %s", err, m.hostRuntime.ns.nsUID)
		return result
	}

	for _, nic := range nics {
		if nic == nil || nic.IfIndex == 0 {
			continue
		}
		result.peers[nic.IfIndex] = nic
		if nic.IfLink > 0 {
			result.peerCandidates[nic.IfLink] = append(result.peerCandidates[nic.IfLink], nic)
		}
		if !nic.IsVirtual || (nic.IsLoopback && m.hostRuntime.ns.nns.includeLo) {
			result.captureNICs = append(result.captureNICs, nic)
		}
	}

	for ifindex := range result.peerCandidates {
		sort.Slice(result.peerCandidates[ifindex], func(i, j int) bool {
			left := result.peerCandidates[ifindex][i]
			right := result.peerCandidates[ifindex][j]
			if left == nil || right == nil {
				return left != nil
			}
			if left.NetNsID != right.NetNsID {
				return left.NetNsID < right.NetNsID
			}
			return left.IfIndex < right.IfIndex
		})
	}

	if m.hostRuntime != nil {
		m.hostRuntime.cacheInventory(result)
	}
	return result
}

func (m *netlogMonitor) buildHostPeerSharedRoutes() map[int]*TCPConns {
	routes := map[int]*TCPConns{}
	for nsUID := range m.netnsInfo {
		for ifindex, conns := range m.namespaceHostPeerRoutes(nsUID) {
			routes[ifindex] = conns
		}
	}
	return routes
}

func hostPeerRouteIfIndexes(routes map[int]*TCPConns) []int {
	if len(routes) == 0 {
		return nil
	}
	ifindexes := make([]int, 0, len(routes))
	for ifindex := range routes {
		ifindexes = append(ifindexes, ifindex)
	}
	sort.Ints(ifindexes)
	return ifindexes
}

func newHostPeerSharedSocket(hostNSInfo *netnsInformation) (*afpacket.TPacket, error) {
	if hostNSInfo == nil || hostNSInfo.nns == nil {
		return nil, fmt.Errorf("missing host netns for shared capture")
	}

	var (
		h   *afpacket.TPacket
		err error
	)
	if callErr := CallWithNetNS(hostNSInfo.nns.netns, func() {
		h, err = newRawsocket(newBPFFilter(), afpacket.OptNumBlocks(sharedCaptureSocketBlocks))
	}); callErr != nil {
		return nil, callErr
	}
	if err != nil {
		return nil, fmt.Errorf("new host peer shared socket: %w", err)
	}
	return h, nil
}

func (m *netlogMonitor) syncHostPeerSharedCapture(hostNSInfo *netnsInformation) {
	routes := m.buildHostPeerSharedRoutes()
	ifindexes := hostPeerRouteIfIndexes(routes)
	if m.hostRuntime == nil {
		return
	}
	if len(routes) == 0 || hostNSInfo == nil || hostNSInfo.nns == nil {
		m.hostRuntime.stopSharedCapture()
		return
	}

	if m.hostRuntime.sharedCapture == nil {
		h, err := newHostPeerSharedSocket(hostNSInfo)
		if err != nil {
			log.Error(err)
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		m.hostRuntime.sharedCapture = &hostPeerSharedCapture{
			ctx:    ctx,
			cancel: cancel,
			h:      h,
			ns:     hostNSInfo.nsUID,
			routes: map[int]*TCPConns{},
			filter: &sharedHostPeerSocketFilter{},
		}
		log.Infof("start host peer shared socket ns=%s routes=%d", hostNSInfo.nsUID, len(routes))
		go m.hostRuntime.sharedCapture.run()
	}

	if err := m.hostRuntime.sharedCapture.syncFilter(ifindexes); err != nil {
		log.Errorf("sync host peer shared filter: %s", err)
		return
	}
	m.hostRuntime.sharedCapture.setRoutes(routes)
}

func (c *hostPeerSharedCapture) stop() {
	if c == nil {
		return
	}
	if c.filter != nil {
		c.filter.close()
	}
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *hostPeerSharedCapture) setRoutes(routes map[int]*TCPConns) {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.routes = routes
	c.mu.Unlock()
}

func (c *hostPeerSharedCapture) summary() string {
	if c == nil {
		return "host peer shared socket disabled"
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	return fmt.Sprintf("host peer shared socket ns=%s mode=%s routes=%d ifindexes=%s sync_count=%d ebpf_success=%d cbpf_success=%d ebpf_failure=%d",
		c.ns, c.mode, len(c.routes), c.filterFingerprint,
		c.attachStats.syncCount, c.attachStats.ebpfSuccess, c.attachStats.cbpfSuccess, c.attachStats.ebpfFailure)
}

func (c *hostPeerSharedCapture) syncFilter(ifindexes []int) error {
	if c == nil || c.h == nil {
		return fmt.Errorf("shared host-peer socket unavailable")
	}

	fp := sharedHostPeerFilterFingerprint(ifindexes)
	if fp == "" {
		return fmt.Errorf("empty shared host-peer filter fingerprint")
	}
	if fp == c.filterFingerprint {
		return nil
	}

	if c.filter == nil {
		c.filter = &sharedHostPeerSocketFilter{}
	}
	mode, ebpfFailed, err := c.filter.sync(c.h, ifindexes)
	if err != nil {
		c.mu.Lock()
		c.attachStats.syncCount++
		if ebpfFailed {
			c.attachStats.ebpfFailure++
		}
		c.mu.Unlock()
		return err
	}

	c.mu.Lock()
	c.filterFingerprint = fp
	c.mode = mode
	c.attachStats.syncCount++
	switch mode {
	case "ebpf":
		c.attachStats.ebpfSuccess++
	case "cbpf":
		c.attachStats.cbpfSuccess++
		if ebpfFailed {
			c.attachStats.ebpfFailure++
		}
	}
	c.mu.Unlock()
	log.Infof("sync host peer shared filter ns=%s mode=%s ifindexes=%s", c.ns, mode, fp)
	return nil
}

func (c *hostPeerSharedCapture) routeFor(ifindex int) *TCPConns {
	if c == nil || ifindex == 0 {
		return nil
	}
	c.mu.RLock()
	conns := c.routes[ifindex]
	c.mu.RUnlock()
	return conns
}

func (c *hostPeerSharedCapture) run() {
	if c == nil || c.h == nil {
		return
	}

	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	layerLi := make([]gopacket.LayerType, 0, 10)
	decoder := NewPktDecoder()
	stringCache := newPacketStringCache()

	for {
		select {
		case <-ticker.C:
			if _, s3, err := c.h.SocketStats(); err != nil {
				log.Error(err)
			} else {
				observeTPacketStatsDelta("l4log_host_peer_shared", &c.lastTPacketStats,
					uint64(s3.Packets()), uint64(s3.Drops()), uint64(s3.QueueFreezes()))
				log.Infof("%s drops=%d packets=%d freezes=%d",
					c.summary(), s3.Drops(), s3.Packets(), s3.QueueFreezes())
			}
		case <-c.ctx.Done():
			c.h.Close()
			return
		default:
		}

		buf, ci, err := c.h.ZeroCopyReadPacketData()
		if err != nil {
			if isPacketReadTimeout(err) {
				continue
			}
			log.Error(err)
			time.Sleep(time.Millisecond * 300)
			continue
		}

		conns := c.routeFor(ci.InterfaceIndex)
		if conns == nil {
			continue
		}
		conns.handleCapturedPacket(decoder, layerLi, stringCache, buf, ci)
	}
}
