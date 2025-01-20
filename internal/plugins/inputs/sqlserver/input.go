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

// getCustomQueryMetrics collect custom SQL query metrics.
func (ipt *Input) getCustomQueryMetrics() {
	for _, customQuery := range ipt.CustomQuery {
		res, err := ipt.query(customQuery.SQL)
		if err != nil {
			l.Warnf("collect custom query [%s] failed: %s", customQuery.SQL, err.Error())
			continue
		}

		for _, row := range res {
			tags := map[string]string{}
			fields := map[string]interface{}{}

			setHostTagIfNotLoopback(tags, ipt.Host)
			for k, v := range ipt.Tags {
				tags[k] = v
			}

			for _, tag := range customQuery.Tags {
				if _, ok := row[tag]; ok {
					tags[tag] = fmt.Sprintf("%v", *row[tag])
				} else {
					l.Warnf("specified tag %s not found", tag)
				}
			}

			for _, field := range customQuery.Fields {
				if _, ok := row[field]; ok {
					fields[field] = *row[field]
				} else {
					l.Warn("specified field %s not found", field)
				}
			}

			m := MetricMeasurment{
				Measurement: Measurement{
					name:     customQuery.Metric,
					tags:     tags,
					fields:   fields,
					election: ipt.Election,
				},
			}

			collectCache = append(collectCache, m.Point())
		}
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
	setHostTagIfNotLoopback(tags, ipt.Host)
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
			collectCache = append(collectCache, point)
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

	cfg, err := msdsn.Parse(u.String())
	if err != nil {
		return err
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
		case ipt.pause = <-ipt.pauseCh:
			// nil
		}
	}

	defer func() {
		if err := ipt.db.Close(); err != nil {
			l.Warnf("Close: %s", err)
		}
	}()

	for {
		if ipt.pause {
			l.Debugf("not leader, skipped")
		} else {
			ipt.setUpState()
			l.Infof("start to collect")
			ipt.getMetric()
			if len(collectCache) > 0 {
				err := ipt.feeder.FeedV2(point.Metric, collectCache,
					dkio.WithCollectCost(time.Since(ipt.start)),
					dkio.WithElection(ipt.Election),
					dkio.WithInputName(inputName),
				)
				collectCache = collectCache[:0]
				if err != nil {
					ipt.lastErr = err
					l.Errorf(err.Error())
				}
			}

			if len(loggingCollectCache) > 0 {
				err := ipt.feeder.FeedV2(point.Logging, loggingCollectCache,
					dkio.WithCollectCost(time.Since(ipt.start)),
					dkio.WithElection(ipt.Election),
					dkio.WithInputName(loggingFeedName),
				)
				loggingCollectCache = loggingCollectCache[:0]
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
			select {
			case <-tick.C:
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
	now := time.Now()
	collectInterval := 10 * time.Minute
	if !ipt.start.IsZero() {
		collectInterval = now.Sub(ipt.start)
	}
	ipt.start = now

	// simple metric points
	for _, v := range query {
		ipt.handRow(v, now, false)
	}

	// simple logging points
	for _, v := range loggingQuery {
		if strings.Contains(v, "__COLLECT_INTERVAL_SECONDS__") {
			v = strings.ReplaceAll(v, "__COLLECT_INTERVAL_SECONDS__", fmt.Sprintf("%.0f", collectInterval.Seconds()))
		}
		if strings.Contains(v, "__DATABASE__") {
			v = strings.ReplaceAll(v, "__DATABASE__", ipt.Database)
		}
		ipt.handRow(v, now, true)
	}

	// collectFuncs collect metrics that can't be collected by simple SQL query.
	for k, v := range ipt.collectFuncs {
		if err := v(); err != nil {
			l.Warnf("collect measurement [%s] error: %s", k, err.Error())
		}
	}

	// custom query from the config
	ipt.getCustomQueryMetrics()

	// get performance counters metrics
	ipt.getPerformanceCounters()
}

func (ipt *Input) handRow(query string, ts time.Time, isLogging bool) {
	ctx, cancel := context.WithTimeout(context.Background(), ipt.timeoutDuration)
	defer cancel()
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

	OrderedColumns, err := rows.Columns()
	if err != nil {
		l.Error(err.Error())
		ipt.lastErr = err
		return
	}

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
		tags := make(map[string]string)
		setHostTagIfNotLoopback(tags, ipt.Host)
		for k, v := range ipt.Tags {
			tags[k] = v
		}
		fields := make(map[string]interface{})
		for header, val := range columnMap {
			if str, ok := (*val).(string); ok {
				if header == "measurement" {
					measurement = str
					continue
				}

				trimText := strings.TrimSuffix(str, "\\")
				if isLogging {
					fields[header] = trimText
				} else {
					tags[header] = trimText
				}
			} else if t, ok := (*val).(time.Time); ok {
				fields[header] = t.UnixMilli()
			} else {
				if *val == nil {
					continue
				}
				fields[header] = *val
			}
		}
		if len(fields) == 0 {
			continue
		}
		if ipt.filterOutDBName(tags) {
			continue
		}

		var opts []point.Option
		if isLogging {
			tags["status"] = "info"
			opts = point.DefaultLoggingOptions()
		} else {
			opts = point.DefaultMetricOptions()
		}

		if ipt.Election {
			opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
		}

		transformData(measurement, tags, fields)

		point := point.NewPointV2(measurement,
			append(point.NewTags(tags), point.NewKVs(fields)...), opts...)

		if isLogging {
			loggingCollectCache = append(loggingCollectCache, point)
		} else {
			collectCache = append(collectCache, point)
		}
	}
}

// filterOutDBName filters out metrics according to their database_name tag.
// Metrics with database_name tag specified in db_filter are filtered out and not fed to IO.
func (ipt *Input) filterOutDBName(tags map[string]string) bool {
	if len(ipt.dbFilterMap) == 0 {
		return false
	}
	db, has := tags["database_name"]
	if !has {
		return false
	}

	if _, filterOut := ipt.dbFilterMap[db]; filterOut {
		l.Debugf("filter out metric from db: %s", db)
		return true
	}
	return false
}

func (ipt *Input) init() {
	if len(ipt.Database) == 0 {
		ipt.Database = "master"
	}

	ipt.collectFuncs = map[string]func() error{
		"sqlserver_database_files": ipt.getDatabaseFilesMetrics,
	}
	var err error
	ipt.timeoutDuration, err = time.ParseDuration(ipt.Timeout)
	if err != nil {
		ipt.timeoutDuration = 30 * time.Second
	}

	ipt.initDBFilterMap()
}

func (ipt *Input) getDatabaseFilesMetrics() error {
	data, err := ipt.query(fmt.Sprintf(`use [%s];
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

		setHostTagIfNotLoopback(tags, ipt.Host)
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
		collectCache = append(collectCache, m.Point())
	}

	return nil
}

func (ipt *Input) query(sql string) (resRows []map[string]*interface{}, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), ipt.timeoutDuration)
	defer cancel()
	rows, err := ipt.db.QueryContext(ctx, sql)
	if err != nil {
		return
	}
	defer rows.Close() //nolint:errcheck

	if err = rows.Err(); err != nil {
		return
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
