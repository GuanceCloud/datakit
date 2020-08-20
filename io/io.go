package io

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"time"

	ifxcli "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

var (
	l       *logger.Logger
	httpCli *http.Client
	baseURL string

	inputCh    = make(chan *iodata, 1024)
	inputstats = map[string]*InputsStat{}

	qstatsCh = make(chan *qstats)

	cache = map[string][][]byte{
		__MetricDeprecated: nil,
		Metric:             nil,
		KeyEvent:           nil,
		Object:             nil,
		Logging:            nil,
	}

	categoryURLs map[string]string

	outputFile     *os.File
	outputFileSize int64
	cookies        string
)

const ( // categories
	__MetricDeprecated = "/v1/write/metrics"
	Metric             = "/v1/write/metric"
	KeyEvent           = "/v1/write/keyevent"
	Object             = "/v1/write/object"
	Logging            = "/v1/write/logging"
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
	CrashCnt int       `json:"crash_cnt"`
}

type qstats struct {
	ch chan []*InputsStat
}

func ChanInfo() (int, int) {
	return len(inputCh), cap(inputCh)
}

// Deprecated
func Feed(data []byte, category string) error {
	return doFeed(data, category, "")
}

func doFeed(data []byte, category, name string) error {
	switch category {
	case Metric, KeyEvent, Object, Logging:
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

func NamedFeed(data []byte, catagory, name string) error {
	return doFeed(data, catagory, name)
}

// Deprecated
func FeedEx(catagory string, metric string, tags map[string]string, fields map[string]interface{}, t ...time.Time) error {
	return doFeedEx("", catagory, metric, tags, fields, t...)
}

func NamedFeedEx(name, catagory string, metric string, tags map[string]string, fields map[string]interface{}, t ...time.Time) error {
	return doFeedEx(name, catagory, metric, tags, fields, t...)
}

func doFeedEx(name, catagory string, metric string, tags map[string]string, fields map[string]interface{}, t ...time.Time) error {
	data, err := MakeMetric(metric, tags, fields, t...)
	if err != nil {
		return err
	}
	return doFeed(data, catagory, name)
}

func MakeMetric(name string, tags map[string]string, fields map[string]interface{}, t ...time.Time) ([]byte, error) {
	var tm time.Time
	if len(t) > 0 {
		tm = t[0]
	} else {
		tm = time.Now().UTC()
	}

	if len(config.Cfg.MainCfg.GlobalTags) > 0 {
		if tags == nil {
			tags = map[string]string{}
		}

		for k, v := range config.Cfg.MainCfg.GlobalTags {
			if _, ok := tags[k]; !ok { // do not overwrite exists tags
				tags[k] = v
			}
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
	baseURL = "http://" + config.Cfg.MainCfg.DataWay.Host
	if config.Cfg.MainCfg.DataWay.Scheme == "https" {
		baseURL = "https://" + config.Cfg.MainCfg.DataWay.Host
	}

	categoryURLs = map[string]string{

		__MetricDeprecated: baseURL + __MetricDeprecated + "?token=" + config.Cfg.MainCfg.DataWay.Token,
		Metric:             baseURL + Metric + "?token=" + config.Cfg.MainCfg.DataWay.Token,
		KeyEvent:           baseURL + KeyEvent + "?token=" + config.Cfg.MainCfg.DataWay.Token,
		Object:             baseURL + Object + "?token=" + config.Cfg.MainCfg.DataWay.Token,
		Logging:            baseURL + Logging + "?token=" + config.Cfg.MainCfg.DataWay.Token,
	}

	l.Debugf("categoryURLs: %+#v", categoryURLs)

	du, err := time.ParseDuration(config.Cfg.MainCfg.DataWay.Timeout)
	if err != nil {
		l.Warnf("parse dataway timeout failed: %s", err.Error())
		du = time.Second * 30
	}

	httpCli = &http.Client{
		Timeout: du,
	}

	defer ioStop()

	var f rtpanic.RecoverCallback

	f = func(trace []byte, _ error) {
		defer rtpanic.Recover(f, nil)

		tick := time.NewTicker(config.Cfg.MainCfg.IntervalDuration)
		defer tick.Stop()
		l.Debugf("io interval: %v", config.Cfg.MainCfg.IntervalDuration)

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

					l.Debugf("get iodata(%d bytes) from %s|%s", len(d.data), d.category, d.name)
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

func gz(data []byte) ([]byte, error) {
	var z bytes.Buffer
	zw := gzip.NewWriter(&z)
	if _, err := zw.Write(data); err != nil {
		l.Error(err)
		return nil, err
	}

	zw.Flush()
	zw.Close()
	return z.Bytes(), nil
}

func doFlush(bodies [][]byte, url string) error {

	var err error

	if bodies == nil {
		return nil
	}

	if cookies == "" {
		cookies = fmt.Sprintf("uuid=%s;name=%s;hostname=%s;max_post_interval=%s;version=%s;os=%s;arch=%s",
			config.Cfg.MainCfg.UUID,
			config.Cfg.MainCfg.Name,
			config.Cfg.MainCfg.Hostname,
			datakit.MaxLifeCheckInterval,
			git.Version,
			runtime.GOOS,
			runtime.GOARCH)
	}

	body := bytes.Join(bodies, []byte("\n"))
	switch url {
	case Object: // object is json
		all_objs := []map[string]interface{}{}

		for _, data := range bodies {

			var objs []map[string]interface{}
			err := json.Unmarshal(data, &objs)
			if err != nil {
				l.Error(err)
				return err
			}
			all_objs = append(all_objs, objs...)
		}

		body, err = json.Marshal(all_objs)
		if err != nil {
			l.Error(err)
			return err
		}
	default: // others are line-protocol
	}

	if datakit.OutputFile != "" {
		return fileOutput(body)
	}

	gzOn := false
	if len(body) > 1024 {
		gzbody, err := gz(body)
		if err != nil {
			return err
		}

		l.Debugf("gzip %d/%d", len(body), len(gzbody))

		gzOn = true
		body = gzbody
	}

	req, err := http.NewRequest("POST", categoryURLs[url], bytes.NewBuffer(body))
	if err != nil {
		l.Error(err)
		return err
	}

	req.Header.Set("Cookie", cookies)
	if gzOn {
		req.Header.Set("Content-Encoding", "gzip")
	}

	switch url {
	case Object: // object is json
		req.Header.Set("Content-Type", "application/json")
	default: // others are line-protocol
	}

	if datakit.MaxLifeCheckInterval > 0 {
		l.Debugf("set max-post-interval: %v", datakit.MaxLifeCheckInterval)
		req.Header.Set("X-Max-POST-Interval", fmt.Sprintf("%v", datakit.MaxLifeCheckInterval))
	} else {
		l.Debugf("max-post-interval not set: %+#v", datakit.MaxLifeCheckInterval)
	}

	l.Debugf("post to %s...", categoryURLs[url])

	resp, err := httpCli.Do(req)
	if err != nil {
		l.Error(err)
		return err
	}

	l.Debugf("get resp from %s...", categoryURLs[url])
	defer resp.Body.Close()
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Error(err)
		return err
	}

	switch resp.StatusCode / 100 {
	case 2:
		l.Debugf("post to %s ok", url)
	case 4:
		l.Errorf("post to %s failed(HTTP: %d): %s, data dropped", url, resp.StatusCode, string(respbody))
	case 5:
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

func GetStats() ([]*InputsStat, error) {
	q := &qstats{
		ch: make(chan []*InputsStat),
	}

	tick := time.NewTicker(time.Second * 3)
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
