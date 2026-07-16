// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package netpath

import (
	"context"
	"errors"

	tr "github.com/GuanceCloud/cliutils/traceroute"
)

func runUDPE2E(t task) e2eResult {
	return runUDPE2EContext(context.Background(), t)
}

func runUDPE2EContext(ctx context.Context, t task) e2eResult {
	if ctx == nil {
		ctx = context.Background()
	}
	result := e2eResult{protocol: protocolUDP}
	if t.Port == 0 {
		result.err = errors.New("udp e2e target missing port")
		return result
	}
	target, err := resolveE2ETargetContext(ctx, t)
	if err != nil {
		result.err = err
		return result
	}
	result.destIP = target.String()

	timeout := normalizeProbeTimeout(t.Timeout)
	queries := normalizeE2EQueries(t.E2EQueries)
	result.samples = make([]e2eSample, 0, queries)
	prober, err := tr.NewUDPProber()
	if err != nil {
		result.err = err
		return result
	}
	defer func() {
		if err := prober.Close(); err != nil {
			l.Debugf("close udp e2e prober: %s", err.Error())
		}
	}()
	for sequence := 0; sequence < queries; sequence++ {
		sample := e2eSample{sequence: sequence}
		reply, err := prober.Probe(ctx, target, t.Port, 0, timeout)
		sample.sent = reply.Sent
		if err != nil {
			result.err = err
			result.samples = append(result.samples, sample)
			break
		}
		sample.rtt = reply.RTT
		sample.received = reply.Reached
		// A silent UDP application is ambiguous: an open service is not
		// required to respond to an arbitrary datagram, so silence is not loss.
		sample.unknown = reply.TimedOut
		result.samples = append(result.samples, sample)
		if sequence+1 < queries {
			if err := waitE2EDelay(ctx); err != nil {
				result.err = err
				break
			}
		}
	}
	return result
}
