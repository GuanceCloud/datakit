// datakit HTTP server

package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	l              = logger.DefaultSLogger("http")
	ginLog         string
	ginReleaseMode = true
	pprof          bool

	uptime = time.Now()

	mtx = sync.Mutex{}

	dw        *dataway.DataWayCfg
	extraTags = map[string]string{}
	apiConfig *APIConfig
	dcaConfig *DCAConfig

	ginRotate = 32 // MB

	g = datakit.G("http")

	DcaToken = ""
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
	GinLog    string
	GinRotate int
	APIConfig *APIConfig
	DataWay   *dataway.DataWayCfg
	DCAConfig *DCAConfig

	GinReleaseMode bool
	PProf          bool
}

type APIConfig struct {
	RUMOriginIPHeader string   `toml:"rum_origin_ip_header"`
	Listen            string   `toml:"listen"`
	Disable404Page    bool     `toml:"disable_404page"`
	RUMAppIDWhiteList []string `toml:"rum_app_id_white_list"`
}

func Start(o *Option) {
	l = logger.SLogger("http")

	ginLog = o.GinLog
	pprof = o.PProf
	ginReleaseMode = o.GinReleaseMode
	ginRotate = o.GinRotate
	apiConfig = o.APIConfig
	dw = o.DataWay
	dcaConfig = o.DCAConfig

	// start HTTP server
	g.Go(func(ctx context.Context) error {
		HttpStart()
		l.Info("http goroutine exit")
		return nil
	})

	// DCA server
	if dcaConfig.Enable {
		g.Go(func(ctx context.Context) error {
			dcaHttpStart()
			l.Info("DCA http goroutine exit")
			return nil
		})
	}
}

type welcome struct {
	Version string
	BuildAt string
	Uptime  string
	OS      string
	Arch    string
}

func SetAPIConfig(c *APIConfig) {
	apiConfig = c
}

func SetGlobalTags(tags map[string]string) {
	extraTags = tags
}

func page404(c *gin.Context) {
	w := &welcome{
		Version: datakit.Version,
		BuildAt: git.BuildAt,
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}

	c.Writer.Header().Set("Content-Type", "text/html")
	t := template.New(``)
	t, err := t.Parse(welcomeMsgTemplate)
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

func HttpStart() {
	gin.DisableConsoleColor()

	if ginReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	l.Debugf("HTTP bind addr:%s", apiConfig.Listen)

	router := gin.New()

	// set gin logger
	l.Infof("set gin log to %s", ginLog)
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
	router.Use(uhttp.CORSMiddleware)
	if !apiConfig.Disable404Page {
		router.NoRoute(page404)
	}

	applyHTTPRoute(router)

	// internal datakit stats API
	router.GET("/stats", func(c *gin.Context) { apiGetDatakitStats(c) })
	router.GET("/monitor", func(c *gin.Context) { apiGetDatakitMonitor(c) })
	router.GET("/man", func(c *gin.Context) { apiManualTOC(c) })
	router.GET("/man/:name", func(c *gin.Context) { apiManual(c) })
	router.GET("/restart", func(c *gin.Context) { apiRestart(c) })

	router.GET("/v1/ping", func(c *gin.Context) { apiPing(c) })
	router.POST("/v1/write/:category", func(c *gin.Context) { apiWrite(c) })
	router.POST("/v1/query/raw", func(c *gin.Context) { apiQueryRaw(c) })
	router.POST("/v1/object/labels", func(c *gin.Context) { apiCreateOrUpdateObjectLabel(c) })
	router.DELETE("/v1/object/labels", func(c *gin.Context) { apiDeleteObjectLabel(c) })

	srv := &http.Server{
		Addr:    apiConfig.Listen,
		Handler: router,
	}

	g.Go(func(ctx context.Context) error {
		tryStartServer(srv)
		l.Info("http server exit")
		return nil
	})

	// start pprof if enabled
	var pprofSrv *http.Server
	if pprof {
		pprofSrv = &http.Server{
			Addr: ":6060",
		}

		g.Go(func(ctx context.Context) error {
			tryStartServer(pprofSrv)
			l.Info("pprof server exit")
			return nil
		})
	}

	l.Debug("http server started")
	<-datakit.Exit.Wait()

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
}

func tryStartServer(srv *http.Server) {
	retryCnt := 0

	// TODO: test if port available

	for {
		l.Infof("try start server at %s(retrying %d)...", srv.Addr, retryCnt)
		if err := srv.ListenAndServe(); err != nil {
			if !errors.As(err, &http.ErrServerClosed) {
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

func checkToken(r *http.Request) error {
	localTokens := dw.GetToken()
	if len(localTokens) == 0 {
		return ErrInvalidToken
	}

	tkn := r.URL.Query().Get("token")

	if tkn == "" || tkn != localTokens[0] {
		return ErrInvalidToken
	}

	return nil
}
