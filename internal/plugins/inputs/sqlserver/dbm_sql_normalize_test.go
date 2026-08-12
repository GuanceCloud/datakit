// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/util"
)

func TestNormalizeSQLServerStatement(t *testing.T) {
	normalizer := util.NewSQLStatementNormalizer(util.SQLDatabaseSQLServer)

	first, err := normalizer.Normalize(
		"INSERT INTO [dbo].[users] (username, email, status) VALUES (@p1, @p2, @p3)")
	require.NoError(t, err)
	second, err := normalizer.Normalize(
		"INSERT INTO [dbo].[users] (username, email, status) VALUES (@name, @email, @status)")
	require.NoError(t, err)
	multiRow, err := normalizer.Normalize(
		"INSERT INTO [dbo].[users] (username, email, status) VALUES (@p1, @p2, @p3), (@p4, @p5, @p6)")
	require.NoError(t, err)

	// The legacy obfuscator preserves bind parameter names.
	assert.NotEqual(t, first.Text, second.Text)
	assert.NotEqual(t, first.Hash, second.Hash)
	assert.NotEqual(t, first.Hash, multiRow.Hash)
	assert.Contains(t, first.Text, "@p1")
	assert.NotContains(t, first.Text, "[")
	assert.Contains(t, first.Text, "dbo")
	assert.Contains(t, first.Text, "users")
}

func TestNormalizeSQLServerStatementDoesNotHashDatabaseContext(t *testing.T) {
	normalizer := util.NewSQLStatementNormalizer(util.SQLDatabaseSQLServer)

	first, err := normalizer.Normalize("SELECT * FROM users WHERE id = 1")
	require.NoError(t, err)
	second, err := normalizer.Normalize("SELECT * FROM users WHERE id = 2")
	require.NoError(t, err)

	assert.Equal(t, first.Text, second.Text)
	assert.Equal(t, first.Hash, second.Hash)
}

func TestNormalizeSQLServerStatementKeepsAliases(t *testing.T) {
	normalizer := util.NewSQLStatementNormalizer(util.SQLDatabaseSQLServer)

	normalized, err := normalizer.Normalize(
		"SELECT u.id AS user_id FROM dbo.users AS u WHERE u.id = 42")
	require.NoError(t, err)

	assert.Contains(t, normalized.Text, "AS user_id")
	assert.Contains(t, normalized.Text, "AS u")
}

func TestNormalizeSQLServerStatementSupportsCTE(t *testing.T) {
	normalizer := util.NewSQLStatementNormalizer(util.SQLDatabaseSQLServer)

	cte, err := normalizer.Normalize(`
		WITH active_users AS (
			SELECT id FROM [dbo].[users] WHERE status = 'active'
		)
		SELECT u.id FROM active_users AS u WHERE u.id = 42`)
	require.NoError(t, err)
	assert.Contains(t, cte.Text, "WITH active_users AS (")
	assert.NotContains(t, cte.Text, "'active'")
}

func TestNormalizeSQLServerStatementUsesConfiguredQueryTextBytes(t *testing.T) {
	normalizer := util.NewSQLStatementNormalizer(util.SQLDatabaseSQLServer)
	normalized, err := normalizer.Normalize(
		"SELECT user_id, username, email FROM dbo.users WHERE status = 'active'")
	require.NoError(t, err)
	queryText, truncated := util.TruncateUTF8ByBytes(normalized.Text, 16)

	assert.True(t, truncated)
	assert.Equal(t, 16, len(queryText))
	assert.Greater(t, len(normalized.Text), len(queryText))
}

func TestNormalizedQueryHashUsesCompleteStatement(t *testing.T) {
	normalizer := util.NewSQLStatementNormalizer(util.SQLDatabaseSQLServer)
	prefix := "SELECT " + strings.Repeat("column_name, ", 60)

	first, err := normalizer.Normalize(prefix + "final_a FROM dbo.users")
	require.NoError(t, err)
	second, err := normalizer.Normalize(prefix + "final_b FROM dbo.users")
	require.NoError(t, err)
	firstQueryText, firstTruncated := util.TruncateUTF8ByBytes(first.Text, defaultQueryTextMaxBytes)
	secondQueryText, secondTruncated := util.TruncateUTF8ByBytes(second.Text, defaultQueryTextMaxBytes)

	assert.True(t, firstTruncated)
	assert.True(t, secondTruncated)
	assert.Equal(t, firstQueryText, secondQueryText)
	assert.Equal(t, defaultQueryTextMaxBytes, len(firstQueryText))
	assert.NotEqual(t, first.Hash, second.Hash)
}

func TestPrepareNormalizedStatementsLeavesUnavailableSQLUnenriched(t *testing.T) {
	ipt := &Input{
		Dbm: &dbmConfig{Metric: &dbmMetricConfig{}},
	}
	emptySQL := &dbmStatementRow{queryHash: "0x01"}
	encryptedSQL := &dbmStatementRow{
		queryHash:     "0x02",
		statementText: "encrypted statement",
		isEncrypted:   true,
	}
	valid := &dbmStatementRow{
		queryHash:     "0x03",
		statementText: "SELECT * FROM users WHERE id = 1",
	}

	ipt.prepareNormalizedStatements([]*dbmStatementRow{nil, emptySQL, encryptedSQL, valid})

	assert.Empty(t, emptySQL.normalizedQueryHash)
	assert.Empty(t, emptySQL.queryText)
	assert.Empty(t, encryptedSQL.normalizedQueryHash)
	assert.Empty(t, encryptedSQL.queryText)
	require.NotEmpty(t, valid.normalizedQueryHash)
	require.NotEmpty(t, valid.queryText)
}

func TestBuildStatementPointsAddsNormalizedQueryTags(t *testing.T) {
	ipt := &Input{Tags: map[string]string{}}
	row := &dbmStatementRow{
		databaseName:        "testdb",
		queryHash:           "0x01",
		querySignature:      "legacy-signature",
		normalizedQueryHash: "normalized-hash",
		queryText:           "SELECT * FROM users WHERE id = ?",
	}

	points := ipt.buildStatementPoints([]*dbmStatementRow{row}, time.Now())
	require.Len(t, points, 1)

	tags := points[0].Tags()
	assert.Equal(t, "normalized-hash", tags.GetTag("normalized_query_hash"))
	assert.Equal(t, "SELECT * FROM users WHERE id = ?", tags.GetTag("query_text"))
	assert.Equal(t, "false", tags.GetTag("query_truncated"))
	assert.Contains(t, (&sqlserverMeasurement{}).Info().Tags, "query_truncated")
}

func TestBuildStatementPointsPreservesUTF8QueryTextAtMaximumBytes(t *testing.T) {
	queryText, truncated := util.TruncateUTF8ByBytes(
		strings.Repeat("中", maximumQueryTextMaxBytes), maximumQueryTextMaxBytes)
	ipt := &Input{Tags: map[string]string{}}
	row := &dbmStatementRow{
		databaseName:        "testdb",
		queryHash:           "0x01",
		querySignature:      "legacy-signature",
		normalizedQueryHash: "normalized-hash",
		queryText:           queryText,
		queryTextTruncated:  truncated,
	}

	points := ipt.buildStatementPoints([]*dbmStatementRow{row}, time.Now())
	require.Len(t, points, 1)

	actual := points[0].Tags().GetTag("query_text")
	assert.Equal(t, queryText, actual)
	assert.LessOrEqual(t, len(actual), maximumQueryTextMaxBytes)
	assert.True(t, utf8.ValidString(actual))
	assert.Equal(t, "true", points[0].Tags().GetTag("query_truncated"))
}

func TestBuildActivityPointsAddsNormalizedQueryHash(t *testing.T) {
	ipt := &Input{Tags: map[string]string{}}
	row := &dbmActivityRow{
		databaseName:        "testdb",
		normalizedQueryHash: "normalized-hash",
		obfuscatedText:      "SELECT * FROM users WHERE id = ?",
	}

	points := ipt.buildActivityPoints([]*dbmActivityRow{row}, time.Now())
	require.Len(t, points, 1)
	assert.Equal(t, "normalized-hash", points[0].Tags().GetTag("normalized_query_hash"))
}

func TestBuildPlanPointsAddsNormalizedQueryHash(t *testing.T) {
	ipt := &Input{Tags: map[string]string{}}
	row := &statementRowWithPlan{
		dbmStatementRow: &dbmStatementRow{
			querySignature:      "legacy-signature",
			queryPlanHash:       "plan-hash",
			normalizedQueryHash: "normalized-hash",
		},
	}

	points := ipt.buildAndFeedDatabasePlanObjects([]*statementRowWithPlan{row}, time.Now())
	require.Len(t, points, 1)
	assert.Equal(t, "normalized-hash", points[0].Tags().GetTag("normalized_query_hash"))
}
