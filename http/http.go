package http

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	iowrite "io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
)

var (
	l              = logger.DefaultSLogger("http")
	httpBind       string
	ginLog         string
	ginReleaseMode = true
	pprof          bool

	uptime    = time.Now()
	reload    time.Time
	reloadCnt int

	stopCh   = make(chan interface{})
	stopOkCh = make(chan interface{})
	mtx      = sync.Mutex{}
)

const (
	LOGGING_SROUCE     = "source"
	PRECISION          = "precision"
	INPUT              = "input"
	IGNORE_GLOBAL_TAGS = "ignore_global_tags"
	CATEGORY           = "category"

	DEFAULT_PRECISION = "n"
	DEFAULT_INPUT     = "datakit" // 当 API 调用方未亮明自己身份时，默认使用 datakit 作为数据源名称
)

type Option struct {
	Bind   string
	GinLog string

	GinReleaseMode bool
	PProf          bool
}

func Start(o *Option) {

	l = logger.SLogger("http")

	httpBind = o.Bind
	ginLog = o.GinLog
	pprof = o.PProf
	ginReleaseMode = o.GinReleaseMode

	// start HTTP server
	go func() {
		HttpStart()
	}()
}

type welcome struct {
	Version string
	BuildAt string
	Uptime  string
	OS      string
	Arch    string
}

func page404(c *gin.Context) {

	w := &welcome{
		Version: git.Version,
		BuildAt: git.BuildAt,
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}

	c.Writer.Header().Set("Content-Type", "text/html")
	t := template.New(``)
	t, err := t.Parse(config.WelcomeMsgTemplate)
	if err != nil {
		l.Error("parse welcome msg failed: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	buf := &bytes.Buffer{}
	w.Uptime = fmt.Sprintf("%v", time.Since(uptime))
	if err := t.Execute(buf, w); err != nil {
		l.Error("build html failed: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	c.String(http.StatusNotFound, buf.String())
}

func corsMiddleware(c *gin.Context) {
	allowHeaders := []string{
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Authorization",
		"accept",
		"origin",
		"Cache-Control",
		"X-Requested-With",

		// dataflux headers
		"X-Token",
		"X-Datakit-UUID",
		"X-RP",
		"X-Precision",
		"X-Lua",
	}

	c.Writer.Header().Set("Access-Control-Allow-Origin", c.GetHeader("origin"))
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(allowHeaders, ", "))
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}

func HttpStart() {
	gin.DisableConsoleColor()

	l.Infof("set gin log to %s", ginLog)
	f, err := os.Create(ginLog)
	if err != nil {
		l.Fatalf("create gin log failed: %s", err)
	}
	gin.DefaultWriter = iowrite.MultiWriter(f)

	if ginReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	l.Debugf("HTTP bind addr:%s", httpBind)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware)
	router.NoRoute(page404)

	applyHTTPRoute(router)

	// internal datakit stats API
	router.GET("/stats", func(c *gin.Context) { apiGetDatakitStats(c) })
	router.GET("/monitor", func(c *gin.Context) { apiGetDatakitMonitor(c) })

	router.GET("/man", func(c *gin.Context) { apiManualTOC(c) })
	router.GET("/man/:name", func(c *gin.Context) { apiManual(c) })

	router.GET("/reload", func(c *gin.Context) { apiReload(c) })

	router.GET("/v1/ping", func(c *gin.Context) { apiPing(c) })

	router.POST("/v1/write/:category", func(c *gin.Context) { apiWrite(c) })

	router.POST("/v1/query/raw", func(c *gin.Context) { apiQueryRaw(c) })

	srv := &http.Server{
		Addr:    httpBind,
		Handler: router,
	}

	go func() {
		tryStartServer(srv)
		l.Info("http server exit")
	}()

	// start pprof if enabled
	var pprofSrv *http.Server
	if pprof {
		pprofSrv = &http.Server{
			Addr: ":6060",
		}

		go func() {
			tryStartServer(pprofSrv)
			l.Info("pprof server exit")
		}()
	}

	l.Debug("http server started")
	<-stopCh
	l.Debug("stopping http server...")

	if err := srv.Shutdown(context.Background()); err != nil {
		l.Errorf("Failed of http server shutdown, err: %s", err.Error())
	} else {
		l.Info("http server shutdown ok")
	}

	if pprof {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := pprofSrv.Shutdown(ctx); err != nil {
			l.Error(err)
		}
		l.Infof("pprof stopped")
	}

	stopOkCh <- nil
}

func HttpStop() {
	l.Info("trigger HTTP server to stopping...")
	stopCh <- nil
}

func tryStartServer(srv *http.Server) {
	retryCnt := 0

	for {
		l.Infof("try start server at %s(retrying %d)...", srv.Addr, retryCnt)
		if err := srv.ListenAndServe(); err != nil {

			if err != http.ErrServerClosed {
				l.Warnf("start server at %s failed: %s, retrying(%d)...", srv.Addr, err.Error(), retryCnt)
				retryCnt++
			} else {
				l.Debugf("server(%s) stopped on: %s", srv.Addr, err.Error())
				break
			}
		}
		time.Sleep(time.Second)
	}
}
