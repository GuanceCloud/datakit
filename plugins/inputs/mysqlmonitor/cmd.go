package mysqlmonitor

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	influxm "github.com/influxdata/influxdb1-client/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

// These are const but can't be declared as such because golang doesn't allow const maps
var (
	// status counter
	generalThreadStates = map[string]uint32{
		"after create":              uint32(0),
		"altering table":            uint32(0),
		"analyzing":                 uint32(0),
		"checking permissions":      uint32(0),
		"checking table":            uint32(0),
		"cleaning up":               uint32(0),
		"closing tables":            uint32(0),
		"converting heap to myisam": uint32(0),
		"copying to tmp table":      uint32(0),
		"creating sort index":       uint32(0),
		"creating table":            uint32(0),
		"creating tmp table":        uint32(0),
		"deleting":                  uint32(0),
		"executing":                 uint32(0),
		"execution of init_command": uint32(0),
		"end":                       uint32(0),
		"freeing items":             uint32(0),
		"flushing tables":           uint32(0),
		"fulltext initialization":   uint32(0),
		"idle":                      uint32(0),
		"init":                      uint32(0),
		"killed":                    uint32(0),
		"waiting for lock":          uint32(0),
		"logging slow query":        uint32(0),
		"login":                     uint32(0),
		"manage keys":               uint32(0),
		"opening tables":            uint32(0),
		"optimizing":                uint32(0),
		"preparing":                 uint32(0),
		"reading from net":          uint32(0),
		"removing duplicates":       uint32(0),
		"removing tmp table":        uint32(0),
		"reopen tables":             uint32(0),
		"repair by sorting":         uint32(0),
		"repair done":               uint32(0),
		"repair with keycache":      uint32(0),
		"replication master":        uint32(0),
		"rolling back":              uint32(0),
		"searching rows for update": uint32(0),
		"sending data":              uint32(0),
		"sorting for group":         uint32(0),
		"sorting for order":         uint32(0),
		"sorting index":             uint32(0),
		"sorting result":            uint32(0),
		"statistics":                uint32(0),
		"updating":                  uint32(0),
		"waiting for tables":        uint32(0),
		"waiting for table flush":   uint32(0),
		"waiting on cond":           uint32(0),
		"writing to net":            uint32(0),
		"other":                     uint32(0),
	}
	// plaintext statuses
	stateStatusMappings = map[string]string{
		"user sleep":     "idle",
		"creating index": "altering table",
		"committing alter table to storage engine": "altering table",
		"discard or import tablespace":             "altering table",
		"rename":                                   "altering table",
		"setup":                                    "altering table",
		"renaming result table":                    "altering table",
		"preparing for alter table":                "altering table",
		"copying to group table":                   "copying to tmp table",
		"copy to tmp table":                        "copying to tmp table",
		"query end":                                "end",
		"update":                                   "updating",
		"updating main table":                      "updating",
		"updating reference tables":                "updating",
		"system lock":                              "waiting for lock",
		"user lock":                                "waiting for lock",
		"table lock":                               "waiting for lock",
		"deleting from main table":                 "deleting",
		"deleting from reference tables":           "deleting",
	}
)

// Math constants
const (
	picoSeconds = 1e12
)

// MySQL environment.
func (m *MysqlMonitor) gatherGlobalVariables(db *sql.DB, serv string) error {
	// run query
	rows, err := db.Query(globalVariablesQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var key string
	var val sql.RawBytes

	// parse DSN and save server tag
	servtag := getDSNTag(serv)
	tags := map[string]string{"server": servtag}
	tags["metricType"] = "globalVariables"

	for k, v := range m.Tags {
		tags[k] = v
	}

	fields := make(map[string]interface{})
	for rows.Next() {
		if err := rows.Scan(&key, &val); err != nil {
			return err
		}
		key = strings.ToLower(key)

		// parse mysql version and put into field and tag
		if strings.Contains(key, "version") {
			fields[key] = string(val)
		}

		value, err := m.parseGlobalVariables(key, val)
		if err != nil {
			l.Debugf("Error parsing global variable %q: %v", key, err)
		} else {
			if value != nil {
				fields[key] = value
			}
		}
	}

	// Send any remaining fields
	if len(fields) > 0 {
		pt, err := io.MakeMetric(m.MetricName, tags, fields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
		if err != nil {
			l.Errorf("[error] : %s", err.Error())
			return err
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}
	}
	return nil
}

func (m *MysqlMonitor) parseGlobalVariables(key string, value sql.RawBytes) (interface{}, error) {
	return ConvertGlobalVariables(key, value)
}

// gatherSlaveStatuses can be used to get replication analytics
// When the server is slave, then it returns only one row.
// If the multi-source replication is set, then everything works differently
// This code does not work with multi-source replication.
func (m *MysqlMonitor) gatherSlaveStatuses(db *sql.DB, serv string) error {
	// run query
	rows, err := db.Query(slaveStatusQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	servtag := getDSNTag(serv)

	tags := map[string]string{"server": servtag}
	tags["metricType"] = "slaveStatus"
	for k, v := range m.Tags {
		tags[k] = v
	}

	fields := make(map[string]interface{})

	// to save the column names as a field key
	// scanning keys and values separately
	if rows.Next() {
		// get columns names, and create an array with its length
		cols, err := rows.Columns()
		if err != nil {
			return err
		}
		vals := make([]interface{}, len(cols))
		// fill the array with sql.Rawbytes
		for i := range vals {
			vals[i] = &sql.RawBytes{}
		}
		if err = rows.Scan(vals...); err != nil {
			return err
		}
		// range over columns, and try to parse values
		for i, col := range cols {
			if value, ok := m.parseValue(*vals[i].(*sql.RawBytes)); ok {
				fields["slave_"+col] = value
			}
		}

		pt, err := io.MakeMetric(m.MetricName, tags, fields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
		if err != nil {
			l.Errorf("[error] : %s", err.Error())
			return err
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}
	}

	return nil
}

// gatherBinaryLogs can be used to collect size and count of all binary files
// binlogs metric requires the MySQL server to turn it on in configuration
func (m *MysqlMonitor) gatherBinaryLogs(db *sql.DB, serv string) error {
	// run query
	rows, err := db.Query(binaryLogsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	// parse DSN and save host as a tag
	servtag := getDSNTag(serv)
	tags := map[string]string{"server": servtag}
	tags["metricType"] = "binaryLogSize"
	for k, v := range m.Tags {
		tags[k] = v
	}

	var (
		size     uint64 = 0
		count    uint64 = 0
		fileSize uint64
		fileName string
	)

	// iterate over rows and count the size and count of files
	for rows.Next() {
		if err := rows.Scan(&fileName, &fileSize); err != nil {
			return err
		}
		size += fileSize
		count++
	}
	fields := map[string]interface{}{
		"binary_size_bytes":  int(size / 1024),
		"binary_files_count": int(count / 1024),
	}

	pt, err := io.MakeMetric(m.MetricName, tags, fields, time.Now())
	if err != nil {
		l.Errorf("make metric point error %v", err)
	}

	_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
	if err != nil {
		l.Errorf("[error] : %s", err.Error())
		return err
	}

	err = io.NamedFeed([]byte(pt), io.Metric, name)
	if err != nil {
		l.Errorf("push metric point error %v", err)
	}

	return nil
}

// gatherGlobalStatuses can be used to get MySQL status metrics
// the mappings of actual names and names of each status to be exported
// to output is provided on mappings variable
func (m *MysqlMonitor) gatherGlobalStatuses(db *sql.DB, serv string) error {
	// run query
	rows, err := db.Query(globalStatusQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	// parse the DSN and save host name as a tag
	servtag := getDSNTag(serv)
	tags := map[string]string{"server": servtag}
	tags["metricType"] = "globalStatus"
	for k, v := range m.Tags {
		tags[k] = v
	}
	fields := make(map[string]interface{})

	for rows.Next() {
		var key string
		var val sql.RawBytes

		if err = rows.Scan(&key, &val); err != nil {
			return err
		}

		key = strings.ToLower(key)
		value, err := ConvertGlobalStatus(key, val)
		if err != nil {
			l.Debugf("Error parsing global status: %v", err)
		} else {
			if value != nil {
				fields[key] = value
			}
		}
	}

	// Send any remaining fields
	pt, err := io.MakeMetric(m.MetricName, tags, fields, time.Now())
	if err != nil {
		l.Errorf("make metric point error %v", err)
	}

	_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
	if err != nil {
		l.Errorf("[error] : %s", err.Error())
		return err
	}

	err = io.NamedFeed([]byte(pt), io.Metric, name)
	if err != nil {
		l.Errorf("push metric point error %v", err)
	}

	return nil
}

// GatherProcessList can be used to collect metrics on each running command
// and its state with its running count
func (m *MysqlMonitor) GatherProcessListStatuses(db *sql.DB, serv string) error {
	// run query
	rows, err := db.Query(infoSchemaProcessListQuery)
	if err != nil {
		return err
	}
	defer rows.Close()
	var (
		command string
		state   string
		count   uint32
	)

	var servtag string
	fields := make(map[string]interface{})
	servtag = getDSNTag(serv)

	// mapping of state with its counts
	stateCounts := make(map[string]uint32, len(generalThreadStates))
	// set map with keys and default values
	for k, v := range generalThreadStates {
		stateCounts[k] = v
	}

	for rows.Next() {
		err = rows.Scan(&command, &state, &count)
		if err != nil {
			return err
		}
		// each state has its mapping
		foundState := findThreadState(command, state)
		// count each state
		stateCounts[foundState] += count
	}

	tags := map[string]string{"server": servtag}
	tags["metricType"] = "processListStatus"
	for k, v := range m.Tags {
		tags[k] = v
	}

	for s, c := range stateCounts {
		fields[newNamespace("threads", s)] = c
	}

	pt, err := io.MakeMetric(m.MetricName, tags, fields, time.Now())
	if err != nil {
		l.Errorf("make metric point error %v", err)
	}

	_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
	if err != nil {
		l.Errorf("[error] : %s", err.Error())
		return err
	}

	err = io.NamedFeed([]byte(pt), io.Metric, name)
	if err != nil {
		l.Errorf("push metric point error %v", err)
	}

	// get count of connections from each user
	conn_rows, err := db.Query("SELECT user, sum(1) AS connections FROM INFORMATION_SCHEMA.PROCESSLIST GROUP BY user")
	if err != nil {
		return err
	}

	for conn_rows.Next() {
		var user string
		var connections int64

		err = conn_rows.Scan(&user, &connections)
		if err != nil {
			return err
		}

		tags := map[string]string{"server": servtag, "user": user}
		tags["metricType"] = "userConnectionsCount"
		for k, v := range m.Tags {
			tags[k] = v
		}

		fields := make(map[string]interface{})

		fields["connections"] = connections

		pt, err := io.MakeMetric(m.MetricName, tags, fields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
		if err != nil {
			l.Errorf("[error] : %s", err.Error())
			return err
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}
	}

	return nil
}

// GatherUserStatistics can be used to collect metrics on each running command
// and its state with its running count
func (m *MysqlMonitor) GatherUserStatisticsStatuses(db *sql.DB, serv string) error {
	// run query
	rows, err := db.Query(infoSchemaUserStatisticsQuery)
	if err != nil {
		// disable collecting if table is not found (mysql specific error)
		// (suppresses repeat errors)
		if strings.Contains(err.Error(), "nknown table 'user_statistics'") {
			m.GatherUserStatistics = false
		}
		return err
	}
	defer rows.Close()

	cols, err := columnsToLower(rows.Columns())
	if err != nil {
		return err
	}

	read, err := getColSlice(len(cols))
	if err != nil {
		return err
	}

	servtag := getDSNTag(serv)
	for rows.Next() {
		err = rows.Scan(read...)
		if err != nil {
			return err
		}

		tags := map[string]string{"server": servtag, "user": *read[0].(*string)}
		tags["metricType"] = "userStatisticsStatus"
		for k, v := range m.Tags {
			tags[k] = v
		}

		fields := map[string]interface{}{}

		for i := range cols {
			if i == 0 {
				continue // skip "user"
			}
			switch v := read[i].(type) {
			case *int64:
				fields[cols[i]] = *v
			case *float64:
				fields[cols[i]] = *v
			case *string:
				fields[cols[i]] = *v
			default:
				return fmt.Errorf("Unknown column type - %T", v)
			}
		}

		pt, err := io.MakeMetric(m.MetricName, tags, fields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
		if err != nil {
			l.Errorf("[error] : %s", err.Error())
			return err
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}
	}
	return nil
}

// columnsToLower converts selected column names to lowercase.
func columnsToLower(s []string, e error) ([]string, error) {
	if e != nil {
		return nil, e
	}
	d := make([]string, len(s))

	for i := range s {
		d[i] = strings.ToLower(s[i])
	}
	return d, nil
}

// getColSlice returns an in interface slice that can be used in the row.Scan().
func getColSlice(l int) ([]interface{}, error) {
	// list of all possible column names
	var (
		user                        string
		total_connections           int64
		concurrent_connections      int64
		connected_time              int64
		busy_time                   int64
		cpu_time                    int64
		bytes_received              int64
		bytes_sent                  int64
		binlog_bytes_written        int64
		rows_read                   int64
		rows_sent                   int64
		rows_deleted                int64
		rows_inserted               int64
		rows_updated                int64
		select_commands             int64
		update_commands             int64
		other_commands              int64
		commit_transactions         int64
		rollback_transactions       int64
		denied_connections          int64
		lost_connections            int64
		access_denied               int64
		empty_queries               int64
		total_ssl_connections       int64
		max_statement_time_exceeded int64
		// maria specific
		fbusy_time float64
		fcpu_time  float64
		// percona specific
		rows_fetched    int64
		table_rows_read int64
	)

	switch l {
	case 23: // maria5
		return []interface{}{
			&user,
			&total_connections,
			&concurrent_connections,
			&connected_time,
			&fbusy_time,
			&fcpu_time,
			&bytes_received,
			&bytes_sent,
			&binlog_bytes_written,
			&rows_read,
			&rows_sent,
			&rows_deleted,
			&rows_inserted,
			&rows_updated,
			&select_commands,
			&update_commands,
			&other_commands,
			&commit_transactions,
			&rollback_transactions,
			&denied_connections,
			&lost_connections,
			&access_denied,
			&empty_queries,
		}, nil
	case 25: // maria10
		return []interface{}{
			&user,
			&total_connections,
			&concurrent_connections,
			&connected_time,
			&fbusy_time,
			&fcpu_time,
			&bytes_received,
			&bytes_sent,
			&binlog_bytes_written,
			&rows_read,
			&rows_sent,
			&rows_deleted,
			&rows_inserted,
			&rows_updated,
			&select_commands,
			&update_commands,
			&other_commands,
			&commit_transactions,
			&rollback_transactions,
			&denied_connections,
			&lost_connections,
			&access_denied,
			&empty_queries,
			&total_ssl_connections,
			&max_statement_time_exceeded,
		}, nil
	case 21: // mysql 5.5
		return []interface{}{
			&user,
			&total_connections,
			&concurrent_connections,
			&connected_time,
			&busy_time,
			&cpu_time,
			&bytes_received,
			&bytes_sent,
			&binlog_bytes_written,
			&rows_fetched,
			&rows_updated,
			&table_rows_read,
			&select_commands,
			&update_commands,
			&other_commands,
			&commit_transactions,
			&rollback_transactions,
			&denied_connections,
			&lost_connections,
			&access_denied,
			&empty_queries,
		}, nil
	case 22: // percona
		return []interface{}{
			&user,
			&total_connections,
			&concurrent_connections,
			&connected_time,
			&busy_time,
			&cpu_time,
			&bytes_received,
			&bytes_sent,
			&binlog_bytes_written,
			&rows_fetched,
			&rows_updated,
			&table_rows_read,
			&select_commands,
			&update_commands,
			&other_commands,
			&commit_transactions,
			&rollback_transactions,
			&denied_connections,
			&lost_connections,
			&access_denied,
			&empty_queries,
			&total_ssl_connections,
		}, nil
	}

	return nil, fmt.Errorf("Not Supported - %d columns", l)
}

// gatherPerfTableIOWaits can be used to get total count and time
// of I/O wait event for each table and process
func (m *MysqlMonitor) gatherPerfTableIOWaits(db *sql.DB, serv string) error {
	rows, err := db.Query(perfTableIOWaitsQuery)
	if err != nil {
		return err
	}

	defer rows.Close()
	var (
		objSchema, objName, servtag                       string
		countFetch, countInsert, countUpdate, countDelete float64
		timeFetch, timeInsert, timeUpdate, timeDelete     float64
	)

	servtag = getDSNTag(serv)

	for rows.Next() {
		err = rows.Scan(&objSchema, &objName,
			&countFetch, &countInsert, &countUpdate, &countDelete,
			&timeFetch, &timeInsert, &timeUpdate, &timeDelete,
		)

		if err != nil {
			return err
		}

		tags := map[string]string{
			"server": servtag,
			"schema": objSchema,
			"name":   objName,
		}
		tags["metricType"] = "perfTableIOWait"

		for k, v := range m.Tags {
			tags[k] = v
		}

		fields := map[string]interface{}{
			"table_io_waits_total_fetch":          countFetch,
			"table_io_waits_total_insert":         countInsert,
			"table_io_waits_total_update":         countUpdate,
			"table_io_waits_total_delete":         countDelete,
			"table_io_waits_seconds_total_fetch":  timeFetch / picoSeconds,
			"table_io_waits_seconds_total_insert": timeInsert / picoSeconds,
			"table_io_waits_seconds_total_update": timeUpdate / picoSeconds,
			"table_io_waits_seconds_total_delete": timeDelete / picoSeconds,
		}

		pt, err := io.MakeMetric(m.MetricName, tags, fields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
		if err != nil {
			l.Errorf("[error] : %s", err.Error())
			return err
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}
	}
	return nil
}

// gatherPerfIndexIOWaits can be used to get total count and time
// of I/O wait event for each index and process
func (m *MysqlMonitor) gatherPerfIndexIOWaits(db *sql.DB, serv string) error {
	rows, err := db.Query(perfIndexIOWaitsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		objSchema, objName, indexName, servtag            string
		countFetch, countInsert, countUpdate, countDelete float64
		timeFetch, timeInsert, timeUpdate, timeDelete     float64
	)

	servtag = getDSNTag(serv)

	for rows.Next() {
		err = rows.Scan(&objSchema, &objName, &indexName,
			&countFetch, &countInsert, &countUpdate, &countDelete,
			&timeFetch, &timeInsert, &timeUpdate, &timeDelete,
		)

		if err != nil {
			return err
		}

		tags := map[string]string{
			"server": servtag,
			"schema": objSchema,
			"name":   objName,
			"index":  indexName,
		}

		for k, v := range m.Tags {
			tags[k] = v
		}

		tags["metricType"] = "perfIndexIOWait"

		fields := map[string]interface{}{
			"index_io_waits_total_fetch":         countFetch,
			"index_io_waits_seconds_total_fetch": timeFetch / picoSeconds,
		}

		// update write columns only when index is NONE
		if indexName == "NONE" {
			fields["index_io_waits_total_insert"] = countInsert
			fields["index_io_waits_total_update"] = countUpdate
			fields["index_io_waits_total_delete"] = countDelete

			fields["index_io_waits_seconds_total_insert"] = timeInsert / picoSeconds
			fields["index_io_waits_seconds_total_update"] = timeUpdate / picoSeconds
			fields["index_io_waits_seconds_total_delete"] = timeDelete / picoSeconds
		}

		pt, err := io.MakeMetric(m.MetricName, tags, fields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
		if err != nil {
			l.Errorf("[error] : %s", err.Error())
			return err
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}
	}
	return nil
}

// gatherInfoSchemaAutoIncStatuses can be used to get auto incremented values of the column
func (m *MysqlMonitor) gatherInfoSchemaAutoIncStatuses(db *sql.DB, serv string) error {
	rows, err := db.Query(infoSchemaAutoIncQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		schema, table, column string
		incValue, maxInt      int64
	)

	servtag := getDSNTag(serv)

	for rows.Next() {
		if err := rows.Scan(&schema, &table, &column, &incValue, &maxInt); err != nil {
			return err
		}
		tags := map[string]string{
			"server": servtag,
			"schema": schema,
			"table":  table,
			"column": column,
		}

		for k, v := range m.Tags {
			tags[k] = v
		}

		tags["metricType"] = "schemaAutoIncStatus"

		fields := make(map[string]interface{})
		fields["auto_increment_column"] = incValue
		fields["auto_increment_column_max"] = maxInt

		pt, err := io.MakeMetric(m.MetricName, tags, fields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
		if err != nil {
			l.Errorf("[error] : %s", err.Error())
			return err
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}
	}
	return nil
}

// gatherInnoDBMetrics can be used to fetch enabled metrics from
// information_schema.INNODB_METRICS
func (m *MysqlMonitor) gatherInnoDBMetrics(db *sql.DB, serv string) error {
	// run query
	rows, err := db.Query(innoDBMetricsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	// parse DSN and save server tag
	servtag := getDSNTag(serv)
	tags := map[string]string{"server": servtag}
	tags["metricType"] = "innoDBMetric"

	for k, v := range m.Tags {
		tags[k] = v
	}

	fields := make(map[string]interface{})
	for rows.Next() {
		var key string
		var val sql.RawBytes
		if err := rows.Scan(&key, &val); err != nil {
			return err
		}
		key = strings.ToLower(key)
		if value, ok := m.parseValue(val); ok {
			fields[key] = value
		}
	}
	// Send any remaining fields
	if len(fields) > 0 {
		pt, err := io.MakeMetric(m.MetricName, tags, fields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
		if err != nil {
			l.Errorf("[error] : %s", err.Error())
			return err
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}
	}
	return nil
}

// gatherPerfTableLockWaits can be used to get
// the total number and time for SQL and external lock wait events
// for each table and operation
// requires the MySQL server to be enabled to save this metric
func (m *MysqlMonitor) gatherPerfTableLockWaits(db *sql.DB, serv string) error {
	// check if table exists,
	// if performance_schema is not enabled, tables do not exist
	// then there is no need to scan them
	var tableName string
	err := db.QueryRow(perfSchemaTablesQuery, "table_lock_waits_summary_by_table").Scan(&tableName)
	switch {
	case err == sql.ErrNoRows:
		return nil
	case err != nil:
		return err
	}

	rows, err := db.Query(perfTableLockWaitsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	servtag := getDSNTag(serv)

	var (
		objectSchema               string
		objectName                 string
		countReadNormal            float64
		countReadWithSharedLocks   float64
		countReadHighPriority      float64
		countReadNoInsert          float64
		countReadExternal          float64
		countWriteAllowWrite       float64
		countWriteConcurrentInsert float64
		countWriteLowPriority      float64
		countWriteNormal           float64
		countWriteExternal         float64
		timeReadNormal             float64
		timeReadWithSharedLocks    float64
		timeReadHighPriority       float64
		timeReadNoInsert           float64
		timeReadExternal           float64
		timeWriteAllowWrite        float64
		timeWriteConcurrentInsert  float64
		timeWriteLowPriority       float64
		timeWriteNormal            float64
		timeWriteExternal          float64
	)

	for rows.Next() {
		err = rows.Scan(
			&objectSchema,
			&objectName,
			&countReadNormal,
			&countReadWithSharedLocks,
			&countReadHighPriority,
			&countReadNoInsert,
			&countReadExternal,
			&countWriteAllowWrite,
			&countWriteConcurrentInsert,
			&countWriteLowPriority,
			&countWriteNormal,
			&countWriteExternal,
			&timeReadNormal,
			&timeReadWithSharedLocks,
			&timeReadHighPriority,
			&timeReadNoInsert,
			&timeReadExternal,
			&timeWriteAllowWrite,
			&timeWriteConcurrentInsert,
			&timeWriteLowPriority,
			&timeWriteNormal,
			&timeWriteExternal,
		)

		if err != nil {
			return err
		}
		tags := map[string]string{
			"server": servtag,
			"schema": objectSchema,
			"table":  objectName,
		}

		for k, v := range m.Tags {
			tags[k] = v
		}

		tags["metricType"] = "perfTableLockWait"

		sqlLWTags := copyTags(tags)
		sqlLWTags["perf_query"] = "sql_lock_waits_total"
		sqlLWFields := map[string]interface{}{
			"read_normal":             countReadNormal,
			"read_with_shared_locks":  countReadWithSharedLocks,
			"read_high_priority":      countReadHighPriority,
			"read_no_insert":          countReadNoInsert,
			"write_normal":            countWriteNormal,
			"write_allow_write":       countWriteAllowWrite,
			"write_concurrent_insert": countWriteConcurrentInsert,
			"write_low_priority":      countWriteLowPriority,
		}

		pt, err := io.MakeMetric(m.MetricName, sqlLWTags, sqlLWFields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
		if err != nil {
			l.Errorf("[error] : %s", err.Error())
			return err
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}

		externalLWTags := copyTags(tags)
		externalLWTags["perf_query"] = "external_lock_waits_total"
		externalLWFields := map[string]interface{}{
			"read":  countReadExternal,
			"write": countWriteExternal,
		}

		pt, err = io.MakeMetric(m.MetricName, externalLWTags, externalLWFields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
		if err != nil {
			l.Errorf("[error] : %s", err.Error())
			return err
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}

		sqlLWSecTotalTags := copyTags(tags)
		sqlLWSecTotalTags["perf_query"] = "sql_lock_waits_seconds_total"
		sqlLWSecTotalFields := map[string]interface{}{
			"read_normal":             timeReadNormal / picoSeconds,
			"read_with_shared_locks":  timeReadWithSharedLocks / picoSeconds,
			"read_high_priority":      timeReadHighPriority / picoSeconds,
			"read_no_insert":          timeReadNoInsert / picoSeconds,
			"write_normal":            timeWriteNormal / picoSeconds,
			"write_allow_write":       timeWriteAllowWrite / picoSeconds,
			"write_concurrent_insert": timeWriteConcurrentInsert / picoSeconds,
			"write_low_priority":      timeWriteLowPriority / picoSeconds,
		}

		pt, err = io.MakeMetric(m.MetricName, sqlLWSecTotalTags, sqlLWSecTotalFields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
		if err != nil {
			l.Errorf("[error] : %s", err.Error())
			return err
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}

		externalLWSecTotalTags := copyTags(tags)
		externalLWSecTotalTags["perf_query"] = "external_lock_waits_seconds_total"
		externalLWSecTotalFields := map[string]interface{}{
			"read":  timeReadExternal / picoSeconds,
			"write": timeWriteExternal / picoSeconds,
		}

		pt, err = io.MakeMetric(m.MetricName, externalLWSecTotalTags, externalLWSecTotalFields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
		if err != nil {
			l.Errorf("[error] : %s", err.Error())
			return err
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}
	}
	return nil
}

// gatherPerfEventWaits can be used to get total time and number of event waits
func (m *MysqlMonitor) gatherPerfEventWaits(db *sql.DB, serv string) error {
	rows, err := db.Query(perfEventWaitsQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		event               string
		starCount, timeWait float64
	)

	servtag := getDSNTag(serv)
	tags := map[string]string{
		"server": servtag,
	}

	for k, v := range m.Tags {
		tags[k] = v
	}

	tags["metricType"] = "perfEventWaits"

	for rows.Next() {
		if err := rows.Scan(&event, &starCount, &timeWait); err != nil {
			return err
		}
		tags["event_name"] = event
		fields := map[string]interface{}{
			"events_waits_total":         starCount,
			"events_waits_seconds_total": timeWait / picoSeconds,
		}

		pt, err := io.MakeMetric(m.MetricName, tags, fields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
		if err != nil {
			l.Errorf("[error] : %s", err.Error())
			return err
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}
	}
	return nil
}

// gatherPerfFileEvents can be used to get stats on file events
func (m *MysqlMonitor) gatherPerfFileEventsStatuses(db *sql.DB, serv string) error {
	rows, err := db.Query(perfFileEventsQuery)
	if err != nil {
		return err
	}

	defer rows.Close()

	var (
		eventName                                 string
		countRead, countWrite, countMisc          float64
		sumTimerRead, sumTimerWrite, sumTimerMisc float64
		sumNumBytesRead, sumNumBytesWrite         float64
	)

	servtag := getDSNTag(serv)
	tags := map[string]string{
		"server": servtag,
	}

	for k, v := range m.Tags {
		tags[k] = v
	}

	tags["metricType"] = "perfFileEventsStatus"

	for rows.Next() {
		err = rows.Scan(
			&eventName,
			&countRead, &sumTimerRead, &sumNumBytesRead,
			&countWrite, &sumTimerWrite, &sumNumBytesWrite,
			&countMisc, &sumTimerMisc,
		)
		if err != nil {
			return err
		}

		tags["event_name"] = eventName
		miscfields := make(map[string]interface{})

		miscTags := copyTags(tags)
		miscTags["mode"] = "misc"
		miscfields["file_events_total"] = countWrite
		miscfields["file_events_seconds_total"] = sumTimerMisc / picoSeconds

		pt, err := io.MakeMetric(m.MetricName, miscTags, miscfields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}

		readTags := copyTags(tags)
		readTags["mode"] = "read"
		readfields := make(map[string]interface{})

		readfields["file_events_total"] = countRead
		readfields["file_events_seconds_total"] = sumTimerRead / picoSeconds
		readfields["file_events_bytes_totals"] = sumNumBytesRead

		pt, err = io.MakeMetric(m.MetricName, readTags, readfields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
		if err != nil {
			l.Errorf("[error] : %s", err.Error())
			return err
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}

		writeTags := copyTags(tags)
		writeTags["mode"] = "write"
		writefields := make(map[string]interface{})

		writefields["file_events_total"] = countWrite
		writefields["file_events_seconds_total"] = sumTimerWrite / picoSeconds
		writefields["file_events_bytes_totals"] = sumNumBytesWrite

		pt, err = io.MakeMetric(m.MetricName, tags, writefields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
		if err != nil {
			l.Errorf("[error] : %s", err.Error())
			return err
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}

	}
	return nil
}

// gatherPerfEventsStatements can be used to get attributes of each event
func (m *MysqlMonitor) gatherPerfEventsStatements(db *sql.DB, serv string) error {
	query := fmt.Sprintf(
		perfEventsStatementsQuery,
		m.PerfEventsStatementsDigestTextLimit,
		m.PerfEventsStatementsTimeLimit,
		m.PerfEventsStatementsLimit,
	)

	rows, err := db.Query(query)
	if err != nil {
		return err
	}

	defer rows.Close()

	var (
		schemaName, digest, digest_text      string
		count, queryTime, errors, warnings   float64
		rowsAffected, rowsSent, rowsExamined float64
		tmpTables, tmpDiskTables             float64
		sortMergePasses, sortRows            float64
		noIndexUsed                          float64
	)

	servtag := getDSNTag(serv)
	tags := map[string]string{
		"server": servtag,
	}

	for k, v := range m.Tags {
		tags[k] = v
	}

	for k, v := range m.Tags {
		tags[k] = v
	}

	tags["metricType"] = "perfEventsStatements"

	for rows.Next() {
		err = rows.Scan(
			&schemaName, &digest, &digest_text,
			&count, &queryTime, &errors, &warnings,
			&rowsAffected, &rowsSent, &rowsExamined,
			&tmpTables, &tmpDiskTables,
			&sortMergePasses, &sortRows,
			&noIndexUsed,
		)

		if err != nil {
			return err
		}
		tags["schema"] = schemaName
		tags["digest"] = digest
		tags["digest_text"] = digest_text

		fields := map[string]interface{}{
			"events_statements_total":                   count,
			"events_statements_seconds_total":           queryTime / picoSeconds,
			"events_statements_errors_total":            errors,
			"events_statements_warnings_total":          warnings,
			"events_statements_rows_affected_total":     rowsAffected,
			"events_statements_rows_sent_total":         rowsSent,
			"events_statements_rows_examined_total":     rowsExamined,
			"events_statements_tmp_tables_total":        tmpTables,
			"events_statements_tmp_disk_tables_total":   tmpDiskTables,
			"events_statements_sort_merge_passes_total": sortMergePasses,
			"events_statements_sort_rows_total":         sortRows,
			"events_statements_no_index_used_total":     noIndexUsed,
		}

		pt, err := io.MakeMetric(m.MetricName, tags, fields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
		if err != nil {
			l.Errorf("[error] : %s", err.Error())
			return err
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}
	}
	return nil
}

// gatherTableSchema can be used to gather stats on each schema
func (m *MysqlMonitor) gatherTableSchema(db *sql.DB, serv string) error {
	var dbList []string
	servtag := getDSNTag(serv)

	// if the list of databases if empty, then get all databases
	if len(m.TableSchemaDatabases) == 0 {
		rows, err := db.Query(dbListQuery)
		if err != nil {
			return err
		}
		defer rows.Close()

		var database string
		for rows.Next() {
			err = rows.Scan(&database)
			if err != nil {
				return err
			}

			dbList = append(dbList, database)
		}
	} else {
		dbList = m.TableSchemaDatabases
	}

	for _, database := range dbList {
		rows, err := db.Query(fmt.Sprintf(tableSchemaQuery, database))
		if err != nil {
			return err
		}
		defer rows.Close()
		var (
			tableSchema   string
			tableName     string
			tableType     string
			engine        string
			version       float64
			rowFormat     string
			tableRows     float64
			dataLength    float64
			indexLength   float64
			dataFree      float64
			createOptions string
		)

		for rows.Next() {
			err = rows.Scan(
				&tableSchema,
				&tableName,
				&tableType,
				&engine,
				&version,
				&rowFormat,
				&tableRows,
				&dataLength,
				&indexLength,
				&dataFree,
				&createOptions,
			)
			if err != nil {
				return err
			}
			tags := map[string]string{"server": servtag}
			tags["schema"] = tableSchema
			tags["table"] = tableName
			tags["metricType"] = "tableSchemaStat"

			for k, v := range m.Tags {
				tags[k] = v
			}

			fields := make(map[string]interface{})

			versionTags := copyTags(tags)
			versionTags["type"] = tableType
			versionTags["engine"] = engine
			versionTags["row_format"] = rowFormat
			versionTags["create_options"] = createOptions

			fields["rows"] = tableRows
			fields["data_length"] = dataLength
			fields["index_length"] = indexLength
			fields["data_free"] = dataFree
			fields["table_version"] = version

			pt, err := io.MakeMetric(m.MetricName, versionTags, fields, time.Now())
			if err != nil {
				l.Errorf("make metric point error %v", err)
			}

			_, err = influxm.ParsePointsWithPrecision(pt, time.Now().UTC(), "")
			if err != nil {
				l.Errorf("[error] : %s", err.Error())
				return err
			}

			err = io.NamedFeed([]byte(pt), io.Metric, name)
			if err != nil {
				l.Errorf("push metric point error %v", err)
			}
		}
	}
	return nil
}

func (m *MysqlMonitor) parseValue(value sql.RawBytes) (interface{}, bool) {
	return parseValue(value)
}

// parseValue can be used to convert values such as "ON","OFF","Yes","No" to 0,1
func parseValue(value sql.RawBytes) (interface{}, bool) {
	if bytes.EqualFold(value, []byte("YES")) || bytes.Compare(value, []byte("ON")) == 0 {
		return 1, true
	}

	if bytes.EqualFold(value, []byte("NO")) || bytes.Compare(value, []byte("OFF")) == 0 {
		return 0, true
	}

	if val, err := strconv.ParseInt(string(value), 10, 64); err == nil {
		return val, true
	}
	if val, err := strconv.ParseFloat(string(value), 64); err == nil {
		return val, true
	}

	if len(string(value)) > 0 {
		return string(value), true
	}
	return nil, false
}

// findThreadState can be used to find thread state by command and plain state
func findThreadState(rawCommand, rawState string) string {
	var (
		// replace '_' symbol with space
		command = strings.Replace(strings.ToLower(rawCommand), "_", " ", -1)
		state   = strings.Replace(strings.ToLower(rawState), "_", " ", -1)
	)
	// if the state is already valid, then return it
	if _, ok := generalThreadStates[state]; ok {
		return state
	}

	// if state is plain, return the mapping
	if mappedState, ok := stateStatusMappings[state]; ok {
		return mappedState
	}
	// if the state is any lock, return the special state
	if strings.Contains(state, "waiting for") && strings.Contains(state, "lock") {
		return "waiting for lock"
	}

	if command == "sleep" && state == "" {
		return "idle"
	}

	if command == "query" {
		return "executing"
	}

	if command == "binlog dump" {
		return "replication master"
	}
	// if no mappings found and state is invalid, then return "other" state
	return "other"
}

func copyTags(in map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range in {
		out[k] = v
	}
	return out
}

func dsnAddTimeout(dsn string) (string, error) {
	conf, err := mysql.ParseDSN(dsn)
	if err != nil {
		l.Errorf("mysql.ParseDSN(): %s", err.Error())
		return "", err
	}

	if conf.Timeout == 0 {
		conf.Timeout = time.Second * 5
	}

	return conf.FormatDSN(), nil
}

// newNamespace can be used to make a namespace
func newNamespace(words ...string) string {
	return strings.Replace(strings.Join(words, "_"), " ", "_", -1)
}

func getDSNTag(dsn string) string {
	conf, err := mysql.ParseDSN(dsn)
	if err != nil {
		return "127.0.0.1:3306"
	}
	return conf.Addr
}

func (m *MysqlMonitor) gatherExtend(db *sql.DB, serv string) {
	for key, item := range metricMap {
		resMap, err := m.query(db, serv, item)
		if err != nil {
			l.Errorf("mysql query faild %v", err)
		}

		servtag := getDSNTag(serv)

		m.handleResponse(key, servtag, resMap)
	}
}

func (m *MysqlMonitor) handleResponse(mm string, servtag string, response []map[string]interface{}) error {
	for _, item := range response {
		tags := map[string]string{"server": servtag}
		tags["metricType"] = mm

		for k, v := range m.Tags {
			tags[k] = v
		}

		pt, err := io.MakeMetric(m.MetricName, tags, item, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}

		err = io.NamedFeed([]byte(pt), io.Metric, name)
		if err != nil {
			l.Errorf("push metric point error %v", err)
		}
	}

	return nil
}

func (r *MysqlMonitor) query(db *sql.DB, serv string, sql string) ([]map[string]interface{}, error) {
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	columnLength := len(columns)
	cache := make([]interface{}, columnLength)
	for idx, _ := range cache {
		var a interface{}
		cache[idx] = &a
	}
	var list []map[string]interface{}
	for rows.Next() {
		_ = rows.Scan(cache...)

		item := make(map[string]interface{})
		for i, data := range cache {
			key := strings.ToLower(columns[i])
			val := *data.(*interface{})

			if val != nil {
				vType := reflect.TypeOf(val)

				switch vType.String() {
				case "int64":
					item[key] = val.(int64)
				case "string":
					var data interface{}
					data, err := strconv.ParseFloat(val.(string), 64)
					if err != nil {
						data = val
					}
					item[key] = data
				case "time.Time":
					item[key] = val.(time.Time)
				case "[]uint8":
					item[key] = string(val.([]uint8))
				default:
					return nil, fmt.Errorf("unsupport data type '%s' now\n", vType)
				}
			}
		}

		list = append(list, item)
	}
	return list, nil
}
