// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"database/sql"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

// SQLProcess for Oracle 11g+.
const SQLProcess = `SELECT 
	name pdb_name,  -- v$containers.name
	pid,
	program,
	nvl(pga_used_mem,0) pga_used_mem, 
	nvl(pga_alloc_mem,0) pga_alloc_mem, 
	nvl(pga_freeable_mem,0) pga_freeable_mem, 
	nvl(pga_max_mem,0) pga_max_mem,
	nvl(CPU_USED,0) cpu_used
FROM v$process p, v$containers c
WHERE c.con_id(+) = p.con_id`

// SQLProcessOld for Oracle 11g and 11g-.
const SQLProcessOld = `SELECT 
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
	CPUUsed        int64          `db:"CPU_USED"`
	PGAMaxMem      float64        `db:"PGA_MAX_MEM"`
}

func (ipt *Input) collectOracleProcess() {
	var (
		start      = time.Now()
		metricName = "oracle_process"
		rows       = []processesRowDB{}
		pts        []*point.Point
	)

	if sql, ok := ipt.cacheSQL[metricName]; ok {
		if err := selectWrapper(ipt, &rows, sql, getMetricName(metricName, "oracle_process")); err != nil {
			l.Error("failed to collect processes info: %s", err)
			return
		}
	} else {
		err := selectWrapper(ipt, &rows, SQLProcess, getMetricName(metricName, "oracle_process"))
		ipt.cacheSQL[metricName] = SQLProcess
		l.Infof("Query for metric [%s], sql: %s", metricName, SQLProcess)
		if err != nil {
			if strings.Contains(err.Error(), "ORA-00942: table or view does not exist") {
				ipt.cacheSQL[metricName] = SQLProcessOld
				l.Infof("Query for metric [%s], sql: %s", metricName, SQLProcessOld)
				// oracle old version. 11g
				if err := selectWrapper(ipt, &rows, SQLProcessOld, getMetricName(metricName, "oracle_process_old")); err != nil {
					l.Errorf("failed to collect old processes info: %s", err)
					return
				}
			} else {
				l.Errorf("failed to collect processes info: %s", err)
				return
			}
		}
	}

	opts := ipt.getKVsOpts()
	for _, row := range rows {
		kvs := ipt.getKVs()
		if row.PdbName.Valid {
			kvs = kvs.AddTag("pdb_name", row.PdbName.String)
		}

		if row.Program.Valid {
			kvs = kvs.AddTag("program", row.Program.String)
		}

		kvs = kvs.AddV2("pga_alloc_mem", row.PGAAllocMem, true).
			AddV2("pga_freeable_mem", row.PGAFreeableMem, true).
			AddV2("pga_max_mem", row.PGAMaxMem, true).
			AddV2("pga_used_mem", row.PGAUsedMem, true).
			AddV2("cpu_used", row.CPUUsed, true)

		if row.PID > 0 {
			kvs = kvs.Add("pid", row.PID, false, true)
		}

		pts = append(pts, point.NewPointV2(metricName, kvs, opts...))
	}

	if err := ipt.feeder.Feed(point.Metric,
		pts,
		dkio.WithCollectCost(time.Since(start)),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(inputName)); err != nil {
		l.Warnf("feeder.Feed: %s, ignored", err)
	}
}
