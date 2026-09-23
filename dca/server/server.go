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
	HTTPPort                 string // HTTPPort is the port of HTTP server
	PromListen               string // Prometheus metric export URL
	ConsoleWebURL            string // ConsoleWebURL is the URL of console web page
	ConsoleAPIURL            string // ConsoleAPIURL is the URL of console API
	StaticBaseURL            string
	ConsoleAPIProxy          string
	UploadHostStatus         bool
	UploadHostStatusInterval time.Duration
	DBPath                   string
	TLSEnable                bool
	TLSCertFile              string
	TLSKeyFile               string
}

// Manager define a ws server manager.
var Manager = ClientManager{
	// A DCA restart makes every datakit reconnect at once: keep enough room in
	// the queue so a fleet does not get 429 (old clients treat it as a failure
	// and back off, which made the list look like it recovered very slowly).
	Register:       make(chan *Client, clientQueueSize),
	Unregister:     make(chan *Client, clientQueueSize),
	Clients:        make(map[string]*Client),
	WebsocketConns: make(map[string]chan *websocket.Conn),
}

const clientQueueSize = 2048

var (
	dbPath                          = DefaultDBPath
	enableTLS                       = false
	tlsCertFile                     string
	tlsKeyFile                      string
	datakitDB                       = NewDB()
	defaultUploadHostStatusInterval = 30 * time.Second
	datakitCleanupInterval          = time.Hour
	g                               = goroutine.NewGroup(goroutine.Option{Name: "dca-server"})
	consoleClient                   = http.Client{
		Timeout: 30 * time.Second,
	}
)

func Start(opt *ServerOptions) error {
	l = logger.SLogger("server")

	closeCh := make(chan struct{})
	defer close(closeCh)

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

	if opt.UploadHostStatus {
		interval := opt.UploadHostStatusInterval
		if interval <= 0 {
			interval = defaultUploadHostStatusInterval
		}
		g.Go(func(ctx context.Context) error {
			UploadHostStatus(interval, closeCh)
			return nil
		})
	}

	// Clean up the datakits that stopped reporting, a long running DCA process
	// would otherwise keep them forever (it is done once at startup as well).
	g.Go(func(ctx context.Context) error {
		ticker := time.NewTicker(datakitCleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-closeCh:
				return nil
			case <-ticker.C:
				if err := datakitDB.DeleteExpired(); err != nil {
					l.Warnf("failed to clean expired datakits: %s", err.Error())
				}
			}
		}
	})

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
