// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux

package traceroute

import (
	"context"
	"encoding/binary"
	"errors"
	"math"
	"net"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type udpLeaseKey struct {
	target     [net.IPv4len]byte
	targetPort uint16
	sourcePort uint16
}

type udpLeaseRegistry struct {
	mu           sync.Mutex
	leases       map[udpLeaseKey]time.Time
	cleanupTimer *time.Timer
}

type udpProbeSocket struct {
	conn       *net.UDPConn
	packetConn *ipv4.PacketConn
	key        udpLeaseKey
	sent       bool
	appBuffer  [2048]byte
	icmpBuffer [2048]byte
}

type udpWaitResult struct {
	reply probeReply
	err   error
}

// UDPProber reuses one raw ICMP receiver across sequential UDP probes. Probe
// calls are serialized so replies cannot be consumed by the wrong caller.
type UDPProber struct {
	mu   sync.Mutex
	icmp *icmp.PacketConn
}

var udpLeases = udpLeaseRegistry{leases: make(map[udpLeaseKey]time.Time)} //nolint:gochecknoglobals

func traceUDP(ctx context.Context, target net.IP, cfg options) (Result, error) {
	icmpConn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return Result{}, err
	}
	defer icmpConn.Close() //nolint:errcheck

	result := Result{Routes: make([]*Route, 0, cfg.maxTTL)}
	for ttl := 1; ttl <= cfg.maxTTL; ttl++ {
		replies := make([]probeReply, 0, cfg.attempts)
		hopReached := false
		for attempt := 0; attempt < cfg.attempts; attempt++ {
			probeID := (ttl-1)*cfg.attempts + attempt + 1
			reply, err := sendIsolatedUDPProbe(ctx, icmpConn, target, cfg.port, ttl, probeID, cfg.timeout)
			if err != nil {
				if len(replies) > 0 {
					result.Routes = append(result.Routes, repliesToRoute(replies))
				}
				return result, err
			}
			replies = append(replies, reply)
			hopReached = hopReached || reply.reached
		}
		result.Routes = append(result.Routes, repliesToRoute(replies))
		if hopReached {
			result.Reached = true
			return result, nil
		}
	}
	return result, nil
}

// ProbeUDP sends one IPv4 UDP probe. A TTL of zero keeps the operating
// system's default TTL, which is useful for endpoint reachability checks.
func ProbeUDP(ctx context.Context, target net.IP, port uint16, ttl int,
	timeout time.Duration,
) (ProbeResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return ProbeResult{}, err
	}
	target, timeout, err := validateUDPProbe(target, port, ttl, timeout)
	if err != nil {
		return ProbeResult{}, err
	}
	prober, err := NewUDPProber()
	if err != nil {
		return ProbeResult{}, err
	}
	defer prober.Close() //nolint:errcheck
	return prober.probe(ctx, target, port, ttl, timeout)
}

// NewUDPProber opens a reusable raw ICMP receiver for UDP probes.
func NewUDPProber() (*UDPProber, error) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, err
	}
	return &UDPProber{icmp: conn}, nil
}

// Close closes the reusable raw ICMP receiver.
func (prober *UDPProber) Close() error {
	if prober == nil {
		return nil
	}
	prober.mu.Lock()
	defer prober.mu.Unlock()
	if prober.icmp == nil {
		return nil
	}
	conn := prober.icmp
	prober.icmp = nil
	return conn.Close()
}

// Probe sends one IPv4 UDP probe. A TTL of zero keeps the operating system's
// default TTL, which is useful for endpoint reachability checks.
func (prober *UDPProber) Probe(ctx context.Context, target net.IP, port uint16, ttl int,
	timeout time.Duration,
) (ProbeResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return ProbeResult{}, err
	}
	target, timeout, err := validateUDPProbe(target, port, ttl, timeout)
	if err != nil {
		return ProbeResult{}, err
	}
	return prober.probe(ctx, target, port, ttl, timeout)
}

func (prober *UDPProber) probe(ctx context.Context, target net.IP, port uint16, ttl int,
	timeout time.Duration,
) (ProbeResult, error) {
	if prober == nil {
		return ProbeResult{}, errors.New("UDP prober is closed")
	}
	prober.mu.Lock()
	defer prober.mu.Unlock()
	if prober.icmp == nil {
		return ProbeResult{}, errors.New("UDP prober is closed")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	socket, err := openUDPProbeSocket(ctx, target, port)
	if err != nil {
		return ProbeResult{}, err
	}
	defer socket.close()
	reply, err := sendUDPProbe(ctx, prober.icmp, socket, target, port, ttl, 1, timeout)
	result := ProbeResult{
		RTT:      reply.rtt,
		Reached:  reply.reached,
		TimedOut: reply.timedOut,
		Sent:     reply.sent,
	}
	if reply.ip != nil {
		result.IP = reply.ip.String()
	}
	return result, err
}

func validateUDPProbe(target net.IP, port uint16, ttl int,
	timeout time.Duration,
) (net.IP, time.Duration, error) {
	target = target.To4()
	if target == nil {
		return nil, 0, errors.New("udp probe target must be IPv4")
	}
	if target.IsUnspecified() || target.IsMulticast() || target.Equal(net.IPv4bcast) {
		return nil, 0, errors.New("udp probe target must be a unicast address")
	}
	if port == 0 {
		return nil, 0, errors.New("udp probe requires a destination port")
	}
	if ttl < 0 || ttl > MaxUDPHops {
		return nil, 0, errors.New("udp probe TTL is out of range")
	}
	if timeout <= 0 {
		timeout = defaultTimeout
	} else if timeout > MaxTimeout {
		timeout = MaxTimeout
	}
	return target, timeout, nil
}

func openUDPProbeSocket(ctx context.Context, target net.IP, targetPort uint16) (*udpProbeSocket, error) {
	var targetBytes [net.IPv4len]byte
	copy(targetBytes[:], target.To4())
	for attempt := 0; attempt < 128; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero})
		if err != nil {
			return nil, err
		}
		addr, ok := conn.LocalAddr().(*net.UDPAddr)
		if !ok || addr.Port <= 0 || addr.Port > math.MaxUint16 {
			_ = conn.Close()
			return nil, errors.New("get UDP traceroute source port")
		}
		key := udpLeaseKey{target: targetBytes, targetPort: targetPort, sourcePort: uint16(addr.Port)}
		if !reserveUDPLease(key, time.Now()) {
			_ = conn.Close()
			continue
		}
		return &udpProbeSocket{conn: conn, packetConn: ipv4.NewPacketConn(conn), key: key}, nil
	}
	return nil, errors.New("allocate isolated UDP traceroute source port")
}

func reserveUDPLease(key udpLeaseKey, now time.Time) bool {
	return udpLeases.reserve(key, now)
}

func (registry *udpLeaseRegistry) reserve(key udpLeaseKey, now time.Time) bool {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if expires, ok := registry.leases[key]; ok && (expires.IsZero() || now.Before(expires)) {
		return false
	}
	registry.leases[key] = time.Time{}
	return true
}

func (socket *udpProbeSocket) close() {
	udpLeases.finish(socket.key, socket.sent, time.Now())
	_ = socket.conn.Close()
}

func (registry *udpLeaseRegistry) finish(key udpLeaseKey, sent bool, now time.Time) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if sent {
		registry.leases[key] = now.Add(MaxTimeout)
		registry.scheduleCleanupLocked()
	} else {
		delete(registry.leases, key)
	}
}

func (registry *udpLeaseRegistry) scheduleCleanupLocked() {
	if registry.cleanupTimer != nil {
		return
	}
	registry.cleanupTimer = time.AfterFunc(MaxTimeout, func() {
		registry.cleanupExpiredAt(time.Now())
	})
}

func (registry *udpLeaseRegistry) cleanupExpiredAt(now time.Time) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.cleanupTimer = nil
	hasQuarantine := false
	for key, expires := range registry.leases {
		if expires.IsZero() {
			continue
		}
		if !now.Before(expires) {
			delete(registry.leases, key)
			continue
		}
		hasQuarantine = true
	}
	if hasQuarantine {
		registry.scheduleCleanupLocked()
	}
}

func sendIsolatedUDPProbe(ctx context.Context, icmpConn *icmp.PacketConn, target net.IP,
	targetPort uint16, ttl, probeID int, timeout time.Duration,
) (probeReply, error) {
	socket, err := openUDPProbeSocket(ctx, target, targetPort)
	if err != nil {
		return probeReply{}, err
	}
	defer socket.close()
	return sendUDPProbe(ctx, icmpConn, socket, target, targetPort, ttl, probeID, timeout)
}

func sendUDPProbe(ctx context.Context, icmpConn *icmp.PacketConn, socket *udpProbeSocket,
	target net.IP, targetPort uint16, ttl, probeID int, timeout time.Duration,
) (probeReply, error) {
	if err := ctx.Err(); err != nil {
		return probeReply{}, err
	}
	if ttl > 0 {
		if err := socket.packetConn.SetTTL(ttl); err != nil {
			return probeReply{}, err
		}
	}
	payload, udpLength, err := udpPayload(probeID)
	if err != nil {
		return probeReply{}, err
	}
	startedAt := time.Now()
	if _, err := socket.conn.WriteToUDP(payload, &net.UDPAddr{IP: target, Port: int(targetPort)}); err != nil {
		return probeReply{}, err
	}
	socket.sent = true
	reply, err := waitUDPReply(ctx, icmpConn, socket.conn, target, socket.key.sourcePort,
		targetPort, udpLength, startedAt, timeout, socket)
	reply.sent = true
	return reply, err
}

func udpPayload(probeID int) ([]byte, uint16, error) {
	const udpHeaderLength = 8
	if probeID <= 0 || probeID > MaxUDPProbePayload {
		return nil, 0, errors.New("invalid UDP traceroute probe ID")
	}
	payload := make([]byte, probeID)
	return payload, uint16(udpHeaderLength + len(payload)), nil
}

func waitUDPReply(ctx context.Context, icmpConn *icmp.PacketConn, udpConn *net.UDPConn,
	target net.IP, sourcePort, targetPort, udpLength uint16, startedAt time.Time,
	timeout time.Duration, socket *udpProbeSocket,
) (probeReply, error) {
	waitCtx, cancel := context.WithCancel(ctx)
	results := make(chan udpWaitResult, 2)
	go func() {
		reply, err := waitUDPICMPReplyBuffer(waitCtx, icmpConn, socket.icmpBuffer[:],
			target, sourcePort, targetPort, udpLength, startedAt, timeout)
		results <- udpWaitResult{reply: reply, err: err}
	}()
	go func() {
		reply, err := waitUDPApplicationReplyBuffer(waitCtx, udpConn, socket.appBuffer[:],
			target, targetPort, startedAt, timeout)
		results <- udpWaitResult{reply: reply, err: err}
	}()

	var firstErr error
	for pending := 2; pending > 0; pending-- {
		result := <-results
		if result.err == nil && !result.reply.timedOut {
			cancel()
			_ = icmpConn.SetReadDeadline(time.Now())
			_ = udpConn.SetReadDeadline(time.Now())
			for pending--; pending > 0; pending-- {
				<-results
			}
			return result.reply, nil
		}
		if result.err != nil && firstErr == nil {
			firstErr = result.err
		}
	}
	cancel()
	if err := ctx.Err(); err != nil {
		return probeReply{}, err
	}
	if firstErr != nil {
		return probeReply{}, firstErr
	}
	return probeReply{timedOut: true}, nil
}

func waitUDPApplicationReply(ctx context.Context, conn *net.UDPConn, target net.IP,
	targetPort uint16, startedAt time.Time, timeout time.Duration,
) (probeReply, error) {
	return waitUDPApplicationReplyBuffer(ctx, conn, make([]byte, 2048), target,
		targetPort, startedAt, timeout)
}

func waitUDPApplicationReplyBuffer(ctx context.Context, conn *net.UDPConn, buffer []byte,
	target net.IP, targetPort uint16, startedAt time.Time, timeout time.Duration,
) (probeReply, error) {
	deadline := probeDeadline(ctx, startedAt, timeout)
	for {
		if err := conn.SetReadDeadline(probeReadDeadline(ctx, deadline)); err != nil {
			return probeReply{}, err
		}
		_, peer, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if ctx.Err() != nil {
				return probeReply{}, ctx.Err()
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				if time.Now().Before(deadline) {
					continue
				}
				return probeReply{timedOut: true}, nil
			}
			return probeReply{}, err
		}
		if peer == nil || peer.Port != int(targetPort) || !peer.IP.Equal(target) {
			continue
		}
		return probeReply{ip: peer.IP, rtt: time.Since(startedAt), reached: true}, nil
	}
}

func waitUDPICMPReplyBuffer(ctx context.Context, conn *icmp.PacketConn, buffer []byte,
	target net.IP, sourcePort, targetPort, udpLength uint16, startedAt time.Time,
	timeout time.Duration,
) (probeReply, error) {
	deadline := probeDeadline(ctx, startedAt, timeout)
	for {
		if err := conn.SetReadDeadline(probeReadDeadline(ctx, deadline)); err != nil {
			return probeReply{}, err
		}
		n, peer, err := conn.ReadFrom(buffer)
		if err != nil {
			if ctx.Err() != nil {
				return probeReply{}, ctx.Err()
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				if time.Now().Before(deadline) {
					continue
				}
				return probeReply{timedOut: true}, nil
			}
			return probeReply{}, err
		}
		reply, matched := parseUDPICMPReply(buffer[:n], peer, target, sourcePort, targetPort,
			udpLength, time.Since(startedAt))
		if matched {
			return reply, nil
		}
	}
}

func parseUDPICMPReply(packet []byte, peer net.Addr, target net.IP,
	sourcePort, targetPort, udpLength uint16, rtt time.Duration,
) (probeReply, bool) {
	message, err := icmp.ParseMessage(1, packet)
	if err != nil {
		return probeReply{}, false
	}
	quoted := quotedDatagram(message)
	if quoted == nil {
		return probeReply{}, false
	}
	header, err := ipv4.ParseHeader(quoted)
	if err != nil || header.Protocol != 17 || len(quoted) < header.Len+8 || !header.Dst.Equal(target) {
		return probeReply{}, false
	}
	udpHeader := quoted[header.Len : header.Len+8]
	if binary.BigEndian.Uint16(udpHeader[0:2]) != sourcePort ||
		binary.BigEndian.Uint16(udpHeader[2:4]) != targetPort ||
		binary.BigEndian.Uint16(udpHeader[4:6]) != udpLength {
		return probeReply{}, false
	}
	from, ok := ipFromAddr(peer)
	if !ok {
		return probeReply{}, false
	}
	_, destinationUnreachable := message.Body.(*icmp.DstUnreach)
	return probeReply{
		ip:      from,
		rtt:     rtt,
		reached: destinationUnreachable && from.Equal(target),
	}, true
}
