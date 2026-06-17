// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cibmi collects IBM i metrics through Db2 for i SQL services.
package cibmi

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/jmoiron/sqlx"

	cp "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/colorprint"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ibm_i/collect/ccommon"
)

const (
	inputName          = "ibm_i"
	passwordEnv        = "ENV_INPUT_IBM_I_PASSWORD"
	defaultDriver      = "odbc"
	defaultODBCDriver  = "IBM i Access ODBC Driver 64-bit"
	defaultRetryPeriod = 3 * time.Second
	minInterval        = 10 * time.Second
	maxInterval        = 15 * time.Minute
)

var l = logger.DefaultSLogger(inputName)

type MetricConfig struct {
	Enabled bool
}

type Input struct {
	DSN                  string
	Host                 string
	User                 string
	Password             string
	ODBCDriver           string
	Driver               string
	Interval             time.Duration
	QueryTimeout         time.Duration
	JobQueryTimeout      time.Duration
	SystemMQQueryTimeout time.Duration
	Election             bool
	Tags                 map[string]string
	Queries              []string
	SeverityThreshold    int
	MessageQueues        []string

	Metric MetricConfig

	db            *sqlx.DB
	systemInfo    *systemInfo
	metricPostURL string
	writeData     func(*logger.Logger, []byte, string) error
	connectDBFunc func() (*sqlx.DB, error)
	retryPeriod   time.Duration
}

func NewInput(opt *ccommon.Option) (*Input, error) {
	logPath := opt.Log
	if logPath == "" {
		logPath = filepath.Join(datakit.InstallDir, "externals", inputName+".log")
	}
	if err := logger.InitRoot(&logger.Option{
		Path: logPath, Level: opt.LogLevel, Flags: logger.OPT_DEFAULT,
	}); err != nil {
		cp.Println("set root log failed:", err.Error())
	}
	l = logger.SLogger(inputName)

	interval, err := time.ParseDuration(opt.Interval)
	if err != nil {
		return nil, fmt.Errorf("parse interval: %w", err)
	}
	queryTimeout, err := time.ParseDuration(opt.QueryTimeout)
	if err != nil {
		return nil, fmt.Errorf("parse query timeout: %w", err)
	}
	jobQueryTimeout, err := time.ParseDuration(opt.JobQueryTimeout)
	if err != nil {
		return nil, fmt.Errorf("parse job query timeout: %w", err)
	}
	systemMQQueryTimeout, err := time.ParseDuration(opt.SystemMQQueryTimeout)
	if err != nil {
		return nil, fmt.Errorf("parse system mq query timeout: %w", err)
	}
	metricEnabled, err := parseEnabled("metric-enabled", opt.MetricEnabled)
	if err != nil {
		return nil, err
	}
	queries, err := normalizeQueries(opt.Queries)
	if err != nil {
		return nil, err
	}

	password := opt.Password
	if envPassword := os.Getenv(passwordEnv); envPassword != "" {
		password = envPassword
	}

	ipt := &Input{
		DSN: opt.DSN, Host: opt.Host, User: opt.Username,
		Password: password, ODBCDriver: opt.Driver, Driver: defaultDriver,
		Interval: interval, QueryTimeout: queryTimeout, JobQueryTimeout: jobQueryTimeout,
		SystemMQQueryTimeout: systemMQQueryTimeout, Election: opt.Election,
		Tags: parseTags(opt.Tags), Queries: queries, SeverityThreshold: opt.SeverityThreshold,
		MessageQueues: normalizeNames(opt.MessageQueues),
		Metric:        MetricConfig{Enabled: metricEnabled},
		metricPostURL: ccommon.GetPostURL(opt.Election, ccommon.CategoryMetric,
			inputName, opt.DatakitHTTPHost, opt.DatakitHTTPPort),
		writeData: ccommon.WriteData,
	}
	if err := ipt.setup(); err != nil {
		return nil, err
	}
	return ipt, nil
}

func (ipt *Input) setup() error {
	if ipt.Host == "" && ipt.DSN != "" {
		ipt.Host = firstNonEmpty(connValue(ipt.DSN, "System"), connValue(ipt.DSN, "HOSTNAME"))
	}
	if ipt.DSN == "" && ipt.Host == "" {
		return fmt.Errorf("host is required")
	}
	if ipt.DSN == "" && ipt.User == "" {
		return fmt.Errorf("username is required")
	}
	if ipt.Driver == "" {
		ipt.Driver = defaultDriver
	}
	if ipt.ODBCDriver == "" {
		ipt.ODBCDriver = defaultODBCDriver
	}
	if ipt.Tags == nil {
		ipt.Tags = map[string]string{}
	}
	if len(ipt.Queries) == 0 {
		ipt.Queries = append([]string{}, defaultQueries...)
	}
	if ipt.SeverityThreshold <= 0 {
		ipt.SeverityThreshold = 50
	}
	if ipt.Interval <= 0 {
		ipt.Interval = time.Minute
	}
	ipt.Interval = protectInterval(ipt.Interval)
	if ipt.QueryTimeout <= 0 {
		ipt.QueryTimeout = 30 * time.Second
	}
	if ipt.JobQueryTimeout <= 0 {
		ipt.JobQueryTimeout = 240 * time.Second
	}
	if ipt.SystemMQQueryTimeout <= 0 {
		ipt.SystemMQQueryTimeout = 80 * time.Second
	}
	if ipt.writeData == nil {
		ipt.writeData = ccommon.WriteData
	}
	return nil
}

func protectInterval(value time.Duration) time.Duration {
	if value < minInterval {
		return minInterval
	}
	if value > maxInterval {
		return maxInterval
	}
	return value
}

func defaultInput() *Input {
	return &Input{
		Driver: defaultDriver, ODBCDriver: defaultODBCDriver, Interval: time.Minute,
		QueryTimeout: 30 * time.Second, JobQueryTimeout: 240 * time.Second,
		SystemMQQueryTimeout: 80 * time.Second, SeverityThreshold: 50,
		Election: true, Tags: map[string]string{}, Queries: append([]string{}, defaultQueries...),
		Metric: MetricConfig{Enabled: true}, writeData: ccommon.WriteData,
	}
}

func (ipt *Input) Run(stop <-chan os.Signal) {
	l.Infof("IBM i external collector start: host=%q driver=%q odbc_driver=%q interval=%s query_timeout=%s job_query_timeout=%s system_mq_query_timeout=%s queries=%v metric_enabled=%t",
		ipt.Host, ipt.Driver, ipt.ODBCDriver, ipt.Interval, ipt.QueryTimeout,
		ipt.JobQueryTimeout, ipt.SystemMQQueryTimeout, ipt.Queries, ipt.Metric.Enabled)
	if !ipt.connectUntilReady(stop) {
		return
	}
	defer ipt.closeDB()

	metricTicker := time.NewTicker(ipt.Interval)
	defer metricTicker.Stop()

	ipt.collectMetricOnce(time.Now())

	for {
		select {
		case ts := <-metricTicker.C:
			ipt.collectMetricOnce(ts)
		case <-stop:
			l.Info("IBM i external collector stopped")
			return
		}
	}
}

func (ipt *Input) connectUntilReady(stop <-chan os.Signal) bool {
	for {
		db, err := ipt.openDB()
		if err == nil {
			ipt.db = db
			l.Infof("IBM i connection established: host=%q driver=%q odbc_driver=%q", ipt.Host, ipt.Driver, ipt.ODBCDriver)
			return true
		}
		retryPeriod := ipt.retryPeriod
		if retryPeriod <= 0 {
			retryPeriod = defaultRetryPeriod
		}
		ccommon.ReportErrorf(inputName, l, "connect failed: %v; retry in %s", err, retryPeriod)
		timer := time.NewTimer(retryPeriod)
		select {
		case <-timer.C:
		case <-stop:
			timer.Stop()
			return false
		}
	}
}

func (ipt *Input) openDB() (*sqlx.DB, error) {
	if ipt.connectDBFunc != nil {
		return ipt.connectDBFunc()
	}
	return ipt.connectDB()
}

func (ipt *Input) connectDB() (*sqlx.DB, error) {
	connStr := ipt.connectionString()
	l.Debugf("opening IBM i connection: host=%q driver=%q odbc_driver=%q timeout=%s dsn_provided=%t",
		ipt.Host, ipt.Driver, ipt.ODBCDriver, ipt.QueryTimeout, ipt.DSN != "")
	db, err := sql.Open(ipt.Driver, connStr)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(2)
	ctx, cancel := context.WithTimeout(context.Background(), ipt.QueryTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close() //nolint:errcheck
		return nil, err
	}
	l.Debugf("IBM i ping succeeded: host=%q", ipt.Host)
	return sqlx.NewDb(db, ipt.Driver), nil
}

func (ipt *Input) connectionString() string {
	if ipt.DSN != "" {
		return ipt.DSN
	}
	driver := strings.Trim(ipt.ODBCDriver, "{}")
	connStr := fmt.Sprintf("Driver={%s};", driver)
	if strings.TrimSpace(ipt.Host) != "" {
		connStr += "System=" + strings.TrimSpace(ipt.Host) + ";"
	}
	if strings.TrimSpace(ipt.User) != "" {
		connStr += "UID=" + strings.TrimSpace(ipt.User) + ";"
	}
	if ipt.Password != "" {
		connStr += "PWD=" + ipt.Password + ";"
	}
	return connStr
}

func (ipt *Input) closeDB() {
	if ipt.db != nil {
		if err := ipt.db.Close(); err != nil {
			l.Warnf("close IBM i connection: %v", err)
		}
		ipt.db = nil
	}
}

func (ipt *Input) reportFailure(category string, err error) {
	ccommon.ReportErrorf(inputName, l, "%s failed: %v", category, err)
	if ipt.db == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), ipt.QueryTimeout)
	defer cancel()
	if pingErr := ipt.db.PingContext(ctx); pingErr != nil {
		ipt.closeDB()
	}
}

func (ipt *Input) ensureConnection() bool {
	if ipt.db != nil {
		return true
	}
	db, err := ipt.openDB()
	if err != nil {
		ccommon.ReportErrorf(inputName, l, "reconnect failed: %v", err)
		return false
	}
	ipt.db = db
	l.Infof("IBM i connection re-established: host=%q driver=%q odbc_driver=%q", ipt.Host, ipt.Driver, ipt.ODBCDriver)
	return true
}

func (ipt *Input) collectMetricOnce(ts time.Time) {
	if !ipt.Metric.Enabled || !ipt.ensureConnection() {
		return
	}
	start := time.Now()
	l.Debugf("collect metric start: queries=%v", ipt.Queries)
	ctx := context.Background()
	metrics, _, err := ipt.collectMetricPoints(ctx, ts)
	if err != nil {
		ipt.reportFailure("collect metric", err)
	}
	if err := ipt.reportPoints(metrics, ipt.metricPostURL); err != nil {
		ccommon.ReportErrorf(inputName, l, "write metric failed: %v", err)
		return
	}
	l.Debugf("collect metric done: points=%d cost=%s", len(metrics), time.Since(start))
}

func (ipt *Input) reportPoints(pts []*point.Point, target string) error {
	if len(pts) == 0 {
		l.Debug("skip write metric: no points")
		return nil
	}
	lines := make([][]byte, 0, len(pts))
	for _, pt := range pts {
		lines = append(lines, []byte(pt.LineProto()))
	}
	start := time.Now()
	err := ipt.writeData(l, bytes.Join(lines, []byte("\n")), target)
	if err != nil {
		return err
	}
	l.Debugf("write metric succeeded: points=%d cost=%s", len(pts), time.Since(start))
	return nil
}

func (ipt *Input) selectContext(parent context.Context, dest any, query string, timeout time.Duration, args ...any) error {
	if timeout <= 0 {
		timeout = ipt.QueryTimeout
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	queryName := queryNameFromSQL(query)
	start := time.Now()
	l.Debugf("query start: name=%s timeout=%s args=%d", queryName, timeout, len(args))
	err := ipt.db.SelectContext(ctx, dest, query, args...)
	cost := time.Since(start)
	if err != nil {
		l.Debugf("query failed: name=%s cost=%s err=%v", queryName, cost, err)
		return err
	}
	rows := selectedRows(dest)
	l.Debugf("query done: name=%s rows=%d cost=%s", queryName, rows, cost)
	return nil
}

func selectedRows(dest any) int {
	value := reflect.ValueOf(dest)
	if value.Kind() != reflect.Pointer || value.IsNil() {
		return 0
	}
	elem := value.Elem()
	if elem.Kind() != reflect.Slice {
		return 0
	}
	return elem.Len()
}

func queryNameFromSQL(query string) string {
	switch query {
	case systemInfoSQL:
		return "system_info"
	case baseDiskUsage72SQL:
		return "base_disk_usage_72"
	case baseDiskUsage73SQL:
		return "base_disk_usage_73"
	case diskUsageSQL:
		return queryDiskUsage
	case cpuUsageSQL:
		return queryCPUUsage
	case jobQJobStatusSQL:
		return queryJobQJobStatus
	case activeJobStatusSQL:
		return queryActiveJobStatus
	case jobMemoryUsageSQL:
		return queryJobMemoryUsage
	case memoryInfoSQL:
		return queryMemoryInfo
	case subsystemSQL:
		return querySubsystem
	case jobQueueSQL:
		return queryJobQueue
	default:
		if strings.Contains(query, "MESSAGE_QUEUE_INFO") {
			return queryMessageQueue
		}
		return "custom"
	}
}

func parseTags(value string) map[string]string {
	tags := map[string]string{}
	for _, item := range strings.Split(value, ";") {
		key, val, ok := strings.Cut(item, "=")
		if ok && strings.TrimSpace(key) != "" {
			tags[strings.TrimSpace(key)] = strings.TrimSpace(val)
		}
	}
	return tags
}

func normalizeNames(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, raw := range values {
		value := strings.ToUpper(strings.TrimSpace(raw))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func parseEnabled(name, value string) (bool, error) {
	if value == "" {
		return true, nil
	}
	enabled, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("parse %s: %w", name, err)
	}
	return enabled, nil
}

func (ipt *Input) buildUpPoint(ts time.Time, up int64) *point.Point {
	tags := ipt.commonTags()
	tags["job"] = inputName
	tags["instance"] = firstNonEmpty(ipt.Host, ipt.systemHostName())
	var kvs point.KVs
	addTags(&kvs, tags)
	kvs = kvs.Set("up", up)
	return point.NewPoint("collector", kvs, metricOptions(ts)...)
}
