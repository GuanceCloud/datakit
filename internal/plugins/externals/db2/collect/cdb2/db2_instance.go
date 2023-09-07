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

type instanceMetrics struct {
	x collectParameters
}

var _ ccommon.DBMetricsCollector = (*instanceMetrics)(nil)

func newInstanceMetrics(opts ...collectOption) *instanceMetrics {
	m := &instanceMetrics{}

	for _, opt := range opts {
		if opt != nil {
			opt(&m.x)
		}
	}

	return m
}

func (m *instanceMetrics) Collect() (*point.Point, error) {
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
type instanceRowDB struct {
	Total_connections sql.NullInt64 `db:"TOTAL_CONNECTIONS"`
}

// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.sql.rtn.doc/doc/r0060770.html
// SELECT total_connections FROM TABLE(MON_GET_INSTANCE(-1))
//
//nolint:lll,stylecheck
var (
	INSTANCE_TABLE_COLUMNS = []string{
		"total_connections",
	}
	INSTANCE_TABLE = "SELECT " + strings.Join(INSTANCE_TABLE_COLUMNS, ",") + " FROM TABLE(MON_GET_INSTANCE(-1))"
)

//nolint:lll,stylecheck
func (m *instanceMetrics) collect() (*ccommon.TagField, error) {
	tf := ccommon.NewTagField()
	rows := []instanceRowDB{}

	err := selectWrapper(m.x.Ipt, &rows, INSTANCE_TABLE)
	if err != nil {
		return nil, fmt.Errorf("failed to collect instance: %w", err)
	}

	for _, r := range rows {
		tf.AddField("connection_active", r.Total_connections.Int64, nil)
	}

	return tf, nil
}
