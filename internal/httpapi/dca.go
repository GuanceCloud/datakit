// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package httpapi is datakit's HTTP server
package httpapi

import (
	// nolint:gosec
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/gin-gonic/gin"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/usagetrace"
)

var (
	workspaceUUID = ""
	datakitToken  = ""
)

type WorkspaceQueryResponse struct {
	Content []struct {
		Token struct {
			WorkspaceUUID string `json:"ws_uuid"`
		}
	}
}

func setWorkspace(hs *httpServerConf) error {
	if arr := hs.dw.GetTokens(); len(arr) > 0 {
		datakitToken = arr[0]
		if resp, err := hs.dw.WorkspaceQuery(
			[]byte(fmt.Sprintf(`{"token":["%s"]}`, datakitToken))); err != nil {
			return fmt.Errorf("workspace query failed: %w", err)
		} else {
			defer resp.Body.Close() //nolint:errcheck
			respBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("read response body %w", err)
			}

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("workspace query failed: %s", string(respBytes))
			}

			var wqr WorkspaceQueryResponse

			if err := json.Unmarshal(respBytes, &wqr); err != nil {
				return fmt.Errorf("unmarshal response body %w", err)
			}

			if len(wqr.Content) > 0 {
				workspaceUUID = wqr.Content[0].Token.WorkspaceUUID
			}
		}
	}
	return nil
}

func createDCARouter(router *gin.Engine, hs *httpServerConf) {
	if router == nil {
		l.Warnf("createDCARouter failed: router is nil")
		return
	}

	dcaRouter := router.Group("/v1/dca")
	dcaRouter.Use(func(ctx *gin.Context) {
		clientIP := ctx.ClientIP()
		if v, err := IsInternalHost(clientIP, nil); err != nil || !v {
			uhttp.HttpErr(ctx, uhttp.Errorf(ErrPublicAccessDisabled,
				"api %s disabled from IP %s, only internal network allowed",
				ctx.Request.URL.Path, clientIP))
			ctx.Abort()
			return
		}
		ctx.Next()
	})

	dcaRouter.GET("/reload", dcaReload)
	dcaRouter.GET("/info", createGetDCAInfoHandler(hs))
	dcaRouter.GET("/stat", getDCAStat)
	dcaRouter.POST("/saveConfig", dcaSaveConfig)
	dcaRouter.DELETE("/config", dcaDeleteConfig)
	dcaRouter.GET("/config", dcaGetConfig)
}

func createGetDCAInfoHandler(hs *httpServerConf) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if workspaceUUID == "" {
			if err := setWorkspace(hs); err != nil {
				l.Warnf("set workspace failed: %s", err.Error())
			}
		}
		getDCAInfo(ctx)
	}
}

type DCAInfo struct {
	DataKit *ws.DataKit   `json:"datakit"`
	Config  config.Config `json:"config"`
}

func getDCAInfo(c *gin.Context) {
	dk := getDatakitData()

	dcaCtx := getContext(c)

	dcaCtx.success(DCAInfo{
		DataKit: dk,
		Config:  *config.Cfg,
	})
}

func getDCAStat(c *gin.Context) {
	dcaCtx := getContext(c)

	if stats, err := getDatakitStats(); err != nil {
		l.Warnf("get datakit stats failed: %s", err.Error())
		dcaCtx.fail()
	} else {
		dcaCtx.success(stats)
	}
}

func dcaSaveConfig(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	context := &dcaContext{c: c}
	if err != nil {
		l.Errorf("Read request body error: %s", err.Error())
		context.fail()
		return
	}

	defer c.Request.Body.Close() //nolint:errcheck

	param := saveConfigParam{}

	if err := json.Unmarshal(body, &param); err != nil {
		l.Errorf("Json unmarshal error: %s", err.Error())
		context.fail(&ws.ResponseError{Code: 400, ErrorCode: "param.json.invalid", ErrorMsg: "body invalid json format"})
		return
	}

	if err := doSaveConfig(&param); err != nil {
		context.fail(err)
	} else {
		context.success(map[string]string{"path": param.Path})
	}
}

func dcaGetConfig(c *gin.Context) {
	context := getContext(c)
	if content, err := doGetConfig(c.Query("path")); err != nil {
		context.fail(err)
	} else {
		context.success(content)
	}
}

func dcaDeleteConfig(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close() //nolint:errcheck

	context := &dcaContext{c: c}
	if err != nil {
		l.Errorf("Read request body error: %s", err.Error())
		context.fail(&ws.ResponseError{Code: 400, ErrorMsg: "read request body error"})
		return
	}

	param := &struct {
		Path      string `json:"path"`
		InputName string `json:"inputName"`
	}{}
	if err := json.Unmarshal(body, &param); err != nil {
		l.Errorf("Json unmarshal error: %s", err.Error())
		context.fail(&ws.ResponseError{Code: 400, ErrorCode: "param.json.invalid", ErrorMsg: "body invalid json format"})
		return
	}

	if err := doDeleteConfig(param.InputName, param.Path); err != nil {
		context.fail(err)
	} else {
		context.success()
	}
}

func getDatakitStats() (*DCAstats, error) {
	if stats, err := GetStats(); err != nil {
		return nil, fmt.Errorf("get stats failed: %w", err)
	} else {
		return &DCAstats{
			DatakitStats: stats,
			ConfigInfo:   inputs.ConfigInfo,
		}, nil
	}
}

func getDatakitData() *ws.DataKit {
	ip, err := datakit.LocalIP()
	if err != nil {
		l.Warnf("get local ip failed: %s", err.Error())
	}

	datakitInstance := &ws.DataKit{
		Arch:           runtime.GOARCH,
		OS:             runtime.GOOS,
		HostName:       datakit.DKHost,
		Version:        datakit.Version,
		RunTimeID:      datakit.RuntimeID,
		IP:             ip,
		StartTime:      metrics.Uptime.UnixMilli(),
		RunInContainer: datakit.Docker,
		Status:         ws.StatusRunning,
		WorkspaceUUID:  workspaceUUID,
		DataKitRuntimeInfo: ws.DataKitRuntimeInfo{
			GlobalHostTags: datakit.GlobalHostTags(),
			DataDir:        datakit.DataDir,
			ConfdDir:       datakit.ConfdDir,
			PipelineDir:    datakit.PipelineDir,
			InstallDir:     datakit.InstallDir,
			Log:            config.Cfg.Logging.Log,
			GinLog:         config.Cfg.Logging.GinLog,
		},
	}

	ut := usagetrace.GetUsageTraceInstance()
	if ut != nil {
		datakitInstance.RunMode = ut.RunMode
		datakitInstance.UsageCores = ut.UsageCores
	}

	if config.Cfg.DCAConfig != nil && config.Cfg.DCAConfig.WebsocketServer != "" {
		datakitInstance.ConnID = datakitInstance.GetConnID(config.Cfg.DCAConfig.WebsocketServer)
	}

	return datakitInstance
}

var dcaErrorMessage = map[string]string{
	"server.error": "server error",
}

func dcaGetMessage(errCode string) string {
	if errMsg, ok := dcaErrorMessage[errCode]; ok {
		return errMsg
	} else {
		return "server error"
	}
}

type dcaContext struct {
	c    *gin.Context
	data interface{}
}

func (d *dcaContext) send(response *ws.DCAResponse) {
	body, err := json.Marshal(response)
	if err != nil {
		d.fail()
		return
	}

	status := d.c.Writer.Status()

	d.c.Data(status, "application/json", body)
}

func (d *dcaContext) success(datas ...interface{}) {
	var data interface{}

	if len(datas) > 0 {
		data = datas[0]
	}

	if data == nil {
		data = d.data
	}

	response := &ws.DCAResponse{
		Code:    200,
		Content: data,
		Success: true,
	}

	d.send(response)
}

func (d *dcaContext) fail(errors ...*ws.ResponseError) {
	var e *ws.ResponseError
	if len(errors) > 0 {
		e = errors[0]
	} else {
		e = &ws.ResponseError{
			Code:      http.StatusInternalServerError,
			ErrorCode: "server.error",
			ErrorMsg:  "",
		}
	}

	code := e.Code
	errorCode := e.ErrorCode
	errorMsg := e.ErrorMsg

	if code == 0 {
		code = http.StatusInternalServerError
	}

	if errorCode == "" {
		errorCode = "server.error"
	}

	if errorMsg == "" {
		errorMsg = dcaGetMessage(errorCode)
	}

	response := &ws.DCAResponse{
		Code:      code,
		ErrorCode: errorCode,
		Message:   errorMsg,
		Success:   false,
	}

	d.send(response)
}

var errDCAReloadError = &ws.ResponseError{
	ErrorCode: "system.reload.error",
	ErrorMsg:  "reload datakit error",
}

// dca reload.
func dcaReload(c *gin.Context) {
	dcaCtx := getContext(c)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	if err := ReloadDataKit(ctx); err != nil {
		l.Error("reloadDataKit: %s", err)
		dcaCtx.fail(errDCAReloadError)
		return
	}

	dcaCtx.success()
}

type DCAstats struct {
	*DatakitStats
	ConfigInfo inputs.ConfigInfoItem `json:"config_info"`
}

func getContext(c *gin.Context) dcaContext {
	return dcaContext{c: c}
}

// start dca service.
// only support docker mode and dk_upgrader service will start dca service in host mode.
func startDCA(hs *httpServerConf) {
	if hs != nil &&
		hs.dcaConfig != nil &&
		hs.dcaConfig.Enable &&
		hs.dcaConfig.WebsocketServer != "" &&
		datakit.Docker {
		tick := time.NewTicker(10 * time.Second)
		defer tick.Stop()

		// init workspace uuid.
		for {
			if workspaceUUID == "" {
				if err := setWorkspace(hs); err != nil {
					l.Warnf("set workspace failed: %s", err.Error())
				}
			} else {
				break
			}

			select {
			case <-datakit.Exit.Wait():
				l.Info("stop dca")
				return

			case <-tick.C:
			}
		}

		dk := getDatakitData()
		if client, err := ws.NewClient(
			ws.WithWebsocketAddress(hs.dcaConfig.WebsocketServer),
			ws.WithLogger(l),
			ws.WithDataKit(dk),
			ws.WithActionHandlers(HostActionHandlerMap),
		); err != nil {
			l.Errorf("create dca client failed: %s", err.Error())
		} else {
			client.Start()
			g.Go(func(ctx context.Context) error {
				// wait for datakit exit.
				<-datakit.Exit.Wait()
				client.Stop()
				return nil
			})
		}
	}
}
