package http

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"strings"
)

type RegHttpInfo struct {
	Method string
	Path   string
	Handler gin.HandlerFunc
}

var (
	httpRegList = make([]*RegHttpInfo, 0, 16)
)

func RegHttpHandler(method, path string, handler gin.HandlerFunc) {
	regInfo := &RegHttpInfo{
		method,
		path,
		handler,
	}
	httpRegList = append(httpRegList, regInfo)
}


func RegPathToHttpServ(router *gin.Engine) {
	for _, regInfo := range httpRegList {
		method  := strings.ToUpper(regInfo.Method)
		path    := regInfo.Path
		handler := regInfo.Handler
		l.Infof("register %s %s to HTTP server", method, path)
		switch method {
		case http.MethodPost:
			router.POST(path, handler)
		case http.MethodGet:
			router.GET(path, handler)
		case http.MethodHead:
			router.HEAD(path, handler)
		case http.MethodPut:
			router.PUT(path, handler)
		case http.MethodPatch:
			router.PATCH(path, handler)
		case http.MethodDelete:
			router.DELETE(path, handler)
		case http.MethodOptions:
			router.OPTIONS(path, handler)
		}
	}
}