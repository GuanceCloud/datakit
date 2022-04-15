// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink"
)

var (
	extraTags = map[string]string{}
	defaultIO = &IO{
		FeedChanSize:              1024,
		HighFreqFeedChanSize:      2048,
		MaxCacheCount:             1024,
		CacheDumpThreshold:        512,
		MaxDynamicCacheCount:      1024,
		DynamicCacheDumpThreshold: 512,
		FlushInterval:             10 * time.Second,
	}
)

type IOOption func(io *IO)

func SetMaxCacheCount(max int64) IOOption {
	return func(io *IO) {
		io.MaxCacheCount = max
	}
}

func SetCacheDumpThreshold(threshold int64) IOOption {
	return func(io *IO) {
		io.CacheDumpThreshold = threshold
	}
}

func SetMaxDynamicCacheCount(max int64) IOOption {
	return func(io *IO) {
		io.MaxDynamicCacheCount = max
	}
}

func SetDynamicCacheDumpThreshold(threshold int64) IOOption {
	return func(io *IO) {
		io.DynamicCacheDumpThreshold = threshold
	}
}

func SetFlushInterval(s string) IOOption {
	return func(io *IO) {
		if len(s) == 0 {
			io.FlushInterval = 10 * time.Second
		} else {
			if d, err := time.ParseDuration(s); err != nil {
				log.Errorf("parse io flush interval failed, %s", err.Error())
				io.FlushInterval = 10 * time.Second
			} else {
				io.FlushInterval = d
			}
		}
	}
}

func SetOutputFile(output string) IOOption {
	return func(io *IO) {
		io.OutputFile = output
	}
}

func SetOutputFileInput(outputFileInputs []string) IOOption {
	return func(io *IO) {
		io.OutputFileInput = outputFileInputs
	}
}

func SetDataway(dw *dataway.DataWayCfg) IOOption {
	return func(io *IO) {
		io.dw = dw
	}
}

func SetFeedChanSize(size int) IOOption {
	return func(io *IO) {
		io.FeedChanSize = size
	}
}

func SetHighFreqFeedChanSize(size int) IOOption {
	return func(io *IO) {
		io.HighFreqFeedChanSize = size
	}
}

func SetEnableCache(enable bool) IOOption {
	return func(io *IO) {
		io.EnableCache = enable
	}
}

func ConfigDefaultIO(opts ...IOOption) {
	for _, opt := range opts {
		opt(defaultIO)
	}
}

func Start(sincfg []map[string]interface{}) error {
	log = logger.SLogger("io")

	log.Debugf("default io config: %v", defaultIO)

	defaultIO.in = make(chan *iodata, defaultIO.FeedChanSize)
	defaultIO.in2 = make(chan *iodata, defaultIO.HighFreqFeedChanSize)
	defaultIO.inLastErr = make(chan *lastError, 128)
	defaultIO.inputstats = map[string]*InputsStat{}
	defaultIO.cache = map[string][]*Point{}
	defaultIO.dynamicCache = map[string][]*Point{}

	if err := sink.Init(sincfg, defaultIO.dw.Write); err != nil {
		log.Errorf("InitSink failed: %v", err)
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
