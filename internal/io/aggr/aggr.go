// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package aggr is a aggregator for datakit.
package aggr

import (
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/GuanceCloud/cliutils/aggregate"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

type endpointRoute string

const (
	aggregateRoute          endpointRoute = "aggregate"
	tailSamplingRoute       endpointRoute = "tail_sampling"
	tailSamplingConfigRoute endpointRoute = "tail_sampling_config"
)

var (
	log = logger.DefaultSLogger("aggr")

	routePathByType = map[endpointRoute]string{
		aggregateRoute:          datakit.Aggregate,
		tailSamplingRoute:       datakit.TailSampling,
		tailSamplingConfigRoute: datakit.TailSamplingConfig,
	}
)

type Aggregator struct {
	Endpoints                   []string                          `toml:"endpoints"`
	Timeout                     time.Duration                     `toml:"timeout"`
	MaxRawBodySize              int                               `toml:"max_raw_body_size"`
	UseLocalConfig              bool                              `toml:"use_local_config"`
	LocalConfigDir              string                            `toml:"local_config_dir"`
	LocalMetricConfigFile       string                            `toml:"local_metric_config_file"`
	LocalTailSamplingConfigFile string                            `toml:"local_tail_sampling_config_file"`
	DW                          *dataway.Dataway                  `toml:"-"`
	Internal                    time.Duration                     `toml:"-"`
	PullFunc                    func(args string) ([]byte, error) `toml:"-"`

	metricConfig        *aggregate.AggregatorConfigure
	tailSamplingConfig  *aggregate.TailSamplingConfigs
	metricEnabled       bool
	tailSamplingEnabled bool

	aggrURL               string
	tailSamplingURL       string
	tailSamplingConfigURL string
	Transport             *http.Transport `toml:"-"`
	lastSendTime          time.Time

	stateMu        sync.RWMutex
	tsConfigSendMu sync.Mutex
}

type ProcessResult struct {
	Points []*point.Point
	// SelectedPoints is the total points selected by aggregation or tail sampling.
	SelectedPoints int
	// BatchPackages is the count of metric aggregation batches selected.
	BatchPackages int
	// TailSamplingPackages is the count of tail-sampling packages selected.
	TailSamplingPackages int
	Consumed             bool
}

func (ag *Aggregator) StartAggr() {
	log = logger.SLogger("aggr")
	log.Infof("start aggregator config reloader")
	ag.initHTTP()
	ag.reloadConfigs()
	aggregate.SetLogging(log)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ag.reloadConfigs()
		case <-datakit.Exit.Wait():
			return
		}
	}
}

func (ag *Aggregator) reloadConfigs() {
	ag.applyConfigSourceDefaults()

	var (
		metricConfigBody []byte
		tsConfigBody     []byte
		metricLoaded     bool
		tsLoaded         bool
	)

	if ag.UseLocalConfig {
		metricLoaded = true
		tsLoaded = true
	} else {
		if ag.PullFunc == nil {
			log.Debugf("skip config reload: PullFunc is nil")
			return
		}
		bts, err := ag.PullFunc("aggr=true")
		if err != nil {
			log.Errorf("pull metric config failed: %v", err)
		} else {
			metricConfigBody = bts
			metricLoaded = true
		}

		tsBts, err := ag.PullFunc("tail-sampling=true")
		if err != nil {
			log.Errorf("pull tail sampling config failed: %v", err)
		} else {
			tsConfigBody = tsBts
			tsLoaded = true
		}
	}

	ag.stateMu.Lock()
	if ag.UseLocalConfig {
		if metricLoaded {
			ag.loadMetricConfigFromFile(filepath.Join(ag.LocalConfigDir, ag.LocalMetricConfigFile))
		}
		if tsLoaded {
			ag.loadTailSamplingConfigFromFile(filepath.Join(ag.LocalConfigDir, ag.LocalTailSamplingConfigFile))
		}
	} else {
		if metricLoaded {
			ag.loadMetricConfigFromDataway(metricConfigBody)
		}
		if tsLoaded {
			ag.loadTailSamplingConfigFromDataway(tsConfigBody)
		}
	}

	ag.metricEnabled = false
	ag.tailSamplingEnabled = false

	if ag.metricConfig != nil {
		if err := ag.metricConfig.Setup(); err != nil {
			log.Errorf("setup metric config failed: %v", err)
		}
		if len(ag.metricConfig.AggregateRules) > 0 {
			ag.metricEnabled = true
		}
	}
	if ag.tailSamplingConfig != nil {
		ag.tailSamplingEnabled = true
	}
	metricEnabled := ag.metricEnabled
	tailSamplingEnabled := ag.tailSamplingEnabled
	metricConfig := ag.metricConfig
	tailSamplingConfig := ag.tailSamplingConfig
	ag.stateMu.Unlock()

	if tailSamplingEnabled {
		ag.sendTSConfigToDW()
	}

	tailSamplingCfgStr := "<nil>"
	if tailSamplingConfig != nil {
		tailSamplingCfgStr = tailSamplingConfig.ToString()
	}

	log.Debugf("config reload done: metric_enabled=%t tail_sampling_enabled=%t tail sampling config=%s \n aggr config is %+v",
		metricEnabled, tailSamplingEnabled, tailSamplingCfgStr, metricConfig)
}

func (ag *Aggregator) Process(cat point.Category, input string, pts []*point.Point) (*ProcessResult, error) {
	res := &ProcessResult{Points: pts}
	snapshot := ag.snapshotState()

	if snapshot.metricEnabled {
		batchMap := ag.pickMetricWithConfig(snapshot.metricConfig, cat.String(), pts)
		res.SelectedPoints += countSelectedMetricPoints(batchMap)
		res.BatchPackages = countMetricBatchPackages(batchMap)
		if len(batchMap) > 0 {
			log.Debugf("metric enable,pick metric batch len=%d for category =%s", len(batchMap), cat.String())
		}

		if err := ag.SendMetricBatches(batchMap); err != nil {
			log.Errorf("send metric batches failed: %v", err)
			// Metrics are already recorded in SendMetricBatches
		}
	}

	if !snapshot.tailSamplingEnabled || snapshot.tailSamplingConfig == nil {
		return res, nil
	}

	switch cat { // nolint:exhaustive
	case point.Tracing:
		if snapshot.tailSamplingConfig.Tracing == nil {
			return res, nil
		}

		packages := ag.pickTraceWithConfig(input, pts, snapshot.tailSamplingConfig)
		res.SelectedPoints += countTailSamplingPoints(packages)
		res.TailSamplingPackages = countTailSamplingPackages(packages)
		if err := ag.SendTailSamplingPackages(packages); err != nil {
			log.Errorf("process tracing points failed: %v", err)
			return res, err // Return original result instead of nil to avoid data loss
		}
		log.Debugf("process tracing points done: input=%s pts=%d packages=%d consumed=true", input, len(pts), len(packages))
		res.Points = nil
		res.Consumed = true
		return res, nil

	case point.Logging:
		if snapshot.tailSamplingConfig.Logging == nil {
			return res, nil
		}

		packages, passedThrough := ag.pickLoggingWithConfig(input, pts, snapshot.tailSamplingConfig)
		res.SelectedPoints += countTailSamplingPoints(packages)
		res.TailSamplingPackages = countTailSamplingPackages(packages)
		if err := ag.SendTailSamplingPackages(packages); err != nil {
			log.Errorf("process logging points failed: %v", err)
			return nil, err
		}
		log.Debugf("process logging points done: input=%s pts=%d packages=%d passthrough=%d", input, len(pts), len(packages), len(passedThrough))
		res.Points = passedThrough
		return res, nil

	case point.RUM:
		if snapshot.tailSamplingConfig.RUM == nil {
			return res, nil
		}

		packages, passedThrough := ag.pickRUMWithConfig(input, pts, snapshot.tailSamplingConfig)
		res.SelectedPoints += countTailSamplingPoints(packages)
		res.TailSamplingPackages = countTailSamplingPackages(packages)
		if err := ag.SendTailSamplingPackages(packages); err != nil {
			log.Errorf("process rum points failed: %v", err)
			return nil, err
		}
		log.Debugf("process rum points done: input=%s pts=%d packages=%d passthrough=%d", input, len(pts), len(packages), len(passedThrough))
		res.Points = passedThrough
		return res, nil

	default:
		return res, nil
	}
}

func countSelectedMetricPoints(batchMap map[uint64]*aggregate.Batchs) int {
	cnt := 0
	for _, batch := range batchMap {
		if batch == nil || len(batch.Batchs) == 0 {
			continue
		}
		for _, one := range batch.Batchs {
			if one == nil || one.Points == nil {
				continue
			}
			cnt += len(one.Points.Arr)
		}
	}
	return cnt
}

func countMetricBatchPackages(batchMap map[uint64]*aggregate.Batchs) int {
	cnt := 0
	for _, batch := range batchMap {
		if batch == nil || len(batch.Batchs) == 0 {
			continue
		}
		cnt += len(batch.Batchs)
	}
	return cnt
}

func countTailSamplingPoints(packages map[uint64]*aggregate.DataPacket) int {
	cnt := 0
	for _, pkg := range packages {
		if pkg == nil {
			continue
		}

		cnt += int(pkg.PointCount)
	}
	return cnt
}

func countTailSamplingPackages(packages map[uint64]*aggregate.DataPacket) int {
	cnt := 0
	for _, pkg := range packages {
		if pkg == nil {
			continue
		}

		if pkg.PointCount <= 0 || len(pkg.PointsPayload) == 0 {
			continue
		}
		cnt++
	}
	return cnt
}

func countSelectedMetricPointsInBatch(batch *aggregate.Batchs) int {
	if batch == nil || len(batch.Batchs) == 0 {
		return 0
	}

	cnt := 0
	for _, b := range batch.Batchs {
		if b == nil || b.Points == nil {
			continue
		}
		cnt += len(b.Points.Arr)
	}
	return cnt
}

type aggrStateSnapshot struct {
	metricConfig        *aggregate.AggregatorConfigure
	tailSamplingConfig  *aggregate.TailSamplingConfigs
	metricEnabled       bool
	tailSamplingEnabled bool
}

func (ag *Aggregator) snapshotState() aggrStateSnapshot {
	ag.stateMu.RLock()
	defer ag.stateMu.RUnlock()

	return aggrStateSnapshot{
		metricConfig:        ag.metricConfig,
		tailSamplingConfig:  ag.tailSamplingConfig,
		metricEnabled:       ag.metricEnabled,
		tailSamplingEnabled: ag.tailSamplingEnabled,
	}
}

func (ag *Aggregator) tailSamplingConfigVersion() int64 {
	ag.stateMu.RLock()
	defer ag.stateMu.RUnlock()

	if ag.tailSamplingConfig == nil {
		return 0
	}
	return ag.tailSamplingConfig.Version
}

func (ag *Aggregator) PickMetric(category string, pts []*point.Point) map[uint64]*aggregate.Batchs {
	ag.stateMu.RLock()
	cfg := ag.metricConfig
	ag.stateMu.RUnlock()

	return ag.pickMetricWithConfig(cfg, category, pts)
}

func (ag *Aggregator) pickMetricWithConfig(cfg *aggregate.AggregatorConfigure, category string, pts []*point.Point) map[uint64]*aggregate.Batchs {
	if cfg == nil || len(cfg.AggregateRules) == 0 {
		return nil
	}
	batchMap := cfg.PickPoints(category, pts)

	if len(batchMap) > 0 {
		log.Debugf("PickMetric batch=%d category is %s  from pts len=%d", len(batchMap), category, len(pts))
	}
	return batchMap
}

func (ag *Aggregator) PickTrace(input string, pts []*point.Point) map[uint64]*aggregate.DataPacket {
	ag.stateMu.RLock()
	cfg := ag.tailSamplingConfig
	ag.stateMu.RUnlock()

	return ag.pickTraceWithConfig(input, pts, cfg)
}

func (ag *Aggregator) pickTraceWithConfig(input string, pts []*point.Point, cfg *aggregate.TailSamplingConfigs) map[uint64]*aggregate.DataPacket {
	if cfg == nil {
		return nil
	}

	packages := aggregate.PickTrace(input, pts, cfg.Version)
	for _, pkg := range packages {
		pkg.Token = ag.defaultToken()
	}

	log.Debugf("PickTrace packages=%d", len(packages))
	return packages
}

func (ag *Aggregator) PickLogging(input string, pts []*point.Point) (map[uint64]*aggregate.DataPacket, []*point.Point) {
	ag.stateMu.RLock()
	cfg := ag.tailSamplingConfig
	ag.stateMu.RUnlock()

	return ag.pickLoggingWithConfig(input, pts, cfg)
}

func (ag *Aggregator) pickLoggingWithConfig(input string, pts []*point.Point,
	cfg *aggregate.TailSamplingConfigs,
) (map[uint64]*aggregate.DataPacket, []*point.Point) {
	if cfg == nil || cfg.Logging == nil {
		return nil, pts
	}

	packages := map[uint64]*aggregate.DataPacket{}
	passedThrough := pts
	for _, group := range cfg.Logging.GroupDimensions {
		groupPackages, nextPassedThrough := group.PickLogging(input, passedThrough)
		for _, pkg := range groupPackages {
			pkg.Token = ag.defaultToken()
			pkg.ConfigVersion = cfg.Version
			packages[tailSamplingPacketPickKey(pkg)] = pkg
		}
		passedThrough = nextPassedThrough
	}

	log.Debugf("PickLogging packages=%d passthrough=%d", len(packages), len(passedThrough))
	return packages, passedThrough
}

func (ag *Aggregator) PickRUM(input string, pts []*point.Point) (map[uint64]*aggregate.DataPacket, []*point.Point) {
	ag.stateMu.RLock()
	cfg := ag.tailSamplingConfig
	ag.stateMu.RUnlock()

	return ag.pickRUMWithConfig(input, pts, cfg)
}

func (ag *Aggregator) pickRUMWithConfig(input string, pts []*point.Point,
	cfg *aggregate.TailSamplingConfigs,
) (map[uint64]*aggregate.DataPacket, []*point.Point) {
	if cfg == nil || cfg.RUM == nil {
		return nil, pts
	}

	packages := map[uint64]*aggregate.DataPacket{}
	passedThrough := pts
	for _, group := range cfg.RUM.GroupDimensions {
		groupPackages, nextPassedThrough := group.PickRUM(input, passedThrough)
		for _, pkg := range groupPackages {
			pkg.Token = ag.defaultToken()
			pkg.ConfigVersion = cfg.Version
			packages[tailSamplingPacketPickKey(pkg)] = pkg
		}
		passedThrough = nextPassedThrough
	}

	log.Debugf("PickRUM packages=%d passthrough=%d", len(packages), len(passedThrough))
	return packages, passedThrough
}

func tailSamplingPacketPickKey(packet *aggregate.DataPacket) uint64 {
	if packet == nil {
		return 0
	}

	key := aggregate.HashToken(packet.Token, packet.GroupIdHash)
	key = aggregate.HashCombine(key, xxhash.Sum64String(packet.DataType))
	key = aggregate.HashCombine(key, xxhash.Sum64String(packet.GroupKey))

	return key
}

// Metrics wrapper methods for Aggregator.
func (ag *Aggregator) recordSendSuccess(sendType, category string) {
	recordSendSuccess(sendType, category)
}

func (ag *Aggregator) recordSendFailed(sendType, category, reason string) {
	recordSendFailed(sendType, category, reason)
}

func (ag *Aggregator) recordSendPoints(sendType, category string, points int) {
	recordSendPoints(sendType, category, points)
}

func (ag *Aggregator) recordLostPoints(sendType, category, reason string, points int) {
	recordLostPoints(sendType, category, reason, points)
}

func (ag *Aggregator) recordSendLatency(sendType, category string, latency time.Duration) {
	recordSendLatency(sendType, category, latency)
}
