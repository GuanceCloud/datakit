// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package coracle

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/oracle/collect/ccommon"
)

type systemMetrics struct {
	x collectParameters
}

var _ ccommon.DBMetricsCollector = (*systemMetrics)(nil)

func newSystemMetrics(opts ...collectOption) *systemMetrics {
	m := &systemMetrics{}

	for _, opt := range opts {
		if opt != nil {
			opt(&m.x)
		}
	}

	return m
}

func (m *systemMetrics) Collect() ([]*point.Point, error) {
	l.Debug("Collect entry")
	return m.sysMetrics()
}

// SYSMETRICS_QUERY is the get system info SQL query for Oracle for 12c2 and 12c2+.
// https://docs.oracle.com/en/database/oracle/oracle-database/12.2/refrn/V-CON_SYSMETRIC.html#GUID-B7A51ACA-952B-47FE-9BC5-B265A120A9F7
//
//nolint:stylecheck
const SYSMETRICS_QUERY = `SELECT 
	metric_name,
	value, 
	metric_unit, 
	--(end_time - begin_time)*24*3600 interval_length,
	name pdb_name 
  FROM %s s, v$containers c 
  WHERE s.con_id = c.con_id(+)`

// SYSMETRICS_QUERY_OLD is the get system info SQL query for Oracle for for 12c1 and 12c1-.
//
//nolint:stylecheck
const SYSMETRICS_QUERY_OLD = `SELECT 
    VALUE, 
    METRIC_NAME 
  FROM GV$SYSMETRIC ORDER BY BEGIN_TIME`

type sysmetricsRowDB struct {
	MetricName string         `db:"METRIC_NAME"`
	Value      float64        `db:"VALUE"`
	MetricUnit string         `db:"METRIC_UNIT"`
	PdbName    sql.NullString `db:"PDB_NAME"`
}

type sysMetricsDefinition struct {
	DDmetric string
	DBM      bool
}

var (
	//nolint:stylecheck
	SYSMETRICS_COLS = map[string]sysMetricsDefinition{
		"Average Active Sessions":                       {DDmetric: "active_sessions"},
		"Average Synchronous Single-Block Read Latency": {DDmetric: "avg_synchronous_single_block_read_latency", DBM: true},
		"Background CPU Usage Per Sec":                  {DDmetric: "active_background_on_cpu", DBM: true},
		"Background Time Per Sec":                       {DDmetric: "active_background", DBM: true},
		"Branch Node Splits Per Sec":                    {DDmetric: "branch_node_splits", DBM: true},
		"Buffer Cache Hit Ratio":                        {DDmetric: "buffer_cachehit_ratio"},
		"Consistent Read Changes Per Sec":               {DDmetric: "consistent_read_changes", DBM: true},
		"Consistent Read Gets Per Sec":                  {DDmetric: "consistent_read_gets", DBM: true},
		"CPU Usage Per Sec":                             {DDmetric: "active_sessions_on_cpu", DBM: true},
		"Current OS Load":                               {DDmetric: "os_load", DBM: true},
		"Database CPU Time Ratio":                       {DDmetric: "database_cpu_time_ratio", DBM: true},
		"Database Wait Time Ratio":                      {DDmetric: "database_wait_time_ratio"},
		"DB Block Changes Per Sec":                      {DDmetric: "db_block_changes", DBM: true},
		"DB Block Gets Per Sec":                         {DDmetric: "db_block_gets", DBM: true},
		"DBWR Checkpoints Per Sec":                      {DDmetric: "dbwr_checkpoints", DBM: true},
		"Disk Sort Per Sec":                             {DDmetric: "disk_sorts"},
		"Enqueue Deadlocks Per Sec":                     {DDmetric: "enqueue_deadlocks", DBM: true},
		"Enqueue Timeouts Per Sec":                      {DDmetric: "enqueue_timeouts"},
		"Execute Without Parse Ratio":                   {DDmetric: "execute_without_parse", DBM: true},
		"GC CR Block Received Per Second":               {DDmetric: "gc_cr_block_received"},
		"GC Current Block Received Per Second":          {DDmetric: "gc_current_block_received", DBM: true},
		"Global Cache Average CR Get Time":              {DDmetric: "gc_average_cr_get_time", DBM: true},
		"Global Cache Average Current Get Time":         {DDmetric: "gc_average_current_get_time", DBM: true},
		"Global Cache Blocks Corrupted":                 {DDmetric: "cache_blocks_corrupt"},
		"Global Cache Blocks Lost":                      {DDmetric: "cache_blocks_lost"},
		"Hard Parse Count Per Sec":                      {DDmetric: "hard_parses", DBM: true},
		"Host CPU Utilization (%)":                      {DDmetric: "host_cpu_utilization", DBM: true},
		"Leaf Node Splits Per Sec":                      {DDmetric: "leaf_nodes_splits", DBM: true},
		"Library Cache Hit Ratio":                       {DDmetric: "library_cachehit_ratio"},
		"Logical Reads Per Sec":                         {DDmetric: "logical_reads", DBM: true},
		"Logons Per Sec":                                {DDmetric: "logons"},
		"Long Table Scans Per Sec":                      {DDmetric: "long_table_scans"},
		"Memory Sorts Ratio":                            {DDmetric: "memory_sorts_ratio"},
		"Network Traffic Volume Per Sec":                {DDmetric: "network_traffic_volume", DBM: true},
		"PGA Cache Hit %":                               {DDmetric: "pga_cache_hit", DBM: true},
		"Parse Failure Count Per Sec":                   {DDmetric: "parse_failures", DBM: true},
		"Physical Read Bytes Per Sec":                   {DDmetric: "physical_read_bytes", DBM: true},
		"Physical Read IO Requests Per Sec":             {DDmetric: "physical_read_io_requests", DBM: true},
		"Physical Read Total IO Requests Per Sec":       {DDmetric: "physical_read_total_io_requests", DBM: true},
		"Physical Reads Direct Lobs Per Sec":            {DDmetric: "physical_reads_direct_lobs", DBM: true},
		"Physical Read Total Bytes Per Sec":             {DDmetric: "physical_read_total_bytes", DBM: true},
		"Physical Reads Direct Per Sec":                 {DDmetric: "physical_reads_direct", DBM: true},
		"Physical Reads Per Sec":                        {DDmetric: "physical_reads"},
		"Physical Write Bytes Per Sec":                  {DDmetric: "physical_write_bytes", DBM: true},
		"Physical Write IO Requests Per Sec":            {DDmetric: "physical_write_io_requests", DBM: true},
		"Physical Write Total Bytes Per Sec":            {DDmetric: "physical_write_total_bytes", DBM: true},
		"Physical Write Total IO Requests Per Sec":      {DDmetric: "physical_write_total_io_requests", DBM: true},
		"Physical Writes Direct Lobs Per Sec":           {DDmetric: "physical_writes_direct_lobs", DBM: true},
		"Physical Writes Direct Per Sec":                {DDmetric: "physical_writes_direct", DBM: true},
		"Physical Writes Per Sec":                       {DDmetric: "physical_writes"},
		"Process Limit %":                               {DDmetric: "process_limit", DBM: true},
		"Redo Allocation Hit Ratio":                     {DDmetric: "redo_allocation_hit_ratio", DBM: true},
		"Redo Generated Per Sec":                        {DDmetric: "redo_generated", DBM: true},
		"Redo Writes Per Sec":                           {DDmetric: "redo_writes", DBM: true},
		"Row Cache Hit Ratio":                           {DDmetric: "row_cache_hit_ratio", DBM: true},
		"Rows Per Sort":                                 {DDmetric: "rows_per_sort"},
		"SQL Service Response Time":                     {DDmetric: "service_response_time"},
		"Session Count":                                 {DDmetric: "session_count"},
		"Session Limit %":                               {DDmetric: "session_limit_usage"},
		"Shared Pool Free %":                            {DDmetric: "shared_pool_free"},
		"Soft Parse Ratio":                              {DDmetric: "soft_parse_ratio", DBM: true},
		"Temp Space Used":                               {DDmetric: "temp_space_used"},
		"Total Parse Count Per Sec":                     {DDmetric: "total_parse_count", DBM: true},
		"Total Sorts Per User Call":                     {DDmetric: "sorts_per_user_call"},
		"User Commits Per Sec":                          {DDmetric: "user_commits", DBM: true},
		"User Rollbacks Per Sec":                        {DDmetric: "user_rollbacks"},
	}

	dic = map[string]string{
		"buffer_cache_hit_ratio":       "buffer_cachehit_ratio",
		"cursor_cache_hit_ratio":       "cursor_cachehit_ratio",
		"library_cache_hit_ratio":      "library_cachehit_ratio",
		"shared_pool_free_%":           "shared_pool_free",
		"physical_read_bytes_per_sec":  "physical_reads",
		"physical_write_bytes_per_sec": "physical_writes",
		"enqueue_timeouts_per_sec":     "enqueue_timeouts",

		"gc_cr_block_received_per_second": "gc_cr_block_received",
		"global_cache_blocks_corrupted":   "cache_blocks_corrupt",
		"global_cache_blocks_lost":        "cache_blocks_lost",
		"average_active_sessions":         "active_sessions",
		"sql_service_response_time":       "service_response_time",
		"user_rollbacks_per_sec":          "user_rollbacks",
		"total_sorts_per_user_call":       "sorts_per_user_call",
		"rows_per_sort":                   "rows_per_sort",
		"disk_sort_per_sec":               "disk_sorts",
		"memory_sorts_ratio":              "memory_sorts_ratio",
		"database_wait_time_ratio":        "database_wait_time_ratio",
		"session_limit_%":                 "session_limit_usage",
		"session_count":                   "session_count",
		"temp_space_used":                 "temp_space_used",
	}
)

func (*systemMetrics) addMetric(r sysmetricsRowDB, seen map[string]bool) point.KVs {
	var kvs point.KVs

	if metric, ok := SYSMETRICS_COLS[r.MetricName]; ok {
		value := r.Value
		if r.MetricUnit == "CentiSeconds Per Second" {
			value /= 100
		}

		l.Debugf("%s: %f", metric.DDmetric, value)
		if r.PdbName.Valid {
			kvs = kvs.AddTag(pdbName, r.PdbName.String)
		}

		alias, ok := dic[metric.DDmetric]
		if ok {
			kvs = kvs.Add(alias, value, false, false)
		} else {
			kvs = kvs.Add(metric.DDmetric, value, false, false)
		}
		seen[r.MetricName] = true
	}

	return kvs
}

func (m *systemMetrics) sysMetrics() ([]*point.Point, error) {
	var pts []*point.Point
	seenInContainerMetrics := make(map[string]bool)

	hostTag := ccommon.GetHostTag(l, m.x.Ipt.host)

	metricRows := []sysmetricsRowDB{}
	err := selectWrapper(m.x.Ipt, &metricRows, fmt.Sprintf(SYSMETRICS_QUERY, "v$con_sysmetric"))
	if err != nil {
		if strings.Contains(err.Error(), "dpiStmt_execute: ORA-00942: table or view does not exist") {
			l.Debug("system: dpiStmt_execute: ORA-00942: table or view does not exist. Maybe Oracle version lower than 12.")

			// oracle old version. 12c2-
			if err = selectWrapper(m.x.Ipt, &metricRows, SYSMETRICS_QUERY_OLD); err != nil {
				return nil, fmt.Errorf("failed to collect old sysmetrics: %w", err)
			}

			var kvs point.KVs
			for _, r := range metricRows {
				newKVs := m.addMetric(r, seenInContainerMetrics)
				if newKVs.FieldCount() > 0 {
					kvs = appendKVs(kvs, newKVs)
				}
			}

			if kvs.FieldCount() > 0 {
				kvs, err = m.getOverAllocationCount(kvs)
				if err != nil {
					return nil, err
				}

				pts = append(pts, ccommon.BuildPointMetric(
					kvs, m.x.MetricName,
					m.x.Ipt.tags, hostTag,
				))
			}

			return pts, nil
		} else {
			return nil, fmt.Errorf("failed to collect container sysmetrics: %w", err)
		}
	}
	l.Debugf("system 1: len(metricRows) = %d", len(metricRows))

	{
		var kvs point.KVs
		for _, r := range metricRows {
			newKVs := m.addMetric(r, seenInContainerMetrics)
			if newKVs.FieldCount() > 0 {
				kvs = appendKVs(kvs, newKVs)
			}
		}

		if kvs.FieldCount() > 0 {
			kvs, err = m.getOverAllocationCount(kvs)
			if err != nil {
				return nil, err
			}

			pts = append(pts, ccommon.BuildPointMetric(
				kvs, m.x.MetricName,
				m.x.Ipt.tags, hostTag,
			))
		}
	}

	seenInGlobalMetrics := make(map[string]bool)
	err = selectWrapper(m.x.Ipt, &metricRows, fmt.Sprintf(SYSMETRICS_QUERY, "v$sysmetric")+" ORDER BY begin_time ASC, metric_name ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to collect sysmetrics: %w", err)
	}
	l.Debugf("system 2: len(metricRows) = %d", len(metricRows))

	{
		var kvs point.KVs
		for _, r := range metricRows {
			if _, ok := seenInContainerMetrics[r.MetricName]; !ok {
				if _, ok := seenInGlobalMetrics[r.MetricName]; ok {
					break
				} else {
					newKVs := m.addMetric(r, seenInGlobalMetrics)
					if newKVs.FieldCount() > 0 {
						kvs = appendKVs(kvs, newKVs)
					}
				}
			}
		}

		if kvs.FieldCount() > 0 {
			kvs, err = m.getOverAllocationCount(kvs)
			if err != nil {
				return nil, err
			}

			pts = append(pts, ccommon.BuildPointMetric(
				kvs, m.x.MetricName,
				m.x.Ipt.tags, hostTag,
			))
		}
	}

	return pts, nil
}

func appendKVs(in, toAdd point.KVs) point.KVs {
	for _, v := range toAdd {
		in = in.AddKV(v, false)
	}
	return in
}

type pgaOverAllocationCount struct {
	value float64
	valid bool
}

var previousPGAOverAllocationCount pgaOverAllocationCount

func (m *systemMetrics) getOverAllocationCount(kvs point.KVs) (point.KVs, error) {
	var overAllocationCount float64
	if err := getWrapper(m.x.Ipt, &overAllocationCount, "SELECT value FROM v$pgastat WHERE name = 'over allocation count'"); err != nil {
		return nil, fmt.Errorf("failed to get PGA over allocation count: %w", err)
	}

	if previousPGAOverAllocationCount.valid {
		v := overAllocationCount - previousPGAOverAllocationCount.value
		kvs = kvs.Add("pga_over_allocation_count", v, false, false)
		previousPGAOverAllocationCount.value = overAllocationCount
	} else {
		kvs = kvs.Add("pga_over_allocation_count", float64(0), false, false)
		previousPGAOverAllocationCount = pgaOverAllocationCount{value: overAllocationCount, valid: true}
	}

	return kvs, nil
}
