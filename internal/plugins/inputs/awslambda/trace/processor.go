// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package lambdatrace

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/logger"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/inferred"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/model"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/telemetry"
)

var traceLog = logger.DefaultSLogger("awslambda-trace")

const (
	// Datadog's Lambda extension keeps invocation state in bounded FIFO buffers
	// instead of time-based TTL caches. 500 is their default context capacity.
	processorContextCapacity       = 500
	processorPendingItemCapacity   = 500
	processorPendingTracerCapacity = 500
	maxCarrierScanDepth            = 16
	maxCarrierEncodedStringBytes   = 1 << 20
)

func SetLogger(l *logger.Logger) {
	if l == nil {
		return
	}
	traceLog = l
}

type InvocationContext struct {
	RequestID        string
	TraceID          uint64
	InvocationSpanID uint64
	InferredSpanID   uint64
	ColdStartSpanID  uint64
	RestoreSpanID    uint64
	HasInferred      bool
	ColdStart        bool
	UniversalStarted bool
	UniversalEnded   bool
	ReportReceived   bool
	RuntimeDone      bool
	Error            bool
	StartTime        time.Time
	Duration         time.Duration
	RuntimeDuration  time.Duration
	InitDuration     time.Duration
	RestoreDuration  time.Duration
	RequestPayload   []byte
	ResponsePayload  []byte
	Trigger          *inferred.Info
	ManagedInstance  bool
	Upstream         *upstreamContext
	TracerMeta       map[string]string
	TracerMetrics    map[string]float64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Processor struct {
	mu                sync.Mutex
	contexts          map[string]*InvocationContext
	contextOrder      []string
	deferredColdStart map[string]deferredColdStart
	sink              Sink
	firstInvocation   bool
	pendingInitDur    time.Duration
	pendingRestoreDur time.Duration
	managedInstance   bool
	pendingInvokeIDs  []pendingRequestID
	pendingStarts     []*InvocationContext
	pendingRuntimeIDs []pendingRequestID
	pendingEnds       []pendingUniversalEnd
	pendingTracer     map[tracerPlaceholderKey]*tracerPlaceholder
}

type pendingRequestID struct {
	id        string
	createdAt time.Time
}

type pendingUniversalEnd struct {
	headers   map[string]string
	payload   []byte
	createdAt time.Time
}

type tracerPlaceholderKey struct {
	traceID          uint64
	invocationSpanID uint64
}

type tracerPlaceholder struct {
	meta      map[string]string
	metrics   map[string]float64
	requestID string
	createdAt time.Time
}

type deferredColdStart struct {
	RequestID           string
	TraceID             uint64
	SpanID              uint64
	Service             string
	InvocationStartUnix int64
}

type upstreamContext struct {
	TraceID  uint64
	ParentID uint64
	Meta     map[string]string
}

type w3cContext struct {
	TraceIDHigh string
	TraceIDLow  uint64
	ParentID    uint64
	Meta        map[string]string
}

var activeProcessor atomic.Pointer[Processor]

func SetActiveProcessor(p *Processor) {
	activeProcessor.Store(p)
}

func CaptureTracerPlaceholder(traceID, invocationSpanID uint64, meta map[string]string, metrics map[string]float64) {
	if p := activeProcessor.Load(); p != nil {
		p.OnTracerPlaceholder(traceID, invocationSpanID, meta, metrics)
	}
}

func NewProcessor(sink Sink, managed bool) *Processor {
	return &Processor{
		contexts:          map[string]*InvocationContext{},
		contextOrder:      make([]string, 0, 4),
		deferredColdStart: map[string]deferredColdStart{},
		sink:              sink,
		firstInvocation:   true,
		managedInstance:   managed,
		pendingInvokeIDs:  make([]pendingRequestID, 0, 4),
		pendingStarts:     make([]*InvocationContext, 0, 4),
		pendingRuntimeIDs: make([]pendingRequestID, 0, 4),
		pendingEnds:       make([]pendingUniversalEnd, 0, 4),
		pendingTracer:     map[tracerPlaceholderKey]*tracerPlaceholder{},
	}
}

func (p *Processor) OnTracerPlaceholder(traceID, invocationSpanID uint64, meta map[string]string, metrics map[string]float64) {
	if traceID == 0 || invocationSpanID == 0 {
		return
	}

	now := time.Now()
	placeholder := &tracerPlaceholder{
		createdAt: now,
		meta:      cloneMap(meta),
		metrics:   cloneFloatMap(metrics),
		requestID: meta["request_id"],
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.cleanupLocked()

	if ctx := p.findContextByTraceAndSpanLocked(traceID, invocationSpanID); ctx != nil {
		mergeTracerPlaceholderLocked(ctx, placeholder)
		return
	}

	if placeholder.requestID != "" {
		if ctx, ok := p.contexts[placeholder.requestID]; ok {
			mergeTracerPlaceholderLocked(ctx, placeholder)
			return
		}
	}

	key := tracerPlaceholderKey{traceID: traceID, invocationSpanID: invocationSpanID}
	p.pendingTracer[key] = placeholder
	p.cleanupLocked()
}

func (p *Processor) OnInvokeEvent(requestID string) {
	if requestID == "" {
		return
	}
	now := time.Now()
	p.mu.Lock()
	defer p.mu.Unlock()
	ctx := p.ensureContextLocked(requestID)
	if len(p.pendingStarts) > 0 {
		p.mergePendingStartLocked(ctx, p.pendingStarts[0])
		p.pendingStarts = p.pendingStarts[1:]
	} else {
		p.pendingInvokeIDs = append(p.pendingInvokeIDs, pendingRequestID{id: requestID, createdAt: now})
	}
	touchContextLocked(ctx, now)
	p.cleanupLocked()
	_, existed := p.contexts[requestID]
	traceLog.Debugf(
		"OnInvokeEvent request_id=%s existed=%t trace_id=%d invocation_span_id=%d universal_started=%t universal_ended=%t report_received=%t",
		requestID, existed, ctx.TraceID, ctx.InvocationSpanID, ctx.UniversalStarted, ctx.UniversalEnded, ctx.ReportReceived,
	)
}

func (p *Processor) OnTelemetryEvent(event *telemetry.Event) error {
	switch record := event.Record.(type) {
	case *telemetry.PlatformInitReport:
		p.OnPlatformInitReport(record.Metrics.DurationMs)
	case *telemetry.PlatformRestoreReport:
		p.OnPlatformRestoreReport(record.Metrics.DurationMs)
	case *telemetry.PlatformStart:
		p.OnPlatformStart(record.RequestID, event.Time)
	case *telemetry.PlatformRuntimeDone:
		duration := 0.0
		if record.Metrics != nil {
			duration = record.Metrics.DurationMs
		}
		return p.OnPlatformRuntimeDone(record.RequestID, duration, string(record.Status))
	case *telemetry.PlatformReport:
		duration := record.Metrics.DurationMs
		if record.Metrics.InitDurationMs != nil {
			if err := p.OnPlatformReportInitDuration(record.RequestID, *record.Metrics.InitDurationMs); err != nil {
				return err
			}
		}
		return p.OnPlatformReport(record.RequestID, duration, string(record.Status))
	}
	return nil
}

func (p *Processor) OnPlatformInitReport(durationMs float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pendingInitDur = time.Duration(durationMs * float64(time.Millisecond))
}

func (p *Processor) OnPlatformRestoreReport(durationMs float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pendingRestoreDur = time.Duration(durationMs * float64(time.Millisecond))
}

func (p *Processor) OnPlatformReportInitDuration(requestID string, durationMs float64) error {
	if requestID == "" || durationMs <= 0 {
		return nil
	}

	initDuration := time.Duration(durationMs * float64(time.Millisecond))

	p.mu.Lock()
	if ctx, ok := p.contexts[requestID]; ok {
		if ctx.InitDuration == 0 {
			ctx.InitDuration = initDuration
		}
		touchContextLocked(ctx, time.Now())
		p.mu.Unlock()
		return nil
	}

	deferred, ok := p.deferredColdStart[requestID]
	if !ok {
		p.mu.Unlock()
		return nil
	}
	delete(p.deferredColdStart, requestID)
	p.mu.Unlock()

	span := Span{
		TraceID:  deferred.TraceID,
		SpanID:   deferred.SpanID,
		Name:     "aws.lambda.cold_start",
		Resource: "aws.lambda.cold_start",
		Service:  deferred.Service,
		Type:     "serverless",
		Start:    deferred.InvocationStartUnix - initDuration.Nanoseconds(),
		Duration: initDuration.Nanoseconds(),
		Meta: map[string]string{
			"request_id": requestIDOrUnknown(deferred.RequestID),
		},
		Metrics: map[string]float64{},
	}

	traceLog.Debugf(
		"deferred cold start emitted request_id=%s trace_id=%d span_id=%d init_duration_ms=%.2f",
		requestID, deferred.TraceID, deferred.SpanID, durationMs,
	)

	return p.sink.Consume(context.Background(), []Span{span})
}

func (p *Processor) OnPlatformStart(requestID string, ts time.Time) {
	if requestID == "" {
		return
	}
	now := time.Now()
	p.mu.Lock()
	defer p.mu.Unlock()
	ctx := p.ensureContextLocked(requestID)
	ctx.StartTime = ts
	touchContextLocked(ctx, now)
}

func (p *Processor) OnPlatformRuntimeDone(requestID string, durationMs float64, status string) error {
	if requestID == "" {
		return nil
	}
	now := time.Now()
	p.mu.Lock()
	ctx := p.ensureContextLocked(requestID)
	ctx.RuntimeDone = true
	ctx.RuntimeDuration = time.Duration(durationMs * float64(time.Millisecond))
	if ctx.Duration == 0 {
		ctx.Duration = ctx.RuntimeDuration
	}
	if status != "" && status != string(model.StatusSuccess) {
		ctx.Error = true
	}
	if len(p.pendingEnds) > 0 {
		p.processUniversalEndLocked(requestID, p.pendingEnds[0].headers, p.pendingEnds[0].payload)
		p.pendingEnds = p.pendingEnds[1:]
	} else {
		p.pendingRuntimeIDs = append(p.pendingRuntimeIDs, pendingRequestID{id: requestID, createdAt: now})
	}
	touchContextLocked(ctx, now)
	p.cleanupLocked()
	ready := ctx.UniversalEnded
	traceLog.Debugf(
		"OnPlatformRuntimeDone request_id=%s duration_ms=%.2f status=%s ready=%t "+
			"trace_id=%d invocation_span_id=%d universal_started=%t universal_ended=%t report_received=%t",
		requestID, durationMs, status, ready, ctx.TraceID, ctx.InvocationSpanID, ctx.UniversalStarted, ctx.UniversalEnded, ctx.ReportReceived,
	)
	p.mu.Unlock()
	if ready {
		return p.flush(requestID)
	}
	return nil
}

func (p *Processor) OnPlatformReport(requestID string, durationMs float64, status string) error {
	if requestID == "" {
		return nil
	}
	now := time.Now()
	p.mu.Lock()
	ctx, ok := p.contexts[requestID]
	if !ok {
		traceLog.Debugf("OnPlatformReport ignored unknown request_id=%s duration_ms=%.2f status=%s", requestID, durationMs, status)
		p.mu.Unlock()
		return nil
	}
	ctx.ReportReceived = true
	if ctx.Duration == 0 {
		ctx.Duration = time.Duration(durationMs * float64(time.Millisecond))
	}
	if status != "" && status != string(model.StatusSuccess) {
		ctx.Error = true
	}
	touchContextLocked(ctx, now)
	ready := ctx.ManagedInstance && ctx.UniversalStarted && ctx.UniversalEnded && !ctx.RuntimeDone
	traceLog.Debugf(
		"OnPlatformReport request_id=%s duration_ms=%.2f status=%s ready=%t "+
			"trace_id=%d invocation_span_id=%d universal_started=%t universal_ended=%t report_received=%t",
		requestID, durationMs, status, ready, ctx.TraceID, ctx.InvocationSpanID, ctx.UniversalStarted, ctx.UniversalEnded, ctx.ReportReceived,
	)
	p.mu.Unlock()
	if ready {
		return p.flush(requestID)
	}
	return nil
}

func (p *Processor) OnUniversalStart(requestID string, headers map[string]string, payload []byte) (uint64, uint64) {
	now := time.Now()
	p.mu.Lock()
	defer p.mu.Unlock()

	if requestID == "" && len(p.pendingInvokeIDs) > 0 {
		requestID = p.pendingInvokeIDs[0].id
		p.pendingInvokeIDs = p.pendingInvokeIDs[1:]
	}

	ctx := p.buildUniversalStartContextLocked(requestID, headers, payload)
	if requestID == "" {
		p.pendingStarts = append(p.pendingStarts, ctx)
		p.cleanupLocked()
		traceLog.Debugf(
			"OnUniversalStart buffered pending start trace_id=%d invocation_span_id=%d pending_starts=%d",
			ctx.TraceID, ctx.InvocationSpanID, len(p.pendingStarts),
		)
		return ctx.TraceID, ctx.InvocationSpanID
	}

	target := p.ensureContextLocked(requestID)
	p.mergePendingStartLocked(target, ctx)
	touchContextLocked(target, now)
	traceLog.Debugf(
		"OnUniversalStart request_id=%s trace_id=%d invocation_span_id=%d universal_started=%t universal_ended=%t report_received=%t",
		requestID, ctx.TraceID, ctx.InvocationSpanID, ctx.UniversalStarted, ctx.UniversalEnded, ctx.ReportReceived,
	)
	return target.TraceID, target.InvocationSpanID
}

func (p *Processor) OnUniversalEnd(requestID string, headers map[string]string, payload []byte) error {
	now := time.Now()
	p.mu.Lock()
	if requestID == "" && len(p.pendingRuntimeIDs) > 0 {
		requestID = p.pendingRuntimeIDs[0].id
		p.pendingRuntimeIDs = p.pendingRuntimeIDs[1:]
	}
	if requestID == "" {
		p.pendingEnds = append(p.pendingEnds, pendingUniversalEnd{
			headers:   cloneMap(headers),
			payload:   cloneBytes(payload),
			createdAt: now,
		})
		p.cleanupLocked()
		traceLog.Debugf("OnUniversalEnd buffered pending end pending_ends=%d", len(p.pendingEnds))
		p.mu.Unlock()
		return nil
	}
	ready := p.processUniversalEndLocked(requestID, headers, payload)
	p.mu.Unlock()
	if ready {
		return p.flush(requestID)
	}
	return nil
}

func (p *Processor) OnShutdown() error {
	p.mu.Lock()
	ids := make([]string, 0, len(p.contexts))
	for id, ctx := range p.contexts {
		if ctx.UniversalStarted || ctx.ReportReceived {
			ids = append(ids, id)
		}
	}
	p.mu.Unlock()
	for _, id := range ids {
		if err := p.flush(id); err != nil {
			return err
		}
	}
	return nil
}

func (p *Processor) flush(requestID string) error {
	p.mu.Lock()
	ctx, ok := p.contexts[requestID]
	if !ok {
		p.mu.Unlock()
		return nil
	}
	p.mergePendingTracerLocked(ctx)
	traceLog.Debugf(
		"flush request_id=%s trace_id=%d invocation_span_id=%d "+
			"duration_ms=%.2f runtime_duration_ms=%.2f "+
			"universal_started=%t universal_ended=%t report_received=%t runtime_done=%t",
		requestID, ctx.TraceID, ctx.InvocationSpanID,
		ctx.Duration.Seconds()*1000, ctx.RuntimeDuration.Seconds()*1000,
		ctx.UniversalStarted, ctx.UniversalEnded, ctx.ReportReceived, ctx.RuntimeDone,
	)
	p.captureDeferredColdStartLocked(ctx)
	spans := p.buildTraceLocked(ctx)
	delete(p.contexts, requestID)
	p.mu.Unlock()
	if len(spans) == 0 {
		return nil
	}
	return p.sink.Consume(context.Background(), spans)
}

func (p *Processor) buildTraceLocked(ctx *InvocationContext) []Span {
	start := ctx.StartTime
	if start.IsZero() {
		start = time.Now()
	}
	startNS := start.UnixNano()
	durationNS := ctx.Duration.Nanoseconds()
	if durationNS == 0 {
		durationNS = (100 * time.Millisecond).Nanoseconds()
	}

	spans := make([]Span, 0, 4)
	if ctx.HasInferred && ctx.Trigger != nil {
		inferredParentID := uint64(0)
		if ctx.Upstream != nil {
			inferredParentID = ctx.Upstream.ParentID
		}
		inferredMeta := cloneMap(ctx.Trigger.Tags)
		if ctx.Upstream != nil {
			for k, v := range ctx.Upstream.Meta {
				inferredMeta[k] = v
			}
		}
		spans = append(spans, Span{
			TraceID:  ctx.TraceID,
			SpanID:   ctx.InferredSpanID,
			ParentID: inferredParentID,
			Name:     ctx.Trigger.Operation,
			Resource: emptyFallback(ctx.Trigger.Resource, ctx.Trigger.Operation),
			Service:  emptyFallback(ctx.Trigger.Service, "aws-trigger"),
			Type:     "serverless",
			Start:    startNS,
			Duration: durationNS,
			Meta:     inferredMeta,
			Metrics:  map[string]float64{},
		})
	}

	parentID := uint64(0)
	if ctx.HasInferred {
		parentID = ctx.InferredSpanID
	} else if ctx.Upstream != nil {
		parentID = ctx.Upstream.ParentID
	}

	service, resource := lambdaInvocationServiceAndResource()

	meta := map[string]string{
		"request_id": requestIDOrUnknown(ctx.RequestID),
	}
	for k, v := range ctx.TracerMeta {
		meta[k] = v
	}
	meta["request_id"] = requestIDOrUnknown(ctx.RequestID)
	if ctx.ColdStart {
		meta["cold_start"] = "true"
	}
	if ctx.ManagedInstance {
		meta["init_type"] = "lambda-managed-instances"
	}
	if ctx.Upstream != nil && !ctx.HasInferred {
		for k, v := range ctx.Upstream.Meta {
			meta[k] = v
		}
	}

	spans = append(spans, Span{
		TraceID:  ctx.TraceID,
		SpanID:   ctx.InvocationSpanID,
		ParentID: parentID,
		Name:     "aws.lambda",
		Resource: resource,
		Service:  service,
		Type:     "serverless",
		Start:    startNS,
		Duration: durationNS,
		Error:    boolToError(ctx.Error),
		Meta:     meta,
		Metrics:  cloneFloatMap(ctx.TracerMetrics),
	})

	if ctx.ColdStart && !ctx.ManagedInstance {
		initNS := ctx.InitDuration.Nanoseconds()
		if initNS > 0 {
			spans = append(spans, Span{
				TraceID:  ctx.TraceID,
				SpanID:   newID(),
				Name:     "aws.lambda.cold_start",
				Resource: "aws.lambda.cold_start",
				Service:  service,
				Type:     "serverless",
				Start:    startNS - initNS,
				Duration: initNS,
				Meta: map[string]string{
					"request_id": requestIDOrUnknown(ctx.RequestID),
				},
				Metrics: map[string]float64{},
			})
		}
	}

	if ctx.RestoreDuration > 0 {
		restoreNS := ctx.RestoreDuration.Nanoseconds()
		spans = append(spans, Span{
			TraceID:  ctx.TraceID,
			SpanID:   newID(),
			Name:     "aws.lambda.snapstart_restore",
			Resource: "aws.lambda.snapstart_restore",
			Service:  service,
			Type:     "serverless",
			Start:    startNS - restoreNS,
			Duration: restoreNS,
			Meta: map[string]string{
				"request_id": requestIDOrUnknown(ctx.RequestID),
			},
			Metrics: map[string]float64{},
		})
	}

	return spans
}

func (p *Processor) ensureContextLocked(requestID string) *InvocationContext {
	if ctx, ok := p.contexts[requestID]; ok {
		touchContextLocked(ctx, time.Now())
		return ctx
	}
	now := time.Now()
	ctx := &InvocationContext{
		RequestID:       requestID,
		ColdStart:       p.firstInvocation && !p.managedInstance,
		InitDuration:    p.pendingInitDur,
		RestoreDuration: p.pendingRestoreDur,
		ManagedInstance: p.managedInstance,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	p.pendingInitDur = 0
	p.pendingRestoreDur = 0
	p.firstInvocation = false
	p.contexts[requestID] = ctx
	p.contextOrder = append(p.contextOrder, requestID)
	p.cleanupLocked()
	return ctx
}

func (p *Processor) buildUniversalStartContextLocked(requestID string, headers map[string]string, payload []byte) *InvocationContext {
	now := time.Now()
	ctx := &InvocationContext{
		RequestID:       requestID,
		ColdStart:       p.firstInvocation && !p.managedInstance,
		InitDuration:    p.pendingInitDur,
		RestoreDuration: p.pendingRestoreDur,
		ManagedInstance: p.managedInstance,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	ctx.UniversalStarted = true
	ctx.RequestPayload = cloneBytes(payload)
	ctx.StartTime = now
	if upstream := extractDatadogContext(headers, payload); upstream != nil {
		ctx.TraceID = upstream.TraceID
		ctx.Upstream = upstream
	} else if traceID, ok := parseHeaderUint(headers, "x-datadog-trace-id"); ok {
		ctx.TraceID = traceID
	} else {
		ctx.TraceID = newID()
	}
	ctx.InvocationSpanID = newID()
	ctx.Trigger = inferred.Detect(payload)
	if ctx.Trigger != nil {
		ctx.HasInferred = true
		ctx.InferredSpanID = newID()
	}
	return ctx
}

func (p *Processor) mergePendingStartLocked(dst, src *InvocationContext) {
	if dst == nil || src == nil {
		return
	}
	if dst.TraceID == 0 {
		dst.TraceID = src.TraceID
	}
	if dst.InvocationSpanID == 0 {
		dst.InvocationSpanID = src.InvocationSpanID
	}
	if dst.InferredSpanID == 0 {
		dst.InferredSpanID = src.InferredSpanID
	}
	if !dst.HasInferred {
		dst.HasInferred = src.HasInferred
		dst.Trigger = src.Trigger
	}
	if dst.StartTime.IsZero() {
		dst.StartTime = src.StartTime
	}
	if dst.RequestPayload == nil {
		dst.RequestPayload = src.RequestPayload
	}
	if !dst.UniversalStarted {
		dst.UniversalStarted = src.UniversalStarted
	}
	if !dst.ColdStart {
		dst.ColdStart = src.ColdStart
	}
	if dst.InitDuration == 0 {
		dst.InitDuration = src.InitDuration
	}
	if dst.RestoreDuration == 0 {
		dst.RestoreDuration = src.RestoreDuration
	}
	if dst.Upstream == nil {
		dst.Upstream = src.Upstream
	}
	touchContextLocked(dst, time.Now())
	p.mergePendingTracerLocked(dst)
}

func (p *Processor) processUniversalEndLocked(requestID string, headers map[string]string, payload []byte) bool {
	ctx := p.ensureContextLocked(requestID)
	ctx.UniversalEnded = true
	ctx.ResponsePayload = cloneBytes(payload)
	if headers["x-datadog-invocation-error"] == "true" {
		ctx.Error = true
	}
	touchContextLocked(ctx, time.Now())
	ready := ctx.RuntimeDone || ctx.ReportReceived
	traceLog.Debugf(
		"OnUniversalEnd request_id=%s ready=%t trace_id=%d invocation_span_id=%d "+
			"duration_ms=%.2f universal_started=%t universal_ended=%t report_received=%t",
		requestID, ready, ctx.TraceID, ctx.InvocationSpanID,
		ctx.Duration.Seconds()*1000, ctx.UniversalStarted, ctx.UniversalEnded, ctx.ReportReceived,
	)
	return ready
}

func parseUint(v string) (uint64, bool) {
	if v == "" {
		return 0, false
	}
	value, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, false
	}
	return value, true
}

func parseHeaderUint(headers map[string]string, key string) (uint64, bool) {
	return parseUint(headerValue(headers, key))
}

func extractDatadogContext(headers map[string]string, payload []byte) *upstreamContext {
	for _, carrier := range upstreamCarriers(payload, headers) {
		if upstream := extractDatadogContextFromCarrier(carrier); upstream != nil {
			return upstream
		}
	}
	for _, carrier := range upstreamCarriers(payload, headers) {
		if w3c := extractW3CContextFromCarrier(carrier); w3c != nil {
			return &upstreamContext{
				TraceID:  w3c.TraceIDLow,
				ParentID: w3c.ParentID,
				Meta:     w3c.Meta,
			}
		}
		if b3 := extractB3ContextFromCarrier(carrier); b3 != nil {
			return &upstreamContext{
				TraceID:  b3.TraceIDLow,
				ParentID: b3.ParentID,
				Meta:     b3.Meta,
			}
		}
		if xray := extractXRayContextFromCarrier(carrier); xray != nil {
			return &upstreamContext{
				TraceID:  xray.TraceIDLow,
				ParentID: xray.ParentID,
				Meta:     xray.Meta,
			}
		}
	}
	return nil
}

func extractDatadogContextFromCarrier(carrier map[string]string) *upstreamContext {
	traceID, ok := parseHeaderUint(carrier, "x-datadog-trace-id")
	if !ok || traceID == 0 {
		return nil
	}
	parentID, _ := parseHeaderUint(carrier, "x-datadog-parent-id")

	meta := datadogMetaFromCarrier(carrier)

	return &upstreamContext{
		TraceID:  traceID,
		ParentID: parentID,
		Meta:     meta,
	}
}

func datadogMetaFromCarrier(carrier map[string]string) map[string]string {
	meta := map[string]string{}
	if samplingPriority := headerValue(carrier, "x-datadog-sampling-priority"); samplingPriority != "" {
		meta["_sampling_priority_v1"] = samplingPriority
	}
	if origin := headerValue(carrier, "x-datadog-origin"); origin != "" {
		meta["_dd.origin"] = origin
	}
	for k, v := range parseDatadogTags(headerValue(carrier, "x-datadog-tags")) {
		meta[k] = v
	}
	return meta
}

func extractW3CContextFromCarrier(carrier map[string]string) *w3cContext {
	traceparent := headerValue(carrier, "traceparent")
	if traceparent == "" {
		return nil
	}

	w3c := parseTraceparent(traceparent)
	if w3c == nil || w3c.TraceIDLow == 0 {
		return nil
	}
	for k, v := range parseTracestateDatadog(headerValue(carrier, "tracestate")) {
		w3c.Meta[k] = v
	}
	for k, v := range datadogMetaFromCarrier(carrier) {
		w3c.Meta[k] = v
	}
	if w3c.TraceIDHigh != "" {
		if _, ok := w3c.Meta["_dd.p.tid"]; !ok {
			w3c.Meta["_dd.p.tid"] = w3c.TraceIDHigh
		}
	}
	return w3c
}

func extractB3ContextFromCarrier(carrier map[string]string) *w3cContext {
	traceID := headerValue(carrier, "x-b3-traceid")
	parentIDHex := headerValue(carrier, "x-b3-spanid")
	sampled := headerValue(carrier, "x-b3-sampled")
	if traceID == "" || parentIDHex == "" {
		traceID, parentIDHex, sampled = parseB3Single(headerValue(carrier, "b3"))
	}
	if traceID == "" || parentIDHex == "" {
		return nil
	}

	w3c := parseHexTraceContext(traceID, parentIDHex)
	if w3c == nil || w3c.TraceIDLow == 0 {
		return nil
	}
	if sampled != "" {
		switch strings.ToLower(sampled) {
		case "1", "true", "d":
			w3c.Meta["_sampling_priority_v1"] = "1"
		case "0", "false":
			w3c.Meta["_sampling_priority_v1"] = "0"
		}
	}
	ensureTraceIDHighMeta(w3c)
	return w3c
}

func extractXRayContextFromCarrier(carrier map[string]string) *w3cContext {
	value := headerValue(carrier, "x-amzn-trace-id")
	if value == "" {
		return nil
	}
	parts := map[string]string{}
	for _, item := range strings.Split(value, ";") {
		key, val, ok := strings.Cut(strings.TrimSpace(item), "=")
		if ok && key != "" {
			parts[key] = val
		}
	}
	root := parts["Root"]
	parent := parts["Parent"]
	if root == "" || parent == "" {
		return nil
	}
	rootParts := strings.Split(root, "-")
	if len(rootParts) != 3 {
		return nil
	}

	w3c := parseHexTraceContext(rootParts[1]+rootParts[2], parent)
	if w3c == nil || w3c.TraceIDLow == 0 {
		return nil
	}
	switch parts["Sampled"] {
	case "1":
		w3c.Meta["_sampling_priority_v1"] = "1"
	case "0":
		w3c.Meta["_sampling_priority_v1"] = "0"
	}
	ensureTraceIDHighMeta(w3c)
	return w3c
}

func upstreamCarriers(payload []byte, headers map[string]string) []map[string]string {
	carriers := make([]map[string]string, 0, 4)
	carriers = append(carriers, datadogCarriersFromPayload(payload)...)
	if len(headers) > 0 {
		carriers = append(carriers, headers)
	}
	return carriers
}

func datadogCarriersFromPayload(payload []byte) []map[string]string {
	if len(payload) == 0 {
		return nil
	}
	var event map[string]any
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil
	}
	return carrierMapsFromAny(event)
}

func carrierMapsFromAny(value any) []map[string]string {
	return carrierMapsFromAnyDepth(value, 0)
}

func carrierMapsFromAnyDepth(value any, depth int) []map[string]string {
	if depth > maxCarrierScanDepth {
		return nil
	}
	input, ok := value.(map[string]any)
	if !ok {
		return nil
	}

	out := make([]map[string]string, 0, 2)
	if carrier := stringMapFromAny(input); hasTraceCarrier(carrier) {
		out = append(out, carrier)
	}

	for k, v := range input {
		switch strings.ToLower(k) {
		case "headers":
			if carrier := stringMapFromAny(v); len(carrier) > 0 {
				out = append(out, carrier)
			}
			if carrier := firstStringMapFromAny(v); len(carrier) > 0 {
				out = append(out, carrier)
			}
			if carrier := stringMapFromKafkaHeaders(v); len(carrier) > 0 {
				out = append(out, carrier)
			}
			if carrier := stringMapFromDynamoDBAttributeMap(v); len(carrier) > 0 {
				out = append(out, carrier)
			}
		case "multivalueheaders":
			if carrier := firstStringMapFromAny(v); len(carrier) > 0 {
				out = append(out, carrier)
			}
		case "messageattributes", "attributes":
			if carrier := stringMapFromMessageAttributes(v); len(carrier) > 0 {
				out = append(out, carrier)
			}
			if carrier := stringMapFromDynamoDBAttributeMap(v); len(carrier) > 0 {
				out = append(out, carrier)
			}
		default:
			out = append(out, carrierMapsFromAnyDepth(v, depth+1)...)
			if values, ok := v.([]any); ok {
				for _, item := range values {
					out = append(out, carrierMapsFromAnyDepth(item, depth+1)...)
				}
			}
			if s, ok := v.(string); ok {
				out = append(out, carrierMapsFromEncodedString(s, depth+1)...)
			}
		}
	}
	return out
}

func carrierMapsFromEncodedString(value string, depth int) []map[string]string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maxCarrierEncodedStringBytes || depth > maxCarrierScanDepth {
		return nil
	}
	if carriers := carrierMapsFromJSONString(value, depth); len(carriers) > 0 {
		return carriers
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(value)
	}
	if err != nil {
		return nil
	}
	return carrierMapsFromJSONString(string(decoded), depth)
}

func carrierMapsFromJSONString(value string, depth int) []map[string]string {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "{") && !strings.HasPrefix(value, "[") {
		return nil
	}
	var decoded any
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return nil
	}
	if carriers := carrierMapsFromAnyDepth(decoded, depth+1); len(carriers) > 0 {
		return carriers
	}
	values, ok := decoded.([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]string, 0, len(values))
	for _, item := range values {
		out = append(out, carrierMapsFromAnyDepth(item, depth+1)...)
	}
	return out
}

func hasTraceCarrier(headers map[string]string) bool {
	return headerValue(headers, "x-datadog-trace-id") != "" ||
		headerValue(headers, "traceparent") != "" ||
		headerValue(headers, "b3") != "" ||
		headerValue(headers, "x-b3-traceid") != "" ||
		headerValue(headers, "x-amzn-trace-id") != ""
}

func stringMapFromAny(value any) map[string]string {
	input, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	out := make(map[string]string, len(input))
	for k, v := range input {
		if s, ok := v.(string); ok && s != "" {
			out[k] = s
		}
	}
	return out
}

func firstStringMapFromAny(value any) map[string]string {
	input, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	out := make(map[string]string, len(input))
	for k, v := range input {
		values, ok := v.([]any)
		if !ok || len(values) == 0 {
			continue
		}
		if s, ok := values[0].(string); ok && s != "" {
			out[k] = s
		}
	}
	return out
}

func stringMapFromKafkaHeaders(value any) map[string]string {
	values, ok := value.([]any)
	if !ok {
		return nil
	}
	out := map[string]string{}
	for _, item := range values {
		header, ok := item.(map[string]any)
		if !ok {
			continue
		}
		for k, raw := range header {
			if s := byteArrayString(raw); s != "" {
				out[k] = s
			}
		}
	}
	return out
}

func stringMapFromMessageAttributes(value any) map[string]string {
	input, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	out := make(map[string]string, len(input))
	for k, raw := range input {
		switch attr := raw.(type) {
		case string:
			if attr != "" {
				out[k] = attr
			}
		case map[string]any:
			for _, field := range []string{"stringValue", "StringValue", "value", "Value", "S"} {
				if s, ok := attr[field].(string); ok && s != "" {
					out[k] = s
					break
				}
			}
		}
	}
	return out
}

func stringMapFromDynamoDBAttributeMap(value any) map[string]string {
	input, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	if nested, ok := input["M"].(map[string]any); ok {
		input = nested
	}
	out := make(map[string]string, len(input))
	for k, raw := range input {
		switch attr := raw.(type) {
		case string:
			if attr != "" {
				out[k] = attr
			}
		case map[string]any:
			if s, ok := attr["S"].(string); ok && s != "" {
				out[k] = s
			}
		}
	}
	return out
}

func byteArrayString(value any) string {
	values, ok := value.([]any)
	if !ok {
		return ""
	}
	bytes := make([]byte, 0, len(values))
	for _, item := range values {
		switch v := item.(type) {
		case float64:
			if v < 0 || v > 255 || v != float64(byte(v)) {
				return ""
			}
			bytes = append(bytes, byte(v))
		case int:
			if v < 0 || v > 255 {
				return ""
			}
			bytes = append(bytes, byte(v))
		}
	}
	return string(bytes)
}

func headerValue(headers map[string]string, key string) string {
	for k, v := range headers {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}

func parseDatadogTags(value string) map[string]string {
	out := map[string]string{}
	for _, item := range strings.Split(value, ",") {
		key, val, ok := strings.Cut(strings.TrimSpace(item), "=")
		if !ok || key == "" {
			continue
		}
		out[key] = val
	}
	return out
}

func parseTraceparent(value string) *w3cContext {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) != 4 || parts[0] == "" {
		return nil
	}
	return parseHexTraceContext(parts[1], parts[2])
}

func parseHexTraceContext(traceID, parentIDHex string) *w3cContext {
	traceID = strings.ToLower(strings.TrimSpace(traceID))
	parentIDHex = strings.ToLower(strings.TrimSpace(parentIDHex))
	if len(traceID) == 16 {
		traceID = strings.Repeat("0", 16) + traceID
	}
	if len(traceID) != 32 || len(parentIDHex) != 16 || traceID == strings.Repeat("0", 32) || parentIDHex == strings.Repeat("0", 16) {
		return nil
	}
	if _, ok := parseHexUint64(traceID[:16]); !ok {
		return nil
	}
	traceLow, ok := parseHexUint64(traceID[16:])
	if !ok {
		return nil
	}
	parentID, ok := parseHexUint64(parentIDHex)
	if !ok {
		return nil
	}
	return &w3cContext{
		TraceIDHigh: traceID[:16],
		TraceIDLow:  traceLow,
		ParentID:    parentID,
		Meta:        map[string]string{},
	}
}

func parseHexUint64(value string) (uint64, bool) {
	parsed, err := strconv.ParseUint(value, 16, 64)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func parseTracestateDatadog(value string) map[string]string {
	meta := map[string]string{}
	for _, entry := range strings.Split(value, ",") {
		key, val, ok := strings.Cut(strings.TrimSpace(entry), "=")
		if !ok || key != "dd" {
			continue
		}
		for _, item := range strings.Split(val, ";") {
			itemKey, itemVal, ok := strings.Cut(strings.TrimSpace(item), ":")
			if !ok || itemKey == "" {
				continue
			}
			switch itemKey {
			case "s":
				meta["_sampling_priority_v1"] = itemVal
			case "o":
				meta["_dd.origin"] = itemVal
			default:
				if strings.HasPrefix(itemKey, "t.") {
					meta["_dd.p."+strings.TrimPrefix(itemKey, "t.")] = itemVal
				}
			}
		}
	}
	return meta
}

func ensureTraceIDHighMeta(w3c *w3cContext) {
	if w3c == nil || w3c.TraceIDHigh == "" || w3c.TraceIDHigh == strings.Repeat("0", 16) {
		return
	}
	if w3c.Meta == nil {
		w3c.Meta = map[string]string{}
	}
	if _, ok := w3c.Meta["_dd.p.tid"]; !ok {
		w3c.Meta["_dd.p.tid"] = w3c.TraceIDHigh
	}
}

func parseB3Single(value string) (traceID, spanID, sampled string) {
	if value == "" || value == "0" || value == "1" || strings.EqualFold(value, "d") {
		return "", "", ""
	}
	parts := strings.Split(value, "-")
	if len(parts) < 2 {
		return "", "", ""
	}
	traceID = parts[0]
	spanID = parts[1]
	if len(parts) >= 3 {
		sampled = parts[2]
	}
	return traceID, spanID, sampled
}

func cloneMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}

func cloneFloatMap(input map[string]float64) map[string]float64 {
	if len(input) == 0 {
		return map[string]float64{}
	}
	out := make(map[string]float64, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}

func (p *Processor) findContextByTraceAndSpanLocked(traceID, invocationSpanID uint64) *InvocationContext {
	for _, ctx := range p.contexts {
		if ctx.TraceID == traceID && ctx.InvocationSpanID == invocationSpanID {
			return ctx
		}
	}
	return nil
}

func (p *Processor) mergePendingTracerLocked(ctx *InvocationContext) {
	if ctx == nil || ctx.TraceID == 0 || ctx.InvocationSpanID == 0 {
		return
	}
	key := tracerPlaceholderKey{traceID: ctx.TraceID, invocationSpanID: ctx.InvocationSpanID}
	if placeholder, ok := p.pendingTracer[key]; ok {
		mergeTracerPlaceholderLocked(ctx, placeholder)
		delete(p.pendingTracer, key)
	}
}

func mergeTracerPlaceholderLocked(ctx *InvocationContext, placeholder *tracerPlaceholder) {
	if ctx == nil || placeholder == nil {
		return
	}
	if ctx.TracerMeta == nil {
		ctx.TracerMeta = map[string]string{}
	}
	for k, v := range placeholder.meta {
		ctx.TracerMeta[k] = v
	}
	if ctx.TracerMetrics == nil {
		ctx.TracerMetrics = map[string]float64{}
	}
	for k, v := range placeholder.metrics {
		ctx.TracerMetrics[k] = v
	}
}

func cloneBytes(input []byte) []byte {
	if len(input) == 0 {
		return nil
	}
	out := make([]byte, len(input))
	copy(out, input)
	return out
}

func emptyFallback(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func requestIDOrUnknown(value string) string {
	if value == "" {
		return "unknown"
	}
	return value
}

func touchContextLocked(ctx *InvocationContext, now time.Time) {
	if ctx == nil {
		return
	}
	if now.IsZero() {
		now = time.Now()
	}
	if ctx.CreatedAt.IsZero() {
		ctx.CreatedAt = now
	}
	ctx.UpdatedAt = now
}

func (p *Processor) captureDeferredColdStartLocked(ctx *InvocationContext) {
	if ctx == nil || !ctx.ColdStart || ctx.ManagedInstance || ctx.InitDuration > 0 || ctx.RequestID == "" {
		return
	}
	if _, exists := p.deferredColdStart[ctx.RequestID]; exists {
		return
	}

	service, _ := lambdaInvocationServiceAndResource()
	start := ctx.StartTime
	if start.IsZero() {
		start = time.Now()
	}

	p.deferredColdStart[ctx.RequestID] = deferredColdStart{
		RequestID:           ctx.RequestID,
		TraceID:             ctx.TraceID,
		SpanID:              newID(),
		Service:             service,
		InvocationStartUnix: start.UnixNano(),
	}
}

func (p *Processor) cleanupLocked() {
	for len(p.contexts) > processorContextCapacity && len(p.contextOrder) > 0 {
		oldestID := p.contextOrder[0]
		p.contextOrder = p.contextOrder[1:]
		delete(p.contexts, oldestID)
	}
	if len(p.contextOrder) > 0 && len(p.contextOrder) > len(p.contexts)+processorContextCapacity {
		p.compactContextOrderLocked()
	}

	p.pendingInvokeIDs = cleanupPendingRequestIDs(p.pendingInvokeIDs)
	p.pendingRuntimeIDs = cleanupPendingRequestIDs(p.pendingRuntimeIDs)
	p.pendingStarts = cleanupPendingStarts(p.pendingStarts)
	p.pendingEnds = cleanupPendingEnds(p.pendingEnds)

	for key, placeholder := range p.pendingTracer {
		if placeholder == nil {
			delete(p.pendingTracer, key)
		}
	}
	for len(p.pendingTracer) > processorPendingTracerCapacity {
		oldestKey, ok := oldestPendingTracerKey(p.pendingTracer)
		if !ok {
			break
		}
		delete(p.pendingTracer, oldestKey)
	}
	for len(p.deferredColdStart) > processorContextCapacity {
		for requestID := range p.deferredColdStart {
			delete(p.deferredColdStart, requestID)
			break
		}
	}
}

func (p *Processor) compactContextOrderLocked() {
	out := p.contextOrder[:0]
	for _, id := range p.contextOrder {
		if _, ok := p.contexts[id]; ok {
			out = append(out, id)
		}
	}
	p.contextOrder = out
}

func cleanupPendingRequestIDs(input []pendingRequestID) []pendingRequestID {
	out := input
	if len(out) > processorPendingItemCapacity {
		out = out[len(out)-processorPendingItemCapacity:]
	}
	return out
}

func cleanupPendingStarts(input []*InvocationContext) []*InvocationContext {
	out := input[:0]
	for _, ctx := range input {
		if ctx == nil {
			continue
		}
		out = append(out, ctx)
	}
	if len(out) > processorPendingItemCapacity {
		out = out[len(out)-processorPendingItemCapacity:]
	}
	return out
}

func cleanupPendingEnds(input []pendingUniversalEnd) []pendingUniversalEnd {
	out := input
	if len(out) > processorPendingItemCapacity {
		out = out[len(out)-processorPendingItemCapacity:]
	}
	return out
}

func oldestPendingTracerKey(input map[tracerPlaceholderKey]*tracerPlaceholder) (tracerPlaceholderKey, bool) {
	var (
		oldestKey tracerPlaceholderKey
		oldest    time.Time
		ok        bool
	)
	for key, placeholder := range input {
		if placeholder == nil {
			return key, true
		}
		if !ok || placeholder.createdAt.Before(oldest) {
			oldestKey = key
			oldest = placeholder.createdAt
			ok = true
		}
	}
	return oldestKey, ok
}

func boolToError(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

func newID() uint64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return uint64(time.Now().UnixNano())
	}
	return binary.LittleEndian.Uint64(b[:])
}
