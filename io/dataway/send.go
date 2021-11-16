package dataway

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/ddtrace/tracer"
	dktracer "gitlab.jiagouyun.com/cloudcare-tools/datakit/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

func (dc *endPoint) send(category string, data []byte, gz bool) error {
	dktracer.GlobalTracer.Start(tracer.WithLogger(&tracer.SimpleLogger{}))
	defer dktracer.GlobalTracer.Stop()

	requrl, ok := dc.categoryURL[category]
	if !ok {
		// for dialtesting, there are user-defined url to post
		if x, err := url.ParseRequestURI(category); err != nil {
			return fmt.Errorf("invalid url %s", category)
		} else {
			l.Debugf("try use URL %+#v", x)
			requrl = category
		}
	}

	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(data))
	if err != nil {
		l.Error(err)
		return err
	}

	if gz {
		req.Header.Set("Content-Encoding", "gzip")
	}

	// append extra headers
	for k, v := range ExtraHeaders {
		req.Header.Set(k, v)
	}

	if dc.ontest {
		l.Debug("Datakit client on test")
		return nil
	}

	// start trace span from request context
	span, _ := dktracer.GlobalTracer.StartSpanFromContext(req.Context(),
		"datakit.dataway.send", req.RequestURI, ext.SpanTypeHTTP)
	defer dktracer.GlobalTracer.FinishSpan(span, tracer.WithFinishTime(time.Now()))

	// inject span into http header
	if err := dktracer.GlobalTracer.Inject(span, req.Header); err != nil {
		l.Warnf("GlobalTracer.Inject: %s, ignored", err.Error())
	}

	postbeg := time.Now()
	resp, err := dc.dw.sendReq(req)
	if err != nil {
		dktracer.GlobalTracer.SetTag(span, "http_client_do_error", err.Error())
		l.Errorf("request url %s failed(proxy: %s): %s", requrl, dc.proxy, err)
		dc.fails++

		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		dktracer.GlobalTracer.SetTag(span, "io_read_request_body_error", err.Error())
		l.Error(err)

		return err
	}

	switch resp.StatusCode / 100 {
	case 2:
		dc.fails = 0
		l.Debugf("post %d to %s ok(gz: %v), cost %v, response: %s",
			len(data), requrl, gz, time.Since(postbeg), string(respbody))
		return nil

	case 4:
		dc.fails = 0
		dktracer.GlobalTracer.SetTag(span, "http_request_400_error",
			fmt.Errorf("%d: %s", resp.StatusCode, http.StatusText(resp.StatusCode)))

		l.Warnf("post %d to %s failed(HTTP: %s): %s, cost %v, data dropped",
			len(data),
			requrl,
			resp.Status,
			string(respbody),
			time.Since(postbeg))
		return nil

	case 5:
		dc.fails++
		dktracer.GlobalTracer.SetTag(span, "http_request_500_error",
			fmt.Errorf("%d: %s", resp.StatusCode, http.StatusText(resp.StatusCode)))

		l.Errorf("[%d] post %d to %s failed(HTTP: %s): %s, cost %v", dc.fails,
			len(data),
			requrl,
			resp.Status,
			string(respbody),
			time.Since(postbeg))
		return fmt.Errorf("dataway internal error")
	}

	return nil
}

func (dw *DataWayCfg) sendReq(req *http.Request) (*http.Response, error) {
	l.Debugf("send request %s, proxy: %s, dwcli: %p", req.URL.String(), dw.HTTPProxy, dw.httpCli.Transport)
	return dw.httpCli.Do(req)
}

func (dw *DataWayCfg) Send(category string, data []byte, gz bool) error {
	for i, ep := range dw.endPoints {
		l.Debugf("send to %dth dataway, fails: %d/%d", i, ep.fails, dw.MaxFails)
		// 判断 fails
		if ep.fails > dw.MaxFails && len(AvailableDataways) > 0 {
			rand.Seed(time.Now().UnixNano())
			index := rand.Intn(len(AvailableDataways)) //nolint:gosec

			var err error
			url := fmt.Sprintf(`%s?%s`, AvailableDataways[index], ep.urlValues.Encode())
			ep, err = dw.initEndpoint(url)
			if err != nil {
				l.Error(err)
				return err
			}

			dw.endPoints[i] = ep
		}

		if err := ep.send(category, data, gz); err != nil {
			return err
		}
	}

	return nil
}
