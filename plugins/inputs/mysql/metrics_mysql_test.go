package mysql

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

//----------------------------------------------------------------------
// mysql_schema

// go test -v -timeout 30s -run ^TestGetCleanSchemaData$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/mysql
func TestGetCleanSchemaData(t *testing.T) {
	cases := []struct {
		name   string
		rows   *mockRows
		expect map[string]interface{}
	}{
		// mysql 5
		{
			name: "mysql_5_schema_size",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{"information_schema", "0.15625000"},
					{"mysql", "7.73961163"},
					{"performance_schema", "0.00000000"},
					{"sys", "0.01562500"},
				},
			},
			expect: map[string]interface{}{
				"information_schema": "0.15625000",
				"mysql":              "7.73961163",
				"performance_schema": "0.00000000",
				"sys":                "0.01562500",
			},
		},

		{
			name: "mysql_5_schema_time",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{"mysqlslap", "4905625"},
				},
			},
			expect: map[string]interface{}{
				"mysqlslap": "4905625",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := getCleanSchemaData(tc.rows)

			for k, v := range tc.expect {
				switch x := v.(type) {
				case int64:
					tu.Assert(t, res[k] != nil, "key %s should not be nil", k)
					tu.Equals(t, x, res[k].(int64))
				case string:
					tu.Equals(t, x, res[k].(string))
				default:
					t.Logf("%s is type %s", k, reflect.TypeOf(v))
				}
			}
		})
	}
}

//----------------------------------------------------------------------
// mysql_innodb

// go test -v -timeout 30s -run ^TestGetCleanInnodb$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/mysql
func TestGetCleanInnodb(t *testing.T) {
	cases := []struct {
		name   string
		rows   *mockRows
		expect map[string]interface{}
	}{
		// mysql 5
		{
			name: "mysql_5_innodb",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{"table_schema", "mysqlslap"},
					{"lock_deadlocks", "0"},
					{"lock_timeouts", "0"},
					{"lock_row_lock_current_waits", "0"},
					{"lock_row_lock_time", "0"},
					{"lock_row_lock_time_max", "0"},
					{"lock_row_lock_waits", "0"},
					{"lock_row_lock_time_avg", "0"},
					{"buffer_pool_size", "134217728"},
					{"buffer_pool_reads", "589"},
					{"buffer_pool_read_requests", "1741"},
				},
			},
			expect: map[string]interface{}{
				"table_schema":                "mysqlslap",
				"lock_deadlocks":              "0",
				"lock_timeouts":               "0",
				"lock_row_lock_current_waits": "0",
				"lock_row_lock_time":          "0",
				"lock_row_lock_time_max":      "0",
				"lock_row_lock_waits":         "0",
				"lock_row_lock_time_avg":      "0",
				"buffer_pool_size":            "134217728",
				"buffer_pool_reads":           "589",
				"buffer_pool_read_requests":   "1741",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := getCleanSchemaData(tc.rows)

			for k, v := range tc.expect {
				switch x := v.(type) {
				case int:
					tu.Assert(t, res[k] != nil, "key %s should not be nil", k)
					tu.Equals(t, x, res[k].(int))
				case int64:
					tu.Assert(t, res[k] != nil, "key %s should not be nil", k)
					tu.Equals(t, x, res[k].(int64))
				case string:
					tu.Equals(t, x, res[k].(string))
				case bool:
					tu.Equals(t, x, res[k].(bool))
				case nil:
					tu.Equals(t, x, res[k])
				default:
					t.Logf("%s is type %s", k, reflect.TypeOf(v))
				}
			}
		})
	}
}

//----------------------------------------------------------------------
// mysql_table_schema

// go test -v -timeout 30s -run ^TestGetCleanTableSchema$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/mysql
func TestGetCleanTableSchema(t *testing.T) {
	cases := []struct {
		name   string
		rows   *mockRows
		expect map[string]interface{}
	}{
		// mysql 5
		{
			name: "mysql_5_table_schema",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{"table_schema", "mysqlslap"},
					{"table_name", "t1"},
					{"table_type", "BASE TABLE"},
					{"engine", "InnoDB"},
					{"version", "10"},
					{"table_rows", "571"},
					{"data_length", "13156352"},
					{"index_length", "0"},
					{"data_free", "4194304"},
				},
			},
			expect: map[string]interface{}{
				"table_schema": "mysqlslap",
				"table_name":   "t1",
				"table_type":   "BASE TABLE",
				"engine":       "InnoDB",
				"version":      "10",
				"table_rows":   "571",
				"data_length":  "13156352",
				"index_length": "0",
				"data_free":    "4194304",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := getCleanSchemaData(tc.rows)

			for k, v := range tc.expect {
				switch x := v.(type) {
				case int64:
					tu.Assert(t, res[k] != nil, "key %s should not be nil", k)
					tu.Equals(t, x, res[k].(int64))
				case string:
					tu.Equals(t, x, res[k].(string))
				default:
					t.Logf("%s is type %s", k, reflect.TypeOf(v))
				}
			}
		})
	}
}

//----------------------------------------------------------------------
// mysql_user_status

// go test -v -timeout 30s -run ^TestGetCleanUserStatusName$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/mysql
func TestGetCleanUserStatusName(t *testing.T) {
	cases := []struct {
		name   string
		rows   *mockRows
		expect map[string]interface{}
	}{
		// mysql 5
		{
			name: "mysql_5_user_name",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{"root"},
					{"mysql.session"},
					{"mysql.sys"},
				},
			},
			expect: map[string]interface{}{
				"root":          true,
				"mysql.session": nil,
				"mysql.sys":     nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := getCleanUserStatusName(tc.rows)

			t.Logf("res = %v", res)

			for k, v := range tc.expect {
				switch x := v.(type) {
				case int64:
					tu.Assert(t, res[k] != nil, "key %s should not be nil", k)
					tu.Equals(t, x, res[k].(int64))
				case string:
					tu.Equals(t, x, res[k].(string))
				case bool:
					tu.Equals(t, x, res[k].(bool))
				case nil:
					tu.Equals(t, x, res[k])
				default:
					t.Logf("%s is type %s", k, reflect.TypeOf(v))
				}
			}
		})
	}
}

// go test -v -timeout 30s -run ^TestGetCleanUserStatusVariable$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/mysql
func TestGetCleanUserStatusVariable(t *testing.T) {
	cases := []struct {
		name   string
		rows   *mockRows
		expect map[string]interface{}
	}{
		// mysql 5
		{
			name: "mysql_5_user_variable",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{"bytes_received", "9067974"},
					{"bytes_sent", "4345363594"},
				},
			},
			expect: map[string]interface{}{
				"bytes_received": int64(9067974),
				"bytes_sent":     int64(4345363594),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := getCleanUserStatusVariable(tc.rows)

			t.Logf("res = %v", res)

			for k, v := range tc.expect {
				switch x := v.(type) {
				case int:
					tu.Assert(t, res[k] != nil, "key %s should not be nil", k)
					tu.Equals(t, x, res[k].(int))
				case int64:
					tu.Assert(t, res[k] != nil, "key %s should not be nil", k)
					tu.Equals(t, x, res[k].(int64))
				case string:
					tu.Equals(t, x, res[k].(string))
				case bool:
					tu.Equals(t, x, res[k].(bool))
				case nil:
					tu.Equals(t, x, res[k])
				default:
					t.Logf("%s is type %s", k, reflect.TypeOf(v))
				}
			}
		})
	}
}

// go test -v -timeout 30s -run ^TestGetCleanUserStatusConnection$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/mysql
func TestGetCleanUserStatusConnection(t *testing.T) {
	cases := []struct {
		name   string
		rows   *mockRows
		expect map[string]interface{}
	}{
		// mysql 5
		{
			name: "mysql_5_user_connection",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{"root", int64(1), int64(109)},
				},
			},
			expect: map[string]interface{}{
				"current_connect": int64(1),
				"total_connect":   int64(109),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := getCleanUserStatusConnection(tc.rows)

			for k, v := range tc.expect {
				switch x := v.(type) {
				case int:
					tu.Assert(t, res[k] != nil, "key %s should not be nil", k)
					tu.Equals(t, x, res[k].(int))
				case int64:
					tu.Assert(t, res[k] != nil, "key %s should not be nil", k)
					tu.Equals(t, x, res[k].(int64))
				case string:
					tu.Equals(t, x, res[k].(string))
				case bool:
					tu.Equals(t, x, res[k].(bool))
				case nil:
					tu.Equals(t, x, res[k])
				default:
					t.Logf("%s is type %s", k, reflect.TypeOf(v))
				}
			}
		})
	}
}

//----------------------------------------------------------------------
// mysql_dbm_metric

// go test -v -timeout 30s -run ^TestGetCleanSummaryRows$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/mysql
func TestGetCleanSummaryRows(t *testing.T) {
	cases := []struct {
		name   string
		rows   *mockRows
		expect []dbmRow
	}{
		// mysql 5
		{
			name: "mysql_5_summary_rows",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{nil, "feaff321c54a9c8e1e9508628f7a5a05", "SHOW WARNINGS ", "18", "8658974000", "0", "0", "0", "0", "0", "0", "0", "0", "0"},
					{nil, "2cf2906bafd55ee74be1b693d8234ae8", "SELECT SCHEMA ( ) ", "6", "3062148000", "0", "0", "0", "6", "0", "0", "0", "0", "0"},
					{nil, "d13d9fe69dd20a9855c0807894171379", "SELECT @@SESSION . `transaction_isolation` ", "4", "2356537000", "0", "0", "0", "4", "0", "0", "0", "0", "0"},
					{nil, "e6c32877475171cb7eed4878606e2f9c", "SET SESSION TRANSACTION READ WRITE ", "3", "1109201000", "0", "0", "0", "0", "0", "0", "0", "0", "0"},
					{nil, "94ea89a88d21e9b8268539fc20c018b3", "SELECT @@SESSION . `transaction_read_only` ", "3", "1237410000", "0", "0", "0", "3", "0", "0", "0", "0", "0"},
					{nil, "79626ab1f2c0945f90348821eaa1d4a6", "SET `net_write_timeout` = ? ", "3", "1858841000", "0", "0", "0", "0", "0", "0", "0", "0", "0"},
					{nil, "ba203e60d976926b307982e69455be3e", "SET `SQL_SELECT_LIMIT` = ? ", "3", "1723062000", "0", "0", "0", "0", "0", "0", "0", "0", "0"},
					{nil, "3dea1a31bbe89e1c205e3ca43d44512b", "SELECT NAME , `COUNT` FROM `information_schema` . `INNODB_METRICS` WHERE STATUS = ? ", "2", "3228510000", "382000000", "0", "0", "130", "470", "2", "0", "2,0"},
					{nil, "1db261454fc2f5418dcd0aab83a1b444", "SET `SQL_SELECT_LIMIT` = DEFAULT ", "2", "952665000", "0", "0", "0", "0", "0", "0", "0", "0", "0"},
					{nil, "1fa9e9651540b0e65d6c065f4bdfb036", "SELECT @@SESSION . `auto_increment_increment` AS `auto_increment_increment` , @@`character_set_client` AS `character_set_client` , @@`character_set_connection` AS `character_set_connection` , @@`character_set_results` AS `character_set_results` , @@`character_set_server` AS `character_set_server` , @@`collation_server` AS `collation_server` , @@`collation_connection` AS `collation_connection` , @@`init_connect` AS `init_connect` , @@`interactive_timeout` AS `interactive_timeout` , @@`license` AS `license` , @@`lower_case_table_names` AS `lower_case_table_names` , @@`max_allowed_packet` AS `max_allowed_packet` , @@`net_write_timeout` AS `net_write_timeout` , @@`performance_schema` AS `performance_schema` , @@`query_cache_size` AS `query_cache_size` , @@`query_cache_type` AS `query_cache_type` , @@`sql_mode` AS `sql_mode` , @@`system_time_zone` AS `system_time_zone` , @@`time_zone` AS `time_zone` , @@`transaction_isolation` AS `transaction_isolation` , @@", "1", "605229000", "0", "0", "0", "1", "0", "0", "0", "0", "0"},
				},
			},
			expect: []dbmRow{
				{
					/*"schemaName":        */ "",
					/*"digest":            */ "",
					/*"digestText":        */ "",
					/*"querySignature":    */ "d41d8cd98f00b204e9800998ecf8427e",
					/*"countStar":         */ 0,
					/*"sumTimerWait":      */ 0,
					/*"sumLockTime":       */ 0,
					/*"sumErrors":         */ 0,
					/*"sumRowsAffected":   */ 0,
					/*"sumRowsSent":       */ 0,
					/*"sumRowsExamined":   */ 0,
					/*"sumSelectScan":     */ 0,
					/*"sumSelectFullJoin": */ 0,
					/*"sumNoIndexUsed":    */ 0,
					/*"sumNoGoodIndexUsed":*/ 0,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := getCleanSummaryRows(tc.rows)

			for k, v := range tc.expect {
				switch reflect.TypeOf(v).Name() {
				case "string", "uint64":
					tu.Equals(t, v, res[k])
				default:
					t.Logf("%d is type %s", k, reflect.TypeOf(v))
				}
			}
		})
	}
}

//----------------------------------------------------------------------
// mysql_dbm_sample

// go test -v -timeout 30s -run ^TestGetCleanEnabledPerformanceSchemaConsumers$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/mysql
func TestGetCleanEnabledPerformanceSchemaConsumers(t *testing.T) {
	cases := []struct {
		name   string
		rows   *mockRows
		expect map[string]interface{}
	}{
		// mysql 5
		{
			name: "mysql_5_dbm_sample_consumers",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{"events_statements_current"},
					{"events_statements_history"},
					{"events_statements_history_long"},
				},
			},
			expect: map[string]interface{}{
				"events_statements_current":      true,
				"events_statements_history":      true,
				"events_statements_history_long": true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := getCleanEnabledPerformanceSchemaConsumers(tc.rows)
			assert.NotEmpty(t, res, "get empty!")

			for _, v := range res {
				if _, ok := tc.expect[v]; !ok {
					t.Fail()
				}
			}
		})
	}
}

// go test -v -timeout 30s -run ^TestGetCleanMysqlVersion$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/mysql
func TestGetCleanMysqlVersion(t *testing.T) {
	cases := []struct {
		name   string
		rows   *mockRows
		expect *mysqlVersion
	}{
		// mysql 5
		{
			name: "mysql_5_version",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{"5.7.36"},
				},
			},
			expect: &mysqlVersion{
				version: "5.7.36",
				flavor:  "MySQL",
				build:   "unspecified",
			},
		},

		// mysql 8
		{
			name: "mysql_8_version",
			rows: &mockRows{
				t: t,
				data: [][]interface{}{
					{"8.0.27"},
				},
			},
			expect: &mysqlVersion{
				version: "8.0.27",
				flavor:  "MySQL",
				build:   "unspecified",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := getCleanMysqlVersion(tc.rows)

			assert.NotNil(t, res)

			tu.Equals(t, res.version, tc.expect.version)
			tu.Equals(t, res.flavor, tc.expect.flavor)
			tu.Equals(t, res.build, tc.expect.build)
		})
	}
}

//----------------------------------------------------------------------
// mysql_custom_queries

// go test -v -timeout 30s -run ^TestGetCleanMysqlCustomQueries$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/mysql
func TestGetCleanMysqlCustomQueries(t *testing.T) {
	// cases := []struct {
	// 	name   string
	// 	rows   *mockRows
	// 	expect map[string]interface{}
	// }{
	// mysql 5
	// 	{
	// 		name: "mysql_5_custom_queries",
	// 		rows: &mockRows{
	// 			t: t,
	// 			data: [][]interface{}{
	// 				{"table_schema", "mysqlslap"},
	// 				{"table_name", "t1"},
	// 				{"table_type", "BASE TABLE"},
	// 				{"engine", "InnoDB"},
	// 				{"version", "10"},
	// 				{"table_rows", "571"},
	// 				{"data_length", "13156352"},
	// 				{"index_length", "0"},
	// 				{"data_free", "4194304"},
	// 			},
	// 		},
	// 		expect: map[string]interface{}{
	// 			"table_schema": "mysqlslap",
	// 			"table_name":   "t1",
	// 			"table_type":   "BASE TABLE",
	// 			"engine":       "InnoDB",
	// 			"version":      "10",
	// 			"table_rows":   "571",
	// 			"data_length":  "13156352",
	// 			"index_length": "0",
	// 			"data_free":    "4194304",
	// 		},
	// 	},
	// }

	// for _, tc := range cases {
	// 	t.Run(tc.name, func(t *testing.T) {
	// 		res := getCleanSchemaData(tc.rows)

	// 		for k, v := range tc.expect {
	// 			switch x := v.(type) {
	// 			case int64:
	// 				tu.Assert(t, res[k] != nil, "key %s should not be nil", k)
	// 				tu.Equals(t, x, res[k].(int64))
	// 			case string:
	// 				tu.Equals(t, x, res[k].(string))
	// 			default:
	// 				t.Logf("%s is type %s", k, reflect.TypeOf(v))
	// 			}
	// 		}
	// 	})
	// }
}
