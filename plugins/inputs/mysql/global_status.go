package mysql

import (
	"database/sql"
	"math"
	"strconv"
)

func globalStatusMetrics(r rows) map[string]interface{} {
	if r == nil {
		return nil
	}

	defer closeRows(r)

	res := map[string]interface{}{}

	for r.Next() {
		var key string
		var val sql.RawBytes

		if err := r.Scan(&key, &val); err != nil {
			l.Warnf("Scan: %s, ignored", err)
			continue
		}

		raw := string(val)
		if v, err := strconv.ParseUint(raw, 10, 64); err == nil {
			if v > uint64(math.MaxInt64) {
				l.Warnf("%s exceed maxint64: %d > %d, ignored", key, v, math.MaxInt64)
				continue
			}
			res[key] = int64(v)
		} else {
			res[key] = raw
		}
	}

	return res
}
