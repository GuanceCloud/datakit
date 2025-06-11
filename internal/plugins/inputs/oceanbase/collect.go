// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oceanbase

import (
	"crypto/md5" //nolint:gosec
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/obfuscate"
)

const (
	MysqlServerSchemaInfo = `select SVR_IP, SVR_PORT from gv$server_schema_info`
	MysqlShowVariables    = `SHOW VARIABLES`
)

//nolint:stylecheck
type DB_STATUS int

const (
	Down DB_STATUS = iota
	OK
)

func (ipt *Input) collectCustomQuery() (category point.Category, pts []*point.Point, err error) {
	category = point.Metric

	pts = make([]*point.Point, 0)
	for _, item := range ipt.Query {
		arr := getCleanMysqlCustomQueries(ipt.q(item.SQL))
		if len(arr) == 0 {
			continue
		}

		for _, row := range arr {
			fields := make(map[string]interface{})
			tags := make(map[string]string)

			for _, tgKey := range item.Tags {
				if value, ok := row[tgKey]; ok {
					tags[tgKey] = cast.ToString(value)
					delete(row, tgKey)
				}
			}

			for _, fdKey := range item.Fields {
				if value, ok := row[fdKey]; ok {
					// transform all fields to float64
					fields[fdKey] = cast.ToFloat64(value)
				}
			}

			if len(fields) > 0 {
				pts = append(pts, ipt.buildPoint(item.Metric, tags, fields, false))
			}
		}
	}

	return category, pts, nil
}

const SQLStat = `
select 
  /*+ MONITOR_AGENT READ_CONSISTENCY(WEAK) */
  con_id tenant_id, 
  SVR_IP, 
  stat_id, 
  replace(name, " ", "_") as metric_name, 
  value as metric_value 
from 
  gv$sysstat 
where 
  stat_id in (
    40000, 40002, 40004, 40006, 40008, 40018, 
    40001, 40003, 40005, 40007, 40009, 
    40019, 40010, 40011, 40012, 40116, 
    40117, 40118, 20000, 20001, 20002, 
    140005, 140006, 140003, 140002, 130000, 
    130001, 130002, 130004, 10005, 10006, 
    10000, 10002, 10001, 10003, 40030, 
    30005, 30011, 30009, 30007, 30006, 
    30008, 30010, 30012, 30002, 80057, 
    30000, 30001, 80040, 80041, 60022, 
    60021, 60023, 60000, 60003, 60001, 
    60002, 50008, 50009, 50001, 50000, 
    50038, 50037, 50005, 50004, 50011, 
    50010,120000,120001,120008,120009
  ) 
  and (
    con_id > 1000 
    or con_id = 1
  ) 
  and class < 1000;
`

type obStat struct {
	MetricValue sql.NullFloat64 `db:"metric_value"`
	TenantID    sql.NullString  `db:"tenant_id"`
	MetricName  sql.NullString  `db:"metric_name"`
	SvrIP       sql.NullString  `db:"SVR_IP"`
	StatID      sql.NullString  `db:"stat_id"`
}

func (ipt *Input) collectStat() (category point.Category, pts []*point.Point, err error) {
	category = point.Metric
	rows := []obStat{}
	if err = selectWrapper(ipt, &rows, SQLStat); err != nil {
		err = fmt.Errorf("selectWrapper: %w", err)
		return
	}
	pts = make([]*point.Point, 0)
	for _, row := range rows {
		tags := map[string]string{
			"cluster":     ipt.Cluster,
			"tenant_id":   row.TenantID.String,
			"tenant_name": ipt.getTenantNameByID(row.TenantID.String),
			"metric_name": row.MetricName.String,
			"svr_ip":      row.SvrIP.String,
			"stat_id":     row.StatID.String,
		}
		fields := map[string]interface{}{
			"metric_value": row.MetricValue.Float64,
		}

		pts = append(pts, ipt.buildPoint("oceanbase_stat", tags, fields, false))
	}
	return category, pts, nil
}

func (ipt *Input) collectEvent() (category point.Category, pts []*point.Point, err error) {
	category = point.Metric
	rows := []obEvent{}
	if err = selectWrapper(ipt, &rows, SQLEvent); err != nil {
		err = fmt.Errorf("selectWrapper: %w", err)
		return
	}
	pts = make([]*point.Point, 0)
	for _, row := range rows {
		tags := map[string]string{
			"cluster":     ipt.Cluster,
			"tenant_id":   row.TenantID.String,
			"tenant_name": ipt.getTenantNameByID(row.TenantID.String),
			"svr_ip":      row.SvrIP.String,
			"event_group": row.EventGroup.String,
		}
		fields := map[string]interface{}{
			"total_waits": row.TotalWaits.Float64,
			"time_waited": row.TimeWaited.Float64,
		}

		pts = append(pts, ipt.buildPoint("oceanbase_event", tags, fields, false))
	}
	return category, pts, nil
}

const SQLSession = `
select
  /*+ MONITOR_AGENT READ_CONSISTENCY(WEAK) */
  case
    when all_cnt is null then 0
    else all_cnt
  end as all_cnt,
  case
    when active_cnt is null then 0
    else active_cnt
  end as active_cnt,
  tenant_name,
  tenant_id,
  svr_ip,
  svr_port
from
  (
    select
      __all_tenant.tenant_name,
      __all_tenant.tenant_id,
      all_cnt,
      active_cnt,
      svr_ip,
      svr_port
    from
      __all_tenant
      left join (
        select
          count(state = 'ACTIVE' OR NULL) as active_cnt,
          count(1) as all_cnt,
          tenant as tenant_name,
          svr_ip,
          svr_port
        from
          __all_virtual_processlist
        group by
          tenant,
          svr_ip,
          svr_port
      ) t1 on __all_tenant.tenant_name = t1.tenant_name
  ) t2
`

type obSession struct {
	TenantID   sql.NullString `db:"tenant_id"`
	TenantName sql.NullString `db:"tenant_name"`
	ActiveCnt  sql.NullInt64  `db:"active_cnt"`
	AllCnt     sql.NullInt64  `db:"all_cnt"`
	SvrIP      sql.NullString `db:"svr_ip"`
	SvrPort    sql.NullString `db:"svr_port"`
}

func (ipt *Input) collectSession() (category point.Category, pts []*point.Point, err error) {
	category = point.Metric
	rows := []obSession{}
	if err = selectWrapper(ipt, &rows, SQLSession); err != nil {
		err = fmt.Errorf("selectWrapper: %w", err)
		return
	}
	pts = make([]*point.Point, 0)
	for _, row := range rows {
		tags := map[string]string{
			"cluster":     ipt.Cluster,
			"tenant_id":   row.TenantID.String,
			"tenant_name": row.TenantName.String,
			"svr_ip":      row.SvrIP.String,
			"svr_port":    row.SvrPort.String,
		}
		fields := map[string]interface{}{
			"active_cnt": row.ActiveCnt.Int64,
			"all_cnt":    row.AllCnt.Int64,
		}

		pts = append(pts, ipt.buildPoint("oceanbase_session", tags, fields, false))
	}
	return category, pts, nil
}

const SQLPlanCache = `
select 
  /*+ MONITOR_AGENT READ_CONSISTENCY(WEAK) */
  tenant_id, 
  access_count, 
  hit_count, 
  svr_ip,
	svr_port
from 
  gv$plan_cache_stat
`

const SQLBlockCache = `
select
  /*+ MONITOR_AGENT READ_CONSISTENCY(WEAK) */
  tenant_id,
  cache_name,
  cache_size,
  svr_ip,
	svr_port
from
  __all_virtual_kvcache_info
`

type obPlanCache struct {
	TenantID    sql.NullString `db:"tenant_id"`
	AccessCount sql.NullInt64  `db:"access_count"`
	HitCount    sql.NullInt64  `db:"hit_count"`
	SvrIP       sql.NullString `db:"svr_ip"`
	SvrPort     sql.NullString `db:"svr_port"`
}

type obBlockCache struct {
	TenantID  sql.NullString `db:"tenant_id"`
	CacheName sql.NullString `db:"cache_name"`
	CacheSize sql.NullInt64  `db:"cache_size"`
	SvrIP     sql.NullString `db:"svr_ip"`
	SvrPort   sql.NullString `db:"svr_port"`
}

func (ipt *Input) collectCache() (category point.Category, pts []*point.Point, err error) {
	category = point.Metric
	pts = make([]*point.Point, 0)

	// collect plan cache
	{
		rows := []obPlanCache{}
		if err = selectWrapper(ipt, &rows, SQLPlanCache); err != nil {
			l.Warnf("selectWrapper: %s", err.Error())
		} else {
			for _, row := range rows {
				tags := map[string]string{
					"cluster":     ipt.Cluster,
					"tenant_id":   row.TenantID.String,
					"tenant_name": ipt.getTenantNameByID(row.TenantID.String),
					"svr_ip":      row.SvrIP.String,
					"svr_port":    row.SvrPort.String,
				}
				fields := map[string]interface{}{
					"access_count": row.AccessCount.Int64,
					"hit_count":    row.HitCount.Int64,
				}

				pts = append(pts, ipt.buildPoint("oceanbase_cache_plan", tags, fields, false))
			}
		}
	}

	// collcect block cache
	{
		rows := []obBlockCache{}
		if err = selectWrapper(ipt, &rows, SQLBlockCache); err != nil {
			l.Warnf("selectWrapper: %s", err.Error())
		}

		for _, row := range rows {
			tags := map[string]string{
				"cluster":     ipt.Cluster,
				"tenant_id":   row.TenantID.String,
				"tenant_name": ipt.getTenantNameByID(row.TenantID.String),
				"svr_ip":      row.SvrIP.String,
				"svr_port":    row.SvrPort.String,
				"cache_name":  row.CacheName.String,
			}
			fields := map[string]interface{}{
				"cache_size": row.CacheSize.Int64,
			}

			pts = append(pts, ipt.buildPoint("oceanbase_cache_block", tags, fields, false))
		}
	}
	return category, pts, nil
}

type obTenant struct {
	TenantID   sql.NullString `db:"tenant_id"`
	TenantName sql.NullString `db:"tenant_name"`
}

func (ipt *Input) initTenantNames() (err error) {
	rows := []obTenant{}
	if err = selectWrapper[*[]obTenant](ipt, &rows, "select tenant_id, tenant_name from gv$tenant"); err != nil {
		err = fmt.Errorf("selectWrapper: %w", err)
		return
	}

	names := make(map[string]string)
	for _, row := range rows {
		if len(row.TenantID.String) > 0 {
			names[row.TenantID.String] = row.TenantName.String
		}
	}
	ipt.tenantNames = names

	return nil
}

func (ipt *Input) getTenantNameByID(tenantID string) string {
	if len(ipt.tenantNames) > 0 {
		return ipt.tenantNames[tenantID]
	}

	return ""
}

const SQLClog = `  
select
  /*+ MONITOR_AGENT READ_CONSISTENCY(WEAK) */
  stat.table_id >> 40 tenant_id,
  stat.replica_type,
  stat.svr_ip,
  stat.svr_port,
  max(stat.next_replay_ts_delta) / 1000000 as max_clog_sync_delay_seconds
from
  __all_virtual_clog_stat stat
  left join (
    select
      meta.table_id,
      meta.partition_id,
      meta.svr_ip,
      meta.svr_port
    from
      __all_virtual_meta_table meta
    where
      meta.status = 'REPLICA_STATUS_NORMAL'
      and meta.table_id not in (
        select
          table_id
        from
          __all_virtual_partition_migration_status mig
        where
          mig.action <> 'END'
      )
  ) meta on stat.table_id = meta.table_id
  and stat.partition_idx = meta.partition_id
  and stat.svr_ip = meta.svr_ip
  and stat.svr_port = meta.svr_port
group by
  tenant_id,
  replica_type,
  stat.svr_ip,
  stat.svr_port
`

type obClog struct {
	TenantID                sql.NullString  `db:"tenant_id"`
	ReplicaType             sql.NullString  `db:"replica_type"`
	SvrIP                   sql.NullString  `db:"svr_ip"`
	SvrPort                 sql.NullInt64   `db:"svr_port"`
	MaxClogSyncDelaySeconds sql.NullFloat64 `db:"max_clog_sync_delay_seconds"`
}

func (ipt *Input) collectClog() (category point.Category, pts []*point.Point, err error) {
	category = point.Metric
	rows := []obClog{}

	pts = make([]*point.Point, 0)
	if err = selectWrapper(ipt, &rows, SQLClog); err != nil {
		err = fmt.Errorf("selectWrapper: %w", err)
		return
	}
	for _, row := range rows {
		tags := map[string]string{
			"cluster":      ipt.Cluster,
			"tenant_id":    row.TenantID.String,
			"tenant_name":  ipt.getTenantNameByID(row.TenantID.String),
			"svr_ip":       row.SvrIP.String,
			"svr_port":     fmt.Sprint(row.SvrPort.Int64),
			"replica_type": row.ReplicaType.String,
		}
		fields := map[string]interface{}{
			"max_clog_sync_delay_seconds": row.MaxClogSyncDelaySeconds.Float64,
		}

		pts = append(pts, ipt.buildPoint("oceanbase_clog", tags, fields, false))
	}
	return category, pts, nil
}

func (ipt *Input) q(s string) rows {
	rows, err := ipt.db.Query(s)
	if err != nil {
		l.Errorf(`query failed, sql (%q), error: %s, ignored`, s, err.Error())
		return nil
	}

	if err := rows.Err(); err != nil {
		closeRows(rows)
		l.Errorf(`query row failed, sql (%q), error: %s, ignored`, s, err.Error())
		return nil
	}

	return rows
}

type rows interface {
	Next() bool
	Scan(...interface{}) error
	Close() error
	Err() error
	Columns() ([]string, error)
}

func closeRows(r rows) {
	if err := r.Close(); err != nil {
		l.Warnf("Close: %s, ignored", err)
	}
}

func getCleanMysqlCustomQueries(r rows) []map[string]interface{} {
	l.Debugf("getCleanMysqlCustomQueries entry")

	if r == nil {
		l.Debug("r == nil")
		return nil
	}

	defer closeRows(r)

	var list []map[string]interface{}

	columns, err := r.Columns()
	if err != nil {
		l.Errorf("Columns() failed: %v", err)
	}
	l.Debugf("columns = %v", columns)
	columnLength := len(columns)
	l.Debugf("columnLength = %d", columnLength)

	cache := make([]interface{}, columnLength)
	for idx := range cache {
		var a interface{}
		cache[idx] = &a
	}

	for r.Next() {
		l.Debug("Next() entry")

		if err := r.Scan(cache...); err != nil {
			l.Errorf("Scan failed: %v", err)
		}

		l.Debugf("len(cache) = %d", len(cache))

		item := make(map[string]interface{})
		for i, data := range cache {
			key := columns[i]
			val := *data.(*interface{})

			if val != nil {
				vType := reflect.TypeOf(val)

				l.Debugf("key = %s, vType = %s, %s", key, vType.String(), vType.Name())

				switch vType.String() {
				case "int64":
					if v, ok := val.(int64); ok {
						item[key] = v
					} else {
						l.Warn("expect int64, ignored")
					}
				case "string":
					var data interface{}
					data, err := strconv.ParseFloat(val.(string), 64)
					if err != nil {
						data = val
					}
					item[key] = data
				case "time.Time":
					if v, ok := val.(time.Time); ok {
						item[key] = v
					} else {
						l.Warn("expect time.Time, ignored")
					}
				case "[]uint8":
					item[key] = string(val.([]uint8))
				default:
					l.Warn("unsupport data type '%s', ignored", vType)
				}
			}
		}

		list = append(list, item)
	}

	if err := r.Err(); err != nil {
		l.Errorf("Err() failed: %v", err)
	}

	l.Debugf("len(list) = %d", len(list))

	return list
}

func GetMD5String32(bt []byte) string {
	return fmt.Sprintf("%X", md5.Sum(bt)) // nolint:gosec
}

const MySQLSlowQuery = `SELECT 
	 /*+ QUERY_TIMEOUT(60000000) */
  TENANT_ID,
	TENANT_NAME,
	DB_NAME,
	CLIENT_IP,
	USER_CLIENT_IP,
  SVR_IP,
  SVR_PORT,
  PLAN_ID,
  SQL_ID,
	IS_INNER_SQL,
  DB_ID,
	QUERY_SQL,
	USER_NAME,
	REQUEST_TIME,
  ELAPSED_TIME
FROM GV$SQL_AUDIT
WHERE REQUEST_TIME  > %d AND ELAPSED_TIME > %d`

func (ipt *Input) collectSlowQuery() (category point.Category, pts []*point.Point, err error) {
	category = point.Logging

	l.Debugf("This request time = %s", ipt.slowQueryStartTime)
	query := fmt.Sprintf(MySQLSlowQuery, ipt.slowQueryStartTime.UnixMicro(), ipt.slowQueryTime.Microseconds())

	mRes, err := selectMapWrapper(ipt, query)
	if err != nil {
		err = fmt.Errorf("selectMapWrapper() failed: %w", err)
		return
	}
	ipt.slowQueryStartTime = time.Now()

	if len(mRes) == 0 {
		return
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
		return
	}

	pts = make([]*point.Point, 0)
	for _, v := range mResults {
		if sql, ok := v["query_sql"]; ok {
			if sqlStr, ok := sql.(string); ok {
				v["query_sql"] = obfuscateSQL(sqlStr)
			}
		}
		jsn, err := json.Marshal(v)
		if err != nil {
			l.Warnf("Marshal json failed: %s, ignore this result", err.Error())
			continue
		}
		tags := map[string]string{
			"cluster": ipt.Cluster,
		}

		fields := map[string]interface{}{
			"status":  "warning",
			"message": string(jsn),
		}

		pts = append(pts, ipt.buildPoint("oceanbase_log", tags, fields, true))
	}

	return category, pts, nil
}

func (ipt *Input) buildPoint(name string, tags map[string]string, fields map[string]interface{}, isLogging bool) *point.Point {
	var opts []point.Option

	if isLogging {
		opts = point.DefaultLoggingOptions()
	} else {
		opts = point.DefaultMetricOptions()
	}

	opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	opts = append(opts, point.WithTime(ipt.ptsTime))

	kvs := point.NewTags(tags)

	// add extended tags
	for k, v := range ipt.mergedTags {
		kvs = kvs.AddTag(k, v)
	}

	for k, v := range fields {
		kvs = kvs.Add(k, v, false, true)
	}

	return point.NewPointV2(name, kvs, opts...)
}

func selectMapWrapper(ipt *Input, sqlText string) ([]map[string]interface{}, error) {
	now := time.Now()
	mRet, err := selectMap(ipt, sqlText)
	l.Debugf("executed sql: %s, cost: %v, length: %d, err: %v\n", sqlText, time.Since(now), len(mRet), err)
	return mRet, err
}

func selectMap(ipt *Input, sqlText string) ([]map[string]interface{}, error) {
	var err error
	var rows *sql.Rows
	var cols []string

	defer func() {
		if err != nil {
			l.Errorf("SQL failed, err = %v, sql = %s", err, sqlText)
		}
	}()

	rows, err = ipt.db.Query(sqlText)
	if err != nil {
		l.Errorf("db.Query() failed: %v", err)
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	cols, err = rows.Columns()
	if err != nil {
		l.Errorf("rows.Columns() failed: %v", err)
		return nil, err
	}

	mRet := make([]map[string]interface{}, 0)

	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err = rows.Scan(columnPointers...); err != nil {
			l.Errorf("Scan() failed: %v", err)
			return nil, err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}

		// Outputs: map[columnName:value columnName2:value2 columnName3:value3 ...]
		mRet = append(mRet, m)
	}

	if err = rows.Err(); err != nil {
		l.Errorf("rows.Err() failed: %v", err)
		return nil, err
	}

	return mRet, nil
}

func selectWrapper[T any](ipt *Input, s T, sql string) error {
	now := time.Now()

	err := ipt.db.Select(s, sql)
	if err != nil && (strings.Contains(err.Error(), "ORA-01012") || strings.Contains(err.Error(), "database is closed")) {
		if err := ipt.initDBConnect(); err != nil {
			_ = ipt.db.Close()
		}
	}

	if err != nil {
		l.Errorf("executed sql: %s, cost: %v, err: %v\n", sql, time.Since(now), err)
	} else {
		l.Debugf("executed sql: %s, cost: %v, err: %v\n", sql, time.Since(now), err)
	}

	return err
}

const SQLLock = `SELECT
  COUNT(1), MAX(CTIME)
FROM GV$LOCK`

const SQLConcurrentLimitSQL = `SELECT
  COUNT(1)
FROM GV$CONCURRENT_LIMIT_SQL`

const SQLInstance = `SELECT
  HOST_NAME,
  VERSION,
  DATABASE_STATUS
FROM GV$INSTANCE`

const SQLMemory = `SELECT
  SUM(COUNT),
  SUM(USED)
FROM GV$MEMORY`

const SQLMemstore = `SELECT
case WHEN(NVL(SUM(TOTAL), 0))!=0 THEN
  round(SUM(ACTIVE) / SUM(TOTAL), 3)
ELSE 0
END
RESULT
FROM GV$MEMSTORE`

const SQLObSQLWorkareaMemoryInfo = `SELECT 
MAX_AUTO_WORKAREA_SIZE,
MEM_TARGET,
GLOBAL_MEM_BOUND
FROM GV$OB_SQL_WORKAREA_MEMORY_INFO`

const SQLPlanCacheStat = `SELECT 
SUM(MEM_USED),
SUM(MEM_HOLD),
AVG(HIT_RATE),
SUM(PLAN_NUM)
FROM GV$PLAN_CACHE_STAT`

const SQLPsStat = `SELECT 
case WHEN(NVL(SUM(ACCESS_COUNT), 0))!=0 THEN 
  round(SUM(HIT_COUNT) / SUM(ACCESS_COUNT), 3) 
ELSE 0 
END 
RESULT 
FROM GV$PS_STAT`

const SQLSessionWait = `SELECT 
AVG(WAIT_TIME_MICRO)
FROM GV$SESSION_WAIT`

//nolint:stylecheck
const SQLEvent = `
select 
  /*+ MONITOR_AGENT READ_CONSISTENCY(WEAK) */
  con_id tenant_id, 
  SVR_IP, 
  case when event_id = 10000 then 'INTERNAL' when event_id = 13000 then 'SYNC_RPC' when event_id = 14003 then 'ROW_LOCK_WAIT' when (
    event_id >= 10001 
    and event_id <= 11006
  ) 
  or (
    event_id >= 11008 
    and event_id <= 11011
  ) then 'IO' when event like 'latch:%' then 'LATCH' else 'OTHER' END event_group, 
  sum(total_waits) as total_waits, 
  sum(time_waited_micro / 1000000) as time_waited 
from 
  gv$system_event 
where 
  gv$system_event.wait_class <> 'IDLE' 
  and (
    con_id > 1000 
    or con_id = 1
  ) 
group by 
  tenant_id, 
  event_group, 
  SVR_IP
`

type obEvent struct {
	TenantID   sql.NullString  `db:"tenant_id"`
	SvrIP      sql.NullString  `db:"SVR_IP"`
	EventGroup sql.NullString  `db:"event_group"`
	TotalWaits sql.NullFloat64 `db:"total_waits"`
	TimeWaited sql.NullFloat64 `db:"time_waited"`
}

func normalizeResultArray(in []map[string]interface{}) {
	for k, r := range in {
		for name, val := range r {
			switch tp := val.(type) {
			case []uint8:
				if v, err := strconv.ParseFloat(string(tp), 64); err != nil {
					l.Debugf("parse float err:%s, using string instead", err.Error())
					in[k][name] = string(tp)
				} else {
					in[k][name] = v
				}
			case time.Time, int64, string, float64:
			case nil:
				in[k][name] = "NULL"
			default:
				l.Warnf("%s unhandled type: %s", name, reflect.TypeOf(tp).String())
			}
		}
	}
}

var reg = regexp.MustCompile(`\n|\s+`)

func obfuscateSQL(text string) (sql string) {
	defer func() {
		sql = strings.TrimSpace(reg.ReplaceAllString(sql, " "))
	}()

	if out, err := obfuscate.NewObfuscator(nil).Obfuscate("sql", text); err != nil {
		return fmt.Sprintf("ERROR: failed to obfuscate: %s", err.Error())
	} else {
		return out.Query
	}
}
