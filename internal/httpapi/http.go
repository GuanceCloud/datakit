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
	"runtime"
	"strings"
	"sync"

	// nolint:gosec
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/metrics"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/pipeline-go/constants"
	"github.com/GuanceCloud/timeout"
	"github.com/didip/tollbooth/v6/limiter"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	l = logger.DefaultSLogger("http")

	pprofServer *http.Server

	g = datakit.G("http")

	semReload          *cliutils.Sem // [http server](the normal one, not dca nor pprof) reload signal
	semReloadCompleted *cliutils.Sem // [http server](the normal one, not dca nor pprof) reload completed signal

	httpConfMtx sync.Mutex
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

	reqLimiter *limiter.Limiter
}

func defaultHTTPServerConf() *httpServerConf {
	return &httpServerConf{
		apiConfig: &config.APIConfig{
			PublicAPIs: []string{"/v1/ping", "/v1/ntp"}, // Default enable APIs.
		},
	}
}

func Start(opts ...option) {
	l = logger.SLogger("http")

	// inject reload http server function to kv
	config.GetKV().SetHTTPServerRestart(ReloadHTTPServer)

	// register golang runtime metrics
	metrics.MustAddGolangMetrics()

	hs := defaultHTTPServerConf()

	for _, opt := range opts {
		if opt != nil {
			opt(hs)
		}
	}

	if hs.apiConfig.RequestRateLimit > 0.0 {
		ttl := hs.apiConfig.RequestRateLimitTTL
		if ttl <= 0 {
			ttl = time.Minute // default 1min
		}

		hs.reqLimiter = setupLimiter(hs.apiConfig.RequestRateLimit, ttl)

		if hs.apiConfig.RequestRateLimitBurst > 0 {
			hs.reqLimiter.SetBurst(hs.apiConfig.RequestRateLimitBurst)
		}

		l.Infof("set up request limit at %f, ttl: %s, burst: %d",
			hs.apiConfig.RequestRateLimit, ttl, hs.apiConfig.RequestRateLimitBurst)
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

	startDCA(hs)

	// start HTTP server
	g.Go(func(ctx context.Context) error {
		HTTPStart(hs)
		l.Info("http goroutine exit")
		return nil
	})

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
	c.Header("X-DataKit", fmt.Sprintf("%s/%s", datakit.Version, datakit.DKHost))
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

	addNewRegistedAPIs(hs)
	// use whitelist config
	if len(hs.apiConfig.PublicAPIs) != 0 {
		router.Use(apiWhiteListMiddleware(hs.apiConfig.PublicAPIs))
	}

	applyRegistedAPIs(router)

	createDCARouter(router, hs)

	wraper1, wraper2 := &HandlerWrapper{WrappedResponse: true}, &HandlerWrapper{WrappedResponse: false}

	reqLimiter := hs.reqLimiter

	// For ntp api, we should keep the same response struct like
	// dataway API /v1/ntp, and there is no outter content wrapper.
	router.GET("/v1/ntp", wraper2.RawHTTPWrapper(reqLimiter, apiNTP))

	router.GET("/v1/ping", wraper1.RawHTTPWrapper(reqLimiter, apiPing))
	router.POST("/v1/write/:category", wraper1.RawHTTPWrapper(reqLimiter, apiWrite, &apiWriteImpl{}))

	router.POST("/v1/query/raw", wraper1.RawHTTPWrapper(reqLimiter, apiQueryRaw, hs.dw))

	router.POST("/v1/object/labels", wraper1.RawHTTPWrapper(reqLimiter, apiCreateOrUpdateObjectLabel, hs.dw))
	router.DELETE("/v1/object/labels", wraper1.RawHTTPWrapper(reqLimiter, apiDeleteObjectLabel, hs.dw))

	router.POST("/v1/pipeline/debug", wraper1.RawHTTPWrapper(reqLimiter, apiPipelineDebugHandler))

	router.POST("/v1/lasterror", wraper1.RawHTTPWrapper(reqLimiter, apiPutLastError, dkio.DefaultFeeder()))
	router.GET("/restart", wraper1.RawHTTPWrapper(reqLimiter, apiRestart, apiRestartImpl{conf: hs}))

	router.GET("/metrics", ginLimiter(reqLimiter), metrics.HTTPGinHandler(promhttp.HandlerOpts{}))

	router.GET("/v1/global/host/tags", ginLimiter(reqLimiter), getHostTags)
	router.POST("/v1/global/host/tags", ginLimiter(reqLimiter), postHostTags)
	router.DELETE("/v1/global/host/tags", ginLimiter(reqLimiter), deleteHostTags)
	router.GET("/v1/global/election/tags", ginLimiter(reqLimiter), getElectionTags)
	router.POST("/v1/global/election/tags", ginLimiter(reqLimiter), postElectionTags)
	router.DELETE("/v1/global/election/tags", ginLimiter(reqLimiter), deleteElectionTags)

	return router
}

func isLoopbackClient(c *gin.Context) bool {
	xff := c.GetHeader("X-Forwarded-For")
	xri := c.GetHeader("X-Real-IP")
	if xff == "" && xri == "" {
		if c.Request.RemoteAddr == "@" {
			return true
		}
		return net.ParseIP(c.ClientIP()).IsLoopback()
	}

	if xff != "" {
		if ip := net.ParseIP(xff); ip != nil {
			if ip.IsLoopback() { // fake loopback
				l.Warnf("forwarded loopback IP(forwarded-ip) not accepted")
				return false
			}
		}
	}

	if xri != "" {
		if ip := net.ParseIP(xri); ip != nil {
			if ip.IsLoopback() { // fake loopback
				l.Warnf("forwarded loopback(x-real-ip) IP not accepted")
				return false
			}
		}
	}

	return false
}

func apiWhiteListMiddleware(apis []string) gin.HandlerFunc {
	publicAPITable := make(map[string]struct{}, len(apis))
	for _, apiPath := range apis {
		l.Infof("apply API %q to white list(maybe duplicated)", apiPath)

		apiPath = strings.TrimSpace(apiPath)
		if len(apiPath) > 0 && apiPath[0] != '/' {
			apiPath = "/" + apiPath
		}
		publicAPITable[apiPath] = struct{}{}
	}

	return func(c *gin.Context) {
		if _, ok := publicAPITable[c.Request.URL.Path]; !ok && !isLoopbackClient(c) {
			uhttp.HttpErr(c, uhttp.Errorf(ErrPublicAccessDisabled,
				"api %s disabled from external IP, only loopback(localhost) allowed",
				c.Request.URL.Path))
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

	if hs.apiConfig.ListenSocket != "" {
		g.Go(func(ctx context.Context) error {
			tryStartUDSServer(hs.apiConfig.ListenSocket,
				srv, true, semReload, semReloadCompleted)
			l.Info("http(uds) server exit")
			return nil
		})
	}

	l.Debug("http server started")

	stopFunc := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
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

func tryStartUDSServer(udsPath string, srv *http.Server,
	canReload bool,
	semReload,
	semReloadCompleted *cliutils.Sem,
) {
	if runtime.GOOS == datakit.OSWindows {
		l.Errorf("Unix domain socket not available on Windows, ignored")
		return
	}

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

	// serve udsPath
	udsListener, err := initUnixListener(udsPath)
	if err != nil {
		l.Errorf("init uds listener failed: %s", err)
		return
	}

	defer func() {
		if udsListener != nil {
			err = udsListener.Close()
			if err != nil {
				l.Warnf("close uds listener failed: %s", err)
			}
		}
	}()

	l.Infof("try start uds server, path %s ...", srv.Addr)
	if err = srv.Serve(udsListener); err != nil {
		l.Errorf("start server failed: %s", err.Error())
	}
	l.Info("http server exit")
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

func initUnixListener(udsPath string) (net.Listener, error) {
	var (
		listener net.Listener
		err      error
	)

	if filepath.IsAbs(udsPath) {
		_ = os.MkdirAll(filepath.Dir(udsPath), 0o755) //nolint:gosec
		if fi, err := os.Stat(udsPath); err == nil {
			if fi.Mode()&os.ModeSocket == 0 {
				return nil, fmt.Errorf("reuse %s faild: file mode %s", udsPath,
					fi.Mode().String())
			}
			if err = os.Remove(udsPath); err != nil && !os.IsNotExist(err) {
				return nil, fmt.Errorf("remove %s: %w", udsPath, err)
			}
		}

		if listener, err = net.Listen("unix", udsPath); err != nil {
			return nil, fmt.Errorf(`net.Listen("unix"): %w`, err)
		}
		if err := os.Chmod(udsPath, 0o722); err != nil { //nolint:gosec
			return nil, fmt.Errorf("setting socket permissions failed: %w", err)
		}

		return listener, nil
	} else {
		return nil, fmt.Errorf("UDS path %s is not absolute", udsPath)
	}
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

// ReloadDataKit will reload datakit modules wihout restart datakit process.
func ReloadDataKit(ctx context.Context) error {
	round := 0 // 循环次数
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("reload timeout")

		default:
			switch round {
			case 0:
				l.Info("before ReloadCheckInputCfg")

				_, err := config.ReloadCheckInputCfg()
				if err != nil {
					l.Errorf("ReloadCheckInputCfg failed: %v", err)
					return err
				}

				l.Info("before ReloadCheckPipelineCfg")

			case 1:
				l.Info("before StopInputs")

				if err := inputs.StopInputs(); err != nil {
					l.Errorf("StopInputs failed: %v", err)
					return err
				}

			case 2:
				l.Info("before ReloadInputConfig")

				if err := config.ReloadInputConfig(); err != nil {
					l.Errorf("ReloadInputConfig failed: %v", err)
					return err
				}

			case 3:
				l.Info("before set pipelines")
				if m, ok := plval.GetManager(); ok && m != nil {
					// git
					if config.GitHasEnabled() {
						m.LoadScriptsFromWorkspace(constants.NSGitRepo,
							filepath.Join(datakit.GitReposRepoFullPath, "pipeline"), nil)
					}
					// local
					plPath := filepath.Join(datakit.InstallDir, "pipeline")
					m.LoadScriptsFromWorkspace(constants.NSDefault, plPath, nil)
				}

			case 4:
				l.Info("before RunInputs")

				CleanHTTPHandler()
				if err := inputs.RunInputs(inputs.AllInputsInfo); err != nil {
					l.Errorf("RunInputs failed: %v", err)
					return err
				}

			case 5:
				l.Info("before ReloadTheNormalServer")

				ReloadHTTPServer()
			}
		}

		round++
		if round > 6 {
			return nil
		}
	}
}

func ReloadHTTPServer() {
	ReloadTheNormalServer(
		WithAPIConfig(config.Cfg.HTTPAPI),
		WithDCAConfig(config.Cfg.DCAConfig),
		WithGinLog(config.Cfg.Logging.GinLog),
		WithGinRotateMB(config.Cfg.Logging.Rotate),
		WithGinReleaseMode(strings.ToLower(config.Cfg.Logging.Level) != "debug"),
		WithDataway(config.Cfg.Dataway),
		WithPProf(config.Cfg.EnablePProf),
		WithPProfListen(config.Cfg.PProfListen),
	)
}
