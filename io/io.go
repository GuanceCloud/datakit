package io

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"reflect"
	"time"

	influxm "github.com/influxdata/influxdb1-client/models"
	ifxcli "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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

type IO struct {
	DatawayHost   string
	HTTPTimeout   time.Duration
	MaxCacheCnt   int64
	OutputFile    string
	StrictMode    bool
	FlushInterval time.Duration

	httpCli *http.Client
	dw      *datakit.DataWayCfg

	in  chan *iodata
	in2 chan *iodata // high-freq chan

	inputstats map[string]*InputsStat
	qstatsCh   chan *qstats

	cache        map[string][][]byte
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

		in:  make(chan *iodata, 128),
		in2: make(chan *iodata, 128*8),

		inputstats: map[string]*InputsStat{},
		qstatsCh:   make(chan *qstats), // blocking

		cache:        map[string][][]byte{},
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

	minGZSize = 1024
)

type iodata struct {
	category, name string
	opt            *Option
	data           []byte // line-protocol or json or others
	url            string
	isProxy        bool
}

type InputsStat struct {
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Frequency string    `json:"frequency,omitempty"`
	AvgSize   int64     `json:"avg_size"`
	Total     int64     `json:"total"`
	Count     int64     `json:"count"`
	First     time.Time `json:"first"`
	Last      time.Time `json:"last"`

	totalCost time.Duration `json:"-"`

	AvgCollectCost time.Duration `json:"avg_collect_cost"`
}

func TestOutput() {
	testAssert = true
}

func SetTest() {
	testAssert = true
}

func (x *IO) doFeed(data []byte, category, name string, opt *Option) error {

	switch category {
	case Metric, KeyEvent, Object, Logging, Tracing:
		// metric line check
		if err := x.checkMetric(data); err != nil {
			return fmt.Errorf("invalid line protocol data %v", err)
		}
	case Rum: // do not check RUM data structure, too complecated

	default:
		return fmt.Errorf("invalid category %s", category)
	}

	ch := x.in

	if opt != nil && opt.HighFreq {
		ch = x.in2
	}

	if opt == nil {
		select {
		case ch <- &iodata{
			category: category,
			data:     data,
			name:     name,
			opt:      opt,
		}:
		case <-datakit.Exit.Wait():
			l.Warnf("%s/%s feed skipped on global exit", category, name)
		}
	}

	return nil
}

func (x *IO) checkMetric(data []byte) error {
	if !x.StrictMode {
		return nil
	}

	_, err := influxm.ParsePointsWithPrecision(data, time.Now().UTC(), "n")
	if err != nil {
		l.Errorf("[error] : %s", err.Error())
	}

	return err
}

func MakeMetric(name string, tags map[string]string,
	fields map[string]interface{}, t ...time.Time) ([]byte, error) {

	var tm time.Time
	if len(t) > 0 {
		tm = t[0]
	} else {
		tm = time.Now().UTC()
	}

	if len(datakit.Cfg.MainCfg.GlobalTags) > 0 {
		if tags == nil {
			tags = map[string]string{}
		}

		for k, v := range datakit.Cfg.MainCfg.GlobalTags {
			if _, ok := tags[k]; !ok { // do not overwrite exists tags
				tags[k] = v
			}
		}
	}

	for k, v := range tags { // remove any suffix `\` in all tag values
		tags[k] = datakit.TrimSuffixAll(v, `\`)
	}

	for k, v := range fields { // convert uint to int
		switch v.(type) {
		case uint64:
			if v.(uint64) > uint64(math.MaxInt64) {
				l.Warnf("on input `%s', filed %s, get uint64 %d > MaxInt64(%d), dropped", name, k, v.(uint64), uint64(math.MaxInt64))
				delete(fields, k)
			} else { // convert uint64 -> int64
				fields[k] = int64(v.(uint64))
			}
		case uint32, uint16, uint8,
			int, int8, int16, int32, int64,
			bool,
			string,
			float32, float64:
		default:
			l.Errorf("invalid filed type `%s', from `%s', on filed `%s', got value `%+#v'",
				reflect.TypeOf(v).String(), name, k, fields[k])
			return nil, fmt.Errorf("invalid field type")
		}
	}

	pt, err := ifxcli.NewPoint(name, tags, fields, tm)
	if err != nil {
		return nil, err
	}
	return []byte(pt.String()), nil
}

func (x *IO) ioStop() {
	if x.fd != nil {
		if err := x.fd.Close(); err != nil {
			l.Error(err)
		}
	}
}

func (x *IO) updateStats(d *iodata) {
	now := time.Now()
	stat, ok := x.inputstats[d.name]

	if !ok {
		stat := &InputsStat{
			Name:     d.name,
			Category: d.category,
			Total:    int64(len(d.data)),
			First:    now,
			Count:    1,
			Last:     now,
		}

		if d.opt != nil {
			stat.totalCost = d.opt.CollectCost
			stat.AvgCollectCost = d.opt.CollectCost
		}
		x.inputstats[d.name] = stat
	} else {
		stat.Total += int64(len(d.data))
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
		}
	}
}

func (x *IO) cacheData(d *iodata, tryClean bool) {

	if d == nil {
		l.Warn("get empty data, ignored")
		return
	}

	l.Debugf("get iodata(%d bytes) from %s|%s", len(d.data), d.category, d.name)

	x.updateStats(d)

	if d.opt.HTTPHost != "" {
		x.dynamicCache = append(x.dynamicCache, d)
	} else {
		x.cache[d.category] = append(x.cache[d.category], d.data)
	}

	x.cacheCnt++

	if x.cacheCnt > x.MaxCacheCnt && tryClean {
		x.flushAll()
	}
}

func (x *IO) cleanHighFreqIOData() {
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
		Rum:      x.dw.RumURL(),
	}

	return nil
}

func (x *IO) startIO(recoverable bool) {

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

		if trace != nil {
			l.Warnf("recover from %s", string(trace))
		}

		for {
			select {
			case d := <-x.in:
				x.cacheData(d, true)

			case q := <-x.qstatsCh:
				statRes := []*InputsStat{}
				for _, v := range x.inputstats {
					statRes = append(statRes, v)
				}
				select {
				case q.ch <- statRes: // maybe blocking(i.e., client canceled)
				default:
					l.Warn("client canceled")
					// pass
				}

			case <-highFreqRecvTicker.C:
				x.cleanHighFreqIOData()

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
			l.Debugf("clean %d/%d cache on %s", len(v), x.cacheCnt, k)
			x.cache[k] = nil
		}
		l.Debugf("clean %d/%d cache on %s", len(v), x.cacheCnt, k)
	}

	// flush dynamic cache: __not__ post to default dataway
	left := []*iodata{}
	for _, v := range x.dynamicCache {
		if err := x.doFlush([][]byte{v.data}, v.url); err != nil {
			l.Errorf("post %d to %s failed", len(v.data), v.url)
			left = append(left, v)
			continue
		}

		if len(v.data) > 0 {
			x.cacheCnt -= int64(len(v.data))
		}
	}

	l.Debugf("clean %d/%d dynamic cache", len(x.dynamicCache), len(left))

	x.dynamicCache = left
}

func buildBody(bodies [][]byte) (body []byte, gzon bool, err error) {
	body = bytes.Join(bodies, []byte("\n"))
	if len(body) > minGZSize && datakit.OutputFile == "" { // should not gzip on file output
		if body, err = datakit.GZip(body); err != nil {
			l.Errorf("gz: %s", err.Error())
			return
		}
		gzon = true
	}

	return
}

func (x *IO) doFlush(bodies [][]byte, url string) error {

	if testAssert {
		return nil
	}

	if bodies == nil {
		return nil
	}

	body, gz, err := buildBody(bodies)
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

	postbeg := time.Now()

	resp, err := x.httpCli.Do(req)
	if err != nil {
		l.Error(err)
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
		l.Debugf("post %d to %s ok(gz: %v), cost %v, response: %s",
			len(body), url, gz, time.Since(postbeg), string(respbody))
		return nil

	case 4:
		l.Debugf("post %d to %s failed(HTTP: %s): %s, cost %v, data dropped",
			len(body), url, resp.StatusCode, string(respbody), time.Since(postbeg))
		return nil

	case 5:
		l.Errorf("post %d to %s failed(HTTP: %s): %s, cost %v",
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
