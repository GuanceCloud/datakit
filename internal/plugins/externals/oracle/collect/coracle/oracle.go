// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package coracle contains collect Oracle code.
package coracle

import (
	"bytes"
	"fmt"
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

	pdbName        = "pdb_name"
	tablespaceName = "tablespace_name"
	programName    = "program"
)

var (
	l   = logger.DefaultSLogger(inputName)
	dic = map[string]string{
		"buffer_cache_hit_ratio":       "buffer_cachehit_ratio",
		"cursor_cache_hit_ratio":       "cursor_cachehit_ratio",
		"library_cache_hit_ratio":      "library_cachehit_ratio",
		"shared_pool_free_%":           "shared_pool_free",
		"physical_read_bytes_per_sec":  "physical_reads",
		"physical_write_bytes_per_sec": "physical_writes",
		"enqueue_timeouts_per_sec":     "enqueue_timeouts",

		"gc_cr_block_received_per_second": "gc_cr_block_received",
		"global_cache_blocks_corrupted":   "cache_blocks_corrupt",
		"global_cache_blocks_lost":        "cache_blocks_lost",
		"average_active_sessions":         "active_sessions",
		"sql_service_response_time":       "service_response_time",
		"user_rollbacks_per_sec":          "user_rollbacks",
		"total_sorts_per_user_call":       "sorts_per_user_call",
		"rows_per_sort":                   "rows_per_sort",
		"disk_sort_per_sec":               "disk_sorts",
		"memory_sorts_ratio":              "memory_sorts_ratio",
		"database_wait_time_ratio":        "database_wait_time_ratio",
		"session_limit_%":                 "session_limit_usage",
		"session_count":                   "session_count",
		"temp_space_used":                 "temp_space_used",
	}

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
	tags        map[string]string
	election    bool

	db               *sqlx.DB
	intervalDuration time.Duration
	collectors       []ccommon.DBMetricsCollector
	datakitPostURL   string
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

	ipt.collectors = append(ipt.collectors, proMetric, tsMetric, sysMetric)
	l.Infof("collectors len = %d", len(ipt.collectors))

	ipt.datakitPostURL = ccommon.GetPostURL(
		opt.Election,
		ccommon.CategoryMetric, inputName, opt.DatakitHTTPHost,
		opt.DatakitHTTPPort,
	)
	l.Infof("Datakit post metric URL: %s", ipt.datakitPostURL)

	return ipt
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
	if err != nil && (strings.Contains(err.Error(), "ORA-01012") || strings.Contains(err.Error(), "database is closed")) {
		if err := ipt.ConnectDB(); err != nil {
			ipt.CloseDB()
			return err
		}
	}

	l.Debugf("executed sql: %s, cost: %v\n", sql, time.Since(now))
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
