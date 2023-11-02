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

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/oceanbase/collect/ccommon"
)

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

func (m *obMetrics) Collect() (*point.Point, error) {
	l.Debug("Collect entry")

	tf, err := m.collectOB()
	if err != nil {
		return nil, err
	}

	if tf == nil {
		return nil, nil
	}

	if tf.IsEmpty() {
		return nil, fmt.Errorf("ob metrics empty")
	}

	opt := &ccommon.BuildPointOpt{
		TF:         tf,
		MetricName: m.x.MetricName,
		Tags:       m.x.Ipt.tags,
		Host:       m.x.Ipt.host,
	}
	return ccommon.BuildPoint(l, opt), nil
}

////////////////////////////////////////////////////////////////////////////////

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
result
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
result 
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

////////////////////////////////////////////////////////////////////////////////

func (m *obMetrics) collectOB() (*ccommon.TagField, error) {
	tf := ccommon.NewTagField()

	{
		rows := []obLock{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_LOCK); err != nil {
			return nil, err
		}
		for _, v := range rows {
			tf.AddField("ob_lock_count", v.Count.Int64, nil)
			tf.AddField("ob_lock_max_ctime", v.MaxCtime.Int64, nil)
		}
	}

	{
		rows := []obConcurrentLimitSQL{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_CONCURRENT_LIMIT_SQL); err != nil {
			return nil, err
		}
		for _, v := range rows {
			tf.AddField("ob_concurrent_limit_sql_count", v.Count.Int64, nil)
		}
	}

	{
		rows := []obInstance{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_INSTANCE); err != nil {
			return nil, err
		}
		for _, v := range rows {
			tf.AddTag("ob_host_name", v.Hostname.String)
			tf.AddTag("ob_version", v.Version.String)

			switch v.DatabaseStatus.String { //nolint:gocritic
			case "active":
				tf.AddField("ob_database_status", 1, nil)
			default:
				l.Errorf("Database status error: %s", v.DatabaseStatus.String)
				tf.AddField("ob_database_status", 0, nil)
			}
		}
	}

	{
		rows := []obMemory{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_MEMORY); err != nil {
			return nil, err
		}
		for _, v := range rows {
			tf.AddField("ob_mem_sum_count", v.SumCount.Int64, nil)
			tf.AddField("ob_mem_sum_used", v.SumUsed.Int64, nil)
		}
	}

	{
		rows := []obMemStore{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_MEMSTORE); err != nil {
			return nil, err
		}
		for _, v := range rows {
			tf.AddField("ob_memstore_active_rate", getThreeDecimal(v.ActiveRate.Float64)*100, nil)
		}
	}

	{
		rows := []obSQLWorkareaMemoryInfo{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_OB_SQL_WORKAREA_MEMORY_INFO); err != nil {
			return nil, err
		}
		for _, v := range rows {
			tf.AddField("ob_workarea_max_auto_workarea_size", v.MaxAutoWorkareaSize.Int64, nil)
			tf.AddField("ob_workarea_mem_target", v.MemTarget.Int64, nil)
			tf.AddField("ob_workarea_global_mem_bound", v.GlobalMemBound.Int64, nil)
		}
	}

	{
		rows := []obPlanCacheStat{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_PLAN_CACHE_STAT); err != nil {
			return nil, err
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
		}
	}

	{
		rows := []obPsStat{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_PS_STAT); err != nil {
			return nil, err
		}
		for _, v := range rows {
			tf.AddField("ob_ps_hit_rate", getThreeDecimal(v.HitRate.Float64)*100, nil)
		}
	}

	{
		rows := []obSessionWait{}
		if err := selectWrapper(m.x.Ipt, &rows, SQL_SESSION_WAIT); err != nil {
			return nil, err
		}
		for _, v := range rows {
			tf.AddField("ob_session_avg_wait_time", getOneDecimal(v.AvgWaitTimeMicro.Float64), nil)
		}
	}

	return tf, nil
}

func getOneDecimal(in float64) float64 {
	return math.Round(in*10) / 10
}

func getThreeDecimal(in float64) float64 {
	return math.Round(in*1000) / 1000
}
