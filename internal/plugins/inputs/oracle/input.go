// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package oracle collect Oracle metrics
package oracle

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/jmoiron/sqlx"
	go_ora "github.com/sijms/go-ora/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var _ inputs.ElectionInput = (*Input)(nil)

const (
	maxInterval = 15 * time.Minute
	minInterval = 10 * time.Second
	inputName   = "oracle"
	catalogName = "db"
)

var l = logger.DefaultSLogger(inputName)

type customQuery struct {
	SQL    string   `toml:"sql"`
	Metric string   `toml:"metric"`
	Tags   []string `toml:"tags"`
	Fields []string `toml:"fields"`
}

type Input struct {
	Host            string `toml:"host"`
	Port            int    `toml:"port"`
	User            string `toml:"user"`
	Password        string `toml:"password"`
	Interval        datakit.Duration
	Timeout         string `toml:"connect_timeout"`
	Service         string `toml:"service"`
	timeoutDuration time.Duration
	Query           []*customQuery    `toml:"custom_queries"`
	SlowQueryTime   string            `toml:"slow_query_time"`
	Election        bool              `toml:"election"`
	Tags            map[string]string `toml:"tags"`

	semStop        *cliutils.Sem // start stop signal
	pauseCh        chan bool
	feeder         dkio.Feeder
	tagger         datakit.GlobalTagger
	mergedTags     map[string]string
	db             *sqlx.DB
	pause          bool
	start          time.Time
	slowQueryTime  time.Duration
	lastActiveTime string
	collectors     map[string]func() (point.Category, []*point.Point, error)
	cacheSQL       map[string]string
}

func (ipt *Input) setupDB() error {
	var err error
	ipt.timeoutDuration, err = time.ParseDuration(ipt.Timeout)
	if err != nil {
		ipt.timeoutDuration = 30 * time.Second
	}

	connStr := ipt.getConnString()
	db, err := sqlx.Open("oracle", connStr)
	if err != nil {
		l.Errorf("sql.Open(): %s", err.Error())
		return err
	} else {
		ipt.db = db
	}

	db.SetConnMaxLifetime(ipt.Interval.Duration) // avoid max cursor problem

	ctx, cancel := context.WithTimeout(context.Background(), ipt.timeoutDuration)
	defer cancel()

	if err := ipt.db.PingContext(ctx); err != nil {
		l.Errorf("init config connect error %v", err)
		ipt.db.Close() //nolint:errcheck,gosec
		return err
	}

	return nil
}

func (ipt *Input) getConnString() string {
	opt := map[string]string{
		"timeout": fmt.Sprintf("%d", ipt.timeoutDuration/time.Second),
	}

	connStr := go_ora.BuildUrl(ipt.Host, ipt.Port, ipt.Service, ipt.User, ipt.Password, opt)

	return connStr
}

// init db connect.
func (ipt *Input) initDBConnect() error {
	isNeedConnect := false

	if ipt.db == nil {
		isNeedConnect = true
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), ipt.timeoutDuration)
		defer func() {
			cancel()
		}()

		if err := ipt.db.PingContext(ctx); err != nil {
			isNeedConnect = true
		}
	}

	if isNeedConnect {
		if err := ipt.setupDB(); err != nil {
			return err
		}
	}

	return nil
}

func (ipt *Input) Collect() (map[point.Category][]*point.Point, error) {
	var err error
	allPts := make(map[point.Category][]*point.Point)

	ipt.start = time.Now()

	if err := ipt.initDBConnect(); err != nil {
		return nil, err
	}

	for name, collector := range ipt.collectors {
		category, pts, err := collector()
		if err != nil {
			l.Warnf("collect %s failed: %s", name, err.Error())
			continue
		}

		allPts[category] = append(allPts[category], pts...)
	}

	return allPts, err
}

func (ipt *Input) Init() {
	var err error

	l = logger.SLogger(inputName)
	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	setHost := false
	host := strings.ToLower(ipt.Host)
	switch host {
	case "", "localhost":
		setHost = true
	default:
		if net.ParseIP(host).IsLoopback() {
			setHost = true
		}
	}
	if setHost {
		host, err = os.Hostname()
		if err != nil {
			l.Errorf("os.Hostname failed: %v", err)
		}
	}

	ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, host)
	ipt.mergedTags["oracle_service"] = ipt.Service
	ipt.mergedTags["oracle_server"] = fmt.Sprintf("%s:%d", ipt.Host, ipt.Port)

	ipt.collectors = map[string]func() (point.Category, []*point.Point, error){
		"oracle_process":    ipt.collectOracleProcess,
		"oracle_tablespace": ipt.collectOracleTableSpace,
		"oracle_system":     ipt.collectOracleSystem,
	}
	// cache sql
	ipt.cacheSQL = make(map[string]string)
	// slow query
	if len(ipt.SlowQueryTime) > 0 {
		du, err := time.ParseDuration(ipt.SlowQueryTime)
		if err != nil {
			l.Warnf("bad slow query %s: %s, disable slow query", ipt.SlowQueryTime, err.Error())
		} else {
			if du >= time.Millisecond {
				ipt.slowQueryTime = du
				ipt.collectors["slow_query"] = ipt.collectSlowQuery
			} else {
				l.Warnf("slow query time %v less than 1 millisecond, skip", du)
			}
		}
	}

	if len(ipt.Query) > 0 {
		ipt.collectors["custom_query"] = ipt.collectCustomQuery
	}

	// Try until init OK.
	for {
		if err := ipt.setupDB(); err != nil {
			l.Warnf("init config error: %s", err.Error())
			ipt.feeder.FeedLastError(err.Error(),
				dkio.WithLastErrorInput(inputName),
				dkio.WithLastErrorCategory(point.Metric),
			)
		} else {
			break
		}

		select {
		case <-datakit.Exit.Wait():
			return

		case <-ipt.semStop.Wait():
			return

		case <-tick.C:
		}
	}
}

func (ipt *Input) Run() {
	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()
	defer func() {
		l.Info("oracle exit")
	}()

	ipt.Init()

	l.Infof("collecting each %v", ipt.Interval.Duration)

	for {
		if ipt.pause {
			l.Info("not leader, skipped")
		} else {
			l.Info("oracle input gathering...")

			mpts, err := ipt.Collect()
			if err != nil {
				l.Warnf("i.Collect failed: %v", err)
				ipt.feeder.FeedLastError(err.Error(),
					dkio.WithLastErrorInput(inputName),
					dkio.WithLastErrorCategory(point.Metric),
				)
			}

			for category, pts := range mpts {
				if len(pts) > 0 {
					if err := ipt.feeder.FeedV2(category, pts,
						dkio.WithCollectCost(time.Since(ipt.start)),
						dkio.WithElection(ipt.Election),
						dkio.WithInputName(inputName),
					); err != nil {
						ipt.feeder.FeedLastError(err.Error(),
							dkio.WithLastErrorInput(inputName),
							dkio.WithLastErrorCategory(point.Metric),
						)
						l.Errorf("feed : %s", err)
					}
				}
			}
		}

		select {
		case <-datakit.Exit.Wait():
			return

		case <-ipt.semStop.Wait():
			return

		case <-tick.C:

		case ipt.pause = <-ipt.pauseCh:
			// nil
		}
	}
}

func (ipt *Input) Catalog() string { return catalogName }

func (ipt *Input) SampleConfig() string { return configSample }

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&processMeasurement{},
		&tablespaceMeasurement{},
		&systemMeasurement{},
	}
}

func (ipt *Input) AvailableArchs() []string {
	return datakit.AllOSWithElection
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

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func defaultInput() *Input {
	return &Input{
		Tags:     make(map[string]string),
		Timeout:  "10s",
		pauseCh:  make(chan bool, inputs.ElectionPauseChannelLength),
		Election: true,
		feeder:   dkio.DefaultFeeder(),
		tagger:   datakit.DefaultGlobalTagger(),
		semStop:  cliutils.NewSem(),
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
