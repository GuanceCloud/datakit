package mysql

import (
	"fmt"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type userMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// 生成行协议.
func (m *userMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

// 指定指标.
//nolint:lll
func (m *userMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Desc: "MySQL 用户指标",
		Name: "mysql_user_status",
		Fields: map[string]interface{}{
			// status
			"bytes_received": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of bytes received this user",
			},
			"bytes_sent": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of bytes sent this user",
			},
			"max_execution_time_exceeded": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of SELECT statements for which the execution timeout was exceeded.",
			},
			"max_execution_time_set": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of SELECT statements for which a nonzero execution timeout was set. This includes statements that include a nonzero MAX_EXECUTION_TIME optimizer hint, and statements that include no such hint but execute while the timeout indicated by the max_execution_time system variable is nonzero.",
			},
			"max_execution_time_set_failed": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of SELECT statements for which the attempt to set an execution timeout failed.",
			},
			"sort_rows": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of sorted rows.",
			},
			"sort_scan": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of sorts that were done by scanning the table.",
			},
			"table_open_cache_hits": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of hits for open tables cache lookups.",
			},
			"table_open_cache_misses": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of misses for open tables cache lookups.",
			},
			"table_open_cache_overflows": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of overflows for the open tables cache. This is the number of times, after a table is opened or closed, a cache instance has an unused entry and the size of the instance is larger than table_open_cache / table_open_cache_instances.",
			},
			"slow_queries": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of queries that have taken more than long_query_time seconds. This counter increments regardless of whether the slow query log is enabled",
			},
			"current_connect": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of current connect",
			},
			"total_connect": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "The number of total connect",
			},
		},
		Tags: map[string]interface{}{
			"user": &inputs.TagInfo{
				Desc: "user",
			},
		},
	}
}

func (i *Input) getUserData() ([]inputs.Measurement, error) {
	var resData []inputs.Measurement

	filterMap := map[string]bool{
		"mysql.session": true,
		"mysql.sys":     true,
	}

	userSQL := `select DISTINCT(user) from mysql.user`

	if len(i.Users) > 0 {
		var arr []string
		for _, user := range i.Users {
			arr = append(arr, fmt.Sprintf("'%s'", user))
		}

		filterStr := strings.Join(arr, ",")
		userSQL = fmt.Sprintf("%s where user in (%s);", userSQL, filterStr)
	}

	// run query
	rows, err := i.db.Query(userSQL)
	if err != nil {
		l.Errorf("query %v error %v", userSQL, err)
		return nil, err
	}
	defer func() {
		_ = rows.Close() //nolint:errcheck
		_ = rows.Err()   //nolint:errcheck
	}()

	for rows.Next() {
		var user string
		err = rows.Scan(
			&user,
		)

		if err != nil {
			return nil, err
		}

		if _, ok := filterMap[user]; ok {
			continue
		}

		ms, err := i.getUserStatus(user)
		if err != nil {
			continue
		}

		resData = append(resData, ms...)
	}

	return resData, nil
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

// 数据源获取数据.
func (i *Input) getUserStatus(user string) ([]inputs.Measurement, error) {
	var collectCache []inputs.Measurement

	userQuerySQL := `
	select VARIABLE_NAME, VARIABLE_VALUE
	from performance_schema.status_by_user
	where user='%s';
	`

	userConnSQL := `select USER, CURRENT_CONNECTIONS, TOTAL_CONNECTIONS
	from performance_schema.users
	where user = '%s';
    `

	m := &userMeasurement{
		tags:   make(map[string]string),
		fields: make(map[string]interface{}),
	}

	m.name = "mysql_user_status"

	for key, value := range i.Tags {
		m.tags[key] = value
	}

	m.tags["user"] = user

	// run query
	rows, err := i.db.Query(fmt.Sprintf(userQuerySQL, user))
	if err != nil {
		l.Errorf("query %v error %v", userQuerySQL, err)
		return nil, err
	}
	defer func() {
		_ = rows.Close() //nolint:errcheck
		_ = rows.Err()   //nolint:errcheck
	}()

	for rows.Next() {
		var (
			item  string
			value string
		)

		err = rows.Scan(
			&item,
			&value,
		)

		if err != nil {
			return nil, err
		}

		key := strings.ToLower(item)

		if _, ok := filterMetric[key]; ok {
			if m.fields[key], err = Conv(value, inputs.Int); err != nil {
				l.Warnf("convert '%s: %v' to int failed: %s, ignored", key, value, err.Error())
			}
		}
	}

	// two different 'rows'and 'rows.Close()' cannot exist in a method at the same time.
	rows1, err := i.db.Query(fmt.Sprintf(userConnSQL, user))
	if err != nil {
		l.Errorf("query %v error %v", userConnSQL, err)
		return nil, err
	}
	defer func() {
		_ = rows1.Close() //nolint:errcheck
		_ = rows1.Err()   //nolint:errcheck
	}()

	if err := rows1.Err(); err != nil {
		l.Errorf("rows.Err: %s", err)
		return nil, err
	}

	for rows1.Next() {
		var (
			curUser   string
			curConn   int
			totalConn int
		)

		err = rows1.Scan(
			&curUser,
			&curConn,
			&totalConn,
		)

		if err != nil {
			return nil, err
		}

		m.fields["current_connect"] = curConn
		m.fields["total_connect"] = totalConn
	}

	if len(m.fields) > 0 {
		collectCache = append(collectCache, m)
	}

	return collectCache, nil
}
