package mysql

import (
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var filterMap = map[string]bool{
	"mysql.session": true,
	"mysql.sys":     true,
}

// port from getUserData.
func getCleanUserStatusName(r rows) map[string]interface{} {
	if r == nil {
		return nil
	}

	defer closeRows(r)

	res := map[string]interface{}{}

	for r.Next() {
		var user string
		if err := r.Scan(&user); err != nil {
			l.Warnf("Scan: %s, ignored", err)
			continue
		}

		if _, ok := filterMap[user]; ok {
			continue
		}

		res[user] = true
	}

	return res
}

var filterMetric = map[string]bool{
	"bytes_received":                true,
	"bytes_sent":                    true,
	"max_execution_time_exceeded":   true,
	"max_execution_time_set":        true,
	"max_execution_time_set_failed": true,
	"sort_rows":                     true,
	"sort_scan":                     true,
	"table_open_cache_hits":         true,
	"table_open_cache_misses":       true,
	"table_open_cache_overflows":    true,
	"slow_queries":                  true,
}

func getCleanUserStatusVariable(r rows) map[string]interface{} {
	if r == nil {
		return nil
	}

	defer closeRows(r)

	res := make(map[string]interface{})

	for r.Next() {
		var item, value string

		if err := r.Scan(&item, &value); err != nil {
			l.Warnf("Scan: %s, ignored", err)
			continue
		}

		key := strings.ToLower(item)

		if _, ok := filterMetric[key]; ok {
			if result, err := Conv(value, inputs.Int); err != nil {
				l.Warnf("convert '%s: %v' to int failed: %s, ignored", key, value, err.Error())
			} else {
				res[key] = result
			}
		}
	}

	if len(res) == 0 {
		return nil
	}
	return res
}

func getCleanUserStatusConnection(r rows) map[string]interface{} {
	if r == nil {
		return nil
	}

	defer closeRows(r)

	res := make(map[string]interface{})

	for r.Next() {
		var curUser string
		var curConn, totalConn int64

		if err := r.Scan(&curUser, &curConn, &totalConn); err != nil {
			l.Warnf("Scan: %s, ignored", err)
			continue
		}

		res["current_connect"] = curConn
		res["total_connect"] = totalConn
	}

	if len(res) == 0 {
		return nil
	}
	return res
}
