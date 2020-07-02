package io

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
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

	inited = false
)

const ( // categories
	MetricDeprecated = "/v1/write/metrics"
	Metric           = "/v1/write/metric"
	KeyEvent         = "/v1/write/keyevent"
	Object           = "/v1/write/object"
	Logging          = "/v1/write/logging"

	initErr = "you should call io.Init() before starting any services under io"
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

// there is more than 1 service under io module, we should init some data before these services starting
func Init() {
	l = logger.SLogger("io")
	inited = true
}

func Start() {

	if !inited {
		panic(initErr)
	}

	baseURL = "http://" + config.Cfg.MainCfg.DataWay.Host
	if config.Cfg.MainCfg.DataWay.Scheme == "https" {
		baseURL = "https://" + config.Cfg.MainCfg.DataWay.Host
	}

	categoryURLs = map[string]string{

		MetricDeprecated: baseURL + MetricDeprecated + "?token=" + config.Cfg.MainCfg.DataWay.Token,
		Metric:           baseURL + Metric + "?token=" + config.Cfg.MainCfg.DataWay.Token,
		KeyEvent:         baseURL + KeyEvent + "?token=" + config.Cfg.MainCfg.DataWay.Token,
		Object:           baseURL + Object + "?token=" + config.Cfg.MainCfg.DataWay.Token,
		Logging:          baseURL + Logging + "?token=" + config.Cfg.MainCfg.DataWay.Token,
	}

	l.Debugf("categoryURLs: %+#v", categoryURLs)

	cache := map[string][][]byte{
		MetricDeprecated: nil,
		Metric:           nil,
		KeyEvent:         nil,
		Object:           nil,
		Logging:          nil,
	}

	var f rtpanic.RecoverCallback

	f = func(trace []byte, _ error) {
		defer rtpanic.Recover(f, nil)

		tick := time.NewTicker(config.Cfg.MainCfg.Interval.Duration)
		defer tick.Stop()
		l.Debugf("io interval: %v", config.Cfg.MainCfg.Interval.Duration)

		if trace != nil {
			l.Warn("recover ok")
		}

		for {
			select {
			case d := <-input:
				if d == nil {
					l.Warn("get empty data, ignored")
				} else {

					l.Debugf("get iodata(%d bytes) from %s", len(d.data), d.category)
					cache[d.category] = append(cache[d.category], d.data)
				}

			case <-tick.C:
				flush(cache)

			case <-config.Exit.Wait():
				l.Info("exit")
				return
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

	if bodies == nil {
		return nil
	}

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
