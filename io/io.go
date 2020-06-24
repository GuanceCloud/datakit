package io

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/system/rtpanic"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

var (
	input   chan *iodata
	l       *zap.SugaredLogger
	baseURL string

	httpCli      *http.Client
	categoryURLs map[string]string
)

const ( // categories
	Metric   = "/v1/write/metrics"
	KeyEvent = "/v1/write/keyevent"
	Object   = "/v1/write/object"
	Logging  = "/v1/write/logging"
)

type iodata struct {
	category string
	data     []byte // line-protocol or json or others
}

func init() {
	input = make(chan *iodata, 128)
	httpCli = &http.Client{}
}

func Feed(data []byte, category string) error {
	switch category {
	case Metric, KeyEvent, Object, Logging:
	default:
		return fmt.Errorf("invalid category %s", category)
	}

	input <- &iodata{
		category: category,
		data:     data,
	} // XXX: blocking

	return nil
}

func Start() {
	tick := time.NewTicker(time.Second * 10) // FIXME: duration should configurable
	l = logger.SLogger("io")

	baseURL = "http://" + path.Join(config.Cfg.MainCfg.DataWay.Host)
	if config.Cfg.MainCfg.DataWay.Scheme == "https" {
		baseURL = "https://" + path.Join(config.Cfg.MainCfg.DataWay.Host)
	}

	categoryURLs = map[string]string{
		Metric:   path.Join(baseURL, Metric) + "?token=" + config.Cfg.MainCfg.DataWay.Token,
		KeyEvent: path.Join(baseURL, KeyEvent) + "?token=" + config.Cfg.MainCfg.DataWay.Token,
		Object:   path.Join(baseURL, Object) + "?token=" + config.Cfg.MainCfg.DataWay.Token,
		Logging:  path.Join(baseURL, Logging) + "?token=" + config.Cfg.MainCfg.DataWay.Token,
	}

	cache := map[string][][]byte{
		Metric:   [][]byte{},
		KeyEvent: [][]byte{},
		Object:   [][]byte{},
		Logging:  [][]byte{},
	}

	var f rtpanic.RecoverCallback

	f = func(trace []byte, _ error) {
		defer rtpanic.Recover(f, nil)

		if trace != nil {
			l.Warn("recover ok")
		}

		for {
			select {
			case d := <-input:
				// TODO
				cache[d.category] = append(cache[d.category], d.data)

			case <-tick.C:
				// TODO: Flush to file/http
				flush(cache)

				// TODO: add global exit entry
			case <-config.Exit.Wait():
				l.Info("io exit.")
			}
		}
	}

	l.Info("starting...")
	f(nil, nil)
}

func flush(cache map[string][][]byte) {
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

func doFlush(bodies [][]byte, url string) error {

	body := bytes.Join(bodies, []byte("\n"))

	gz := false
	if len(body) > 1024 { // Gzip ?
	}

	req, err := http.NewRequest("POST", categoryURLs[url], bytes.NewBuffer(body))
	if err != nil {
		l.Error(err)
		return err
	}

	req.Header.Set("X-Datakit-UUID", config.Cfg.MainCfg.UUID)
	req.Header.Set("X-Version", git.Version)
	req.Header.Set("X-Version", config.DKUserAgent)
	if gz {
		req.Header.Set("Content-Encoding", "gzip")
	}

	switch url {
	case Object: // object is json
		req.Header.Set("Content-Type", "application/json")
	default: // others are line-protocol
	}

	if config.MaxLifeCheckInterval > 0 {
		req.Header.Set("X-Max-POST-Interval", fmt.Sprintf("%v", config.MaxLifeCheckInterval))
	}

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
		l.Debugf("post to %s ok", url)
	case 4:
		l.Errorf("post to %s failed(HTTP: %d): %s, data dropped", url, resp.StatusCode, string(respbody))
	case 5:
		l.Warnf("post to %s failed(HTTP: %d): %s", url, resp.StatusCode, string(respbody))
		return fmt.Errorf("dataway internal error")
	}

	return nil
}
