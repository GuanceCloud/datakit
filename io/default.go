package io

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

var (
	defaultMaxCacheCnt = int64(1024)

	defaultIO     = NewIO(defaultMaxCacheCnt)
	outputFile    = ""
	flushInterval = 10 * time.Second
	extraTags     = map[string]string{}
)

func SetOutputFile(f string) {
	defaultIO.OutputFile = f
}

func SetFlushInterval(x time.Duration) {
	defaultIO.FlushInterval = x
}

func SetDataWay(dw *dataway.DataWayCfg) {
	defaultIO.dw = dw
}

func SetExtraTags(x map[string]string) {
	extraTags = x
}

func Start() error {
	l = logger.SLogger("io")

	datakit.WG.Add(1)
	go func() {
		defer datakit.WG.Done()
		defaultIO.StartIO(true)
	}()

	l.Debugf("io: %+#v", defaultIO)

	return nil
}

func GetStats(timeout time.Duration) (map[string]*InputsStat, error) {
	q := &qstats{
		qid: cliutils.XID("statqid_"),
		ch:  make(chan map[string]*InputsStat),
	}

	defer close(q.ch)

	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	tick := time.NewTicker(timeout)
	defer tick.Stop()

	select {
	case defaultIO.qstatsCh <- q:
	case <-tick.C:
		return nil, fmt.Errorf("default IO busy(qid: %s, %v)", q.qid, timeout)
	}

	select {
	case res := <-q.ch:
		return res, nil
	case <-tick.C:
		return nil, fmt.Errorf("default IO response timeout(qid: %s, %v)", q.qid, timeout)
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

	return defaultIO.DoFeed(pts, category, name, opt)
}

func FeedLastError(inputName string, err string) error {
	select {
	case defaultIO.inLastErr <- &lastErr{
		from: inputName,
		err:  err,
		ts:   time.Now(),
	}:
	case <-datakit.Exit.Wait():
		l.Warnf("%s feed last error skipped on global exit", inputName)
	}
	return nil
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
			ExtraTags: extraTags,
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

	return defaultIO.DoFeed(x, category, name, nil)
}

// Deprecated
func HighFreqFeedEx(name, category, metric string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time) error {

	var ts time.Time
	if len(t) > 0 {
		ts = t[0]
	} else {
		ts = time.Now().UTC()
	}

	pt, err := lp.MakeLineProtoPoint(metric, tags, fields,
		&lp.Option{
			ExtraTags: extraTags,
			Strict:    true,
			Time:      ts,
			Precision: "n"})
	if err != nil {
		return err
	}

	return defaultIO.DoFeed([]*Point{&Point{pt}}, category, name, &Option{HighFreq: true})
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

	pt, err := lp.MakeLineProtoPoint(metric, tags, fields,
		&lp.Option{
			ExtraTags: extraTags,
			Strict:    true,
			Time:      ts,
			Precision: "n"})
	if err != nil {
		return err
	}

	return defaultIO.DoFeed([]*Point{&Point{pt}}, category, name, nil)
}
