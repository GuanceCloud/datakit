package http

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"
)

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

	//cli := http.Client{}

	n := 1000
	wg := sync.WaitGroup{}
	wg.Add(n)

	for i := 0; i < n; i++ {

		go func() {

			defer wg.Done()

			cli := HTTPCli(&Options{
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
				//resp, err := SendRequest(req)
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
