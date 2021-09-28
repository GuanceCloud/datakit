package http

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gopkg.in/natefinch/lumberjack.v2"
)

type DCAConfig struct {
	Enable    bool     `toml:"enable" json:"enable"`
	Listen    string   `toml:"listen" json:"listen"`
	WhiteList []string `toml:"white_list" json:"white_list"`
}

func dcaHttpStart() {
	gin.DisableConsoleColor()

	if ginReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	l.Debugf("DCA HTTP bind addr:%s", dcaConfig.Listen)

	router := gin.New()

	// set gin logger
	l.Infof("set DCA server log to %s", ginLog)
	var ginlogger io.Writer
	if ginLog == "stdout" {
		ginlogger = os.Stdout
	} else {
		ginlogger = &lumberjack.Logger{
			Filename:   ginLog,
			MaxSize:    ginRotate, // MB
			MaxBackups: 5,
			MaxAge:     30, // day
		}
	}

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
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH")

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

	router.GET("/v1/dca/stats", func(c *gin.Context) { dcaStats(c) })
	router.GET("/v1/dca/inputDoc", func(c *gin.Context) { dcaInputDoc(c) })
	router.GET("/v1/dca/reload", func(c *gin.Context) { dcaReload(c) })
	// conf
	router.POST("/v1/dca/saveConfig", func(c *gin.Context) { dcaSaveConfig(c) })
	router.GET("/v1/dca/getConfig", func(c *gin.Context) { dcaGetConfig(c) })
	// pipelines
	router.GET("/v1/dca/pipelines", func(c *gin.Context) { dcaGetPipelines(c) })
	router.GET("/v1/dca/pipelines/detail", func(c *gin.Context) { dcaGetPipelinesDetail(c) })
	router.POST("/v1/dca/pipelines/test", func(c *gin.Context) { dcaTestPipelines(c) })
	router.POST("/v1/dca/pipelines", func(c *gin.Context) { dcaCreatePipeline(c) })
	router.PATCH("/v1/dca/pipelines", func(c *gin.Context) { dcaUpdatePipeline(c) })

	srv := &http.Server{
		Addr:    dcaConfig.Listen,
		Handler: router,
	}

	g.Go(func(ctx context.Context) error {
		tryStartServer(srv)
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
	isValid := true
	context := dcaContext{c: c}
	clientRawIp := c.ClientIP()
	whiteList := dcaConfig.WhiteList

	if len(whiteList) > 0 {
		isValid = false
		for _, v := range whiteList {
			l.Debugf("check cidr %s, client ip: %s", v, clientRawIp)
			clientIp := net.ParseIP(clientRawIp)
			_, ipNet, err := net.ParseCIDR(v)
			if err != nil {
				ip := net.ParseIP(v)
				if ip == nil {
					l.Warnf("parse ip error, %s, ignore", v)
					continue
				}
				if string(ip) == string(clientIp) {
					isValid = true
					break
				}
			} else {
				if ipNet.Contains(clientIp) {
					isValid = true
					break
				}
			}
		}
	} else {
		isValid = false
		l.Warn("DCA service is enabled, but the white list is empty!!")
	}

	if isValid {
		c.Next()
	} else {
		context.fail(dcaError{Code: 401, ErrorCode: "whiteList.check.error", ErrorMsg: "your cient is not in the white list"})
		c.Abort()
	}
}
