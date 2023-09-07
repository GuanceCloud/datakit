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

type transactionLogMetrics struct {
	x collectParameters
}

var _ ccommon.DBMetricsCollector = (*transactionLogMetrics)(nil)

func newTransactionLogMetrics(opts ...collectOption) *transactionLogMetrics {
	m := &transactionLogMetrics{}

	for _, opt := range opts {
		if opt != nil {
			opt(&m.x)
		}
	}

	return m
}

func (m *transactionLogMetrics) Collect() (*point.Point, error) {
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
type transactionLogRowDB struct {
	Log_reads           sql.NullInt64 `db:"LOG_READS"`
	Log_writes          sql.NullInt64 `db:"LOG_WRITES"`
	Total_log_available sql.NullInt64 `db:"TOTAL_LOG_AVAILABLE"`
	Total_log_used      sql.NullInt64 `db:"TOTAL_LOG_USED"`
}

// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.sql.rtn.doc/doc/r0059253.html
// SELECT log_reads,log_writes,total_log_available,total_log_used FROM TABLE(MON_GET_TRANSACTION_LOG(-1))
//
//nolint:lll,stylecheck
var (
	TRANSACTION_LOG_TABLE_COLUMNS = []string{
		"log_reads",
		"log_writes",
		"total_log_available",
		"total_log_used",
	}
	TRANSACTION_LOG_TABLE = "SELECT " + strings.Join(TRANSACTION_LOG_TABLE_COLUMNS, ",") + " FROM TABLE(MON_GET_TRANSACTION_LOG(-1))"
)

// https://www.ibm.com/support/knowledgecenter/en/SSEPGG_11.1.0/com.ibm.db2.luw.admin.config.doc/doc/r0000239.html
//
//nolint:lll,stylecheck
const block_size = 4096

//nolint:lll,stylecheck
func (m *transactionLogMetrics) collect() (*ccommon.TagField, error) {
	tf := ccommon.NewTagField()
	rows := []transactionLogRowDB{}

	err := selectWrapper(m.x.Ipt, &rows, TRANSACTION_LOG_TABLE)
	if err != nil {
		return nil, fmt.Errorf("failed to collect instance: %w", err)
	}

	// Only 1 transaction log
	for _, r := range rows {
		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0002530.html
		used := r.Total_log_used.Int64
		tf.AddField("log_used", float64(used)/float64(block_size), nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0002531.html
		available := r.Total_log_available.Int64

		// Handle infinite log space
		var utilized float64
		if available == -1 {
			utilized = 0
		} else {
			utilized = float64(used) / float64(available) * 100
			available /= block_size
		}

		tf.AddField("log_available", available, nil)
		tf.AddField("log_utilized", utilized, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001278.html
		tf.AddField("log_reads", r.Log_reads.Int64, nil)

		// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001279.html
		tf.AddField("log_writes", r.Log_writes.Int64, nil)
	}

	return tf, nil
}
