// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package main to benchmark proxy input.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"syscall"

	"github.com/gin-gonic/gin"
)

var (
	flagListen  = flag.String("listen", "localhost:54321", "HTTP listen")
	flagGinLog  = flag.Bool("gin-log", false, "enable or disable gin log")
	flagTLSCert = flag.String("cert", "ca-cert.pem", "TLS cert file")
	flagTLSKey  = flag.String("key", "ca-key.pem", "TLS key file")
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

	router.POST("/v1/write/:category",
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

	srv := &http.Server{
		Addr:    *flagListen,
		Handler: router,
	}

	if *flagTLSCert != "" && *flagTLSKey != "" {
		if err := srv.ListenAndServeTLS(*flagTLSCert, *flagTLSKey); err != nil {
			panic(err)
		}
	} else {
		if err := srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}
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
	benchHTTPServer()
}
