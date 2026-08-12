// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/util"
)

func TestNormalizePostgreSQLStatement(t *testing.T) {
	normalizer := util.NewSQLStatementNormalizer(util.SQLDatabasePostgreSQL)

	first, err := normalizer.Normalize("SELECT * FROM users WHERE id = 1 AND status = 'active'")
	require.NoError(t, err)
	second, err := normalizer.Normalize("SELECT * FROM users WHERE id = 2 AND status = 'inactive'")
	require.NoError(t, err)

	assert.Equal(t, first.Text, second.Text)
	assert.Equal(t, first.Hash, second.Hash)
	assert.Len(t, first.Hash, 16)
}

func TestPostgreSQLMetricAndActivityUseSameNormalizedHash(t *testing.T) {
	normalizer := util.NewSQLStatementNormalizer(util.SQLDatabasePostgreSQL)

	metric, err := normalizer.Normalize("SELECT * FROM users WHERE id = $1 AND status = $2")
	require.NoError(t, err)
	activity, err := normalizer.Normalize("SELECT * FROM users WHERE id = 42 AND status = 'active'")
	require.NoError(t, err)

	assert.Equal(t, metric.Text, activity.Text)
	assert.Equal(t, metric.Hash, activity.Hash)
}

func TestPostgreSQLNormalizedQueryHashUsesCompleteStatement(t *testing.T) {
	normalizer := util.NewSQLStatementNormalizer(util.SQLDatabasePostgreSQL)
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

func TestPostgreSQLQueryTextLimit(t *testing.T) {
	queryText, truncated := util.TruncateUTF8ByBytes(
		strings.Repeat("中", maximumQueryTextMaxBytes), maximumQueryTextMaxBytes)

	assert.True(t, truncated)
	assert.LessOrEqual(t, len(queryText), maximumQueryTextMaxBytes)
	assert.True(t, utf8.ValidString(queryText))
}

func TestPostgreSQLDefaultQueryTextLimit(t *testing.T) {
	ipt := defaultInput()

	assert.Equal(t, defaultQueryTextMaxBytes, ipt.DbmMetric.QueryTextMaxBytes)
}
