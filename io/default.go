// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"os"
	"path/filepath"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/convertutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sender"
)

var defaultIO = getDefault()

func getDefault() *IO {
	return &IO{
		conf: &IOConfig{
			FeedChanSize:         128,
			MaxCacheCount:        64,
			MaxDynamicCacheCount: 128,

			FlushInterval: "10s",
		},
	}
}

func SetDataway(dw dataway.DataWay) {
	defaultIO.dw = dw
}

func ConfigDefaultIO(c *IOConfig) {
	defaultIO.conf = c
}

func (x *IO) init() error {
	if x.conf.OutputFile != "" {
		f, err := os.OpenFile(x.conf.OutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644) //nolint:gosec
		if err != nil {
			return err
		}

		x.fd = f
	}

	x.inLastErr = make(chan *lastError, datakit.CommonChanCap)
	x.inputstats = map[string]*InputsStat{}
	x.chans = map[string]chan *iodata{}
	x.fcs = make(map[string]*failCache)
	for _, c := range []string{
		datakit.Metric,
		datakit.Network,
		datakit.KeyEvent,
		datakit.Object,
		datakit.CustomObject,
		datakit.Logging,
		datakit.Tracing,
		datakit.RUM,
		datakit.Security,
		datakit.Profiling,
		datakit.DynamicDatawayCategory,
	} {
		x.chans[c] = make(chan *iodata, x.conf.FeedChanSize)

		if x.conf.EnableCache && c != datakit.DynamicDatawayCategory {
			cg, err := convertutil.GetMapCategoryFullToShort(c)
			if err != nil {
				return err
			}
			fc, err := initFailCache(filepath.Join(datakit.CacheDir, cg), int64(x.conf.CacheSizeGB*1024*1024*1024))
			if err != nil {
				return err
			}
			x.fcs[c] = fc
		}
	}

	du, err := time.ParseDuration(x.conf.FlushInterval)
	if err != nil {
		log.Warnf("time.ParseDuration: %s, ignored", err)
		du = time.Second * 10
	}

	x.flushInterval = du

	if sender, err := sender.NewSender(
		&sender.Option{
			Cache:              x.conf.EnableCache,
			FlushCacheInterval: du,
			ErrorCallback:      nil,
		}); err != nil {
		log.Errorf("init sender error: %s", err.Error())
	} else {
		x.sender = sender
	}

	return nil
}

func Start() error {
	log = logger.SLogger("io")
	log.Debugf("default io config: %v", defaultIO)

	if err := defaultIO.init(); err != nil {
		log.Errorf("io init failed: %s", err)
		return err
	}

	defaultIO.StartIO(true)

	return nil
}
