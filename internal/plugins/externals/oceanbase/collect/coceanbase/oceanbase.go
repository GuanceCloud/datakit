// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

// Package coceanbase contains collect OceanBase code.
package coceanbase

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/oceanbase/collect/ccommon"
)

const (
	inputName  = "oceanbase"
	metricName = "oceanbase"
	logName    = "oceanbase_log"
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
	Tenant        string
	Cluster       string
	ConnectString string
	Database      string
	Mode          string
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
		interval:      opt.Interval,
		user:          opt.Username,
		password:      opt.Password,
		desc:          opt.InstanceDesc,
		host:          opt.Host,
		port:          opt.Port,
		serviceName:   opt.ServiceName,
		tags:          make(map[string]string),
		election:      opt.Election,
		Tenant:        opt.Tenant,
		Cluster:       opt.Cluster,
		ConnectString: opt.ConnectString,
		Database:      opt.Database,
		Mode:          opt.Mode,
	}

	switch ipt.Mode {
	case "mysql":
		tenantMode = modeMySQL
	case "oracle":
		tenantMode = modeOracle
	default:
		l.Errorf("Unknown mode: %s", ipt.Mode)
	}

	// ENV_INPUT_OCEANBASE_PASSWORD
	pwd := os.Getenv("ENV_INPUT_OCEANBASE_PASSWORD")
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
				l.Warnf("slow query time %v less than 1 millisecond, skip", du)
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

	proMetric := newOBMetrics(withInput(ipt), withMetricName(metricName))
	customMetric := newCustomQueryCollector(withInput(ipt))

	ipt.collectors = append(ipt.collectors, proMetric, customMetric)

	ipt.datakitPostURL = ccommon.GetPostURL(
		opt.Election,
		ccommon.CategoryMetric, inputName, opt.DatakitHTTPHost,
		opt.DatakitHTTPPort,
	)
	l.Infof("Datakit post metric URL: %s", ipt.datakitPostURL)

	if ipt.SlowQueryTime > 0 {
		slowQueryLoggingR := newSlowQueryLogging(withInput(ipt), withMetricName(logName))
		ipt.collectorsLogging = append(ipt.collectorsLogging, slowQueryLoggingR)

		ipt.datakitPostLoggingURL = ccommon.GetPostURL(
			opt.Election,
			ccommon.CategoryLogging, inputName, opt.DatakitHTTPHost,
			opt.DatakitHTTPPort,
		)
	}

	l.Infof("collectors len = %d", len(ipt.collectors))

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

// ConnectDB establishes a connection to an Database instance and returns an open connection to the database.
func (ipt *Input) ConnectDB() error {
	// -usys@oraclet#obcluster
	// -u用户名@租户名#集群名

	// dataSourceName := fmt.Sprintf("%s/%s@%s:%s/%s",
	// 	ipt.user, ipt.password, ipt.host, ipt.port, ipt.serviceName)
	// fmt.Println(dataSourceName)

	// mysql: root:passwd@tcp(0.0.0.0:3306)/user
	// sqlconn := "datakit@oraclet#obcluster/123456@10.200.14.240:2883"

	sqlconn := ipt.ConnectString
	if len(sqlconn) == 0 {
		user := ipt.user

		if len(ipt.Tenant) > 0 {
			user += ("@" + ipt.Tenant)
		}

		if len(ipt.Cluster) > 0 {
			user += ("#" + ipt.Cluster)
		}

		var safeOutput string

		//nolint:exhaustive
		switch tenantMode {
		case modeMySQL:
			password := ipt.password
			cfg := mysql.Config{
				User:                 user,
				Passwd:               password,
				Net:                  "tcp",
				Addr:                 net.JoinHostPort(ipt.host, ipt.port),
				DBName:               ipt.Database,
				AllowNativePasswords: true,
			}
			sqlconn = cfg.FormatDSN()

			cfg.Passwd = "***"
			safeOutput = cfg.FormatDSN()

		case modeOracle:
			password := url.QueryEscape(ipt.password)
			sqlconn = user
			safeOutput = user

			sqlconn += fmt.Sprintf("/%s@%s:%s", password, ipt.host, ipt.port)
			safeOutput += fmt.Sprintf("/%s@%s:%s", "***", ipt.host, ipt.port) // Used for output.
		}

		l.Infof("sqlconn = %s", safeOutput)
	}

	var db *sqlx.DB
	var err error

	//nolint:exhaustive
	switch tenantMode {
	case modeMySQL:
		db, err = sqlx.Open("mysql", sqlconn)
	case modeOracle:
		db, err = sqlx.Open("oci8", sqlconn)
	}

	if err != nil {
		l.Errorf("sqlx.Open failed: %v\n", err)
		return err
	}
	if err := db.Ping(); err != nil {
		l.Errorf("Ping failed: %v\n", err)
		return err
	}

	ipt.db = db
	return nil
}

// CloseDB cleans up database resources used.
func (ipt *Input) CloseDB() {
	if ipt.db != nil {
		if err := ipt.db.Close(); err != nil {
			l.Warnf("failed to close "+inputName+" connection | server=[%s]: %s", ipt.host, err.Error())
		}
	}
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

func closeRows(r rows) {
	if err := r.Close(); err != nil {
		l.Warnf("Close: %s, ignored", err)
	}
}

type rows interface {
	Next() bool
	Scan(...interface{}) error
	Close() error
	Err() error
	Columns() ([]string, error)
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
				ba.Add(line)
			}
		}
	}

	if ba.Len() > 0 {
		if err := ccommon.WriteData(l, bytes.Join(ba.Get(), []byte("\n")), reportURL); err != nil {
			l.Errorf("writeData failed: %v", err)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func selectMapWrapper(ipt *Input, sqlText string) ([]map[string]interface{}, error) {
	now := time.Now()
	mRet, err := selectMap(ipt, sqlText)
	l.Debugf("executed sql: %s, cost: %v, length: %d, err: %v\n", sqlText, time.Since(now), len(mRet), err)
	return mRet, err
}

func selectMap(ipt *Input, sqlText string) ([]map[string]interface{}, error) {
	var err error
	var rows *sql.Rows
	var cols []string

	defer func() {
		if err != nil {
			l.Errorf("SQL failed, err = %v, sql = %s", err, sqlText)
		}
	}()

	rows, err = ipt.db.Query(sqlText)
	if err != nil {
		l.Errorf("db.Query() failed: %v", err)
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	cols, err = rows.Columns()
	if err != nil {
		l.Errorf("rows.Columns() failed: %v", err)
		return nil, err
	}

	mRet := make([]map[string]interface{}, 0)

	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		if err = rows.Scan(columnPointers...); err != nil {
			l.Errorf("Scan() failed: %v", err)
			return nil, err
		}

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}

		// Outputs: map[columnName:value columnName2:value2 columnName3:value3 ...]
		mRet = append(mRet, m)
	}

	if err = rows.Err(); err != nil {
		l.Errorf("rows.Err() failed: %v", err)
		return nil, err
	}

	return mRet, nil
}

func selectWrapper[T any](ipt *Input, s T, sql string) error {
	now := time.Now()

	err := ipt.db.Select(s, sql)
	if err != nil && (strings.Contains(err.Error(), "ORA-01012") || strings.Contains(err.Error(), "database is closed")) {
		if err := ipt.ConnectDB(); err != nil {
			ipt.CloseDB()
		}
	}

	if err != nil {
		l.Errorf("executed sql: %s, cost: %v, err: %v\n", sql, time.Since(now), err)
	} else {
		l.Debugf("executed sql: %s, cost: %v, err: %v\n", sql, time.Since(now), err)
	}

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
