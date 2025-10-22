// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package server is DCA's HTTP server
package server

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
)

var l = logger.DefaultSLogger("server")

type ServerOptions struct {
	HTTPPort        string // HTTPPort is the port of HTTP server
	PromListen      string // Prometheus metric export URL
	ConsoleWebURL   string // ConsoleWebURL is the URL of console web page
	ConsoleAPIURL   string // ConsoleAPIURL is the URL of console API
	StaticBaseURL   string
	ConsoleAPIProxy string
	DBPath          string
	TLSEnable       bool
	TLSCertFile     string
	TLSKeyFile      string
}

// Manager define a ws server manager.
var Manager = ClientManager{
	Register:       make(chan *Client, 100),
	Unregister:     make(chan *Client, 100),
	Clients:        make(map[string]*Client),
	WebsocketConns: make(map[string]chan *websocket.Conn),
}

var (
	dbPath        = DefaultDBPath
	enableTLS     = false
	tlsCertFile   string
	tlsKeyFile    string
	datakitDB     = NewDB()
	g             = goroutine.NewGroup(goroutine.Option{Name: "dca-server"})
	consoleClient = http.Client{
		Timeout: 30 * time.Second,
	}
)

func Start(opt *ServerOptions) error {
	l = logger.SLogger("server")

	if opt != nil {
		if opt.HTTPPort != "" {
			dcaHTTPPort = opt.HTTPPort
		}

		if opt.ConsoleWebURL != "" {
			consoleWebURL = strings.TrimRight(opt.ConsoleWebURL, "/")
		}

		if opt.ConsoleAPIURL != "" {
			consoleAPIURL = strings.TrimRight(opt.ConsoleAPIURL, "/")
		}

		if opt.StaticBaseURL != "" {
			staticBaseURL = strings.TrimRight(opt.StaticBaseURL, "/")
		}

		if opt.ConsoleAPIProxy != "" {
			if p, err := url.Parse(opt.ConsoleAPIProxy); err != nil {
				l.Errorf("invalid proxy URL: %s, ignore proxy", err.Error())
			} else {
				consoleClient.Transport = &http.Transport{
					Proxy: http.ProxyURL(p),
				}
			}
		}

		if opt.DBPath != "" {
			dbPath = opt.DBPath
		}

		// tls setting
		enableTLS = opt.TLSEnable
		tlsCertFile = opt.TLSCertFile
		tlsKeyFile = opt.TLSKeyFile
	}

	router := gin.Default()

	if err := setupRouter(router); err != nil {
		return fmt.Errorf("failed to setup HTTP server: %w", err)
	}

	if err := datakitDB.Init(); err != nil {
		return fmt.Errorf("failed to init DB: %w", err)
	}

	g.Go(func(ctx context.Context) error {
		s := metrics.NewMetricServer()
		if opt.PromListen != "" {
			s.Listen = opt.PromListen
		}

		l.Infof("PromListen on: %q, and the metrics route: %s", s.Listen, s.URL)

		if err := s.Start(); err != nil {
			l.Warnf("start metric server failed: %s", err.Error())
		}
		return nil
	})

	g.Go(func(ctx context.Context) error {
		Manager.Start()
		return nil
	})
	l.Infof("start HTTP server on port %s", dcaHTTPPort)
	addr := fmt.Sprintf(":%s", dcaHTTPPort)

	if enableTLS {
		l.Infof("enable TLS mode")
		l.Debugf("tls cert: %s, tls key: %s", tlsCertFile, tlsKeyFile)
		return router.RunTLS(addr, tlsCertFile, tlsKeyFile)
	}
	return router.Run(addr)
}

func getConsoleAPIURL(path string) string {
	return fmt.Sprintf("%s/api/v1%s", consoleAPIURL, path)
}
