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
	"sync/atomic"
	"syscall"
	"time"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/dustin/go-humanize"
	limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/gin"
)

var (
	flagListen  = flag.String("listen", "localhost:54321", "HTTP listen")
	flagGinLog  = flag.Bool("gin-log", false, "enable or disable gin log")
	flagMaxBody = flag.Int("max-body", 0, "set max body size(kb)")
	flagDecode  = flag.Bool("decode", false, "try decode request")

	MPts, LPts atomic.Int64
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

	router.POST("/v1/write/:category",
		func(c *gin.Context) {
			log.Printf("************************************************")

			if len(c.Errors) > 0 {
				log.Printf("context error: %s, skipped", c.Errors[0].Error())
				return
			}

			var (
				start    = time.Now()
				encoding point.Encoding
				dec      *point.Decoder
			)

			if body, err := io.ReadAll(c.Request.Body); err != nil {
				c.Status(http.StatusInternalServerError)
				return
			} else {
				elapsed := time.Since(start)
				if len(body) > 0 {
					log.Printf("************************************************")
					log.Printf("copy elapsed %s, bandwidth %fKB/S", elapsed, float64(len(body))/(float64(elapsed)/float64(time.Second))/1024.0)
				}

				if !*flagDecode {
					goto end
				}

				for k, _ := range c.Request.Header {
					log.Printf("%s: %s", k, c.Request.Header.Get(k))
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

				encoding = point.HTTPContentType(c.Request.Header.Get("Content-Type"))
				switch encoding {
				case point.Protobuf:
					dec = point.GetDecoder(point.WithDecEncoding(point.Protobuf))
					defer point.PutDecoder(dec)

				case point.LineProtocol:
					dec = point.GetDecoder(point.WithDecEncoding(point.LineProtocol))
					defer point.PutDecoder(dec)

				default: // not implemented
					log.Printf("unknown encoding %s", encoding)
				}

				if dec != nil {
					if pts, err := dec.Decode(body); err != nil {
						log.Printf("decode on %s error: %s", encoding, err)
					} else {
						nwarns := 0
						for _, pt := range pts {
							if len(pt.Warns()) > 0 {
								//fmt.Printf(pt.Pretty())
								nwarns++
							}
						}

						cat := point.CatURL(c.Request.URL.Path)

						switch cat {
						case point.Logging:
							LPts.Add(int64(len(pts)))

						case point.Metric:
							MPts.Add(int64(len(pts)))
						}

						log.Printf("decode %d points, %d with warnnings", len(pts), nwarns)
						log.Printf("total M/%s, L/%s", humanize.SI(float64(MPts.Load()), ""), humanize.SI(float64(LPts.Load()), ""))
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

func showENVs() {
	for _, env := range os.Environ() {
		log.Println(env)
	}
}

// nolint: typecheck
func main() {
	showENVs()

	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		panic(fmt.Sprintf("Error Getting Rlimit: %s", err))
	}

	fmt.Println(rLimit)
	rLimit.Max = 10240
	rLimit.Cur = 10240
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		panic(fmt.Sprintf("Error Setting Rlimit: %s", err))
	}
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		panic(fmt.Sprintf("Error Getting Rlimit: %s", err))
	}

	flag.Parse()
	benchHTTPServer()
}
