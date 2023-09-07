// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux && amd64

package cdb2

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/db2/collect/ccommon"
)

type bufferPoolMetrics struct {
	x collectParameters
}

var _ ccommon.DBMetricsCollector = (*bufferPoolMetrics)(nil)

func newBufferPoolMetrics(opts ...collectOption) *bufferPoolMetrics {
	m := &bufferPoolMetrics{}

	for _, opt := range opts {
		if opt != nil {
			opt(&m.x)
		}
	}

	return m
}

func (m *bufferPoolMetrics) Collect() (*point.Point, error) {
	l.Debug("Collect entry")

	tf, err := m.collect()
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

//nolint:lll,stylecheck
type bufferPoolRowDB struct {
	BP_name                          sql.NullString `db:"BP_NAME"`
	Pool_async_col_lbp_pages_found   sql.NullInt64  `db:"POOL_ASYNC_COL_LBP_PAGES_FOUND"`
	Pool_async_data_lbp_pages_found  sql.NullInt64  `db:"POOL_ASYNC_DATA_LBP_PAGES_FOUND"`
	Pool_async_index_lbp_pages_found sql.NullInt64  `db:"POOL_ASYNC_INDEX_LBP_PAGES_FOUND"`
	Pool_async_xda_lbp_pages_found   sql.NullInt64  `db:"POOL_ASYNC_XDA_LBP_PAGES_FOUND"`
	Pool_col_gbp_l_reads             sql.NullInt64  `db:"POOL_COL_GBP_L_READS"`
	Pool_col_gbp_p_reads             sql.NullInt64  `db:"POOL_COL_GBP_P_READS"`
	Pool_col_l_reads                 sql.NullInt64  `db:"POOL_COL_L_READS"`
	Pool_col_lbp_pages_found         sql.NullInt64  `db:"POOL_COL_LBP_PAGES_FOUND"`
	Pool_col_p_reads                 sql.NullInt64  `db:"POOL_COL_P_READS"`
	Pool_data_gbp_l_reads            sql.NullInt64  `db:"POOL_DATA_GBP_L_READS"`
	Pool_data_gbp_p_reads            sql.NullInt64  `db:"POOL_DATA_GBP_P_READS"`
	Pool_data_l_reads                sql.NullInt64  `db:"POOL_DATA_L_READS"`
	Pool_data_lbp_pages_found        sql.NullInt64  `db:"POOL_DATA_LBP_PAGES_FOUND"`
	Pool_data_p_reads                sql.NullInt64  `db:"POOL_DATA_P_READS"`
	Pool_index_gbp_l_reads           sql.NullInt64  `db:"POOL_INDEX_GBP_L_READS"`
	Pool_index_gbp_p_reads           sql.NullInt64  `db:"POOL_INDEX_GBP_P_READS"`
	Pool_index_l_reads               sql.NullInt64  `db:"POOL_INDEX_L_READS"`
	Pool_index_lbp_pages_found       sql.NullInt64  `db:"POOL_INDEX_LBP_PAGES_FOUND"`
	Pool_index_p_reads               sql.NullInt64  `db:"POOL_INDEX_P_READS"`
	Pool_temp_col_l_reads            sql.NullInt64  `db:"POOL_TEMP_COL_L_READS"`
	Pool_temp_col_p_reads            sql.NullInt64  `db:"POOL_TEMP_COL_P_READS"`
	Pool_temp_data_l_reads           sql.NullInt64  `db:"POOL_TEMP_DATA_L_READS"`
	Pool_temp_data_p_reads           sql.NullInt64  `db:"POOL_TEMP_DATA_P_READS"`
	Pool_temp_index_l_reads          sql.NullInt64  `db:"POOL_TEMP_INDEX_L_READS"`
	Pool_temp_index_p_reads          sql.NullInt64  `db:"POOL_TEMP_INDEX_P_READS"`
	Pool_temp_xda_l_reads            sql.NullInt64  `db:"POOL_TEMP_XDA_L_READS"`
	Pool_temp_xda_p_reads            sql.NullInt64  `db:"POOL_TEMP_XDA_P_READS"`
	Pool_xda_gbp_l_reads             sql.NullInt64  `db:"POOL_XDA_GBP_L_READS"`
	Pool_xda_gbp_p_reads             sql.NullInt64  `db:"POOL_XDA_GBP_P_READS"`
	Pool_xda_l_reads                 sql.NullInt64  `db:"POOL_XDA_L_READS"`
	Pool_xda_lbp_pages_found         sql.NullInt64  `db:"POOL_XDA_LBP_PAGES_FOUND"`
	Pool_xda_p_reads                 sql.NullInt64  `db:"POOL_XDA_P_READS"`
}

// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.sql.rtn.doc/doc/r0053942.html
// SELECT bp_name,pool_async_col_lbp_pages_found,pool_async_data_lbp_pages_found,pool_async_index_lbp_pages_found,pool_async_xda_lbp_pages_found,pool_col_gbp_l_reads,pool_col_gbp_p_reads,pool_col_l_reads,pool_col_lbp_pages_found,pool_col_p_reads,pool_data_gbp_l_reads,pool_data_gbp_p_reads,pool_data_l_reads,pool_data_lbp_pages_found,pool_data_p_reads,pool_index_gbp_l_reads,pool_index_gbp_p_reads,pool_index_l_reads,pool_index_lbp_pages_found,pool_index_p_reads,pool_temp_col_l_reads,pool_temp_col_p_reads,pool_temp_data_l_reads,pool_temp_data_p_reads,pool_temp_index_l_reads,pool_temp_index_p_reads,pool_temp_xda_l_reads,pool_temp_xda_p_reads,pool_xda_gbp_l_reads,pool_xda_gbp_p_reads,pool_xda_l_reads,pool_xda_lbp_pages_found,pool_xda_p_reads FROM TABLE(MON_GET_BUFFERPOOL(NULL, -1))
//
//nolint:lll,stylecheck
var (
	BUFFER_POOL_TABLE_COLUMNS = []string{
		"bp_name",
		"pool_async_col_lbp_pages_found",
		"pool_async_data_lbp_pages_found",
		"pool_async_index_lbp_pages_found",
		"pool_async_xda_lbp_pages_found",
		"pool_col_gbp_l_reads",
		"pool_col_gbp_p_reads",
		"pool_col_l_reads",
		"pool_col_lbp_pages_found",
		"pool_col_p_reads",
		"pool_data_gbp_l_reads",
		"pool_data_gbp_p_reads",
		"pool_data_l_reads",
		"pool_data_lbp_pages_found",
		"pool_data_p_reads",
		"pool_index_gbp_l_reads",
		"pool_index_gbp_p_reads",
		"pool_index_l_reads",
		"pool_index_lbp_pages_found",
		"pool_index_p_reads",
		"pool_temp_col_l_reads",
		"pool_temp_col_p_reads",
		"pool_temp_data_l_reads",
		"pool_temp_data_p_reads",
		"pool_temp_index_l_reads",
		"pool_temp_index_p_reads",
		"pool_temp_xda_l_reads",
		"pool_temp_xda_p_reads",
		"pool_xda_gbp_l_reads",
		"pool_xda_gbp_p_reads",
		"pool_xda_l_reads",
		"pool_xda_lbp_pages_found",
		"pool_xda_p_reads",
	}
	BUFFER_POOL_TABLE = "SELECT " + strings.Join(BUFFER_POOL_TABLE_COLUMNS, ",") + " FROM TABLE(MON_GET_BUFFERPOOL(NULL, -1))"
)

//nolint:lll,stylecheck
func (m *bufferPoolMetrics) collect() (*ccommon.TagField, error) {
	tf := ccommon.NewTagField()
	rows := []bufferPoolRowDB{}

	err := selectWrapper(m.x.Ipt, &rows, BUFFER_POOL_TABLE)
	if err != nil {
		return nil, fmt.Errorf("failed to collect buffer pool: %w", err)
	}

	// Hit ratio formulas:
	// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0056871.html
	for _, r := range rows {
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0002256.html
		tf.AddTag("bp_name", r.BP_name.String)

		// Column-organized pages

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0060858.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0060874.html
		column_reads_physical := r.Pool_col_p_reads.Int64 + r.Pool_temp_col_p_reads.Int64
		tf.AddField("bufferpool_column_reads_physical", column_reads_physical, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0060763.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0060873.html
		column_reads_logical := r.Pool_col_l_reads.Int64 + r.Pool_temp_col_l_reads.Int64
		tf.AddField("bufferpool_column_reads_logical", column_reads_logical, nil)

		// Submit total
		tf.AddField("bufferpool_column_reads_total", column_reads_physical+column_reads_logical, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0060857.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0060850.html
		column_pages_found := r.Pool_col_lbp_pages_found.Int64 - r.Pool_async_col_lbp_pages_found.Int64

		var column_hit_percent float64
		if column_reads_logical > 0 {
			column_hit_percent = float64(column_pages_found) / float64(column_reads_logical) * 100
		} else {
			column_hit_percent = 0
		}
		tf.AddField("bufferpool_column_hit_percent", column_hit_percent, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0060855.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0060856.html
		group_column_reads_logical := r.Pool_col_gbp_l_reads.Int64
		group_column_pages_found := group_column_reads_logical - r.Pool_col_gbp_p_reads.Int64

		//  Submit group ratio if in a pureScale environment
		if group_column_reads_logical > 0 {
			group_column_hit_percent := group_column_pages_found / group_column_reads_logical * 100
			tf.AddField("bufferpool_group_column_hit_percent", group_column_hit_percent, nil)
		}

		// Data pages

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001236.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0011300.html
		data_reads_physical := r.Pool_data_p_reads.Int64 + r.Pool_temp_data_p_reads.Int64
		tf.AddField("bufferpool_data_reads_physical", data_reads_physical, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001235.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0011302.html
		data_reads_logical := r.Pool_data_l_reads.Int64 + r.Pool_temp_data_l_reads.Int64
		tf.AddField("bufferpool_data_reads_logical", data_reads_logical, nil)

		// Submit total
		tf.AddField("bufferpool_data_reads_total", data_reads_physical+data_reads_logical, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0056487.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0056493.html
		data_pages_found := r.Pool_data_lbp_pages_found.Int64 - r.Pool_async_data_lbp_pages_found.Int64

		var data_hit_percent float64
		if data_reads_logical > 0 {
			data_hit_percent = float64(data_pages_found) / float64(data_reads_logical) * 100
		} else {
			data_hit_percent = 0
		}
		tf.AddField("bufferpool_data_hit_percent", data_hit_percent, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0056485.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0056486.html
		group_data_reads_logical := r.Pool_data_gbp_l_reads.Int64
		group_data_pages_found := group_data_reads_logical - r.Pool_data_gbp_p_reads.Int64

		// Submit group ratio if in a pureScale environment
		var group_data_hit_percent float64
		if group_data_reads_logical > 0 {
			group_data_hit_percent = float64(group_data_pages_found) / float64(group_data_reads_logical) * 100
			tf.AddField("bufferpool_group_data_hit_percent", group_data_hit_percent, nil)
		}

		// Index pages

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001239.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0011301.html
		index_reads_physical := r.Pool_index_p_reads.Int64 + r.Pool_temp_index_p_reads.Int64
		tf.AddField("bufferpool_index_reads_physical", index_reads_physical, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001238.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0011303.html
		index_reads_logical := r.Pool_index_l_reads.Int64 + r.Pool_temp_index_l_reads.Int64
		tf.AddField("bufferpool_index_reads_logical", index_reads_logical, nil)

		// Submit total
		tf.AddField("bufferpool_index_reads_total", index_reads_physical+index_reads_logical, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0056243.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0056496.html
		index_pages_found := r.Pool_index_lbp_pages_found.Int64 - r.Pool_async_index_lbp_pages_found.Int64

		var index_hit_percent float64
		if index_reads_logical > 0 {
			index_hit_percent = float64(index_pages_found) / float64(index_reads_logical) * 100
		} else {
			index_hit_percent = 0
		}
		tf.AddField("bufferpool_index_hit_percent", index_hit_percent, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0056488.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0056489.html
		group_index_reads_logical := r.Pool_index_gbp_l_reads.Int64
		group_index_pages_found := group_index_reads_logical - r.Pool_index_gbp_p_reads.Int64

		// Submit group ratio if in a pureScale environment
		var group_index_hit_percent float64
		if group_index_reads_logical > 0 {
			group_index_hit_percent = float64(group_index_pages_found) / float64(group_index_reads_logical) * 100
			tf.AddField("bufferpool_group_index_hit_percent", group_index_hit_percent, nil)
		}

		// XML storage object (XDA) pages

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0022730.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0022739.html
		xda_reads_physical := r.Pool_xda_p_reads.Int64 + r.Pool_temp_xda_p_reads.Int64
		tf.AddField("bufferpool_xda_reads_physical", xda_reads_physical, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0022731.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0022738.html
		xda_reads_logical := r.Pool_xda_l_reads.Int64 + r.Pool_temp_xda_l_reads.Int64
		tf.AddField("bufferpool_xda_reads_logical", xda_reads_logical, nil)

		// Submit total
		tf.AddField("bufferpool_xda_reads_total", xda_reads_physical+xda_reads_logical, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0058666.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0058670.html
		xda_pages_found := r.Pool_xda_lbp_pages_found.Int64 - r.Pool_async_xda_lbp_pages_found.Int64

		var xda_hit_percent float64
		if xda_reads_logical > 0 {
			xda_hit_percent = float64(xda_pages_found) / float64(xda_reads_logical) * 100
		} else {
			xda_hit_percent = 0
		}
		tf.AddField("bufferpool_xda_hit_percent", xda_hit_percent, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0058664.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0058665.html
		group_xda_reads_logical := r.Pool_xda_gbp_l_reads.Int64
		group_xda_pages_found := group_xda_reads_logical - r.Pool_xda_gbp_p_reads.Int64

		// Submit group ratio if in a pureScale environment
		var group_xda_hit_percent float64
		if group_xda_reads_logical > 0 {
			group_xda_hit_percent = float64(group_xda_pages_found) / float64(group_xda_reads_logical) * 100
			tf.AddField("bufferpool_group_xda_hit_percent", group_xda_hit_percent, nil)
		}

		// Compute overall stats
		reads_physical := column_reads_physical + data_reads_physical + index_reads_physical + xda_reads_physical
		tf.AddField("bufferpool_reads_physical", reads_physical, nil)

		reads_logical := column_reads_logical + data_reads_logical + index_reads_logical + xda_reads_logical
		tf.AddField("bufferpool_reads_logical", reads_logical, nil)

		reads_total := reads_physical + reads_logical
		tf.AddField("bufferpool_reads_total", reads_total, nil)

		var hit_percent float64
		if reads_logical > 0 {
			pages_found := column_pages_found + data_pages_found + index_pages_found + xda_pages_found
			hit_percent = float64(pages_found) / float64(reads_logical) * 100
		} else {
			hit_percent = 0
		}
		tf.AddField("bufferpool_hit_percent", hit_percent, nil)

		// Submit group ratio if in a pureScale environment
		group_reads_logical := group_column_reads_logical + group_data_reads_logical + group_index_reads_logical + group_xda_reads_logical

		if group_reads_logical > 0 {
			group_pages_found := group_column_pages_found + group_data_pages_found + group_index_pages_found + group_xda_pages_found
			group_hit_percent := float64(group_pages_found) / float64(group_reads_logical) * 100
			tf.AddField("bufferpool_group_hit_percent", group_hit_percent, nil)
		}
	}

	return tf, nil
}
