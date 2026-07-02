// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package net

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	dnscache "go.mercari.io/go-dnscache"
	"go.uber.org/zap"
)

const (
	IPFamilyPolicyAuto       = "auto"
	IPFamilyPolicyPreferIPv6 = "prefer_ipv6"
	IPFamilyPolicyIPv6Only   = "ipv6_only"
	IPFamilyPolicyIPv4Only   = "ipv4_only"

	DefaultIPv4FallbackDelay = 250 * time.Millisecond
)

type DialFunc func(ctx context.Context, network, addr string) (net.Conn, error)

type DNSCacheDialOptions struct {
	Freq             time.Duration
	LookupTimeout    time.Duration
	FallbackDelay    time.Duration
	IPFamilyPolicy   string
	DialTimeout      time.Duration
	KeepAlive        time.Duration
	DialDone         func(family, result string, elapsed time.Duration)
	FallbackStarted  func()
	ConnectionOpened func(family, remote string)
	ConnectionClosed func(family string)
}

func ValidateIPFamilyPolicy(policy string) error {
	switch policy {
	case IPFamilyPolicyAuto, IPFamilyPolicyPreferIPv6, IPFamilyPolicyIPv6Only, IPFamilyPolicyIPv4Only:
		return nil
	default:
		return fmt.Errorf("invalid IP family policy %q", policy)
	}
}

func GetDNSCacheDialContext(opts DNSCacheDialOptions) (DialFunc, error) {
	if err := ValidateIPFamilyPolicy(opts.IPFamilyPolicy); err != nil {
		return nil, err
	}

	resolver, err := dnscache.New(opts.Freq, opts.LookupTimeout, zap.NewNop())
	if err != nil {
		return nil, err
	}

	dialTimeout := opts.DialTimeout
	if dialTimeout <= 0 {
		dialTimeout = 30 * time.Second
	}
	keepAlive := opts.KeepAlive
	if keepAlive <= 0 {
		keepAlive = 30 * time.Second
	}
	fallbackDelay := opts.FallbackDelay
	if fallbackDelay <= 0 {
		fallbackDelay = DefaultIPv4FallbackDelay
	}

	d := &dnsCacheDialer{
		fetch:            resolver.Fetch,
		dial:             (&net.Dialer{Timeout: dialTimeout, KeepAlive: keepAlive}).DialContext,
		lookupTimeout:    opts.LookupTimeout,
		fallbackDelay:    fallbackDelay,
		ipFamilyPolicy:   opts.IPFamilyPolicy,
		dialDone:         opts.DialDone,
		fallbackStarted:  opts.FallbackStarted,
		connectionOpened: opts.ConnectionOpened,
		connectionClosed: opts.ConnectionClosed,
	}

	return d.DialContext, nil
}

type fetchIPFunc func(context.Context, string) ([]net.IP, error)

type dnsCacheDialer struct {
	fetch          fetchIPFunc
	dial           DialFunc
	lookupTimeout  time.Duration
	fallbackDelay  time.Duration
	ipFamilyPolicy string

	dialDone         func(family, result string, elapsed time.Duration)
	fallbackStarted  func()
	connectionOpened func(family, remote string)
	connectionClosed func(family string)
}

type dialResult struct {
	conn   net.Conn
	err    error
	family string
}

func (d *dnsCacheDialer) DialContext(ctx context.Context, _ string, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	lookupCtx := ctx
	if d.lookupTimeout > 0 {
		var cancel context.CancelFunc
		lookupCtx, cancel = context.WithTimeout(ctx, d.lookupTimeout)
		defer cancel()
	}

	ips, err := d.fetch(lookupCtx, host)
	if err != nil {
		return nil, err
	}

	ipv4, ipv6 := splitIPs(ips)
	var result dialResult
	switch d.ipFamilyPolicy {
	case IPFamilyPolicyAuto:
		result = d.dialAuto(ctx, ips, port)
	case IPFamilyPolicyPreferIPv6:
		result = d.dialPreferIPv6(ctx, ipv6, ipv4, port)
	case IPFamilyPolicyIPv6Only:
		if len(ipv6) == 0 {
			return nil, fmt.Errorf("no IPv6 address found for %q", host)
		}
		result = d.dialSerial(ctx, ipv6, port)
	case IPFamilyPolicyIPv4Only:
		if len(ipv4) == 0 {
			return nil, fmt.Errorf("no IPv4 address found for %q", host)
		}
		result = d.dialSerial(ctx, ipv4, port)
	}
	if result.err != nil {
		return nil, result.err
	}

	if d.connectionOpened != nil {
		d.connectionOpened(result.family, result.conn.RemoteAddr().String())
	}
	if d.connectionClosed != nil {
		result.conn = &observedConn{
			Conn:    result.conn,
			family:  result.family,
			onClose: d.connectionClosed,
		}
	}

	return result.conn, nil
}

func splitIPs(ips []net.IP) (ipv4, ipv6 []net.IP) {
	for _, ip := range ips {
		if ip.To4() != nil {
			ipv4 = append(ipv4, ip)
		} else if ip.To16() != nil {
			ipv6 = append(ipv6, ip)
		}
	}
	return ipv4, ipv6
}

func (d *dnsCacheDialer) dialAuto(ctx context.Context, ips []net.IP, port string) dialResult {
	validIPs := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		if ip.To16() != nil {
			validIPs = append(validIPs, ip)
		}
	}
	if len(validIPs) == 0 {
		return dialResult{err: errors.New("DNS lookup returned no IP addresses")}
	}

	ordered := make([]net.IP, 0, len(validIPs))
	for _, idx := range rand.Perm(len(validIPs)) { // Preserve the legacy random address selection policy.
		ordered = append(ordered, validIPs[idx])
	}
	return d.dialSerial(ctx, ordered, port)
}

func (d *dnsCacheDialer) dialPreferIPv6(ctx context.Context, ipv6, ipv4 []net.IP, port string) dialResult {
	if len(ipv6) == 0 {
		if len(ipv4) == 0 {
			return dialResult{err: errors.New("DNS lookup returned no IP addresses")}
		}
		return d.dialSerial(ctx, ipv4, port)
	}
	if len(ipv4) == 0 {
		return d.dialSerial(ctx, ipv6, port)
	}

	primaryCtx, primaryCancel := context.WithCancel(ctx)
	defer primaryCancel()
	fallbackCtx, fallbackCancel := context.WithCancel(ctx)
	defer fallbackCancel()

	results := make(chan dialResult)
	returned := make(chan struct{})
	defer close(returned)

	start := func(dialCtx context.Context, addresses []net.IP) {
		result := d.dialSerial(dialCtx, addresses, port)
		select {
		case results <- result:
		case <-returned:
			if result.conn != nil {
				_ = result.conn.Close()
			}
		}
	}

	go start(primaryCtx, ipv6)
	timer := time.NewTimer(d.fallbackDelay)
	defer timer.Stop()

	primaryDone := false
	fallbackDone := false
	fallbackRunning := false
	var primaryErr, fallbackErr error

	startFallback := func() {
		if fallbackRunning {
			return
		}
		fallbackRunning = true
		if d.fallbackStarted != nil {
			d.fallbackStarted()
		}
		go start(fallbackCtx, ipv4)
	}

	for {
		select {
		case <-ctx.Done():
			return dialResult{err: ctx.Err()}
		case <-timer.C:
			startFallback()
		case result := <-results:
			if result.err == nil {
				return result
			}
			if result.family == "ipv6" {
				primaryDone = true
				primaryErr = result.err
				if !fallbackRunning {
					if timer.Stop() {
						startFallback()
					}
				}
			} else {
				fallbackDone = true
				fallbackErr = result.err
			}

			if primaryDone && fallbackDone {
				return dialResult{err: errors.Join(primaryErr, fallbackErr)}
			}
		}
	}
}

func (d *dnsCacheDialer) dialSerial(ctx context.Context, ips []net.IP, port string) dialResult {
	if len(ips) == 0 {
		return dialResult{err: errors.New("no IP addresses to dial")}
	}

	var firstErr error
	for _, ip := range ips {
		if err := ctx.Err(); err != nil {
			return dialResult{err: err, family: familyOf(ips)}
		}

		family := "ipv6"
		network := "tcp6"
		if ip.To4() != nil {
			family = "ipv4"
			network = "tcp4"
		}

		start := time.Now()
		conn, err := d.dial(ctx, network, net.JoinHostPort(ip.String(), port))
		if d.dialDone != nil {
			result := "success"
			if err != nil {
				result = "failed"
				if errors.Is(err, context.Canceled) ||
					(errors.Is(err, context.DeadlineExceeded) && ctx.Err() != nil) {
					result = "canceled"
				}
			}
			d.dialDone(family, result, time.Since(start))
		}
		if err == nil {
			return dialResult{conn: conn, family: family}
		}
		if firstErr == nil {
			firstErr = err
		}
	}

	return dialResult{err: firstErr, family: familyOf(ips)}
}

func familyOf(ips []net.IP) string {
	if len(ips) > 0 && ips[0].To4() != nil {
		return "ipv4"
	}
	return "ipv6"
}

type observedConn struct {
	net.Conn
	family  string
	onClose func(string)
	once    sync.Once
}

func (c *observedConn) Close() error {
	err := c.Conn.Close()
	c.once.Do(func() { c.onClose(c.family) })
	return err
}
