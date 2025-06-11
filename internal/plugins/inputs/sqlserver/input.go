// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sqlserver collects SQL Server metrics.
package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	mssql "github.com/microsoft/go-mssqldb"
	"github.com/microsoft/go-mssqldb/msdsn"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var _ inputs.ElectionInput = (*Input)(nil)

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (ipt *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case ipt.pauseCh <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (*Input) SampleConfig() string {
	return sample
}

func (*Input) Catalog() string {
	return catalogName
}

func (*Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
}

func (*Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pScrpit,
	}
	return pipelineMap
}

//nolint:lll
func (ipt *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"SQLServer log": `2021-05-28 10:46:07.78 spid10s     0 transactions rolled back in database 'msdb' (4:0). This is an informational message only. No user action is required`,
		},
	}
}

func (ipt *Input) getPerformanceCounters() {
	ctx, cancel := context.WithTimeout(context.Background(), ipt.timeoutDuration)
	defer cancel()
	rows, err := ipt.db.QueryContext(ctx, sqlServerPerformanceCounters)
	if err != nil {
		l.Error(err.Error())
		ipt.lastErr = err
		return
	}

	defer func() {
		if err := rows.Close(); err != nil {
			l.Warnf("Close: %s, ignored", err)
		}
	}()

	// measurement, sqlserver_host, object_name, counter_name, instance_name, cntr_value
	columns, err := rows.Columns()
	if err != nil {
		l.Error(err.Error())
		ipt.lastErr = err
		return
	}

	rawBytesColumns := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(columns))

	for i := range rawBytesColumns {
		scanArgs[i] = &rawBytesColumns[i]
	}

	tags := make(map[string]string)
	for k, v := range ipt.Tags {
		tags[k] = v
	}

	var measurement string
	var counterName string

	// metric collect
	for rows.Next() {
		fields := make(map[string]interface{})
		// scan for every column
		if err := rows.Scan(scanArgs...); err != nil {
			l.Warnf("Scan: %s, ignored", err)
			continue
		}

		found, index := findInSlice(columns, "measurement")
		if found {
			measurement = string(rawBytesColumns[index])
		} else {
			measurement = "sqlserver_performance"
			l.Warnf("measurement not found, use default: sqlserver_performance")
		}

		found, index = findInSlice(columns, "counter_name")
		if found {
			counterName = string(rawBytesColumns[index])
			if mappedName, ok := counterNameMap[counterName]; ok {
				counterName = mappedName
			}
		} else {
			counterName = "cntr_value"
			l.Warnf("counter_name not found")
		}

		for i, key := range columns {
			if key == "measurement" {
				continue
			}

			raw := string(rawBytesColumns[i])

			if key == "cntr_value" {
				// the raw value is a number and the key is cntr_value, store in fields
				if v, err := strconv.ParseFloat(raw, 64); err == nil {
					if v > float64(math.MaxInt64) {
						l.Warnf("%s exceed maxint64: %d > %d, ignored", key, v, int64(math.MaxInt64))
						continue
					}
					// store the counter_name and cntr_value as fields
					if counterName == "buffer_cache_hit_ratio" {
						fields[counterName] = v
					} else {
						fields[counterName] = int64(v)
					}

					// remain the original format for "cntr_value": cntr_value
					fields[key] = int64(v)
				} else {
					l.Warnf("parse %s failed: %s", key, raw)
				}
			} else {
				str := strings.TrimSuffix(raw, "\\")
				tags[key] = str
			}
		}
		if len(fields) == 0 {
			continue
		}

		var opts []point.Option

		opts = point.DefaultMetricOptions()

		if ipt.Election {
			opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
		}

		// metric build
		m := Performance{}
		fields = getMetricFields(fields, m.Info())

		point := point.NewPointV2(measurement,
			append(point.NewTags(tags), point.NewKVs(fields)...), opts...)

		if len(fields) > 0 {
			ipt.collectCache = append(ipt.collectCache, point)
		}
	}
}

func findInSlice(slice []string, key string) (bool, int) {
	for index, value := range slice {
		if value == key {
			return true, index
		}
	}
	return false, -1
}

func getMetricFields(fields map[string]interface{}, info *inputs.MeasurementInfo) map[string]interface{} {
	if info == nil {
		return fields
	}
	newFields := map[string]interface{}{}

	for k, v := range fields {
		if _, ok := info.Fields[k]; ok {
			newFields[k] = v
		}
	}

	return newFields
}

func (ipt *Input) GetPipeline() []tailer.Option {
	opts := []tailer.Option{
		tailer.WithSource(inputName),
		tailer.WithService(inputName),
	}
	if ipt.Log != nil {
		opts = append(opts, tailer.WithPipeline(ipt.Log.Pipeline))
	}
	return opts
}

func (ipt *Input) initDB() error {
	query := url.Values{}
	if ipt.AllowTLS10 {
		// Because go1.18 defaults client-sids's TLS minimum version to TLS 1.2,
		// we need to configure MinVersion manually to enable TLS 1.0 and TLS 1.1.
		query.Add("tlsmin", "1.0")
	}

	if len(ipt.ConnectionParameters) > 0 {
		paramsQuery, err := url.ParseQuery(ipt.ConnectionParameters)
		if err != nil {
			return fmt.Errorf("parse connection_parameters failed: %w", err)
		}

		for k, v := range paramsQuery {
			query.Set(k, v[0])
		}
	}

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(ipt.User, ipt.Password),
		Host:     ipt.Host,
		RawQuery: query.Encode(),
	}

	dsn := u.String()
	cfg, err := msdsn.Parse(dsn)
	if err != nil {
		return fmt.Errorf("msdsn.Parse(%s): %w", dsn, err)
	}

	if ipt.InstanceName != "" {
		cfg.Instance = ipt.InstanceName
	}

	conn := mssql.NewConnectorConfig(cfg)
	db := sql.OpenDB(conn)

	ctx, cancel := context.WithTimeout(context.Background(), ipt.timeoutDuration)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close() //nolint:errcheck,gosec
		return err
	}

	ipt.db = db
	return nil
}

func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	opts := []tailer.Option{
		tailer.WithSource(inputName),
		tailer.WithService(inputName),
		tailer.WithPipeline(ipt.Log.Pipeline),
		tailer.WithIgnoreStatus(ipt.Log.IgnoreStatus),
		tailer.WithCharacterEncoding(ipt.Log.CharacterEncoding),
		tailer.EnableMultiline(true),
		tailer.WithMaxMultilineLength(int64(float64(config.Cfg.Dataway.MaxRawBodySize) * 0.8)),
		tailer.WithMultilinePatterns([]string{`^\d{4}-\d{2}-\d{2}`}),
		tailer.WithGlobalTags(inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")),
		tailer.EnableDebugFields(config.Cfg.EnableDebugFields),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opts...)
	if err != nil {
		l.Error(err)
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
		)
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_sqlserver"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

func (ipt *Input) runCustomQuery(query *customQuery) {
	if query == nil {
		return
	}

	// use input interval as default
	// use custom query interval if set
	if query.Interval.Duration > 0 {
		query.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, query.Interval.Duration)
	}

	tick := time.NewTicker(query.Interval.Duration)
	defer tick.Stop()

	ptsTime := ntp.Now()
	for {
		if ipt.pause {
			l.Debugf("not leader, custom query skipped")
		} else {
			start := time.Now()
			// collect custom query
			l.Debugf("start collecting custom query, metric name: %s", query.Metric)

			if res, err := ipt.query(query.Metric, query.SQL); err != nil {
				l.Warnf("collect custom query [%s] failed: %s", query.SQL, err.Error())
			} else {
				opt := ipt.getKVsOpts()
				opt = append(opt, point.WithTimestamp(ptsTime.UnixNano()))
				pts := []*point.Point{}
				for _, row := range res {
					kvs := ipt.getKVs()
					for _, tag := range query.Tags {
						if _, ok := row[tag]; ok {
							kvs = kvs.AddTag(tag, fmt.Sprintf("%v", *row[tag]))
						} else {
							l.Warnf("specified tag %s not found", tag)
						}
					}

					for _, field := range query.Fields {
						if _, ok := row[field]; ok {
							kvs = kvs.Add(field, *row[field], false, true)
						} else {
							l.Warn("specified field %s not found", field)
						}
					}

					if kvs.FieldCount() > 0 {
						pts = append(pts, point.NewPointV2(query.Metric, kvs, opt...))
					}
				}

				if len(pts) > 0 {
					if err := ipt.feeder.FeedV2(point.Metric, pts,
						dkio.WithCollectCost(time.Since(start)),
						dkio.WithElection(ipt.Election),
						dkio.WithInputName(customQueryFeedName),
					); err != nil {
						ipt.feeder.FeedLastError(err.Error(),
							metrics.WithLastErrorInput(customQueryFeedName),
							metrics.WithLastErrorCategory(point.Metric),
						)
						l.Errorf("feed custom query failed: %s", err.Error())
					}
				}
			}
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info("custom query exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Info("custom query return")
			return

		case tt := <-tick.C:
			ptsTime = inputs.AlignTime(tt, ptsTime, query.Interval.Duration)
		}
	}
}

func (ipt *Input) runCustomQueries() {
	if len(ipt.CustomQuery) == 0 {
		return
	}

	l.Infof("start to run custom queries, total %d queries", len(ipt.CustomQuery))

	g := goroutine.NewGroup(goroutine.Option{
		Name:         "sqlserver_custom_query",
		PanicTimes:   6,
		PanicTimeout: 10 * time.Second,
	})
	for _, q := range ipt.CustomQuery {
		func(q *customQuery) {
			g.Go(func(ctx context.Context) error {
				ipt.runCustomQuery(q)
				return nil
			})
		}(q)
	}
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("sqlserver start")

	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)

	if ipt.Election {
		ipt.opt = point.WithExtraTags(datakit.GlobalElectionTags())
	} else {
		ipt.opt = point.WithExtraTags(datakit.GlobalHostTags())
	}

	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	ipt.init()

	// Init DB until OK.
	for {
		if err := ipt.initDB(); err != nil {
			ipt.FeedErrUpMetric()
			ipt.FeedCoByErr(err)
			l.Errorf("initDB: %s", err.Error())
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
			)
		} else {
			break
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Info("sqlserver exit")
			return
		case <-ipt.semStop.Wait():
			l.Info("sqlserver exit")
			return
		case ipt.pause = <-ipt.pauseCh:
			// nil
		}
	}

	defer func() {
		if err := ipt.db.Close(); err != nil {
			l.Warnf("Close: %s", err)
		}
	}()

	// run custom queries
	ipt.runCustomQueries()

	ipt.ptsTime = ntp.Now()

	for {
		if ipt.pause {
			l.Debugf("not leader, skipped")
		} else {
			ipt.setUpState()
			l.Debugf("start to collect")
			ipt.getMetric()
			if len(ipt.collectCache) > 0 {
				err := ipt.feeder.FeedV2(point.Metric, ipt.collectCache,
					dkio.WithCollectCost(time.Since(ipt.start)),
					dkio.WithElection(ipt.Election),
					dkio.WithInputName(inputName),
				)
				ipt.collectCache = ipt.collectCache[:0]
				if err != nil {
					ipt.lastErr = err
					l.Errorf(err.Error())
				}
			}

			if len(ipt.loggingCollectCache) > 0 {
				err := ipt.feeder.FeedV2(point.Logging, ipt.loggingCollectCache,
					dkio.WithCollectCost(time.Since(ipt.start)),
					dkio.WithElection(ipt.Election),
					dkio.WithInputName(loggingFeedName),
				)
				ipt.loggingCollectCache = ipt.loggingCollectCache[:0]
				if err != nil {
					ipt.lastErr = err
					l.Errorf(err.Error())
				}
			}

			if ipt.lastErr != nil {
				ipt.feeder.FeedLastError(ipt.lastErr.Error(),
					metrics.WithLastErrorInput(inputName),
				)
				ipt.lastErr = nil

				ipt.setErrUpState()
			}
			ipt.FeedUpMetric()
			ipt.FeedCoPts()
		}

		select {
		case tt := <-tick.C:
			ipt.ptsTime = inputs.AlignTime(tt, ipt.ptsTime, ipt.Interval.Duration)

		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Info("sqlserver exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Info("sqlserver return")
			return
		}
	}
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
		l.Info("sqlserver log exit")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) getMetric() {
	now := ntp.Now()
	collectInterval := 10 * time.Minute
	if !ipt.start.IsZero() {
		collectInterval = now.Sub(ipt.start)
	}
	ipt.start = now

	// simple metric points
	for k, v := range ipt.collectQuery {
		ipt.handRow(k, v, false)
	}

	// simple logging points
	for k, v := range ipt.collectLoggingQuery {
		if strings.Contains(v, "__COLLECT_INTERVAL_SECONDS__") {
			v = strings.ReplaceAll(v, "__COLLECT_INTERVAL_SECONDS__", fmt.Sprintf("%.0f", collectInterval.Seconds()))
		}
		if strings.Contains(v, "__DATABASE__") {
			v = strings.ReplaceAll(v, "__DATABASE__", ipt.Database)
		}
		ipt.handRow(k, v, true)
	}

	// collectFuncs collect metrics that can't be collected by simple SQL query.
	for k, v := range ipt.collectFuncs {
		if err := v(); err != nil {
			l.Warnf("collect measurement [%s] error: %s", k, err.Error())
		}
	}

	// get performance counters metrics
	ipt.getPerformanceCounters()
}

func (ipt *Input) handRow(name, query string, isLogging bool) {
	ctx, cancel := context.WithTimeout(context.Background(), ipt.timeoutDuration)
	defer cancel()
	start := time.Now()
	rows, err := ipt.db.QueryContext(ctx, query)
	if err != nil {
		l.Error(err.Error())
		ipt.lastErr = err
		return
	}

	defer rows.Close() //nolint:errcheck

	if err := rows.Err(); err != nil {
		l.Errorf("rows.Err: %s", err)
		return
	}

	metricName, sqlName := getMetricNames(name)

	if len(metricName) > 0 {
		sqlQueryCostSummary.WithLabelValues(metricName, sqlName).Observe(float64(time.Since(start)) / float64(time.Second))
	}

	OrderedColumns, err := rows.Columns()
	if err != nil {
		l.Error(err.Error())
		ipt.lastErr = err
		return
	}

	category := point.Metric
	if isLogging {
		category = point.Logging
	}
	opts := ipt.getKVsOpts(category)
	for rows.Next() {
		var columnVars []interface{}
		// var fields = make(map[string]interface{})
		// store the column name with its *interface{}
		columnMap := make(map[string]*interface{})

		for _, column := range OrderedColumns {
			columnMap[column] = new(interface{})
		}
		// populate the array of interface{} with the pointers in the right order
		for i := 0; i < len(columnMap); i++ {
			columnVars = append(columnVars, columnMap[OrderedColumns[i]])
		}
		// deconstruct array of variables and send to Scan
		err := rows.Scan(columnVars...)
		if err != nil {
			l.Error(err.Error())
			ipt.lastErr = err
			return
		}
		measurement := ""
		kvs := ipt.getKVs()
		for header, val := range columnMap {
			if str, ok := (*val).(string); ok {
				if header == "measurement" {
					measurement = str
					continue
				}

				trimText := strings.TrimSuffix(str, "\\")
				if isLogging {
					kvs = kvs.Add(header, trimText, false, true)
				} else {
					kvs = kvs.AddTag(header, trimText)
				}
			} else if t, ok := (*val).(time.Time); ok {
				kvs = kvs.Add(header, t.UnixMilli(), false, true)
			} else {
				if *val == nil {
					continue
				}
				kvs = kvs.Add(header, *val, false, true)
			}
		}
		if kvs.FieldCount() == 0 {
			continue
		}
		if ipt.filterOutDBName(kvs.GetTag("database_name")) {
			continue
		}

		if isLogging {
			kvs = kvs.AddTag("status", "info")
		}

		kvs = transformData(measurement, kvs)

		point := point.NewPointV2(measurement,
			kvs, opts...)

		if isLogging {
			ipt.loggingCollectCache = append(ipt.loggingCollectCache, point)
		} else {
			ipt.collectCache = append(ipt.collectCache, point)
		}
	}
}

// filterOutDBName filters out metrics according to their database_name tag.
// Metrics with database_name tag specified in db_filter are filtered out and not fed to IO.
func (ipt *Input) filterOutDBName(name string) bool {
	if len(ipt.dbFilterMap) == 0 {
		return false
	}

	if _, filterOut := ipt.dbFilterMap[name]; filterOut {
		l.Debugf("filter out metric from db: %s", name)
		return true
	}
	return false
}

func (ipt *Input) init() {
	if len(ipt.Database) == 0 {
		ipt.Database = "master"
	}

	collectFuncs := map[string]func() error{
		"sqlserver_database_files": ipt.getDatabaseFilesMetrics,
	}

	ipt.collectFuncs = map[string]func() error{}
	ipt.collectLoggingQuery = map[string]string{}
	ipt.collectQuery = map[string]string{}

	for k, v := range query {
		ipt.collectQuery[k] = v
	}

	for k, v := range loggingQuery {
		ipt.collectLoggingQuery[k] = v
	}

	for k, v := range collectFuncs {
		ipt.collectFuncs[k] = v
	}

	// exclude metric
	for _, v := range ipt.MetricExcludeList {
		delete(ipt.collectQuery, v)
		delete(ipt.collectLoggingQuery, v)
		delete(ipt.collectFuncs, v)
	}

	var err error
	ipt.timeoutDuration, err = time.ParseDuration(ipt.Timeout)
	if err != nil {
		ipt.timeoutDuration = 30 * time.Second
	}

	ipt.initDBFilterMap()

	// init Tags and set host tag
	if ipt.Tags == nil {
		ipt.Tags = make(map[string]string)
	}
	if _, ok := ipt.Tags["host"]; !ok {
		host := getHostTagIfNotLoopback(ipt.Host)
		if len(host) > 0 {
			ipt.Tags["host"] = host
		}
	}
}

func (ipt *Input) getDatabaseFilesMetrics() error {
	data, err := ipt.query("sqlserver_database_files", fmt.Sprintf(`use [%s];
		select file_id,type as file_type,physical_name,state_desc,size,state
		from sys.database_files
	`, ipt.Database))
	if err != nil {
		return err
	}

	m := &DatabaseFilesMeasurement{
		MetricMeasurment: MetricMeasurment{},
	}
	info := m.Info()

	for _, row := range data {
		tags := make(map[string]string)
		fields := make(map[string]interface{})

		for k, v := range ipt.Tags {
			tags[k] = v
		}

		tags["database"] = ipt.Database

		for k, v := range row {
			switch k {
			case "size":
				if size, err := getValue[int64](v); err != nil {
					return err
				} else {
					fields[k] = 8 * size
				}
			default:
				if _, ok := info.Tags[k]; ok {
					tags[k] = fmt.Sprintf("%v", *v)
				}
			}
		}

		m.Measurement.fields = fields
		m.Measurement.tags = tags
		m.Measurement = Measurement{
			tags:     tags,
			fields:   fields,
			name:     info.Name,
			election: ipt.Election,
		}
		ipt.collectCache = append(ipt.collectCache, m.Point())
	}

	return nil
}

func (ipt *Input) query(name, sql string) (resRows []map[string]*interface{}, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), ipt.timeoutDuration)
	defer cancel()
	start := time.Now()
	rows, err := ipt.db.QueryContext(ctx, sql)
	if err != nil {
		return
	}
	defer rows.Close() //nolint:errcheck

	if err = rows.Err(); err != nil {
		return
	}

	metricName, sqlName := getMetricNames(name)

	if len(metricName) > 0 {
		sqlQueryCostSummary.WithLabelValues(metricName, sqlName).Observe(float64(time.Since(start)) / float64(time.Second))
	}

	columns, err := rows.Columns()
	if err != nil {
		return
	}

	resRows = make([]map[string]*interface{}, 0)

	for rows.Next() {
		var columnVars []interface{}
		columnMap := make(map[string]*interface{})

		for _, column := range columns {
			item := new(interface{})
			columnMap[column] = item
			columnVars = append(columnVars, item)
		}

		err = rows.Scan(columnVars...)
		if err != nil {
			return
		}
		resRows = append(resRows, columnMap)
	}

	return resRows, err
}

func (ipt *Input) initDBFilterMap() {
	if ipt.dbFilterMap == nil {
		ipt.dbFilterMap = make(map[string]struct{}, len(ipt.DBFilter))
	}
	for _, db := range ipt.DBFilter {
		ipt.dbFilterMap[db] = struct{}{}
	}
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&SqlserverMeasurment{},
		&Performance{},
		&WaitStatsCategorized{},
		&DatabaseIO{},
		&Schedulers{},
		&VolumeSpace{},
		&LockRow{},
		&LockTable{},
		&LockDead{},
		&LogicalIO{},
		&WorkerTime{},
		&DatabaseSize{},
		&DatabaseBackupMeasurement{},
		&DatabaseFilesMeasurement{},
		&customerObjectMeasurement{},
		&inputs.UpMeasurement{},
	}
}

// getValue return the of value with the type of the specified T.
func getValue[T any](rawValue interface{}) (res T, err error) {
	if v, ok := rawValue.(*interface{}); !ok {
		err = fmt.Errorf("value is not *interface{}")
		return
	} else if res, ok = (*v).(T); !ok {
		err = fmt.Errorf("value is not specified type")
		return
	}

	return
}

func getMetricNames(name string) (string, string) {
	names := strings.SplitN(name, ":", 2)
	metricName := ""
	sqlName := ""
	if len(names) == 1 {
		metricName = names[0]
		sqlName = names[0]
	} else if len(names) == 2 {
		metricName = names[0]
		sqlName = names[1]
	}

	return metricName, sqlName
}

func defaultInput() *Input {
	return &Input{
		Interval:    datakit.Duration{Duration: time.Second * 10},
		Election:    true,
		pauseCh:     make(chan bool, inputs.ElectionPauseChannelLength),
		semStop:     cliutils.NewSem(),
		dbFilterMap: make(map[string]struct{}, 0),
		feeder:      dkio.DefaultFeeder(),
		tagger:      datakit.DefaultGlobalTagger(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
