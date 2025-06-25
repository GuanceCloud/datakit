// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (

	//nolint:gosec
	"database/sql"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

// SQLTableSpace for Oracle 11g+.
const SQLTableSpace = `SELECT
  c.name pdb_name,
  t.tablespace_name tablespace_name,
  NVL(m.used_space * t.block_size, 0) used,
  NVL(m.tablespace_size * t.block_size, 0) size_,
  NVL(m.used_percent, 0) in_use,
  NVL2(m.used_space, 0, 1) offline_
FROM
  cdb_tablespace_usage_metrics m, cdb_tablespaces t, v$containers c
WHERE
  m.tablespace_name(+) = t.tablespace_name and c.con_id(+) = t.con_id`

// SQLTableSpaceOld for Oracle 11g and 11g-.
const SQLTableSpaceOld = `SELECT
  m.tablespace_name,
  NVL(m.used_space * t.block_size, 0) as used,
  m.tablespace_size * t.block_size as size_,
  NVL(m.used_percent, 0) as in_use,
  NVL2(m.used_space, 0, 1) as offline_
FROM
  dba_tablespace_usage_metrics m
  join dba_tablespaces t on m.tablespace_name = t.tablespace_name`

type tableSpaceRowDB struct {
	PdbName        sql.NullString `db:"PDB_NAME"`
	TablespaceName string         `db:"TABLESPACE_NAME"`
	Used           float64        `db:"USED"`
	Size           float64        `db:"SIZE_"`
	InUse          float64        `db:"IN_USE"`
	Offline        float64        `db:"OFFLINE_"`
}

func (ipt *Input) collectOracleTableSpace() {
	var (
		start      = time.Now()
		metricName = "oracle_tablespace"
		rows       = []tableSpaceRowDB{}
		pts        []*point.Point
	)

	if sql, ok := ipt.cacheSQL[metricName]; ok {
		if err := selectWrapper(ipt, &rows, sql, getMetricName(metricName, "oracle_tablespace")); err != nil {
			l.Errorf("failed to collect table space: %s", err)
			return
		}
	} else {
		err := selectWrapper(ipt, &rows, SQLTableSpace, getMetricName(metricName, "oracle_tablespace"))
		ipt.cacheSQL[metricName] = SQLTableSpace
		l.Infof("Query for metric [%s], sql: %s", metricName, SQLTableSpace)
		if err != nil {
			if strings.Contains(err.Error(), "ORA-00942: table or view does not exist") {
				ipt.cacheSQL[metricName] = SQLTableSpaceOld
				l.Infof("Query for metric [%s], sql: %s", metricName, SQLTableSpaceOld)
				// oracle old version. 11g
				if err := selectWrapper(ipt, &rows, SQLTableSpaceOld, getMetricName(metricName, "oracle_tablespace_old")); err != nil {
					l.Errorf("failed to collect old table space info: %s", err)
					return
				}
			} else {
				l.Errorf("failed to collect table space info: %s", err)
				return
			}
		}
	}

	opts := ipt.getKVsOpts()
	for _, row := range rows {
		kvs := ipt.getKVs()
		kvs = kvs.AddTag("tablespace_name", row.TablespaceName)
		if row.PdbName.Valid {
			kvs = kvs.AddTag("pdb_name", row.PdbName.String)
		}

		kvs = kvs.Add("in_use", row.InUse, false, true)
		kvs = kvs.Add("off_use", row.Offline, false, true)
		kvs = kvs.Add("ts_size", row.Size, false, true)
		kvs = kvs.Add("used_space", row.Used, false, true)

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
