// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestHasSampledSinceCompletion(t *testing.T) {
	tests := []struct {
		name             string
		row              eventRow
		eventTimestampMs int64
		windowMs         int64
		want             bool
	}{
		{
			name: "invalid end_event_id should not skip",
			row: eventRow{
				endEventID: sql.NullInt64{Int64: 0, Valid: false},
			},
			eventTimestampMs: 1_000_000,
			windowMs:         60_000,
			want:             false,
		},
		{
			name: "cannot calculate end time should not skip",
			row: eventRow{
				endEventID: sql.NullInt64{Int64: 1, Valid: true},
				now:        sql.NullInt64{Int64: 1_000, Valid: true},
				// missing uptime/timerEnd makes calculateTimerEndMs return false
			},
			eventTimestampMs: 1_000_000,
			windowMs:         60_000,
			want:             false,
		},
		{
			name: "within window should not skip",
			row: eventRow{
				endEventID: sql.NullInt64{Int64: 1, Valid: true},
				now:        sql.NullInt64{Int64: 1_000, Valid: true},
				uptime:     sql.NullString{String: "100", Valid: true},
				// queryEndMs = (1000 - 100 + 50) * 1000 = 950000
				timerEnd: sql.NullInt64{Int64: 50_000_000_000_000, Valid: true}, // 50s in picoseconds
			},
			eventTimestampMs: 980_000,
			windowMs:         40_000,
			want:             false,
		},
		{
			name: "past window should skip",
			row: eventRow{
				endEventID: sql.NullInt64{Int64: 1, Valid: true},
				now:        sql.NullInt64{Int64: 1_000, Valid: true},
				uptime:     sql.NullString{String: "100", Valid: true},
				// queryEndMs = 950000
				timerEnd: sql.NullInt64{Int64: 50_000_000_000_000, Valid: true},
			},
			eventTimestampMs: 1_020_001,
			windowMs:         70_000,
			want:             true,
		},
		{
			name: "absolute diff applies when event time earlier",
			row: eventRow{
				endEventID: sql.NullInt64{Int64: 1, Valid: true},
				now:        sql.NullInt64{Int64: 1_000, Valid: true},
				uptime:     sql.NullString{String: "100", Valid: true},
				// queryEndMs = 950000
				timerEnd: sql.NullInt64{Int64: 50_000_000_000_000, Valid: true},
			},
			eventTimestampMs: 870_000,
			windowMs:         70_000,
			want:             true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := hasSampledSinceCompletion(tc.row, tc.eventTimestampMs, tc.windowMs)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestComputePlanSignature(t *testing.T) {
	sig1 := ComputeSQLPlanSignature(`{"plan":"a"}`)
	sig2 := ComputeSQLPlanSignature(`{"plan":"a"}`)
	sig3 := ComputeSQLPlanSignature(`{"plan":"b"}`)

	assert.Equal(t, sig1, sig2)
	assert.NotEqual(t, sig1, sig3)
	assert.Len(t, sig1, 16) // %016x
}

func TestGeneratePlanCacheKey(t *testing.T) {
	key1 := generatePlanCacheKey("query-1", "plan-1")
	key2 := generatePlanCacheKey("query-1", "plan-1")
	key3 := generatePlanCacheKey("query-2", "plan-1")

	assert.Equal(t, key1, key2)
	assert.NotEqual(t, key1, key3)
	assert.Len(t, key1, 16)
}

func TestMySQLPlanQuerySignatureMatchesMetricPath(t *testing.T) {
	o := obfuscate.NewObfuscator(obfuscate.Config{})
	digestText := "SELECT * FROM orders WHERE id = 42"
	planDigest := digestText
	if obfResult, err := o.ObfuscateSQLString(digestText); err == nil {
		planDigest = obfResult.Query
	}
	metricDigest := digestText
	if obfResult, err := o.ObfuscateSQLString(digestText); err == nil {
		metricDigest = obfResult.Query
	}

	require.NotEmpty(t, planDigest)
	assert.Equal(t, metricDigest, planDigest)
	assert.Equal(t, generateQuerySignature("app", metricDigest), generateQuerySignature("app", planDigest))
}

func TestCalculateTimerEndMs(t *testing.T) {
	t.Run("valid row should return calculated ms", func(t *testing.T) {
		row := eventRow{
			now:      sql.NullInt64{Int64: 1_000, Valid: true},
			uptime:   sql.NullString{String: "100", Valid: true},
			timerEnd: sql.NullInt64{Int64: 50_000_000_000_000, Valid: true}, // 50s
		}
		got, ok := calculateTimerEndMs(row)
		assert.True(t, ok)
		assert.Equal(t, int64(950_000), got)
	})

	t.Run("invalid row should return false", func(t *testing.T) {
		row := eventRow{
			now:      sql.NullInt64{Int64: 1_000, Valid: true},
			uptime:   sql.NullString{String: "", Valid: false},
			timerEnd: sql.NullInt64{Int64: 1, Valid: true},
		}
		got, ok := calculateTimerEndMs(row)
		assert.False(t, ok)
		assert.Equal(t, int64(0), got)
	})
}

func TestIsTruncated(t *testing.T) {
	assert.True(t, isTruncated("SELECT * FROM t ..."))
	assert.False(t, isTruncated("SELECT * FROM t"))
}

func TestCheckLimitRate(t *testing.T) {
	ipt := &Input{
		DbmSample: dbmSample{
			ExplainCacheTTL: datakit.Duration{Duration: time.Minute},
		},
	}

	assert.True(t, checkLimitRate(ipt, "k1"))
	assert.False(t, checkLimitRate(ipt, "k1")) // same key should be rate-limited
	assert.True(t, checkLimitRate(ipt, "k2"))
}

func TestCheckPlanSignatureRate(t *testing.T) {
	ipt := &Input{
		DbmSample: dbmSample{
			PlanCacheTTL: datakit.Duration{Duration: time.Minute},
		},
	}

	assert.False(t, checkPlanSignatureRate(ipt, ""))

	assert.True(t, checkPlanSignatureRate(ipt, "p1"))
	ipt.recordReportedPlanSignature("p1")
	assert.False(t, checkPlanSignatureRate(ipt, "p1"))

	assert.True(t, checkPlanSignatureRate(ipt, "p2"))
}
