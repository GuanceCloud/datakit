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
	GetStats       func() (*DatakitStats, error)
	RestartDataKit func() error
	TestPipeline   func(string, string) (string, error)
}

var dcaAPI = &apiList{
	GetStats:       GetStats,
	RestartDataKit: restartDataKit,
	TestPipeline:   pipelineTest,
}

var ignoreAuthURI = []string{
	"/v1/rum/sourcemap",
}

func dcaHTTPStart() {
	gin.DisableConsoleColor()

	if apiServer.ginReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	l.Debugf("DCA HTTP bind addr:%s", apiServer.dcaConfig.Listen)

	router := setupDcaRouter()

	srv := &http.Server{
		Addr:    apiServer.dcaConfig.Listen,
		Handler: router,
	}

	g.Go(func(ctx context.Context) error {
		tryStartServer(srv, false, nil, nil)
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

func whiteListCheck(c *gin.Context) {
	var isValid bool
	context := dcaContext{c: c}
	clientIP := net.ParseIP(c.ClientIP())
	whiteList := apiServer.dcaConfig.WhiteList

	// ignore loopback
	if clientIP.IsLoopback() {
		l.Debugf("loopback ip: %s, ignore check whitelist", clientIP)
		c.Next()
		return
	}

	if len(whiteList) == 0 {
		l.Warn("DCA service is enabled, but the white list is empty!!")
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

func setupDcaRouter() *gin.Engine {
	// set gin logger
	var ginlogger io.Writer

	l.Infof("set DCA server log to %s", apiServer.ginLog)

	if apiServer.ginLog == "stdout" {
		ginlogger = os.Stdout
	} else {
		ginlogger = &lumberjack.Logger{
			Filename:   apiServer.ginLog,
			MaxSize:    apiServer.ginRotate, // MB
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
	router.Use(whiteListCheck)

	// auth check
	router.Use(dcaAuthMiddleware)

	router.NoRoute(dcaDefault)

	router.GET("/v1/dca/stats", dcaStats)
	router.GET("/v1/dca/reload", dcaReload)
	// conf
	router.POST("/v1/dca/saveConfig", dcaSaveConfig)
	router.DELETE("/v1/dca/deleteConfig", dcaDeleteConfig)
	router.GET("/v1/dca/getConfig", dcaGetConfig)
	// pipelines
	router.GET("/v1/dca/pipelines", dcaGetPipelines)
	router.DELETE("/v1/dca/pipelines", dcaDeletePipelines)
	router.GET("/v1/dca/pipelines/detail", dcaGetPipelinesDetail)
	router.POST("/v1/dca/pipelines/test", dcaTestPipelines)
	router.POST("/v1/dca/pipelines", dcaCreatePipeline)
	router.PATCH("/v1/dca/pipelines", dcaUpdatePipeline)

	router.GET("/v1/filter", dcaGetFilter)
	router.GET("/v1/stats/:type", dcaStatsByType)

	router.GET("/v1/log/tail", dcaGetLogTail)
	router.GET("/v1/log/download", dcaDownloadLog)

	return router
}
