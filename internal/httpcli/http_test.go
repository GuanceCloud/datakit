// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpcli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	tu "github.com/GuanceCloud/cliutils/testutil"
	"github.com/elazarl/goproxy"
	"github.com/gin-gonic/gin"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var reqs = map[string]string{
	datakit.MetricDeprecated:  "POST",
	datakit.Metric:            "POST",
	datakit.Network:           "POST",
	datakit.KeyEvent:          "POST",
	datakit.Object:            "POST",
	datakit.CustomObject:      "POST",
	datakit.Logging:           "POST",
	datakit.LogFilter:         "GET",
	datakit.Tracing:           "POST",
	datakit.RUM:               "POST",
	datakit.Security:          "POST",
	datakit.HeartBeat:         "POST",
	datakit.Election:          "POST",
	datakit.ElectionHeartbeat: "POST",
	datakit.QueryRaw:          "POST",
	datakit.ListDataWay:       "GET",
	datakit.ObjectLabel:       "POST",
}

func TestProxy(t *testing.T) {
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello, tls client")
	}))

	defer tlsServer.Close()

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	proxyAddr := "0.0.0.0:54321"

	proxysrv := &http.Server{
		Addr:    proxyAddr,
		Handler: proxy,
	}

	go func() {
		if err := proxysrv.ListenAndServe(); err != nil {
			t.Logf("%s", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	defer proxysrv.Shutdown(ctx) //nolint:errcheck

	time.Sleep(time.Second) // wait proxy server OK

	cases := []struct {
		cli  *http.Client
		fail bool
	}{
		{
			cli: Cli(&Options{
				InsecureSkipVerify: true,
				ProxyURL: func() *url.URL {
					//nolint:golint
					if u, err := url.Parse("http://" + proxyAddr); err != nil {
						t.Error(err)
						return nil
					} else {
						return u
					}
				}(),
			}),
		},

		{
			cli: Cli(&Options{
				InsecureSkipVerify: false,
				ProxyURL: func() *url.URL {
					u, err := url.Parse("http://" + proxyAddr)
					if err != nil { //nolint:golint
						t.Error(err)
						return nil
					}
					return u
				}(),
			}),
			fail: true,
		},
	}

	t.Logf("https request: %s", tlsServer.URL)
	for _, tc := range cases {
		resp, err := tc.cli.Get(tlsServer.URL)
		if tc.fail {
			tu.NotOk(t, err, "")
			t.Logf("error %s", err)
			continue
		} else {
			tu.Ok(t, err)

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Error(err)
			}

			defer resp.Body.Close() //nolint:errcheck
			t.Logf("resp: %s", string(body))
		}
	}
}

func TestInsecureSkipVerify(t *testing.T) {
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello, tls client")
	}))

	defer tlsServer.Close() //nolint:errcheck

	cases := []struct {
		cli  *http.Client
		fail bool
	}{
		{
			cli: Cli(&Options{
				InsecureSkipVerify: true,
			}),
		},

		{
			cli: Cli(&Options{
				InsecureSkipVerify: false,
			}),
			fail: true,
		},
	}

	for _, tc := range cases {
		resp, err := tc.cli.Get(tlsServer.URL)
		if tc.fail {
			tu.NotOk(t, err, "")
			t.Logf("error %s", err)
			continue
		} else {
			tu.Ok(t, err)

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Error(err)
			}

			defer resp.Body.Close() //nolint:errcheck
			t.Logf("resp: %s", string(body))
		}
	}
}

func httpGinServer(t *testing.T, host string) *http.Server {
	t.Helper()
	router := gin.New()

	for u, m := range reqs {
		if m == "GET" {
			router.GET(u, func(c *gin.Context) {})
		} else {
			router.POST(u, func(c *gin.Context) {})
		}
	}

	srv := &http.Server{
		Addr:    host,
		Handler: router,
	}

	retryCnt := 0
	go func() {
		for {
			if err := srv.ListenAndServe(); err != nil {
				if !errors.As(err, &http.ErrServerClosed) {
					t.Errorf("start server at %s failed: %s, retrying(%d)...",
						srv.Addr, err.Error(), retryCnt)
					retryCnt++
				} else {
					t.Logf("server(%s) stopped on: %s", srv.Addr, err.Error())
					break
				}
			}
			time.Sleep(time.Second)
		}
	}()
	return srv
}

func runTestClientConnections(t *testing.T, host string, tc *testClientConnectionsCase) {
	t.Helper()

	wg := sync.WaitGroup{}

	nreq := 100

	wg.Add(tc.nclients)
	for i := 0; i < tc.nclients; i++ {
		go func() {
			defer wg.Done()
			var cli *http.Client
			if tc.defaultOptions {
				cli = &http.Client{}
			} else {
				cli = Cli(nil)
			}

			for j := 0; j < nreq; j++ {
				for u, m := range reqs {
					req, err := http.NewRequest(m, host+u, nil)
					if err != nil {
						t.Error(err)
					}

					resp, err := cli.Do(req)
					if err != nil {
						t.Error(err)
					}

					if resp != nil {
						io.Copy(ioutil.Discard, resp.Body) //nolint:errcheck
						if err := resp.Body.Close(); err != nil {
							t.Error(err)
						}
					}
					if tc.closeIdleManually {
						cli.CloseIdleConnections()
					}
				}
			}
		}()
	}

	wg.Wait()
}

type testClientConnectionsCase struct {
	nclients          int
	defaultOptions    bool
	closeIdleManually bool
	ginServer         bool
}

func TestClientConnections(t *testing.T) {
	if runtime.GOOS != "linux" {
		return
	}

	cases := []testClientConnectionsCase{
		{10, false, false, true},
		{10, false, true, true},
		{10, true, false, true},
		{1, false, false, true},
		{1, false, true, true},
		{4, true, false, true},

		{10, false, false, false},
		{10, false, true, false},
		{10, true, false, false},
		{1, false, false, false},
		{1, false, true, false},
		{4, true, false, false},
	}

	ts := httptest.NewUnstartedServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(time.Millisecond)
			fmt.Fprintf(w, "{}")
		}))

	ts.Start()
	ginHost := "0.0.0.0:12345"
	ginserver := httpGinServer(t, ginHost)

	time.Sleep(time.Second) // wait HTTP server...

	for _, tc := range cases {
		var cw ConnWatcher

		host := ts.URL
		if tc.ginServer {
			host = "http://" + ginHost
			ginserver.ConnState = cw.OnStateChange
		} else {
			ts.Config.ConnState = cw.OnStateChange
		}

		runTestClientConnections(t, host, &tc)

		t.Logf("[server: %s, gin: %v, clients: %d, defaultOptions: %v, closeIdleManually: %v]\ncw: %s\n",
			ts.URL, tc.ginServer, tc.nclients, tc.defaultOptions, tc.closeIdleManually, cw.String())

		if tc.defaultOptions || tc.closeIdleManually {
			tu.Assert(t, cw.Max >= int64(tc.nclients),
				"by using default transport, %d should > %d",
				cw.Max, tc.nclients)
		} else {
			tu.Assert(t, cw.Max <= int64(tc.nclients),
				"by using specified transport, %d should == %d",
				cw.Max, tc.nclients)
		}
	}
}

func TestClientTimeWait(t *testing.T) {
	r := gin.New()
	r.GET("/hello", func(c *gin.Context) {
		time.Sleep(time.Millisecond)
		fmt.Fprintf(c.Writer, "hello\n")
	})

	ts := httptest.NewServer(r)
	defer ts.Close()
	time.Sleep(time.Second)

	n := 10
	wg := sync.WaitGroup{}
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()

			cli := Cli(&Options{ // new fresh client
				DialTimeout:           30 * time.Second,
				DialKeepAlive:         30 * time.Second,
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   n,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: time.Second,
			})

			for j := 0; j < 1; j++ {
				for u, m := range reqs {
					req, err := http.NewRequest(m, ts.URL+u, nil)
					if err != nil {
						t.Error(err)
					}

					resp, err := cli.Do(req)
					if err != nil {
						t.Error(err)
					}

					if resp != nil {
						io.Copy(ioutil.Discard, resp.Body) //nolint:errcheck

						if err := resp.Body.Close(); err != nil {
							t.Error(err)
						}
					}
				}
			}

			cli.CloseIdleConnections()
		}()
	}

	wg.Wait()
}

// test what error if client request timeout
func TestClientTimeout(t *testing.T) {
	r := gin.New()

	sec := 3

	r.GET("/test", func(c *gin.Context) {
		time.Sleep(time.Second * time.Duration(sec+1))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()
	time.Sleep(time.Second)

	cli := http.Client{
		Timeout: time.Second,
	}

	req, err := http.NewRequest("GET", ts.URL+"/test", nil)
	if err != nil {
		t.Error(err)
	}

	resp, err := cli.Do(req)
	if err != nil {
		t.Logf("Do: %s, type: %s, %+#v, Err: %+#v", err, reflect.TypeOf(err), err, err.(*url.Error).Err) //nolint:errorlint
	} else {
		defer resp.Body.Close()
	}
}

type slowReader struct {
	buf *bytes.Buffer
}

var sleepms = time.Duration(1000)

func (r *slowReader) Read(p []byte) (int, error) {
	if r.buf == nil {
		return 0, nil
	}

	// slow reader
	time.Sleep(sleepms * time.Millisecond) //nolint:durationcheck
	return r.buf.Read(p)
}

func TestServerTimeout(t *testing.T) {
	r := gin.New()

	r.GET("/test", func(c *gin.Context) {
		sr := &slowReader{buf: bytes.NewBufferString("response body")}
		c.DataFromReader(200, int64(sr.buf.Len()), "application/json", sr, nil)
	})

	ts := httptest.NewUnstartedServer(r)
	// ts.Config.ReadTimeout = sleepms * time.Millisecond // nolint:durationcheck
	ts.Config.WriteTimeout = (sleepms * 10) * time.Millisecond //nolint:durationcheck
	ts.Start()

	defer ts.Close()
	time.Sleep(time.Second)

	cli := http.Client{
		// Timeout: time.Second, not set
	}

	// req, err := http.NewRequest("GET", ts.URL+"/test", &slowReader{buf: bytes.NewBufferString("body string")})
	req, err := http.NewRequest("GET", ts.URL+"/test", nil)
	if err != nil {
		t.Error(err)
		return
	}

	start := time.Now()
	resp, err := cli.Do(req)
	if err != nil {
		t.Logf("Do: %s\ntype: %s, %+#v, Err: %+#v\ncost: %s", err, reflect.TypeOf(err), err, err.(*url.Error).Err, time.Since(start)) //nolint:errorlint
	} else {
		defer resp.Body.Close()
	}
}
