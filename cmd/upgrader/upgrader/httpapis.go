// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package upgrader

import (
	"net/http"

	"github.com/GuanceCloud/cliutils/logger"
	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"

	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
)

var (
	httpServClosed = make(chan error, 4)
	ui             = &upgraderImpl{
		upgradeStatus: atomic.NewInt32(0),
	}
)

type RegisterData struct {
	Datakit *ws.DataKit       `json:"datakit"`
	DCA     *config.DCAConfig `json:"dca"`
}

func apiDKVersion(w http.ResponseWriter, r *http.Request, args ...interface{}) (interface{}, error) {
	if args == nil || len(args) != 1 {
		l.Error("invalid handler")
		return nil, httpapi.ErrInvalidAPIHandler
	}

	var u *upgraderImpl

	for _, arg := range args {
		switch x := arg.(type) {
		case *upgraderImpl:
			u = x
		default:
			return nil, httpapi.ErrInvalidAPIHandler
		}
	}

	x, err := u.fetchCurrentDKVersion()
	if err != nil {
		return nil, uhttp.Errorf(httpapi.ErrUpgradeGetVersionFailed, "%s", err.Error())
	} else {
		return x, nil
	}
}

func apiUpgrade(w http.ResponseWriter, r *http.Request, args ...interface{}) (interface{}, error) {
	if args == nil || len(args) != 1 {
		l.Error("invalid handler")
		return nil, httpapi.ErrInvalidAPIHandler
	}

	var u upgrader

	for _, arg := range args {
		switch x := arg.(type) {
		case upgrader:
			u = x
		default:
			return nil, httpapi.ErrInvalidAPIHandler
		}
	}

	version := r.URL.Query().Get("version")
	force := (r.URL.Query().Get("force") != "")
	if err := u.upgrade(withVersion(version), withForce(force)); err != nil {
		return nil, err
	}

	return uhttp.RawJSONBody(`{"msg": "success"}`), nil
}

func DebugRun() {
	if err := Cfg.LoadMainTOML(MainConfigFile); err != nil {
		l.Warnf("unable to load main config file: %s", err)
	}
	Cfg.SetLogging()
	l = logger.SLogger("main")

	_ = startHTTPServer()
	if err := startDCA(&serviceImpl{done: make(chan struct{})}); err != nil {
		l.Errorf("startDCA failed: %s", err.Error())
	}
}

func startHTTPServer() *http.Server {
	gin.DefaultErrorWriter = getGinErrLogger()
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	router := gin.New()
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: uhttp.GinLogFormatter,
		Output:    getGinLog(),
	}))
	router.Use(gin.Recovery())

	if len(Cfg.IPWhiteList) > 0 {
		router.Use(getIPVerifyMiddleware())
	}

	ui.c = Cfg

	router.POST("/v1/datakit/upgrade",
		httpapi.RawHTTPWrapper(nil, apiUpgrade, ui))

	router.GET("/v1/datakit/version",
		httpapi.RawHTTPWrapper(nil, apiDKVersion, ui))

	serv := &http.Server{
		Addr:    Cfg.Listen,
		Handler: router,
	}

	go func() {
		if err := serv.ListenAndServe(); err != nil {
			l.Infof("ListenAndServe: %s", err.Error())
			httpServClosed <- err
		}
	}()

	return serv
}
