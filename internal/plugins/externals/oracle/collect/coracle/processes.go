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

type processMetrics struct {
	x collectParameters
}

var _ ccommon.DBMetricsCollector = (*processMetrics)(nil)

func newProcessMetrics(opts ...collectOption) *processMetrics {
	m := &processMetrics{}

	for _, opt := range opts {
		if opt != nil {
			opt(&m.x)
		}
	}

	return m
}

func (m *processMetrics) Collect() ([]*point.Point, error) {
	l.Debug("Collect entry")

	return m.processMemory()
}

// PGA_QUERY is the get process info SQL query for Oracle 11g+.
//
//nolint:stylecheck
const PGA_QUERY = `SELECT 
	name pdb_name, 
	pid, 
	program, 
	nvl(pga_used_mem,0) pga_used_mem, 
	nvl(pga_alloc_mem,0) pga_alloc_mem, 
	nvl(pga_freeable_mem,0) pga_freeable_mem, 
	nvl(pga_max_mem,0) pga_max_mem
  FROM v$process p, v$containers c
  WHERE
  	c.con_id(+) = p.con_id`

// PGA_QUERY_OLD is the get process info SQL query for Oracle 11g and 11g-.
//
//nolint:stylecheck
const PGA_QUERY_OLD = `SELECT 
    PROGRAM, 
    PGA_USED_MEM, 
    PGA_ALLOC_MEM, 
    PGA_FREEABLE_MEM, 
    PGA_MAX_MEM
  FROM GV$PROCESS`

type processesRowDB struct {
	PdbName        sql.NullString `db:"PDB_NAME"`
	PID            uint64         `db:"PID"`
	Program        sql.NullString `db:"PROGRAM"`
	PGAUsedMem     float64        `db:"PGA_USED_MEM"`
	PGAAllocMem    float64        `db:"PGA_ALLOC_MEM"`
	PGAFreeableMem float64        `db:"PGA_FREEABLE_MEM"`
	PGAMaxMem      float64        `db:"PGA_MAX_MEM"`
}

func (m *processMetrics) processMemory() ([]*point.Point, error) {
	rows := []processesRowDB{}

	err := selectWrapper(m.x.Ipt, &rows, PGA_QUERY)
	if err != nil {
		l.Debug("process: dpiStmt_execute: ORA-00942: table or view does not exist")

		if strings.Contains(err.Error(), "dpiStmt_execute: ORA-00942: table or view does not exist") {
			// oracle old version. 11g
			if err = selectWrapper(m.x.Ipt, &rows, PGA_QUERY_OLD); err != nil {
				return nil, fmt.Errorf("failed to collect old processes info: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to collect processes info: %w", err)
		}
	}

	var pts []*point.Point
	hostTag := ccommon.GetHostTag(l, m.x.Ipt.host)

	for _, r := range rows {
		var kvs point.KVs
		if r.PdbName.Valid {
			kvs = kvs.AddTag(pdbName, r.PdbName.String)
		}

		if r.Program.Valid {
			kvs = kvs.AddTag(programName, r.Program.String)
		}

		if r.PID > 0 {
			kvs = kvs.Add("pid", r.PID, false, false)
		}

		kvs = kvs.Add("pga_alloc_mem", r.PGAAllocMem, false, false)
		kvs = kvs.Add("pga_freeable_mem", r.PGAFreeableMem, false, false)
		kvs = kvs.Add("pga_max_mem", r.PGAMaxMem, false, false)
		kvs = kvs.Add("pga_used_mem", r.PGAUsedMem, false, false)

		pts = append(pts, ccommon.BuildPointMetric(
			kvs, m.x.MetricName,
			m.x.Ipt.tags, hostTag,
		))
	}

	return pts, nil
}
