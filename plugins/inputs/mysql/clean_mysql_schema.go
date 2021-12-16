package mysql

import (
	"database/sql"
)

// port from getSchemaSize.
func getCleanSchemaData(r rows) map[string]interface{} {
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

		res[key] = string(val)
	}

	return res
}
