// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows

package traceroute

import (
	"context"
	"errors"
	"net"
	"time"
)

// UDPProber is unavailable on Windows.
type UDPProber struct{}

func trace(context.Context, net.IP, options) (Result, error) {
	return Result{}, errors.New("traceroute is not supported on windows")
}

// ProbeUDP is unavailable on Windows.
func ProbeUDP(context.Context, net.IP, uint16, int, time.Duration) (ProbeResult, error) {
	return ProbeResult{}, errors.New("UDP probe is not supported on windows")
}

// NewUDPProber is unavailable on Windows.
func NewUDPProber() (*UDPProber, error) {
	return nil, errors.New("UDP probe is not supported on windows")
}

// Probe is unavailable on Windows.
func (*UDPProber) Probe(context.Context, net.IP, uint16, int, time.Duration) (ProbeResult, error) {
	return ProbeResult{}, errors.New("UDP probe is not supported on windows")
}

// Close is a no-op on Windows.
func (*UDPProber) Close() error { return nil }
