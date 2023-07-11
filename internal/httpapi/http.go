// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package httpapi is datakit's HTTP server
package httpapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/netip"
	"path/filepath"
	"strings"

	// nolint:gosec
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/metrics"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/timeout"
	"github.com/gin-gonic/gin"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	dkm "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

var (
	l = logger.DefaultSLogger("http")

	apiServer = &httpServer{
		apiConfig: &config.APIConfig{},
	}

	pprofServer *http.Server

	g = datakit.G("http")

	semReload          *cliutils.Sem // [http server](the normal one, not dca nor pprof) reload signal
	semReloadCompleted *cliutils.Sem // [http server](the normal one, not dca nor pprof) reload completed signal
)

type httpServer struct {
	ginLog         string
	ginRotate      int
	ginReleaseMode bool

	apiConfig *config.APIConfig
	dw        *dataway.Dataway
	dcaConfig *config.DCAConfig

	timeout time.Duration

	pprof       bool
	pprofListen string
}

func Start(opts ...option) {
	l = logger.SLogger("http")

	for _, opt := range opts {
		if opt != nil {
			opt(apiServer)
		}
	}

	if apiServer.apiConfig.RequestRateLimit > 0.0 {
		l.Infof("set request limit to %f", apiServer.apiConfig.RequestRateLimit)
		reqLimiter = setupLimiter(apiServer.apiConfig.RequestRateLimit)
	} else {
		l.Infof("set request limit not set: %f", apiServer.apiConfig.RequestRateLimit)
	}

	apiServer.timeout = 30 * time.Second
	switch apiServer.apiConfig.Timeout {
	case "":
	default:
		du, err := time.ParseDuration(apiServer.apiConfig.Timeout)
		if err == nil {
			apiServer.timeout = du
		}
	}

	// start HTTP server
	g.Go(func(ctx context.Context) error {
		HTTPStart()
		l.Info("http goroutine exit")
		return nil
	})

	// DCA server require dataway
	if apiServer.dcaConfig != nil {
		if apiServer.dcaConfig.Enable {
			if apiServer.dw == nil {
				l.Warn("Ignore to start DCA server because dataway is not set!")
			} else {
				g.Go(func(ctx context.Context) error {
					dcaHTTPStart()
					l.Info("DCA http goroutine exit")
					return nil
				})
			}
		}
	}

	// start pprof if enabled
	if apiServer.pprof {
		pprofServer = &http.Server{
			Addr: apiServer.pprofListen,
		}

		l.Infof("start pprof on %s", apiServer.pprofListen)
		g.Go(func(ctx context.Context) error {
			tryStartServer(pprofServer, true, semReload, semReloadCompleted)
			l.Info("pprof server exit")
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
	w.Uptime = fmt.Sprintf("%v", time.Since(dkm.Uptime))
	if err := t.Execute(buf, w); err != nil {
		l.Error("build html failed: %s", err.Error())
		uhttp.HttpErr(c, err)
		return
	}

	apiCountVec.WithLabelValues("404-page",
		c.Request.Method,
		http.StatusText(http.StatusNotFound)).Inc()

	apiReqSizeVec.WithLabelValues("404-page",
		c.Request.Method,
		http.StatusText(http.StatusNotFound)).Observe(float64(approximateRequestSize(c.Request)))

	c.String(http.StatusNotFound, buf.String())
}

func setupGinLogger() (gl io.Writer) {
	// set gin logger
	l.Infof("set gin log to %s", apiServer.ginLog)
	if apiServer.ginLog == "stdout" {
		gl = os.Stdout
	} else {
		gl = &lumberjack.Logger{
			Filename:   apiServer.ginLog,
			MaxSize:    apiServer.ginRotate, // MB
			MaxBackups: 5,
			MaxAge:     30, // day
		}
	}

	return
}

func setDKInfo(c *gin.Context) {
	c.Header("X-DataKit", fmt.Sprintf("%s/%s", datakit.Version, datakit.DatakitHostName))
}

func timeoutResponse(c *gin.Context) {
	c.String(http.StatusRequestTimeout, fmt.Sprintf("timeout(%s)", apiServer.timeout))
}

func dkHTTPTimeout() gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(apiServer.timeout),
		timeout.WithHandler(func(c *gin.Context) {
			c.Next()
		}),
		timeout.WithResponse(timeoutResponse),
	)
}

func setupRouter() *gin.Engine {
	if apiServer.ginReleaseMode {
		l.Info("set gin in release mode")
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// use whitelist config
	if len(apiServer.apiConfig.PublicAPIs) != 0 {
		router.Use(getAPIWhiteListMiddleware())
	}

	router.Use(setDKInfo)

	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: uhttp.GinLogFormmatter,
		Output:    setupGinLogger(),
	}))

	router.Use(gin.Recovery())
	router.Use(uhttp.CORSMiddleware)
	router.Use(dkHTTPTimeout())

	if !apiServer.apiConfig.Disable404Page {
		router.NoRoute(page404)
	}

	// if enableRequestLogger {
	//	router.Use(uhttp.RequestLoggerMiddleware)
	//}

	applyHTTPRoute(router)

	router.GET("/stats", rawHTTPWraper(reqLimiter, apiGetDatakitStats)) // Deprecated, use /metrics

	// limit metrics
	router.GET("/metrics", ginLimiter(reqLimiter), metrics.HTTPGinHandler(promhttp.HandlerOpts{}))

	router.GET("/stats/:type", rawHTTPWraper(reqLimiter, apiGetDatakitStatsByType))
	// router.GET("/stats/input", rawHTTPWraper(reqLimiter, apiGetInputStats))

	router.GET("/restart", apiRestart)

	router.GET("/v1/workspace", ginLimiter(reqLimiter), apiWorkspace)
	router.GET("/v1/ping", rawHTTPWraper(reqLimiter, apiPing))
	router.POST("/v1/lasterror", ginLimiter(reqLimiter), apiGetDatakitLastError)

	router.POST("/v1/write/:category", rawHTTPWraper(reqLimiter, apiWrite, &apiWriteImpl{}))

	router.POST("/v1/query/raw", ginLimiter(reqLimiter), apiQueryRaw)
	router.POST("/v1/object/labels", ginLimiter(reqLimiter), apiCreateOrUpdateObjectLabel)
	router.DELETE("/v1/object/labels", ginLimiter(reqLimiter), apiDeleteObjectLabel)

	router.POST("/v1/pipeline/debug", rawHTTPWraper(reqLimiter, apiPipelineDebugHandler))
	router.POST("/v1/dialtesting/debug", rawHTTPWraper(reqLimiter, apiDebugDialtestingHandler))
	return router
}

func getAPIWhiteListMiddleware() gin.HandlerFunc {
	publicAPITable := make(map[string]struct{}, len(apiServer.apiConfig.PublicAPIs))
	for _, apiPath := range apiServer.apiConfig.PublicAPIs {
		apiPath = strings.TrimSpace(apiPath)
		if len(apiPath) > 0 && apiPath[0] != '/' {
			apiPath = "/" + apiPath
		}
		publicAPITable[apiPath] = struct{}{}
	}

	return func(c *gin.Context) {
		cliIP := net.ParseIP(c.ClientIP())
		if _, ok := publicAPITable[c.Request.URL.Path]; !ok && !cliIP.IsLoopback() {
			uhttp.HttpErr(c, uhttp.Errorf(ErrPublicAccessDisabled,
				"api %s disabled from IP %s, only loopback(localhost) allowed",
				c.Request.URL.Path, cliIP.String()))
			c.Abort()
			return
		}
		c.Next()
	}
}

func HTTPStart() {
	refreshRebootSem()
	l.Debugf("HTTP bind addr:%s", apiServer.apiConfig.Listen)
	srv := &http.Server{
		Addr:    apiServer.apiConfig.Listen,
		Handler: setupRouter(),
	}

	if apiServer.apiConfig.CloseIdleConnection {
		srv.ReadTimeout = apiServer.timeout
	}

	g.Go(func(ctx context.Context) error {
		tryStartServer(srv, true, semReload, semReloadCompleted)
		l.Info("http server exit")
		return nil
	})

	l.Debug("http server started")

	stopFunc := func() {
		if err := srv.Shutdown(context.Background()); err != nil {
			l.Errorf("Failed of http server shutdown, err: %s", err.Error())
		} else {
			l.Info("http server shutdown ok")
		}

		if apiServer.pprof {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := pprofServer.Shutdown(ctx); err != nil {
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
				l.Debug("[HttpServer] before reload completed")
				semReloadCompleted.Close()
				l.Debug("[HttpServer] after reload completed")
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
			l.Debug("[HttpServer] check wait")

			<-semReloadCompleted.Wait()
			l.Info("[HttpServer] reload stopped")

			// start HTTP server
			g.Go(func(ctx context.Context) error {
				HTTPStart()
				l.Info("http goroutine exit")
				return nil
			})

			return
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

	listener, err := initListener(srv.Addr)
	if err != nil {
		l.Errorf("initListener failed: %v", err)
		return
	}

	closeListener := func() {
		if listener != nil {
			err = listener.Close()
			if err != nil {
				l.Warnf("listener.Close failed: %v", err)
			}
		}
	}

	defer closeListener()

	for {
		l.Infof("try start server at %s(retrying %d)...", srv.Addr, retryCnt)
		if err = srv.Serve(listener); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				l.Warnf("start server at %s failed: %s, retrying(%d)...", srv.Addr, err.Error(), retryCnt)
				retryCnt++
			} else {
				l.Debugf("server(%s) stopped on: %s", srv.Addr, err.Error())
				closeListener()
				break
			}

			// retry
			time.Sleep(time.Second)
		}
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
	if apiServer.dw == nil {
		return ErrInvalidToken
	}
	localTokens := apiServer.dw.GetTokens()
	if len(localTokens) == 0 {
		return ErrInvalidToken
	}

	tkn := r.URL.Query().Get("token")

	if tkn == "" || tkn != localTokens[0] {
		return ErrInvalidToken
	}

	return nil
}

func initListener(lsn string) (net.Listener, error) {
	var (
		listener net.Listener
		err      error
	)

	if filepath.IsAbs(lsn) {
		if err = os.RemoveAll(lsn); err != nil {
			return nil, fmt.Errorf("os.RemoveAll: %w", err)
		}

		if listener, err = net.Listen("unix", lsn); err != nil {
			return nil, fmt.Errorf(`net.Listen("unix"): %w`, err)
		}
		return listener, nil
	}

	// netip.ParseAddrPort can't parse `localhost', see:
	//  https://pkg.go.dev/net/netip#ParseAddrPort
	if strings.Contains(lsn, "localhost") {
		lsn = strings.ReplaceAll(lsn, "localhost", "127.0.0.1")
	}

	// ipv6 or ipv6
	if addrPort, err := netip.ParseAddrPort(lsn); err != nil {
		return nil, fmt.Errorf("netip.ParseAddrPort: %w", err)
	} else {
		switch {
		case addrPort.Addr().Is6():
			listener, err = net.Listen("tcp6", lsn)
			if err != nil {
				return nil, fmt.Errorf("net.Listen(tcp6): %w", err)
			}
		default: // ipv4 or ipv6:
			listener, err = net.Listen("tcp", lsn)
			if err != nil {
				return nil, fmt.Errorf("net.Listen(tcp): %w", err)
			}
		}
	}

	return listener, nil
}
