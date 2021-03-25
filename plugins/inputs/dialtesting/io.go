package dialtesting

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"time"

	influxm "github.com/influxdata/influxdb1-client/models"
	ifxcli "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	httpCli *http.Client
	// inputCh = make(chan *iodata, datakit.CommonChanCap)
	// cache   = map[string][][]byte{}

	maxCacheCnt = 128

	MaxPostFailCache = 1024
	outputFile       *os.File
	outputFileSize   int64
)

type iodata struct {
	name string
	data []byte // line-protocol or json or others
	url  string
}

type IO struct {
	curCacheCnt int
	cache       map[string][][]byte
	inputCh     chan *iodata
}

func NewIO() *IO {
	return &IO{
		curCacheCnt: 0,
		cache:       map[string][][]byte{},
		inputCh:     make(chan *iodata, datakit.CommonChanCap),
	}

}

func (i *IO) startIO() {

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
			case d := <-i.inputCh:
				if d == nil {
					l.Warn("get empty data, ignored")
				} else {

					var isProxy bool

					// 考虑到推送至不同的dataway地址

					u, err := url.Parse(d.url)
					if err != nil {
						l.Warn("get invalid url, ignored")
						continue
					}
					if u.Path == "/proxy" {
						isProxy = true
					}
					u.Path = u.Path + `/v1/write/metric`
					d.url = u.String()

					// disable cache under proxied mode, to prevent large packages in proxing lua module
					if isProxy {
						if err := doFlush([][]byte{d.data}, d.url); err != nil {
							l.Errorf("post %s failed, drop %d packages", d.url, len(d.data))
						}
					} else {
						i.cache[d.url] = append(i.cache[d.url], d.data)
						i.curCacheCnt++

						if i.curCacheCnt > maxCacheCnt {
							i.flushAll()
						}
					}
				}

			case <-tick.C:
				i.flushAll()

			case <-datakit.Exit.Wait():
				l.Info("io exit on exit")
				return
			}
		}
	}

	l.Info("starting...")
	f(nil, nil)
}

func ioStop() {
	if outputFile != nil {
		if err := outputFile.Close(); err != nil {
			l.Error(err)
		}
	}
}

func buildBody(url string, bodies [][]byte) (body []byte, gzon bool, err error) {
	body = bytes.Join(bodies, []byte("\n"))

	if body, err = datakit.GZip(body); err != nil {
		l.Errorf("gz: %s", err.Error())
		return
	}
	gzon = true

	return
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

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
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

func (i *IO) flush() {

	defer httpCli.CloseIdleConnections()
	for k, v := range i.cache {
		if err := doFlush(v, k); err != nil {
			l.Errorf("post %d to %s failed", len(v), k)
		} else {
			i.curCacheCnt -= len(v)
			i.cache[k] = nil
		}
	}
}

func (i *IO) flushAll() {
	i.flush()

	if i.curCacheCnt > 0 {
		l.Warnf("post failed cache count: %d", i.curCacheCnt)
	}

	if i.curCacheCnt > MaxPostFailCache {
		l.Warnf("failed cache count reach max limit(%d), cleanning cache...", MaxPostFailCache)
		for k, _ := range i.cache {
			i.cache[k] = nil
		}
		i.curCacheCnt = 0
	}
}

func (i *IO) doFeed(inputName string, data []byte, url string) error {

	if err := checkMetric(data); err != nil {
		return fmt.Errorf("invalid line protocol data %v", err)
	}

	select {
	case i.inputCh <- &iodata{
		data: data,
		name: inputName,
		url:  url,
	}: // XXX: blocking

	case <-datakit.Exit.Wait():
		l.Warnf("%s feed skipped on global exit", inputName)
	}

	return nil
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
			if v.(uint64) > uint64(math.MaxInt64) {
				l.Warnf("on input `%s', filed %s, get uint64 %d > MaxInt64(%d), dropped", name, k, v.(uint64), int64(math.MaxInt64))
				delete(fields, k)
			} else { // convert uint64 -> int64
				fields[k] = int64(v.(uint64))
			}
		case uint32, uint16, uint8, int64, int32, int16, int8, bool, string, float32, float64:
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
