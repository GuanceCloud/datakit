package io

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	influxm "github.com/influxdata/influxdb1-client/models"
	ifxcli "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	l          = logger.DefaultSLogger("io")
	testAssert = false
	httpCli    *http.Client
	baseURL    string

	inputCh    = make(chan *iodata, datakit.CommonChanCap)
	inputstats = map[string]*InputsStat{}

	qstatsCh = make(chan *qstats)

	cache = map[string][][]byte{
		MetricDeprecated: nil,
		Metric:           nil,
		KeyEvent:         nil,
		Object:           nil,
		Logging:          nil,
		Tracing:          nil,
		Rum:              nil,
	}

	curCacheCnt = 0

	categoryURLs map[string]string

	outputFile     *os.File
	outputFileSize int64
)

const ( // categories

	maxCacheCnt      = 128
	MaxPostFailCache = 1024

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
}

type InputsStat struct {
	Name     string    `json:"name"`
	Category string    `json:"category"`
	Total    int64     `json:"total"`
	Count    int64     `json:"count"`
	First    time.Time `json:"first"`
	Last     time.Time `json:"last"`

	totalCost time.Duration `json:"-"`

	AvgCollectCost time.Duration `json:"avg_collect_cost"`
}

type qstats struct {
	ch chan []*InputsStat
}

func TestOutput() {
	testAssert = true
}

func ChanInfo() (l, c int) {
	l = len(inputCh)
	c = cap(inputCh)
	return
}

// Deprecated
func Feed(data []byte, category, name string, opt *Option) error {
	return doFeed(data, category, name, opt)
}

func doFeed(data []byte, category, name string, opt *Option) error {

	switch category {
	case Metric, KeyEvent, Object, Logging, Tracing:
		// metric line check
		if err := checkMetric(data); err != nil {
			return fmt.Errorf("invalid line protocol data %v", err)
		}
	case Rum: // do not check RUM data structure, too complecated
	default:
		return fmt.Errorf("invalid category %s", category)
	}

	if testAssert {
		l.Infof("[%s] source: `%s', data: %s", category, name, data)
		return nil
	}

	select {
	case inputCh <- &iodata{
		category: category,
		opt:      opt,
		data:     data,
		name:     name,
	}: // XXX: blocking

	case <-datakit.Exit.Wait():
		l.Warnf("%s/%s feed skipped on global exit", category, name)
	}

	return nil
}

func checkMetric(data []byte) error {
	if datakit.Cfg.MainCfg.StrictMode {
		_, err := influxm.ParsePointsWithPrecision(data, time.Now().UTC(), "n")
		if err != nil {
			l.Errorf("[error] : %s", err.Error())
			return err
		}
	}
	return nil
}

type Option struct {
	CollectCost time.Duration
}

func NamedFeed(data []byte,
	category, name string) error {
	return doFeed(data, category, name, nil)
}

func NamedFeedPoints(pts []influxm.Point, category, name string) error {
	if len(pts) == 0 {
		return nil
	}

	lines := []string{}
	for _, p := range pts {
		lines = append(lines, p.String())
	}

	return NamedFeed([]byte(strings.Join(lines, "\n")), category, name)
}

func NamedFeedEx(name, category, metric string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time) error {
	return doFeedEx(name, category, metric, tags, fields, t...)
}

func doFeedEx(name, category, metric string,
	tags map[string]string,
	fields map[string]interface{}, t ...time.Time) error {
	data, err := MakeMetric(metric, tags, fields, t...)
	if err != nil {
		return err
	}
	return doFeed(data, category, name, nil)
}

func MakeMetric(name string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time) ([]byte, error) {
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
		case int, uint32, uint16, uint8, int64, int32, int16, int8, bool, string, float32, float64:
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

func ioStop() {
	if outputFile != nil {
		if err := outputFile.Close(); err != nil {
			l.Error(err)
		}
	}
}

func startIO() {
	categoryURLs = map[string]string{
		Metric:   datakit.Cfg.MainCfg.DataWay.MetricURL(),
		KeyEvent: datakit.Cfg.MainCfg.DataWay.KeyEventURL(),
		Object:   datakit.Cfg.MainCfg.DataWay.ObjectURL(),
		Logging:  datakit.Cfg.MainCfg.DataWay.LoggingURL(),
		Tracing:  datakit.Cfg.MainCfg.DataWay.TracingURL(),
		Rum:      datakit.Cfg.MainCfg.DataWay.RumURL(),
	}

	l.Debugf("categoryURLs: %+#v", categoryURLs)
	var du time.Duration
	var err error

	if datakit.Cfg.MainCfg.DataWay.Timeout != "" {
		du, err = time.ParseDuration(datakit.Cfg.MainCfg.DataWay.Timeout)
		if err != nil {
			l.Warnf("parse dataway timeout failed: %s", err.Error())
			du = time.Second * 30
		}
	}

	httpCli = &http.Client{
		Timeout: du,
	}

	defer ioStop()

	var f rtpanic.RecoverCallback

	f = func(trace []byte, _ error) {
		defer rtpanic.Recover(f, nil)

		tick := time.NewTicker(datakit.IntervalDuration)
		defer tick.Stop()
		l.Debugf("io interval: %v", datakit.IntervalDuration)

		if trace != nil {
			l.Warn("recover ok")
		}

		for {
			select {
			case d := <-inputCh:
				if d == nil {
					l.Warn("get empty data, ignored")
				} else {

					now := time.Now()

					if d.name == "tailf" && datakit.Cfg.MainCfg.LogUpload {
					} else {
						l.Debugf("get iodata(%d bytes) from %s|%s", len(d.data), d.category, d.name)
					}

					stat, ok := inputstats[d.name]
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
						inputstats[d.name] = stat

					} else {
						stat.Total += int64(len(d.data))
						stat.Count++
						stat.Last = now
						stat.Category = d.category

						if d.opt != nil {
							stat.totalCost += d.opt.CollectCost
							stat.AvgCollectCost = (stat.totalCost) / time.Duration(stat.Count)
						}
					}

					// disable cache under proxied mode, to prevent large packages in proxing lua module
					if datakit.Cfg.MainCfg.DataWay.Proxy {
						if err := doFlush([][]byte{d.data}, d.category); err != nil {
							l.Errorf("post %s failed, drop %d packages", d.category, len(d.data))
						}
					} else {
						cache[d.category] = append(cache[d.category], d.data)
						curCacheCnt++

						if curCacheCnt > maxCacheCnt {
							flushAll(cache)
						}
					}
				}

			case q := <-qstatsCh:
				statRes := []*InputsStat{}
				for _, v := range inputstats {
					statRes = append(statRes, v)
				}
				select {
				case q.ch <- statRes: // maybe blocking(i.e., client canceled)
				default:
					l.Warn("client canceled")
					// pass
				}

			case <-tick.C:
				flushAll(cache)

			case <-datakit.Exit.Wait():
				l.Info("io exit on exit")
				return
			}
		}
	}

	l.Info("starting...")
	f(nil, nil)
}

func Start() {

	l = logger.SLogger("io")

	datakit.WG.Add(1)
	go func() {
		defer datakit.WG.Done()
		startIO()
	}()

	datakit.WG.Add(1)
	go func() {
		defer datakit.WG.Done()
		GRPCServer()
	}()
}

func flushAll(cache map[string][][]byte) {
	flush(cache)

	if curCacheCnt > 0 {
		l.Warnf("post failed cache count: %d", curCacheCnt)
	}

	if curCacheCnt > MaxPostFailCache {
		l.Warnf("failed cache count reach max limit(%d), cleanning cache...", MaxPostFailCache)
		for k, _ := range cache {
			cache[k] = nil
		}
		curCacheCnt = 0
	}
}

func flush(cache map[string][][]byte) {

	defer httpCli.CloseIdleConnections()
	for k, v := range cache {
		if err := doFlush(v, k); err != nil {
			l.Errorf("post %d to %s failed", len(v), k)
		} else {
			if len(v) > 0 {
				curCacheCnt -= len(v)
				l.Debugf("clean %d/%d cache on %s", len(v), curCacheCnt, k)
				cache[k] = nil
			}
		}
	}
}

func buildBody(url string, bodies [][]byte) (body []byte, gzon bool, err error) {
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

func doFlush(bodies [][]byte, url string) error {

	if bodies == nil {
		return nil
	}

	body, gz, err := buildBody(url, bodies)
	if err != nil {
		return err
	}

	if datakit.OutputFile != "" {
		return fileOutput(body)
	}

	req, err := http.NewRequest("POST", categoryURLs[url], bytes.NewBuffer(body))
	if err != nil {
		l.Error(err)
		return err
	}

	if gz {
		req.Header.Set("Content-Encoding", "gzip")
	}

	postbeg := time.Now()

	resp, err := httpCli.Do(req)
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
		l.Debugf("post %d to %s ok(gz: %v), cost %v",
			len(body), url, gz, time.Since(postbeg))
		return nil

	case 4:
		l.Debugf("post %d to %s failed(HTTP: %s): %s, cost %v, data dropped",
			len(body), url, resp.StatusCode, string(respbody), time.Since(postbeg))
		return nil

	case 5:
		l.Debugf("post %d to %s failed(HTTP: %s): %s, cost %v",
			len(body), url, resp.StatusCode, string(respbody), time.Since(postbeg))
		return fmt.Errorf("dataway internal error")
	}

	return nil
}

func fileOutput(body []byte) error {

	if outputFile == nil {
		f, err := os.OpenFile(datakit.OutputFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			l.Error(err)
			return err
		}

		outputFile = f
	}

	if _, err := outputFile.Write(append(body, '\n')); err != nil {
		l.Error(err)
		return err
	}

	outputFileSize += int64(len(body))
	if outputFileSize > 4*1024*1024 {
		if err := outputFile.Truncate(0); err != nil {
			l.Error(err)
			return err
		}
		outputFileSize = 0
	}

	return nil
}

var (
	statsTimeout = time.Second * 3
)

func GetStats() ([]*InputsStat, error) {
	q := &qstats{
		ch: make(chan []*InputsStat),
	}

	tick := time.NewTicker(statsTimeout)
	defer tick.Stop()

	select {
	case qstatsCh <- q:
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

func SetTest() {
	testAssert = true
}
