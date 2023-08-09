// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sqlserver collects SQL Server metrics.
package sqlserver

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/denisenkom/go-mssqldb/msdsn"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

var _ inputs.ElectionInput = (*Input)(nil)

func (i *Input) ElectionEnabled() bool {
	return i.Election
}

func (i *Input) Pause() error {
	tick := time.NewTicker(inputs.ElectionPauseTimeout)
	defer tick.Stop()
	select {
	case i.pauseCh <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(inputs.ElectionResumeTimeout)
	defer tick.Stop()
	select {
	case i.pauseCh <- false:
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
		inputName: pipeline,
	}
	return pipelineMap
}

//nolint:lll
func (i *Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"SQLServer log": `2021-05-28 10:46:07.78 spid10s     0 transactions rolled back in database 'msdb' (4:0). This is an informational message only. No user action is required`,
		},
	}
}

// getCustomQueryMetrics collect custom SQL query metrics.
func (i *Input) getCustomQueryMetrics() {
	for _, customQuery := range i.CustomQuery {
		res, err := i.query(customQuery.SQL)
		if err != nil {
			l.Warnf("collect custom query [%s] failed: %s", customQuery.SQL, err.Error())
			continue
		}

		for _, row := range res {
			tags := map[string]string{}
			fields := map[string]interface{}{}

			setHostTagIfNotLoopback(tags, i.Host)

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
					election: i.Election,
				},
			}

			collectCache = append(collectCache, m.Point())
		}
	}
}

func (i *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if i.Log != nil {
					return i.Log.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (i *Input) initDB() error {
	connStr := fmt.Sprintf("sqlserver://%s:%s@%s?dial+timeout=3", url.PathEscape(i.User), url.PathEscape(i.Password), url.PathEscape(i.Host))
	cfg, _, err := msdsn.Parse(connStr)
	if err != nil {
		return err
	}
	if i.AllowTLS10 {
		// Because go1.18 defaults client-sids's TLS minimum version to TLS 1.2,
		// we need to configure MinVersion manually to enable TLS 1.0 and TLS 1.1.
		cfg.TLSConfig.MinVersion = tls.VersionTLS10
	}
	conn := mssql.NewConnectorConfig(cfg)
	db := sql.OpenDB(conn)

	if err := db.Ping(); err != nil {
		db.Close() //nolint:errcheck,gosec
		return err
	}

	i.db = db
	return nil
}

func (i *Input) RunPipeline() {
	if i.Log == nil || len(i.Log.Files) == 0 {
		return
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		Pipeline:          i.Log.Pipeline,
		GlobalTags:        i.Tags,
		IgnoreStatus:      i.Log.IgnoreStatus,
		CharacterEncoding: i.Log.CharacterEncoding,
		MultilinePatterns: []string{`^\d{4}-\d{2}-\d{2}`},
		Done:              i.semStop.Wait(),
	}

	var err error
	i.tail, err = tailer.NewTailer(i.Log.Files, opt)
	if err != nil {
		l.Error(err)
		i.feeder.FeedLastError(err.Error(),
			io.WithLastErrorInput(inputName),
		)
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_sqlserver"})
	g.Go(func(ctx context.Context) error {
		i.tail.Start()
		return nil
	})
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)
	l.Info("sqlserver start")

	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)

	if i.Election {
		i.opt = point.WithExtraTags(dkpt.GlobalElectionTags())
	} else {
		i.opt = point.WithExtraTags(dkpt.GlobalHostTags())
	}

	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()

	i.init()

	// Init DB until OK.
	for {
		if err := i.initDB(); err != nil {
			l.Errorf("initDB: %s", err.Error())
			i.feeder.FeedLastError(i.lastErr.Error(),
				io.WithLastErrorInput(inputName),
			)
		} else {
			break
		}

		select {
		case <-tick.C:
		case <-datakit.Exit.Wait():
			l.Info("sqlserver exit")
			return
		case i.pause = <-i.pauseCh:
			// nil
		}
	}

	defer func() {
		if err := i.db.Close(); err != nil {
			l.Warnf("Close: %s", err)
		}

		if i.tail != nil {
			i.tail.Close()
		}
	}()

	for {
		if i.pause {
			l.Debugf("not leader, skipped")
		} else {
			i.getMetric()
			if len(collectCache) > 0 {
				err := i.feeder.Feed(inputName, point.Metric, collectCache, &io.Option{CollectCost: time.Since(i.start)})
				collectCache = collectCache[:0]
				if err != nil {
					i.lastErr = err
					l.Errorf(err.Error())
				}
			}

			if len(loggingCollectCache) > 0 {
				err := i.feeder.Feed(inputName, point.Logging, loggingCollectCache, &io.Option{CollectCost: time.Since(i.start)})
				loggingCollectCache = loggingCollectCache[:0]
				if err != nil {
					i.lastErr = err
					l.Errorf(err.Error())
				}
			}

			if i.lastErr != nil {
				i.feeder.FeedLastError(i.lastErr.Error(),
					io.WithLastErrorInput(inputName),
				)
				i.lastErr = nil
			}

			select {
			case <-tick.C:
			case <-datakit.Exit.Wait():
				l.Info("sqlserver exit")
				return

			case <-i.semStop.Wait():
				i.exit()
				l.Info("sqlserver return")
				return
			}
		}
	}
}

func (i *Input) exit() {
	if i.tail != nil {
		i.tail.Close()
		l.Info("sqlserver log exit")
	}
}

func (i *Input) Terminate() {
	if i.semStop != nil {
		i.semStop.Close()
	}
}

func (i *Input) getMetric() {
	now := time.Now()
	collectInterval := 10 * time.Minute
	if !i.start.IsZero() {
		collectInterval = now.Sub(i.start)
	}
	i.start = now

	// simple metric points
	for _, v := range query {
		i.handRow(v, now, false)
	}

	// simple logging points
	for _, v := range loggingQuery {
		if strings.Contains(v, "__COLLECT_INTERVAL_SECONDS__") {
			v = strings.ReplaceAll(v, "__COLLECT_INTERVAL_SECONDS__", fmt.Sprintf("%.0f", collectInterval.Seconds()))
		}
		if strings.Contains(v, "__DATABASE__") {
			v = strings.ReplaceAll(v, "__DATABASE__", i.Database)
		}
		i.handRow(v, now, true)
	}

	// collectFuncs collect metrics that can't be collected by simple SQL query.
	for k, v := range i.collectFuncs {
		if err := v(); err != nil {
			l.Warnf("collect measurement [%s] error: %s", k, err.Error())
		}
	}

	// custom query from the config
	i.getCustomQueryMetrics()
}

func (i *Input) handRow(query string, ts time.Time, isLogging bool) {
	rows, err := i.db.Query(query)
	if err != nil {
		l.Error(err.Error())
		i.lastErr = err
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
		i.lastErr = err
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
			i.lastErr = err
			return
		}
		measurement := ""
		tags := make(map[string]string)
		setHostTagIfNotLoopback(tags, i.Host)
		for k, v := range i.Tags {
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
		if i.filterOutDBName(tags) {
			continue
		}

		var opts []point.Option
		if isLogging {
			tags["status"] = "info"
			opts = point.DefaultLoggingOptions()
		} else {
			opts = point.DefaultMetricOptions()
		}

		if i.Election {
			opts = append(opts, point.WithExtraTags(dkpt.GlobalElectionTags()))
		}

		transformData(measurement, tags, fields)

		point := point.NewPointV2([]byte(measurement),
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
func (i *Input) filterOutDBName(tags map[string]string) bool {
	if len(i.dbFilterMap) == 0 {
		return false
	}
	db, has := tags["database_name"]
	if !has {
		return false
	}

	if _, filterOut := i.dbFilterMap[db]; filterOut {
		l.Debugf("filter out metric from db: %s", db)
		return true
	}
	return false
}

func (i *Input) init() {
	if len(i.Database) == 0 {
		i.Database = "master"
	}

	i.collectFuncs = map[string]func() error{
		"sqlserver_database_files": i.getDatabaseFilesMetrics,
	}

	i.initDBFilterMap()
}

func (i *Input) getDatabaseFilesMetrics() error {
	data, err := i.query(fmt.Sprintf(`use [%s];
		select file_id,type as file_type,physical_name,state_desc,size,state
		from sys.database_files
	`, i.Database))
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

		setHostTagIfNotLoopback(tags, i.Host)
		for k, v := range i.Tags {
			tags[k] = v
		}

		tags["database"] = i.Database

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
			election: i.Election,
		}
		collectCache = append(collectCache, m.Point())
	}

	return nil
}

func (i *Input) query(sql string) (resRows []map[string]*interface{}, err error) {
	rows, err := i.db.Query(sql)
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

func (i *Input) initDBFilterMap() {
	if i.dbFilterMap == nil {
		i.dbFilterMap = make(map[string]struct{}, len(i.DBFilter))
	}
	for _, db := range i.DBFilter {
		i.dbFilterMap[db] = struct{}{}
	}
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
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
		feeder:      io.DefaultFeeder(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
