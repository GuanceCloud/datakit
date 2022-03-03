package mysql

import "database/sql"

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

		digestTextStr = obfuscateSQL(digestTextStr)

		querySignature := computeSQLSignature(digestTextStr)

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
