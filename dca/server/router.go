// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	tollbooth "github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	authLogoutPath         = "authLogout"
	workspaceListPath      = "workspaceList"
	currentAccountPath     = "currentAccount"
	accountPermissionsPath = "accountPermissions"
	changeWorkspacePath    = "changeWorkspace"
	currentWorkspacePath   = "currentWorkspace"
	datakitListPath        = "datakitList"
	getTokenByCodePath     = "getTokenByCode"
	frontTokenPath         = "frontToken"

	cookieWorkspaceUUID = "workspace_uuid"
	cookieFrontToken    = "front_token"
)

var (
	consoleWebURL = fmt.Sprintf("https://console.%s", datakit.BrandDomain)
	consoleAPIURL = fmt.Sprintf("https://console-api.%s", datakit.BrandDomain)
	staticBaseURL = fmt.Sprintf("https://static.%s", datakit.BrandDomain)
	dcaHTTPPort   = "80"
	ignoreAuthURI = []string{
		"/ws",
		"/console/dca",
		"/sso/login",
		"/login",
		"/metrics",
		"/api/console/logout",
	}

	consoleAPIInfo = map[string][2]string{
		authLogoutPath:         {http.MethodGet, "/auth-token/logout"},
		workspaceListPath:      {http.MethodGet, "/workspace/query_list"},
		currentAccountPath:     {http.MethodGet, "/account/current"},
		accountPermissionsPath: {http.MethodGet, "/workspace/account/permissions"},
		changeWorkspacePath:    {http.MethodPost, "/workspace/change"},
		currentWorkspacePath:   {http.MethodGet, "/workspace/current"},
		datakitListPath:        {http.MethodGet, "/datakit/list"},
		getTokenByCodePath:     {http.MethodGet, "/auth-token/get"},
		frontTokenPath:         {http.MethodGet, "/workspace/token/get"},
	}
)

var nlimiter = tollbooth.NewLimiter(1000, &limiter.ExpirableOptions{
	DefaultExpirationTTL: time.Second,
})

func limiterHandler(lmt *limiter.Limiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		h := newHandler(ctx)
		httpError := tollbooth.LimitByRequest(lmt, ctx.Writer, ctx.Request)
		if httpError != nil {
			h.fail(http.StatusTooManyRequests, "too.many.requests", "too many requests")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

func setupRouter(router *gin.Engine) error {
	router.Use(gin.Recovery())
	router.GET("/ws", limiterHandler(nlimiter), websocketHandler)

	apiRouter := router.Group("/api")
	apiRouter.GET("/lastDatakitVersion", getLastDatakitVersionHandler)
	consoleRouter := apiRouter.Group("/console")
	datakitRouter := apiRouter.Group("/datakit")

	consoleRouter.GET("/workspaceList", consoleHandler(workspaceListPath))
	consoleRouter.GET("/currentWorkspace", consoleHandler(currentWorkspacePath))
	consoleRouter.GET("/currentAccount", consoleHandler(currentAccountPath))
	consoleRouter.GET("/accountPermissions", consoleHandler(accountPermissionsPath))
	consoleRouter.POST("/changeWorkspace", consoleChangeWorkspaceHandler)
	consoleRouter.POST("/logout", logoutHandler)

	datakitRouter.Use(auth)
	datakitRouter.GET("/ws/log", websocketLogHandler)
	datakitRouter.GET("/stats", datakitHandler(ws.GetDatakitStatsAction))
	datakitRouter.GET("/list", datakitListHandler)
	datakitRouter.GET("/listByID", datakitByIDHandler)
	datakitRouter.GET("/searchValue", datakitSearchValueHandler)
	datakitRouter.PUT("/reload", datakitHandler(ws.ReloadDatakitAction))
	datakitRouter.PUT("/restart", datakitHandler(ws.RestartDatakitAction))
	datakitRouter.POST("/upgrade", datakitHandler(ws.UpgradeDatakitAction))
	datakitRouter.POST("/operation/:type", datakitOperationHandler)

	// config
	datakitRouter.GET("/getConfig", datakitHandler(ws.GetDatakitConfigAction))
	datakitRouter.POST("/saveConfig", datakitHandler(ws.SaveDatakitConfigAction))
	datakitRouter.DELETE("/deleteConfig", datakitHandler(ws.DeleteDatakitConfigAction))

	// pipeline
	datakitRouter.GET("/pipelines", datakitHandler(ws.GetDatakitPipelineAction))
	datakitRouter.GET("/pipelines/detail", datakitHandler(ws.GetDatakitPipelineDetailAction))
	datakitRouter.PATCH("/pipelines", datakitHandler(ws.PatchDatakitPipelineAction))
	datakitRouter.POST("/pipelines", datakitHandler(ws.CreateDatakitPipelineAction))
	datakitRouter.DELETE("/pipelines", datakitHandler(ws.DeleteDatakitPipelineAction))
	datakitRouter.POST("/pipelines/test", datakitHandler(ws.TestDatakitPipelineAction))

	// filter
	datakitRouter.GET("/filter", datakitHandler(ws.GetDatakitFilterAction))

	// log
	datakitRouter.GET("/log/download", datakitLogDownladHandler)

	router.GET("/sso/login", ssoLoginHandler)
	router.GET("/console/dca", consoleRedirectHandler)

	router.StaticFile("/", "./public/index.html")
	router.Static("/public", "./public")
	router.NoRoute(func(ctx *gin.Context) {
		ctx.File("./public/index.html")
	})

	return nil
}

type accountPermissionsContent struct {
	Permissions []string `json:"permissions"`
	Role        string   `json:"role"`
	Roles       []string `json:"roles"`
}

func auth(ctx *gin.Context) {
	if ctx.Request.Method == "OPTIONS" {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}

	fullPath := ctx.Request.URL.Path
	// public static files
	if strings.HasPrefix(fullPath, "/public") {
		ctx.Next()
		return
	}
	for _, uri := range ignoreAuthURI {
		if uri == fullPath {
			ctx.Next()
			return
		}
	}
	h := newHandler(ctx)
	api := consoleAPIInfo[accountPermissionsPath]

	req, err := http.NewRequest(api[0], getConsoleAPIURL(api[1]), nil)
	if err != nil {
		l.Errorf("failed to new request: %s", err.Error())
		h.fail(500, "auth.failed", "auth failed")
		ctx.Abort()
		return
	}

	fontToken, _ := ctx.Cookie(cookieFrontToken)
	workspaceUUID, _ := ctx.Cookie(cookieWorkspaceUUID)

	req.Header.Set("X-FT-Auth-Token", fontToken)
	req.Header.Set("X-Workspace-Uuid", workspaceUUID)

	respbody, err := h.doRequest(req)
	if err != nil {
		l.Errorf("failed to do request: %s", err.Error())
		h.fail(500, "auth.failed", err.Error())
		ctx.Abort()
		return
	}

	content := accountPermissionsContent{}
	resp := Response{
		DCAResponse: ws.DCAResponse{
			Content: &content,
		},
	}

	if err := json.Unmarshal(respbody, &resp); err != nil {
		l.Errorf("failed to decode response from console: %s", err.Error())
		h.fail(500, "auth.failed", err.Error())
		ctx.Abort()
		return
	}

	if !resp.Success {
		h.send(&ws.DCAResponse{
			Success:   false,
			Message:   resp.Message,
			ErrorCode: resp.ErrorCode,
			Code:      401,
		})
		ctx.Abort()
		return
	}

	if ctx.Request.Method == http.MethodGet {
		ctx.Next()
		return
	}

	// check container mode, only readonly operation is allowed
	if dk := h.getDatakit(); dk != nil && dk.RunInContainer {
		h.send(&ws.DCAResponse{
			Success:   false,
			Message:   "no permission in container mode",
			ErrorCode: "permission.denied.container",
			Code:      403,
		})
		ctx.Abort()
		return
	}

	// check dca permission
	for _, p := range content.Permissions {
		if p == "dca.*" {
			ctx.Next()
			return
		}
	}

	h.send(&ws.DCAResponse{
		Success:   false,
		Message:   "no permission",
		ErrorCode: "permission.denied",
		Code:      403,
	})
	ctx.Abort()
}
