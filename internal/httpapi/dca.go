// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type apiList struct {
	GetStats      func() (*DatakitStats, error)
	ReloadDataKit func(context.Context) error
	TestPipeline  func(string, string) (string, error)
}

var (
	dcaAPI = &apiList{
		GetStats:      GetStats,
		ReloadDataKit: ReloadDataKit,
		TestPipeline:  pipelineTest,
	}

	ignoreAuthURI = []string{
		"/v1/rum/sourcemap",
	}
)

func dcaHTTPStart(hs *httpServerConf) {
	gin.DisableConsoleColor()

	if hs.ginReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	l.Debugf("DCA HTTP bind addr:%s", hs.dcaConfig.Listen)

	router := setupDcaRouter(hs)

	srv := &http.Server{
		Addr:    hs.dcaConfig.Listen,
		Handler: router,
	}

	g.Go(func(ctx context.Context) error {
		tryStartServer(hs, srv, false, nil, nil)
		l.Info("DCA server exit")
		return nil
	})

	l.Debug("DCA server started")
	<-datakit.Exit.Wait()
	l.Debug("stopping DCA server...")

	if err := srv.Shutdown(context.Background()); err != nil {
		l.Errorf("Failed of DCA server shutdown, err: %s", err.Error())
	} else {
		l.Info("DCA server shutdown ok")
	}
}

func whiteListCheck(whiteList []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var isValid bool
		context := dcaContext{c: c}
		clientIP := net.ParseIP(c.ClientIP())

		// ignore loopback
		if clientIP.IsLoopback() {
			l.Debugf("loopback ip: %s, ignore check whitelist", clientIP)
			c.Next()
			return
		}

		if len(whiteList) == 0 {
			c.Next()
			return
		}

		isValid = false
		for _, v := range whiteList {
			l.Debugf("check cidr %s, client ip: %s", v, clientIP)
			_, ipNet, err := net.ParseCIDR(v)
			if err != nil {
				ip := net.ParseIP(v)
				if ip == nil {
					l.Warnf("parse ip error, %s, ignore", v)
					continue
				}

				if string(ip) == string(clientIP) {
					isValid = true
					break
				}
			} else if ipNet.Contains(clientIP) {
				isValid = true
				break
			}
		}

		if isValid {
			c.Next()
		} else {
			context.fail(dcaError{
				Code:      401,
				ErrorCode: "whiteList.check.error",
				ErrorMsg:  "your cient is not in the white list",
			})
			c.Abort()
		}
	}
}

func setupDcaRouter(hs *httpServerConf) *gin.Engine {
	// set gin logger
	var ginlogger io.Writer

	l.Infof("set DCA server log to %s", hs.ginLog)

	if hs.ginLog == "stdout" {
		ginlogger = os.Stdout
	} else {
		ginlogger = &lumberjack.Logger{
			Filename:   hs.ginLog,
			MaxSize:    hs.ginRotate, // MB
			MaxBackups: 5,
			MaxAge:     30, // day
		}
	}

	router := gin.New()

	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: nil, // not set, use the default
		Output:    ginlogger,
	}))

	router.Use(gin.Recovery())

	// cors
	router.Use(func(c *gin.Context) {
		allowHeaders := []string{
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Authorization",
			"accept",
			"origin",
			"Cache-Control",
			"X-Requested-With",
			"X-Token",
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(allowHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// white list check
	if len(hs.dcaConfig.WhiteList) > 0 {
		router.Use(whiteListCheck(hs.dcaConfig.WhiteList))
	} else {
		l.Warn("DCA service is enabled, but the white list is empty!!")
	}

	// auth check
	router.Use(dcaAuthMiddleware(hs.dw.GetTokens()))

	router.NoRoute(dcaDefault)

	v1 := router.Group("/v1")

	v1.GET("/stats", dcaStats)
	v1.GET("/reload", dcaReload)
	// conf
	v1.POST("/saveConfig", dcaSaveConfig)
	v1.DELETE("/deleteConfig", dcaDeleteConfig)
	v1.GET("/getConfig", dcaGetConfig)
	// pipelines
	v1.GET("/pipelines", dcaGetPipelines)
	v1.DELETE("/pipelines", dcaDeletePipelines)
	v1.GET("/pipelines/detail", dcaGetPipelinesDetail)
	v1.POST("/pipelines/test", dcaTestPipelines)
	v1.POST("/pipelines", dcaCreatePipeline)
	v1.PATCH("/pipelines", dcaUpdatePipeline)

	v1.GET("/filter", dcaGetFilter)

	v1.GET("/log/tail", dcaGetLogTail)
	v1.GET("/log/download", dcaDownloadLog)

	return router
}
