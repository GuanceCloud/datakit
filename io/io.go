package io

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"encoding/json"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

var (
	testAssert = false
	l          = logger.DefaultSLogger("io")

	highFreqCleanInterval = time.Millisecond * 500
)

type Option struct {
	CollectCost time.Duration
	HighFreq    bool

	HTTPHost    string
	PostTimeout time.Duration
}

type lastErr struct {
	from, err string
	ts        time.Time
}

type qstats struct {
	qid string
	ch  chan map[string]*InputsStat
}

type IO struct {
	DatawayHost   string
	HTTPTimeout   time.Duration
	MaxCacheCnt   int64
	OutputFile    string
	FlushInterval time.Duration

	httpCli *http.Client
	dw      *datakit.DataWayCfg

	in        chan *iodata
	in2       chan *iodata // high-freq chan
	inLastErr chan *lastErr

	inputstats map[string]*InputsStat
	qstatsCh   chan *qstats

	cache        map[string][]*Point
	dynamicCache []*iodata

	cacheCnt       int64
	fd             *os.File
	outputFileSize int64
	categoryURLs   map[string]string
}

func NewIO() *IO {
	return &IO{
		HTTPTimeout:   30 * time.Second,
		MaxCacheCnt:   1024,
		FlushInterval: time.Second * 10,

		in:        make(chan *iodata, 128),
		in2:       make(chan *iodata, 128*8),
		inLastErr: make(chan *lastErr, 128),

		inputstats: map[string]*InputsStat{},
		qstatsCh:   make(chan *qstats), // blocking

		cache:        map[string][]*Point{},
		dynamicCache: []*iodata{},
	}
}

const ( // categories
	MetricDeprecated = "/v1/write/metrics"
	Metric           = "/v1/write/metric"
	KeyEvent         = "/v1/write/keyevent"
	Object           = "/v1/write/object"
	Logging          = "/v1/write/logging"
	Tracing          = "/v1/write/tracing"
	Rum              = "/v1/write/rum"
	Security         = "/v1/write/security"
	Telegraf         = "/v1/write/telegraf"
	HeartBeat        = "/v1/write/heartbeat"

	minGZSize = 1024
)

type iodata struct {
	category, name string
	opt            *Option
	pts            []*Point
	url            string
	isProxy        bool
}

type InputsStat struct {
	//Name      string    `json:"name"`
	Category  string    `json:"category"`
	Frequency string    `json:"frequency,omitempty"`
	AvgSize   int64     `json:"avg_size"`
	Total     int64     `json:"total"`
	Count     int64     `json:"count"`
	First     time.Time `json:"first"`
	Last      time.Time `json:"last"`

	LastErr   string    `json:"last_error,omitempty"`
	LastErrTS time.Time `json:"last_error_ts,omitempty"`

	MaxCollectCost time.Duration `json:"max_collect_cost"`
	AvgCollectCost time.Duration `json:"avg_collect_cost"`

	totalCost time.Duration `json:"-"`
}

func TestOutput() {
	testAssert = true
}

func SetTest() {
	testAssert = true
}

func (x *IO) DoFeed(pts []*Point, category, name string, opt *Option) error {

	ch := x.in

	if opt != nil && opt.HighFreq {
		ch = x.in2
	}

	switch category {
	case MetricDeprecated:
	case Metric:
	case KeyEvent:
	case Object:
	case Logging:
	case Tracing:
	case Security:
	case Rum:
	default:
		return fmt.Errorf("invalid category `%s'", category)
	}

	l.Debugf("io feed %s", name)

	select {
	case ch <- &iodata{
		category: category,
		pts:      pts,
		name:     name,
		opt:      opt,
	}:
	case <-datakit.Exit.Wait():
		l.Warnf("%s/%s feed skipped on global exit", category, name)
	}

	return nil
}

func (x *IO) ioStop() {
	if x.fd != nil {
		if err := x.fd.Close(); err != nil {
			l.Error(err)
		}
	}
}

func (x *IO) updateLastErr(e *lastErr) {
	stat, ok := x.inputstats[e.from]
	if !ok {
		stat = &InputsStat{}
		x.inputstats[e.from] = stat
	}

	stat.LastErr = e.err
	stat.LastErrTS = e.ts
}

func (x *IO) updateStats(d *iodata) {
	now := time.Now()
	stat, ok := x.inputstats[d.name]

	if !ok {
		stat = &InputsStat{
			Total: int64(len(d.pts)),
			First: now,
		}
		x.inputstats[d.name] = stat
	}

	stat.Total += int64(len(d.pts))
	stat.Count++
	stat.Last = now
	stat.Category = d.category

	if (stat.Last.Unix() - stat.First.Unix()) > 0 {
		stat.Frequency = fmt.Sprintf("%.02f/min",
			float64(stat.Count)/(float64(stat.Last.Unix()-stat.First.Unix())/60))
	}
	stat.AvgSize = (stat.Total) / stat.Count

	if d.opt != nil {
		stat.totalCost += d.opt.CollectCost
		stat.AvgCollectCost = (stat.totalCost) / time.Duration(stat.Count)
		if d.opt.CollectCost > stat.MaxCollectCost {
			stat.MaxCollectCost = d.opt.CollectCost
		}
	}
}

func (x *IO) cacheData(d *iodata, tryClean bool) {

	if d == nil {
		l.Warn("get empty data, ignored")
		return
	}

	l.Debugf("get iodata(%d points) from %s|%s", len(d.pts), d.category, d.name)

	x.updateStats(d)

	if d.opt != nil && d.opt.HTTPHost != "" {
		x.dynamicCache = append(x.dynamicCache, d)
	} else {
		x.cache[d.category] = append(x.cache[d.category], d.pts...)
	}

	x.cacheCnt += int64(len(d.pts))

	if x.cacheCnt > x.MaxCacheCnt && tryClean {
		x.flushAll()
	}
}

func (x *IO) cleanHighFreqIOData() {

	if len(x.in2) > 0 {
		l.Debugf("clean %d cache on high-freq-chan", len(x.in2))
	}

	for {
		select {
		case d := <-x.in2: // eat all cached data
			x.cacheData(d, false)
		default:
			return
		}
	}
}

func (x *IO) init() error {
	if x.OutputFile != "" {
		f, err := os.OpenFile(datakit.OutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			l.Error(err)
			return err
		}

		x.fd = f
	}

	x.httpCli = &http.Client{
		Timeout: x.HTTPTimeout,
	}

	dw, err := datakit.ParseDataway(x.DatawayHost)
	if err != nil {
		return err
	}

	x.dw = dw
	x.categoryURLs = map[string]string{
		Metric:   x.dw.MetricURL(),
		KeyEvent: x.dw.KeyEventURL(),
		Object:   x.dw.ObjectURL(),
		Logging:  x.dw.LoggingURL(),
		Tracing:  x.dw.TracingURL(),
		Security: x.dw.SecurityURL(),
		Rum:      x.dw.RumURL(),
	}

	return nil
}

func (x *IO) StartIO(recoverable bool) {

	if err := x.init(); err != nil {
		return
	}

	defer x.ioStop()

	var f rtpanic.RecoverCallback

	f = func(trace []byte, _ error) {
		if recoverable {
			defer rtpanic.Recover(f, nil)
		}

		tick := time.NewTicker(x.FlushInterval)
		defer tick.Stop()

		highFreqRecvTicker := time.NewTicker(highFreqCleanInterval)
		defer highFreqRecvTicker.Stop()

		heartBeatTick := time.NewTicker(time.Second * 30)
		defer heartBeatTick.Stop()

		if trace != nil {
			l.Warnf("recover from %s", string(trace))
		}

		for {
			select {
			case d := <-x.in:
				x.cacheData(d, true)

			case e := <-x.inLastErr:
				x.updateLastErr(e)

			case q := <-x.qstatsCh:

				res := dumpStats(x.inputstats)
				select {
				case <-q.ch:
					l.Warnf("qid(%s) client canceled, ignored", q.qid)
				case q.ch <- res: // XXX: reference
					l.Debugf("qid(%s) response ok", q.qid)
				}

			case <-highFreqRecvTicker.C:
				x.cleanHighFreqIOData()

			case <-heartBeatTick.C:
				x.dkHeartbeat()

			case <-tick.C:
				l.Debugf("chan stat: %s", ChanStat())
				x.flushAll()

			case <-datakit.Exit.Wait():
				l.Info("io exit on exit")
				return
			}
		}
	}

	l.Info("starting...")
	f(nil, nil)
}

func dumpStats(is map[string]*InputsStat) (res map[string]*InputsStat) {
	res = map[string]*InputsStat{}
	for x, y := range is {
		res[x] = y
	}
	return
}

func (x *IO) dkHeartbeat() {
	body := map[string]interface{}{
		"dk_uuid":   datakit.Cfg.MainCfg.UUID,
		"heartbeat": time.Now().Unix(),
		"host":      datakit.Cfg.MainCfg.Hostname,
		"token":     datakit.Cfg.MainCfg.DataWay.GetToken(),
	}
	bodyByte, err := json.Marshal(body)
	if err != nil {
		l.Errorf("[error] heartbeat json marshal err:%s", err.Error())
		return
	}

	req, err := http.NewRequest("POST", datakit.Cfg.MainCfg.DataWay.HeartBeatURL(), bytes.NewBuffer(bodyByte))
	resp, err := x.httpCli.Do(req)
	if err != nil {
		l.Error(err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		l.Errorf("heart beat resp err: %+#v", resp)
	}
}

func (x *IO) flushAll() {
	x.flush()

	if x.cacheCnt > 0 {
		l.Warnf("post failed cache count: %d", x.cacheCnt)
	}

	if x.cacheCnt > x.MaxCacheCnt {
		l.Warnf("failed cache count reach max limit(%d), cleanning cache...", x.MaxCacheCnt)
		for k, _ := range x.cache {
			x.cache[k] = nil
		}
		x.cacheCnt = 0
	}
}

func (x *IO) flush() {

	if x.httpCli != nil {
		defer x.httpCli.CloseIdleConnections()
	}

	for k, v := range x.cache {

		if err := x.doFlush(v, x.categoryURLs[k]); err != nil {
			l.Errorf("post %d to %s failed", len(v), k)
			continue
		}

		if len(v) > 0 {
			x.cacheCnt -= int64(len(v))
			l.Debugf("clean %d cache on %s, remain: %d", len(v), k, x.cacheCnt)
			x.cache[k] = nil
		}
	}

	// flush dynamic cache: __not__ post to default dataway
	left := []*iodata{}
	for _, v := range x.dynamicCache {
		v.url = v.opt.HTTPHost
		if err := x.doFlush(v.pts, v.url); err != nil {
			l.Errorf("post %d to %s failed", len(v.pts), v.url)
			left = append(left, v)
			continue
		}

		l.Debugf("dynamic Cache %s", v.url)

		if len(v.pts) > 0 {
			x.cacheCnt -= int64(len(v.pts))
		}
	}

	if len(x.dynamicCache) > 0 {
		l.Debugf("clean %d dynamic cache, remain: %d", len(x.dynamicCache), len(left))
	}

	x.dynamicCache = left
}

func (x *IO) buildBody(pts []*Point) (body []byte, gzon bool, err error) {

	lines := []string{}
	for _, pt := range pts {
		lines = append(lines, pt.String())
	}

	raw := strings.Join(lines, "\n")
	if len(raw) > minGZSize && x.OutputFile == "" { // should not gzip on file output
		if body, err = datakit.GZipStr(raw); err != nil {
			l.Errorf("gz: %s", err.Error())
			return
		}
		gzon = true
	} else {
		body = []byte(raw)
	}

	return
}

func (x *IO) doFlush(pts []*Point, url string) error {

	if testAssert {
		return nil
	}

	if pts == nil {
		return nil
	}

	body, gz, err := x.buildBody(pts)
	if err != nil {
		return err
	}

	if x.OutputFile != "" {
		return x.fileOutput(body)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		l.Error(err)
		return err
	}

	if gz {
		req.Header.Set("Content-Encoding", "gzip")
	}

	// append datakit info
	req.Header.Set("X-Datakit-Info",
		fmt.Sprintf("%s; %s", datakit.Cfg.MainCfg.Hostname, git.Version))

	postbeg := time.Now()

	resp, err := x.httpCli.Do(req)
	if err != nil {
		l.Errorf("request url %s failed: %s", url, err)
		return err
	}

	defer resp.Body.Close()
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Error(err)
		return err
	}

	switch resp.StatusCode / 100 {
	case 2:
		l.Debugf("post %d bytes to %s ok(gz: %v), cost %v, response: %s",
			len(body), url, gz, time.Since(postbeg), string(respbody))
		return nil

	case 4:
		l.Debugf("post %d bytes to %s failed(HTTP: %s): %s, cost %v, data dropped",
			len(body), url, resp.StatusCode, string(respbody), time.Since(postbeg))
		return nil

	case 5:
		l.Errorf("post %d bytes to %s failed(HTTP: %s): %s, cost %v",
			len(body), url, resp.Status, string(respbody), time.Since(postbeg))
		return fmt.Errorf("dataway internal error")
	}

	return nil
}

func (x *IO) fileOutput(body []byte) error {

	if _, err := x.fd.Write(append(body, '\n')); err != nil {
		l.Error(err)
		return err
	}

	x.outputFileSize += int64(len(body))
	if x.outputFileSize > 4*1024*1024 {
		if err := x.fd.Truncate(0); err != nil {
			l.Error(err)
			return err
		}
		x.outputFileSize = 0
	}

	return nil
}
