// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"context"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_merge(t *T.T) {
	conf := defaultHotBitKeyConf()
	conf.TopN = 3

	t.Run("hot", func(t *T.T) {
		scanner := newHotBigKeyScanner(conf)
		scanner.mergeHotKeys(&keyInfo{key: "key1", freq: 6})
		scanner.mergeHotKeys(&keyInfo{key: "key2", freq: 0})
		scanner.mergeHotKeys(&keyInfo{key: "key3", freq: 3})
		scanner.mergeHotKeys(&keyInfo{key: "key4", freq: 2})
		scanner.mergeHotKeys(&keyInfo{key: "key5", freq: 2})

		assert.Equal(t, int64(2), scanner.hot[2].freq)
		assert.Equal(t, "key4", scanner.hot[2].key) // key5 not inserted
		assert.Equal(t, 1, scanner.hot[2].cnt)      // extra same freq key counts
	})

	t.Run("big", func(t *T.T) {
		scanner := newHotBigKeyScanner(conf)
		scanner.bigString = scanner.mergeBigKeys(scanner.bigString, &keyInfo{key: "key1", vlen: 6})
		scanner.bigString = scanner.mergeBigKeys(scanner.bigString, &keyInfo{key: "key2", vlen: 0})
		scanner.bigString = scanner.mergeBigKeys(scanner.bigString, &keyInfo{key: "key3", vlen: 3})
		scanner.bigString = scanner.mergeBigKeys(scanner.bigString, &keyInfo{key: "key4", vlen: 2})
		scanner.bigString = scanner.mergeBigKeys(scanner.bigString, &keyInfo{key: "key5", vlen: 2})

		assert.Equal(t, int64(2), scanner.bigString[2].vlen)
		assert.Equal(t, "key4", scanner.bigString[2].key)
		assert.Equal(t, 1, scanner.bigString[2].cnt)
	})
}

func Test_hbscanners(t *T.T) {
	t.Run("scan-on-master-slave", func(t *T.T) {
		t.Skip("skip on real redis instance")
		ipt := defaultInput()
		ipt.DBs = []int{0, 1}
		ipt.MasterSlave = &redisMasterSlave{
			Hosts: []string{
				"centos.orb.local:26380",
				"centos.orb.local:26381",
				"centos.orb.local:26382",
			},
			Sentinel: &redisSentinel{
				Password:   "123456abc",
				MasterName: "mymaster",
			},
		}

		ipt.HotBigKeys.TargetRole = "replica"
		ipt.HotBigKeys.BatchSize = 1000
		ipt.HotBigKeys.ScanSleep = 0
		ipt.Password = "abc123456"

		assert.NoError(t, ipt.initCfg())
		require.Len(t, ipt.instances, 1)

		inst := ipt.instances[0]
		inst.setupHotBigKeyScanners()

		t.Logf("scanners: %+#v", inst.hbScanners)

		ctx := context.Background()
		for _, s := range inst.hbScanners {
			pts := inst.doCollectHotBigKeys(ctx, s)
			for _, pt := range pts {
				t.Logf("%s", pt.Pretty())
				assert.NotNil(t, pt.Get("host"))
				assert.NotNil(t, pt.Get("server"))
			}
		}
	})

	t.Run("scan-on-cluster", func(t *T.T) {
		t.Skip("skip on real redis instance")
		ipt := defaultInput()
		ipt.DBs = []int{0, 1}
		ipt.Cluster = &redisCluster{Hosts: []string{
			"centos.orb.local:7001",
		}}
		ipt.HotBigKeys.TargetRole = "any"
		ipt.HotBigKeys.BatchSize = 1000
		ipt.HotBigKeys.ScanSleep = 0
		ipt.Password = "abc123456"

		assert.NoError(t, ipt.initCfg())

		require.Len(t, ipt.instances, 3)

		ctx := context.Background()
		for _, inst := range ipt.instances {
			inst.setupHotBigKeyScanners()

			for _, s := range inst.hbScanners {
				t.Logf("scanner: %s", s.String())
			}

			for _, s := range inst.hbScanners {
				pts := inst.doCollectHotBigKeys(ctx, s)
				for _, pt := range pts {
					t.Logf("%s", pt.Pretty())
					assert.NotNil(t, pt.Get("host"))
					assert.NotNil(t, pt.Get("server"))
				}
			}
		}
	})
}

func Test_hotbigkey(t *T.T) {
	t.Run("big-hot-keys", func(t *T.T) {
		t.Skip("skip-real-redis-host-test")
		reg := prometheus.NewRegistry()
		reg.MustRegister(p8sMetrics()...)
		defer resetP8sMetrics()

		redisServer := "centos.orb.local:7001"
		rdb := redis.NewClient(&redis.Options{
			Addr:     redisServer,
			Password: "abc123456",
		})

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		_, err := rdb.Ping(ctx).Result()
		assert.NoError(t, err)
		conf := defaultHotBitKeyConf()

		conf.BigkeyThreshouldLen = 0
		conf.BigkeyThreshouldMem = 0
		conf.ScanSleep = 0
		conf.BigkeyInterval = 0
		conf.HotkeyInterval = 0

		scanner := newHotBigKeyScanner(conf) // top n hot keys
		scanner.cc = &colclient{rdb}
		scanner.flagHotkey = true
		scanner.flagBigkey = true
		scanner.setup()

		inst := newInstance()
		assert.NoError(t, inst.hotbigkeyScanDB(ctx, 0, scanner))

		if scanner != nil {
			for i, hk := range scanner.hot {
				t.Logf("[HOT] %d. key: %s(%s), freq: %d, cnt: %d", i+1, hk.key, hk.keyt, hk.freq, hk.cnt)
			}

			for i, bk := range scanner.bigString {
				t.Logf("[BIG str] %d. key: %s, len: %d, cnt: %d", i+1, bk.key, bk.vlen, bk.cnt)
			}

			for i, bk := range scanner.bigList {
				t.Logf("[BIG list] %d. key: %s, len: %d, cnt: %d", i+1, bk.key, bk.vlen, bk.cnt)
			}

			for i, bk := range scanner.bigSet {
				t.Logf("[BIG set] %d. key: %s, len: %d, cnt: %d", i+1, bk.key, bk.vlen, bk.cnt)
			}

			for i, bk := range scanner.bigZset {
				t.Logf("[BIG zset] %d. key: %s, len: %d, cnt: %d", i+1, bk.key, bk.vlen, bk.cnt)
			}

			for i, bk := range scanner.bigHash {
				t.Logf("[BIG hash] %d. key: %s, len: %d, cnt: %d", i+1, bk.key, bk.vlen, bk.cnt)
			}

			for i, bk := range scanner.bigMem {
				t.Logf(
					"[BIG mem] %d. key: %s(%s), len: %d, mem: %d, cnt: %d",
					i+1,
					bk.key,
					bk.keyt,
					bk.vlen,
					bk.vmem,
					bk.cnt,
				)
			}
		}

		mfs, err := reg.Gather()
		assert.NoError(t, err)

		t.Logf("metrics: \n%s", metrics.MetricFamily2Text(mfs))
	})
}

func Test_hotbigPoints(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		t.Skip("skip on fake redis server")
		redisServer := "facked.server"
		ipt := defaultInput()
		ipt.Host = redisServer
		ipt.Port = 7001

		assert.NoError(t, ipt.initCfg())
		require.Len(t, ipt.instances, 1)

		inst := ipt.instances[0]

		t.Logf("inst: %s", inst)

		ipt.HotBigKeys.BigkeyThreshouldLen = 0
		ipt.HotBigKeys.BigkeyThreshouldMem = 0
		ipt.HotBigKeys.ScanSleep = 0
		ipt.HotBigKeys.BigkeyInterval = 0
		ipt.HotBigKeys.HotkeyInterval = 0

		scanner := newHotBigKeyScanner(ipt.HotBigKeys) // top n hot keys

		scanner.db = 0
		scanner.sampled = 100
		scanner.flagHotkey = true
		scanner.flagBigkey = true
		scanner.addr = redisServer

		var (
			bigmem = int64(1 << 20)
			biglen = int64(1 << 10)
			freq   = int64(3)
		)

		keyname := "all-the-same"

		scanner.hot = append(scanner.hot, &keyInfo{key: keyname, keyt: "string", cnt: 1, freq: freq})
		scanner.bigMem = append(scanner.bigMem, &keyInfo{key: keyname, keyt: "string", cnt: 1, vmem: bigmem})
		scanner.bigString = append(scanner.bigString, &keyInfo{key: keyname, keyt: "string", cnt: 1, vlen: biglen})
		scanner.bigList = append(scanner.bigList, &keyInfo{key: keyname, keyt: "list", cnt: 1, vlen: biglen})
		scanner.bigHash = append(scanner.bigHash, &keyInfo{key: keyname, keyt: "hash", cnt: 1, vlen: biglen})
		scanner.bigSet = append(scanner.bigSet, &keyInfo{key: keyname, keyt: "set", cnt: 1, vlen: biglen})
		scanner.bigZset = append(scanner.bigZset, &keyInfo{key: keyname, keyt: "zset", cnt: 1, vlen: biglen})

		// build points
		pts := inst.hotbigPoints(scanner)

		for _, pt := range pts {
			t.Logf("%s", pt.Pretty())

			switch pt.Name() {
			case "redis_hotkey":
				require.NotNil(t, pt.Get("key_count"))
			case "redis_bigkey":
				require.True(t, nil != pt.Get("value_length") || nil != pt.Get("value_mem"))
			}

			require.Equal(t, "info", pt.Get("status"))
			require.Equal(t, int64(100), pt.Get("keys_sampled"))
			require.Equal(t, "db0", pt.Get("db_name"))
			require.Equal(t, keyname, pt.Get("key"))

			assert.NotNil(t, pt.Get("host"))
			assert.NotNil(t, pt.Get("server"))

			keyt := pt.Get("key_type")
			require.Truef(t, "string" == keyt ||
				"list" == keyt ||
				"hash" == keyt ||
				"set" == keyt ||
				"zset" == keyt, "key type: %s", keyt)
		}

		// set > 0
		scanner.BigkeyThreshouldLen = 1
		scanner.BigkeyThreshouldMem = 1
		pts = inst.hotbigPoints(scanner)
		for _, pt := range pts {
			if pt.Name() == "redis_bigkey" {
				require.Equal(t, "warn", pt.Get("status"))
			}

			assert.NotNil(t, pt.Get("host"))
			assert.NotNil(t, pt.Get("server"))
		}
	})
}
