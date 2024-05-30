// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"database/sql"
	"math"
	"strconv"
)

func replicationMetrics(r rows) map[string]interface{} {
	if r == nil {
		return nil
	}

	res := map[string]interface{}{}

	defer closeRows(r)

	columns, err := r.Columns()
	if err != nil {
		l.Warnf("Error getting columns: %s", err)
		return nil
	}

	rawBytesColumns := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(columns))

	for i := range rawBytesColumns {
		scanArgs[i] = &rawBytesColumns[i]
	}

	for r.Next() {
		// scan for every column
		if err := r.Scan(scanArgs...); err != nil {
			l.Warnf("Scan: %s, ignored", err)
			continue
		}
		for i, key := range columns {
			raw := string(rawBytesColumns[i])
			if v, err := strconv.ParseUint(raw, 10, 64); err == nil {
				if v > uint64(math.MaxInt64) {
					l.Warnf("%s exceed maxint64: %d > %d, ignored", key, v, int64(math.MaxInt64))
					continue
				}
				res[key] = int64(v)
			} else {
				res[key] = raw
			}
		}
	}

	return res
}
