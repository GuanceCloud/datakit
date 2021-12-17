package mysql

import (
	"database/sql"
	"math"
	"strconv"
)

func binlogMetrics(r rows) map[string]interface{} {
	if r == nil {
		return nil
	}

	res := map[string]interface{}{}
	defer closeRows(r)

	var usage int64

	for r.Next() {
		var key string
		var val sql.RawBytes
		if n, err := r.Columns(); err != nil {
			l.Warnf("Columns(): %s, ignored", err.Error())
			continue
		} else {
			length := len(n)
			switch length {
			case 3:
				var encrypted string
				if err := r.Scan(&key, &val, &encrypted); err != nil {
					l.Warnf("Scan(): %s, ignored", err.Error())
					continue
				}
			default: // 2
				if err := r.Scan(&key, &val); err != nil {
					l.Warnf("Scan(): %s, ignored", err.Error())
					continue
				}
			}
		}

		raw := string(val)

		if v, err := strconv.ParseUint(raw, 10, 64); err == nil {
			if v > uint64(math.MaxInt64) {
				l.Warnf("%s exceed maxint64: %d > %d, ignored", key, v, math.MaxInt64)
				continue
			}

			usage += int64(v)
		} else {
			l.Warnf("invalid binlog usage: (%s: %s), ignored", key, raw)
		}
	}

	res["Binlog_space_usage_bytes"] = usage
	return res
}
