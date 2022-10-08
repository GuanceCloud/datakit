// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package main

// 本工具可集成出一个 datakit 的命令行工具，用于模拟出一个 dataway，并且 mock 巨量的数据推送给 datakit，以测试 datakit 的基本数据吞吐能力。

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
)

var (
	flagDataway        = flag.String("dataway", "localhost:12345", `set dataway to "" to disable dataway`)
	flagDatawayLatency = flag.Duration("dataway-duration", time.Millisecond, "dataway response latency")
	flagWorker         = flag.Int("worker", 10, "set to 0 to disable workers")
	flagWrokerSleep    = flag.Duration("worker-sleep", 10*time.Millisecond, "worker sleep during request")
)

func startHTTP() {
	router := gin.New()

	router.Use(gin.Recovery())

	router.POST("/v1/write/:category",
		func(c *gin.Context) {
			time.Sleep(*flagDatawayLatency)

			body, err := ioutil.ReadAll(c.Request.Body)
			if err != nil {
				log.Println(err)
				return
			}

			if c.Request.Header.Get("Content-Encoding") == "gzip" {
				unzipbody, err := uhttp.Unzip(body)
				if err != nil {
					log.Printf("unzip: %s", err)
					return
				}

				log.Printf("unzip body: %d => %d(%.4f)", len(body), len(unzipbody), float64(len(body))/float64(len(unzipbody)))

				body = unzipbody
			}

			pts, err := lp.ParsePoints(body, &lp.Option{EnablePointInKey: true})
			if err != nil {
				log.Printf("ParsePoints: %s, points: %q", err, body)
			} else {
				log.Printf("accept %d points from %s: %s", len(pts), c.Request.URL.Path, pts[0].Name())
			}

			c.Status(http.StatusOK)
		})

	srv := &http.Server{
		Addr:    *flagDataway,
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}

func setulimit() {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Println("Error Getting Rlimit ", err)
	}

	rLimit.Max = 999999
	rLimit.Cur = 999999
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Println("Error Setting Rlimit ", err)
	}
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Println("Error Getting Rlimit ", err)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	setulimit()
	flag.Parse()

	wg := sync.WaitGroup{}

	if *flagDataway != "" {
		wg.Add(1)
		go func() {
			startHTTP()
		}()
	}

	time.Sleep(time.Second)

	reqs := map[string][]byte{
		fmt.Sprintf("http://localhost:9529/v1/write/logstreaming?type=influxdb&size=%d",
			len(loggingData)): []byte(loggingData),

		fmt.Sprintf("http://localhost:9529/v1/write/logging?input=post-v1-write-logging&size=%d",
			len(loggingData)): []byte(loggingData),

		fmt.Sprintf("http://localhost:9529/v1/write/logging?input=post-v1-write-logging-large&size=%d",
			len(loggingData)*100): []byte(strings.Repeat(loggingData, 100)),

		fmt.Sprintf("http://localhost:9529/v1/write/metric?input=post-v1-write-metric&size=%d",
			len(metricData)): []byte(metricData),

		fmt.Sprintf("http://localhost:9529/v1/write/metric?input=post-v1-write-metric-large&size=%d",
			len(metricData)*500): []byte(strings.Repeat(metricData, 500)),

		fmt.Sprintf("http://localhost:9529/v1/write/object?input=post-v1-write&size=%d",
			len(objectData)): []byte(objectData),

		fmt.Sprintf("http://localhost:9529/v1/write/rum?precision=ms&input=post-v1-write-rum&size=%d",
			len(rumData)): []byte(rumData),
	}

	if *flagWorker <= 0 {
		wg.Wait()
	}

	wg.Add(*flagWorker)
	for i := 0; i < *flagWorker; i++ {
		go func() {
			defer wg.Done()
			cli := http.Client{}

			n := 0
			for {
				for k, v := range reqs {
					if len(v) > 4*1024*1024 { // slow send on large body
						time.Sleep(10 * time.Second)
					}

					req, err := http.NewRequest("POST", k, bytes.NewBuffer(v))
					if err != nil {
						log.Printf("http.NewRequest: %s", err)
						time.Sleep(time.Second)
						continue
					}

					resp, err := cli.Do(req)
					if err != nil {
						log.Printf("cli.Do: %s", err)
						time.Sleep(time.Second)
						continue
					}

					if body, err := ioutil.ReadAll(resp.Body); err != nil {
						log.Printf("ioutil.ReadAll: %s", err)
					} else {
						switch resp.StatusCode / 100 {
						case 2:
						default:
							log.Printf("request %s failed:: %s", k, string(body))
						}
					}

					resp.Body.Close() //nolint:errcheck,gosec
					n++

					if n%100 == 0 {
						time.Sleep(time.Second * 10)
					} else {
						time.Sleep(*flagWrokerSleep)
					}
				}
			}
		}()
	}

	wg.Wait()
}

var (
	//go:embed metric.data
	metricData string

	//go:embed logging.data
	loggingData string

	//go:embed object.data
	objectData string

	//go:embed rum.data
	rumData string
)
