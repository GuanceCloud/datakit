// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"database/sql"
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
			res[key] = v
			l.Debugf("get status %s:%v", key, v)
		} else {
			res[key] = raw
			l.Debugf("get status %s:%v", key, v)
		}
	}

	return res
}
