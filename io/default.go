package io

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cache"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

var (
	extraTags = map[string]string{}
	defaultIO = &IO{
		conf: &IOConfig{
			FeedChanSize:              1024,
			HighFreqFeedChanSize:      2048,
			MaxCacheCount:             1024,
			CacheDumpThreshold:        512,
			MaxDynamicCacheCount:      1024,
			DynamicCacheDumpThreshold: 512,
			FlushInterval:             "10s",
		},
	}
)

func SetDataway(dw *dataway.DataWayCfg) {
	defaultIO.dw = dw
}

func ConfigDefaultIO(c *IOConfig) {
	defaultIO.conf = c
}

func Start() error {
	log = logger.SLogger("io")

	log.Debugf("default io config: %v", defaultIO)

	defaultIO.in = make(chan *iodata, defaultIO.conf.FeedChanSize)
	defaultIO.in2 = make(chan *iodata, defaultIO.conf.HighFreqFeedChanSize)
	defaultIO.inputstats = map[string]*InputsStat{}
	defaultIO.cache = map[string][]*Point{}
	defaultIO.dynamicCache = map[string][]*Point{}

	defaultIO.StartIO(true)

	if defaultIO.conf.EnableCache {
		if err := cache.Initialize(datakit.CacheDir, nil); err != nil {
			log.Warn("initialized cache: %s, ignored", err)
		} else { //nolint
			if err := cache.CreateBucketIfNotExists(cacheBucket); err != nil {
				log.Warn("create bucket: %s", err)
			}
		}
	}

	log.Debugf("io: %+#v", defaultIO)

	return nil
}

func GetStats() (map[string]*InputsStat, error) {
	defaultIO.lock.RLock()
	defer defaultIO.lock.RUnlock()

	return dumpStats(defaultIO.inputstats), nil
}

func GetIoStats() IoStat {
	stats := IoStat{
		SentBytes: defaultIO.SentBytes,
	}
	return stats
}

func ChanStat() string {
	l := len(defaultIO.in)
	c := cap(defaultIO.in)

	l2 := len(defaultIO.in2)
	c2 := cap(defaultIO.in2)
	return fmt.Sprintf("inputCh: %d/%d, highFreqInputCh: %d/%d", l, c, l2, c2)
}

func DroppedTotal() int64 {
	return defaultIO.DroppedTotal()
}
