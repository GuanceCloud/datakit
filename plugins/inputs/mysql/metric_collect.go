package mysql

import (
	"errors"
	"fmt"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hashcode"
)

// collect metric

func (i *Input) collectMysql() error {
	var err error
	defer func() {
		if err != nil {
			i.globalStatus = map[string]interface{}{}
			i.globalVariables = map[string]interface{}{}
			i.binlog = map[string]interface{}{}
		}
	}()

	// We should first collect global MySQL metrics
	if res := globalStatusMetrics(i.q("SHOW /*!50002 GLOBAL */ STATUS;")); res != nil {
		i.globalStatus = res
	} else {
		err = fmt.Errorf("collect_show_status_failed")
		return err
	}

	if res := globalVariablesMetrics(i.q("SHOW GLOBAL VARIABLES;")); res != nil {
		i.globalVariables = res

		// Detect if binlog enabled
		switch v := i.globalVariables["log_bin"].(type) {
		case string:
			i.binLogOn = (v == "on" || v == "ON")
		default:
			i.binLogOn = false
		}
	} else {
		err = fmt.Errorf("collect_show_variables_failed")
		return err
	}

	if i.binLogOn {
		if res := binlogMetrics(i.q("SHOW BINARY LOGS;")); res != nil {
			i.binlog = res
		} else {
			err = fmt.Errorf("collect_show_binlog_failed")
			return err
		}
	}

	return err
}

func (i *Input) collectMysqlSchema() error {
	var err error
	defer func() {
		if err != nil {
			i.mSchemaSize = map[string]interface{}{}
			i.mSchemaQueryExecTime = map[string]interface{}{}
		}
	}()

	querySizePerschemaSQL := `
		SELECT   table_schema, IFNULL(SUM(data_length+index_length)/1024/1024,0) AS total_mb
		FROM     information_schema.tables
		GROUP BY table_schema;
	`
	if res := getCleanSchemaData(i.q(querySizePerschemaSQL)); res != nil {
		i.mSchemaSize = res
	} else {
		err = fmt.Errorf("collect_schema_size_failed")
		return err
	}

	queryExecPerTimeSQL := `
	SELECT schema_name, ROUND((SUM(sum_timer_wait) / SUM(count_star)) / 1000000) AS avg_us
	FROM performance_schema.events_statements_summary_by_digest
	WHERE schema_name IS NOT NULL
	GROUP BY schema_name;
	`
	if res := getCleanSchemaData(i.q(queryExecPerTimeSQL)); res != nil {
		i.mSchemaQueryExecTime = res
	} else {
		err = fmt.Errorf("collect_schema_failed")
		return err
	}

	return err
}

func (i *Input) collectMysqlInnodb() error {
	var err error
	defer func() {
		if err != nil {
			i.mInnodb = map[string]interface{}{}
		}
	}()

	globalInnodbSQL := `SELECT NAME, COUNT FROM information_schema.INNODB_METRICS WHERE status='enabled'`

	if res := getCleanInnodb(i.q(globalInnodbSQL)); res != nil {
		i.mInnodb = res
	} else {
		err = fmt.Errorf("collect_innodb_failed")
		return err
	}

	return err
}

func (i *Input) collectMysqlTableSchema() error {
	var err error
	defer func() {
		if err != nil {
			i.mTableSchema = []map[string]interface{}{}
		}
	}()

	tableSchemaSQL := `
	SELECT
        TABLE_SCHEMA,
        TABLE_NAME,
        TABLE_TYPE,
        ifnull(ENGINE, 'NONE') as ENGINE,
        ifnull(VERSION, '0') as VERSION,
        ifnull(ROW_FORMAT, 'NONE') as ROW_FORMAT,
        ifnull(TABLE_ROWS, '0') as TABLE_ROWS,
        ifnull(DATA_LENGTH, '0') as DATA_LENGTH,
        ifnull(INDEX_LENGTH, '0') as INDEX_LENGTH,
        ifnull(DATA_FREE, '0') as DATA_FREE
    FROM information_schema.tables
    WHERE TABLE_SCHEMA NOT IN ('mysql', 'performance_schema', 'information_schema', 'sys')
	`

	if len(i.Tables) > 0 {
		var arr []string
		for _, table := range i.Tables {
			arr = append(arr, fmt.Sprintf("'%s'", table))
		}

		filterStr := strings.Join(arr, ",")
		tableSchemaSQL = fmt.Sprintf("%s and TABLE_NAME in (%s);", tableSchemaSQL, filterStr)
	}

	if res := getCleanTableSchema(i.q(tableSchemaSQL)); res != nil {
		i.mTableSchema = res
	} else {
		err = fmt.Errorf("collect_table_schema_failed")
		return err
	}

	return err
}

func (i *Input) collectMysqlUserStatus() error {
	var err error
	defer func() {
		if err != nil {
			i.mUserStatusName = map[string]interface{}{}
			i.mUserStatusVariable = map[string]map[string]interface{}{}
			i.mUserStatusConnection = map[string]map[string]interface{}{}
		}
	}()

	userSQL := `select DISTINCT(user) from mysql.user`

	if len(i.Users) > 0 {
		var arr []string
		for _, user := range i.Users {
			arr = append(arr, fmt.Sprintf("'%s'", user))
		}

		filterStr := strings.Join(arr, ",")
		userSQL = fmt.Sprintf("%s where user in (%s);", userSQL, filterStr)
	}

	if res := getCleanUserStatusName(i.q(userSQL)); res != nil {
		i.mUserStatusName = res
	} else {
		err = fmt.Errorf("collect_user_name_failed")
		return err
	}

	userQuerySQL := `
	select VARIABLE_NAME, VARIABLE_VALUE
	from performance_schema.status_by_user
	where user='%s';
	`

	userConnSQL := `select USER, CURRENT_CONNECTIONS, TOTAL_CONNECTIONS
	from performance_schema.users
	where user = '%s';
    `

	for user := range i.mUserStatusName {
		if res := getCleanUserStatusVariable(i.q(fmt.Sprintf(userQuerySQL, user))); res != nil {
			i.mUserStatusVariable = make(map[string]map[string]interface{})
			i.mUserStatusVariable[user] = res
		}

		if res := getCleanUserStatusConnection(i.q(fmt.Sprintf(userConnSQL, user))); res != nil {
			i.mUserStatusConnection = make(map[string]map[string]interface{})
			i.mUserStatusConnection[user] = res
		}
	}

	if len(i.mUserStatusVariable) == 0 {
		err = fmt.Errorf("collect_user_variable_failed")
		return err
	}

	if len(i.mUserStatusConnection) == 0 {
		err = fmt.Errorf("collect_user_connection_failed")
		return err
	}

	return err
}

func (i *Input) collectMysqlCustomQueries() error {
	var err error
	defer func() {
		if err != nil {
			i.mCustomQueries = map[string][]map[string]interface{}{}
		}
	}()

	for _, item := range i.Query {
		arr := getCleanMysqlCustomQueries(i.q(item.sql))
		if arr == nil {
			continue
		}
		hs := hashcode.GetMD5String32([]byte(item.sql))
		i.mCustomQueries[hs] = make([]map[string]interface{}, 0)
		i.mCustomQueries[hs] = arr
	}

	return err
}

func (i *Input) collectMysqlDbmMetric() error {
	var err error
	defer func() {
		if err != nil {
			i.dbmMetricRows = []dbmRow{}

			// dbmCache cannot reset
		}
	}()

	statementSummarySQL := `
		SELECT schema_name,digest,digest_text,count_star,
		sum_timer_wait,sum_lock_time,sum_errors,sum_rows_affected,
		sum_rows_sent,sum_rows_examined,sum_select_scan,sum_select_full_join,
		sum_no_index_used,sum_no_good_index_used
		FROM performance_schema.events_statements_summary_by_digest
		WHERE digest_text NOT LIKE 'EXPLAIN %' OR digest_text IS NULL
		ORDER BY count_star DESC LIMIT 10000`

	dbmRows := getCleanSummaryRows(i.q(statementSummarySQL))
	if dbmRows == nil {
		err = fmt.Errorf("collect_summary_rows_failed")
		return err
	}

	metricRows, newDbmCache := getMetricRows(dbmRows, &i.dbmCache)
	i.dbmMetricRows = metricRows

	// save dbm rows
	i.dbmCache = newDbmCache

	return err
}

// mysql_dbm_sample
//----------------------------------------------------------------------

func (i *Input) collectMysqlDbmSample() error {
	var err error
	defer func() {
		if err != nil {
			i.dbmSamplePlans = []planObj{}
		}
	}()

	if len(i.dbmSampleCache.globalStatusTable) == 0 {
		if len(i.dbmSampleCache.version.version) == 0 {
			const sqlSelect = "SELECT VERSION();"
			version := getCleanMysqlVersion(i.q(sqlSelect))
			if version == nil {
				err = errors.New("version_nil")
				return err
			}
			i.dbmSampleCache.version = *version
		}

		if i.dbmSampleCache.version.flavor == strMariaDB || !(i.dbmSampleCache.version.versionCompatible([]int{5, 7, 0})) {
			i.dbmSampleCache.globalStatusTable = "information_schema.global_status"
		} else {
			i.dbmSampleCache.globalStatusTable = "performance_schema.global_status"
		}
	}

	strategy, err := getSampleCollectionStrategy(i)
	if err != nil {
		return err
	}

	rows, err := getNewEventsStatements(i, strategy.table, 5000)
	if err != nil {
		return err
	}

	rows = filterValidStatementRows(i, rows)

	plans := collectPlanForStatements(i, rows)
	if len(plans) > 0 {
		i.dbmSamplePlans = plans
	}

	return nil
}

//----------------------------------------------------------------------
