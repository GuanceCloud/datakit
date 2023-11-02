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
	"regexp"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/obfuscate"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/oceanbase/collect/ccommon"
)

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

func (m *slowQueryLogging) Collect() (*point.Point, error) {
	l.Debug("Collect entry")

	tf, err := m.slowQuery()
	if err != nil {
		return nil, err
	}

	if tf == nil {
		return nil, nil
	}

	if tf.IsEmpty() {
		return nil, fmt.Errorf("ob logs empty")
	}

	opt := &ccommon.BuildPointOpt{
		TF:         tf,
		MetricName: m.x.MetricName,
		Tags:       m.x.Ipt.tags,
		Host:       m.x.Ipt.host,
	}
	return ccommon.BuildPointLogging(l, opt), nil
}

// GV$SQL_AUDIT
// GV$PLAN_CACHE_PLAN_STAT

// SLOW_QUERY selects table GV$PLAN_CACHE_PLAN_STAT.
//
// exclueded: EXPECT_WORKER_COUNT (deprecated)
// exclueded: PS_STMT_ID (deprecated)
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
"IS_BIND_SENSITIVE",
"IS_BIND_AWARE",
"STATEMENT",
"QUERY_SQL",
"SPECIAL_PARAMS",
"PARAM_INFOS",
"SYS_VARS",
TO_NUMBER(PLAN_HASH) AS PLAN_HASH,
"FIRST_LOAD_TIME",
"SCHEMA_VERSION",
"MERGED_VERSION",
"LAST_ACTIVE_TIME",
"AVG_EXE_USEC",
"SLOWEST_EXE_TIME",
"SLOWEST_EXE_USEC",
"SLOW_COUNT",
"HIT_COUNT",
"PLAN_SIZE",
"EXECUTIONS",
"DISK_READS",
"DIRECT_WRITES",
"BUFFERS_GETS",
"APPLICATION_WATI_TIME",
"CONCURRENCY_WAIT_TIME",
"USER_IO_WAIT_TIME",
"ROWS_PROCESSED",
"ELAPSED_TIME",
"CPU_TIME",
"LARGE_QUERYS",
"DELAYED_LARGE_QUERYS",
"DELAYED_PX_QUERYS",
"OUTLINE_VERSION",
"OUTLINE_ID",
"OUTLINE_DATA",
"HINTS_INFO",
"HINTS_ALL_WORKED",
"ACS_SEL_INFO",
"TABLE_SCAN",
"EVOLUTION",
"EVO_EXECUTIONS",
"EVO_CPU_TIME",
"TIMEOUT_COUNT",
"PS_STMT_ID",
TO_NUMBER(SESSID) AS SESSID,
"TEMP_TABLES",
"IS_USE_JIT",
"OBJECT_TYPE",
"PL_SCHEMA_ID",
"IS_BATCHED_MULTI_STMT"
FROM GV$PLAN_CACHE_PLAN_STAT
WHERE LAST_ACTIVE_TIME > '%s' AND LAST_ACTIVE_TIME <= '%s' AND ELAPSED_TIME > %d`

// SQL_QUERY_MAX_TIME gets the max LAST_ACTIVE_TIME in the table GV$PLAN_CACHE_PLAN_STAT.
//
//nolint:stylecheck
const SQL_QUERY_MAX_TIME = `SELECT 
  TO_CHAR(MAX(LAST_ACTIVE_TIME)) AS MAX_TIME 
FROM GV$PLAN_CACHE_PLAN_STAT`

func (m *slowQueryLogging) getMaxActiveTime() (string, error) {
	mRes, err := selectMapWrapper(m.x.Ipt, SQL_QUERY_MAX_TIME)
	if err != nil {
		return "", fmt.Errorf("selectMapWrapper() failed: %w", err)
	}

	var maxTime string

	for _, r := range mRes {
		switch tp := r["MAX_TIME"].(type) {
		case string:
			if len(tp) == 0 {
				l.Errorf("MAX_TIME empty")
				continue
			}

			maxTime = tp
		case nil:
			l.Warnf("MAX_TIME NULL")
			continue
		default:
			l.Errorf("MAX_TIME not valid: %s", reflect.TypeOf(tp).String())
			continue
		}
	}

	return maxTime, nil
}

func (m *slowQueryLogging) slowQuery() (*ccommon.TagField, error) {
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

	tf := ccommon.NewTagField()

	query := fmt.Sprintf(SLOW_QUERY, m.firstTime, thisMaxTime, m.x.Ipt.SlowQueryTime.Microseconds())

	mRes, err := selectMapWrapper(m.x.Ipt, query)
	if err != nil {
		return nil, fmt.Errorf("selectMapWrapper() failed: %w", err)
	}

	l.Debugf("got rows: %d", len(mRes))

	if len(mRes) == 0 {
		return nil, nil
	}

	tf.AddField("status", "warning", nil)

	mResults := make([]map[string]any, 0)

	for _, r := range mRes {
		l.Debugf("got result row = %#v", r)

		// Check QUERY_SQL.
		var querySQL string
		switch tp := r["QUERY_SQL"].(type) {
		case string:
			if len(tp) == 0 {
				l.Warnf("QUERY_SQL empty")
				continue
			}

			querySQL = tp
		case nil:
			l.Warnf("QUERY_SQL NULL")
			continue
		default:
			l.Errorf("QUERY_SQL not valid: %s", reflect.TypeOf(tp).String())
			continue
		}

		// Processing fields.
		l.Debugf("passed row = %#v", r)

		mMapUnit := make(map[string]any, len(r))
		for columnName, columnValue := range r {
			switch columnName {
			case "QUERY_SQL":
				fullText, err := obfuscateSQL(querySQL)
				if err != nil {
					mMapUnit["failed_obfuscate"] = err.Error()
				}
				mMapUnit["query_sql"] = fullText

			default:
				name := strings.ToLower(columnName)
				if columnValue == nil {
					mMapUnit[name] = "NULL"
				} else {
					mMapUnit[name] = columnValue
				}
			}
		}

		// Combine.
		mResults = append(mResults, mMapUnit)
	}

	if len(mResults) == 0 {
		return nil, nil
	}

	jsn, err := json.Marshal(mResults)
	if err != nil {
		return nil, err
	}

	tf.AddField("message", string(jsn), nil)

	return tf, nil
}

func obfuscateSQL(text string) (string, error) {
	reg, err := regexp.Compile(`\n|\s+`) //nolint:gocritic
	if err != nil {
		l.Debugf("Failed to obfuscate, err: %s \n", err.Error())
		return text, err
	}

	sql := strings.TrimSpace(reg.ReplaceAllString(text, " "))

	if out, err := obfuscate.NewObfuscator(nil).Obfuscate("sql", sql); err != nil {
		l.Debugf("Failed to obfuscate, err: %s \n", err.Error())
		return text, err
	} else {
		return out.Query, nil
	}
}
