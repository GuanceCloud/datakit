// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows

package traceroute

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

var nextICMPID = uint32(time.Now().UnixNano()) //nolint:gochecknoglobals,gosec

func traceICMP(ctx context.Context, target net.IP, cfg options) (Result, error) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return Result{}, err
	}
	defer conn.Close() //nolint:errcheck

	result := Result{Routes: make([]*Route, 0, cfg.maxTTL)}
	buffer := make([]byte, 2048)
	for ttl := 1; ttl <= cfg.maxTTL; ttl++ {
		replies := make([]probeReply, 0, cfg.attempts)
		hopReached := false
		for attempt := 0; attempt < cfg.attempts; attempt++ {
			if err := ctx.Err(); err != nil {
				return result, err
			}
			token := atomic.AddUint32(&nextICMPID, 1)
			id := int(uint16(token >> 16))
			sequence := int(uint16(token))
			reply, err := sendICMPProbe(ctx, conn, buffer, target, ttl, id, sequence, cfg.timeout)
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

func sendICMPProbe(ctx context.Context, conn *icmp.PacketConn, buffer []byte, target net.IP,
	ttl, id, sequence int, timeout time.Duration,
) (probeReply, error) {
	if err := conn.IPv4PacketConn().SetTTL(ttl); err != nil {
		return probeReply{}, err
	}
	message, err := (&icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Body: &icmp.Echo{ID: id, Seq: sequence},
	}).Marshal(nil)
	if err != nil {
		return probeReply{}, err
	}
	startedAt := time.Now()
	if _, err := conn.WriteTo(message, &net.IPAddr{IP: target}); err != nil {
		return probeReply{}, err
	}
	deadline := probeDeadline(ctx, startedAt, timeout)
	for {
		if err := conn.SetReadDeadline(probeReadDeadline(ctx, deadline)); err != nil {
			return probeReply{}, err
		}
		n, peer, err := conn.ReadFrom(buffer)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return probeReply{}, ctxErr
			}
			if errors.Is(err, net.ErrClosed) {
				return probeReply{}, err
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				if time.Now().Before(deadline) {
					continue
				}
				return probeReply{timedOut: true}, nil
			}
			return probeReply{}, err
		}
		from, ok := ipFromAddr(peer)
		if !ok || !matchesICMPProbe(buffer[:n], from, target, id, sequence) {
			continue
		}
		return probeReply{
			ip:      from,
			rtt:     time.Since(startedAt),
			reached: from.Equal(target),
		}, nil
	}
}

func matchesICMPProbe(packet []byte, from, target net.IP, id, sequence int) bool {
	message, err := icmp.ParseMessage(1, packet)
	if err != nil {
		return false
	}
	if message.Type == ipv4.ICMPTypeEchoReply {
		echo, ok := message.Body.(*icmp.Echo)
		return ok && from.Equal(target) && echo.ID == id && echo.Seq == sequence
	}
	quoted := quotedDatagram(message)
	if len(quoted) < ipv4.HeaderLen || quoted[0]>>4 != ipv4.Version {
		return false
	}
	header, err := ipv4.ParseHeader(quoted)
	if err != nil || header.Protocol != 1 || len(quoted) < header.Len+8 || !header.Dst.Equal(target) {
		return false
	}
	inner, err := icmp.ParseMessage(1, quoted[header.Len:])
	if err != nil {
		return false
	}
	echo, ok := inner.Body.(*icmp.Echo)
	return ok && echo.ID == id && echo.Seq == sequence
}

func quotedDatagram(message *icmp.Message) []byte {
	switch body := message.Body.(type) {
	case *icmp.TimeExceeded:
		return body.Data
	case *icmp.DstUnreach:
		return body.Data
	case *icmp.ParamProb:
		return body.Data
	default:
		return nil
	}
}

func ipFromAddr(addr net.Addr) (net.IP, bool) {
	switch value := addr.(type) {
	case *net.IPAddr:
		if value != nil && value.IP != nil {
			return value.IP, true
		}
	case *net.UDPAddr:
		if value != nil && value.IP != nil {
			return value.IP, true
		}
	}
	return nil, false
}

func probeDeadline(ctx context.Context, startedAt time.Time, timeout time.Duration) time.Time {
	deadline := startedAt.Add(timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		return ctxDeadline
	}
	return deadline
}

func probeReadDeadline(ctx context.Context, deadline time.Time) time.Time {
	if ctx.Done() == nil {
		return deadline
	}
	pollDeadline := time.Now().Add(100 * time.Millisecond)
	if pollDeadline.Before(deadline) {
		return pollDeadline
	}
	return deadline
}
