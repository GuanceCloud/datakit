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

func (m *processMetrics) Collect() (*point.Point, error) {
	l.Debug("Collect entry")

	tf, err := m.processMemory()
	if err != nil {
		return nil, err
	}

	if tf.IsEmpty() {
		return nil, fmt.Errorf("process empty data")
	}

	opt := &ccommon.BuildPointOpt{
		TF:         tf,
		MetricName: m.x.MetricName,
		Tags:       m.x.Ipt.tags,
		Host:       m.x.Ipt.host,
	}
	return ccommon.BuildPoint(l, opt), nil
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

func (m *processMetrics) processMemory() (*ccommon.TagField, error) {
	tf := ccommon.NewTagField()
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

	for _, r := range rows {
		if r.PdbName.Valid {
			tf.AddTag(pdbName, r.PdbName.String)
		}

		if r.Program.Valid {
			tf.AddTag(programName, r.Program.String)
		}

		if r.PID > 0 {
			tf.AddField("pid", r.PID, dic)
		}

		tf.AddField("pga_alloc_mem", r.PGAAllocMem, dic)
		tf.AddField("pga_freeable_mem", r.PGAFreeableMem, dic)
		tf.AddField("pga_max_mem", r.PGAMaxMem, dic)
		tf.AddField("pga_used_mem", r.PGAUsedMem, dic)
	}

	return tf, nil
}
