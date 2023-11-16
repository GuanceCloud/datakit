// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package coceanbase

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/oceanbase/collect/ccommon"
)

const nullString = "NULL"

type slowQueryLogging struct {
	x         collectParameters
	firstTime string
}

var _ ccommon.DBMetricsCollector = (*slowQueryLogging)(nil)

func newSlowQueryLogging(opts ...collectOption) *slowQueryLogging {
	m := &slowQueryLogging{}

	for _, opt := range opts {
		if opt != nil {
			opt(&m.x)
		}
	}

	return m
}

func (m *slowQueryLogging) Collect() ([]*point.Point, error) {
	l.Debug("Collect entry")

	tfs, err := m.slowQuery()
	if err != nil {
		return nil, err
	}

	if len(tfs) == 0 {
		return nil, nil
	}

	var pts []*point.Point
	for _, tf := range tfs {
		opt := &ccommon.BuildPointOpt{
			TF:         tf,
			MetricName: m.x.MetricName,
			Tags:       m.x.Ipt.tags,
			Host:       m.x.Ipt.host,
		}
		pt := ccommon.BuildPointLogging(l, opt)
		pts = append(pts, pt)
	}
	return pts, nil
}

// GV$SQL_AUDIT
// GV$PLAN_CACHE_PLAN_STAT https://www.oceanbase.com/docs/enterprise-oceanbase-database-cn-10000000000376413

// const SLOW_QUERY = `SELECT
// "TENANT_ID",
// "SVR_IP",
// "SVR_PORT",
// "PLAN_ID",
// "SQL_ID",
// "TYPE",
// "DB_ID",
// "IS_BIND_SENSITIVE",
// "IS_BIND_AWARE",
// "STATEMENT",
// "QUERY_SQL",
// "SPECIAL_PARAMS",
// "PARAM_INFOS",
// "SYS_VARS",
// TO_NUMBER(PLAN_HASH) AS PLAN_HASH,
// "FIRST_LOAD_TIME",
// "SCHEMA_VERSION",
// "MERGED_VERSION",
// "LAST_ACTIVE_TIME",
// "AVG_EXE_USEC",
// "SLOWEST_EXE_TIME",
// "SLOWEST_EXE_USEC",
// "SLOW_COUNT",
// "HIT_COUNT",
// "PLAN_SIZE",
// "EXECUTIONS",
// "DISK_READS",
// "DIRECT_WRITES",
// "BUFFERS_GETS",
// "APPLICATION_WATI_TIME",
// "CONCURRENCY_WAIT_TIME",
// "USER_IO_WAIT_TIME",
// "ROWS_PROCESSED",
// "ELAPSED_TIME",
// "CPU_TIME",
// "LARGE_QUERYS",
// "DELAYED_LARGE_QUERYS",
// "DELAYED_PX_QUERYS",
// "OUTLINE_VERSION",
// "OUTLINE_ID",
// "OUTLINE_DATA",
// "HINTS_INFO",
// "HINTS_ALL_WORKED",
// "ACS_SEL_INFO",
// "TABLE_SCAN",
// "EVOLUTION",
// "EVO_EXECUTIONS",
// "EVO_CPU_TIME",
// "TIMEOUT_COUNT",
// "PS_STMT_ID",
// TO_NUMBER(SESSID) AS SESSID,
// "TEMP_TABLES",
// "IS_USE_JIT",
// "OBJECT_TYPE",
// "PL_SCHEMA_ID",
// "IS_BATCHED_MULTI_STMT"
// FROM GV$PLAN_CACHE_PLAN_STAT
// WHERE LAST_ACTIVE_TIME > '%s' AND LAST_ACTIVE_TIME <= '%s' AND ELAPSED_TIME > %d`

// SLOW_QUERY selects table GV$PLAN_CACHE_PLAN_STAT.
//
// exclueded: EXPECT_WORKER_COUNT (deprecated)
// exclueded: PS_STMT_ID (deprecated)
// exclueded: PARAM_INFOS (too long)
// exclueded: QUERY_SQL (too long)
// exclueded: OUTLINE_DATA (too long)
// exclueded: object_type (too long)
// exclueded: SYS_VARS (too long)
//
//nolint:stylecheck
const SLOW_QUERY = `SELECT
  "TENANT_ID",
  "SVR_IP",
  "SVR_PORT",
  "PLAN_ID",
  "SQL_ID",
  "TYPE",
  "DB_ID",
  SUBSTR(TO_CHAR(STATEMENT), 1, 65535) AS STATEMENT,
  TO_NUMBER(PLAN_HASH) AS PLAN_HASH,
  "LAST_ACTIVE_TIME",
  "ELAPSED_TIME"
FROM GV$PLAN_CACHE_PLAN_STAT
WHERE LAST_ACTIVE_TIME > '%s' AND LAST_ACTIVE_TIME <= '%s' AND ELAPSED_TIME > %d`

// const MYSQL_SLOW_QUERY = `SELECT
//   TENANT_ID AS TENANT_ID,
//   SVR_IP AS SVR_IP,
//   SVR_PORT AS SVR_PORT,
//   PLAN_ID AS PLAN_ID,
//   SQL_ID AS SQL_ID,
//   TYPE AS TYPE,
//   DB_ID AS DB_ID,
//   IS_BIND_SENSITIVE AS IS_BIND_SENSITIVE,
//   IS_BIND_AWARE AS IS_BIND_AWARE,
//   STATEMENT AS STATEMENT,
//   SPECIAL_PARAMS AS SPECIAL_PARAMS,
//   PLAN_HASH AS PLAN_HASH,
//   FIRST_LOAD_TIME AS FIRST_LOAD_TIME,
//   SCHEMA_VERSION AS SCHEMA_VERSION,
//   MERGED_VERSION AS MERGED_VERSION,
//   LAST_ACTIVE_TIME AS LAST_ACTIVE_TIME,
//   AVG_EXE_USEC AS AVG_EXE_USEC,
//   SLOWEST_EXE_TIME AS SLOWEST_EXE_TIME,
//   SLOWEST_EXE_USEC AS SLOWEST_EXE_USEC,
//   SLOW_COUNT AS SLOW_COUNT,
//   HIT_COUNT AS HIT_COUNT,
//   PLAN_SIZE AS PLAN_SIZE,
//   EXECUTIONS AS EXECUTIONS,
//   DISK_READS AS DISK_READS,
//   DIRECT_WRITES AS DIRECT_WRITES,
//   BUFFER_GETS AS BUFFER_GETS,
//   APPLICATION_WAIT_TIME AS APPLICATION_WAIT_TIME,
//   CONCURRENCY_WAIT_TIME AS CONCURRENCY_WAIT_TIME,
//   USER_IO_WAIT_TIME AS USER_IO_WAIT_TIME,
//   ROWS_PROCESSED AS ROWS_PROCESSED,
//   ELAPSED_TIME AS ELAPSED_TIME,
//   CPU_TIME AS CPU_TIME,
//   LARGE_QUERYS AS LARGE_QUERYS,
//   DELAYED_LARGE_QUERYS AS DELAYED_LARGE_QUERYS,
//   DELAYED_PX_QUERYS AS DELAYED_PX_QUERYS,
//   OUTLINE_VERSION AS OUTLINE_VERSION,
//   OUTLINE_ID AS OUTLINE_ID,
//   HINTS_INFO AS HINTS_INFO,
//   HINTS_ALL_WORKED AS HINTS_ALL_WORKED,
//   ACS_SEL_INFO AS ACS_SEL_INFO,
//   TABLE_SCAN AS TABLE_SCAN,
//   EVOLUTION AS EVOLUTION,
//   EVO_EXECUTIONS AS EVO_EXECUTIONS,
//   EVO_CPU_TIME AS EVO_CPU_TIME,
//   TIMEOUT_COUNT AS TIMEOUT_COUNT,
//   PS_STMT_ID AS PS_STMT_ID,
//   SESSID AS SESSID,
//   TEMP_TABLES AS TEMP_TABLES,
//   IS_USE_JIT AS IS_USE_JIT,
//   PL_SCHEMA_ID AS PL_SCHEMA_ID,
//   IS_BATCHED_MULTI_STMT AS IS_BATCHED_MULTI_STMT
// FROM GV$PLAN_CACHE_PLAN_STAT
// WHERE LAST_ACTIVE_TIME > '%s' AND LAST_ACTIVE_TIME <= '%s' AND ELAPSED_TIME > %d`

//nolint:stylecheck
const MYSQL_SLOW_QUERY = `SELECT 
  TENANT_ID AS TENANT_ID,
  SVR_IP AS SVR_IP,
  SVR_PORT AS SVR_PORT,
  PLAN_ID AS PLAN_ID,
  SQL_ID AS SQL_ID,
  TYPE AS TYPE,
  DB_ID AS DB_ID,
  STATEMENT AS STATEMENT,
  PLAN_HASH AS PLAN_HASH,
  LAST_ACTIVE_TIME AS LAST_ACTIVE_TIME,
  ELAPSED_TIME AS ELAPSED_TIME
FROM GV$PLAN_CACHE_PLAN_STAT
WHERE LAST_ACTIVE_TIME > '%s' AND LAST_ACTIVE_TIME <= '%s' AND ELAPSED_TIME > %d`

// SQL_QUERY_MAX_TIME gets the max LAST_ACTIVE_TIME in the table GV$PLAN_CACHE_PLAN_STAT.
//
//nolint:stylecheck
const SQL_QUERY_MAX_TIME = `SELECT 
  TO_CHAR(MAX(LAST_ACTIVE_TIME)) AS MAX_TIME 
FROM GV$PLAN_CACHE_PLAN_STAT`

// SQL_QUERY_MAX_TIME_MYSQL gets the max LAST_ACTIVE_TIME in the table GV$PLAN_CACHE_PLAN_STAT.
//
//nolint:stylecheck
const MYSQL_QUERY_MAX_TIME = `SELECT 
  date_format(MAX(LAST_ACTIVE_TIME),'%Y-%m-%d %H:%i:%S') AS MAX_TIME 
FROM GV$PLAN_CACHE_PLAN_STAT`

func (m *slowQueryLogging) getMaxActiveTime() (string, error) {
	var query string

	//nolint:exhaustive
	switch tenantMode {
	case modeMySQL:
		query = MYSQL_QUERY_MAX_TIME
	case modeOracle:
		query = SQL_QUERY_MAX_TIME
	default:
		return "", fmt.Errorf("not prepared")
	}

	mRes, err := selectMapWrapper(m.x.Ipt, query)
	if err != nil {
		return "", fmt.Errorf("selectMapWrapper() failed: %w", err)
	}

	var maxTime string

	normalizeResultArray(mRes)

	for _, r := range mRes {
		iMaxTime := r["MAX_TIME"]
		val, ok := iMaxTime.(string)
		if !ok {
			err := fmt.Errorf("MAX_TIME not valid: %s", reflect.TypeOf(iMaxTime).String())
			l.Error(err.Error())
			return "", err
		}

		maxTime = val
		break //nolint:staticcheck
	}

	return maxTime, nil
}

func (m *slowQueryLogging) slowQuery() ([]*ccommon.TagField, error) {
	if len(m.firstTime) == 0 {
		var err error
		m.firstTime, err = m.getMaxActiveTime()
		if err != nil {
			l.Errorf("[m.firstTime] getMaxActiveTime failed: %v", err)
			return nil, err
		}

		l.Debugf("First active time = %s", m.firstTime)

		return nil, nil
	}

	thisMaxTime, err := m.getMaxActiveTime()
	if err != nil {
		l.Errorf("[thisMaxTime] getMaxActiveTime failed: %v", err)
		return nil, err
	}

	l.Debugf("This active time = %s", thisMaxTime)

	var query string
	//nolint:exhaustive
	switch tenantMode {
	case modeMySQL:
		query = fmt.Sprintf(MYSQL_SLOW_QUERY, m.firstTime, thisMaxTime, m.x.Ipt.SlowQueryTime.Microseconds())
	case modeOracle:
		query = fmt.Sprintf(SLOW_QUERY, m.firstTime, thisMaxTime, m.x.Ipt.SlowQueryTime.Microseconds())
	}

	mRes, err := selectMapWrapper(m.x.Ipt, query)
	if err != nil {
		return nil, fmt.Errorf("selectMapWrapper() failed: %w", err)
	}

	if len(mRes) == 0 {
		return nil, nil
	}

	mResults := make([]map[string]any, 0)

	normalizeResultArray(mRes)

	for _, r := range mRes {
		l.Debugf("got result row = %#v", r)

		mMapUnit := make(map[string]any, len(r))
		for columnName, columnValue := range r {
			name := strings.ToLower(columnName)
			mMapUnit[name] = columnValue
		}

		// Combine.
		mResults = append(mResults, mMapUnit)
	}

	if len(mResults) == 0 {
		return nil, nil
	}

	tfs := make([]*ccommon.TagField, 0)
	for _, v := range mResults {
		jsn, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		tf := ccommon.NewTagField()
		tf.AddField("status", "warning", nil)
		tf.AddField("message", string(jsn), nil)

		tfs = append(tfs, tf)
	}

	return tfs, nil
}

func normalizeResultArray(in []map[string]interface{}) {
	for k, r := range in {
		for name, val := range r {
			switch tp := val.(type) {
			case []uint8:
				in[k][name] = string(tp)
			case time.Time, int64, string, float64:
			case nil:
				in[k][name] = nullString
			default:
				l.Warnf("%s unhandled type: %s", name, reflect.TypeOf(tp).String())
			}
		}
	}
}
