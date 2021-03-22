package io

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	defaultIO = NewIO()
)

func Start() error {
	l = logger.SLogger("io")

	defaultIO.DatawayHost = datakit.Cfg.MainCfg.DataWay.URL

	if datakit.Cfg.MainCfg.DataWay.Timeout != "" {
		du, err := time.ParseDuration(datakit.Cfg.MainCfg.DataWay.Timeout)
		if err != nil {
			l.Warnf("parse dataway timeout failed: %s, default 30s", err.Error())
		} else {
			defaultIO.HTTPTimeout = du
		}
	}

	if datakit.OutputFile != "" {
		defaultIO.OutputFile = datakit.OutputFile
	}

	if datakit.Cfg.MainCfg.StrictMode {
		defaultIO.StrictMode = true
	}

	defaultIO.FlushInterval = datakit.IntervalDuration

	datakit.WG.Add(1)
	go func() {
		defer datakit.WG.Done()
		defaultIO.startIO(true)
	}()

	return nil
}

type qstats struct {
	ch chan []*InputsStat
}

func GetStats() ([]*InputsStat, error) {
	q := &qstats{
		ch: make(chan []*InputsStat),
	}

	tick := time.NewTicker(time.Second * 3)
	defer tick.Stop()

	select {
	case defaultIO.qstatsCh <- q:
	case <-tick.C:
		return nil, fmt.Errorf("send stats request timeout")
	}

	select {
	case res := <-q.ch:
		return res, nil
	case <-tick.C:
		return nil, fmt.Errorf("get stats timeout")
	}
}

func ChanStat() string {
	l := len(defaultIO.in)
	c := cap(defaultIO.in)

	l2 := len(defaultIO.in2)
	c2 := cap(defaultIO.in2)
	return fmt.Sprintf("inputCh: %d/%d, highFreqInputCh: %d/%d", l, c, l2, c2)
}

func Feed(data []byte, name, category string, opt *Option) error {
	return defaultIO.doFeed(data, category, name, opt)
}
