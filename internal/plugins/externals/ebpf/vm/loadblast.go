//go:build ignore
// +build ignore

package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type result struct {
	Name        string  `json:"name"`
	URL         string  `json:"url"`
	Concurrency int     `json:"concurrency"`
	DurationSec int     `json:"duration_sec"`
	Attempts    uint64  `json:"attempts"`
	Success     uint64  `json:"success"`
	Fail        uint64  `json:"fail"`
	QPS         float64 `json:"qps"`
}

func main() {
	url := flag.String("url", "http://127.0.0.1:18080/", "target url")
	dur := flag.Duration("duration", 10*time.Second, "run duration")
	conc := flag.Int("concurrency", 32, "worker count")
	name := flag.String("name", "worker", "report name")
	timeout := flag.Duration("timeout", 2*time.Second, "request timeout")
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	tr := &http.Transport{
		Proxy:                 nil,
		DialContext:           (&net.Dialer{Timeout: *timeout, KeepAlive: 30 * time.Second}).DialContext,
		MaxIdleConns:          *conc * 4,
		MaxIdleConnsPerHost:   *conc * 4,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   *timeout,
		ExpectContinueTimeout: time.Second,
		DisableCompression:    true,
		ForceAttemptHTTP2:     false,
	}
	client := &http.Client{Transport: tr, Timeout: *timeout}

	ctx, cancel := context.WithTimeout(context.Background(), *dur)
	defer cancel()

	var attempts uint64
	var success uint64
	var fail uint64
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < *conc; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				if ctx.Err() != nil {
					return
				}
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, *url, nil)
				if err != nil {
					atomic.AddUint64(&fail, 1)
					continue
				}
				atomic.AddUint64(&attempts, 1)
				resp, err := client.Do(req)
				if err != nil {
					atomic.AddUint64(&fail, 1)
					continue
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					atomic.AddUint64(&success, 1)
				} else {
					atomic.AddUint64(&fail, 1)
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start).Seconds()
	out := result{
		Name:        *name,
		URL:         *url,
		Concurrency: *conc,
		DurationSec: int(dur.Seconds()),
		Attempts:    atomic.LoadUint64(&attempts),
		Success:     atomic.LoadUint64(&success),
		Fail:        atomic.LoadUint64(&fail),
		QPS:         float64(atomic.LoadUint64(&success)) / elapsed,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(out)
}
