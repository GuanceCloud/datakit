package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/gin-gonic/gin"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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
	datakit.Rum:               "POST",
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

func hello(w http.ResponseWriter, req *http.Request) {
	time.Sleep(time.Millisecond)
	fmt.Fprintf(w, "hello\n")
}

func TestClientTimeWait(t *testing.T) {
	http.HandleFunc("/hello", hello)

	server := &http.Server{
		Addr: ":8090",
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			t.Log(err)
		}
	}()

	time.Sleep(time.Second) // wait server ok

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
					req, err := http.NewRequest(m, "http://:8090"+u, nil)
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

	time.Sleep(time.Second * 8)
	if err := server.Shutdown(context.Background()); err != nil {
		t.Log(err)
	}
}
