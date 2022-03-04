package dataway

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gopkg.in/CodapeWild/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/CodapeWild/dd-trace-go.v1/ddtrace/tracer"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

func (dc *endPoint) send(category string, data []byte, gz bool) error {
	var err error
	span, _ := tracer.StartSpanFromContext(context.Background(), "io.dataway.send", tracer.SpanType(ext.SpanTypeHTTP))
	defer func() {
		span.SetTag("fails", dc.fails)
		span.Finish(tracer.WithError(err))
	}()
	span.SetTag("category", category)
	span.SetTag("data_size", len(data))
	span.SetTag("is_gz", gz)

	requrl, ok := dc.categoryURL[category]
	if !ok {
		// for dialtesting, there are user-defined url to post
		if x, err := url.ParseRequestURI(category); err != nil {
			return fmt.Errorf("invalid url %s", category)
		} else {
			log.Debugf("try use URL %+#v", x)
			requrl = category
		}
	}

	var req *http.Request
	if req, err = http.NewRequest("POST", requrl, bytes.NewBuffer(data)); err != nil {
		log.Error(err)

		return err
	}
	span.SetTag("method", req.Method)
	span.SetTag("url", requrl)

	if gz {
		req.Header.Set("Content-Encoding", "gzip")
	}
	for k, v := range ExtraHeaders {
		req.Header.Set(k, v)
	}

	if dc.ontest {
		log.Debug("Datakit client on test")

		return nil
	}

	tracer.Inject(span.Context(), tracer.HTTPHeadersCarrier(req.Header)) //nolint:errcheck,gosec

	var resp *http.Response
	if resp, err = dc.dw.sendReq(req); err != nil {
		dc.fails++
		log.Errorf("request url %s failed(proxy: %s): %s", requrl, dc.proxy, err)

		return err
	}
	span.SetTag("status", resp.Status)

	defer resp.Body.Close() //nolint:errcheck
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Error(err)

		return err
	}

	postbeg := time.Now()
	switch resp.StatusCode / 100 {
	case 2:
		dc.fails = 0
		log.Debugf("post %d to %s ok(gz: %v), cost %v, response: %s",
			len(data), requrl, gz, time.Since(postbeg), string(body))
	case 4:
		dc.fails = 0
		log.Warnf("post %d to %s failed(HTTP: %s): %s, cost %v, data dropped",
			len(data),
			requrl,
			resp.Status,
			string(body),
			time.Since(postbeg))
	case 5:
		dc.fails++
		log.Errorf("[%d] post %d to %s failed(HTTP: %s): %s, cost %v", dc.fails,
			len(data),
			requrl,
			resp.Status,
			string(body),
			time.Since(postbeg))
		err = fmt.Errorf("dataway internal error")
	}

	return err
}

func (dw *DataWayCfg) sendReq(req *http.Request) (*http.Response, error) {
	log.Debugf("send request %s, proxy: %s, dwcli: %p", req.URL.String(), dw.HTTPProxy, dw.httpCli.Transport)

	return dw.httpCli.Do(req)
}

func (dw *DataWayCfg) Send(category string, data []byte, gz bool) error {
	for i, ep := range dw.endPoints {
		log.Debugf("send to %dth dataway, fails: %d/%d", i, ep.fails, dw.MaxFails)
		// 判断 fails
		if ep.fails > dw.MaxFails && len(AvailableDataways) > 0 {
			rand.Seed(time.Now().UnixNano())
			index := rand.Intn(len(AvailableDataways)) //nolint:gosec

			var err error
			url := fmt.Sprintf(`%s?%s`, AvailableDataways[index], ep.urlValues.Encode())
			ep, err = dw.initEndpoint(url)
			if err != nil {
				log.Error(err)
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

func (dw *DataWayCfg) Write(category string, pts []*influxdb.Point) error {
	if len(pts) == 0 {
		return nil
	}

	bodies, err := dw.buildBody(pts, true)
	if err != nil {
		return err
	}

	for _, body := range bodies {
		if err := dw.Send(category, body.buf, body.gzon); err != nil {
			log.Error(err)
			return err
		}
	}

	return nil
}

const (
	minGZSize   = 1024
	maxKodoPack = 10 * 1000 * 1000
)

type body struct {
	buf  []byte
	gzon bool
}

func (dw *DataWayCfg) buildBody(pts []*influxdb.Point, isGzip bool) ([]*body, error) {
	lines := bytes.Buffer{}
	var (
		gz = func(lines []byte) (*body, error) {
			var (
				body = &body{buf: lines}
				err  error
			)
			log.Debugf("### io body size before GZ: %dM %dK", len(body.buf)/1000/1000, len(body.buf)/1000)
			if len(lines) > minGZSize && isGzip {
				if body.buf, err = datakit.GZip(body.buf); err != nil {
					log.Errorf("gz: %s", err.Error())

					return nil, err
				}
				body.gzon = true
			}

			return body, nil
		}
		// lines  bytes.Buffer
		bodies []*body
	)
	lines.Reset()
	for _, pt := range pts {
		ptstr := pt.String()
		if lines.Len()+len(ptstr)+1 >= maxKodoPack {
			if body, err := gz(lines.Bytes()); err != nil {
				return nil, err
			} else {
				log.Warn(string(body.buf))
				bodies = append(bodies, body)
			}
			lines.Reset()
		}
		lines.WriteString(ptstr)
		lines.WriteString("\n")
	}
	if body, err := gz(lines.Bytes()); err != nil {
		return nil, err
	} else {
		return append(bodies, body), nil
	}
}
