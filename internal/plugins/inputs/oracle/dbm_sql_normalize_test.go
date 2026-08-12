// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/util"
)

func TestNormalizeOracleStatement(t *testing.T) {
	normalizer := util.NewSQLStatementNormalizer(util.SQLDatabaseOracle)

	first, err := normalizer.Normalize("SELECT * FROM users WHERE id = 1 AND status = 'active'")
	require.NoError(t, err)
	second, err := normalizer.Normalize("SELECT * FROM users WHERE id = 2 AND status = 'inactive'")
	require.NoError(t, err)

	assert.Equal(t, first.Text, second.Text)
	assert.Equal(t, first.Hash, second.Hash)
	assert.Len(t, first.Hash, 16)
}

func TestGetFullSQLTextUsesCache(t *testing.T) {
	cache, err := lru.New[string, string](fullSQLTextCacheSize)
	require.NoError(t, err)
	cache.Add("abc123", "SELECT * FROM users")

	ipt := &Input{fullSQLTextCache: cache}
	var statement string
	require.NoError(t, ipt.getFullSQLText(context.Background(), &statement, "sql_id", "abc123"))
	assert.Equal(t, "SELECT * FROM users", statement)
}

func TestFullSQLTextCacheEvictsOldEntriesWhenWorkingSetExceedsCapacity(t *testing.T) {
	cache, err := lru.New[string, string](fullSQLTextCacheSize)
	require.NoError(t, err)
	for i := 0; i <= fullSQLTextCacheSize; i++ {
		cache.Add(fmt.Sprintf("sql-%d", i), fmt.Sprintf("SELECT %d FROM dual", i))
	}

	_, foundOldest := cache.Get("sql-0")
	newest, foundNewest := cache.Get(fmt.Sprintf("sql-%d", fullSQLTextCacheSize))
	assert.False(t, foundOldest)
	assert.True(t, foundNewest)
	assert.Equal(t, fmt.Sprintf("SELECT %d FROM dual", fullSQLTextCacheSize), newest)
}

func TestOracleNormalizedQueryHashUsesCompleteStatement(t *testing.T) {
	normalizer := util.NewSQLStatementNormalizer(util.SQLDatabaseOracle)
	prefix := "SELECT " + strings.Repeat("column_name, ", 60)

	first, err := normalizer.Normalize(prefix + "final_a FROM users")
	require.NoError(t, err)
	second, err := normalizer.Normalize(prefix + "final_b FROM users")
	require.NoError(t, err)
	firstQueryText, firstTruncated := util.TruncateUTF8ByBytes(first.Text, defaultQueryTextMaxBytes)
	secondQueryText, secondTruncated := util.TruncateUTF8ByBytes(second.Text, defaultQueryTextMaxBytes)

	assert.True(t, firstTruncated)
	assert.True(t, secondTruncated)
	assert.Equal(t, firstQueryText, secondQueryText)
	assert.NotEqual(t, first.Hash, second.Hash)
}

func TestOracleQueryTextLimit(t *testing.T) {
	queryText, truncated := util.TruncateUTF8ByBytes(
		strings.Repeat("中", defaultQueryTextMaxBytes), defaultQueryTextMaxBytes)
	assert.True(t, truncated)
	assert.LessOrEqual(t, len(queryText), defaultQueryTextMaxBytes)
	assert.True(t, utf8.ValidString(queryText))
}

func TestPrepareNormalizedStatementsUsesDefaultQueryTextLimitForInvalidConfig(t *testing.T) {
	statement := "SELECT " + strings.Repeat("column_name, ", 80) + "final_column FROM users"

	for _, queryTextMaxBytes := range []int{0, -1, maximumQueryTextMaxBytes + 1} {
		ipt := &Input{
			Dbm: &dbmConfig{Metric: &dbmMetricConfig{QueryTextMaxBytes: queryTextMaxBytes}},
		}
		row := &OracleRow{
			querySignature: "legacy-signature",
			RawData: StatementMetricsDB{
				SQLText:       statement,
				SQLTextLength: 999,
			},
		}

		ipt.prepareNormalizedStatements(context.Background(), []*OracleRow{row})

		assert.Len(t, row.queryText, defaultQueryTextMaxBytes)
		assert.True(t, row.queryTextTruncated)
	}
}

func TestPrepareNormalizedStatementsDoesNotReuseLegacySignature(t *testing.T) {
	ipt := &Input{
		Dbm: &dbmConfig{Metric: &dbmMetricConfig{QueryTextMaxBytes: 32}},
	}
	first := &OracleRow{
		querySignature: "legacy-signature",
		RawData: StatementMetricsDB{
			SQLText:       "SELECT * FROM users WHERE id = 1",
			SQLTextLength: 32,
		},
	}
	ipt.prepareNormalizedStatements(context.Background(), []*OracleRow{first})

	require.NotEmpty(t, first.normalizedQueryHash)
	assert.LessOrEqual(t, len(first.queryText), 32)

	second := &OracleRow{
		querySignature: "legacy-signature",
		RawData: StatementMetricsDB{
			SQLText:       "SELECT * FROM orders WHERE id = 2",
			SQLTextLength: 33,
		},
	}
	ipt.prepareNormalizedStatements(context.Background(), []*OracleRow{second})

	assert.NotEqual(t, first.normalizedQueryHash, second.normalizedQueryHash)
	assert.NotEqual(t, first.queryText, second.queryText)
}

func TestPrepareNormalizedStatementsLeavesUnavailableSQLUnenriched(t *testing.T) {
	ipt := &Input{
		Dbm: &dbmConfig{Metric: &dbmMetricConfig{}},
	}
	missingSignature := &OracleRow{
		RawData: StatementMetricsDB{SQLText: "SELECT 1", SQLTextLength: 8},
	}
	missingSQL := &OracleRow{querySignature: "missing-sql"}
	valid := &OracleRow{
		querySignature: "legacy-signature",
		RawData: StatementMetricsDB{
			SQLText:       "SELECT * FROM users WHERE id = 1",
			SQLTextLength: 32,
		},
	}

	ipt.prepareNormalizedStatements(context.Background(), []*OracleRow{
		nil,
		missingSignature,
		missingSQL,
		valid,
	})

	assert.Empty(t, missingSignature.normalizedQueryHash)
	assert.Empty(t, missingSignature.queryText)
	assert.Empty(t, missingSQL.normalizedQueryHash)
	assert.Empty(t, missingSQL.queryText)
	require.NotEmpty(t, valid.normalizedQueryHash)
}

func TestPrepareNormalizedStatementsFallsBackForColdCacheRows(t *testing.T) {
	contexts := map[string]func() context.Context{
		"canceled": func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		},
		"expired deadline": func() context.Context {
			ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
			t.Cleanup(cancel)
			return ctx
		},
	}

	for name, newContext := range contexts {
		t.Run(name, func(t *testing.T) {
			cache, err := lru.New[string, string](fullSQLTextCacheSize)
			require.NoError(t, err)
			ipt := &Input{
				Dbm:              &dbmConfig{Metric: &dbmMetricConfig{}},
				fullSQLTextCache: cache,
			}
			firstLongSQL := newLongOracleRow("long-query-1", strings.Repeat("x", int(MaxSQLFullTextVSQLStats)))
			secondLongSQL := newLongOracleRow("long-query-2", strings.Repeat("y", int(MaxSQLFullTextVSQLStats)))
			shortSQL := &OracleRow{
				querySignature: "short-signature",
				RawData: StatementMetricsDB{
					SQLText:       "SELECT * FROM users WHERE id = 42",
					SQLTextLength: 33,
				},
			}

			ipt.prepareNormalizedStatements(newContext(), []*OracleRow{firstLongSQL, secondLongSQL, shortSQL})

			assert.NotEmpty(t, firstLongSQL.normalizedQueryHash)
			assert.NotEmpty(t, secondLongSQL.normalizedQueryHash)
			assert.NotEqual(t, firstLongSQL.normalizedQueryHash, secondLongSQL.normalizedQueryHash)
			assert.NotEmpty(t, shortSQL.normalizedQueryHash)
			assert.Equal(t, 0, cache.Len())
		})
	}
}

func TestPrepareNormalizedStatementsUsesCachedFullSQLTextAfterCancellation(t *testing.T) {
	const sqlID = "cached-long-query"
	fullSQLText := "SELECT * FROM users WHERE id = 42 AND status = 'active'"
	cache, err := lru.New[string, string](fullSQLTextCacheSize)
	require.NoError(t, err)
	cache.Add(sqlID, fullSQLText)
	ipt := &Input{
		Dbm:              &dbmConfig{Metric: &dbmMetricConfig{}},
		fullSQLTextCache: cache,
	}
	row := newLongOracleRow(sqlID, strings.Repeat("x", int(MaxSQLFullTextVSQLStats)))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ipt.prepareNormalizedStatements(ctx, []*OracleRow{row})

	normalized, err := util.NewSQLStatementNormalizer(util.SQLDatabaseOracle).Normalize(fullSQLText)
	require.NoError(t, err)
	assert.Equal(t, normalized.Hash, row.normalizedQueryHash)
	assert.Equal(t, normalized.Text, row.normalizedText)
}

func newLongOracleRow(sqlID, sqlText string) *OracleRow {
	return &OracleRow{
		querySignature: "legacy-" + sqlID,
		RawData: StatementMetricsDB{
			StatementMetricsKeyDB: StatementMetricsKeyDB{SQLID: sqlID},
			SQLText:               sqlText,
			SQLTextLength:         MaxSQLFullTextVSQLStats,
		},
	}
}
