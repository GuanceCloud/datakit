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

func (ipt *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:  inputName,
			Service: inputName,
			Pipeline: func() string {
				if ipt.Log != nil {
					return ipt.Log.Pipeline
				}
				return ""
			}(),
		},
	}
}

func (ipt *Input) initDB() error {
	connStr := fmt.Sprintf("sqlserver://%s:%s@%s?dial+timeout=3", url.PathEscape(ipt.User), url.PathEscape(ipt.Password), url.PathEscape(ipt.Host))
	cfg, _, err := msdsn.Parse(connStr)
	if err != nil {
		return err
	}
	if ipt.AllowTLS10 {
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

	ipt.db = db
	return nil
}

func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		Pipeline:          ipt.Log.Pipeline,
		GlobalTags:        ipt.Tags,
		IgnoreStatus:      ipt.Log.IgnoreStatus,
		CharacterEncoding: ipt.Log.CharacterEncoding,
		MultilinePatterns: []string{`^\d{4}-\d{2}-\d{2}`},
		Done:              ipt.semStop.Wait(),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opt)
	if err != nil {
		l.Error(err)
		ipt.feeder.FeedLastError(err.Error(),
			io.WithLastErrorInput(inputName),
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
			l.Errorf("initDB: %s", err.Error())
			ipt.feeder.FeedLastError(ipt.lastErr.Error(),
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
		case ipt.pause = <-ipt.pauseCh:
			// nil
		}
	}

	defer func() {
		if err := ipt.db.Close(); err != nil {
			l.Warnf("Close: %s", err)
		}

		if ipt.tail != nil {
			ipt.tail.Close()
		}
	}()

	for {
		if ipt.pause {
			l.Debugf("not leader, skipped")
		} else {
			ipt.getMetric()
			if len(collectCache) > 0 {
				err := ipt.feeder.Feed(inputName, point.Metric, collectCache, &io.Option{CollectCost: time.Since(ipt.start)})
				collectCache = collectCache[:0]
				if err != nil {
					ipt.lastErr = err
					l.Errorf(err.Error())
				}
			}

			if len(loggingCollectCache) > 0 {
				err := ipt.feeder.Feed(inputName, point.Logging, loggingCollectCache, &io.Option{CollectCost: time.Since(ipt.start)})
				loggingCollectCache = loggingCollectCache[:0]
				if err != nil {
					ipt.lastErr = err
					l.Errorf(err.Error())
				}
			}

			if ipt.lastErr != nil {
				ipt.feeder.FeedLastError(ipt.lastErr.Error(),
					io.WithLastErrorInput(inputName),
				)
				ipt.lastErr = nil
			}

			select {
			case <-tick.C:
			case <-datakit.Exit.Wait():
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
}

func (ipt *Input) handRow(query string, ts time.Time, isLogging bool) {
	rows, err := ipt.db.Query(query)
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
	rows, err := ipt.db.Query(sql)
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
