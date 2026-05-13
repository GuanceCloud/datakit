// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package doris collects Doris object data and documents Doris Prometheus metrics.
//
//nolint:lll
package doris

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/go-sql-driver/mysql"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	inputName      = "doris"
	catalogName    = "db"
	dorisType      = "Doris"
	defaultSQLPort = 9030
)

var (
	l                   = logger.DefaultSLogger(inputName)
	objectFeedName      = dkio.FeedSource(inputName, "O")
	customQueryFeedName = dkio.FeedSource(inputName, "custom_query")
)

type dorisObject struct {
	Enable   bool             `toml:"enabled"`
	Interval datakit.Duration `toml:"interval"`

	lastCollectionTime time.Time
}

type TLS struct {
	TLSKey             string `toml:"tls_key"`
	TLSCert            string `toml:"tls_cert"`
	TLSCA              string `toml:"tls_ca"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`
	AllowTLS10         bool   `toml:"allow_tls10,omitempty"`
}

type Input struct {
	Host           string            `toml:"host"`
	Port           int               `toml:"port"`
	User           string            `toml:"user"`
	Password       string            `toml:"password"`
	Interval       datakit.Duration  `toml:"interval"`
	ConnectTimeout datakit.Duration  `toml:"connect_timeout"`
	FEURL          string            `toml:"fe_metric_url"`
	Election       bool              `toml:"election"`
	Object         dorisObject       `toml:"object"`
	Query          []*customQuery    `toml:"custom_queries"`
	TLS            *TLS              `toml:"tls"`
	MetricTLS      *TLS              `toml:"metric_tls"`
	Tags           map[string]string `toml:"tags"`

	UpState int

	databaseInstance string
	version          string

	db         *sql.DB
	pause      atomic.Bool
	semStop    *cliutils.Sem
	feeder     dkio.Feeder
	tagger     datakit.GlobalTagger
	mergedTags map[string]string
	ptsTime    time.Time
}

var (
	_ inputs.InputV2       = (*Input)(nil)
	_ inputs.ElectionInput = (*Input)(nil)
)

func (ipt *Input) Run() {
	ipt.setup()

	tick := time.NewTicker(ipt.Interval.Duration)
	defer tick.Stop()

	l.Infof("%s input started", inputName)

	for {
		if err := ipt.initDB(); err != nil {
			l.Errorf("init doris db failed: %s", err.Error())
		} else {
			break
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.close()
			l.Infof("%s input exit", inputName)
			return

		case <-ipt.semStop.Wait():
			ipt.close()
			l.Infof("%s input return", inputName)
			return

		case <-tick.C:
		}
	}

	ipt.runCustomQueries()

	ipt.ptsTime = ntp.Now()
	for {
		if !ipt.pause.Load() {
			if err := ipt.collect(); err != nil {
				ipt.setErrUpState()
				l.Errorf("collect doris object failed: %s", err.Error())
			} else {
				ipt.setUpState()
			}
			ipt.feedUpMetric()
		} else {
			l.Debugf("not leader, skipped")
		}

		select {
		case <-datakit.Exit.Wait():
			ipt.close()
			l.Infof("%s input exit", inputName)
			return

		case <-ipt.semStop.Wait():
			ipt.close()
			l.Infof("%s input return", inputName)
			return

		case tt := <-tick.C:
			ipt.ptsTime = inputs.AlignTime(tt, ipt.ptsTime, ipt.Interval.Duration)
		}
	}
}

func (*Input) Catalog() string { return catalogName }

func (*Input) SampleConfig() string { return sampleCfg }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&feMeasurement{},
		&beMeasurement{},
		&commonMeasurement{},
		&jvmMeasurement{},
		&dorisObjectMeasurement{},
		&inputs.UpMeasurement{},
	}
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) Pause() error {
	ipt.pause.Store(true)
	return nil
}

func (ipt *Input) Resume() error {
	ipt.pause.Store(false)
	return nil
}

func (ipt *Input) setup() {
	l = logger.SLogger(inputName)

	if ipt.Host == "" {
		ipt.Host = "127.0.0.1"
	}
	if ipt.Port == 0 {
		ipt.Port = defaultSQLPort
	}
	if ipt.Interval.Duration <= 0 {
		ipt.Interval.Duration = 10 * time.Second
	}

	if ipt.ConnectTimeout.Duration <= 0 {
		ipt.ConnectTimeout.Duration = 10 * time.Second
	}
	if ipt.Object.Interval.Duration <= 0 {
		ipt.Object.Interval.Duration = 10 * time.Minute
	}

	if ipt.Election {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.ElectionTags(), ipt.Tags, ipt.Host)
	} else {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, ipt.Host)
	}
	if _, ok := ipt.mergedTags["server"]; !ok {
		ipt.mergedTags["server"] = ipt.server()
	}

	if instanceTag, ok := ipt.mergedTags["database_instance"]; ok && instanceTag != "" {
		ipt.databaseInstance = instanceTag
	}
}

func (ipt *Input) collect() error {
	if !ipt.Object.Enable {
		return nil
	}

	if !ipt.Object.lastCollectionTime.IsZero() &&
		ipt.Object.lastCollectionTime.Add(ipt.Object.Interval.Duration).After(time.Now()) {
		l.Debugf("skip doris object collection, time interval not reached")
		return nil
	}

	if err := ipt.collectDatabaseObject(); err != nil {
		return err
	}
	ipt.Object.lastCollectionTime = time.Now()

	return nil
}

func (ipt *Input) initDB() error {
	dsn, err := ipt.getDSNString()
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("open doris connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ipt.ConnectTimeout.Duration)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("ping doris %s failed: %w", ipt.server(), err)
	}

	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)

	ipt.db = db
	if err := ipt.initDatabaseInfo(ctx); err != nil {
		_ = db.Close()
		ipt.db = nil
		return err
	}

	return nil
}

func (ipt *Input) getDSNString() (string, error) {
	cfg := mysql.Config{
		User:                 ipt.User,
		Passwd:               ipt.Password,
		Net:                  "tcp",
		Addr:                 ipt.server(),
		AllowNativePasswords: true,
		CheckConnLiveness:    true,
		Timeout:              ipt.ConnectTimeout.Duration,
		Params:               map[string]string{},
	}

	if ipt.TLS != nil {
		tlsConfig, err := createTLSConf(ipt.TLS.TLSCA, ipt.TLS.TLSCert, ipt.TLS.TLSKey)
		if err != nil {
			return "", err
		}

		tlsConfig.InsecureSkipVerify = ipt.TLS.InsecureSkipVerify
		if ipt.TLS.AllowTLS10 {
			tlsConfig.MinVersion = tls.VersionTLS10
		}

		tlsConfigName := fmt.Sprintf("%s-%s", inputName, ipt.server())
		if err := mysql.RegisterTLSConfig(tlsConfigName, tlsConfig); err != nil {
			return "", fmt.Errorf("register tls config failed: %w", err)
		}
		cfg.Params["tls"] = tlsConfigName
	}

	return cfg.FormatDSN(), nil
}

func createTLSConf(caFile, certFile, keyFile string) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if certFile != "" || keyFile != "" {
		if certFile == "" || keyFile == "" {
			return nil, errors.New("both tls_cert and tls_key are required")
		}
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if caFile == "" {
		return tlsConfig, nil
	}

	caCert, err := os.ReadFile(filepath.Clean(caFile))
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("failed to append certs from PEM")
	}

	tlsConfig.RootCAs = caCertPool

	return tlsConfig, nil
}

func (ipt *Input) close() {
	if ipt.db != nil {
		if err := ipt.db.Close(); err != nil {
			l.Warnf("close doris connection failed: %s", err.Error())
		}
		ipt.db = nil
	}
}

func (ipt *Input) server() string {
	return fmt.Sprintf("%s:%d", ipt.Host, ipt.Port)
}

func (ipt *Input) setUpState() {
	ipt.UpState = 1
}

func (ipt *Input) setErrUpState() {
	ipt.UpState = 0
}

func (ipt *Input) feedUpMetric() {
	kvs := point.NewTags(map[string]string{
		"job":      inputName,
		"instance": ipt.server(),
	})
	kvs = kvs.Add("up", ipt.UpState)

	opts := point.DefaultMetricOptions()
	opts = append(opts, point.WithTimestamp(ipt.ptsTime.UnixNano()))
	pt := point.NewPoint(inputs.CollectorUpMeasurement, kvs, opts...)

	if err := ipt.feeder.Feed(point.Metric, []*point.Point{pt},
		dkio.WithElection(ipt.Election),
		dkio.WithSource(inputName),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Metric),
		)
		l.Errorf("feed up metric: %s", err.Error())
	}
}

func defaultInput() *Input {
	return &Input{
		Interval:       datakit.Duration{Duration: 10 * time.Second},
		ConnectTimeout: datakit.Duration{Duration: 10 * time.Second},
		Election:       true,
		Tags:           make(map[string]string),
		semStop:        cliutils.NewSem(),
		feeder:         dkio.DefaultFeeder(),
		tagger:         datakit.DefaultGlobalTagger(),
		Object: dorisObject{
			Enable:   true,
			Interval: datakit.Duration{Duration: 10 * time.Minute},
		},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
