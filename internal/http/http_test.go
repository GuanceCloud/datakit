package http

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/elazarl/goproxy"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestProxy(t *testing.T) {
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello, tls client")
	}))

	defer tlsServer.Close()

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	proxyAddr := "0.0.0.0:12345"

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
	defer proxysrv.Shutdown(ctx)

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

			defer resp.Body.Close()
			t.Logf("resp: %s", string(body))
		}
	}
}

func TestInsecureSkipVerify(t *testing.T) {
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hello, tls client")
	}))

	defer tlsServer.Close()

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

			defer resp.Body.Close()
			t.Logf("resp: %s", string(body))
		}
	}
}

type connWatcher struct {
	nNew      int64
	nClose    int64
	nMax      int64
	nIdle     int64
	nActive   int64
	nHijacked int64
}

func (cw *connWatcher) OnStateChange(conn net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		atomic.AddInt64(&cw.nNew, 1)
		atomic.AddInt64(&cw.nMax, 1)
	case http.StateHijacked:
		atomic.AddInt64(&cw.nHijacked, 1)
		atomic.AddInt64(&cw.nNew, -1)
	case http.StateClosed:
		atomic.AddInt64(&cw.nNew, -1)
		atomic.AddInt64(&cw.nClose, 1)
	case http.StateIdle:
		atomic.AddInt64(&cw.nIdle, 1)
	case http.StateActive:
		atomic.AddInt64(&cw.nActive, 1)
	}
}

func (cw *connWatcher) Count() int {
	return int(atomic.LoadInt64(&cw.nNew))
}

func (cw *connWatcher) String() string {
	return fmt.Sprintf("connections: new: %d, closed: %d, Max: %d, hijacked: %d, idle: %d, active: %d",
		cw.nNew, cw.nClose, cw.nMax, cw.nHijacked, cw.nIdle, cw.nActive)
}

func TestClientConnections(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond * 50)
		fmt.Fprintf(w, "hello\n")
	}))

	var cw connWatcher
	ts.Config.ConnState = cw.OnStateChange

	defer ts.Close()

	wg := sync.WaitGroup{}

	nclients := 10
	nreq := 1000

	go func() {
		for {
			time.Sleep(time.Second)
			fmt.Printf("server: %s, clients: %d, cw: %s\n", ts.URL, nclients, cw.String())
		}
	}()

	wg.Add(nclients)
	for i := 0; i < nclients; i++ {
		go func() {
			defer wg.Done()

			cli := Cli(nil)

			for j := 0; j < nreq; j++ {
				req, err := http.NewRequest("POST", fmt.Sprintf("%s/hello", ts.URL), nil)
				if err != nil {
					t.Error(err)
				}

				resp, err := cli.Do(req)
				if err != nil {
					t.Error(err)
				}

				io.Copy(ioutil.Discard, resp.Body)

				if err := resp.Body.Close(); err != nil {
					t.Error(err)
				}
			}

			cli.CloseIdleConnections()
		}()
	}

	wg.Wait()
}

func hello(w http.ResponseWriter, req *http.Request) {
	time.Sleep(time.Millisecond * 50)
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
				req, err := http.NewRequest("GET", "http://:8090/hello", nil)
				if err != nil {
					t.Error(err)
				}

				resp, err := cli.Do(req)
				if err != nil {
					t.Error(err)
				}

				io.Copy(ioutil.Discard, resp.Body)

				if err := resp.Body.Close(); err != nil {
					t.Error(err)
				}
			}

			cli.CloseIdleConnections()
		}()
	}

	wg.Wait()

	time.Sleep(time.Second * 10)
	if err := server.Shutdown(context.Background()); err != nil {
		t.Log(err)
	}
}
