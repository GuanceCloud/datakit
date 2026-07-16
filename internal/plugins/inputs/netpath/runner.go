// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package netpath

import (
	"context"
	"fmt"
	"strconv"
	"time"

	tr "github.com/GuanceCloud/cliutils/traceroute"
)

type probeResult struct {
	task        task
	tags        map[string]string
	fields      map[string]interface{}
	duration    time.Duration
	scheduledAt time.Time
	startedAt   time.Time
	finishedAt  time.Time
	testRunID   string
	failType    string
	err         error
}

type probeTask interface {
	Check() error
	Run() error
	GetResults() (map[string]string, map[string]interface{})
}

type contextProbeTask interface {
	RunContext(ctx context.Context) error
}

type probeBuilder func(t task) (probeTask, error)

type e2eProbeRunner func(t task) e2eResult

func runProbe(t task) (res probeResult) {
	return runProbeWithContext(context.Background(), t, runE2EProbe, buildProbeTask)
}

func runProbeContext(ctx context.Context, t task) probeResult {
	if ctx == nil {
		ctx = context.Background()
	}
	probeCtx, cancel := context.WithTimeout(ctx, maxProbeRunDuration)
	defer cancel()
	return runProbeWithContext(probeCtx, t, func(t task) e2eResult {
		return runProtocolE2EContext(probeCtx, t)
	}, buildProbeTaskContext)
}

func runProbeWithContext(ctx context.Context, t task, runE2E e2eProbeRunner,
	buildProbe probeBuilder,
) (res probeResult) {
	if ctx == nil {
		ctx = context.Background()
	}
	t.MaxTTL = normalizeMaxTTL(t.MaxTTL, t.Protocol)
	t.Timeout = normalizeProbeTimeout(t.Timeout)
	start := time.Now()
	var e2eCh <-chan e2eResult
	if t.E2EQueries > 0 {
		ch := make(chan e2eResult, 1)
		e2eCh = ch
		go func() {
			ch <- runE2E(t)
		}()
	}
	res = probeResult{
		task:        t,
		tags:        map[string]string{},
		fields:      map[string]interface{}{},
		scheduledAt: t.ScheduledAt,
		startedAt:   start,
		testRunID:   makeRunID(t, start),
	}
	if res.scheduledAt.IsZero() {
		res.scheduledAt = start
	}
	defer func() {
		if e2eCh != nil {
			e2e := <-e2eCh
			// Preserve completed samples even when the overall probe deadline
			// cancels the remaining E2E work. The E2E runner records the
			// cancellation in its result, so callers still see why it stopped.
			mergeE2EResult(res.tags, res.fields, e2e, t.E2EQueries)
		}
		if res.finishedAt.IsZero() {
			res.finishedAt = time.Now()
		}
		res.duration = res.finishedAt.Sub(start)
		if res.err != nil && res.failType == "" {
			res.failType = classifyFailure(res.err.Error())
		}
	}()

	if err := ctx.Err(); err != nil {
		res.err = err
		return
	}
	probe, err := buildProbe(t)
	if err != nil {
		res.err = err
		return
	}
	if err := probe.Check(); err != nil {
		res.err = err
		return
	}
	if err := ctx.Err(); err != nil {
		res.err = err
		return
	}
	if contextProbe, ok := probe.(contextProbeTask); ok {
		err = contextProbe.RunContext(ctx)
	} else {
		err = probe.Run()
	}
	if err != nil {
		res.err = err
	} else if ctxErr := ctx.Err(); ctxErr != nil {
		res.err = ctxErr
	}
	tags, fields := probe.GetResults()
	res.tags = tags
	res.fields = normalizeRunnerFields(fields)
	if _, ok := res.tags["traceroute_protocol"]; !ok {
		res.tags["traceroute_protocol"] = tracerouteProtocol(t.Protocol)
	}
	if res.err != nil {
		if res.tags["traceroute_status"] == "" {
			res.tags["traceroute_status"] = "failed"
		}
		if _, ok := res.fields["traceroute_fail_reason"]; !ok {
			res.fields["traceroute_fail_reason"] = res.err.Error()
		}
	}
	return
}

func routesToRun(routes []*tr.Route, runID int) tracerouteRun {
	run := tracerouteRun{RunID: strconv.Itoa(runID), Hops: make([]tracerouteHop, 0, len(routes))}
	for idx, route := range routes {
		item := routeItem{IP: "*"}
		if route != nil && len(route.Items) > 0 && route.Items[0] != nil {
			item.IP = route.Items[0].IP
			item.ResponseTime = route.Items[0].ResponseTime
		}
		run.Hops = append(run.Hops, routeItemToHop(idx+1, item))
	}
	return run
}

func tracerouteProtocol(protocol string) string {
	switch protocol {
	case protocolTCP, protocolUDP, protocolICMP:
		return protocol
	default:
		return protocolICMP
	}
}

func normalizeRunnerFields(fields map[string]interface{}) map[string]interface{} {
	if fields == nil {
		return map[string]interface{}{}
	}
	normalizeTracerouteFields(fields)
	return fields
}

func makeRunID(t task, startedAt time.Time) string {
	if startedAt.IsZero() {
		startedAt = time.Now()
	}
	return "run-" + shortHash(fmt.Sprintf("%s|%s|%d", t.ID, t.ScheduleKey, startedAt.UnixNano()))
}

func buildProbeTask(t task) (probeTask, error) {
	switch t.Protocol {
	case protocolTCP:
		if t.Port == 0 {
			return nil, fmt.Errorf("tcp netpath target %q missing port", t.Target)
		}
	case protocolUDP:
		if t.Port == 0 {
			return nil, fmt.Errorf("udp netpath target %q missing port", t.Target)
		}
	case protocolICMP:
	default:
		return nil, fmt.Errorf("unsupported netpath protocol %q", t.Protocol)
	}
	return newTracerouteProbe(t, resolverForTask(t)), nil
}

func buildProbeTaskContext(t task) (probeTask, error) {
	probe, err := buildProbeTask(t)
	if err != nil {
		return nil, err
	}
	resolver := contextResolverForTask(t)
	if probe, ok := probe.(*tracerouteProbe); ok {
		probe.contextResolver = resolver
	}
	return probe, nil
}

func normalizeMaxTTL(ttl int, protocol string) int {
	if ttl <= 0 {
		ttl = defaultMaxTTL
	}
	limit := maxTTL
	if protocol != protocolUDP && tr.MaxHops < limit {
		limit = tr.MaxHops
	}
	if ttl > limit {
		return limit
	}
	return ttl
}
