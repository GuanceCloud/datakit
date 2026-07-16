// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !linux && !windows

package traceroute

import (
	"context"
	"errors"
	"net"
	"time"
)

// UDPProber is unavailable outside Linux.
type UDPProber struct{}

func traceUDP(context.Context, net.IP, options) (Result, error) {
	return Result{}, errors.New("UDP traceroute is only supported on linux")
}

// ProbeUDP is unavailable outside Linux.
func ProbeUDP(context.Context, net.IP, uint16, int, time.Duration) (ProbeResult, error) {
	return ProbeResult{}, errors.New("UDP probe is only supported on linux")
}

// NewUDPProber is unavailable outside Linux.
func NewUDPProber() (*UDPProber, error) {
	return nil, errors.New("UDP probe is only supported on linux")
}

// Probe is unavailable outside Linux.
func (*UDPProber) Probe(context.Context, net.IP, uint16, int, time.Duration) (ProbeResult, error) {
	return ProbeResult{}, errors.New("UDP probe is only supported on linux")
}

// Close is a no-op outside Linux.
func (*UDPProber) Close() error { return nil }
