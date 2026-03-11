// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"
)

func TestGeneratePlanCacheKey(t *testing.T) {
	tests := []struct {
		name           string
		querySignature string
		planHashValue  string
		validate       func(t *testing.T, result string)
	}{
		{
			name:           "normal case",
			querySignature: "abc123def456",
			planHashValue:  "987654321",
			validate: func(t *testing.T, result string) {
				t.Helper()
				// generatePlanCacheKey returns a hash, not a string concatenation
				assert.NotEmpty(t, result)
				assert.Len(t, result, 16) // xxhash produces 16 hex characters
			},
		},
		{
			name:           "empty query_signature",
			querySignature: "",
			planHashValue:  "987654321",
			validate: func(t *testing.T, result string) {
				t.Helper()
				assert.NotEmpty(t, result)
				assert.Len(t, result, 16)
			},
		},
		{
			name:           "empty plan_hash_value",
			querySignature: "abc123def456",
			planHashValue:  "",
			validate: func(t *testing.T, result string) {
				t.Helper()
				assert.NotEmpty(t, result)
				assert.Len(t, result, 16)
			},
		},
		{
			name:           "both empty",
			querySignature: "",
			planHashValue:  "",
			validate: func(t *testing.T, result string) {
				t.Helper()
				assert.NotEmpty(t, result)
				assert.Len(t, result, 16)
			},
		},
		{
			name:           "same inputs produce same hash",
			querySignature: "abc123def456",
			planHashValue:  "987654321",
			validate: func(t *testing.T, result string) {
				t.Helper()
				// Same inputs should produce same hash
				result2 := generatePlanCacheKey("abc123def456", "987654321")
				assert.Equal(t, result, result2)
			},
		},
		{
			name:           "different inputs produce different hash",
			querySignature: "abc123def456",
			planHashValue:  "987654321",
			validate: func(t *testing.T, result string) {
				t.Helper()
				// Different inputs should produce different hash
				result2 := generatePlanCacheKey("abc123def456", "987654322")
				assert.NotEqual(t, result, result2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generatePlanCacheKey(tt.querySignature, tt.planHashValue)
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandlePredicate(t *testing.T) {
	obfuscator := obfuscate.NewObfuscator(obfuscate.Config{})

	tests := []struct {
		name          string
		predicate     sql.NullString
		predicateType string
		row           *OracleRow
		validate      func(t *testing.T, result string)
	}{
		{
			name: "valid predicate",
			predicate: sql.NullString{
				String: "id = 123",
				Valid:  true,
			},
			predicateType: "access",
			row: &OracleRow{
				RawData: StatementMetricsDB{
					StatementMetricsKeyDB: StatementMetricsKeyDB{
						SQLID: "abc123",
					},
				},
			},
			validate: func(t *testing.T, result string) {
				t.Helper()
				assert.NotEmpty(t, result)
				// Obfuscated result should not contain the literal value
				assert.NotContains(t, result, "123")
			},
		},
		{
			name: "NULL predicate",
			predicate: sql.NullString{
				Valid: false,
			},
			predicateType: "access",
			row: &OracleRow{
				RawData: StatementMetricsDB{
					StatementMetricsKeyDB: StatementMetricsKeyDB{
						SQLID: "abc123",
					},
				},
			},
			validate: func(t *testing.T, result string) {
				t.Helper()
				assert.Empty(t, result)
			},
		},
		{
			name: "empty predicate",
			predicate: sql.NullString{
				String: "",
				Valid:  true,
			},
			predicateType: "access",
			row: &OracleRow{
				RawData: StatementMetricsDB{
					StatementMetricsKeyDB: StatementMetricsKeyDB{
						SQLID: "abc123",
					},
				},
			},
			validate: func(t *testing.T, result string) {
				t.Helper()
				assert.Empty(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			handlePredicate(tt.predicateType, tt.predicate, &result, tt.row, obfuscator)
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestBuildAndFeedDatabasePlanObjects(t *testing.T) {
	ipt := &Input{
		Object: oracleObject{
			name: "testhost:1521",
		},
		cdbName: "TESTCDB",
	}

	tests := []struct {
		name     string
		rows     []*statementRowWithPlan
		wantLen  int
		validate func(t *testing.T, pts []*point.Point)
	}{
		{
			name:    "empty rows",
			rows:    []*statementRowWithPlan{},
			wantLen: 0,
		},
		{
			name: "single row with plan",
			rows: []*statementRowWithPlan{
				{
					OracleRow: &OracleRow{
						querySignature: "abc123def456",
						RawData: StatementMetricsDB{
							StatementMetricsKeyDB: StatementMetricsKeyDB{
								ConID:                  1,
								PDBName:                "TESTPDB",
								ForceMatchingSignature: "1234567890",
								SQLID:                  "abc123",
								PlanHashValue:          987654321,
							},
						},
					},
					planObfuscated: `[{"operation":"SELECT","id":1}]`,
					timestamp:      "2024-01-01T00:00:00Z",
					optimizerMode:  "ALL_ROWS",
					other:          "test",
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, pts []*point.Point) {
				t.Helper()
				assert.Len(t, pts, 1)
				pt := pts[0]
				tags := pt.Tags()
				fields := pt.Fields()

				planKey := generatePlanCacheKey("abc123def456", "987654321")
				assert.Equal(t, planKey, tags.GetTag("name"))
				assert.Equal(t, "abc123def456", tags.GetTag("query_signature"))
				assert.Equal(t, "987654321", tags.GetTag("plan_hash_value"))
				assert.Equal(t, "Oracle", tags.GetTag("database_type"))
				assert.Equal(t, "JSON", tags.GetTag("plan_type"))
				assert.Equal(t, "1", tags.GetTag("con_id"))
				assert.Equal(t, "TESTPDB", tags.GetTag("pdb_name"))
				assert.Equal(t, "TESTCDB", tags.GetTag("cdb_name"))
				assert.Equal(t, "1234567890", tags.GetTag("force_matching_signature"))
				assert.Equal(t, "abc123", tags.GetTag("sql_id"))

				if message := fields.Get("message"); message != nil {
					var planJSON []map[string]interface{}
					err := json.Unmarshal([]byte(message.GetS()), &planJSON)
					assert.NoError(t, err)
					assert.Len(t, planJSON, 1)
					assert.Equal(t, "SELECT", planJSON[0]["operation"])
				}
			},
		},
		{
			name: "row with zero con_id",
			rows: []*statementRowWithPlan{
				{
					OracleRow: &OracleRow{
						querySignature: "abc123def456",
						RawData: StatementMetricsDB{
							StatementMetricsKeyDB: StatementMetricsKeyDB{
								ConID:         0,
								PDBName:       "TESTPDB",
								PlanHashValue: 987654321,
							},
						},
					},
					planObfuscated: `[{"operation":"SELECT"}]`,
				},
			},
			wantLen: 1,
			validate: func(t *testing.T, pts []*point.Point) {
				t.Helper()
				assert.Len(t, pts, 1)
				pt := pts[0]
				tags := pt.Tags()
				// con_id should not be set when 0
				assert.Empty(t, tags.GetTag("con_id"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptsTime := time.Now()
			pts := ipt.buildAndFeedDatabasePlanObjects(tt.rows, ptsTime)
			assert.Len(t, pts, tt.wantLen)
			if tt.validate != nil {
				tt.validate(t, pts)
			}
		})
	}
}
