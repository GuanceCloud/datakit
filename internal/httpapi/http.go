// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package httpapi is datakit's HTTP server
package httpapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"path/filepath"
	"reflect"
	"strings"

	// nolint:gosec
	_ "net/http/pprof"
	"os"
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
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

var (
	l = logger.DefaultSLogger("http")

	pprofServer *http.Server

	g = datakit.G("http")

	semReload          *cliutils.Sem // [http server](the normal one, not dca nor pprof) reload signal
	semReloadCompleted *cliutils.Sem // [http server](the normal one, not dca nor pprof) reload completed signal
)

type httpServerConf struct {
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

func defaultHTTPServerConf() *httpServerConf {
	return &httpServerConf{
		apiConfig: &config.APIConfig{},
	}
}

func Start(opts ...option) {
	l = logger.SLogger("http")

	// register golang runtime metrics
	metrics.MustAddGolangMetrics()

	hs := defaultHTTPServerConf()

	for _, opt := range opts {
		if opt != nil {
			opt(hs)
		}
	}

	if hs.apiConfig.RequestRateLimit > 0.0 {
		l.Infof("set request limit to %f", hs.apiConfig.RequestRateLimit)
		reqLimiter = setupLimiter(hs.apiConfig.RequestRateLimit)
	} else {
		l.Infof("set request limit not set: %f", hs.apiConfig.RequestRateLimit)
	}

	hs.timeout = 30 * time.Second
	switch hs.apiConfig.Timeout {
	case "":
	default:
		du, err := time.ParseDuration(hs.apiConfig.Timeout)
		if err == nil {
			hs.timeout = du
		}
	}

	// start HTTP server
	g.Go(func(ctx context.Context) error {
		HTTPStart(hs)
		l.Info("http goroutine exit")
		return nil
	})

	// DCA server require dataway
	if hs.dcaConfig != nil {
		if hs.dcaConfig.Enable {
			if hs.dw == nil {
				l.Warn("Ignore to start DCA server because dataway is not set!")
			} else {
				g.Go(func(ctx context.Context) error {
					dcaHTTPStart(hs)
					l.Info("DCA http goroutine exit")
					return nil
				})
			}
		}
	}

	// start pprof if enabled
	if hs.pprof {
		pprofServer = &http.Server{
			Addr: hs.pprofListen,
		}

		l.Infof("start pprof on %s", hs.pprofListen)
		g.Go(func(ctx context.Context) error {
			tryStartServer(hs, pprofServer, true, semReload, semReloadCompleted)
			l.Info("pprof server exit")
			return nil
		})
	}
}

func setupGinLogger(hs *httpServerConf) (gl io.Writer) {
	// set gin logger
	l.Infof("set gin log to %s", hs.ginLog)
	if hs.ginLog == "stdout" {
		gl = os.Stdout
	} else {
		gl = &lumberjack.Logger{
			Filename:   hs.ginLog,
			MaxSize:    hs.ginRotate, // MB
			MaxBackups: 5,
			MaxAge:     30, // day
		}
	}

	return
}

func setDKInfo(c *gin.Context) {
	c.Header("X-DataKit", fmt.Sprintf("%s/%s", datakit.Version, datakit.DatakitHostName))
}

// dkHTTPTimeout Caution: this middleware must be registered as the first one.
func dkHTTPTimeout(du time.Duration) gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(du),

		timeout.WithHandler(func(c *gin.Context) {
			c.Next()
		}),

		timeout.WithResponse(func(c *gin.Context) {
			c.String(http.StatusRequestTimeout, fmt.Sprintf("timeout(%s)", du))
		}),
	)
}

func setupRouter(hs *httpServerConf) *gin.Engine {
	if hs.ginReleaseMode {
		l.Info("set gin in release mode")
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Caution: timeout middleware MUST be registered as the first one, or may crash the process.
	// DON'T CHANGE ITS ORDER!
	router.Use(dkHTTPTimeout(hs.timeout))

	// use whitelist config
	if len(hs.apiConfig.PublicAPIs) != 0 {
		router.Use(apiWhiteListMiddleware(hs.apiConfig.PublicAPIs))
	}

	router.Use(setDKInfo)

	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: uhttp.GinLogFormatter,
		Output:    setupGinLogger(hs),
	}))

	router.Use(gin.Recovery())

	router.Use(uhttp.CORSMiddlewareV2(hs.apiConfig.AllowedCORSOrigins))

	if !hs.apiConfig.Disable404Page {
		router.NoRoute(page404)
	}

	applyHTTPRoute(router)

	router.GET("/v1/ping", RawHTTPWrapper(reqLimiter, apiPing))
	router.POST("/v1/write/:category", RawHTTPWrapper(reqLimiter, apiWrite, &apiWriteImpl{}))

	router.POST("/v1/query/raw", RawHTTPWrapper(reqLimiter, apiQueryRaw, hs.dw))

	router.POST("/v1/object/labels", RawHTTPWrapper(reqLimiter, apiCreateOrUpdateObjectLabel, hs.dw))
	router.DELETE("/v1/object/labels", RawHTTPWrapper(reqLimiter, apiDeleteObjectLabel, hs.dw))

	router.POST("/v1/pipeline/debug", RawHTTPWrapper(reqLimiter, apiPipelineDebugHandler))

	router.GET("/v1/workspace", RawHTTPWrapper(reqLimiter, apiWorkspace, hs.dw))

	router.POST("/v1/lasterror", RawHTTPWrapper(reqLimiter, apiPutLastError, dkio.DefaultFeeder()))
	router.GET("/restart", RawHTTPWrapper(reqLimiter, apiRestart, apiRestartImpl{conf: hs}))

	router.GET("/metrics", ginLimiter(reqLimiter), metrics.HTTPGinHandler(promhttp.HandlerOpts{}))

	router.GET("/v1/global/host/tags", ginLimiter(reqLimiter), getHostTags)
	router.POST("/v1/global/host/tags", ginLimiter(reqLimiter), postHostTags)
	router.DELETE("/v1/global/host/tags", ginLimiter(reqLimiter), deleteHostTags)
	router.GET("/v1/global/election/tags", ginLimiter(reqLimiter), getElectionTags)
	router.POST("/v1/global/election/tags", ginLimiter(reqLimiter), postElectionTags)
	router.DELETE("/v1/global/election/tags", ginLimiter(reqLimiter), deleteElectionTags)

	return router
}

func apiWhiteListMiddleware(apis []string) gin.HandlerFunc {
	publicAPITable := make(map[string]struct{}, len(apis))
	for _, apiPath := range apis {
		apiPath = strings.TrimSpace(apiPath)
		if len(apiPath) > 0 && apiPath[0] != '/' {
			apiPath = "/" + apiPath
		}
		publicAPITable[apiPath] = struct{}{}
	}

	l.Debugf("public API list: %+#v", publicAPITable)

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

func HTTPStart(hs *httpServerConf) {
	refreshRebootSem()
	l.Debugf("HTTP bind addr:%s", hs.apiConfig.Listen)

	srv := &http.Server{
		Addr:    hs.apiConfig.Listen,
		Handler: setupRouter(hs),
	}

	if hs.apiConfig.CloseIdleConnection {
		srv.ReadTimeout = hs.timeout
	}

	g.Go(func(ctx context.Context) error {
		tryStartServer(hs, srv, true, semReload, semReloadCompleted)
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

		if hs.pprof {
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

func ReloadTheNormalServer(opts ...option) {
	if semReload != nil {
		hs := &httpServerConf{
			apiConfig: &config.APIConfig{},
		}

		for _, opt := range opts {
			if opt != nil {
				opt(hs)
			}
		}

		semReload.Close()

		// wait stop completed
		if semReloadCompleted != nil {
			l.Debug("[HttpServer] check wait")

			<-semReloadCompleted.Wait()
			l.Info("[HttpServer] reload stopped")

			// start HTTP server
			g.Go(func(ctx context.Context) error {
				HTTPStart(hs)
				l.Info("http goroutine exit")
				return nil
			})

			return
		}
	}
}

func tryStartServer(hs *httpServerConf,
	srv *http.Server,
	canReload bool,
	semReload,
	semReloadCompleted *cliutils.Sem,
) {
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
			l.Warnf("start server at %s, port is already used", srv.Addr)
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

	tryTLS := hs.apiConfig.HTTPSEnabled()
	for {
		if tryTLS {
			l.Infof("try start server with tls at %s cert: %s privkey: %s",
				srv.Addr,
				hs.apiConfig.TLSConf.Cert,
				hs.apiConfig.TLSConf.PrivKey)

			if err = srv.ServeTLS(listener,
				hs.apiConfig.TLSConf.Cert,
				hs.apiConfig.TLSConf.PrivKey); err != nil {
				l.Warn(err.Error())
			}
		}

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

func checkTokens(dw *dataway.Dataway, req *http.Request) error {
	if dw == nil {
		return ErrInvalidToken
	}

	localTokens := dw.GetTokens()
	if len(localTokens) == 0 {
		return ErrInvalidToken
	}

	tkn := req.URL.Query().Get("token")
	if tkn == "" || tkn != localTokens[0] {
		return ErrInvalidToken
	}

	return nil
}

// IsNil test if x is a nil pointer or nil interface.
func IsNil(x any) bool {
	return x == nil || (reflect.ValueOf(x).Kind() == reflect.Ptr && reflect.ValueOf(x).IsNil())
}
