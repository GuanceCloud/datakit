// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package collect

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
)

type processMetrics struct {
	x collectParameters
}

var _ dbMetricsCollector = (*processMetrics)(nil)

func newProcessMetrics(opts ...collectOption) *processMetrics {
	m := &processMetrics{}

	for _, opt := range opts {
		if opt != nil {
			opt(&m.x)
		}
	}

	return m
}

func (m *processMetrics) collect() (*point.Point, error) {
	l.Debug("processMetrics Collect entry")

	tf, err := m.processMemory()
	if err != nil {
		return nil, err
	}

	if tf.isEmpty() {
		return nil, fmt.Errorf("process empty data")
	}

	opt := &buildPointOpt{
		tf:         tf,
		metricName: metricNameProcess,
		m:          m.x.m,
	}
	return buildPoint(opt), nil
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

func (m *processMetrics) processMemory() (*tagField, error) {
	tf := newTagField()
	rows := []processesRowDB{}

	err := selectWrapper(m.x.m, &rows, PGA_QUERY)
	if err != nil {
		l.Debug("process: dpiStmt_execute: ORA-00942: table or view does not exist")

		if strings.Contains(err.Error(), "dpiStmt_execute: ORA-00942: table or view does not exist") {
			// oracle old version. 11g
			if err = selectWrapper(m.x.m, &rows, PGA_QUERY_OLD); err != nil {
				return nil, fmt.Errorf("failed to collect old processes info: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to collect processes info: %w", err)
		}
	}

	for _, r := range rows {
		if r.PdbName.Valid {
			tf.addTag(pdbName, r.PdbName.String)
		}

		if r.Program.Valid {
			tf.addTag(programName, r.Program.String)
		}

		if r.PID > 0 {
			tf.addField("pid", r.PID)
		}

		tf.addField("pga_alloc_mem", r.PGAAllocMem)
		tf.addField("pga_freeable_mem", r.PGAFreeableMem)
		tf.addField("pga_max_mem", r.PGAMaxMem)
		tf.addField("pga_used_mem", r.PGAUsedMem)
	}

	return tf, nil
}
