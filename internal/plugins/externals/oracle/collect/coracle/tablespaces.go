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

type tablespaceMetrics struct {
	x collectParameters
}

var _ ccommon.DBMetricsCollector = (*tablespaceMetrics)(nil)

func newTablespaceMetrics(opts ...collectOption) *tablespaceMetrics {
	m := &tablespaceMetrics{}

	for _, opt := range opts {
		if opt != nil {
			opt(&m.x)
		}
	}

	return m
}

func (m *tablespaceMetrics) Collect() (*point.Point, error) {
	l.Debug("Collect entry")

	tf, err := m.tablespaces()
	if err != nil {
		return nil, err
	}

	if tf.IsEmpty() {
		return nil, fmt.Errorf("tablespace empty data")
	}

	opt := &ccommon.BuildPointOpt{
		TF:         tf,
		MetricName: m.x.MetricName,
		Tags:       m.x.Ipt.tags,
		Host:       m.x.Ipt.host,
	}
	return ccommon.BuildPoint(l, opt), nil
}

// QUERY is the get tablespace info SQL query for Oracle 11g+.
//
//nolint:stylecheck
const QUERY = `SELECT
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

// QUERY_OLD is the get tablespace info SQL query for Oracle 11g and 11g-.
//
//nolint:stylecheck
const QUERY_OLD = `SELECT
  m.tablespace_name,
  NVL(m.used_space * t.block_size, 0) as used_space,
  m.tablespace_size * t.block_size as ts_size,
  NVL(m.used_percent, 0) as in_use,
  NVL2(m.used_space, 0, 1) as off_use
FROM
  dba_tablespace_usage_metrics m
  join dba_tablespaces t on m.tablespace_name = t.tablespace_name`

// RowDB is for Oracle 11g+.
type RowDB struct {
	PdbName        sql.NullString `db:"PDB_NAME"`
	TablespaceName string         `db:"TABLESPACE_NAME"`
	Used           float64        `db:"USED"`
	Size           float64        `db:"SIZE_"`
	InUse          float64        `db:"IN_USE"`
	Offline        float64        `db:"OFFLINE_"`
}

// RowDBOld is for Oracle 11g.
type RowDBOld struct {
	PdbName        sql.NullString `db:"PDB_NAME"`
	TablespaceName string         `db:"TABLESPACE_NAME"`
	UsedSpace      float64        `db:"USED_SPACE"`
	TSSize         float64        `db:"TS_SIZE"`
	InUse          float64        `db:"IN_USE"`
	OffUse         float64        `db:"OFF_USE"`
}

func (m *tablespaceMetrics) tablespaces() (*ccommon.TagField, error) {
	tf := ccommon.NewTagField()

	rows := []RowDB{}
	err := selectWrapper(m.x.Ipt, &rows, QUERY)
	if err != nil {
		l.Debug("tablespace: dpiStmt_execute: ORA-00942: table or view does not exist")

		if strings.Contains(err.Error(), "dpiStmt_execute: ORA-00942: table or view does not exist") {
			// oracle old version. 11g
			rowsOld := []RowDBOld{}
			if err = selectWrapper(m.x.Ipt, &rowsOld, QUERY_OLD); err != nil {
				return nil, fmt.Errorf("failed to collect old tablespace info: %w", err)
			}

			// map to new struct.
			rows = make([]RowDB, len(rowsOld))
			for k, v := range rowsOld {
				if v.PdbName.Valid {
					rows[k].PdbName = v.PdbName
				}
				rows[k].TablespaceName = v.TablespaceName
				rows[k].Used = v.UsedSpace
				rows[k].Size = v.TSSize
				rows[k].InUse = v.InUse
				rows[k].Offline = v.OffUse
			}
		} else {
			return nil, fmt.Errorf("failed to collect tablespace info: %w", err)
		}
	}

	for _, r := range rows {
		if r.PdbName.Valid {
			tf.AddTag(pdbName, r.PdbName.String)
		}

		tf.AddTag(tablespaceName, r.TablespaceName)

		tf.AddField("in_use", r.InUse, dic)
		tf.AddField("off_use", r.Offline, dic)
		tf.AddField("ts_size", r.Size, dic)
		tf.AddField("used_space", r.Used, dic)
	}

	return tf, nil
}
