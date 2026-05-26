// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
)

func TestGenerateQuerySignature(t *testing.T) {
	tests := []struct {
		name      string
		pdbName   string
		queryHash string
		validate  func(t *testing.T, signature string)
	}{
		{
			name:      "with pdb_name and query_hash",
			pdbName:   "TESTPDB",
			queryHash: "1234567890",
			validate: func(t *testing.T, signature string) {
				t.Helper()
				assert.NotEmpty(t, signature)
				assert.Len(t, signature, 16) // xxhash produces 16 hex characters
			},
		},
		{
			name:      "empty pdb_name",
			pdbName:   "",
			queryHash: "1234567890",
			validate: func(t *testing.T, signature string) {
				t.Helper()
				assert.NotEmpty(t, signature)
				assert.Len(t, signature, 16)
			},
		},
		{
			name:      "empty query_hash",
			pdbName:   "TESTPDB",
			queryHash: "",
			validate: func(t *testing.T, signature string) {
				t.Helper()
				assert.NotEmpty(t, signature)
				assert.Len(t, signature, 16)
			},
		},
		{
			name:      "both empty",
			pdbName:   "",
			queryHash: "",
			validate: func(t *testing.T, signature string) {
				t.Helper()
				assert.NotEmpty(t, signature)
				assert.Len(t, signature, 16)
			},
		},
		{
			name:      "same inputs produce same signature",
			pdbName:   "TESTPDB",
			queryHash: "1234567890",
			validate: func(t *testing.T, signature string) {
				t.Helper()
				sig1 := generateQuerySignature("TESTPDB", "1234567890")
				sig2 := generateQuerySignature("TESTPDB", "1234567890")
				assert.Equal(t, sig1, sig2)
			},
		},
		{
			name:      "different inputs produce different signatures",
			pdbName:   "TESTPDB",
			queryHash: "1234567890",
			validate: func(t *testing.T, signature string) {
				t.Helper()
				sig1 := generateQuerySignature("TESTPDB", "1234567890")
				sig2 := generateQuerySignature("TESTPDB2", "1234567890")
				assert.NotEqual(t, sig1, sig2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature := generateQuerySignature(tt.pdbName, tt.queryHash)
			if tt.validate != nil {
				tt.validate(t, signature)
			}
		})
	}
}

func TestGetPdbName(t *testing.T) {
	tests := []struct {
		name     string
		ipt      *Input
		pdbName  string
		expected string
	}{
		{
			name: "non-empty pdb_name",
			ipt: &Input{
				cdbName: "TESTCDB",
			},
			pdbName:  "TESTPDB",
			expected: "TESTPDB",
		},
		{
			name: "empty pdb_name with multitenant",
			ipt: &Input{
				cdbName:       "TESTCDB",
				isMultitenant: true,
			},
			pdbName:  "",
			expected: "CDB$ROOT",
		},
		{
			name: "empty pdb_name without multitenant",
			ipt: &Input{
				cdbName:       "TESTCDB",
				isMultitenant: false,
			},
			pdbName:  "",
			expected: "TESTCDB",
		},
		{
			name: "empty pdb_name and empty cdbName",
			ipt: &Input{
				cdbName:       "",
				isMultitenant: false,
			},
			pdbName:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ipt.getPdbName(tt.pdbName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildStatementPoints(t *testing.T) {
	ipt := &Input{
		Object: oracleObject{
			name: "testhost:1521",
		},
		cdbName: "TESTCDB",
	}

	tests := []struct {
		name     string
		rows     []*OracleRow
		wantLen  int
		validate func(t *testing.T, pts []*point.Point)
	}{
		{
			name:    "empty rows",
			rows:    []*OracleRow{},
			wantLen: 0,
		},
		{
			name: "single row with all fields",
			rows: []*OracleRow{
				{
					querySignature: "abc123def456",
					RawData: StatementMetricsDB{
						StatementMetricsKeyDB: StatementMetricsKeyDB{
							ConID:                  1,
							PDBName:                "TESTPDB",
							ForceMatchingSignature: "1234567890",
							PlanHashValue:          987654321,
						},
						StatementMetricsMonotonicCountDB: StatementMetricsMonotonicCountDB{
							Executions:    100,
							ElapsedTime:   5000000,
							CPUTime:       3000000,
							BufferGets:    2000,
							RowsProcessed: 500,
						},
						StatementMetricsGaugeDB: StatementMetricsGaugeDB{
							VersionCount: 1,
							SharableMem:  10000,
							TypecheckMem: 5000,
						},
					},
					DeltaData: OracleRowMonotonicCount{
						Executions:    10,
						ElapsedTime:   500000,
						CPUTime:       300000,
						BufferGets:    200,
						RowsProcessed: 50,
					},
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, pts []*point.Point) {
				t.Helper()
				assert.Len(t, pts, 1)
				pt := pts[0]
				tags := pt.Tags()
				fields := pt.Fields()

				assert.Equal(t, "abc123def456", tags.GetTag("query_signature"))
				assert.Equal(t, "1", tags.GetTag("con_id"))
				assert.Equal(t, "TESTCDB", tags.GetTag("cdb_name"))
				assert.Equal(t, "TESTPDB", tags.GetTag("pdb_name"))
				assert.Equal(t, "1234567890", tags.GetTag("force_matching_signature"))
				assert.Equal(t, "987654321", tags.GetTag("plan_hash_value"))

				// Check total fields
				if totalExec := fields.Get("total_executions"); totalExec != nil {
					assert.Equal(t, int64(100), totalExec.Raw())
				}
				if totalElapsed := fields.Get("total_elapsed_time"); totalElapsed != nil {
					assert.Equal(t, int64(5000000), totalElapsed.Raw())
				}

				// Check delta fields
				if exec := fields.Get("executions"); exec != nil {
					assert.Equal(t, int64(10), exec.Raw())
				}
				if elapsed := fields.Get("elapsed_time"); elapsed != nil {
					assert.Equal(t, int64(500000), elapsed.Raw())
				}
			},
		},
		{
			name: "row with zero force_matching_signature",
			rows: []*OracleRow{
				{
					querySignature: "abc123def456",
					RawData: StatementMetricsDB{
						StatementMetricsKeyDB: StatementMetricsKeyDB{
							ForceMatchingSignature: "",
							PlanHashValue:          987654321,
						},
					},
					DeltaData: OracleRowMonotonicCount{},
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, pts []*point.Point) {
				t.Helper()
				assert.Len(t, pts, 1)
				pt := pts[0]
				tags := pt.Tags()
				assert.Equal(t, "0", tags.GetTag("force_matching_signature"))
			},
		},
		{
			name: "row with avg_elapsed_time calculation",
			rows: []*OracleRow{
				{
					querySignature: "abc123def456",
					RawData: StatementMetricsDB{
						StatementMetricsKeyDB: StatementMetricsKeyDB{
							PlanHashValue: 987654321,
						},
					},
					DeltaData: OracleRowMonotonicCount{
						Executions:  10,
						ElapsedTime: 500000,
					},
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, pts []*point.Point) {
				t.Helper()
				assert.Len(t, pts, 1)
				pt := pts[0]
				fields := pt.Fields()
				if avgElapsed := fields.Get("avg_elapsed_time"); avgElapsed != nil {
					assert.Equal(t, int64(50000), avgElapsed.Raw()) // 500000 / 10
				}
			},
		},
		{
			name: "row with zero executions (no avg_elapsed_time)",
			rows: []*OracleRow{
				{
					querySignature: "abc123def456",
					RawData: StatementMetricsDB{
						StatementMetricsKeyDB: StatementMetricsKeyDB{
							PlanHashValue: 987654321,
						},
					},
					DeltaData: OracleRowMonotonicCount{
						Executions:  0,
						ElapsedTime: 500000,
					},
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, pts []*point.Point) {
				t.Helper()
				assert.Len(t, pts, 1)
				pt := pts[0]
				fields := pt.Fields()
				// avg_elapsed_time should not be set when executions is 0
				assert.Nil(t, fields.Get("avg_elapsed_time"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptsTime := time.Now()
			pts := ipt.buildStatementPoints(tt.rows, ptsTime)
			assert.Len(t, pts, tt.wantLen)
			if tt.validate != nil {
				tt.validate(t, pts)
			}
		})
	}
}

func TestIsAllDeltasZero(t *testing.T) {
	tests := []struct {
		name     string
		diff     OracleRowMonotonicCount
		expected bool
	}{
		{
			name:     "all zeros",
			diff:     OracleRowMonotonicCount{},
			expected: true,
		},
		{
			name: "non-zero executions",
			diff: OracleRowMonotonicCount{
				Executions: 10,
			},
			expected: false,
		},
		{
			name: "non-zero elapsed_time",
			diff: OracleRowMonotonicCount{
				ElapsedTime: 1000,
			},
			expected: false,
		},
		{
			name: "non-zero cpu_time",
			diff: OracleRowMonotonicCount{
				CPUTime: 500,
			},
			expected: false,
		},
		{
			name: "non-zero buffer_gets",
			diff: OracleRowMonotonicCount{
				BufferGets: 100,
			},
			expected: false,
		},
		{
			name: "multiple non-zero values",
			diff: OracleRowMonotonicCount{
				Executions:  5,
				ElapsedTime: 1000,
				CPUTime:     500,
				BufferGets:  50,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAllDeltasZero(tt.diff)
			assert.Equal(t, tt.expected, result)
		})
	}
}
