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

const SQLASMDiskgroup = `SELECT name, free_mb, total_mb, state, offline_disks FROM v$asm_diskgroup`

type asmDiskgroupRowDB struct {
	DiskgroupName sql.NullString  `db:"NAME"`
	FreeMB        sql.NullFloat64 `db:"FREE_MB"`
	TotalMB       sql.NullFloat64 `db:"TOTAL_MB"`
	State         sql.NullString  `db:"STATE"`
	OfflineDisks  sql.NullInt64   `db:"OFFLINE_DISKS"`
}

func asmDiskgroupStateValue(state string) int64 {
	switch strings.ToUpper(strings.TrimSpace(state)) {
	case "CONNECTED":
		return 1
	case "MOUNTED":
		return 2
	case "DISMOUNTED":
		return 3
	case "QUIESCING":
		return 4
	case "BROKEN":
		return 5
	default:
		return 0
	}
}

func (ipt *Input) collectOracleASMDiskgroup(ptsTime time.Time) {
	var (
		start      = time.Now()
		metricName = "oracle_asm_diskgroup"
		rows       = []asmDiskgroupRowDB{}
		pts        []*point.Point
	)

	if ipt.isMetricExclude(metricName) {
		l.Debugf("metric [%s] is excluded, ignored", metricName)
		return
	}

	if err := selectWrapper(ipt, &rows, SQLASMDiskgroup, getMetricName(metricName, "oracle_asm_diskgroup")); err != nil {
		l.Warnf("failed to collect ASM diskgroup info: %s", err)
		return
	}

	opts := ipt.getKVsOptsWithTime(ptsTime)
	for _, row := range rows {
		if !row.DiskgroupName.Valid {
			continue
		}

		kvs := ipt.getKVs().AddTag("asm_diskgroup_name", row.DiskgroupName.String)
		hasField := false
		if row.FreeMB.Valid {
			kvs = kvs.Set("asm_diskgroup_free_mb", row.FreeMB.Float64)
			hasField = true
		}
		if row.TotalMB.Valid {
			kvs = kvs.Set("asm_diskgroup_total_mb", row.TotalMB.Float64)
			hasField = true
		}
		if row.OfflineDisks.Valid {
			kvs = kvs.Set("asm_diskgroup_offline_disks", row.OfflineDisks.Int64)
			hasField = true
		}
		if row.State.Valid {
			kvs = kvs.Set("asm_diskgroup_state", asmDiskgroupStateValue(row.State.String))
			hasField = true
		}
		if !hasField {
			continue
		}

		pts = append(pts, point.NewPoint(metricName, kvs, opts...))
	}

	l.Debugf("%s: get %d points", metricName, len(pts))
	if len(pts) == 0 {
		return
	}

	if err := ipt.feeder.Feed(point.Metric,
		pts,
		dkio.WithCollectCost(time.Since(start)),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(inputName), dkio.WithInput(inputName),
		dkio.WithMeasurement(ipt.overrideMeasurement)); err != nil {
		l.Warnf("feeder.Feed: %s, ignored", err)
	}
}
