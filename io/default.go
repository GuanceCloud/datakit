package io

import (
	"fmt"
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
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

	l.Debugf("io: %+#v", defaultIO)

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

func Feed(name, category string, pts []*Point, opt *Option) error {
	if len(pts) == 0 {
		return fmt.Errorf("no points")
	}

	return defaultIO.doFeed(pts, category, name, opt)
}

func MakePoint(name string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time) (*Point, error) {

	var ts time.Time
	if len(t) > 0 {
		ts = t[0]
	} else {
		ts = time.Now().UTC()
	}

	p, err := lp.MakeLineProtoPoint(name, tags, fields,
		&lp.Option{
			ExtraTags: datakit.Cfg.MainCfg.GlobalTags,
			Strict:    true,
			Time:      ts,
			Precision: "n"})
	if err != nil {
		return nil, err
	}

	return &Point{Point: p}, nil
}

// Deprecated
func MakeMetric(name string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time) ([]byte, error) {
	p, err := MakePoint(name, tags, fields, t...)
	if err != nil {
		return nil, err
	}

	return []byte(p.Point.String()), nil
}

// Deprecated
func NamedFeed(data []byte, category, name string) error {
	pts, err := lp.ParsePoints(data, nil)
	if err != nil {
		return err
	}

	x := []*Point{}
	for _, pt := range pts {
		x = append(x, &Point{Point: pt})
	}

	return defaultIO.doFeed(x, category, name, nil)
}

// Deprecated
func NamedFeedEx(name, category, metric string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time) error {

	var ts time.Time
	if len(t) > 0 {
		ts = t[0]
	} else {
		ts = time.Now().UTC()
	}

	pt, err := lp.MakeLineProtoPoint(name, tags, fields,
		&lp.Option{
			ExtraTags: datakit.Cfg.MainCfg.GlobalTags,
			Strict:    true,
			Time:      ts,
			Precision: "n"})
	if err != nil {
		return err
	}

	return defaultIO.doFeed([]*Point{&Point{pt}}, category, name, nil)
}
