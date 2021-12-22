package mysql

import (
	"database/sql"
	"math"
	"strconv"
)

func globalVariablesMetrics(r rows) (res map[string]interface{}) {
	if r == nil {
		return
	}

	res = map[string]interface{}{}

	defer closeRows(r)

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
				l.Warnf("%s exceed maxint64: %d > %d, ignored", key, v, int64(math.MaxInt64))
				continue
			}
			res[key] = int64(v)
		} else {
			res[key] = raw
		}
	}

	return res
}
