// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/cespare/xxhash/v2"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"
)

// generateQuerySignature generates a unique signature for a SQL statement.
// Similar to SQL Server's implementation, it uses schema and digest to generate the signature.
func generateQuerySignature(schemaName, digest string) string {
	h := xxhash.New()
	_, _ = h.WriteString(schemaName)
	_, _ = h.WriteString(":")
	_, _ = h.WriteString(digest)

	return fmt.Sprintf("%016x", h.Sum64())
}

func newMySQLSQLObfuscator() *obfuscate.Obfuscator {
	return obfuscate.NewObfuscator(obfuscate.Config{})
}

func getCleanSummaryRows(r rows) []dbmRow {
	if r == nil {
		return nil
	}

	defer closeRows(r)

	var (
		schemaName sql.NullString
		digest     sql.NullString
		digestText sql.NullString

		countStar          uint64
		sumTimerWait       uint64
		sumLockTime        uint64
		sumErrors          uint64
		sumRowsAffected    uint64
		sumRowsSent        uint64
		sumRowsExamined    uint64
		sumSelectScan      uint64
		sumSelectFullJoin  uint64
		sumNoIndexUsed     uint64
		sumNoGoodIndexUsed uint64
	)

	dbmRows := []dbmRow{}

	o := newMySQLSQLObfuscator()

	for r.Next() {
		if err := r.Scan(
			&schemaName,
			&digest,
			&digestText,
			&countStar,
			&sumTimerWait,
			&sumLockTime,
			&sumErrors,
			&sumRowsAffected,
			&sumRowsSent,
			&sumRowsExamined,
			&sumSelectScan,
			&sumSelectFullJoin,
			&sumNoIndexUsed,
			&sumNoGoodIndexUsed); err != nil {
			l.Errorf("scan dbm metric row failed: %s", err.Error())
			continue
		}

		var digestStr,
			digestTextStr,
			schemaNameStr string
		if digest.Valid {
			digestStr = digest.String
		}
		if digestText.Valid {
			digestTextStr = digestText.String
		}
		if schemaName.Valid {
			schemaNameStr = schemaName.String
		}

		// Skip rows with empty digest or digest text
		if digestStr == "" || digestTextStr == "" {
			continue
		}

		// Filter out rows that are EXPLAIN statements
		if strings.HasPrefix(strings.ToLower(digestTextStr), "explain") {
			continue
		}

		obfResult, err := o.ObfuscateSQLString(digestTextStr)
		if err != nil {
			l.Warnf("obfuscate digest text failed: %s,digestTextStr: %s", err.Error(), digestTextStr)
			continue
		}
		digestTextStr = obfResult.Query

		// Generate query signature from schema and obfuscated digest text (xxhash)
		querySignature := generateQuerySignature(schemaNameStr, digestTextStr)

		dbmRowItem := dbmRow{
			digest:             digestStr,
			digestText:         digestTextStr,
			schemaName:         schemaNameStr,
			querySignature:     querySignature,
			countStar:          countStar,
			sumTimerWait:       sumTimerWait,
			sumLockTime:        sumLockTime,
			sumErrors:          sumErrors,
			sumRowsAffected:    sumRowsAffected,
			sumRowsSent:        sumRowsSent,
			sumRowsExamined:    sumRowsExamined,
			sumSelectScan:      sumSelectScan,
			sumSelectFullJoin:  sumSelectFullJoin,
			sumNoIndexUsed:     sumNoIndexUsed,
			sumNoGoodIndexUsed: sumNoGoodIndexUsed,
		}
		dbmRows = append(dbmRows, dbmRowItem)
	}

	dbmRows = mergeDuplicateRows(dbmRows)

	return dbmRows
}
