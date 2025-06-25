// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/point"
	limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/gin"
)

var (
	flagListen   = flag.String("listen", "0.0.0.0:54321", "HTTP listen")
	flagGinLog   = flag.Bool("gin-log", false, "enable or disable gin log")
	flagMaxBody  = flag.Int("max-body", 0, "set max body size(kb)")
	flagDecode   = flag.Bool("decode", false, "try decode request")
	flagHeader   = flag.Bool("header", false, "show HTTP request headers")
	flag5XXRatio = flag.Int("5xx-ratio", 0, "fail request ratio(minimal is 1/1000)")
	flagLatency  = flag.Duration("latency", time.Millisecond*10, "latency used on API cost")

	MPts, LPts, TPts, totalReq, req5xx, decErr, decErrPts atomic.Int64
)

func benchHTTPServer() {
	router := gin.New()
	router.Use(gin.Recovery())

	if *flagGinLog {
		router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
			Formatter: nil,
			Output:    os.Stdout,
		}))
	}

	if *flagMaxBody > 0 {
		router.Use(limits.RequestSizeLimiter(int64(*flagMaxBody * 1024)))
	}

	router.GET("/v1/datakit/pull", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", []byte(`{}`))
	})

	go func() {
		showTicker := time.NewTicker(time.Second * 3)
		defer showTicker.Stop()
		for {
			select {
			case <-showTicker.C:
				showInfo()
			}
		}
	}()

	router.GET("/2020-01-01/extension/event/next", func(c *gin.Context) { //mock aws lambda telemetry API
		//c.Data(http.StatusOK, "", []byte(`{ "eventType": "SHUTDOWN" }`)) // mock shut down event
		c.Data(http.StatusOK, "", []byte(`{}`))
	})

	router.PUT("/2022-07-01/telemetry", func(c *gin.Context) { //mock aws lambda telemetry API
		c.Data(http.StatusOK, "", nil)
	})

	router.POST("/2020-01-01/extension/register", // mock aws lambda register API
		func(c *gin.Context) {
			c.Header("Lambda-Extension-Identifier", "dk-lambda-ext")
			c.Data(http.StatusOK, "application/json", []byte(`{
"functionName": "dk-lambda-ext-testing",
"functionVersion": "0.1.0",
"accountId": "tester-007",
"handler": "mock"
			}`))
		})

	router.POST("/v1/write/:category",
		func(c *gin.Context) {
			if *flagLatency > 0 {
				time.Sleep(*flagLatency)
			}

			totalReq.Add(1)

			if *flag5XXRatio > 0 {
				ns := time.Now().UnixMicro()
				if r := ns % 1000; r < int64(*flag5XXRatio) {
					req5xx.Add(1)
					showInfo()
					c.Data(http.StatusInternalServerError, "", []byte(fmt.Sprintf("drop ration within %d(%d: %d)", *flag5XXRatio, ns, r)))
					return
				}
			}

			if len(c.Errors) > 0 {
				log.Printf("context error: %s, skipped", c.Errors[0].Error())
				return
			}

			var (
				//start     = time.Now()
				encoding point.Encoding
				dec      *point.Decoder
			)

			if body, err := io.ReadAll(c.Request.Body); err != nil {
				c.Status(http.StatusInternalServerError)
				return
			} else {
				//elapsed := time.Since(start)
				//if len(body) > 0 {
				//	log.Printf("copy elapsed %s, bandwidth %fKB/S", elapsed, float64(len(body))/(float64(elapsed)/float64(time.Second))/1024.0)
				//}

				var headerPts int64
				if x := c.Request.Header.Get("X-Points"); x != "" {
					if n, err := strconv.ParseInt(x, 10, 64); err == nil {
						headerPts = n
					}
				}

				if *flagHeader {
					showHeaders(c)
				}

				if !*flagDecode {
					goto end
				}

				if c.Request.Header.Get("Content-Encoding") == "gzip" {
					unzipbody, err := uhttp.Unzip(body)
					if err != nil {
						//log.Printf("unzip: %s, body: %q", err, body)
						log.Printf("[ERROR] unzip(header %q): %s", body[:10], err)
						c.Data(http.StatusBadRequest, "", []byte(err.Error()))
						return
					}

					//log.Printf("[INFO] unzip body: %d => %d(%.4f)", len(body), len(unzipbody), float64(len(body))/float64(len(unzipbody)))

					body = unzipbody
				}

				encoding = point.HTTPContentType(c.Request.Header.Get("Content-Type"))
				switch encoding {
				case point.Protobuf:
					dec = point.GetDecoder(point.WithDecEncoding(point.Protobuf))
					defer point.PutDecoder(dec)

				case point.LineProtocol:
					dec = point.GetDecoder(point.WithDecEncoding(point.LineProtocol))
					defer point.PutDecoder(dec)

				default: // not implemented
					log.Printf("[ERROR] unknown encoding %s", encoding)
					return
				}

				if dec != nil {
					if pts, err := dec.Decode(body); err != nil {
						log.Printf("[ERROR] decode on %s error: %s", encoding, err)
						decErr.Add(1)
						decErrPts.Add(headerPts)
						showHeaders(c)
						log.Printf("body: %s", string(body[32]))
					} else {
						nwarns := 0
						for _, pt := range pts {
							if len(pt.Warns()) > 0 {
								log.Printf("[WARN] point warn: %s", pt.Warns()[0].String())
								nwarns++
							}

							//log.Println(pt.LineProto())
						}

						cat := point.CatURL(c.Request.URL.Path)

						switch cat {
						case point.Logging:
							LPts.Add(int64(len(pts)))
						case point.Metric:
							MPts.Add(int64(len(pts)))
						case point.Tracing:
							TPts.Add(int64(len(pts)))
						}

						if nwarns > 0 {
							log.Printf("[WARN] decode %d points, %d with warnnings", len(pts), nwarns)
						}
					}
				}

			end:
				c.Status(http.StatusOK)
			}
		})

	srv := &http.Server{
		Addr:    *flagListen,
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}

func showInfo() {
	log.Printf("total M/%d, L/%d, T/%d req/%d, 5xx/%d, 5xx ratio: %d/1000, decErr: %d, decErrPts: %d",
		MPts.Load(),
		LPts.Load(),
		TPts.Load(),
		totalReq.Load(),
		req5xx.Load(),
		*flag5XXRatio,
		decErr.Load(),
		decErrPts.Load(),
	)
}

func showENVs() {
	for _, env := range os.Environ() {
		log.Println(env)
	}
}

func showHeaders(c *gin.Context) {
	var headerArr []string
	for k, _ := range c.Request.Header {
		headerArr = append(headerArr, fmt.Sprintf("%s: %s", k, c.Request.Header.Get(k)))
	}

	log.Println("-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=")
	log.Printf("URL: %s", c.Request.URL)
	log.Printf("headers:\n%s", strings.Join(headerArr, "\n"))
}

// nolint: typecheck
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	showENVs()

	//var rLimit syscall.Rlimit
	//err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	//if err != nil {
	//	panic(fmt.Sprintf("Error Getting Rlimit: %s", err))
	//}

	//fmt.Println(rLimit)
	//rLimit.Max = 10240
	//rLimit.Cur = 10240
	//err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	//if err != nil {
	//	panic(fmt.Sprintf("Error Setting Rlimit: %s", err))
	//}
	//err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	//if err != nil {
	//	panic(fmt.Sprintf("Error Getting Rlimit: %s", err))
	//}

	flag.Parse()
	benchHTTPServer()
}
