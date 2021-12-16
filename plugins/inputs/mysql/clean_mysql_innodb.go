package mysql

import (
	"database/sql"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func getCleanInnodb(r rows) (res map[string]interface{}) {
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

		value, err := Conv(string(val), inputs.Int)
		if err != nil {
			l.Errorf("innodb get value conv error", err)
		} else {
			res[key] = value
		}
	}

	return res
}
