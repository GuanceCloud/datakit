// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package newrelic handle New Relic APM traces.
package newrelic

import (
	"encoding/hex"
	"math/rand"
	"sync"
)

const (
	traceIDByteLen = 16
	// TraceIDHexStringLen is the length of the trace ID when represented
	// as a hex string.
	TraceIDHexStringLen = 32
	spanIDByteLen       = 8
	maxIDByteLen        = 16
)

// TraceIDGenerator creates identifiers for distributed tracing.
type TraceIDGenerator struct {
	sync.Mutex
	rnd *rand.Rand
}

// NewTraceIDGenerator creates a new trace identifier generator.
func NewTraceIDGenerator(seed int64) *TraceIDGenerator {
	return &TraceIDGenerator{
		rnd: rand.New(rand.NewSource(seed)), // nolint:gosec
	}
}

// Float32 returns a random float32 from its random source.
func (tg *TraceIDGenerator) Float32() float32 {
	tg.Lock()
	defer tg.Unlock()

	return tg.rnd.Float32()
}

// GenerateTraceID creates a new trace identifier, which is a 32 character hex string.
func (tg *TraceIDGenerator) GenerateTraceID() string {
	return tg.generateID(traceIDByteLen)
}

// GenerateSpanID creates a new span identifier, which is a 16 character hex string.
func (tg *TraceIDGenerator) GenerateSpanID() string {
	return tg.generateID(spanIDByteLen)
}

func (tg *TraceIDGenerator) generateID(l int) string {
	var bits [maxIDByteLen]byte
	tg.Lock()
	defer tg.Unlock()
	tg.rnd.Read(bits[:l]) //nolint:gosec
	return hex.EncodeToString(bits[:l])
}
