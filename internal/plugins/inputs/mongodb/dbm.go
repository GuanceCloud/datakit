// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"
	goVersion "github.com/hashicorp/go-version"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/util"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	defaultQueryMetricsInterval          = 60 * time.Second
	defaultQueryMetricsLimit             = 10000
	defaultActivityInterval              = 10 * time.Second
	defaultSlowOperationsInterval        = 10 * time.Second
	defaultSlowOperationsLimit           = 1000
	defaultSlowDatabaseDiscoveryInterval = 10 * time.Minute
	defaultSlowMaxDiscoveredDatabases    = 100
	minimumDBMInterval                   = 5 * time.Second
	maximumDBMInterval                   = 10 * time.Minute
	defaultQueryTextMaxBytes             = 512
	maximumQueryTextMaxBytes             = 1024
	queryObjectCacheSize                 = 10000
	queryObjectCacheTTL                  = 24 * time.Hour
)

var minimumQueryMetricsMongoDBVersion = goVersion.Must(goVersion.NewVersion("8.0.0"))

type queryMetricsConfig struct {
	Enabled           bool             `toml:"enabled"`
	Interval          datakit.Duration `toml:"interval"`
	Limit             int64            `toml:"limit"`
	QueryTextMaxBytes int              `toml:"query_text_max_bytes"`
}

type activityConfig struct {
	Enabled  bool             `toml:"enabled"`
	Interval datakit.Duration `toml:"interval"`
}

type slowOperationsConfig struct {
	Enabled       bool             `toml:"enabled"`
	Interval      datakit.Duration `toml:"interval"`
	MaxOperations int64            `toml:"max_operations"`
}

type dbmConfig struct {
	Enabled        bool                  `toml:"enabled"`
	Databases      []string              `toml:"databases"`
	QueryMetrics   *queryMetricsConfig   `toml:"query_metrics"`
	Activity       *activityConfig       `toml:"activity"`
	SlowOperations *slowOperationsConfig `toml:"slow_operations"`
}

func newMongoDBCommandObfuscator() *obfuscate.Obfuscator {
	return obfuscate.NewObfuscator(obfuscate.Config{
		Mongo: obfuscate.JSONConfig{
			Enabled: true,
			KeepValues: []string{
				"find", "sort", "projection", "skip", "batchSize", "$db", "getMore", "collection",
				"delete", "findAndModify", "insert", "ordered", "update", "aggregate",
			},
		},
	})
}

func obfuscateMongoDBCommand(commandObfuscator *obfuscate.Obfuscator, command bson.D) (string, error) {
	commandJSON, err := bson.MarshalExtJSON(command, false, false)
	if err != nil {
		return "", err
	}
	return commandObfuscator.ObfuscateMongoDBString(string(commandJSON)), nil
}

func computeMongoDBNormalizedQueryHash(obfuscatedCommand string) (string, error) {
	decoder := json.NewDecoder(strings.NewReader(obfuscatedCommand))
	decoder.UseNumber()

	var command interface{}
	if err := decoder.Decode(&command); err != nil {
		return "", fmt.Errorf("decode obfuscated MongoDB command: %w", err)
	}
	canonicalCommand, err := json.Marshal(command)
	if err != nil {
		return "", fmt.Errorf("encode canonical MongoDB command: %w", err)
	}
	return util.ComputeNormalizedSQLHash(string(canonicalCommand)), nil
}

type dbmNodeStatus struct {
	SetName           string `bson:"setName"`
	ArbiterOnly       bool   `bson:"arbiterOnly"`
	IsWritablePrimary bool   `bson:"isWritablePrimary"`
	Secondary         bool   `bson:"secondary"`
}

func (status dbmNodeStatus) collectible() bool {
	if status.ArbiterOnly {
		return false
	}
	return status.SetName == "" || status.IsWritablePrimary || status.Secondary
}

func (ipt *Input) queryTextMaxBytes() int {
	maxBytes := defaultQueryTextMaxBytes
	if ipt.Dbm != nil && ipt.Dbm.QueryMetrics != nil {
		maxBytes = ipt.Dbm.QueryMetrics.QueryTextMaxBytes
	}
	if maxBytes <= 0 || maxBytes > maximumQueryTextMaxBytes {
		return defaultQueryTextMaxBytes
	}
	return maxBytes
}

func (ipt *Input) queryMetricsLimit() int64 {
	if ipt.Dbm != nil && ipt.Dbm.QueryMetrics != nil && ipt.Dbm.QueryMetrics.Limit > 0 {
		return ipt.Dbm.QueryMetrics.Limit
	}
	return defaultQueryMetricsLimit
}

func (ipt *Input) runDBMCollectors() {
	if ipt.Dbm == nil || !ipt.Dbm.Enabled {
		return
	}
	queryMetricsEnabled := ipt.Dbm.QueryMetrics != nil && ipt.Dbm.QueryMetrics.Enabled
	activityEnabled := ipt.Dbm.Activity != nil && ipt.Dbm.Activity.Enabled
	slowOperationsEnabled := ipt.Dbm.SlowOperations != nil && ipt.Dbm.SlowOperations.Enabled
	if !queryMetricsEnabled && !activityEnabled && !slowOperationsEnabled {
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "mongodb_dbm"})
	if queryMetricsEnabled {
		g.Go(func(ctx context.Context) error {
			ipt.runQueryMetrics(ctx)
			return nil
		})
	}
	if activityEnabled {
		g.Go(func(ctx context.Context) error {
			ipt.runActivity(ctx)
			return nil
		})
	}
	if slowOperationsEnabled {
		g.Go(func(ctx context.Context) error {
			ipt.runSlowOperations(ctx)
			return nil
		})
	}
}

func (ipt *Input) runQueryMetrics(ctx context.Context) {
	if len(ipt.mgoSvrs) == 0 {
		return
	}
	svr := ipt.mgoSvrs[0]
	duration := ipt.queryMetricsInterval()
	tick := time.NewTicker(duration)
	defer tick.Stop()

	ptsTime := ntp.Now()
	versionReady := false
	for {
		if ipt.pause.Load() {
			log.Debugf("not leader, MongoDB query metrics collection skipped")
		} else {
			collectCtx, cancel := context.WithTimeout(ctx, duration)
			if !versionReady {
				version, err := mongoDBVersion(collectCtx, svr)
				if err != nil {
					log.Warnf("get MongoDB version for query metrics on %s failed: %s; retrying next collection", svr.host, err)
				} else {
					supported, err := queryMetricsVersionSupported(version)
					if err != nil {
						log.Warnf("parse MongoDB version %q for query metrics on %s failed: %s; stopping query metrics collection", version, svr.host, err)
						cancel()
						return
					}
					if !supported {
						log.Infof("MongoDB query metrics requires MongoDB 8.0 or later; server %s is running %s", svr.host, version)
						cancel()
						return
					}
					versionReady = true
				}
			}
			if versionReady {
				ipt.collectQueryMetrics(collectCtx, ptsTime)
			}
			cancel()
		}

		select {
		case <-ctx.Done():
			log.Info("MongoDB query metrics collection exit")
			return
		case <-datakit.Exit.Wait():
			log.Info("MongoDB query metrics collection exit")
			return
		case <-ipt.semStop.Wait():
			log.Info("MongoDB query metrics collection return")
			return
		case tt := <-tick.C:
			ptsTime = inputs.AlignTime(tt, ptsTime, duration)
		}
	}
}

func mongoDBVersion(ctx context.Context, svr *MongodbServer) (string, error) {
	var status struct {
		Version string `bson:"version"`
	}
	if err := svr.cli.Database("admin").RunCommand(ctx, bson.D{{Key: "serverStatus", Value: 1}}).Decode(&status); err != nil {
		return "", err
	}
	version := strings.TrimSpace(status.Version)
	if version == "" {
		return "", fmt.Errorf("serverStatus returned an empty version")
	}
	return version, nil
}

func (svr *MongodbServer) canCollectDBM(ctx context.Context) bool {
	var status dbmNodeStatus
	if err := svr.cli.Database("admin").RunCommand(ctx, bson.D{{Key: "hello", Value: 1}}).Decode(&status); err != nil {
		log.Warnf("check MongoDB node state for DBM on %s failed: %s; skipping this collection", svr.host, err)
		return false
	}
	if !status.collectible() {
		log.Debugf("skip MongoDB DBM collection on unavailable replica set member %s", svr.host)
		return false
	}
	return true
}

func queryMetricsVersionSupported(rawVersion string) (bool, error) {
	version, err := goVersion.NewVersion(rawVersion)
	if err != nil {
		return false, err
	}
	return version.GreaterThanOrEqual(minimumQueryMetricsMongoDBVersion), nil
}

func (ipt *Input) queryMetricsInterval() time.Duration {
	duration := defaultQueryMetricsInterval
	if ipt.Dbm != nil && ipt.Dbm.QueryMetrics != nil && ipt.Dbm.QueryMetrics.Interval.Duration > 0 {
		duration = ipt.Dbm.QueryMetrics.Interval.Duration
	}

	return config.ProtectedInterval(
		minimumDBMInterval,
		maximumDBMInterval,
		duration,
	)
}

func (ipt *Input) activityInterval() time.Duration {
	duration := defaultActivityInterval
	if ipt.Dbm != nil && ipt.Dbm.Activity != nil && ipt.Dbm.Activity.Interval.Duration > 0 {
		duration = ipt.Dbm.Activity.Interval.Duration
	}

	return config.ProtectedInterval(
		minimumDBMInterval,
		maximumDBMInterval,
		duration,
	)
}

func (ipt *Input) slowOperationsInterval() time.Duration {
	duration := defaultSlowOperationsInterval
	if ipt.Dbm != nil && ipt.Dbm.SlowOperations != nil && ipt.Dbm.SlowOperations.Interval.Duration > 0 {
		duration = ipt.Dbm.SlowOperations.Interval.Duration
	}

	return config.ProtectedInterval(
		minimumDBMInterval,
		maximumDBMInterval,
		duration,
	)
}

func (ipt *Input) slowOperationsLimit() int64 {
	if ipt.Dbm == nil || ipt.Dbm.SlowOperations == nil || ipt.Dbm.SlowOperations.MaxOperations <= 0 {
		return defaultSlowOperationsLimit
	}
	return ipt.Dbm.SlowOperations.MaxOperations
}
