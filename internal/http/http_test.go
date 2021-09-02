package http

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
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
