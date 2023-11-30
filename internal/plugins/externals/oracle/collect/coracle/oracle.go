// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package coracle contains collect Oracle code.
package coracle

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/jmoiron/sqlx"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/oracle/collect/ccommon"
)

const (
	inputName = "oracle"

	metricNameProcess    = "oracle_process"
	metricNameTablespace = "oracle_tablespace"
	metricNameSystem     = "oracle_system"
	metricNameLogging    = "oracle_log"

	pdbName        = "pdb_name"
	tablespaceName = "tablespace_name"
	programName    = "program"
)

var (
	l                = logger.DefaultSLogger(inputName)
	_ ccommon.IInput = (*Input)(nil)
)

type Input struct {
	interval      string
	user          string
	password      string
	desc          string
	host          string
	port          string
	serviceName   string
	tags          map[string]string
	election      bool
	SlowQueryTime time.Duration
	Query         []*customQuery

	db                    *sqlx.DB
	intervalDuration      time.Duration
	collectors            []ccommon.DBMetricsCollector
	collectorsLogging     []ccommon.DBMetricsCollector
	datakitPostURL        string
	datakitPostLoggingURL string

	// collected metrics - mysql custom queries
	mCustomQueries map[string][]map[string]interface{}
}

// customQuery contains custom sql query info.
// Same as struct 'customQuery' in internal/plugins/inputs/external/external.go.
type customQuery struct {
	SQL    string   `toml:"sql"`
	Metric string   `toml:"metric"`
	Tags   []string `toml:"tags"`
	Fields []string `toml:"fields"`

	MD5Hash string
}

func NewInput(infoMsgs []string, opt *ccommon.Option) ccommon.IInput {
	logPath := opt.Log
	if logPath == "" {
		logPath = filepath.Join(datakit.InstallDir, "externals", inputName+".log")
	}

	if err := logger.InitRoot(&logger.Option{
		Path:  logPath,
		Level: opt.LogLevel,
		Flags: logger.OPT_DEFAULT,
	}); err != nil {
		fmt.Println("set root log failed:", err.Error())
	}

	if opt.InstanceDesc != "" { // add description to logger
		l = logger.SLogger(inputName + "-" + opt.InstanceDesc)
	} else {
		l = logger.SLogger(inputName)
	}

	// Print former logs.
	for _, v := range infoMsgs {
		l.Info(v)
	}

	ipt := &Input{
		interval:    opt.Interval,
		user:        opt.Username,
		password:    opt.Password,
		desc:        opt.InstanceDesc,
		host:        opt.Host,
		port:        opt.Port,
		serviceName: opt.ServiceName,
		tags:        make(map[string]string),
		election:    opt.Election,
	}

	// ENV_INPUT_ORACLE_PASSWORD
	pwd := os.Getenv("ENV_INPUT_ORACLE_PASSWORD")
	if len(pwd) > 0 {
		ipt.password = pwd
	}

	if len(opt.SlowQueryTime) > 0 {
		du, err := time.ParseDuration(opt.SlowQueryTime)
		if err != nil {
			l.Warnf("bad slow query %s: %s, disable slow query", opt.SlowQueryTime, err.Error())
		} else {
			if du >= time.Millisecond {
				ipt.SlowQueryTime = du
			} else {
				l.Warnf("slow query time %v larger than 1 millisecond, skip", du)
			}
		}
	}

	l.Infof("opt.CustomQueryFile = %s", opt.CustomQueryFile)
	if len(opt.CustomQueryFile) > 0 {
		ipt.Query = bytesToQuery(opt)
	}

	items := strings.Split(opt.Tags, ";")
	for _, item := range items {
		tagArr := strings.Split(item, "=")

		if len(tagArr) == 2 {
			tagKey := strings.Trim(tagArr[0], " ")
			tagVal := strings.Trim(tagArr[1], " ")
			if tagKey != "" {
				ipt.tags[tagKey] = tagVal
			}
		}
	}

	ipt.tags[inputName+"_service"] = ipt.serviceName
	ipt.tags[inputName+"_server"] = fmt.Sprintf("%s:%s", ipt.host, ipt.port)

	if ipt.interval != "" {
		du, err := time.ParseDuration(ipt.interval)
		if err != nil {
			l.Errorf("bad interval %s: %s, use default: 10m", ipt.interval, err.Error())
			ipt.intervalDuration = 10 * time.Minute
		} else {
			ipt.intervalDuration = du
		}
	}

	for {
		if err := ipt.ConnectDB(); err != nil {
			ccommon.ReportErrorf(inputName, l, inputName+" connect faild %v, retry each 3 seconds...", err)
			time.Sleep(time.Second * 3)
			continue
		}

		break
	}

	proMetric := newProcessMetrics(withInput(ipt), withMetricName(metricNameProcess))
	tsMetric := newTablespaceMetrics(withInput(ipt), withMetricName(metricNameTablespace))
	sysMetric := newSystemMetrics(withInput(ipt), withMetricName(metricNameSystem))
	customMetric := newCustomQueryCollector(withInput(ipt))

	ipt.collectors = append(ipt.collectors, proMetric, tsMetric, sysMetric, customMetric)
	l.Infof("collectors len = %d", len(ipt.collectors))

	ipt.datakitPostURL = ccommon.GetPostURL(
		opt.Election,
		ccommon.CategoryMetric, inputName, opt.DatakitHTTPHost,
		opt.DatakitHTTPPort,
	)
	l.Infof("Datakit post metric URL: %s", ipt.datakitPostURL)

	if ipt.SlowQueryTime > 0 {
		slowQueryLoggingR := newSlowQueryLogging(withInput(ipt), withMetricName(metricNameLogging))
		ipt.collectorsLogging = append(ipt.collectorsLogging, slowQueryLoggingR)

		ipt.datakitPostLoggingURL = ccommon.GetPostURL(
			opt.Election,
			ccommon.CategoryLogging, inputName, opt.DatakitHTTPHost,
			opt.DatakitHTTPPort,
		)
	}

	return ipt
}

func bytesToQuery(opt *ccommon.Option) []*customQuery {
	l.Debug("bytesToQuery entry")

	fi, err := os.Open(opt.CustomQueryFile)
	if err != nil {
		l.Errorf("os.Open() failed: %v", err)
		return nil
	}

	dec := gob.NewDecoder(fi)
	var query []*customQuery
	err = dec.Decode(&query)
	if err != nil {
		l.Errorf("dec.Decode() failed: %v", err)
		return nil
	}

	l.Debugf("query = %#v", query)

	return query
}

// ConnectDB establishes a connection to an Oracle instance and returns an open connection to the database.
func (ipt *Input) ConnectDB() error {
	db, err := sqlx.Open("godror",
		fmt.Sprintf("%s/%s@%s:%s/%s",
			ipt.user, ipt.password, ipt.host, ipt.port, ipt.serviceName))
	if err == nil {
		ipt.db = db
		return err
	}
	return nil
}

func (ipt *Input) q(s string) rows {
	rows, err := ipt.db.Query(s)
	if err != nil {
		l.Errorf(`query failed, sql (%q), error: %s, ignored`, s, err.Error())
		return nil
	}

	if err := rows.Err(); err != nil {
		closeRows(rows)
		l.Errorf(`query row failed, sql (%q), error: %s, ignored`, s, err.Error())
		return nil
	}

	return rows
}

// DB rows interface.
type rows interface {
	Next() bool
	Scan(...interface{}) error
	Close() error
	Err() error
	Columns() ([]string, error)
}

func closeRows(r rows) {
	if err := r.Close(); err != nil {
		l.Warnf("Close: %s, ignored", err)
	}
}

// CloseDB cleans up database resources used.
func (ipt *Input) CloseDB() {
	if ipt.db != nil {
		if err := ipt.db.Close(); err != nil {
			l.Warnf("failed to close "+inputName+" connection | server=[%s]: %s", ipt.host, err.Error())
		}
	}
}

// Run start collecting.
func (ipt *Input) Run() {
	l.Info("starting " + inputName + "...")

	tick := time.NewTicker(ipt.intervalDuration)
	defer ipt.db.Close() //nolint:errcheck
	defer tick.Stop()

	for {
		collectAndReport(&ipt.collectors, ipt.datakitPostURL)
		collectAndReport(&ipt.collectorsLogging, ipt.datakitPostLoggingURL)

		<-tick.C
	}
}

func collectAndReport(collectors *[]ccommon.DBMetricsCollector, reportURL string) {
	ba := ccommon.NewByteArray()

	for idx := range *collectors {
		pts, err := (*collectors)[idx].Collect()
		if err != nil {
			ccommon.ReportErrorf(inputName, l, "Collect failed: %v", err)
		} else {
			if len(pts) == 0 {
				continue
			}

			for _, pt := range pts {
				line := pt.LineProto()
				l.Debugf("line = %s", line)
				if len(line) > 0 {
					ba.Add(line)
				}
			}
		}
	}

	if ba.Len() > 0 {
		if err := ccommon.WriteData(l, bytes.Join(ba.Get(), []byte("\n")), reportURL); err != nil {
			l.Errorf("WriteData failed: %v", err)
		}
	}
}

func selectWrapper[T any](ipt *Input, s T, sql string) error {
	now := time.Now()

	err := ipt.db.Select(s, sql)
	if err != nil && (strings.Contains(err.Error(), "ORA-01012") || strings.Contains(err.Error(), "database is closed")) {
		if err := ipt.ConnectDB(); err != nil {
			ipt.CloseDB()
			return err
		}
	}

	l.Debugf("(selectWrapper) executed sql: %s, cost: %v\n", sql, time.Since(now))
	return err
}

func getWrapper[T any](ipt *Input, s T, sql string, binds ...interface{}) error {
	now := time.Now()

	err := ipt.db.Get(s, sql, binds...)
	if err != nil && (strings.Contains(err.Error(), "ORA-01012") || strings.Contains(err.Error(), "database is closed")) {
		if err := ipt.ConnectDB(); err != nil {
			ipt.CloseDB()
			return err
		}
	}

	l.Debugf("(getWrapper) executed sql: %s, cost: %v\n", sql, time.Since(now))
	return err
}

type collectParameters struct {
	Ipt        *Input
	MetricName string
}

////////////////////////////////////////////////////////////////////////////////

// collectOption used to add various options to collect.
type collectOption func(x *collectParameters)

func withInput(ipt *Input) collectOption {
	return func(x *collectParameters) {
		x.Ipt = ipt
	}
}

func withMetricName(metricName string) collectOption {
	return func(x *collectParameters) {
		x.MetricName = metricName
	}
}
