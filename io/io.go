package io

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"time"

	ifxcli "github.com/influxdata/influxdb1-client/v2"

	influxm "github.com/influxdata/influxdb1-client/models"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
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
	}

	categoryURLs map[string]string

	outputFile     *os.File
	outputFileSize int64
	cookies        string
)

const ( // categories
	MetricDeprecated = "/v1/write/metrics"
	Metric           = "/v1/write/metric"
	KeyEvent         = "/v1/write/keyevent"
	Object           = "/v1/write/object"
	Logging          = "/v1/write/logging"

	minGZSize = 1024

	httpDiv = 100
	httpOk  = 2
	httpBad = 4
	httpErr = 5
)

type iodata struct {
	category, name string
	data           []byte // line-protocol or json or others
}

type InputsStat struct {
	Name     string    `json:"name"`
	Category string    `json:"category"`
	Total    int64     `json:"total"`
	Count    int64     `json:"count"`
	First    time.Time `json:"first"`
	Last     time.Time `json:"last"`
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
func Feed(data []byte, category string) error {
	return doFeed(data, category, "")
}

func doFeed(data []byte, category, name string) error {
	if testAssert {
		l.Infof("[%s] source: %s data: %s", category, name, data)
		return nil
	}

	switch category {
	case Metric, KeyEvent, Logging:
		// metric line check
		if err := checkMetric(data); err != nil {
			return fmt.Errorf("invalid line protocol data %v", err)
		}
	case Object:
	default:
		return fmt.Errorf("invalid category %s", category)
	}

	select {
	case inputCh <- &iodata{
		category: category,
		data:     data,
		name:     name,
	}: // XXX: blocking

	case <-datakit.Exit.Wait():
		l.Warn("feed skipped on global exit")
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

func NamedFeed(data []byte, category, name string) error {
	return doFeed(data, category, name)
}

// Deprecated
func FeedEx(category, metric string, tags map[string]string, fields map[string]interface{}, t ...time.Time) error {
	return doFeedEx("", category, metric, tags, fields, t...)
}

func NamedFeedEx(name, category, metric string, tags map[string]string, fields map[string]interface{}, t ...time.Time) error {
	return doFeedEx(name, category, metric, tags, fields, t...)
}

func doFeedEx(name, category, metric string, tags map[string]string, fields map[string]interface{}, t ...time.Time) error {
	data, err := MakeMetric(metric, tags, fields, t...)
	if err != nil {
		return err
	}
	return doFeed(data, category, name)
}

func MakeMetric(name string, tags map[string]string, fields map[string]interface{}, t ...time.Time) ([]byte, error) {
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
			fields[k] = fmt.Sprintf("%d", v.(uint64)) // convert uint64 to string to avoid overflow
			l.Warnf("force convert uint64 to string(%d -> %s)", v.(uint64), fields[k])
		case uint32:
			fields[k] = int64(v.(uint32))
			l.Warn("force convert uint32 to int64")
		case uint16:
			fields[k] = int64(v.(uint16))
			l.Warn("force convert uint16 to int64")
		case uint8:
			fields[k] = int64(v.(uint8))
			l.Warn("force convert uint8 to int64")
		default:
			// pass
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
	baseURL = "http://" + datakit.Cfg.MainCfg.DataWay.Host
	if datakit.Cfg.MainCfg.DataWay.Scheme == "https" {
		baseURL = "https://" + datakit.Cfg.MainCfg.DataWay.Host
	}

	categoryURLs = map[string]string{

		MetricDeprecated: baseURL + MetricDeprecated + "?token=" + datakit.Cfg.MainCfg.DataWay.Token,
		Metric:           baseURL + Metric + "?token=" + datakit.Cfg.MainCfg.DataWay.Token,
		KeyEvent:         baseURL + KeyEvent + "?token=" + datakit.Cfg.MainCfg.DataWay.Token,
		Object:           baseURL + Object + "?token=" + datakit.Cfg.MainCfg.DataWay.Token,
		Logging:          baseURL + Logging + "?token=" + datakit.Cfg.MainCfg.DataWay.Token,
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

	if datakit.MaxLifeCheckInterval > 0 {
		l.Debugf("max-post-interval: %v", datakit.MaxLifeCheckInterval)
	} else {
		l.Debugf("max-post-interval not set")
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

					cache[d.category] = append(cache[d.category], d.data)

					stat, ok := inputstats[d.name]
					if !ok {
						inputstats[d.name] = &InputsStat{
							Name:     d.name,
							Category: d.category,
							Total:    int64(len(d.data)),
							First:    now,
							Count:    1,
							Last:     now,
						}
					} else {
						stat.Total += int64(len(d.data))
						stat.Count++
						stat.Last = now
						stat.Category = d.category
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
				flush(cache)

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
		GRPCServer(datakit.GRPCDomainSock)
	}()
}

func flush(cache map[string][][]byte) {

	defer httpCli.CloseIdleConnections()

	if err := doFlush(cache[Metric], Metric); err == nil {
		cache[Metric] = nil
	}

	if err := doFlush(cache[KeyEvent], KeyEvent); err == nil {
		cache[KeyEvent] = nil
	}

	if err := doFlush(cache[Object], Object); err == nil {
		cache[Object] = nil
	}

	if err := doFlush(cache[Logging], Logging); err == nil {
		cache[Logging] = nil
	}
}

func initCookies() {
	cookies = fmt.Sprintf("uuid=%s;name=%s;hostname=%s;max_post_interval=%s;version=%s;os=%s;arch=%s",
		datakit.Cfg.MainCfg.UUID,
		datakit.Cfg.MainCfg.Name,
		datakit.Cfg.MainCfg.Hostname,
		datakit.MaxLifeCheckInterval,
		git.Version,
		runtime.GOOS,
		runtime.GOARCH)
}

func buildObjBody(bodies [][]byte) ([]byte, error) {
	allObjs := make([]map[string]interface{}, 0)

	for _, data := range bodies {
		objs := make([]map[string]interface{}, 0)
		if err := json.Unmarshal(data, &objs); err != nil {
			l.Error(err)
			return nil, err
		}
		allObjs = append(allObjs, objs...)
	}

	jbody, err := json.Marshal(allObjs)
	if err != nil {
		l.Error(err)
		return nil, err
	}

	return jbody, nil
}

func buildBody(url string, bodies [][]byte) (body []byte, gzon bool, err error) {
	switch url {
	case Object: // object is json

		body, err = buildObjBody(bodies) // convert raw objects bytes as json array
		if err != nil {
			return
		}

	default: // others are line-protocol
		body = bytes.Join(bodies, []byte("\n"))
	}

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

	if cookies == "" {
		initCookies()
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

	req.Header.Set("Cookie", cookies)
	if gz {
		req.Header.Set("Content-Encoding", "gzip")
	}

	switch url {
	case Object: // object is json
		req.Header.Set("Content-Type", "application/json")
	default: // others are line-protocol
	}

	if datakit.MaxLifeCheckInterval > 0 {
		req.Header.Set("X-Max-POST-Interval", fmt.Sprintf("%v", datakit.MaxLifeCheckInterval))
	}

	l.Debugf("post to %s...", categoryURLs[url])

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

	l.Debugf("post cost %v", time.Since(postbeg))

	switch resp.StatusCode / httpDiv {
	case httpOk:
		l.Debugf("post to %s ok", url)
	case httpBad:
		l.Errorf("post to %s failed(HTTP: %d): %s, data dropped", url, resp.StatusCode, string(respbody))
	case httpErr:
		l.Warnf("post to %s failed(HTTP: %d): %s", url, resp.StatusCode, string(respbody))
		return fmt.Errorf("dataway internal error")
	}

	return nil
}

func fileOutput(body []byte) error {

	if outputFile == nil {
		f, err := os.OpenFile(datakit.OutputFile, os.O_WRONLY|os.O_APPEND, 0644)
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
