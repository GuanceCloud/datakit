// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

var (
	keyBatchScanCostVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_redis",
			Name:      "key_scan_seconds",
			Help:      "Batch key scan cost in seconds.",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"jobs"},
	)

	keyScannedVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_redis",
			Name:      "scanned_keys",
			Help:      "Number of keys scanned.",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"job"},
	)

	dbScanCostVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "datakit",
			Subsystem: "input_redis",
			Name:      "db_scan_seconds",
			Help:      "Total DB scan cost in seconds.",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{"job"},
	)
)

func p8sMetrics() []prometheus.Collector {
	return []prometheus.Collector{
		keyBatchScanCostVec,
		keyScannedVec,
		dbScanCostVec,
	}
}

// nolint:unused
func resetP8sMetrics() {
	keyBatchScanCostVec.Reset()
	keyScannedVec.Reset()
	dbScanCostVec.Reset()
}

//nolint:gochecknoinits
func init() {
	metrics.MustRegister(p8sMetrics()...)
}

// keyInfo stores a key and its size/type/count info.
type keyInfo struct {
	key, keyt        string
	cnt              int // same freq key count
	vlen, vmem, freq int64
}

type hotbigkeyConf struct {
	Enable              bool          `toml:"enable"`
	TopN                int           `toml:"top_n"`
	MemUsageSample      int           `toml:"mem_usage_samples"`
	BatchSize           int64         `toml:"scan_batch_size"`
	BigkeyThreshouldLen int64         `toml:"bigkey_threshold_len"`
	BigkeyThreshouldMem int64         `toml:"bigkey_threshold_bytes"`
	BigkeyInterval      time.Duration `toml:"big_key_interval"`
	HotkeyInterval      time.Duration `toml:"hot_key_interval"`
	ScanSleep           time.Duration `toml:"scan_sleep"`
	TargetRole          string        `toml:"target_role"`
}

type hotbigkeyScanner struct {
	*hotbigkeyConf
	hot,
	bigMem,
	bigString,
	bigList,
	bigHash,
	bigSet,
	bigZset []*keyInfo

	cc collectorClient

	db,
	curBatch,
	sampled int

	cursor uint64
	ctx    context.Context

	flagHotkey,
	flagBigkey bool

	typeCmds []*redis.StatusCmd
	bigLenCmds,
	freqCmds []*redis.Cmd
	bigMemCmds []*redis.IntCmd

	mode, role, replicaof, addr string

	htick, btick *time.Ticker
	stop         chan any
}

func (scanner *hotbigkeyScanner) String() string {
	if scanner.replicaof != "" {
		return fmt.Sprintf("mode: %s(replica of %q) | role: %s | addr: %s",
			scanner.mode, scanner.replicaof, scanner.role, scanner.addr)
	} else {
		return fmt.Sprintf("mode: %s | role: %s | addr: %s",
			scanner.mode, scanner.role, scanner.addr)
	}
}

func defaultHotBitKeyConf() *hotbigkeyConf {
	return &hotbigkeyConf{
		// default close hot/big key collection
		Enable:              false,
		TopN:                10,
		MemUsageSample:      100,
		BatchSize:           100,
		BigkeyThreshouldLen: 5000,
		BigkeyThreshouldMem: 10 * (1 << 20),
		BigkeyInterval:      3 * time.Hour,
		HotkeyInterval:      15 * time.Minute,
		ScanSleep:           200 * time.Millisecond,
		TargetRole:          targetRoleMaster,
	}
}

func newHotBigKeyScanner(conf *hotbigkeyConf) *hotbigkeyScanner {
	scanner := &hotbigkeyScanner{
		hotbigkeyConf: conf,

		flagBigkey: true,
		flagHotkey: true,
		stop:       make(chan any),
	}

	scanner.setup()

	return scanner
}

func (scanner *hotbigkeyScanner) reset() {
	scanner.hot = scanner.hot[:0]
	scanner.bigString = scanner.bigString[:0]
	scanner.bigList = scanner.bigList[:0]
	scanner.bigHash = scanner.bigHash[:0]
	scanner.bigSet = scanner.bigSet[:0]
	scanner.bigZset = scanner.bigZset[:0]
	scanner.sampled = 0
	scanner.cursor = 0
	scanner.curBatch = 0
}

func (scanner *hotbigkeyScanner) setup() {
	if scanner.BigkeyInterval <= 0 {
		l.Warnf("reset big key interval from %v to %v", scanner.BigkeyInterval, 3*time.Hour)
		scanner.BigkeyInterval = 3 * time.Hour
	}

	if scanner.HotkeyInterval <= 0 {
		l.Warnf("reset hot key interval from %v to %v", scanner.HotkeyInterval, 15*time.Minute)
		scanner.HotkeyInterval = 15 * time.Minute
	}

	if scanner.BigkeyInterval > 0 {
		scanner.btick = time.NewTicker(scanner.BigkeyInterval)
	}

	if scanner.HotkeyInterval > 0 {
		scanner.htick = time.NewTicker(scanner.HotkeyInterval)
	}

	if scanner.BatchSize > 0 {
		scanner.typeCmds = make([]*redis.StatusCmd, 0, scanner.BatchSize)
		scanner.freqCmds = make([]*redis.Cmd, 0, scanner.BatchSize)
		scanner.bigLenCmds = make([]*redis.Cmd, 0, scanner.BatchSize)
		scanner.bigMemCmds = make([]*redis.IntCmd, 0, scanner.BatchSize)
	}
}

// scanDB scans Redis and returns the top N hot keys by LFU frequency.
func (i *instance) hotbigkeyScanDB(ctx context.Context, db int, scanner *hotbigkeyScanner) error {
	if scanner.cc == nil {
		l.Warnf("skip on invalid scanner: rdb not set", scanner.addr)
		return nil
	}

	var keyScanned int
	scanner.ctx = ctx
	scanner.db = db
	start := time.Now()

	// check mem policy for hotkeys
	if scanner.flagHotkey {
		l.Debugf("checking mem policy...")
		if x, err := scanner.cc.configGet(ctx, "maxmemory-policy"); err != nil {
			return fmt.Errorf("get maxmemory-policy: %w", err)
		} else if p := x["maxmemory-policy"]; p != "allkeys-lfu" && p != "volatile-lfu" {
			l.Warnf("skip hot key collect, maxmemory-policy is %q, expect LFU", p)
			scanner.flagHotkey = false // disable hotkeys
		}
	}

	if !scanner.flagBigkey && !scanner.flagHotkey { // do nothing
		l.Debugf("do nothing: %+#v", scanner)
		return nil
	}

	defer func() {
		jobLabel := fmt.Sprintf("h:%v,b:%v", scanner.flagHotkey, scanner.flagBigkey)
		keyScannedVec.WithLabelValues(jobLabel).Observe(float64(keyScanned))
		dbScanCostVec.WithLabelValues(jobLabel).Observe(time.Since(start).Seconds())
	}()

	if scanner.mode != modeCluster && db >= 0 {
		if err := scanner.cc.do(ctx, "SELECT", db).Err(); err != nil {
			return fmt.Errorf("SELECT db %d failed: %w", db, err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			l.Infof("hotbigkeyScanDB exited: %s", scanner.String())
			return nil
		default: // pass
		}

		if err := scanner.batchScan(); err != nil {
			l.Errorf("hotbigkeyScanDB batchScan err: %s", err.Error())
			// has error, skip one scan
			return err
		}

		scanner.curBatch++

		// Sleep every 100 batches (similar to Redis official --hotkeys -i implementation)
		// to reduce impact on production while maintaining scan efficiency
		if scanner.ScanSleep > 0 && scanner.curBatch%100 == 0 {
			time.Sleep(scanner.ScanSleep)
		}

		if scanner.cursor == 0 {
			l.Infof("hotbigkeyScanDB scan sampled end: %s,keys: %d", scanner.String(), scanner.sampled)
			break
		}
	}

	keyScanned = scanner.sampled

	return nil
}

func (scanner *hotbigkeyScanner) execPipeline(p redis.Pipeliner) error {
	res, err := p.Exec(scanner.ctx)

	// redis.Nil is not an error, treat as success
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	for _, r := range res {
		if err := r.Err(); err != nil {
			l.Errorf("%s | cmderr: %s => %s", scanner, r.String(), r.Err())
		}
	}

	return nil
}

func (scanner *hotbigkeyScanner) batchScan() error {
	start := time.Now()

	keyList, nextCursor, err := scanner.cc.scanKeys(scanner.ctx, scanner.cursor, "*", scanner.BatchSize)
	if err != nil {
		return fmt.Errorf("SCAN failed: %w", err)
	}

	scanner.cursor = nextCursor
	scanner.sampled += len(keyList)
	l.Debugf("batchScan sampled: %s, keys: %d", scanner.String(), scanner.sampled)

	if len(keyList) == 0 {
		if scanner.cursor == 0 {
			return nil
		}
		return nil
	}

	defer func() {
		keyBatchScanCostVec.WithLabelValues(fmt.Sprintf("h:%v,b:%v", scanner.flagHotkey, scanner.flagBigkey)).
			Observe(time.Since(start).Seconds())
	}()

	pipe := scanner.cc.newPipeline()
	scanner.typeCmds = scanner.typeCmds[:0]
	for _, key := range keyList {
		scanner.typeCmds = append(scanner.typeCmds, pipe.Type(scanner.ctx, key))
	}

	if err := scanner.execPipeline(pipe); err != nil {
		return fmt.Errorf("execPipeline TYPE: %w", err)
	}

	scanner.freqCmds = scanner.freqCmds[:0]
	scanner.bigMemCmds = scanner.bigMemCmds[:0]
	scanner.bigLenCmds = scanner.bigLenCmds[:0]

	// prepare commands
	for i, key := range keyList {
		if scanner.flagHotkey {
			scanner.freqCmds = append(scanner.freqCmds, pipe.Do(scanner.ctx, "OBJECT", "FREQ", key))
		}

		if scanner.flagBigkey {
			scanner.bigMemCmds = append(scanner.bigMemCmds, pipe.MemoryUsage(scanner.ctx, key, scanner.MemUsageSample))

			keyType, err := scanner.typeCmds[i].Result()
			if err != nil {
				l.Warnf("typeCmds.Result: %s", err.Error())
				scanner.bigLenCmds = append(scanner.bigLenCmds, nil)
				continue
			}

			switch keyType {
			case "string":
				scanner.bigLenCmds = append(scanner.bigLenCmds, pipe.Do(scanner.ctx, "STRLEN", key))
			case "list":
				scanner.bigLenCmds = append(scanner.bigLenCmds, pipe.Do(scanner.ctx, "LLEN", key))
			case "hash":
				scanner.bigLenCmds = append(scanner.bigLenCmds, pipe.Do(scanner.ctx, "HLEN", key))
			case "set":
				scanner.bigLenCmds = append(scanner.bigLenCmds, pipe.Do(scanner.ctx, "SCARD", key))
			case "zset":
				scanner.bigLenCmds = append(scanner.bigLenCmds, pipe.Do(scanner.ctx, "ZCARD", key))
			default:
				scanner.bigLenCmds = append(scanner.bigLenCmds, nil)
				l.Debugf("skip key type %q on %q", keyType, key)
			}
		}
	}

	if err := scanner.execPipeline(pipe); err != nil {
		return fmt.Errorf("execPipeline FREQ/MEM: %w", err)
	}

	// get current batch key's frequency and len/count
	for i, key := range keyList {
		keyType, err := scanner.typeCmds[i].Result()
		if err != nil {
			l.Warnf("typeCmds.Result: %s", err.Error())
			continue
		}

		// hot keys
		if scanner.flagHotkey {
			if scanner.freqCmds[i] == nil {
				l.Debugf("skip nil hot-cmd at #%d, key: %s", i, key)
			} else {
				freq, err := scanner.freqCmds[i].Int64()
				if err != nil {
					l.Warnf("freqCmds on %q: %s", key, err.Error())
				} else if freq > 0 {
					// TODO: we should create better policy to define a real hot key.
					scanner.mergeHotKeys(&keyInfo{key: key, keyt: keyType, freq: freq})
				}
			}
		}

		// big len key
		if scanner.flagBigkey {
			if scanner.bigLenCmds[i] == nil {
				l.Debugf("skip nil big-cmd at #%d, key: %s", i, key)
				continue
			}

			var vlen, vmem int64
			if vlen, err = scanner.bigLenCmds[i].Int64(); err != nil {
				l.Warnf("get string key len: %s", err.Error())
			}

			if vmem, err = scanner.bigMemCmds[i].Result(); err != nil {
				l.Warnf("get mem usage on key %q: %s", key, err.Error())
			}

			ki := &keyInfo{key: key, keyt: keyType, vmem: vmem, vlen: vlen}

			if vlen >= scanner.BigkeyThreshouldLen {
				switch keyType {
				case "string":
					scanner.bigString = scanner.mergeBigKeys(scanner.bigString, ki)
				case "list":
					scanner.bigList = scanner.mergeBigKeys(scanner.bigList, ki)
				case "hash":
					scanner.bigHash = scanner.mergeBigKeys(scanner.bigHash, ki)
				case "set":
					scanner.bigSet = scanner.mergeBigKeys(scanner.bigSet, ki)
				case "zset":
					scanner.bigZset = scanner.mergeBigKeys(scanner.bigZset, ki)
				}
			}

			if ki.vmem >= scanner.BigkeyThreshouldMem {
				scanner.mergeBigMemKeys(ki)
			}
		}
	}

	return nil
}

func (scanner *hotbigkeyScanner) mergeBigKeys(bs []*keyInfo, b *keyInfo) []*keyInfo {
	if len(bs) < scanner.TopN {
		l.Debugf("add big len key %s", b.key)
		bs = append(bs, b)
		sort.Slice(bs, func(i, j int) bool {
			return bs[i].vlen > bs[j].vlen
		})

		return bs
	}

	for i, x := range bs {
		if b.vlen == x.vlen {
			x.cnt++
			return bs
		}

		if b.vlen > x.vlen {
			l.Debugf("push big len key %s/%d bytes", b.key, b.vlen)
			bs[i] = b
			break
		}
	}

	return bs
}

func (scanner *hotbigkeyScanner) mergeBigMemKeys(b *keyInfo) {
	if len(scanner.bigMem) < scanner.TopN {
		l.Debugf("add big mem key %s/%d bytes", b.key, b.vlen)
		scanner.bigMem = append(scanner.bigMem, b)
		if len(scanner.bigMem) == scanner.TopN {
			sort.Slice(scanner.bigMem, func(i, j int) bool {
				return scanner.bigMem[i].vmem > scanner.bigMem[j].vmem
			})
		}
		return
	}

	for i, x := range scanner.bigMem {
		if b.vmem == x.vmem {
			x.cnt++
			return
		}

		if b.vmem > x.vmem {
			l.Debugf("push big mem key %s/%d bytes", b.key, b.vlen)
			scanner.bigMem[i] = b
			break
		}
	}
}

func (scanner *hotbigkeyScanner) mergeHotKeys(h *keyInfo) {
	if len(scanner.hot) < scanner.TopN {
		l.Debugf("add hot key %s/%d", h.key, h.freq)
		scanner.hot = append(scanner.hot, h)
		// reverse sort them if top-n elements ready, successive elements
		// are insert and sorted.
		if len(scanner.hot) == scanner.TopN {
			sort.Slice(scanner.hot, func(i, j int) bool {
				return scanner.hot[i].freq > scanner.hot[j].freq
			})
		}
		return
	}

	for i, x := range scanner.hot {
		if h.freq == x.freq {
			x.cnt++
			return
		}

		if h.freq > x.freq {
			scanner.hot[i] = h
			return
		}
	}
}

func (i *instance) hotbigPoints(scanner *hotbigkeyScanner) []*point.Point {
	var (
		pts    []*point.Point
		dbName = fmt.Sprintf("db%d", scanner.db)
	)

	opts := append(point.DefaultLoggingOptions(), point.WithTime(ntp.Now()))

	l.Debugf("new points on %d hot keys", len(scanner.hot))
	for _, ki := range scanner.hot {
		var kvs point.KVs
		kvs = kvs.Set("key_count", ki.freq).
			Set("keys_sampled", scanner.sampled).
			AddTag("db_name", dbName).
			AddTag("key", ki.key).
			Add("status", "info").
			AddTag("key_type", ki.keyt).
			AddTag("server", scanner.addr)

		for k, v := range i.mergedTags {
			kvs = kvs.AddTag(k, v)
		}

		pts = append(pts, point.NewPoint(measureuemtRedisHotKey, kvs, opts...))
	}

	l.Debugf("new points on %d big mem keys", len(scanner.bigMem))
	for _, ki := range scanner.bigMem {
		var kvs point.KVs
		kvs = kvs.Set("value_mem", ki.vmem).
			Set("value_length", ki.vlen).
			Set("keys_sampled", scanner.sampled).
			AddTag("db_name", dbName).
			AddTag("key", ki.key).
			AddTag("key_type", ki.keyt).
			AddTag("server", scanner.addr)

		msg := fmt.Sprintf("memory larger than %d bytes", scanner.BigkeyThreshouldMem)
		if ki.cnt > 0 {
			msg += fmt.Sprintf(", and there are %d same-size keys not collected", ki.cnt)
		}

		kvs = kvs.Set("message", msg)

		if scanner.BigkeyThreshouldMem > 0 {
			kvs = kvs.Add("status", "warn")
		} else {
			kvs = kvs.Add("status", "info")
		}

		for k, v := range i.mergedTags {
			kvs = kvs.AddTag(k, v)
		}

		pts = append(pts, point.NewPoint(measureuemtRedisBigKey, kvs, opts...))
	}

	for _, arr := range [][]*keyInfo{scanner.bigString, scanner.bigList, scanner.bigHash, scanner.bigSet, scanner.bigZset} {
		l.Debugf("new points on %d big keys", len(arr))

		for _, ki := range arr {
			var kvs point.KVs
			kvs = kvs.Set("value_mem", ki.vmem).
				Set("value_length", ki.vlen).
				Set("keys_sampled", scanner.sampled).
				AddTag("db_name", dbName).
				AddTag("key", ki.key).
				AddTag("key_type", ki.keyt).
				AddTag("server", scanner.addr)

			msg := fmt.Sprintf("elements larger than %d", scanner.BigkeyThreshouldLen)
			if ki.cnt > 0 {
				msg += fmt.Sprintf(", and there are %d same-len keys not collected", ki.cnt)
			}

			kvs = kvs.Set("message", msg)

			if scanner.BigkeyThreshouldLen > 0 {
				kvs = kvs.Add("status", "warn")
			} else {
				kvs = kvs.Add("status", "info")
			}

			for k, v := range i.mergedTags {
				kvs = kvs.AddTag(k, v)
			}

			pts = append(pts, point.NewPoint(measureuemtRedisBigKey, kvs, opts...))
		}
	}

	return pts
}

func (i *instance) setupHotBigKeyScanners() {
	var scanners []*hotbigkeyScanner
	// standalone and cluster mode only support master node
	if (i.mode == modeCluster || i.mode == modeStandalone) &&
		i.ipt.HotBigKeys.TargetRole != targetRoleMaster {
		l.Infof("cluster or standalone mode only support master node, skipped")
		return
	}

	switch i.ipt.HotBigKeys.TargetRole {
	case targetRoleReplica: // For targetRoleReplica, we only start scanner on master's replicas.
		if len(i.replicas) == 0 {
			l.Infof("no hot/big key scanner for instance %s: no replicas", i.addr)
			return
		}
		scanner := newHotBigKeyScanner(i.ipt.HotBigKeys)

		scanner.mode = i.mode
		scanner.role = "replica"
		scanner.cc = i.cc
		scanner.addr = i.addr
		scanner.replicaof = i.addr

		l.Infof("add hot/big key scanner on node %s: %s", i.addr, scanner)

		scanners = append(scanners, scanner)

	case targetRoleMaster: // For targetRoleAny, we only start scanner on master and ignore it's replicas.
		if i.cc == nil {
			l.Warnf("skip on invalid master(%s): rdb not set", i.addr)
			return
		}

		scanner := newHotBigKeyScanner(i.ipt.HotBigKeys)

		scanner.mode = i.mode
		scanner.role = "master"
		scanner.cc = i.cc
		scanner.addr = i.addr

		l.Infof("start hot/big key scanner on master node %s: %s", i.addr, scanner)
		scanners = append(scanners, scanner)
	}

	i.hbScanners = scanners
}

func (i *instance) randomScannerReplica(scanner *hotbigkeyScanner) {
	if len(i.replicas) == 0 {
		l.Debugf("no replicas available for mode %s, skip replica scanning", i.mode)
		return
	}
	switch i.ipt.HotBigKeys.TargetRole {
	case targetRoleReplica:
		// master-slave or sentinel mode, switch to replica
		if i.mode == modeMasterSlave || i.mode == modeSentinel {
			rand.Seed(time.Now().UnixNano())
			idx := rand.Intn(len(i.replicas)) // nolint:gosec // 这里使用弱随机数是安全的
			rep := i.replicas[idx]

			if rep.cc == nil {
				l.Warnf("skip on invalid replica: rdb not set", rep.addr)
				return
			}
			scanner.cc = rep.cc
			scanner.addr = rep.addr

			l.Infof("scanner switched to %s", scanner)
		}

	case targetRoleMaster: // pass
	}
}

func (i *instance) doStartCollectHotBigKeys(ctx context.Context, scanner *hotbigkeyScanner) {
	l.Infof("start collecting hot/big keys on %s", scanner.String())

	defer scanner.btick.Stop()
	defer scanner.htick.Stop()

	for {
		i.randomScannerReplica(scanner)
		i.collectHotBigKeys(ctx, scanner)
		scanner.flagHotkey = false
		scanner.flagBigkey = false

		select {
		case <-ctx.Done():
			l.Infof("hot/big key scan exited: %s", scanner.String())
			return

		case <-scanner.btick.C:
			scanner.flagBigkey = true
			select {
			case <-scanner.htick.C:
				scanner.flagHotkey = true
			default:
			}

		case <-scanner.htick.C:
			scanner.flagHotkey = true
			select {
			case <-scanner.btick.C:
				scanner.flagBigkey = true
			default:
			}
		}
	}
}

func (i *instance) collectHotBigKeys(ctx context.Context, scanner *hotbigkeyScanner) {
	collectStart := time.Now()

	if pts := i.doCollectHotBigKeys(ctx, scanner); len(pts) > 0 {
		if err := i.ipt.feeder.Feed(point.Logging, pts,
			dkio.WithCollectCost(time.Since(collectStart)),
			dkio.WithElection(i.ipt.Election),
			dkio.WithSource(dkio.FeedSource(inputName, "hot-big-keys")),
		); err != nil {
			l.Warnf("feed hot/big keys: %s, ignored", err)
		}
	}
}

func (i *instance) doCollectHotBigKeys(ctx context.Context, scanner *hotbigkeyScanner) (pts []*point.Point) {
	if i.ipt.HotBigKeys == nil {
		l.Debugf("skip hot/big key collecting")
		return nil
	}

	var dbs []int
	dbs = i.ipt.DBs
	if i.mode == modeCluster {
		dbs = []int{0}
	}

	for _, db := range dbs {
		l.Debugf("collect hot/big keys on %s db %d, hot: %v, big: %v",
			scanner, db, scanner.flagHotkey, scanner.flagBigkey)

		select {
		case <-ctx.Done():
			l.Infof("doCollectHotBigKeys exited: %s", scanner.String())
			return pts
		default:
		}

		if err := i.hotbigkeyScanDB(ctx, db, scanner); err != nil {
			l.Errorf("findBigHotKeys: %s, skipped", err.Error())
		} else {
			pts = append(pts, i.hotbigPoints(scanner)...)
		}

		scanner.reset() // reset for next db.
	}

	return pts
}

type hotkeyMeasurement struct{}

//nolint:lll
func (m *hotkeyMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   measureuemtRedisHotKey,
		Desc:   "Scan each keys in Redis DBs and find top-N(default top-10) keys that frequently used. Note: Redis's `maxmemory-policy` should configured as `allkeys-lfu` or `volatile-lfu`.",
		DescZh: "通过扫描 Redis DB 中每个 key 的访问频率，返回最高频（默认 top-10）的 key 信息。注意，需 Redis 的 `maxmemory-policy` 设置为 `allkeys-lfu` 或 `volatile-lfu`。",

		Cat: point.Logging,
		Fields: map[string]interface{}{
			"key_count": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Key count times.",
			},
			"keys_sampled": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Sampled keys in the key space.",
			},
			"status": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.EnumType,
				Unit:     inputs.NoUnit,
				Desc:     "`warn` or `info`",
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "Hostname."},
			"server":   &inputs.TagInfo{Desc: "Server addr."},
			"db_name":  &inputs.TagInfo{Desc: "DB name."},
			"key":      &inputs.TagInfo{Desc: "Key name."},
			"key_type": &inputs.TagInfo{Desc: "Key type(`string`/`hash`/`list`/`set`/`zset`)"},
		},
	}
}

type bigKeyMeasurement struct{}

//nolint:lll
func (m *bigKeyMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name:   measureuemtRedisBigKey,
		Cat:    point.Logging,
		Desc:   `Scan each keys in Redis DBs and find keys that larger than specific size.`,
		DescZh: "通过扫描 Redis DB 中每个 key 的大小，筛选出超过指定大小的 key 信息。",
		Fields: map[string]interface{}{
			"value_length": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeByte,
				Desc:     "Key length.",
			},
			"keys_sampled": &inputs.FieldInfo{
				DataType: inputs.Int,
				Type:     inputs.Gauge,
				Unit:     inputs.NCount,
				Desc:     "Sampled keys in the key space.",
			},
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.NoType,
				Unit:     inputs.NCount,
				Desc:     "Big key message details",
			},
			"status": &inputs.FieldInfo{
				DataType: inputs.String,
				Type:     inputs.EnumType,
				Unit:     inputs.NoUnit,
				Desc:     "`warn` or `info`",
			},
		},
		Tags: map[string]interface{}{
			"host":     &inputs.TagInfo{Desc: "Hostname."},
			"server":   &inputs.TagInfo{Desc: "Server addr."},
			"db_name":  &inputs.TagInfo{Desc: "DB name."},
			"key":      &inputs.TagInfo{Desc: "Key name."},
			"key_type": &inputs.TagInfo{Desc: "Key type(`string`/`hash`/`list`/`set`/`zset`)"},
		},
	}
}
