// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package util

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSQLStatementNormalizer(t *testing.T) {
	oracleNormalizer := NewSQLStatementNormalizer(SQLDatabaseOracle)
	first, err := oracleNormalizer.Normalize("SELECT * FROM users WHERE id = 1")
	require.NoError(t, err)
	second, err := oracleNormalizer.Normalize("SELECT * FROM users WHERE id = 2")
	require.NoError(t, err)
	assert.Equal(t, first, second)

	defaultNormalizer := NewSQLStatementNormalizer(SQLDatabase("unsupported"))
	defaultNormalized, err := defaultNormalizer.Normalize("SELECT * FROM users WHERE id = 3")
	require.NoError(t, err)
	assert.Equal(t, first, defaultNormalized)

	postgresqlNormalizer := NewSQLStatementNormalizer(SQLDatabasePostgreSQL)
	postgresqlNormalized, err := postgresqlNormalizer.Normalize("SELECT * FROM users WHERE id = 4")
	require.NoError(t, err)
	postgresqlSecond, err := postgresqlNormalizer.Normalize("SELECT * FROM users WHERE id = 5")
	require.NoError(t, err)
	assert.Equal(t, postgresqlNormalized, postgresqlSecond)

	mysqlNormalizer := NewSQLStatementNormalizer(SQLDatabaseMySQL)
	mysqlNormalized, err := mysqlNormalizer.Normalize("SELECT * FROM users WHERE id = 6")
	require.NoError(t, err)
	mysqlSecond, err := mysqlNormalizer.Normalize("SELECT * FROM users WHERE id = 7")
	require.NoError(t, err)
	assert.Equal(t, mysqlNormalized, mysqlSecond)

	sqlServerNormalizer := NewSQLStatementNormalizer(SQLDatabaseSQLServer)
	normalized, err := sqlServerNormalizer.Normalize(
		"SELECT u.id AS user_id FROM dbo.users AS u WHERE u.id = 42")
	require.NoError(t, err)
	assert.Contains(t, normalized.Text, "AS user_id")
	assert.Contains(t, normalized.Text, "AS u")
}

func TestTruncateUTF8ByBytes(t *testing.T) {
	const maxBytes = 512
	input := strings.Repeat("中", maxBytes) + "结尾"
	result, truncated := TruncateUTF8ByBytes(input, maxBytes)

	assert.True(t, truncated)
	assert.True(t, utf8.ValidString(result))
	assert.LessOrEqual(t, len(result), maxBytes)
	assert.Equal(t, strings.Repeat("中", maxBytes/len("中")), result)

	invalidUTF8 := string([]byte{'S', 'Q', 'L', 0xff, 'X'})
	validPrefix, invalidTruncated := TruncateUTF8ByBytes(invalidUTF8, 4)
	assert.True(t, invalidTruncated)
	assert.True(t, utf8.ValidString(validPrefix))

	exact, exactTruncated := TruncateUTF8ByBytes("1234", 4)
	assert.Equal(t, "1234", exact)
	assert.False(t, exactTruncated)
}

func TestComputeNormalizedSQLHash(t *testing.T) {
	assert.Empty(t, ComputeNormalizedSQLHash(""))
	assert.Equal(t,
		ComputeNormalizedSQLHash("SELECT * FROM users WHERE id = ?"),
		ComputeNormalizedSQLHash("SELECT * FROM users WHERE id = ?"),
	)
	assert.NotEqual(t,
		ComputeNormalizedSQLHash("SELECT * FROM users WHERE id = ?"),
		ComputeNormalizedSQLHash("SELECT * FROM orders WHERE id = ?"),
	)
}
