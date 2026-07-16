// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"sort"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/go-ping/ping"
)

const (
	e2eProbeDelay                  = 20 * time.Millisecond
	maxConcurrentTCPE2EPerTask     = 10
	maxConcurrentTCPE2EAcrossTasks = 64
)

type e2eSample struct {
	sequence int
	rtt      time.Duration
	sent     bool
	received bool
	unknown  bool
	refused  bool
}

type e2eResult struct {
	protocol string
	destIP   string
	samples  []e2eSample
	err      error
}

var runE2EProbe = runProtocolE2E

var dialE2ETCP = func(ctx context.Context, address string) (net.Conn, error) {
	return (&net.Dialer{}).DialContext(ctx, "tcp4", address)
}

var tcpE2EGlobalSlots = make(chan struct{}, maxConcurrentTCPE2EAcrossTasks)

func runProtocolE2E(t task) e2eResult {
	switch t.Protocol {
	case protocolTCP:
		return runTCPE2E(t)
	case protocolUDP:
		return runUDPE2E(t)
	case protocolICMP:
		return runICMPE2E(t)
	default:
		return e2eResult{protocol: t.Protocol, err: fmt.Errorf("unsupported e2e protocol %q", t.Protocol)}
	}
}

func runProtocolE2EContext(ctx context.Context, t task) e2eResult {
	if ctx == nil {
		ctx = context.Background()
	}
	switch t.Protocol {
	case protocolTCP:
		return runTCPE2EContext(ctx, t)
	case protocolUDP:
		return runUDPE2EContext(ctx, t)
	case protocolICMP:
		return runICMPE2EContext(ctx, t)
	default:
		return e2eResult{protocol: t.Protocol, err: fmt.Errorf("unsupported e2e protocol %q", t.Protocol)}
	}
}

func resolveE2ETarget(t task) (net.IP, error) {
	return resolveTaskTarget(t, t.Target, t.Timeout)
}

func resolveE2ETargetContext(ctx context.Context, t task) (net.IP, error) {
	if ctx == nil || ctx.Done() == nil {
		return resolveE2ETarget(t)
	}
	return resolveTaskTargetContext(ctx, t, t.Target, t.Timeout)
}

func runTCPE2E(t task) e2eResult {
	return runTCPE2EContext(context.Background(), t)
}

func runTCPE2EContext(ctx context.Context, t task) e2eResult {
	if ctx == nil {
		ctx = context.Background()
	}
	result := e2eResult{protocol: protocolTCP}
	if t.Port == 0 {
		result.err = errors.New("tcp e2e target missing port")
		return result
	}
	ip, err := resolveE2ETargetContext(ctx, t)
	if err != nil {
		result.err = err
		return result
	}
	result.destIP = ip.String()
	queries := normalizeE2EQueries(t.E2EQueries)
	result.samples = make([]e2eSample, queries)
	dialErrors := make([]error, queries)
	address := net.JoinHostPort(result.destIP, strconv.Itoa(int(t.Port)))
	timeout := normalizeProbeTimeout(t.Timeout)

	var wg sync.WaitGroup
	taskSlots := make(chan struct{}, maxConcurrentTCPE2EPerTask)
	for sequence := 0; sequence < queries; sequence++ {
		result.samples[sequence] = e2eSample{sequence: sequence}
		if err := ctx.Err(); err != nil {
			result.err = err
			break
		}
		if err := acquireTCPE2ESlot(ctx, taskSlots); err != nil {
			result.err = err
			break
		}
		if err := acquireTCPE2ESlot(ctx, tcpE2EGlobalSlots); err != nil {
			<-taskSlots
			result.err = err
			break
		}
		wg.Add(1)
		go func(sequence int) {
			defer wg.Done()
			defer func() {
				<-tcpE2EGlobalSlots
				<-taskSlots
			}()
			probeCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			startedAt := time.Now()
			conn, err := dialE2ETCP(probeCtx, address)
			rtt := time.Since(startedAt)
			if err == nil {
				result.samples[sequence].sent = true
				result.samples[sequence].received = true
				result.samples[sequence].rtt = rtt
				if err := conn.Close(); err != nil {
					l.Debugf("close tcp e2e connection: %s", err.Error())
				}
				return
			}
			if errors.Is(err, syscall.ECONNREFUSED) {
				// A reset proves the probe reached the destination even though the
				// destination service rejected the connection.
				result.samples[sequence].sent = true
				result.samples[sequence].received = true
				result.samples[sequence].refused = true
				result.samples[sequence].rtt = rtt
				return
			}
			// Timeouts normally mean that a SYN was sent but no response arrived.
			// Other dial failures (for example ENETUNREACH or EACCES) happen before
			// a usable probe is emitted and must not be reported as packet loss.
			result.samples[sequence].sent = tcpDialTimedOut(err)
			dialErrors[sequence] = err
		}(sequence)
		if sequence+1 < queries {
			if err := waitE2EDelay(ctx); err != nil {
				result.err = err
				break
			}
		}
	}
	wg.Wait()
	if result.err == nil {
		for _, err := range dialErrors {
			if err != nil {
				result.err = fmt.Errorf("tcp e2e dial %s: %w", address, err)
				break
			}
		}
	}
	if result.err == nil {
		result.err = ctx.Err()
	}
	return result
}

func tcpDialTimedOut(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, syscall.ETIMEDOUT) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func runICMPE2E(t task) e2eResult {
	return runICMPE2EContext(context.Background(), t)
}

func runICMPE2EContext(ctx context.Context, t task) e2eResult {
	if ctx == nil {
		ctx = context.Background()
	}
	result := e2eResult{protocol: protocolICMP}
	ip, err := resolveE2ETargetContext(ctx, t)
	if err != nil {
		result.err = err
		return result
	}
	result.destIP = ip.String()
	queries := normalizeE2EQueries(t.E2EQueries)
	timeout := normalizeProbeTimeout(t.Timeout)

	pinger, err := ping.NewPinger(result.destIP)
	if err != nil {
		result.err = err
		return result
	}
	pinger.Count = queries
	pinger.Interval = e2eProbeDelay
	pinger.Timeout = timeout + time.Duration(queries-1)*e2eProbeDelay
	// Keep the normal ping TTL; MaxTTL only bounds traceroute.
	pinger.SetPrivileged(true)

	samples := make(map[int]e2eSample, queries)
	var mu sync.Mutex
	pinger.OnSend = func(packet *ping.Packet) {
		mu.Lock()
		samples[packet.Seq] = e2eSample{sequence: packet.Seq, sent: true}
		mu.Unlock()
	}
	pinger.OnRecv = func(packet *ping.Packet) {
		mu.Lock()
		sample := samples[packet.Seq]
		sample.sequence = packet.Seq
		sample.sent = true
		sample.received = true
		sample.rtt = packet.Rtt
		samples[packet.Seq] = sample
		mu.Unlock()
	}
	runDone := make(chan struct{})
	stopDone := make(chan struct{})
	go func() {
		defer close(stopDone)
		select {
		case <-ctx.Done():
			pinger.Stop()
		case <-runDone:
		}
	}()
	err = pinger.Run()
	close(runDone)
	<-stopDone
	if ctx.Err() != nil {
		result.err = ctx.Err()
	} else if err != nil {
		result.err = err
	}

	mu.Lock()
	result.samples = make([]e2eSample, 0, len(samples))
	for _, sample := range samples {
		result.samples = append(result.samples, sample)
	}
	mu.Unlock()
	sort.Slice(result.samples, func(i, j int) bool {
		return result.samples[i].sequence < result.samples[j].sequence
	})
	return result
}

func waitE2EDelay(ctx context.Context) error {
	timer := time.NewTimer(e2eProbeDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func acquireTCPE2ESlot(ctx context.Context, slots chan struct{}) error {
	select {
	case slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func mergeE2EResult(tags map[string]string, fields map[string]interface{}, result e2eResult, configuredQueries int) {
	if tags == nil || fields == nil {
		return
	}
	fields["e2e_queries"] = int64(normalizeE2EQueries(configuredQueries))

	sent, received, unknown, refused := 0, 0, 0, 0
	validSamples := make([]e2eSample, 0, len(result.samples))
	for _, sample := range result.samples {
		if !sample.sent {
			continue
		}
		sent++
		if sample.unknown {
			unknown++
		}
		if sample.received {
			received++
			validSamples = append(validSamples, sample)
		}
		if sample.refused {
			refused++
		}
	}
	fields["e2e_packets_sent"] = int64(sent)
	fields["e2e_packets_received"] = int64(received)
	fields["e2e_unknown"] = int64(unknown)
	if refused > 0 {
		fields["e2e_tcp_connection_refused"] = int64(refused)
	}
	if result.err != nil {
		fields["e2e_fail_reason"] = result.err.Error()
	}
	if result.destIP != "" {
		fields["e2e_dest_ip"] = result.destIP
		if tags["probe_dest_ip"] == "" && tags["probe_dest_ip_multiple"] != "true" {
			tags["probe_dest_ip"] = result.destIP
		}
	}

	determinate := sent - unknown
	if determinate > 0 {
		lost := determinate - received
		fields["e2e_probe_loss_percent"] = float64(lost) * 100 / float64(determinate)
	}
	switch {
	case sent > 0 && received == sent && result.err == nil:
		tags["e2e_status"] = "reached"
	case received > 0:
		tags["e2e_status"] = "partial"
	case sent > 0 && unknown == sent:
		tags["e2e_status"] = "unknown"
	default:
		tags["e2e_status"] = "failed"
	}
	addE2ELatencyFields(fields, validSamples)
}

func addE2ELatencyFields(fields map[string]interface{}, samples []e2eSample) {
	if len(samples) == 0 {
		return
	}
	sort.Slice(samples, func(i, j int) bool { return samples[i].sequence < samples[j].sequence })
	minRTT, maxRTT := samples[0].rtt, samples[0].rtt
	var totalRTT time.Duration
	var variationTotal time.Duration
	variationMax := time.Duration(0)
	variationSamples := 0
	for i, sample := range samples {
		totalRTT += sample.rtt
		if sample.rtt < minRTT {
			minRTT = sample.rtt
		}
		if sample.rtt > maxRTT {
			maxRTT = sample.rtt
		}
		if i == 0 || sample.sequence != samples[i-1].sequence+1 {
			continue
		}
		variation := time.Duration(math.Abs(float64(sample.rtt - samples[i-1].rtt)))
		variationTotal += variation
		variationSamples++
		if variation > variationMax {
			variationMax = variation
		}
	}
	fields["e2e_rtt_avg"] = float64(totalRTT.Microseconds()) / float64(len(samples))
	fields["e2e_rtt_min"] = float64(minRTT.Microseconds())
	fields["e2e_rtt_max"] = float64(maxRTT.Microseconds())
	fields["e2e_rtt_variation_samples"] = int64(variationSamples)
	if variationSamples > 0 {
		fields["e2e_rtt_variation_avg"] = float64(variationTotal.Microseconds()) / float64(variationSamples)
		fields["e2e_rtt_variation_max"] = float64(variationMax.Microseconds())
	}
}
