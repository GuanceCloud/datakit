// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package util

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/cespare/xxhash/v2"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"
)

type SQLDatabase string

const (
	SQLDatabaseMySQL      SQLDatabase = "mysql"
	SQLDatabaseOracle     SQLDatabase = "oracle"
	SQLDatabasePostgreSQL SQLDatabase = "postgresql"
	SQLDatabaseSQLServer  SQLDatabase = "sqlserver"
)

type NormalizedStatement struct {
	Text string
	Hash string
}

type SQLStatementNormalizer struct {
	obfuscator *obfuscate.Obfuscator
}

func NewSQLStatementNormalizer(database SQLDatabase) *SQLStatementNormalizer {
	switch database {
	case SQLDatabaseSQLServer:
		return &SQLStatementNormalizer{
			obfuscator: obfuscate.NewObfuscator(obfuscate.Config{
				SQL: obfuscate.SQLConfig{
					DBMS:         obfuscate.DBMSSQLServer,
					KeepSQLAlias: true,
				},
			}),
		}
	case SQLDatabasePostgreSQL:
		return &SQLStatementNormalizer{
			obfuscator: obfuscate.NewObfuscator(obfuscate.Config{
				SQL: obfuscate.SQLConfig{
					DBMS: obfuscate.DBMSPostgres,
				},
			}),
		}
	case SQLDatabaseMySQL:
		return &SQLStatementNormalizer{
			obfuscator: obfuscate.NewObfuscator(obfuscate.Config{}),
		}
	case SQLDatabaseOracle:
		return &SQLStatementNormalizer{
			obfuscator: obfuscate.NewObfuscator(obfuscate.Config{}),
		}
	default:
		return &SQLStatementNormalizer{
			obfuscator: obfuscate.NewObfuscator(obfuscate.Config{}),
		}
	}
}

func (normalizer *SQLStatementNormalizer) Normalize(statement string) (NormalizedStatement, error) {
	if statement == "" {
		return NormalizedStatement{}, nil
	}

	query, err := normalizer.obfuscator.ObfuscateSQLString(statement)
	if err != nil {
		return NormalizedStatement{}, err
	}

	return NormalizedStatement{
		Text: query.Query,
		Hash: ComputeNormalizedSQLHash(query.Query),
	}, nil
}

func ComputeNormalizedSQLHash(normalizedText string) string {
	if normalizedText == "" {
		return ""
	}

	return fmt.Sprintf("%016x", xxhash.Sum64String(normalizedText))
}

func TruncateUTF8ByBytes(value string, maxBytes int) (string, bool) {
	if maxBytes <= 0 {
		return "", value != ""
	}

	value = strings.ToValidUTF8(value, "\uFFFD")
	if len(value) <= maxBytes {
		return value, false
	}

	for maxBytes > 0 && !utf8.RuneStart(value[maxBytes]) {
		maxBytes--
	}

	return value[:maxBytes], true
}
