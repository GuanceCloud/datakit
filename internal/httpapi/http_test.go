// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/gin-gonic/gin"
	"github.com/influxdata/influxdb1-client/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

func TestMetricsAPI(t *T.T) {
	t.Run("/metric", func(t *T.T) {
		r := setupRouter()
		ts := httptest.NewServer(r)

		defer ts.Close()

		vec := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "abc_total",
				Help: "not-set",
			},
			[]string{
				"category",
			},
		)
		metrics.MustRegister(vec)
	})
}

func TestParsePoint(t *T.T) {
	cases := []struct {
		body []byte
		prec string
		npts int
		fail bool
	}{
		{
			body: []byte(`m1,t1=abc f1=123 123`),
			prec: "s",
			npts: 1,
		},

		{
			body: []byte(`m1,t1=abc f1=123 123`),
			prec: "ms",
			npts: 1,
		},

		{
			body: []byte(`m1,t1=abc f1=123 123`),
			prec: "u",
			npts: 1,
		},

		{
			body: []byte(`m1,t1=abc f1=123 123`),
			prec: "n",
			npts: 1,
		},
	}

	for _, tc := range cases {
		pts, err := models.ParsePointsWithPrecision(tc.body, time.Now(), tc.prec)
		if tc.fail {
			assert.Error(t, err)
		} else {
			assert.Equal(t, tc.npts, len(pts))
			for _, pt := range pts {
				t.Log(pt.String())
			}
		}
	}
}

func TestRestartAPI(t *T.T) {
	urls := []string{
		"http://1.2.3.4?token=tkn_abc123",
		"http://4.3.2.1?token=tkn_abc456",
	}

	dw = &dataway.Dataway{URLs: urls}
	assert.NoError(t, dw.Init())

	cases := []struct {
		token string
		fail  bool
	}{
		{
			token: "tkn_abc123",
			fail:  false,
		},
		{
			token: "tkn_abc456",
			fail:  true,
		},
		{
			token: "",
			fail:  true,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := checkToken(r); err != nil {
			w.WriteHeader(ErrInvalidToken.HttpCode)
			if err := json.NewEncoder(w).Encode(err); err != nil {
				t.Error(err)
			}
		} else {
			w.WriteHeader(200)
		}
	}))

	defer ts.Close() //nolint:errcheck

	time.Sleep(time.Second)

	for _, tc := range cases {
		resp, err := http.Post(fmt.Sprintf("%s?token=%s", ts.URL, tc.token), "", nil)
		if err != nil {
			t.Errorf("error: %s", err)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
			continue
		}
		resp.Body.Close() //nolint:errcheck

		if !tc.fail {
			assert.Equal(t, 200, resp.StatusCode)
		} else {
			assert.Equal(t, ErrInvalidToken.HttpCode, resp.StatusCode)
		}

		t.Logf("resp: %s", string(body))
	}
}

func TestApiGetDatakitLastError(t *T.T) {
	const uri string = "/v1/lasterror"

	cases := []struct {
		body []byte
		fail bool
	}{
		{
			[]byte(`{"input":"fakeCPU","err_content":"cpu has broken down"}`),
			false,
		},
		{
			[]byte(`{"input":"fakeCPU","err_content":""}`),
			true,
		},
		{
			[]byte(`{"input":"","err_content":"cpu has broken down"}`),
			true,
		},
		{
			[]byte(`{"input":"","err_content":""}`),
			true,
		},
		{
			[]byte(`{"":"fakeCPU","err_content":"cpu has broken down"}`),
			true,
		},
		{
			[]byte(`{"input":"fakeCPU","":"cpu has broken down"}`),
			true,
		},
		{
			[]byte(`{"":"fakeCPU","":"cpu has broken down"}`),
			true,
		},
		{
			[]byte(``),
			true,
		},
	}

	for _, fakeError := range cases {
		fakeEM := errMessage{}
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", uri, bytes.NewReader(fakeError.body))
		if err != nil {
			t.Errorf("create newrequest failed:%s", err)
		}
		em, err := doAPIGetDatakitLastError(req, rr)
		if err != nil {
			if fakeError.fail {
				t.Logf("expect error: %s", err)
				continue
			}
			t.Errorf("api test failed:%s", err)
		}
		err = json.Unmarshal(fakeError.body, &fakeEM)
		if err != nil {
			t.Errorf("json.Unmarshal: %s", err)
		}
		assert.Equal(t, fakeEM.ErrContent, em.ErrContent)
		assert.Equal(t, fakeEM.Input, em.Input)
	}
}

func TestCORS(t *T.T) {
	router := setupRouter()

	router.POST("/timeout", func(c *gin.Context) {})

	ts := httptest.NewServer(router)
	defer ts.Close()

	time.Sleep(time.Second)

	req, err := http.NewRequest("POST", ts.URL+"/some-404-page", nil)
	if err != nil {
		t.Error(err)
	}

	origin := "http://foobar.com"
	req.Header.Set("Origin", origin)

	c := &http.Client{}

	resp, err := c.Do(req)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	// See: https://stackoverflow.com/a/12179364/342348
	got := resp.Header.Get("Access-Control-Allow-Origin")
	assert.Equal(t, origin, got, "expect %s, got '%s'", origin, got)
}

func TestTimeout(t *T.T) {
	apiServer.timeout = 100 * time.Millisecond

	router := gin.New()
	router.Use(dkHTTPTimeout())

	ts := httptest.NewServer(router)
	defer ts.Close()

	router.POST("/timeout", func(c *gin.Context) {
		x := c.Query("x")
		du, err := time.ParseDuration(x)
		if err != nil {
			du = 10 * time.Millisecond
		}

		time.Sleep(du)
		c.Status(http.StatusOK)
	})

	cases := []struct {
		name             string
		timeout          time.Duration
		expectStatusCode int
	}{
		{
			name:             "timeout",
			timeout:          105 * time.Millisecond,
			expectStatusCode: http.StatusRequestTimeout,
		},

		{
			name:             "ok",
			timeout:          10 * time.Millisecond,
			expectStatusCode: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			req, err := http.NewRequest("POST", fmt.Sprintf("%s/timeout?x=%s", ts.URL, tc.timeout), nil)
			if err != nil {
				t.Errorf("http.NewRequest: %s", err)
				return
			}

			cli := http.Client{}
			resp, err := cli.Do(req)
			if err != nil {
				t.Errorf("cli.Do: %s", err)
				return
			}

			assert.Equal(t, tc.expectStatusCode, resp.StatusCode)

			defer resp.Body.Close()

			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("cli.Do: %s", err)
				return
			}

			t.Logf("body: %s", string(respBody))
		})
	}
}

func setulimit() {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println("Error Getting Rlimit ", err)
	}
	fmt.Println(rLimit)
	rLimit.Max = 999999
	rLimit.Cur = 999999
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println("Error Setting Rlimit ", err)
	}
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Println("Error Getting Rlimit ", err)
	}
}

func TestTimeoutOnConcurrentIdleTCPConnection(t *T.T) {
	setulimit()
	router := gin.New()

	idleSec := time.Duration(1)

	ts := httptest.NewUnstartedServer(router)
	ts.Config.ReadTimeout = idleSec * time.Second //nolint:durationcheck
	ts.Start()
	defer ts.Close()

	router.POST("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	tcpserver := ts.Listener.Addr().String()
	tcpAddr, err := net.ResolveTCPAddr("tcp", tcpserver)
	if err != nil {
		t.Error(err)
	}
	t.Logf("server: %s", tcpserver)

	n := 4096
	wg := &sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// start tcp client connect to HTTP server
			conn, err := net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				t.Logf("Dial %s failed: %s", tcpserver, err.Error())
				return
			}

			defer conn.Close() //nolint:errcheck

			// idle and timeout
			time.Sleep((idleSec + 3) * time.Second) //nolint:durationcheck

			closed := false

			for {
				_, err := conn.Write([]byte(`POST /timeout?x=101ms HTTP/1.1
Host: stackoverflow.com

nothing`))
				if err != nil {
					closed = true
					break
				}
				time.Sleep(time.Second)
			}

			assert.True(t, closed, "expect closed, but not")
		}()
	}

	wg.Wait()
}

//nolint:durationcheck
func TestTimeoutOnIdleTCPConnection(t *T.T) {
	router := gin.New()

	idleSec := time.Duration(3)

	ts := httptest.NewUnstartedServer(router)
	ts.Config.ReadTimeout = idleSec * time.Second
	ts.Start()
	defer ts.Close()

	router.POST("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	tcpserver := ts.Listener.Addr().String()
	tcpAddr, err := net.ResolveTCPAddr("tcp", tcpserver)
	if err != nil {
		t.Error(err)
	}
	t.Logf("server: %s", tcpserver)

	// start tcp client connect to HTTP server
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		t.Errorf("Dial %s failed: %s", tcpserver, err.Error())
	}

	time.Sleep((idleSec + 1) * time.Second) // idle and timeout

	closed := false

	for {
		n, err := conn.Write([]byte(`POST /timeout?x=101ms HTTP/1.1
Host: stackoverflow.com

nothing`))

		if err != nil {
			t.Logf("send %d bytes failed: %s", n, err)
			closed = true
			break
		} else {
			t.Logf("send %d bytes ok", n)
		}
		time.Sleep(time.Second)
	}

	assert.True(t, closed, "expect closed, but not")
}

// go test -v -timeout 30s -run ^TestParseListen$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi
func TestInitListener(t *T.T) {
	cases := []struct {
		name           string
		lsn            string
		expectListener bool
	}{
		// {
		// 	name: "ipv6-loopback",
		// 	lsn:  "[::1]:0",
		// },
		{
			name: "loopback-127.0.0.1-ipv4",
			lsn:  "127.0.0.1:0",
		},

		{
			name: "loopback-localhost-ipv4",
			lsn:  "localhost:0",
		},
		{
			name: "0000-ipv4",
			lsn:  "0.0.0.0:0",
		},
		{
			name: "unix file tmp",
			lsn:  filepath.Join(t.TempDir(), "datakit.sock"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			listener, err := initListener(tc.lsn)
			require.NoError(t, err)

			t.Logf("listener: %s", listener.Addr())

			t.Cleanup(func() {
				listener.Close() //nolint: errcheck
			})
		})
	}
}

func TestHTTPListers(t *T.T) {
	t.Run("domain-socket", func(t *T.T) {
		// To avoid 104-byte-len-of-unix-domain-socket, see:
		//  https://unix.stackexchange.com/questions/367008/why-is-socket-path-length-limited-to-a-hundred-chars
		os.Setenv("TMPDIR", "/tmp/")

		uds := filepath.Join(t.TempDir(), "datakit.sock")
		l, err := initListener(uds)
		require.NoError(t, err, "initListener: %s, len: %d", err, len(uds))

		t.Logf("uds: %s", uds)

		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		}))

		ts.Listener = l
		ts.Start()
		defer ts.Close() //nolint:errcheck

		time.Sleep(time.Second) // wait ts ok

		c := http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", uds)
				},
			},
		}

		urlStr := fmt.Sprintf("http://unix%s", uds)
		resp, err := c.Get(urlStr)
		require.NoError(t, err, "request %q failed: %s", urlStr, err)
		require.Equal(t, 2, resp.StatusCode/100)

		t.Logf("request %s ok", urlStr)

		t.Cleanup(func() {
			os.Unsetenv("TMPDIR")
		})
	})

	// t.Run("loopback-v6", func(t *T.T) {
	// 	l, err := initListener("[::1]:0")
	// 	require.NoError(t, err)

	// 	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 		w.WriteHeader(200)
	// 	}))

	// 	ts.Listener = l
	// 	ts.Start()
	// 	defer ts.Close() //nolint:errcheck

	// 	time.Sleep(time.Second) // wait ts ok

	// 	resp, err := http.Get(ts.URL)
	// 	require.NoError(t, err)
	// 	require.Equal(t, 2, resp.StatusCode/100)

	// 	t.Logf("request %s ok", ts.URL)
	// })

	t.Run("loopback-v4", func(t *T.T) {
		l, err := initListener("localhost:0")
		require.NoError(t, err)

		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
		}))

		ts.Listener = l
		ts.Start()
		defer ts.Close() //nolint:errcheck

		time.Sleep(time.Second) // wait ts ok

		resp, err := http.Get(ts.URL)
		require.NoError(t, err)
		require.Equal(t, 2, resp.StatusCode/100)

		t.Logf("request %s ok", ts.URL)
	})
}
