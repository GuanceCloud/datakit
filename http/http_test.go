// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/influxdata/influxdb1-client/models"
	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
)

func TestParsePoint(t *testing.T) {
	cases := []struct {
		body []byte
		prec string
		npts int
		fail bool
	}{
		{
			body: []byte(`m1,t1=abc f1=123 123`),
			prec: "h",
			npts: 1,
		},

		{
			body: []byte(`m1,t1=abc f1=123 123`),
			prec: "m",
			npts: 1,
		},

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
		points, err := models.ParsePointsWithPrecision(tc.body, time.Now(), tc.prec)
		if tc.fail {
			tu.NotOk(t, err, "")
		} else {
			tu.Equals(t, tc.npts, len(points))
			for _, pt := range points {
				t.Log(pt.String())
			}
		}
	}
}

func TestRestartAPI(t *testing.T) {
	tokens := []string{
		"http://1.2.3.4?token=tkn_abc123",
		"http://4.3.2.1?token=tkn_abc456",
	}

	dwCfg := &dataway.DataWayCfg{URLs: tokens}
	dw = &dataway.DataWayDefault{}
	if err := dw.Init(dwCfg); err != nil {
		t.Error(err)
	}

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
			tu.Equals(t, 200, resp.StatusCode)
		} else {
			tu.Equals(t, ErrInvalidToken.HttpCode, resp.StatusCode)
		}

		t.Logf("resp: %s", string(body))
	}
}

func TestApiGetDatakitLastError(t *testing.T) {
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
		tu.Equals(t, fakeEM.ErrContent, em.ErrContent)
		tu.Equals(t, fakeEM.Input, em.Input)
	}
}

func TestCORS(t *testing.T) {
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
	tu.Assert(t, origin == got, "expect %s, got '%s'", origin, got)
}

func TestTimeout(t *testing.T) {
	apiConfig.timeoutDuration = 100 * time.Millisecond

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
		t.Run(tc.name, func(t *testing.T) {
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

			tu.Equals(t, tc.expectStatusCode, resp.StatusCode)

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

func TestTimeoutOnConcurrentIdleTCPConnection(t *testing.T) {
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

			tu.Assert(t, closed, "expect closed, but not")
		}()
	}

	wg.Wait()
}

//nolint:durationcheck
func TestTimeoutOnIdleTCPConnection(t *testing.T) {
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

	tu.Assert(t, closed, "expect closed, but not")
}
