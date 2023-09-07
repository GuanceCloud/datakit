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
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/db2/collect/ccommon"
)

type tableSpaceMetrics struct {
	x               collectParameters
	tableSpaceState map[string]string // map[table_space]state
}

var _ ccommon.DBMetricsCollector = (*tableSpaceMetrics)(nil)

func newTableSpaceMetrics(opts ...collectOption) *tableSpaceMetrics {
	m := &tableSpaceMetrics{tableSpaceState: make(map[string]string)}

	for _, opt := range opts {
		if opt != nil {
			opt(&m.x)
		}
	}

	return m
}

func (m *tableSpaceMetrics) Collect() (*point.Point, error) {
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
type tableSpaceRowDB struct {
	Tbsp_name         sql.NullString `db:"TBSP_NAME"`
	Tbsp_page_size    sql.NullInt64  `db:"TBSP_PAGE_SIZE"`
	Tbsp_state        sql.NullString `db:"TBSP_STATE"`
	Tbsp_total_pages  sql.NullInt64  `db:"TBSP_TOTAL_PAGES"`
	Tbsp_usable_pages sql.NullInt64  `db:"TBSP_USABLE_PAGES"`
	Tbsp_used_pages   sql.NullInt64  `db:"TBSP_USED_PAGES"`
}

// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.sql.rtn.doc/doc/r0053943.html
// SELECT tbsp_name,tbsp_page_size,tbsp_state,tbsp_total_pages,tbsp_usable_pages,tbsp_used_pages FROM TABLE(MON_GET_TABLESPACE(NULL, -1))
//
//nolint:lll,stylecheck
var (
	TABLE_SPACE_TABLE_COLUMNS = []string{
		"tbsp_name",
		"tbsp_page_size",
		"tbsp_state",
		"tbsp_total_pages",
		"tbsp_usable_pages",
		"tbsp_used_pages",
	}
	TABLE_SPACE_TABLE = "SELECT " + strings.Join(TABLE_SPACE_TABLE_COLUMNS, ",") + " FROM TABLE(MON_GET_TABLESPACE(NULL, -1))"
)

//nolint:lll,stylecheck
func (m *tableSpaceMetrics) collect() (*ccommon.TagField, error) {
	tf := ccommon.NewTagField()
	rows := []tableSpaceRowDB{}

	err := selectWrapper(m.x.Ipt, &rows, TABLE_SPACE_TABLE)
	if err != nil {
		return nil, fmt.Errorf("failed to collect instance: %w", err)
	}

	// Utilization formulas:
	// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.sql.rtn.doc/doc/r0056516.html
	for _, r := range rows {
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001295.html
		tf.AddTag("tablespace_name", r.Tbsp_name.String)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0007534.html
		page_size := r.Tbsp_page_size.Int64

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0007539.html
		total_pages := r.Tbsp_total_pages.Int64
		tf.AddField("tablespace_size", total_pages*page_size, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0007540.html
		usable_pages := r.Tbsp_usable_pages.Int64
		tf.AddField("tablespace_usable", usable_pages*page_size, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0007541.html
		used_pages := r.Tbsp_used_pages.Int64
		tf.AddField("tablespace_used", used_pages*page_size, nil)

		// Percent utilized
		var utilized float64
		if usable_pages > 0 {
			utilized = float64(used_pages) / float64(usable_pages) * 100
		} else {
			utilized = 0
		}
		tf.AddField("tablespace_utilized", utilized, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0007533.html
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.dbobj.doc/doc/c0060111.html
		previousState, ok := m.tableSpaceState[r.Tbsp_name.String]
		if ok {
			if r.Tbsp_state.String != previousState {
				m.tableSpaceState[r.Tbsp_name.String] = r.Tbsp_state.String // update state.

				newTags := map[string]string{
					"source": inputName,
				}
				newTags = ccommon.MergeTags(newTags, m.x.Ipt.tags, m.x.Ipt.host)

				event := &ccommon.DFEvent{
					Name:        inputName,
					EventPrefix: ccommon.EventPrefixDb2,
					Tags:        newTags,
					Date:        time.Now().Unix(),
					Status:      "warning",
					Title:       "Table space state change",
					Message:     fmt.Sprintf("State of `%s` changed from `%s` to `%s`.", r.Tbsp_name.String, previousState, r.Tbsp_state.String),
				}

				go ccommon.FeedEvent(l, event, m.x.Ipt.datakitPostEventURL)
			}
		} else if r.Tbsp_state.Valid && len(r.Tbsp_state.String) > 0 {
			m.tableSpaceState[r.Tbsp_name.String] = r.Tbsp_state.String
		}
	}

	return tf, nil
}
