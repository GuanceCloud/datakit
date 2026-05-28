//go:build linux
// +build linux

package dnsflow

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDNSPendingQueryLimitEnv(t *testing.T) {
	t.Setenv(pendingQueryLimitEnv, "")
	assert.Equal(t, defaultPendingQueryLimit, dnsPendingQueryLimit())

	t.Setenv(pendingQueryLimitEnv, "2")
	assert.Equal(t, 2, dnsPendingQueryLimit())

	t.Setenv(pendingQueryLimitEnv, "0")
	assert.Equal(t, defaultPendingQueryLimit, dnsPendingQueryLimit())

	t.Setenv(pendingQueryLimitEnv, fmt.Sprintf("%d", maxPendingQueryLimit+1))
	assert.Equal(t, maxPendingQueryLimit, dnsPendingQueryLimit())
}

func TestDNSFlowTracerDropsQueryWhenPendingLimitReached(t *testing.T) {
	t.Setenv(pendingQueryLimitEnv, "1")
	tracer := NewDNSFlowTracer()
	now := time.Now()

	tracer.updateDNSStats(&DNSPacketInfo{
		Key:         DNSQAKey{TransactionID: 1, ClientPort: 53001, ServerPort: 53},
		TS:          now,
		QueryDomain: "first.example",
	}, nil)
	tracer.updateDNSStats(&DNSPacketInfo{
		Key:         DNSQAKey{TransactionID: 2, ClientPort: 53002, ServerPort: 53},
		TS:          now,
		QueryDomain: "second.example",
	}, nil)

	if len(tracer.statsMap) != 1 {
		t.Fatalf("pending query map len = %d, want 1", len(tracer.statsMap))
	}
	if _, ok := tracer.statsMap[DNSQAKey{TransactionID: 1, ClientPort: 53001, ServerPort: 53}]; !ok {
		t.Fatal("expected first query to remain when limit is reached")
	}
}

func TestDNSFlowTracerStoresQueryWithoutEmittingStats(t *testing.T) {
	tracer := NewDNSFlowTracer()

	stats := tracer.updateDNSStats(&DNSPacketInfo{
		Key:         DNSQAKey{TransactionID: 1, ClientPort: 53001, ServerPort: 53},
		TS:          time.Now(),
		QueryDomain: "query.example",
		QueryType:   "A",
	}, nil)

	if stats != nil {
		t.Fatalf("expected query to be stored for response matching without immediate stats, got %+v", *stats)
	}
	if len(tracer.statsMap) != 1 {
		t.Fatalf("pending query map len = %d, want 1", len(tracer.statsMap))
	}
}

func TestDNSFlowTracerUpdateHandlesNilAndZeroValue(t *testing.T) {
	var nilTracer *DNSFlowTracer
	if got := nilTracer.updateDNSStats(nil, nil); got != nil {
		t.Fatalf("nil tracer update returned %+v, want nil", got)
	}

	var tracer DNSFlowTracer
	stats := tracer.updateDNSStats(&DNSPacketInfo{
		Key:         DNSQAKey{TransactionID: 1, ClientPort: 53001, ServerPort: 53},
		TS:          time.Now(),
		QueryDomain: "query.example",
	}, nil)
	if stats != nil {
		t.Fatalf("query update returned %+v, want nil", *stats)
	}
	if len(tracer.statsMap) != 1 {
		t.Fatalf("zero-value tracer pending map len = %d, want 1", len(tracer.statsMap))
	}
}

func TestDNSFlowTracerRunHandlesNilInputs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		NewDNSFlowTracer().Run(ctx, nil, nil, nil)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not return for nil TPacket")
	}

	done = make(chan struct{})
	go func() {
		defer close(done)
		var tracer *DNSFlowTracer
		tracer.Run(ctx, nil, nil, nil)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not return for nil tracer")
	}
}

func TestDNSFlowTracerReadPacketHandlesNilInputs(t *testing.T) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		NewDNSFlowTracer().readPacket(context.Background(), nil)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("readPacket did not return for nil TPacket")
	}

	done = make(chan struct{})
	go func() {
		defer close(done)
		var tracer *DNSFlowTracer
		tracer.readPacket(context.Background(), nil)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("readPacket did not return for nil tracer")
	}
}

func TestDNSFlowTracerEmitsStatsOnResponse(t *testing.T) {
	tracer := NewDNSFlowTracer()
	key := DNSQAKey{TransactionID: 1, ClientPort: 53001, ServerPort: 53}
	start := time.Now()

	tracer.updateDNSStats(&DNSPacketInfo{
		Key:         key,
		TS:          start,
		QueryDomain: "query.example",
		QueryType:   "A",
	}, nil)
	stats := tracer.updateDNSStats(&DNSPacketInfo{
		Key:         key,
		QR:          true,
		RCODE:       0,
		TS:          start.Add(5 * time.Millisecond),
		QueryDomain: "query.example",
		QueryType:   "A",
	}, NewDNSRecord())

	if stats == nil {
		t.Fatal("expected response to emit DNS stats")
	}
	if stats.QueryDomain != "query.example" || stats.QueryType != "A" {
		t.Fatalf("unexpected DNS stats metadata: %+v", *stats)
	}
	if stats.RespTime != 5*time.Millisecond {
		t.Fatalf("unexpected response time: %s", stats.RespTime)
	}
	if len(tracer.statsMap) != 0 {
		t.Fatalf("expected matched query to be removed, got %d pending queries", len(tracer.statsMap))
	}
}

func TestDNSFlowTracerEmitsStatsOnResponseWithoutDNSRecord(t *testing.T) {
	tracer := NewDNSFlowTracer()
	key := DNSQAKey{TransactionID: 1, ClientPort: 53001, ServerPort: 53}
	start := time.Now()

	tracer.updateDNSStats(&DNSPacketInfo{
		Key:         key,
		TS:          start,
		QueryDomain: "query.example",
		QueryType:   "A",
	}, nil)
	stats := tracer.updateDNSStats(&DNSPacketInfo{
		Key:         key,
		QR:          true,
		RCODE:       0,
		TS:          start.Add(5 * time.Millisecond),
		QueryDomain: "query.example",
		QueryType:   "A",
	}, nil)

	if stats == nil {
		t.Fatal("expected response to emit DNS stats without DNS record cache")
	}
}

func TestDNSFlowAggAccumulatesLatencyWithoutSamples(t *testing.T) {
	agg := FlowAgg{}
	key := DNSQAKey{
		TransactionID: 1,
		IsUDP:         true,
		IsV4:          true,
		ClientPort:    53001,
		ServerPort:    53,
		ClientIP:      [4]uint32{0, 0, 0, 0x0100000A},
		ServerIP:      [4]uint32{0, 0, 0, 0x08080808},
	}

	if err := agg.Append(key, DNSStats{RCODE: 0, RespTime: time.Millisecond}); err != nil {
		t.Fatalf("append first dns stat: %v", err)
	}
	if err := agg.Append(key, DNSStats{RCODE: 0, RespTime: 3 * time.Millisecond}); err != nil {
		t.Fatalf("append second dns stat: %v", err)
	}

	if len(agg.data) != 1 {
		t.Fatalf("agg len = %d, want 1", len(agg.data))
	}
	for _, value := range agg.data {
		if value.count != 2 {
			t.Fatalf("count = %d, want 2", value.count)
		}
		if value.latencyTotal != int64(4*time.Millisecond) {
			t.Fatalf("latencyTotal = %d, want %d", value.latencyTotal, int64(4*time.Millisecond))
		}
		if value.latencyMax != int64(3*time.Millisecond) {
			t.Fatalf("latencyMax = %d, want %d", value.latencyMax, int64(3*time.Millisecond))
		}
	}
}

func TestDNSFlowAggLimitDropsNewKeysAndKeepsExisting(t *testing.T) {
	agg := FlowAgg{limit: 1}
	first := DNSQAKey{
		TransactionID: 1,
		IsUDP:         true,
		IsV4:          true,
		ClientPort:    53001,
		ServerPort:    53,
		ClientIP:      [4]uint32{0, 0, 0, 0x0100000A},
		ServerIP:      [4]uint32{0, 0, 0, 0x08080808},
	}
	second := first
	second.TransactionID = 2
	second.ClientPort = 53002

	if err := agg.Append(first, DNSStats{RCODE: 0, RespTime: time.Millisecond, QueryDomain: "first.example"}); err != nil {
		t.Fatalf("append first dns stat: %v", err)
	}
	if err := agg.Append(second, DNSStats{RCODE: 0, RespTime: 2 * time.Millisecond, QueryDomain: "second.example"}); err != nil {
		t.Fatalf("append second dns stat: %v", err)
	}
	if err := agg.Append(first, DNSStats{RCODE: 0, RespTime: 3 * time.Millisecond, QueryDomain: "first.example"}); err != nil {
		t.Fatalf("append first dns stat again: %v", err)
	}

	if got := agg.Len(); got != 1 {
		t.Fatalf("agg len = %d, want 1", got)
	}
	for _, value := range agg.data {
		if value.count != 2 {
			t.Fatalf("existing DNS agg key should continue aggregating at limit, got %+v", value)
		}
	}
}

func TestDNSAggLimitEnv(t *testing.T) {
	t.Setenv(dnsAggLimitEnv, "bad")
	if got := dnsAggLimit(); got != defaultDNSAggLimit {
		t.Fatalf("invalid agg limit = %d, want %d", got, defaultDNSAggLimit)
	}

	t.Setenv(dnsAggLimitEnv, fmt.Sprintf("%d", maxDNSAggLimit+1))
	if got := dnsAggLimit(); got != maxDNSAggLimit {
		t.Fatalf("clamped agg limit = %d, want %d", got, maxDNSAggLimit)
	}
}

func TestDNSFlowTracerClearsTimedOutQueriesBeforeApplyingLimit(t *testing.T) {
	t.Setenv(pendingQueryLimitEnv, "1")
	tracer := NewDNSFlowTracer()
	now := time.Now()
	expiredKey := DNSQAKey{TransactionID: 1, ClientPort: 53001, ServerPort: 53}
	newKey := DNSQAKey{TransactionID: 2, ClientPort: 53002, ServerPort: 53}

	tracer.updateDNSStats(&DNSPacketInfo{
		Key:         expiredKey,
		TS:          now.Add(-DNSTIMEOUT - time.Second),
		QueryDomain: "expired.example",
	}, nil)
	tracer.updateDNSStats(&DNSPacketInfo{
		Key:         newKey,
		TS:          now,
		QueryDomain: "new.example",
	}, nil)

	if len(tracer.statsMap) != 1 {
		t.Fatalf("pending query map len = %d, want 1", len(tracer.statsMap))
	}
	if _, ok := tracer.statsMap[expiredKey]; ok {
		t.Fatal("expected expired query to be evicted before enforcing limit")
	}
	if _, ok := tracer.statsMap[newKey]; !ok {
		t.Fatal("expected new query to be accepted after expired query eviction")
	}
}
