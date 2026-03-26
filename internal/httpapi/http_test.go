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
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	T "testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/influxdata/influxdb1-client/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

func TestMetricsAPI(t *T.T) {
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

	hs := defaultHTTPServerConf()

	hs.dw = dataway.NewDefaultDataway()
	hs.dw.URLs = urls
	assert.NoError(t, hs.dw.Init())

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
		if err := checkTokens(hs.dw, r); err != nil {
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

		body, err := io.ReadAll(resp.Body)
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

func TestCORS(t *T.T) {
}

// slowReader wraps a standard reader and adds a delay to every Read call
type slowReader struct {
	r     io.Reader
	delay time.Duration
}

func (s *slowReader) Read(p []byte) (n int, err error) {
	if s.delay > 0 {
		time.Sleep(s.delay) // The delay happens here!
	}

	// For demo purposes, we only read 1 byte at a time to force many Read calls
	// and make the total duration longer.
	if len(p) > 0 {
		return s.r.Read(p[:1])
	}
	return s.r.Read(p)
}

func Test_serverSideTimeout(t *T.T) {
	hs := defaultHTTPServerConf()
	hs.apiConfig.ReadTimeout = 100 * time.Millisecond
	hs.apiConfig.IdleTimeout = time.Second
	hs.apiConfig.ReadHeaderTimeout = time.Second * 2
	hs.apiConfig.WriteTimeout = time.Second * 2

	router := gin.New()

	router.POST("/timeout", func(c *gin.Context) {
		defer c.Request.Body.Close()

		_, err := io.ReadAll(c.Request.Body)
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				c.Status(http.StatusRequestTimeout)
			} else {
				c.Status(http.StatusInternalServerError) // should not been here
			}

			t.Logf("ReadAll: %s", err.Error())
			return
		}

		if f, ok := c.Writer.(http.Flusher); ok {
			f.Flush()
		}

		chunk := make([]byte, 64*1024)

		for i := 0; i < 1000; i++ {
			_, err := c.Writer.Write(chunk)
			if err != nil {
				t.Logf("[%d] server write failed: (timeout triggered): %s", i, err.Error())
				return
			}
		}

		c.Status(http.StatusOK)
	})

	ts := httptest.NewUnstartedServer(router)
	ts.Config = hs.setupServer(ts.Config)

	ts.Start()
	defer ts.Close()

	t.Run("write-timeout", func(t *T.T) {
		// mock a slow read client
		conn, err := net.Dial("tcp", ts.Listener.Addr().String())
		require.NoError(t, err)

		defer conn.Close()

		req := fmt.Sprintf("POST /timeout HTTP/1.1\r\nHost: %s\r\n\r\n", ts.Listener.Addr().String())
		_, err = conn.Write([]byte(req))
		require.NoError(t, err)

		headerBuf := make([]byte, 4096)
		_, err = conn.Read(headerBuf)
		require.NoError(t, err)

		time.Sleep(time.Second * 3) // stop reading, trigger server side wirte timeout

		// check if read fail
		conn.SetReadDeadline(time.Now().Add(time.Second))
		buf := make([]byte, 1024)

		i := 0
		for {
			// because TCP buffering, we should read more times to consume server's send buffer.
			_, err = conn.Read(buf)
			if err != nil {
				t.Logf("[%d] error: %s", i, err.Error())
				require.Equal(t, "EOF", err.Error())
				break
			}
			i++
		}
	})

	t.Run("read-header-timeout", func(t *T.T) {
		conn, err := net.Dial("tcp", ts.Listener.Addr().String())
		require.NoError(t, err)

		defer conn.Close()

		httpParts := []string{
			"POST / HTTP/1.1\r\n",
			"HOST: " + ts.Listener.Addr().String() + "\r\n",
			"User-Agent: Slow-Header-Client\r\n",
		}

		delay := time.Second * 1
		for _, x := range httpParts {
			_, err := conn.Write([]byte(x))
			require.NoError(t, err)
			t.Logf("[%s] -> send: %q, sleeping %v...", time.Now(), x, delay)
			time.Sleep(delay)
		}

		time.Sleep(time.Second)

		// write body
		_, err = conn.Write([]byte("\r\n"))
		assert.Error(t, err)
		t.Logf("write body error: %s", err)
		if errors.Is(err, syscall.ECONNRESET) {
			t.Logf("err: connection reset")
		} else if errors.Is(err, syscall.EPIPE) {
			t.Logf("err: broken pipe")
		}
	})

	t.Run("idle-timeout", func(t *T.T) {
		conn, err := net.Dial("tcp", ts.Listener.Addr().String())
		require.NoError(t, err)

		defer conn.Close()

		time.Sleep(time.Second * 2) // idle

		buf := make([]byte, 1024)
		_, err = conn.Read(buf)
		require.Error(t, err)
		require.Equal(t, "EOF", err.Error())

		if err != nil {
			t.Logf("read: %s", err)
		}
	})

	cases := []struct {
		name             string
		timeout          time.Duration
		expectStatusCode int
	}{
		{
			name:             "read-timeout",
			timeout:          105 * time.Millisecond,
			expectStatusCode: http.StatusRequestTimeout,
		},

		{
			// no timeout
			name:             "ok",
			expectStatusCode: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *T.T) {
			slowbody := &slowReader{
				r:     bytes.NewBuffer([]byte("this is a very long string that will send very slow...")),
				delay: tc.timeout,
			}

			req, err := http.NewRequest("POST", fmt.Sprintf("%s/timeout?x=%s", ts.URL, tc.timeout), slowbody)
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

			_, err = io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("cli.Do: %s", err)
				return
			}
		})
	}
}

func setulimit(t *T.T) {
	t.Helper()

	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		t.Log("Error Getting Rlimit ", err)
	}

	t.Log(rLimit)

	rLimit.Max = 999999
	rLimit.Cur = 999999
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		t.Log("Error Setting Rlimit ", err)
	}
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		t.Log("Error Getting Rlimit ", err)
	}
}

func TestTimeoutOnConcurrentIdleTCPConnection(t *T.T) {
	setulimit(t)
	router := gin.New()

	idleDuration := time.Duration(1)

	ts := httptest.NewUnstartedServer(router)
	ts.Config.ReadTimeout = idleDuration * time.Second //nolint:durationcheck
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
			time.Sleep((idleDuration + 3) * time.Second) //nolint:durationcheck

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

	idleDuration := time.Duration(3)

	ts := httptest.NewUnstartedServer(router)
	ts.Config.ReadTimeout = idleDuration * time.Second
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

	time.Sleep((idleDuration + 1) * time.Second) // idle and timeout

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

func TestIsNil(t *T.T) {
	t.Run("nil", func(t *T.T) {
		var arr []int
		assert.False(t, IsNil(arr))

		var dict map[int]int
		assert.False(t, IsNil(dict))

		assert.True(t, IsNil(nil))

		var x any
		assert.True(t, IsNil(x))

		type obj struct{}
		var ptr *obj
		assert.True(t, IsNil(ptr))

		type callback func()
		var fn callback
		assert.False(t, IsNil(fn))
	})
}

func newPubListener(t *T.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "0.0.0.0:0") // listen for public access

	require.NoError(t, err)
	return ln
}

func TestSetuptRouter(t *T.T) {
	t.Run("api-white-list-on-xforward", func(t *T.T) {
		hs := defaultHTTPServerConf()
		hs.apiConfig = &config.APIConfig{
			PublicAPIs: []string{"/enabled_api"},
		}

		r := setupRouter(hs)
		r.POST("/disabled_api", func(c *gin.Context) {
			c.Data(http.StatusOK, "", []byte("disabled api ok"))
		})
		r.POST("/enabled_api", func(c *gin.Context) {
			c.Data(http.StatusOK, "", []byte("enabled api ok"))
		})

		ts := httptest.NewUnstartedServer(r)
		ts.Listener = newPubListener(t)
		ts.Start()
		defer ts.Close()

		t.Logf("test server: %q", ts.URL)
		time.Sleep(time.Second)

		// use non-loopback IP to request
		ip, err := datakit.GetFirstGlobalUnicastIP()
		assert.NoError(t, err)
		tcpAddr := &net.TCPAddr{
			IP: net.ParseIP(ip),
		}

		transport := &http.Transport{
			Dial: (&net.Dialer{LocalAddr: tcpAddr}).Dial,
		}

		cli := &http.Client{
			Transport: transport,
		}

		// x-Forwarded-for
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/disabled_api", ts.URL), nil)

		req.Header.Set("X-Forwarded-For", "127.0.0.1")

		assert.NoError(t, err)
		resp, err := cli.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		assert.Contains(t, string(respBody), "datakit.publicAccessDisabled")

		t.Logf("resp body: %s", string(respBody))

		// x-real-ip
		req, err = http.NewRequest("POST", fmt.Sprintf("%s/disabled_api", ts.URL), nil)

		req.Header.Set("X-Real-IP", "127.0.0.1")

		assert.NoError(t, err)
		resp, err = cli.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)

		respBody, err = io.ReadAll(resp.Body)
		assert.NoError(t, err)

		assert.Contains(t, string(respBody), "datakit.publicAccessDisabled")

		t.Logf("resp body: %s", string(respBody))

		// forward ip not loopback are blocked
		req, err = http.NewRequest("POST", fmt.Sprintf("%s/disabled_api", ts.URL), nil)
		req.Header.Set("X-Real-IP", "1.1.1.1")

		assert.NoError(t, err)
		resp, err = cli.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)

		respBody, err = io.ReadAll(resp.Body)
		assert.NoError(t, err)

		t.Logf("resp body: %s", string(respBody))

		// real loopback
		req, err = http.NewRequest("POST", fmt.Sprintf("%s/disabled_api", ts.URL), nil)
		assert.NoError(t, err)
		localcli := http.Client{}
		resp, err = localcli.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		respBody, err = io.ReadAll(resp.Body)
		assert.NoError(t, err)

		t.Logf("resp body: %s", string(respBody))

		// on enabled-api
		req, err = http.NewRequest("POST", fmt.Sprintf("%s/enabled_api", ts.URL), nil)

		req.Header.Set("X-Real-IP", "1.1.1.1")

		assert.NoError(t, err)
		resp, err = cli.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		respBody, err = io.ReadAll(resp.Body)
		assert.NoError(t, err)

		t.Logf("resp body: %s", string(respBody))

		// faked loopback on enabled-api
		req, err = http.NewRequest("POST", fmt.Sprintf("%s/enabled_api", ts.URL), nil)

		req.Header.Set("X-Real-IP", "localhost")

		assert.NoError(t, err)
		resp, err = cli.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		respBody, err = io.ReadAll(resp.Body)
		assert.NoError(t, err)

		t.Logf("resp body: %s", string(respBody))
	})

	t.Run("api-white-list", func(t *T.T) {
		hs := defaultHTTPServerConf()
		hs.apiConfig = &config.APIConfig{
			PublicAPIs: []string{"xxx"}, // disable all API
		}

		r := setupRouter(hs)
		r.POST("/disabled_api", func(c *gin.Context) {})

		ts := httptest.NewUnstartedServer(r)
		ts.Listener = newPubListener(t)
		ts.Start()
		defer ts.Close()

		t.Logf("test server: %q", ts.URL)
		time.Sleep(time.Second)

		// use non-loopback IP to request
		ip, err := datakit.GetFirstGlobalUnicastIP()
		assert.NoError(t, err)
		tcpAddr := &net.TCPAddr{
			IP: net.ParseIP(ip),
		}

		transport := &http.Transport{
			Dial: (&net.Dialer{LocalAddr: tcpAddr}).Dial,
		}

		cli := &http.Client{
			Transport: transport,
		}

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/disabled_api", ts.URL), nil)
		assert.NoError(t, err)
		resp, err := cli.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		assert.Contains(t, string(respBody), "datakit.publicAccessDisabled")

		t.Logf("resp body: %s", string(respBody))
	})

	t.Run("CORS", func(t *T.T) {
		hs := defaultHTTPServerConf()
		hs.apiConfig.AllowedCORSOrigins = []string{}
		router := setupRouter(defaultHTTPServerConf())

		router.POST("/timeout", func(c *gin.Context) {})

		ts := httptest.NewUnstartedServer(router)
		ts.Listener = newPubListener(t)
		ts.Start()
		defer ts.Close()

		t.Logf("test server: %q", ts.URL)
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
	})
}
