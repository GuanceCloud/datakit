// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	reflect "reflect"
	"strings"
	"sync/atomic"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/avast/retry-go"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpcli"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	dnet "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
)

type endPoint struct {
	token  string
	host   string
	scheme string

	httpHeaders,
	categoryURL map[string]string

	httpCli *http.Client

	// optionals
	proxy       string
	apis        []string
	httpTimeout time.Duration

	maxHTTPIdleConnectionPerHost int
	maxHTTPConnections           int
	httpIdleTimeout              time.Duration
	maxRetryCount                int
	retryDelay                   time.Duration

	insecureSkipVerify,
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

func withInsecureSkipVerify(on bool) endPointOption {
	return func(ep *endPoint) {
		ep.insecureSkipVerify = on
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

func withMaxRetryCount(count int) endPointOption {
	return func(ep *endPoint) {
		if count > 0 {
			if count > 10 {
				count = 10
			}
			ep.maxRetryCount = count
		}
	}
}

func withRetryDelay(delay time.Duration) endPointOption {
	return func(ep *endPoint) {
		if delay >= 0 {
			ep.retryDelay = delay
		}
	}
}

func withProxy(proxy string) endPointOption {
	return func(ep *endPoint) {
		ep.proxy = proxy
	}
}

func withHTTPHeaders(headers map[string]string) endPointOption {
	return func(ep *endPoint) {
		for k, v := range headers {
			if len(v) > 0 { // ignore empty header value
				ep.httpHeaders[k] = v
			} else {
				l.Warnf("ignore empty value on header %q", k)
			}
		}
	}
}

func newEndpoint(urlstr string, opts ...endPointOption) (*endPoint, error) {
	u, err := url.ParseRequestURI(urlstr)
	if err != nil {
		l.Errorf("parse dataway url %s failed: %s", urlstr, err.Error())
		return nil, err
	}

	ep := &endPoint{
		categoryURL: map[string]string{},
		httpHeaders: map[string]string{},
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

		l.Infof("endpoint regist dataway API %q:%q ok", api, ep.categoryURL[api])
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

func (ep *endPoint) getHTTPCliOpts() *httpcli.Options {
	dialContext, err := dnet.GetDNSCacheDialContext(defaultDNSCacheFreq, defaultDNSCacheLookUpTimeout)
	if err != nil {
		l.Warnf("GetDNSCacheDialContext failed: %v", err)
		dialContext = nil // if failed, then not use dns cache.
	}

	cliOpts := &httpcli.Options{
		DialTimeout:         ep.httpTimeout, // NOTE: should not use http timeout as dial timeout.
		MaxIdleConns:        ep.maxHTTPConnections,
		MaxIdleConnsPerHost: ep.maxHTTPIdleConnectionPerHost,
		IdleConnTimeout:     ep.httpIdleTimeout,
		DialContext:         dialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: ep.insecureSkipVerify, // nolint: gosec
		},
	}

	if ep.proxy != "" { // set proxy
		if u, err := url.ParseRequestURI(ep.proxy); err != nil {
			l.Warnf("parse http proxy %q failed err: %s, ignored and no proxy set", ep.proxy, err.Error())
		} else {
			if ProxyURLOK(u) {
				cliOpts.ProxyURL = u
				l.Infof("set dataway proxy to %q ok", ep.proxy)
			} else {
				l.Warnf("invalid proxy URL: %s, ignored", u)
			}
		}
	}

	return cliOpts
}

func ProxyURLOK(u *url.URL) bool {
	return u.Scheme == "https" || u.Scheme == "http"
}

func (ep *endPoint) setupHTTP() error {
	ep.httpCli = httpcli.Cli(ep.getHTTPCliOpts())
	ep.httpCli.Timeout = ep.httpTimeout

	// Do not override exit valid setting, but protect retry with valid default settings.
	if ep.maxRetryCount <= 0 {
		ep.maxRetryCount = DefaultRetryCount
	}

	return nil
}

func (ep *endPoint) Transport() *http.Transport {
	return httpcli.Transport(ep.getHTTPCliOpts())
}

func (ep *endPoint) writePoints(w *writer) error {
	WithBodyCallback(ep.writePointData)(w)
	return w.buildPointsBody()
}

func (ep *endPoint) writePointData(w *writer, b *body) error {
	httpCodeStr := "unknown"
	requrl, catNotFound := ep.categoryURL[b.cat().URL()]

	if !catNotFound {
		l.Debugf("cat %q not found, w.dynamicURL: %s", b.cat(), w.dynamicURL)

		if w.dynamicURL != "" {
			// for dialtesting, there are dynamic URL to post
			if _, err := url.ParseRequestURI(w.dynamicURL); err != nil {
				return err
			} else {
				l.Debugf("try use dynamic URL %s", w.dynamicURL)
				requrl = w.dynamicURL
			}
		} else {
			return fmt.Errorf("invalid url %s", w.dynamicURL)
		}
	}

	defer func() {
		if w.cacheClean { // ignore metrics on cache clean operation
			l.Debug("on cache clean, no metric applied")
			return
		}

		// /v1/write/metric -> metric
		cat := b.cat().String()

		if b.cat() == point.DynamicDWCategory {
			// NOTE: datakit category deprecated, we use point category
			cat = point.DynamicDWCategory.String()
		}

		bytesCounterVec.WithLabelValues(cat, "gzip", "total").Add(float64(len(b.buf())))
		bytesCounterVec.WithLabelValues(cat, "gzip", httpCodeStr).Add(float64(len(b.buf())))
		bytesCounterVec.WithLabelValues(cat, "raw", "total").Add(float64(b.rawLen()))
		bytesCounterVec.WithLabelValues(cat, "raw", httpCodeStr).Add(float64(b.rawLen()))

		if b.npts() > 0 {
			ptsCounterVec.WithLabelValues(cat, "total").Add(float64(b.npts()))
			ptsCounterVec.WithLabelValues(cat, httpCodeStr).Add(float64(b.npts()))
		} else {
			l.Warnf("npts not set, body from %q", b.from)
		}
	}()

	l.Debugf("post %d bytes to %s...", len(b.buf()), requrl)
	req, err := http.NewRequest("POST", requrl, bytes.NewBuffer(b.buf()))
	if err != nil {
		l.Error("new request to %s: %s", requrl, err)
		return err
	}

	req.Header.Set("X-Points", fmt.Sprintf("%d", b.npts()))
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(b.buf())))
	req.Header.Set("Content-Type", b.enc().HTTPContentType())
	if w.gzip == 1 {
		req.Header.Set("Content-Encoding", "gzip")
	}

	// add package id
	req.Header.Set("X-Pkg-Id", cliutils.XID("dk_"))

	// Common HTTP headers appended, such as User-Agent, X-Global-Tags
	for k, v := range ep.httpHeaders {
		l.Debugf("set %s:%s HTTP header comes from endpoint", k, v)
		req.Header.Set(k, v)
	}

	// Append extra HTTP headers to request.
	// Here may attach X-Global-Tags again.
	for k, v := range w.httpHeaders {
		l.Debugf("set %s:%s HTTP header comes from writer", k, v)
		req.Header.Set(k, v)
	}

	resp, err := ep.sendReq(req)
	// NOTE: resp maybe not nil, we need HTTP status info to fill HTTP metrics before exit.
	if resp != nil {
		httpCodeStr = http.StatusText(resp.StatusCode)
	}

	if err != nil {
		l.Errorf("sendReq: request url %s failed(proxy: %s): %s, resp: %v", requrl, ep.proxy, err, resp)
		// do not return here, we need more details about the fail from @resp.
	}

	if resp == nil {
		return errRequestTerminated
	}

	defer resp.Body.Close() //nolint:errcheck
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("io.ReadAll: %s", err)
		return err
	}

	switch resp.StatusCode / 100 {
	case 2:
		l.Debugf("post %d bytes to %s ok(gz: %v)", len(b.buf()), requrl, w.gzip)

		// Send data ok, it means the error `beyond-usage` error is cleared by kodo server,
		// we have to clear the hint in monitor too.
		if strings.Contains(requrl, "/v1/write/") && atomic.LoadInt64(&metrics.BeyondUsage) > 0 {
			l.Info("clear BeyondUsage")
			atomic.StoreInt64(&metrics.BeyondUsage, 0)
		}

		return nil

	case 4:
		strBody := string(body)
		l.Errorf("post %d to %s failed(HTTP: %s): %s, data dropped",
			len(b.buf()),
			requrl,
			resp.Status,
			strBody)

		switch resp.StatusCode {
		case http.StatusForbidden:
			if strings.Contains(strBody, "beyondDataUsage") {
				atomic.AddInt64(&metrics.BeyondUsage, time.Now().Unix()) // will set `beyond-usage' hint in monitor.
				l.Info("set BeyondUsage")
			}
		default:
			// pass
		}

		return errWritePoints4XX

	default: // 5xx
		l.Errorf("post %d to %s failed(HTTP: %s): %s",
			len(b.buf()),
			requrl,
			resp.Status,
			string(body))

		return fmt.Errorf("dataway internal error")
	}
}

func (ep *endPoint) GetCategoryURL() map[string]string {
	return ep.categoryURL
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

	// Common HTTP headers appended, such as User-Agent, X-Global-Tags
	for k, v := range ep.httpHeaders {
		req.Header.Set(k, v)
	}

	resp, err := ep.sendReq(req)
	if err != nil {
		l.Errorf("datakitPull: %s", err.Error())

		return nil, err
	}

	if resp == nil {
		return nil, errRequestTerminated
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error(err.Error())
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

	// Generally, the req.GetBody in DK should not be nil, while we do this to avoid accidents.
	if ep.maxRetryCount > 1 && req.GetBody == nil && req.Body != nil {
		l.Debugf("setup GetBody() on %q", req.URL.Path)

		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("unable to read body: %w", err)
		}

		if len(b) > 0 {
			req.Body = io.NopCloser(bytes.NewReader(b))
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(b)), nil
			}
		}
	}

	delay := ep.retryDelay

	// We must set retry > 0, or the request will fail immediately.
	maxRetry := uint(ep.maxRetryCount)
	if maxRetry == 0 {
		maxRetry = DefaultRetryCount
	}

	l.Debugf("retry %q with delay %s on %d retrying", req.URL.Path, delay, maxRetry)

	if err := retry.Do(
		func() error {
			defer func() {
				if err == nil {
					return
				}

				if req.GetBody == nil {
					l.Debugf("GetBody() not set for request %q, ignored", req.URL.Path)
					return
				}

				if body, ierr := req.GetBody(); ierr == nil {
					req.Body = body // reset body reader, then we can send the request again.
				} else {
					l.Errorf("GetBody() on %q failed: %s", req.URL.Path, ierr)
				}
			}()

			if resp, err = ep.doSendReq(req); err != nil {
				return err
			}

			if resp.StatusCode/100 == 5 { // server-side error
				status = http.StatusText(resp.StatusCode)
				// Terminate retry on global exit.
				select {
				case <-datakit.Exit.Wait():
					l.Info("retry abort on global exit")
					return nil

				default: // pass
				}
				err = fmt.Errorf("doSendReq: %s", resp.Status)
				return err
			}

			return nil
		},

		retry.Attempts(maxRetry),
		retry.Delay(delay),

		retry.OnRetry(func(n uint, err error) {
			l.Warnf("on %dth retry for %s, error: %s(%s)", n, req.URL, err, reflect.TypeOf(err))

			switch {
			// most of the error is Client.Timeout
			case strings.Contains(err.Error(), "Timeout"):
				status = "timeout"
			default: // Pass
			}

			httpRetry.WithLabelValues(req.URL.Path, status).Inc()
		}),
	); err != nil {
		l.Errorf("retry.Do: %s", err.Error())

		switch {
		case strings.Contains(err.Error(), "All attempts fail"):
			return resp, fmt.Errorf("all-retry-failed")
		default:
			return resp, fmt.Errorf("retry request err: %w", err)
		}
	}

	return resp, err
}

func (ep *endPoint) doSendReq(req *http.Request) (*http.Response, error) {
	l.Debugf("send request %q, proxy: %q, cli: %p, timeout: %s",
		req.URL.String(), ep.proxy, ep.httpCli.Transport, ep.httpTimeout)

	var (
		start       = time.Now()
		httpCodeStr = "unknown"
	)

	defer func() {
		urlPath := req.URL.Path
		// It's a bad-designed API path, we rename it in prometheus metrics.
		// the original URL is `/v1/check/token/tkn_xxxxxxxxxxxxxxxxxxx'
		if strings.HasPrefix(req.URL.Path, "/v1/check/token") {
			urlPath = "/v1/check/token"
		}

		apiSumVec.WithLabelValues(urlPath, httpCodeStr).
			Observe(float64(time.Since(start)) / float64(time.Second))
	}()

	if ep.httpTrace {
		s := httpcli.NewHTTPClientTraceStat("dataway", "")
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
		// To check if the error is a timeout.
		if ue, ok := err.(*url.Error); ok { // nolint: errorlint
			switch {
			case ue.Timeout():
				httpCodeStr = "timeout"
			case strings.Contains(ue.Error(), "reset by peer"):
				httpCodeStr = "reset-by-pear"
			case strings.Contains(ue.Error(), "connection refused"):
				httpCodeStr = "connection-refused"
			case strings.Contains(ue.Error(), "network is unreachable"):
				httpCodeStr = "network-is-unreachable"
			default:
				l.Warnf("unwrapped URL error: %s", err.Error())
				httpCodeStr = "unwrapped-url-error"
			}
		}

		l.Warnf("Do: %s, error type: %s", err.Error(), reflect.TypeOf(err))

		return nil, fmt.Errorf("httpCli.Do: %w, resp: %+#v", err, resp)
	}
	l.Debugf("%s send req ok", req.URL)

end:
	if resp != nil {
		httpCodeStr = http.StatusText(resp.StatusCode)
	}

	return resp, nil
}
