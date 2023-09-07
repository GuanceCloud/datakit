// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux && amd64

// Package cdb2 contains collect IBM Db2 code.
package cdb2

import (
	"bytes"
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	_ "github.com/ibmdb/go_ibm_db"
	"github.com/jmoiron/sqlx"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/db2/collect/ccommon"
)

// Monitor procedures and functions:
// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.sql.rtn.doc/doc/c0053963.html
//
// Monitor views:
// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.sql.rtn.doc/doc/c0061229.html
//
// Monitor element reference:
// https://www.ibm.com/support/knowledgecenter/SSEPGG_11.1.0/com.ibm.db2.luw.admin.mon.doc/doc/r0001140.html

const (
	inputName = "db2"

	metricNameInstance       = "db2_instance"
	metricNameDatabase       = "db2_database"
	metricNameBufferPool     = "db2_buffer_pool"
	metricNameTableSpace     = "db2_table_space"
	metricNameTransactionLog = "db2_transaction_log"
)

var (
	l                = logger.DefaultSLogger(inputName)
	_ ccommon.IInput = (*Input)(nil)
)

type Input struct {
	interval    string
	user        string
	password    string
	desc        string
	host        string
	port        string
	serviceName string
	database    string
	tags        map[string]string
	election    bool

	db                  *sqlx.DB
	intervalDuration    time.Duration
	collectors          []ccommon.DBMetricsCollector
	datakitPostURL      string
	datakitPostEventURL string
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
		database:    opt.Database,
		tags:        make(map[string]string),
		election:    opt.Election,
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

	if len(ipt.serviceName) == 0 {
		ipt.serviceName = inputName
	}
	ipt.tags[inputName+"_service"] = ipt.serviceName
	ipt.tags[inputName+"_server"] = net.JoinHostPort(ipt.host, ipt.port)

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

	instanceMetric := newInstanceMetrics(withInput(ipt), withMetricName(metricNameInstance))
	databaseMetric := newDatabaseMetrics(withInput(ipt), withMetricName(metricNameDatabase))
	bufferPoolMetric := newBufferPoolMetrics(withInput(ipt), withMetricName(metricNameBufferPool))
	tableSpaceMetric := newTableSpaceMetrics(withInput(ipt), withMetricName(metricNameTableSpace))
	transactionLogMetric := newTransactionLogMetrics(withInput(ipt), withMetricName(metricNameTransactionLog))

	ipt.collectors = append(ipt.collectors,
		instanceMetric,
		databaseMetric,
		bufferPoolMetric,
		tableSpaceMetric,
		transactionLogMetric,
	)
	l.Infof("collectors len = %d", len(ipt.collectors))

	ipt.datakitPostURL = ccommon.GetPostURL(
		opt.Election,
		ccommon.CategoryMetric, inputName, opt.DatakitHTTPHost,
		opt.DatakitHTTPPort,
	)
	l.Infof("Datakit post metric URL: %s", ipt.datakitPostURL)

	ipt.datakitPostEventURL = ccommon.GetPostURL(
		opt.Election,
		ccommon.CategoryEvent, inputName, opt.DatakitHTTPHost,
		opt.DatakitHTTPPort,
	)
	l.Infof("Datakit post event URL: %s", ipt.datakitPostEventURL)

	return ipt
}

// ConnectDB establishes a connection to an instance and returns an open connection to the database.
func (ipt *Input) ConnectDB() error {
	con := fmt.Sprintf("HOSTNAME=%s;DATABASE=%s;PORT=%s;UID=%s;PWD=%s",
		ipt.host, ipt.database, ipt.port, ipt.user, ipt.password)
	db, err := sqlx.Open("go_ibm_db", con)
	if err != nil {
		l.Errorf("sqlx.Open failed: %v", err)
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

// Run start collecting.
func (ipt *Input) Run() {
	l.Info("starting " + inputName + "...")

	tick := time.NewTicker(ipt.intervalDuration)
	defer ipt.db.Close() //nolint:errcheck
	defer tick.Stop()

	for {
		ba := ccommon.NewByteArray()

		for idx := range ipt.collectors {
			pt, err := ipt.collectors[idx].Collect()
			if err != nil {
				ccommon.ReportErrorf(inputName, l, "Collect failed: %v", err)
			} else {
				line := pt.LineProto()
				ba.Add(line)
			}
		}

		if ba.Len() > 0 {
			if err := ccommon.WriteData(l, bytes.Join(ba.Get(), []byte("\n")), ipt.datakitPostURL); err != nil {
				l.Errorf("writeData failed: %v", err)
			}
		}

		<-tick.C
	}
}

////////////////////////////////////////////////////////////////////////////////

func selectWrapper[T any](ipt *Input, s T, sql string) error {
	now := time.Now()

	err := ipt.db.Select(s, sql)
	if err != nil {
		fmt.Println("err =", err.Error(), "sql =", sql)

		if strings.Contains(err.Error(), "database is closed") {
			if err := ipt.ConnectDB(); err != nil {
				ipt.CloseDB()
				return err
			}
		}

		return err
	}

	l.Debugf("executed sql: %s, cost: %v\n", sql, time.Since(now))
	return nil
}

////////////////////////////////////////////////////////////////////////////////

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
