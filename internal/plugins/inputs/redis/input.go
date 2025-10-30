// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package redis collects redis metrics.
package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/redis/go-redis/v9"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	dknet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
)

const (
	maxInterval = 60 * time.Minute
	minInterval = 15 * time.Second

	measureuemtRedisConfig        = "redis_config"
	measureuemtRedisHotKey        = "redis_hotkey"
	measureuemtRedisBigKey        = "redis_bigkey"
	measureuemtRedisClient        = "redis_client"
	measureuemtRedisClientsStat   = "redis_clients_stat"
	measureuemtRedisClientLogging = "redis_client"
	measureuemtRedisCluster       = "redis_cluster"
	measureuemtRedisCommandStat   = "redis_command_stat"
	measureuemtRedisDB            = "redis_db"
	measureuemtRedisLatency       = "redis_latency"
	measurementRedisInfo          = "redis_info"
	measureuemtRedisReplica       = "redis_replica"
	measureuemtRedisSlowLog       = "redis_slowlog"
	measureuemtRedisTopology      = "redis_topology"

	measureuemtRedis = "redis"

	modeCluster     = "cluster"
	modeSentinel    = "sentinel"
	modeStandalone  = "standalone"
	modeMasterSlave = "master-slave"

	targetRoleReplica = "replica"
	targetRoleMaster  = "master"
)

var (
	inputName                        = "redis"
	catalogName                      = "db"
	l                                = logger.DefaultSLogger("redis")
	_           inputs.ElectionInput = (*Input)(nil)
)

type redisLog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`

	MatchDeprecated string `toml:"match,omitempty"`
}

type redisCluster struct {
	Hosts []string `toml:"hosts"`
}

type redisMasterSlave struct {
	Hosts    []string       `toml:"hosts"`
	Sentinel *redisSentinel `toml:"sentinel"`
}

type redisSentinel struct {
	Hosts      []string `toml:"hosts"`
	Password   string   `toml:"password"`
	MasterName string   `toml:"master_name"`
}

type Input struct {
	Host           string `toml:"host"`
	Port           int    `toml:"port"`
	UnixSocketPath string `toml:"unix_socket_path"`

	Cluster     *redisCluster     `toml:"cluster"`
	MasterSlave *redisMasterSlave `toml:"master_slave"`

	Username string `toml:"username"`
	Password string `toml:"password"`

	Election bool `toml:"election"`

	Timeout string `toml:"connect_timeout"`

	// deprecated TLS configuration, use *TLSClientConfig.
	TLSOpenDeprecated            bool   `toml:"tls_open"`
	CacertFileDeprecated         string `toml:"tls_ca"`               // use TLSConf.CaCerts
	CertFileDeprecated           string `toml:"tls_cert"`             // use TLSConf.Cert
	KeyFileDeprecated            string `toml:"tls_key"`              // use TLSConf.CertKey
	InsecureSkipVerifyDeprecated bool   `toml:"insecure_skip_verify"` // use TLSConf.InsecureSkipVerify

	*dknet.TLSClientConfig

	ClientListCollector *clientListCollector `toml:"collect_client_list"`

	EnableSlowLog           bool `toml:"slow_log"`
	CollectAllSlowLog       bool `toml:"all_slow_log"`
	SlowlogMaxLen           int  `toml:"slowlog_max_len"`
	SlowlogMaxLenDeprecated int  `toml:"slowlog-max-len"`

	// global metric collect interval
	Interval              time.Duration `toml:"interval"`
	EnableLatencyQuantile bool          `toml:"latency_percentiles"`

	// topology refresh interval for cluster/sentinel mode
	TopologyRefreshInterval time.Duration `toml:"topology_refresh_interval"`

	// config collect interval (less frequent than metrics)
	ConfigCollectInterval time.Duration `toml:"config_collect_interval"`

	// Deprecated: use HotBigKeys
	HotkeyDeprecated bool           `toml:"hotkey"`
	BigKeyDeprecated bool           `toml:"bigkey"`
	HotBigKeys       *hotbigkeyConf `toml:"hot_big_keys"`

	Log *redisLog `toml:"log"`

	DBDeprecated int   `toml:"db"` // use DBs
	DBs          []int `toml:"dbs"`

	mergedTags,
	Tags map[string]string `toml:"tags"`

	MeasurementVersion string `toml:"measurement_version"`

	instances []*instance
	tlsConf   *tls.Config

	// lastCustomerObject *customerObjectMeasurement

	timeoutDuration time.Duration

	tail *tailer.Tailer

	crdb *redis.ClusterClient
	srdb *redis.SentinelClient

	isInitialized, pause bool
	pauseCh              chan bool
	restartCh            chan struct{} // topology change restart signal

	semStop *cliutils.Sem // start stop signal

	// collector management
	collectorGroup    *goroutine.Group
	collectorCancel   context.CancelFunc
	collectorsRunning bool

	startUpUnix int64
	feeder      dkio.Feeder

	tagger  datakit.GlobalTagger
	ptsTime time.Time
}

type redisCPUUsage struct {
	sys,
	user float64
}

func (ipt *Input) ElectionEnabled() bool {
	return ipt.Election
}

func (ipt *Input) setupInstances(ctx context.Context) error {
	if ipt.Cluster != nil {
		return ipt.setupClusterInstances(ctx)
	}
	if ipt.MasterSlave != nil {
		if ipt.MasterSlave.Sentinel != nil {
			return ipt.setupSentinelInstances(ctx)
		}
		return ipt.setupMasterSlaveInstances(ctx)
	}
	return ipt.setupStandaloneInstances(ctx)
}

func (ipt *Input) setupClusterInstances(ctx context.Context) error {
	if err := ipt.setupCluster(ctx); err != nil {
		return err
	}

	arr, err := ipt.scanClusterMasters(ctx)
	if err != nil {
		return err
	}

	l.Infof("cluster found %d instances", len(arr))
	ipt.instances = arr

	// initialize clients
	if err := ipt.initializeClients(ctx); err != nil {
		l.Errorf("initialize clients failed: %s", err)
		return err
	}
	return nil
}

func (ipt *Input) setupSentinelInstances(ctx context.Context) error {
	if err := ipt.setupSentinel(ctx); err != nil {
		return err
	}

	inst, err := ipt.sentinelDiscoverMaster(ctx)
	if err != nil {
		return err
	}

	l.Infof("sentinel found instance %s", inst.String())
	ipt.instances = []*instance{inst}

	// initialize clients
	if err := ipt.initializeClients(ctx); err != nil {
		l.Errorf("initialize clients failed: %s", err)
		return err
	}
	return nil
}

func (ipt *Input) setupMasterSlaveInstances(ctx context.Context) error {
	inst, err := ipt.setupMasterSlave(ctx)
	if err != nil {
		return err
	}

	l.Infof("master-slave found instance %s with %d replicas", inst.String(), len(inst.replicas))
	ipt.instances = []*instance{inst}

	// initialize clients
	if err := ipt.initializeClients(ctx); err != nil {
		l.Errorf("initialize clients failed: %s", err)
		return err
	}
	return nil
}

func (ipt *Input) setupStandaloneInstances(ctx context.Context) error {
	inst := newInstance()
	inst.ipt = ipt
	inst.mode = modeStandalone
	inst.addr = fmt.Sprintf("%s:%d", ipt.Host, ipt.Port)
	inst.host = ipt.Host
	inst.setup()

	// Connect to the standalone instance (no replica discovery)
	if rdb, err := ipt.newClient(ctx, inst.addr); err == nil {
		inst.cc = &colclient{rdb: rdb}
	} else {
		return fmt.Errorf("connect to standalone redis %s failed: %w", inst.addr, err)
	}

	l.Infof("standalone found instance %s", inst.String())
	ipt.instances = []*instance{inst}

	return nil
}

func (ipt *Input) initializeClients(ctx context.Context) error {
	for _, node := range ipt.instances {
		if rdb, err := ipt.newClient(ctx, node.addr); err == nil {
			node.cc = &colclient{rdb: rdb}
			l.Infof("init redis client on master/node %q ok", node.addr)
		} else {
			return fmt.Errorf("connect to master/node %q failed: %w", node.addr, err)
		}

		for _, rep := range node.replicas {
			if rdb, err := ipt.newClient(ctx, rep.addr); err == nil {
				rep.cc = &colclient{rdb: rdb}
				l.Infof("init redis client on replica %q ok", rep.addr)
			} else {
				l.Warnf("init redis client on replica %q failed: %s, ignored", rep.addr, err)
			}
		}
		l.Infof("redis instances: %s has %d replicas", node.addr, len(node.replicas))
	}
	return nil
}

func (ipt *Input) newClient(ctx context.Context, addr string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		ClientName:   "datakit",
		TLSConfig:    ipt.tlsConf,
		Username:     ipt.Username,
		Password:     ipt.Password,     // no password set
		DB:           ipt.DBDeprecated, // use default DB
		PoolSize:     3,
		MinIdleConns: 1,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		DialTimeout:  5 * time.Second,
	})

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		l.Warnf("failed to create client on %q: %s", addr, err)
		_ = rdb.Close()
		return nil, err
	} else {
		return rdb, nil
	}
}

func (ipt *Input) retryInitCfg() error {
	for {
		if err := ipt.initCfg(); err != nil {
			l.Errorf("initCfg failed: %s, will retry", err)

			select {
			case <-datakit.Exit.Wait():
				return fmt.Errorf("datakit is exiting, retryInitCfg failed")
			case <-ipt.semStop.Wait():
				return fmt.Errorf("input is stopping, retryInitCfg failed")
			case <-time.After(5 * time.Second):
			}
		} else {
			l.Info("initCfg successful")
			return nil
		}
	}
}

func (ipt *Input) initCfg() error {
	if ipt.isInitialized {
		return nil
	}

	var err error
	ipt.timeoutDuration, err = time.ParseDuration(ipt.Timeout)
	if err != nil {
		ipt.timeoutDuration = 10 * time.Second
	}

	if ipt.SlowlogMaxLenDeprecated > 0 {
		ipt.SlowlogMaxLen = ipt.SlowlogMaxLenDeprecated // use old value
	}

	if ipt.SlowlogMaxLen == 0 {
		ipt.SlowlogMaxLen = 128
	}

	ipt.TLSClientConfig = dknet.MergeTLSConfig(
		ipt.TLSClientConfig,
		[]string{ipt.CacertFileDeprecated},
		ipt.CertFileDeprecated,
		ipt.KeyFileDeprecated,
		ipt.TLSOpenDeprecated,
		ipt.InsecureSkipVerifyDeprecated,
	)

	ipt.tlsConf, err = ipt.TLSClientConfig.TLSConfigWithBase64()
	if err != nil {
		l.Errorf("TLSConfigWithBase64 error: %s", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), ipt.timeoutDuration)
	defer cancel()

	// setup client
	if err := ipt.setupInstances(ctx); err != nil {
		return err
	}

	ipt.isInitialized = true

	return nil
}

func (*Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipelineCfg,
	}
	return pipelineMap
}

func (*Input) LogExamples() map[string]map[string]string {
	return map[string]map[string]string{
		inputName: {
			"Redis log": `122:M 14 May 2019 19:11:40.164 * Background saving terminated with success`,
		},
	}
}

func (ipt *Input) GetPipeline() []tailer.Option {
	opts := []tailer.Option{
		tailer.WithSource(inputName),
		tailer.WithService(inputName),
	}
	if ipt.Log != nil {
		opts = append(opts, tailer.WithPipeline(ipt.Log.Pipeline))
	}
	return opts
}

func (ipt *Input) Collect(ctx context.Context) error {
	if !ipt.isInitialized {
		return fmt.Errorf("redis collector not initialized")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, ipt.timeoutDuration)
	defer cancel()

	for _, inst := range ipt.instances {
		if err := inst.collect(timeoutCtx); err != nil {
			l.Warnf("instance %s collect failed: %s", inst.String(), err)
		}
	}

	return nil
}

func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	opts := []tailer.Option{
		tailer.WithSource(inputName),
		tailer.WithService(inputName),
		tailer.WithPipeline(ipt.Log.Pipeline),
		tailer.WithIgnoredStatuses(ipt.Log.IgnoreStatus),
		tailer.WithCharacterEncoding(ipt.Log.CharacterEncoding),
		tailer.EnableMultiline(true),
		tailer.WithMaxMultilineLength(int64(float64(config.Cfg.Dataway.MaxRawBodySize) * 0.8)),
		tailer.WithMultilinePatterns([]string{ipt.Log.MultilineMatch}),
		tailer.WithExtraTags(inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, "")),
		tailer.EnableDebugFields(config.Cfg.EnableDebugFields),
	}

	var err error
	ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opts...)
	if err != nil {
		l.Error("NewTailer: %s", err)

		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
		)
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_redis"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

func (ipt *Input) startHotBigKeyCollectors(ctx context.Context) {
	if !ipt.HotBigKeys.Enable {
		l.Info("hot/big key scanner disabled")
		return
	}
	if !ipt.isInitialized {
		l.Warn("hot/big key scanner not initialized")
		return
	}
	if ipt.collectorGroup == nil {
		l.Error("collector group not initialized")
		return
	}

	for _, inst := range ipt.instances {
		inst.setupHotBigKeyScanners()
		if len(inst.hbScanners) == 0 {
			l.Warnf("no hot/big key scanner for instance %s", inst.addr)
			continue
		}
		// start one scanner for each instance
		inst := inst
		scanner := inst.hbScanners[0]
		ipt.collectorGroup.Go(func(goCtx context.Context) error {
			inst.doStartCollectHotBigKeys(ctx, scanner)
			return nil
		})
	}
	l.Infof("start hot big key collectors on %d instances", len(ipt.instances))
}

func (ipt *Input) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ipt.setup()

	// non-election mode: start collectors immediately
	if !ipt.ElectionEnabled() || config.Cfg.Election == nil || !config.Cfg.Election.Enable {
		l.Info("election mode disabled, starting collectors")
		ipt.startCollectors(ctx)
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("received exit signal")
			ipt.stopCollectors()
			ipt.exit()
			return

		case <-ipt.semStop.Wait():
			l.Info("received stop signal")
			ipt.stopCollectors()
			ipt.exit()
			return

		case ipt.pause = <-ipt.pauseCh:
			if ipt.pause {
				// election lost: stop collectors first, then cleanup
				l.Info("election lost")
				ipt.stopCollectors() // wait for all goroutines to exit
				ipt.cleanup()        // safe to cleanup now
			} else {
				// election won: start collectors
				l.Info("election won")
				ipt.startCollectors(ctx)
			}

		case <-ipt.restartCh:
			// topology changed: restart all collectors
			l.Info("topology changed, restarting collectors")
			ipt.stopCollectors()     // stop all collectors first
			ipt.cleanup()            // cleanup old connections
			ipt.startCollectors(ctx) // reinitialize and restart
		}
	}
}

func (ipt *Input) startCollectors(ctx context.Context) {
	if ipt.collectorsRunning {
		l.Debug("collectors already running")
		return
	}

	l.Info("starting all collectors")

	// initialize config
	if !ipt.isInitialized {
		if err := ipt.retryInitCfg(); err != nil {
			return
		}
	}

	subCtx, cancel := context.WithCancel(ctx)
	ipt.collectorCancel = cancel
	ipt.collectorGroup = goroutine.NewGroup(goroutine.Option{Name: "redis_collectors"})

	// metrics collector
	ipt.collectorGroup.Go(func(gCtx context.Context) error {
		ipt.runMetricsCollector(subCtx)
		return nil
	})

	// config collector
	ipt.collectorGroup.Go(func(gCtx context.Context) error {
		ipt.runConfigCollector(subCtx)
		return nil
	})

	// topology refresher
	if ipt.needsTopologyRefresh() {
		ipt.collectorGroup.Go(func(gCtx context.Context) error {
			ipt.runTopologyRefresher(subCtx)
			return nil
		})
	}

	// start hot/big key collectors
	ipt.startHotBigKeyCollectors(subCtx)

	ipt.collectorsRunning = true
	l.Info("all collectors started successfully")
}

func (ipt *Input) stopCollectors() {
	if !ipt.collectorsRunning {
		l.Debug("collectors not running")
		return
	}

	l.Info("stopping all collectors")

	// cancel context to signal all collectors to stop
	if ipt.collectorCancel != nil {
		ipt.collectorCancel()
		ipt.collectorCancel = nil
	}

	// wait for main collector goroutines to exit
	if ipt.collectorGroup != nil {
		if err := ipt.collectorGroup.Wait(); err != nil {
			l.Errorf("collector group wait error: %s", err)
		}
		ipt.collectorGroup = nil
	}

	ipt.collectorsRunning = false

	l.Info("all collectors stopped")
}

func (ipt *Input) needsTopologyRefresh() bool {
	// cluster mode needs topology refresh
	if ipt.Cluster != nil {
		return true
	}
	// sentinel mode needs topology refresh
	if ipt.MasterSlave != nil && ipt.MasterSlave.Sentinel != nil {
		return true
	}
	// standalone mode doesn't need topology refresh
	return false
}

func (ipt *Input) runMetricsCollector(ctx context.Context) {
	ticker := time.NewTicker(ipt.Interval)
	defer ticker.Stop()

	ipt.ptsTime = ntp.Now()

	for {
		select {
		case <-ctx.Done():
			l.Info("metrics collector stopped")
			return
		case tt := <-ticker.C:
			ipt.doMetricsCollect(ctx, tt)
		}
	}
}

func (ipt *Input) doMetricsCollect(ctx context.Context, tickTime time.Time) {
	ipt.ptsTime = inputs.AlignTime(tickTime, ipt.ptsTime, ipt.Interval)

	if err := ipt.Collect(ctx); err != nil {
		l.Errorf("Collect: %s", err)
	}
	ipt.feedUpMetric()
}

func (ipt *Input) runConfigCollector(ctx context.Context) {
	ticker := time.NewTicker(ipt.ConfigCollectInterval)
	defer ticker.Stop()

	for {
		ipt.doConfigCollect(ctx)
		select {
		case <-ctx.Done():
			l.Info("config collector stopped")
			return
		case <-ticker.C:
		}
	}
}

func (ipt *Input) doConfigCollect(ctx context.Context) {
	if !ipt.isInitialized {
		l.Debug("config collector: not initialized, skipped")
		return
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, ipt.timeoutDuration)
	defer cancel()

	for _, inst := range ipt.instances {
		for _, n := range inst.nodes() {
			inst.setCurrentNode(n.cli, n.rep, n.host, n.addr)

			inst.collectConfig(timeoutCtx)
		}
	}
}

func (ipt *Input) runTopologyRefresher(ctx context.Context) {
	ticker := time.NewTicker(ipt.TopologyRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			l.Info("topology refresher stopped")
			return
		case <-ticker.C:
			ipt.doTopologyRefresh(ctx)
		}
	}
}

func (ipt *Input) doTopologyRefresh(ctx context.Context) {
	if !ipt.isInitialized {
		l.Debug("topology refresher: not initialized, skipped")
		return
	}

	if ipt.checkNodesChanged(ctx) {
		l.Info("topology changed, sending restart signal")
		// send restart signal to main loop (non-blocking)
		select {
		case ipt.restartCh <- struct{}{}:
		default:
			l.Debug("restart signal already pending")
		}
	}
}

func (ipt *Input) setup() {
	l = logger.SLogger(inputName)

	l.Infof("%s input started", inputName)

	ipt.startUpUnix = time.Now().Unix()
	ipt.Interval = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval)

	if ipt.Election {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.ElectionTags(), ipt.Tags, ipt.Host)
	} else {
		ipt.mergedTags = inputs.MergeTags(ipt.tagger.HostTags(), ipt.Tags, ipt.Host)
	}

	l.Debugf("merged tags: %+#v", ipt.mergedTags)

	// add deprecated old db.
	if ipt.DBDeprecated != -1 && !isSlicesHave(ipt.DBs, ipt.DBDeprecated) {
		ipt.DBs = append(ipt.DBs, ipt.DBDeprecated)
	}

	if len(ipt.DBs) < 1 && (ipt.HotkeyDeprecated || ipt.BigKeyDeprecated) {
		l.Infof("no DB selected, hot/big key collect disabled")
	}
}

// closeDBConnections closes all database connections.
func (ipt *Input) closeDBConnections() {
	if ipt.crdb != nil {
		if err := ipt.crdb.Close(); err != nil {
			l.Info("redis client.Close: %s, ignored", err.Error())
		}
	}

	if ipt.srdb != nil {
		if err := ipt.srdb.Close(); err != nil {
			l.Info("redis client.Close: %s, ignored", err.Error())
		}
	}
	for _, inst := range ipt.instances {
		inst.stop()
	}
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
		l.Info("redis log exit")
	}

	ipt.closeDBConnections()
}

// cleanup performs cleanup operations.
func (ipt *Input) cleanup() {
	if !ipt.isInitialized {
		l.Debug("cleanup: already cleaned up, skipped")
		return
	}

	// close db connections
	ipt.closeDBConnections()

	// reset state
	ipt.isInitialized = false
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (*Input) Catalog() string { return catalogName }

func (*Input) SampleConfig() string { return configSample }

func (*Input) AvailableArchs() []string { return datakit.AllOSWithElection }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&redisMeasurement{},
		&configMeasurement{},
		&bigKeyMeasurement{},
		&hotkeyMeasurement{},
		// &clientMetricMeasurement{},
		&clientLoggingMeasurement{},
		// &clusterMeasurement{},
		// &commandMeasurement{},
		// &dbMeasurement{},
		// &infoMeasurement{},
		// &replicaMeasurement{},
		&latencyMeasurement{},
		&slowlogMeasurement{},
		&topologyChangeMeasurement{},
		// &customerObjectMeasurement{},
		&inputs.UpMeasurement{},
	}
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

func defaultInput() *Input {
	getReplicaFieldMap()

	return &Input{
		Timeout:   "10s",
		pauseCh:   make(chan bool, inputs.ElectionPauseChannelLength),
		restartCh: make(chan struct{}, 1),
		Tags:      make(map[string]string),

		DBDeprecated: -1,
		DBs:          []int{0},

		Interval:                time.Second * 15,
		TopologyRefreshInterval: 10 * time.Minute,
		ConfigCollectInterval:   1 * time.Hour,
		semStop:                 cliutils.NewSem(),
		Election:                true,
		feeder:                  dkio.DefaultFeeder(),
		tagger:                  datakit.DefaultGlobalTagger(),
		SlowlogMaxLen:           128,
		HotBigKeys:              defaultHotBitKeyConf(),

		ClientListCollector: &clientListCollector{
			CollectLogOnFlags: "bxOR",
		},
	}
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}

func isSlicesHave(s []int, index int) bool {
	for _, i := range s {
		if i == index {
			return true
		}
	}
	return false
}
