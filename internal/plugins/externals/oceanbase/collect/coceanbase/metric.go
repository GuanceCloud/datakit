// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package coceanbase

import (
	"database/sql"
	"fmt"
	"math"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/oceanbase/collect/ccommon"
)

var (
	errEmptyRow    = fmt.Errorf("empty row")
	errUnknownMode = fmt.Errorf("unknown mode")
)

//nolint:stylecheck
type TENANT_MODE int

const (
	modeDefault TENANT_MODE = iota
	modeMySQL
	modeOracle
)

var tenantMode = modeDefault

func getTenantModeName() string {
	//nolint:exhaustive
	switch tenantMode {
	case modeMySQL:
		return "MySQL"
	case modeOracle:
		return "Oracle"
	}

	return "Unknown"
}

//nolint:stylecheck
type DB_STATUS int

const (
	Down DB_STATUS = iota
	OK
)

type dbState struct {
	Hostname string
	Version  string
	Status   DB_STATUS
}

////////////////////////////////////////////////////////////////////////////////

type obMetrics struct {
	x collectParameters
}

var _ ccommon.DBMetricsCollector = (*obMetrics)(nil)

func newOBMetrics(opts ...collectOption) *obMetrics {
	m := &obMetrics{}

	for _, opt := range opts {
		if opt != nil {
			opt(&m.x)
		}
	}

	return m
}

func (m *obMetrics) Collect() ([]*point.Point, error) {
	l.Debug("Collect entry")

	var pts []*point.Point
	collectMap := map[string]func() ([]*ccommon.TagField, error){
		metricName:      m.collectOB,
		statMetricName:  m.collectStat,
		eventMetricName: m.collectEvent,
		cacheMetricName: m.collectCache,
	}

	for metric, fn := range collectMap {
		tfs, err := fn()
		if err != nil {
			l.Warnf("collect %s error: %s", m, err.Error())
			continue
		}

		for _, tf := range tfs {
			if tf != nil && !tf.IsEmpty() {
				opt := &ccommon.BuildPointOpt{
					TF:         tf,
					MetricName: metric,
					Tags:       m.x.Ipt.tags,
					Host:       m.x.Ipt.host,
				}
				pts = append(pts, ccommon.BuildPoint(l, opt))
			}
		}
	}
	return pts, nil
}

// SQL_LOCK selects table GV$LOCK.
//
//nolint:stylecheck
const SQL_LOCK = `SELECT
  COUNT(1), MAX(CTIME)
FROM GV$LOCK`

type obLock struct {
	Count    sql.NullInt64 `db:"COUNT(1)"`   // 数据库行锁的个数。
	MaxCtime sql.NullInt64 `db:"MAX(CTIME)"` // 数据库最大加锁耗时。单位: 秒。
}

// SQL_CONCURRENT_LIMIT_SQL selects table GV$CONCURRENT_LIMIT_SQL.
//
//nolint:stylecheck
const SQL_CONCURRENT_LIMIT_SQL = `SELECT
  COUNT(1)
FROM GV$CONCURRENT_LIMIT_SQL`

type obConcurrentLimitSQL struct {
	Count sql.NullInt64 `db:"COUNT(1)"` // 被限流的 SQL 个数。
}

// SQL_INSTANCE selects table GV$INSTANCE.
//
//nolint:stylecheck
const SQL_INSTANCE = `SELECT
  HOST_NAME,
  VERSION,
  DATABASE_STATUS
FROM GV$INSTANCE`

type obInstance struct {
	Hostname       sql.NullString `db:"HOST_NAME"`       // 数据库实例所在的 Server 地址。
	Version        sql.NullString `db:"VERSION"`         // 数据库实例的版本。
	DatabaseStatus sql.NullString `db:"DATABASE_STATUS"` // 数据库的状态 (如: active=0)。
}

// SQL_MEMORY selects table GV$MEMORY.
//
//nolint:stylecheck
const SQL_MEMORY = `SELECT
  SUM(COUNT),
  SUM(USED)
FROM GV$MEMORY`

type obMemory struct {
	SumCount sql.NullInt64 `db:"SUM(COUNT)"` // 所有租户使用中的内存单元个数。
	SumUsed  sql.NullInt64 `db:"SUM(USED)"`  // 所有租户当前使用的内存数值。单位: Byte。
}

// SQL_MEMSTORE selects table GV$MEMSTORE.
//
//nolint:stylecheck
const SQL_MEMSTORE = `SELECT
case WHEN(NVL(SUM(TOTAL), 0))!=0 THEN
  round(SUM(ACTIVE) / SUM(TOTAL), 3)
ELSE 0
END
RESULT
FROM GV$MEMSTORE`

type obMemStore struct {
	ActiveRate sql.NullFloat64 `db:"RESULT"` // 所有服务器上所有租户的 Memtable 的内存活跃率。
}

// SQL_OB_SQL_WORKAREA_MEMORY_INFO selects table GV$OB_SQL_WORKAREA_MEMORY_INFO.
//
//nolint:stylecheck
const SQL_OB_SQL_WORKAREA_MEMORY_INFO = `SELECT 
MAX_AUTO_WORKAREA_SIZE,
MEM_TARGET,
GLOBAL_MEM_BOUND
FROM GV$OB_SQL_WORKAREA_MEMORY_INFO`

type obSQLWorkareaMemoryInfo struct {
	// auto (workarea 的策略之一，另一是 MANUAL，
	// 详见：https://www.oceanbase.com/docs/enterprise-oceanbase-database-cn-10000000000376699) 时，
	// 预计最大可用内存大小，表示当前workarea情况下，auto 管理的最大内存大小。
	MaxAutoWorkareaSize sql.NullInt64 `db:"MAX_AUTO_WORKAREA_SIZE"`

	MemTarget      sql.NullInt64 `db:"MEM_TARGET"`       // 当前 workarea 可用内存的目标大小。
	GlobalMemBound sql.NullInt64 `db:"GLOBAL_MEM_BOUND"` // auto 模式下，全局最大可用内存大小。
}

// SQL_PLAN_CACHE_STAT selects table GV$PLAN_CACHE_STAT.
//
//nolint:stylecheck
const SQL_PLAN_CACHE_STAT = `SELECT 
SUM(MEM_USED),
SUM(MEM_HOLD),
AVG(HIT_RATE),
SUM(PLAN_NUM)
FROM GV$PLAN_CACHE_STAT`

type obPlanCacheStat struct {
	SumMemUsed sql.NullFloat64 `db:"SUM(MEM_USED)"` // 所有 Server 上 plan_cache 已经使用的总内存。
	SumMemHold sql.NullFloat64 `db:"SUM(MEM_HOLD)"` // 所有 Server 上 plan_cache 持有的总内存。
	AvgHitRate sql.NullFloat64 `db:"AVG(HIT_RATE)"` // 所有 Server 上 plan_cache 的平均命中率。
	SumPlanNum sql.NullInt64   `db:"SUM(PLAN_NUM)"` // 所有 Server 上 plan 的总数。
}

// SQL_PS_STAT selects table GV$PS_STAT.
//
//nolint:stylecheck
const SQL_PS_STAT = `SELECT 
case WHEN(NVL(SUM(ACCESS_COUNT), 0))!=0 THEN 
  round(SUM(HIT_COUNT) / SUM(ACCESS_COUNT), 3) 
ELSE 0 
END 
RESULT 
FROM GV$PS_STAT`

type obPsStat struct {
	HitRate sql.NullFloat64 `db:"RESULT"` // PS(Prepared Statement) Cache 的命中率。
}

// SQL_SESSION_WAIT selects table GV$SESSION_WAIT.
//
//nolint:stylecheck
const SQL_SESSION_WAIT = `SELECT 
AVG(WAIT_TIME_MICRO)
FROM GV$SESSION_WAIT`

type obSessionWait struct {
	// 所有服务器上所有 Session 的当前或者上一次等待事件的平均等待耗时。单位为微秒。
	AvgWaitTimeMicro sql.NullFloat64 `db:"AVG(WAIT_TIME_MICRO)"`
}

//nolint:stylecheck
const SQL_STAT = `
select 
  /*+ MONITOR_AGENT READ_CONSISTENCY(WEAK) */
  con_id tenant_id, 
  SVR_IP, 
  stat_id, 
  tenant.tenant_name,
  replace(name, " ", "_") as metrics_name, 
  value as metrics_value 
from 
  gv$sysstat 
left join gv$tenant tenant on con_id=tenant.tenant_id
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
    50010
  ) 
  and (
    con_id > 1000 
    or con_id = 1
  ) 
  and class < 1000;
`

type obStat struct {
	MetricsValue sql.NullFloat64 `db:"metrics_value"`
	TenantID     sql.NullString  `db:"tenant_id"`
	MetricsName  sql.NullString  `db:"metrics_name"`
	SvrIP        sql.NullString  `db:"SVR_IP"`
	StatID       sql.NullString  `db:"stat_id"`
	TenantName   sql.NullString  `db:"tenant_name"`
}

func (m *obMetrics) collectStat() ([]*ccommon.TagField, error) {
	tfs := []*ccommon.TagField{}
	rows := []obStat{}
	if err := selectWrapper(m.x.Ipt, &rows, SQL_STAT); err != nil {
		return nil, err
	}

	for _, row := range rows {
		tf := ccommon.NewTagField()

		tf.AddTag("cluster", m.x.Ipt.Cluster)
		tf.AddTag("tenant_name", row.TenantName.String)
		tf.AddTag("tenant_id", row.TenantID.String)
		tf.AddTag("stat_id", row.StatID.String)
		tf.AddTag("metrics_name", row.MetricsName.String)
		tf.AddTag("svr_ip", row.SvrIP.String)

		tf.AddField("metrics_value", row.MetricsValue.Float64, nil)

		tfs = append(tfs, tf)
	}
	return tfs, nil
}

//nolint:stylecheck
const SQL_EVENT = `
select 
  /*+ MONITOR_AGENT READ_CONSISTENCY(WEAK) */
  con_id tenant_id, 
  SVR_IP, 
  tenant.tenant_name,
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
left join gv$tenant tenant on con_id=tenant.tenant_id
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
	TenantName sql.NullString  `db:"tenant_name"`
}

func (m *obMetrics) collectEvent() ([]*ccommon.TagField, error) {
	tfs := []*ccommon.TagField{}
	rows := []obEvent{}
	if err := selectWrapper(m.x.Ipt, &rows, SQL_EVENT); err != nil {
		return nil, err
	}

	for _, row := range rows {
		tf := ccommon.NewTagField()

		tf.AddTag("cluster", m.x.Ipt.Cluster)
		tf.AddTag("tenant_name", row.TenantName.String)
		tf.AddTag("tenant_id", row.TenantID.String)
		tf.AddTag("svr_ip", row.SvrIP.String)
		tf.AddTag("event_group", row.EventGroup.String)
		tf.AddField("total_waits", row.TotalWaits.Float64, nil)
		tf.AddField("time_waited", row.TimeWaited.Float64, nil)

		tfs = append(tfs, tf)
	}
	return tfs, nil
}

//nolint:stylecheck
const SQL_CACHE = `
select 
  /*+ MONITOR_AGENT READ_CONSISTENCY(WEAK) */
  tenant_id, 
  access_count, 
  tenant.tenant_name,
  hit_count, 
  SVR_IP 
from 
  gv$plan_cache_stat
left join gv$tenant tenant on con_id=tenant.tenant_id
`

type obCache struct {
	TenantID    sql.NullString `db:"tenant_id"`
	AccessCount sql.NullInt64  `db:"access_count"`
	HitCount    sql.NullInt64  `db:"hit_count"`
	SvrIP       sql.NullString `db:"SVR_IP"`
	TenantName  sql.NullString `db:"tenant_name"`
}

func (m *obMetrics) collectCache() ([]*ccommon.TagField, error) {
	tfs := []*ccommon.TagField{}
	rows := []obCache{}
	if err := selectWrapper(m.x.Ipt, &rows, SQL_CACHE); err != nil {
		return nil, err
	}
	for _, row := range rows {
		tf := ccommon.NewTagField()

		tf.AddTag("cluster", m.x.Ipt.Cluster)
		tf.AddTag("tenant_name", row.TenantName.String)
		tf.AddTag("tenant_id", row.TenantID.String)
		tf.AddTag("svr_ip", row.SvrIP.String)
		tf.AddField("access_count", row.AccessCount.Int64, nil)

		tfs = append(tfs, tf)
	}

	return tfs, nil
}

func (m *obMetrics) collectOB() ([]*ccommon.TagField, error) {
	tf := ccommon.NewTagField()

	var dbs *dbState
	var err error

	if tenantMode == modeDefault {
		return nil, errUnknownMode
	}

	l.Debugf("tenantMode = %s", getTenantModeName())

	//nolint:exhaustive
	switch tenantMode {
	case modeMySQL:
		dbs, err = getMySQLStatus(m.x.Ipt)
		if err != nil {
			l.Errorf("getMySQLStatus() failed: %v", err)
			tf.AddField("ob_database_status", int(Down), nil)
			return []*ccommon.TagField{tf}, nil
		}

	case modeOracle:
		rows := []obInstance{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_INSTANCE); err != nil {
			tf.AddField("ob_database_status", int(Down), nil)
			return []*ccommon.TagField{tf}, nil
		}

		var hostName, version []string
		status := OK
		for _, v := range rows {
			hostName = append(hostName, v.Hostname.String)
			version = append(version, v.Version.String)

			switch v.DatabaseStatus.String {
			case "active":
			default:
				status = Down
			}
		}

		var ds dbState
		ds.Hostname = strings.Join(hostName, ", ")
		ds.Version = strings.Join(version, ", ")
		ds.Status = status
		dbs = &ds
	}

	tf.AddTag("ob_host_name", dbs.Hostname)
	tf.AddTag("ob_version", dbs.Version)
	tf.AddField("ob_database_status", int(dbs.Status), nil)

	{
		rows := []obLock{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_LOCK); err != nil {
			tf.AddField("ob_database_status", int(Down), nil)

			return []*ccommon.TagField{tf}, nil
		}
		for _, v := range rows {
			tf.AddField("ob_lock_count", v.Count.Int64, nil)
			tf.AddField("ob_lock_max_ctime", v.MaxCtime.Int64, nil)
			break
		}
	}

	{
		rows := []obConcurrentLimitSQL{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_CONCURRENT_LIMIT_SQL); err != nil {
			tf.AddField("ob_database_status", int(Down), nil)
			return []*ccommon.TagField{tf}, nil
		}
		for _, v := range rows {
			tf.AddField("ob_concurrent_limit_sql_count", v.Count.Int64, nil)
			break
		}
	}

	{
		rows := []obMemory{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_MEMORY); err != nil {
			tf.AddField("ob_database_status", int(Down), nil)
			return []*ccommon.TagField{tf}, nil
		}
		for _, v := range rows {
			tf.AddField("ob_mem_sum_count", v.SumCount.Int64, nil)
			tf.AddField("ob_mem_sum_used", v.SumUsed.Int64, nil)
			break
		}
	}

	{
		rows := []obMemStore{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_MEMSTORE); err != nil {
			tf.AddField("ob_database_status", int(Down), nil)
			return []*ccommon.TagField{tf}, nil
		}
		for _, v := range rows {
			tf.AddField("ob_memstore_active_rate", getThreeDecimal(v.ActiveRate.Float64)*100, nil)
			break
		}
	}

	{
		rows := []obSQLWorkareaMemoryInfo{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_OB_SQL_WORKAREA_MEMORY_INFO); err != nil {
			tf.AddField("ob_database_status", int(Down), nil)
			return []*ccommon.TagField{tf}, nil
		}
		for _, v := range rows {
			tf.AddField("ob_workarea_max_auto_workarea_size", v.MaxAutoWorkareaSize.Int64, nil)
			tf.AddField("ob_workarea_mem_target", v.MemTarget.Int64, nil)
			tf.AddField("ob_workarea_global_mem_bound", v.GlobalMemBound.Int64, nil)
			break
		}
		if len(rows) == 0 {
			tf.AddField("ob_workarea_max_auto_workarea_size", 0, nil)
			tf.AddField("ob_workarea_mem_target", 0, nil)
			tf.AddField("ob_workarea_global_mem_bound", 0, nil)
		}
	}

	{
		rows := []obPlanCacheStat{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_PLAN_CACHE_STAT); err != nil {
			tf.AddField("ob_database_status", int(Down), nil)
			return []*ccommon.TagField{tf}, nil
		}
		for _, v := range rows {
			{
				used := v.SumMemUsed.Float64
				hold := v.SumMemHold.Float64
				if hold == 0 {
					tf.AddField("ob_plancache_mem_used_rate", float64(0), nil) // 所有 Server 上 plan_cache 的总体内存使用率。（已经使用的内存/持有的内存）
				} else {
					ratio := used / hold
					tf.AddField("ob_plancache_mem_used_rate", getThreeDecimal(ratio)*100, nil) // 所有 Server 上 plan_cache 的总体内存使用率。（已经使用的内存/持有的内存）
				}
			}
			tf.AddField("ob_plancache_avg_hit_rate", getOneDecimal(v.AvgHitRate.Float64), nil)
			tf.AddField("ob_plancache_sum_plan_num", v.SumPlanNum.Int64, nil)
			break
		}
	}

	{
		rows := []obPsStat{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_PS_STAT); err != nil {
			tf.AddField("ob_database_status", int(Down), nil)
			return []*ccommon.TagField{tf}, nil
		}
		for _, v := range rows {
			tf.AddField("ob_ps_hit_rate", getThreeDecimal(v.HitRate.Float64)*100, nil)
			break
		}
	}

	{
		rows := []obSessionWait{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_SESSION_WAIT); err != nil {
			return []*ccommon.TagField{tf}, nil
		}
		for _, v := range rows {
			tf.AddField("ob_session_avg_wait_time", getOneDecimal(v.AvgWaitTimeMicro.Float64), nil)
			break
		}
	}

	return []*ccommon.TagField{tf}, nil
}

func getOneDecimal(in float64) float64 {
	return math.Round(in*10) / 10
}

func getThreeDecimal(in float64) float64 {
	return math.Round(in*1000) / 1000
}
