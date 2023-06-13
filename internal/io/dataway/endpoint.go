// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/avast/retry-go"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	dnet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	pb "google.golang.org/protobuf/proto"
)

var (
	// DatakitUserAgent define HTTP User-Agent header.
	// user-agent format. See
	// 	 https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/User-Agent
	DatakitUserAgent = fmt.Sprintf("datakit-%s-%s/%s", runtime.GOOS, runtime.GOARCH, git.Version)

	httpFailRatio      = 0 // %n
	httpFailStart      time.Time
	httpFailDuration   time.Duration
	httpMockedFailResp *http.Response
)

// nolint: gochecknoinits
func init() {
	if v := datakit.GetEnv("ENV_DEBUG_HTTP_FAIL_RATIO"); v != "" {
		if x, err := strconv.ParseInt(v, 10, 64); err == nil {
			httpFailRatio = int(x)
			httpFailStart = time.Now()

			httpMockedFailResp = &http.Response{
				Status:     http.StatusText(500),
				StatusCode: 500,
			}
		}
	}

	if v := datakit.GetEnv("ENV_DEBUG_HTTP_FAIL_DURATION"); v != "" {
		if x, err := time.ParseDuration(v); err == nil {
			httpFailDuration = x
		}
	}
}

type endPoint struct {
	token       string
	host        string
	scheme      string
	categoryURL map[string]string
	httpCli     *http.Client

	// optionals
	proxy       string
	apis        []string
	httpTimeout time.Duration

	maxHTTPIdleConnectionPerHost int
	maxHTTPConnections           int
	httpIdleTimeout              time.Duration

	httpTrace bool
}

func (ep *endPoint) String() string {
	return fmt.Sprintf("[host: %s][token: %s][apis: %s]",
		ep.host, ep.token, strings.Join(ep.apis, ","))
}

type endPointOption func(*endPoint)

func withAPIs(arr []string) endPointOption {
	return func(ep *endPoint) {
		ep.apis = arr
	}
}

func withHTTPTrace(on bool) endPointOption {
	return func(ep *endPoint) {
		ep.httpTrace = on
	}
}

func withMaxHTTPIdleConnectionPerHost(n int) endPointOption {
	return func(ep *endPoint) {
		if n > 0 {
			ep.maxHTTPIdleConnectionPerHost = n
		}
	}
}

func withMaxHTTPConnections(n int) endPointOption {
	return func(ep *endPoint) {
		if n > 0 {
			ep.maxHTTPConnections = n
		}
	}
}

func withHTTPIdleTimeout(du time.Duration) endPointOption {
	return func(ep *endPoint) {
		if du > 0 {
			ep.httpIdleTimeout = du
		}
	}
}

func withHTTPTimeout(timeout time.Duration) endPointOption {
	return func(ep *endPoint) {
		if timeout > time.Duration(0) {
			ep.httpTimeout = timeout
		}
	}
}

func withProxy(proxy string) endPointOption {
	return func(ep *endPoint) {
		ep.proxy = proxy
	}
}

func newEndpoint(urlstr string, opts ...endPointOption) (*endPoint, error) {
	u, err := url.ParseRequestURI(urlstr)
	if err != nil {
		log.Errorf("parse dataway url %s failed: %s", urlstr, err.Error())
		return nil, err
	}

	ep := &endPoint{
		categoryURL: map[string]string{},
		token:       u.Query().Get("token"),
		host:        u.Host,
		scheme:      u.Scheme,
	}

	// apply options
	for _, opt := range opts {
		if opt != nil {
			opt(ep)
		}
	}

	for _, api := range ep.apis {
		if q := u.Query().Encode(); q != "" {
			ep.categoryURL[api] = fmt.Sprintf("%s://%s%s?%s",
				ep.scheme,
				ep.host,
				api,
				q)
		} else {
			ep.categoryURL[api] = fmt.Sprintf("%s://%s%s",
				ep.scheme,
				ep.host,
				api)
		}
	}

	switch ep.scheme {
	case "http", "https":
		if err := ep.setupHTTP(); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("not supported scheme %q", ep.scheme)
	}

	return ep, nil
}

func (ep *endPoint) setupHTTP() error {
	dialContext, err := dnet.GetDNSCacheDialContext(defaultDNSCacheFreq, defaultDNSCacheLookUpTimeout)
	if err != nil {
		log.Warnf("GetDNSCacheDialContext failed: %v", err)
		dialContext = nil // if failed, then not use dns cache.
	}

	cliopts := &httpcli.Options{
		DialTimeout:         ep.httpTimeout, // NOTE: should not use http timeout as dial timeout.
		MaxIdleConns:        ep.maxHTTPConnections,
		MaxIdleConnsPerHost: ep.maxHTTPIdleConnectionPerHost,
		IdleConnTimeout:     ep.httpIdleTimeout,
		DialContext:         dialContext,
	}

	if ep.proxy != "" { // set proxy
		if u, err := url.ParseRequestURI(ep.proxy); err != nil {
			log.Warnf("parse http proxy %q failed err: %s, ignored and no proxy set", ep.proxy, err.Error())
		} else {
			cliopts.ProxyURL = u
			log.Infof("set dataway proxy to %q ok", ep.proxy)
		}
	}

	ep.httpCli = httpcli.Cli(cliopts)
	ep.httpCli.Timeout = ep.httpTimeout
	return nil
}

func (ep *endPoint) writeBody(w *writer, b *body) {
	w.gzip = b.gzon

	// if send failed, do nothing.
	if err := ep.writePointData(b, w); err != nil {
		log.Warnf("send %d points to %q(gzip: %v) bytes failed: %q, ignored",
			len(w.pts), w.category, w.gzip, err.Error())

		// 4xx error do not cache data.
		// If the error is token-not-found or beyond-usage, datakit
		// will write all data to disk, this may cause unexpected I/O cost
		// on host.
		if errors.Is(err, errWritePoints4XX) {
			return
		}

		if w.fc == nil { // no cache
			return
		}

		// do cache: write them to disk.
		if w.cacheAll {
			if err := doCache(w, b); err != nil {
				log.Errorf("doCache %d pts on %s: %s", b.npts, w.category, err)
			} else {
				log.Infof("ok on doCache %d pts on %s", b.npts, w.category)
			}
		} else {
			//nolint:exhaustive
			switch w.category {
			case point.Metric, // these categories are not cache.
				point.MetricDeprecated,
				point.Object,
				point.CustomObject,
				point.DynamicDWCategory:

				log.Warnf("drop %d pts on %s, not cached", b.npts, w.category)

			default:
				if err := doCache(w, b); err != nil {
					log.Errorf("doCache %v pts on %s: %s", b.npts, w.category, err)
				}
			}
		}
	}
}

func (ep *endPoint) writePoints(w *writer) error {
	var (
		bodies []*body
		err    error
	)

	bodies, err = buildBody(w.pts, MaxKodoBody)
	if err != nil {
		return err
	}

	for _, body := range bodies {
		ep.writeBody(w, body)
	}

	return nil
}

func doCache(w *writer, b *body) error {
	if cachedata, err := pb.Marshal(&CacheData{
		Category:    int32(w.category),
		PayloadType: int32(b.payload),
		Payload:     b.buf,
	}); err != nil {
		return err
	} else {
		return w.fc.Put(cachedata)
	}
}

func (ep *endPoint) writePointData(b *body, w *writer) error {
	var (
		httpCodeStr = "unknown"
		httpCode    int
	)

	requrl, catNotFound := ep.categoryURL[w.category.URL()]

	if !catNotFound {
		if w.dynamicURL != "" {
			// for dialtesting, there are dynamic URL to post
			if _, err := url.ParseRequestURI(w.dynamicURL); err != nil {
				return err
			} else {
				log.Debugf("try use dynamic URL %s", w.dynamicURL)
				requrl = w.dynamicURL

				defer func() {
					// update dial-testing ok/fail info
					updateDTFailInfo(requrl, (httpCode/100 == 2))
				}()
			}
		} else {
			return fmt.Errorf("invalid url %s", w.dynamicURL)
		}
	}

	defer func() {
		// /v1/write/metric -> metric
		cat := w.category.String()

		if w.category == point.DynamicDWCategory {
			// NOTE: datakit category deprecated, we use point category
			cat = point.DynamicDWCategory.String()
		}

		bytesCounterVec.WithLabelValues(cat, "gzip", "total").Add(float64(len(b.buf)))
		bytesCounterVec.WithLabelValues(cat, "gzip", httpCodeStr).Add(float64(len(b.buf)))
		bytesCounterVec.WithLabelValues(cat, "raw", "total").Add(float64(b.rawLen))
		bytesCounterVec.WithLabelValues(cat, "raw", httpCodeStr).Add(float64(b.rawLen))

		ptsCounterVec.WithLabelValues(cat, "total").Add(float64(b.npts))
		ptsCounterVec.WithLabelValues(cat, httpCodeStr).Add(float64(b.npts))

		if w.isSinker {
			sinkPtsVec.WithLabelValues(cat, httpCodeStr).Add(float64(b.npts))
		}
	}()

	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(b.buf))
	if err != nil {
		log.Error(err)
		return err
	}

	if w.gzip {
		req.Header.Set("Content-Encoding", "gzip")
	}

	req.Header.Set("X-Points", fmt.Sprintf("%d", b.npts))

	resp, err := ep.sendReq(req)

	// NOTE: resp maybe not nil, we need HTTP status info to fill HTTP metrics before exit.
	if resp != nil {
		httpCodeStr = http.StatusText(resp.StatusCode)
		httpCode = resp.StatusCode
	}

	if err != nil {
		log.Errorf("sendReq: request url %s failed(proxy: %s): %s, resp: %v", requrl, ep.proxy, err, resp)
		return err
	}

	defer resp.Body.Close() //nolint:errcheck
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("ioutil.ReadAll: %s", err)
		return err
	}

	log.Debugf("post %d bytes to %s...", len(b.buf), requrl)

	switch resp.StatusCode / 100 {
	case 2:
		log.Debugf("post %d bytes to %s ok(gz: %v)", len(b.buf), requrl, w.gzip)

		// Send data ok, it means the error `beyond-usage` error is cleared by kodo server,
		// we have to clear the hint in monitor too.
		if strings.Contains(requrl, "/v1/write/") && atomic.LoadInt64(&metrics.BeyondUsage) > 0 {
			log.Info("clear BeyondUsage")
			atomic.StoreInt64(&metrics.BeyondUsage, 0)
		}

		return nil

	case 4:
		strBody := string(body)
		log.Errorf("post %d to %s failed(HTTP: %s): %s, data dropped",
			len(b.buf),
			requrl,
			resp.Status,
			strBody)

		switch resp.StatusCode {
		case http.StatusForbidden:
			if strings.Contains(strBody, "beyondDataUsage") {
				atomic.AddInt64(&metrics.BeyondUsage, time.Now().Unix()) // will set `beyond-usage' hint in monitor.
				log.Info("set BeyondUsage")
			}
		default:
			// pass
		}

		return errWritePoints4XX

	default: // 5xx
		log.Errorf("post %d to %s failed(HTTP: %s): %s",
			len(b.buf),
			requrl,
			resp.Status,
			string(body))

		return fmt.Errorf("dataway internal error")
	}
}

func (ep *endPoint) GetCategoryURL() map[string]string {
	return ep.categoryURL
}

func (ep *endPoint) getLogFilter() ([]byte, error) {
	url, ok := ep.categoryURL[datakit.LogFilter]
	if !ok {
		return nil, fmt.Errorf("LogFilter API missing, should not been here")
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ep.sendReq(req)
	if err != nil {
		log.Error(err.Error())

		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())

		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getLogFilter failed with status code %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (ep *endPoint) datakitPull(args string) ([]byte, error) {
	url, ok := ep.categoryURL[datakit.DatakitPull]
	if !ok {
		return nil, fmt.Errorf("datakit pull API missing, should not been here")
	}

	req, err := http.NewRequest(http.MethodGet, url+"&"+args, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ep.sendReq(req)
	if err != nil {
		log.Error(err.Error())

		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("datakitPull failed with status code %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (ep *endPoint) sendReq(req *http.Request) (resp *http.Response, err error) {
	status := "unknown"

	if err := retry.Do(
		func() error {
			defer func() {
				if err != nil {
					if body, err := req.GetBody(); err == nil {
						req.Body = body
					}
				}
			}()
			resp, err = ep.doSendReq(req)

			if err != nil {
				status = "unknown"
				return err
			}

			if resp.StatusCode/100 == 5 { // server-side error
				status = http.StatusText(resp.StatusCode)
				return fmt.Errorf("doSendReq: %s", resp.Status)
			}

			return nil
		},

		retry.Attempts(4),
		retry.Delay(time.Second*1),
		retry.OnRetry(func(n uint, err error) {
			log.Warnf("on %dth retry, error: %s", n, err)
			httpRetry.WithLabelValues(req.URL.Path, status).Inc()
		}),
	); err != nil {
		return resp, err
	}

	return resp, err
}

func (ep *endPoint) doSendReq(req *http.Request) (*http.Response, error) {
	log.Debugf("send request %q, proxy: %q, cli: %p, timeout: %s",
		req.URL.String(), ep.proxy, ep.httpCli.Transport, ep.httpTimeout)

	var (
		start       = time.Now()
		httpCodeStr = "unknown"
	)

	req.Header.Set("User-Agent", DatakitUserAgent)

	defer func() {
		urlPath := req.URL.Path
		// It's a bad-designed API path, we rename it in metrics.
		// the original URL is `/v1/check/token/tkn_xxxxxxxxxxxxxxxxxxx'
		if strings.HasPrefix(req.URL.Path, "/v1/check/token") {
			urlPath = "/v1/check/token"
		}

		apiSumVec.WithLabelValues(urlPath, httpCodeStr).
			Observe(float64(time.Since(start)) / float64(time.Second))
	}()

	if ep.httpTrace {
		s := httpcli.NewHTTPClientTraceStat("dataway")
		defer s.Metrics()

		req = req.WithContext(httptrace.WithClientTrace(req.Context(), s.Trace()))
	}

	var (
		resp *http.Response
		err  error
	)

	if httpFailRatio > 0 && (time.Since(httpFailStart) < httpFailDuration || int64(httpFailDuration) == 0) {
		if start.Unix()%100 < int64(httpFailRatio) {
			resp = httpMockedFailResp
			goto end
		}
	}

	resp, err = ep.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("httpCli.Do: %w, resp: %+#v", err, resp)
	}

end:
	if resp != nil {
		httpCodeStr = http.StatusText(resp.StatusCode)
	}

	return resp, nil
}
