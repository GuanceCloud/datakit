// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package doris

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const dorisObjectMeasurementName = "database"

const (
	dorisFrontendsQuery = "SHOW FRONTENDS"
	dorisBackendsQuery  = "SHOW BACKENDS"
)

type dorisObjectMeasurement struct{}

//nolint:lll
func (*dorisObjectMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: dorisObjectMeasurementName,
		Cat:  point.Object,
		Desc: "Doris object metrics.",
		Tags: map[string]interface{}{
			"host":              &inputs.TagInfo{Desc: "The configured Doris FE host."},
			"server":            &inputs.TagInfo{Desc: "The Doris FE SQL endpoint. The value is `host:query_port`."},
			"version":           &inputs.TagInfo{Desc: "The Doris FE version."},
			"database_instance": &inputs.TagInfo{Desc: "Doris instance identifier from configured tag, current FE HostName, or the FE SQL endpoint."},
			"name":              &inputs.TagInfo{Desc: "The object identifier."},
			"database_type":     &inputs.TagInfo{Desc: "The type of the database. The value is `Doris`."},
			"port":              &inputs.TagInfo{Desc: "The FE query port."},
		},
		Fields: map[string]interface{}{
			"message":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Summary of Doris FE/BE and metric source information."},
			"qps":            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Gauge, Desc: "The number of queries executed by Doris per second. Collected from FE metrics."},
			"avg_query_time": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.DurationMS, Desc: "The average query latency in milliseconds. Collected from FE metrics when summary count/sum is available."},
		},
	}
}

type dorisObjectMessage struct {
	Frontends []map[string]string `json:"frontends,omitempty"`
	Backends  []map[string]string `json:"backends,omitempty"`
}

func (m dorisObjectMessage) String() string {
	b, _ := json.Marshal(m)
	return string(b)
}

type dorisFEMetrics struct {
	QPS            float64
	AvgQueryTimeMS float64

	HasQPS            bool
	HasAvgQueryTimeMS bool
}

func (ipt *Input) collectDatabaseObject() error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), ipt.ConnectTimeout.Duration)
	defer cancel()

	frontends, err := ipt.queryRows(ctx, dorisFrontendsQuery)
	if err != nil {
		return fmt.Errorf("show frontends: %w", err)
	}

	backends, err := ipt.queryRows(ctx, dorisBackendsQuery)
	if err != nil {
		l.Warnf("show backends failed: %s", err.Error())
	}

	msg := dorisObjectMessage{
		Frontends: frontends,
		Backends:  backends,
	}

	// FE metrics are optional enrichments for the database object.
	feMetrics, err := ipt.collectFEMetrics()
	if err != nil {
		l.Warnf("collect doris FE metrics failed: %s", err.Error())
	}

	kvs := ipt.getObjectKVs(msg)
	if feMetrics.HasQPS {
		kvs = kvs.Set("qps", feMetrics.QPS)
	}

	if feMetrics.HasAvgQueryTimeMS {
		kvs = kvs.Set("avg_query_time", feMetrics.AvgQueryTimeMS)
	}

	opts := point.DefaultObjectOptions()
	opts = append(opts, point.WithTimestamp(ipt.ptsTime.UnixNano()))
	pt := point.NewPoint(dorisObjectMeasurementName, kvs, opts...)

	if err := ipt.feeder.Feed(point.Object, []*point.Point{pt},
		dkio.WithCollectCost(time.Since(start)),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(objectFeedName),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Object),
		)
		l.Errorf("feed doris object failed: %s", err.Error())
	}

	return nil
}

func (ipt *Input) initDatabaseInfo(ctx context.Context) error {
	frontends, err := ipt.queryRows(ctx, dorisFrontendsQuery)
	if err != nil {
		return fmt.Errorf("show frontends: %w", err)
	} else {
		// CurrentConnected identifies the FE reached by this SQL connection.
		currentFE := selectCurrentFrontend(frontends)
		if v := selectDorisVersion(currentFE); v != "" {
			ipt.version = v
		}
		if ipt.databaseInstance == "" {
			ipt.databaseInstance = selectDorisHostname(currentFE)
		}
	}

	if ipt.databaseInstance == "" {
		ipt.databaseInstance = ipt.server()
	}
	if ipt.mergedTags == nil {
		ipt.mergedTags = map[string]string{}
	}
	ipt.mergedTags["database_instance"] = ipt.databaseInstance

	return nil
}

func (ipt *Input) getObjectKVs(msg dorisObjectMessage) point.KVs {
	kvs := point.KVs{}
	for k, v := range ipt.mergedTags {
		kvs = kvs.AddTag(k, v)
	}

	objectName := ipt.server()
	if ipt.databaseInstance != "" && ipt.databaseInstance != ipt.server() {
		objectName = fmt.Sprintf("%s-%s", ipt.server(), ipt.databaseInstance)
	}

	kvs = kvs.AddTag("database_type", dorisType).
		AddTag("name", objectName).
		AddTag("port", strconv.Itoa(ipt.Port)).
		Set("message", msg.String())

	if ipt.version != "" {
		kvs = kvs.AddTag("version", ipt.version)
	}

	return kvs
}

func (ipt *Input) queryRows(ctx context.Context, query string) ([]map[string]string, error) {
	rows, err := ipt.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query %q: %w", query, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			l.Warnf("close rows failed: %s", err.Error())
		}
	}()

	return rowsToMaps(rows)
}

func rowsToMaps(rows *sql.Rows) ([]map[string]string, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("get columns: %w", err)
	}

	rawColumns := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range rawColumns {
		scanArgs[i] = &rawColumns[i]
	}

	var result []map[string]string
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		row := make(map[string]string, len(columns))
		for i, col := range columns {
			row[col] = string(rawColumns[i])
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return result, nil
}

func (ipt *Input) collectFEMetrics() (dorisFEMetrics, error) {
	if ipt.FEURL == "" {
		return dorisFEMetrics{}, nil
	}

	req, err := http.NewRequest(http.MethodGet, ipt.FEURL, nil)
	if err != nil {
		return dorisFEMetrics{}, fmt.Errorf("new FE metrics request: %w", err)
	}

	client, err := ipt.metricHTTPClient()
	if err != nil {
		return dorisFEMetrics{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return dorisFEMetrics{}, fmt.Errorf("request FE metrics %s: %w", ipt.FEURL, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			l.Warnf("close FE metrics response failed: %s", err.Error())
		}
	}()

	if resp.StatusCode/100 != 2 {
		return dorisFEMetrics{}, fmt.Errorf("request FE metrics %s: status %s", ipt.FEURL, resp.Status)
	}

	metrics, err := parseDorisFEMetrics(resp.Body)
	if err != nil {
		return dorisFEMetrics{}, err
	}

	return metrics, nil
}

func (ipt *Input) metricHTTPClient() (*http.Client, error) {
	client := &http.Client{Timeout: ipt.ConnectTimeout.Duration}
	if ipt.MetricTLS == nil {
		return client, nil
	}

	tlsConfig, err := createTLSConf(ipt.MetricTLS.TLSCA, ipt.MetricTLS.TLSCert, ipt.MetricTLS.TLSKey)
	if err != nil {
		return nil, fmt.Errorf("create metric tls config failed: %w", err)
	}

	tlsConfig.InsecureSkipVerify = ipt.MetricTLS.InsecureSkipVerify
	if ipt.MetricTLS.AllowTLS10 {
		tlsConfig.MinVersion = tls.VersionTLS10
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig
	client.Transport = transport

	return client, nil
}

func parseDorisFEMetrics(r io.Reader) (dorisFEMetrics, error) {
	var parser expfmt.TextParser
	families, err := parser.TextToMetricFamilies(r)
	if err != nil {
		return dorisFEMetrics{}, fmt.Errorf("parse FE metrics: %w", err)
	}

	result := dorisFEMetrics{}
	for name, family := range families {
		switch name {
		case "doris_fe_qps":
			if v, ok := sumMetricFamilyValues(family); ok {
				result.QPS = v
				result.HasQPS = true
			}
		case "doris_fe_query_latency_ms":
			if v, ok := avgSummaryValue(family); ok {
				result.AvgQueryTimeMS = v
				result.HasAvgQueryTimeMS = true
			}
		}
	}

	return result, nil
}

func sumMetricFamilyValues(family *dto.MetricFamily) (float64, bool) {
	var (
		total float64
		found bool
	)
	for _, metric := range family.GetMetric() {
		if v, ok := metricValue(metric); ok {
			total += v
			found = true
		}
	}
	return total, found
}

func avgSummaryValue(family *dto.MetricFamily) (float64, bool) {
	// Prefer the real average from summary sum/count; fall back to p50 if needed.
	for _, metric := range family.GetMetric() {
		summary := metric.GetSummary()
		if summary == nil {
			continue
		}
		if summary.GetSampleCount() > 0 {
			return summary.GetSampleSum() / float64(summary.GetSampleCount()), true
		}
	}

	for _, metric := range family.GetMetric() {
		summary := metric.GetSummary()
		if summary == nil {
			continue
		}
		for _, quantile := range summary.GetQuantile() {
			if quantile.GetQuantile() == 0.5 {
				return quantile.GetValue(), true
			}
		}
	}

	return 0, false
}

func metricValue(metric *dto.Metric) (float64, bool) {
	switch {
	case metric.GetGauge() != nil:
		return metric.GetGauge().GetValue(), true
	case metric.GetCounter() != nil:
		return metric.GetCounter().GetValue(), true
	case metric.GetUntyped() != nil:
		return metric.GetUntyped().GetValue(), true
	}
	return 0, false
}

func selectDorisVersion(currentFE map[string]string) string {
	if version := currentFE["Version"]; version != "" {
		return version
	}
	return ""
}

func selectCurrentFrontend(frontends []map[string]string) map[string]string {
	for _, fe := range frontends {
		if isTruthy(fe["CurrentConnected"]) {
			return fe
		}
	}
	return nil
}

func selectDorisHostname(currentFE map[string]string) string {
	if v := strings.TrimSpace(currentFE["HostName"]); v != "" {
		return v
	}
	return ""
}

func isTruthy(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "yes", "1", "alive", "y":
		return true
	default:
		return false
	}
}
