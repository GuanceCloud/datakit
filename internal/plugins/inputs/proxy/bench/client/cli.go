// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package main to benchmark proxy input.
package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	flagHost       = flag.String("host", "https://localhost:54321/v1/write/xxx", "HTTP(s) server API")
	flagProxy      = flag.String("proxy", "", "http://IP:Port")
	flagConnection = flag.Int("c", 1, "concurrent connections")
	flagReq        = flag.Int("r", 1, "request of each connection")
	flagPostFile   = flag.String("f", "", "file to post")

	elapsed = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "api_elapsed_seconds",
			Help: "Proxied API elapsed seconds",
		},
	)

	proxyPostBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_post_bytes_total",
			Help: "Proxied API post bytes total",
		},
		[]string{
			"api",
			"status",
		},
	)

	proxyReqLatencyVec = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "api_latency_seconds",
			Help: "Proxied API latency",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		[]string{
			"api",
			"status",
		},
	)
)

func runClients() {
	var wg sync.WaitGroup
	wg.Add(*flagConnection)

	start := time.Now()
	reg := prometheus.NewRegistry()
	reg.MustRegister(proxyReqLatencyVec, proxyPostBytes, elapsed)

	u, err := url.Parse(*flagHost)
	if err != nil {
		panic(err.Error())
	}

	for i := 0; i < *flagConnection; i++ {
		go func() {
			defer wg.Done()

			cli := http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
					Proxy: func(req *http.Request) (*url.URL, error) {
						return url.Parse(*flagProxy)
					},
				},
			}

			var (
				data []byte
				err  error
			)

			if *flagPostFile != "" {
				data, err = os.ReadFile(*flagPostFile)
				if err != nil {
					panic(err.Error())
				}
			}

			for j := 0; j < *flagReq; j++ {
				jstart := time.Now()
				req, err := http.NewRequest(http.MethodPost, *flagHost, bytes.NewBuffer(data))
				if err != nil {
					panic(err.Error())
				}

				status := "not-set"
				resp, err := cli.Do(req)
				if err != nil {
					status = err.Error()
				}

				if resp != nil {
					status = resp.Status
					resp.Body.Close() //nolint:errcheck,gosec
				}

				proxyReqLatencyVec.WithLabelValues(
					u.Path,
					status,
				).Observe(float64(time.Since(jstart)) / float64(time.Second))

				proxyPostBytes.WithLabelValues(
					u.Path,
					status,
				).Add(float64(len(data)))
			}
		}()
	}

	wg.Wait()

	elapsed.Set(float64(time.Since(start)) / float64(time.Second))

	mfs, err := reg.Gather()
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Benchmark metrics:\n%s", metrics.MetricFamily2Text(mfs))
}

func main() {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		panic(fmt.Sprintf("Error Getting Rlimit: %s", err))
	}

	rLimit.Max = 10240
	rLimit.Cur = 10240
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(fmt.Sprintf("Error Setting Rlimit: %s", err))
	}

	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(fmt.Sprintf("Error Getting Rlimit: %s", err))
	}

	flag.Parse()

	runClients()
}
