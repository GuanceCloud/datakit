// Package http is datakit's HTTP server
package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"

	// nolint:gosec
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
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
	enablePprof    bool

	uptime = time.Now()

	dw        *dataway.DataWayCfg
	extraTags = map[string]string{}
	apiConfig *APIConfig
	dcaConfig *DCAConfig

	ginRotate = 32 // MB

	g = datakit.G("http")

	DcaToken = ""

	semReload          *cliutils.Sem // [http server](the normal one, not dca nor pprof) reload signal
	semReloadCompleted *cliutils.Sem // [http server](the normal one, not dca nor pprof) reload completed signal
)

//nolint:stylecheck
const (
	LOGGING_SROUCE     = "source"
	PRECISION          = "precision"
	INPUT              = "input"
	IGNORE_GLOBAL_TAGS = "ignore_global_tags"
	CATEGORY           = "category"
	VERSION            = "version"
	DEFAULT_PRECISION  = "n"
	DEFAULT_INPUT      = "datakit" // 当 API 调用方未亮明自己身份时，默认使用 datakit 作为数据源名称
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
	enablePprof = o.PProf
	ginReleaseMode = o.GinReleaseMode
	ginRotate = o.GinRotate
	apiConfig = o.APIConfig
	dw = o.DataWay
	dcaConfig = o.DCAConfig

	// start HTTP server
	g.Go(func(ctx context.Context) error {
		HTTPStart()
		l.Info("http goroutine exit")
		return nil
	})

	// DCA server
	if dcaConfig.Enable {
		g.Go(func(ctx context.Context) error {
			dcaHTTPStart()
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

func HTTPStart() {
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
	router.GET("/stats", apiGetDatakitStats)
	router.GET("/monitor", apiGetDatakitMonitor)
	router.GET("/man", apiManualTOC)
	router.GET("/man/:name", apiManual)
	router.GET("/restart", apiRestart)

	router.GET("/v1/workspace", apiWorkspace)
	router.GET("/v1/ping", apiPing)
	router.POST("/v1/lasterror", apiGetDatakitLastError)
	router.POST("/v1/write/:category", apiWrite)
	router.POST("/v1/query/raw", apiQueryRaw)
	router.POST("/v1/object/labels", apiCreateOrUpdateObjectLabel)
	router.DELETE("/v1/object/labels", apiDeleteObjectLabel)

	refreshRebootSem()

	srv := &http.Server{
		Addr:    apiConfig.Listen,
		Handler: router,
	}

	g.Go(func(ctx context.Context) error {
		tryStartServer(srv, true, semReload, semReloadCompleted)
		l.Info("http server exit")
		return nil
	})

	// start pprof if enabled
	var pprofSrv *http.Server
	if enablePprof {
		pprofSrv = &http.Server{
			Addr: ":6060",
		}

		g.Go(func(ctx context.Context) error {
			tryStartServer(pprofSrv, true, semReload, semReloadCompleted)
			l.Info("pprof server exit")
			return nil
		})
	}

	l.Debug("http server started")

	stopFunc := func() {
		if err := srv.Shutdown(context.Background()); err != nil {
			l.Errorf("Failed of http server shutdown, err: %s", err.Error())
		} else {
			l.Info("http server shutdown ok")
		}

		if enablePprof {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := pprofSrv.Shutdown(ctx); err != nil {
				l.Error(err)
			}
			l.Infof("pprof stopped")
		}
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			stopFunc()
			return
		case <-semReload.Wait():
			l.Info("[HttpServer] reload detected")
			stopFunc()
			if semReloadCompleted != nil {
				semReloadCompleted.Close()
			}
			return
		}
	}
}

func refreshRebootSem() {
	semReload = cliutils.NewSem()
	semReloadCompleted = cliutils.NewSem()
}

func ReloadTheNormalServer() {
	if semReload != nil {
		semReload.Close()

		// wait stop completed
		if semReloadCompleted != nil {
			for range semReloadCompleted.Wait() {
				l.Info("[HttpServer] reload stopped")
				go HTTPStart()
				return
			}
		}
	}
}

func tryStartServer(srv *http.Server, canReload bool, semReload, semReloadCompleted *cliutils.Sem) {
	retryCnt := 0

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("tryStartServer exit")
			return
		default:
			if canReload && semReload != nil {
				select {
				case <-semReload.Wait():
					l.Info("tryStartServer reload detected")

					if semReloadCompleted != nil {
						semReloadCompleted.Close()
					}
					return
				default:
				}
			}
		}

		if portInUse(srv.Addr) {
			l.Warnf("start server at %s ,Port is already used", srv.Addr)
		} else {
			break
		}
		time.Sleep(time.Second)
	}

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

func portInUse(addr string) bool {
	timeout := time.Millisecond * 100
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	defer conn.Close() //nolint:errcheck
	return true
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
