// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux

package traceroute

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	tcpFlagRST = 0x04
	tcpFlagSYN = 0x02
	tcpFlagACK = 0x10
)

type tcpProbeSocket struct {
	localIP    net.IP
	sourcePort uint16
	listener   *net.TCPListener
	conn       *net.IPConn
	raw        *ipv4.RawConn
	icmp       *icmp.PacketConn
	tcpBuffer  [2048]byte
	icmpBuffer [2048]byte
}

type tcpWaitResult struct {
	reply probeReply
	err   error
}

func traceTCP(ctx context.Context, target net.IP, cfg options) (Result, error) {
	socket, err := openTCPProbeSocket(ctx, target, cfg.port)
	if err != nil {
		return Result{}, err
	}
	defer socket.close()

	result := Result{Routes: make([]*Route, 0, cfg.maxTTL)}
	for ttl := 1; ttl <= cfg.maxTTL; ttl++ {
		replies := make([]probeReply, 0, cfg.attempts)
		hopReached := false
		for attempt := 0; attempt < cfg.attempts; attempt++ {
			reply, err := sendTCPProbe(ctx, socket, target, cfg.port, ttl, cfg.timeout)
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

func openTCPProbeSocket(ctx context.Context, target net.IP, targetPort uint16) (*tcpProbeSocket, error) {
	localIP, err := routeSourceIPv4(ctx, target, targetPort)
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: localIP})
	if err != nil {
		return nil, err
	}
	localAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok || localAddr.Port <= 0 || localAddr.Port > 65535 {
		_ = listener.Close()
		return nil, errors.New("get TCP traceroute source port")
	}
	conn, err := net.ListenIP("ip4:tcp", &net.IPAddr{IP: localIP})
	if err != nil {
		_ = listener.Close()
		return nil, err
	}
	raw, err := ipv4.NewRawConn(conn)
	if err != nil {
		_ = conn.Close()
		_ = listener.Close()
		return nil, err
	}
	if err := setTCPPacketFilter(conn, localIP, target, uint16(localAddr.Port), targetPort); err != nil {
		_ = conn.Close()
		_ = listener.Close()
		return nil, err
	}
	icmpConn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		_ = conn.Close()
		_ = listener.Close()
		return nil, err
	}
	return &tcpProbeSocket{
		localIP:    localIP,
		sourcePort: uint16(localAddr.Port),
		listener:   listener,
		conn:       conn,
		raw:        raw,
		icmp:       icmpConn,
	}, nil
}

func routeSourceIPv4(ctx context.Context, target net.IP, port uint16) (net.IP, error) {
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "udp4", net.JoinHostPort(target.String(), integerString(port)))
	if err != nil {
		return nil, err
	}
	defer conn.Close() //nolint:errcheck
	addr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || addr.IP.To4() == nil {
		return nil, errors.New("determine TCP traceroute source IPv4 address")
	}
	return addr.IP.To4(), nil
}

func (socket *tcpProbeSocket) close() {
	_ = socket.icmp.Close()
	_ = socket.conn.Close()
	_ = socket.listener.Close()
}

func sendTCPProbe(ctx context.Context, socket *tcpProbeSocket, target net.IP,
	targetPort uint16, ttl int, timeout time.Duration,
) (probeReply, error) {
	if err := ctx.Err(); err != nil {
		return probeReply{}, err
	}
	var sequenceBytes [4]byte
	if _, err := rand.Read(sequenceBytes[:]); err != nil {
		return probeReply{}, fmt.Errorf("generate TCP traceroute sequence: %w", err)
	}
	sequence := binary.BigEndian.Uint32(sequenceBytes[:])
	segment := marshalTCPSYN(socket.localIP, target, socket.sourcePort, targetPort, sequence)
	header := &ipv4.Header{
		Version:  ipv4.Version,
		Len:      ipv4.HeaderLen,
		TotalLen: ipv4.HeaderLen + len(segment),
		ID:       int(sequence & 0xffff),
		TTL:      ttl,
		Protocol: 6,
		Src:      socket.localIP,
		Dst:      target,
	}
	startedAt := time.Now()
	if err := socket.raw.WriteTo(header, segment, nil); err != nil {
		return probeReply{}, err
	}
	return waitTCPReply(ctx, socket, target, targetPort, sequence, startedAt, timeout)
}

func marshalTCPSYN(sourceIP, targetIP net.IP, sourcePort, targetPort uint16, sequence uint32) []byte {
	segment := make([]byte, 20)
	binary.BigEndian.PutUint16(segment[0:2], sourcePort)
	binary.BigEndian.PutUint16(segment[2:4], targetPort)
	binary.BigEndian.PutUint32(segment[4:8], sequence)
	segment[12] = 5 << 4
	segment[13] = tcpFlagSYN
	binary.BigEndian.PutUint16(segment[14:16], 64240)
	binary.BigEndian.PutUint16(segment[16:18], tcpChecksum(sourceIP, targetIP, segment))
	return segment
}

func tcpChecksum(sourceIP, targetIP net.IP, segment []byte) uint16 {
	pseudo := make([]byte, 12+len(segment))
	copy(pseudo[0:4], sourceIP.To4())
	copy(pseudo[4:8], targetIP.To4())
	pseudo[9] = 6
	binary.BigEndian.PutUint16(pseudo[10:12], uint16(len(segment)))
	copy(pseudo[12:], segment)
	var sum uint32
	for index := 0; index+1 < len(pseudo); index += 2 {
		sum += uint32(binary.BigEndian.Uint16(pseudo[index : index+2]))
	}
	if len(pseudo)%2 != 0 {
		sum += uint32(pseudo[len(pseudo)-1]) << 8
	}
	for sum>>16 != 0 {
		sum = sum&0xffff + sum>>16
	}
	return ^uint16(sum)
}

func waitTCPReply(ctx context.Context, socket *tcpProbeSocket, target net.IP,
	targetPort uint16, sequence uint32, startedAt time.Time, timeout time.Duration,
) (probeReply, error) {
	waitCtx, cancel := context.WithCancel(ctx)
	results := make(chan tcpWaitResult, 2)
	go func() {
		reply, err := waitTCPPacket(waitCtx, socket, target, targetPort, sequence, startedAt, timeout)
		results <- tcpWaitResult{reply: reply, err: err}
	}()
	go func() {
		reply, err := waitTCPICMP(waitCtx, socket, target, targetPort, sequence, startedAt, timeout)
		results <- tcpWaitResult{reply: reply, err: err}
	}()

	var firstErr error
	for pending := 2; pending > 0; pending-- {
		result := <-results
		if result.err == nil && !result.reply.timedOut {
			cancel()
			_ = socket.conn.SetReadDeadline(time.Now())
			_ = socket.icmp.SetReadDeadline(time.Now())
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

func waitTCPPacket(ctx context.Context, socket *tcpProbeSocket, target net.IP,
	targetPort uint16, sequence uint32, startedAt time.Time, timeout time.Duration,
) (probeReply, error) {
	deadline := probeDeadline(ctx, startedAt, timeout)
	buf := socket.tcpBuffer[:]
	for {
		if err := socket.conn.SetReadDeadline(probeReadDeadline(ctx, deadline)); err != nil {
			return probeReply{}, err
		}
		header, payload, _, err := socket.raw.ReadFrom(buf)
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
		if !matchesTCPResponse(header, payload, socket.localIP, target, socket.sourcePort,
			targetPort, sequence) {
			continue
		}
		return probeReply{ip: target, rtt: time.Since(startedAt), reached: true}, nil
	}
}

func matchesTCPResponse(header *ipv4.Header, segment []byte, localIP, target net.IP,
	sourcePort, targetPort uint16, sequence uint32,
) bool {
	if header == nil || !header.Src.Equal(target) || !header.Dst.Equal(localIP) || len(segment) < 20 {
		return false
	}
	if binary.BigEndian.Uint16(segment[0:2]) != targetPort ||
		binary.BigEndian.Uint16(segment[2:4]) != sourcePort {
		return false
	}
	flags := segment[13]
	if flags&(tcpFlagSYN|tcpFlagRST) == 0 || flags&tcpFlagACK == 0 {
		return false
	}
	return binary.BigEndian.Uint32(segment[8:12]) == sequence+1
}

func waitTCPICMP(ctx context.Context, socket *tcpProbeSocket, target net.IP,
	targetPort uint16, sequence uint32, startedAt time.Time, timeout time.Duration,
) (probeReply, error) {
	deadline := probeDeadline(ctx, startedAt, timeout)
	buf := socket.icmpBuffer[:]
	for {
		if err := socket.icmp.SetReadDeadline(probeReadDeadline(ctx, deadline)); err != nil {
			return probeReply{}, err
		}
		n, peer, err := socket.icmp.ReadFrom(buf)
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
		reply, matched := parseTCPICMPReply(buf[:n], peer, target, socket.sourcePort,
			targetPort, sequence, time.Since(startedAt))
		if matched {
			return reply, nil
		}
	}
}

func parseTCPICMPReply(packet []byte, peer net.Addr, target net.IP,
	sourcePort, targetPort uint16, sequence uint32, rtt time.Duration,
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
	if err != nil || header.Protocol != 6 || len(quoted) < header.Len+8 || !header.Dst.Equal(target) {
		return probeReply{}, false
	}
	tcpHeader := quoted[header.Len : header.Len+8]
	if binary.BigEndian.Uint16(tcpHeader[0:2]) != sourcePort ||
		binary.BigEndian.Uint16(tcpHeader[2:4]) != targetPort ||
		binary.BigEndian.Uint32(tcpHeader[4:8]) != sequence {
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

func integerString(value uint16) string {
	const digits = "0123456789"
	if value == 0 {
		return "0"
	}
	var buf [5]byte
	index := len(buf)
	for value > 0 {
		index--
		buf[index] = digits[value%10]
		value /= 10
	}
	return string(buf[index:])
}
