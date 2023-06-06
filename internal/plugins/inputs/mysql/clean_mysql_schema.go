// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

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
