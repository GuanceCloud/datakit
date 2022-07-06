// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"fmt"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sender"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
)

var defaultIO = &IO{
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

func SetDataway(dw dataway.DataWay) {
	defaultIO.dw = dw
}

func ConfigDefaultIO(c *IOConfig) {
	defaultIO.conf = c
}

func Start(sincfg []map[string]interface{}) error {
	log = logger.SLogger("io")

	log.Debugf("default io config: %v", defaultIO)

	defaultIO.in = make(chan *iodata, defaultIO.conf.FeedChanSize)
	defaultIO.in2 = make(chan *iodata, defaultIO.conf.HighFreqFeedChanSize)
	defaultIO.inLastErr = make(chan *lastError, datakit.CommonChanCap)

	defaultIO.inputstats = map[string]*InputsStat{}
	defaultIO.cache = map[string][]*Point{}
	defaultIO.dynamicCache = map[string][]*Point{}

	var writeFunc func(string, []sinkcommon.ISinkPoint) error

	if defaultIO.dw != nil {
		if dw, ok := defaultIO.dw.(sender.Writer); ok {
			writeFunc = dw.Write
		}
	}

	if err := sink.Init(sincfg, writeFunc); err != nil {
		log.Error("InitSink failed: %v", err)
		return err
	}

	defaultIO.StartIO(true)
	log.Debugf("io: %+#v", defaultIO)

	log.Debug("sink.Init succeeded")

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

	return fmt.Sprintf(`inputCh: %d/%d, highFreqInputCh: %d/%d`, l, c, l2, c2)
}

func DroppedTotal() int64 {
	return defaultIO.DroppedTotal()
}
