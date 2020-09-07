package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

type RegHttpInfo struct {
	Method  string
	Path    string
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

func applyHTTPRoute(router *gin.Engine) {
	for _, regInfo := range httpRegList {
		method := strings.ToUpper(regInfo.Method)
		path := regInfo.Path
		handler := regInfo.Handler

		l.Infof("register %s %s by handler %s to HTTP server", method, path, getFunctionName(handler, '/'))

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

func getFunctionName(i interface{}, seps ...rune) string {
	fn := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()

	fields := strings.FieldsFunc(fn, func(sep rune) bool {
		for _, s := range seps {
			if sep == s {
				return true
			}
		}
		return false
	})

	if size := len(fields); size > 0 {
		return fields[size-1]
	}
	return ""
}
