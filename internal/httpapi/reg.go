// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

type httpRouteInfo struct {
	Method string
	Path   string

	handlerDeprecated gin.HandlerFunc
	handler           APIHandler
}

var httpRouteList = make(map[string]*httpRouteInfo)

// RegHTTPHandler deprecated, use RegHTTPRoute instead.
func RegHTTPHandler(method, path string, handler http.HandlerFunc) {
	httpConfMtx.Lock()
	defer httpConfMtx.Unlock()

	method = strings.ToUpper(method)
	if _, ok := httpRouteList[method+path]; ok {
		l.Warnf("failed to register %s %q by handler %s to HTTP server: route exists",
			method, path, getFunctionName(handler, '/'))
	} else {
		l.Infof("register %s %q by handler %s to HTTP server", method, path, getFunctionName(handler, '/'))
		httpRouteList[method+path] = &httpRouteInfo{
			Method:            method,
			Path:              path,
			handlerDeprecated: func(c *gin.Context) { handler(c.Writer, c.Request) },
		}
	}
}

func RegHTTPRoute(method, path string, handler APIHandler) {
	httpConfMtx.Lock()
	defer httpConfMtx.Unlock()

	method = strings.ToUpper(method)
	if _, ok := httpRouteList[method+path]; ok {
		l.Warnf("failed to register %s@%s to router: route exist.", path, method)
	} else {
		httpRouteList[method+path] = &httpRouteInfo{
			Method:  method,
			Path:    path,
			handler: handler,
		}
	}
}

func RemoveHTTPRoute(method, path string) {
	httpConfMtx.Lock()
	defer httpConfMtx.Unlock()
	delete(httpRouteList, method+path)
}

func CleanHTTPHandler() {
	httpConfMtx.Lock()
	defer httpConfMtx.Unlock()
	httpRouteList = make(map[string]*httpRouteInfo)
}

func addNewRegistedAPIs(hs *httpServerConf) {
	httpConfMtx.Lock()
	defer httpConfMtx.Unlock()

	for _, x := range httpRouteList {
		l.Infof("add %q(method %q) to API white list", x.Path, x.Method)
		// Because API whitelist defauled enabled, we should add new registered APIs to white list.
		hs.apiConfig.PublicAPIs = append(hs.apiConfig.PublicAPIs, x.Path)
	}
}

func applyRegistedAPIs(router *gin.Engine) {
	wrapper1 := &HandlerWrapper{WrappedResponse: true}

	for _, routeInfo := range httpRouteList {
		method := routeInfo.Method
		path := routeInfo.Path

		l.Infof("register %s@%s to HTTP server", method, path)

		switch method {
		case http.MethodPost:
			if routeInfo.handler != nil {
				router.POST(path, wrapper1.RawHTTPWrapper(reqLimiter, routeInfo.handler))
			} else {
				router.POST(path, ginLimiter(reqLimiter), routeInfo.handlerDeprecated)
			}

		case http.MethodGet:
			if routeInfo.handler != nil {
				router.GET(path, wrapper1.RawHTTPWrapper(reqLimiter, routeInfo.handler))
			} else {
				router.GET(path, ginLimiter(reqLimiter), routeInfo.handlerDeprecated)
			}

		case http.MethodHead:
			if routeInfo.handler != nil {
				router.HEAD(path, wrapper1.RawHTTPWrapper(reqLimiter, routeInfo.handler))
			} else {
				router.HEAD(path, ginLimiter(reqLimiter), routeInfo.handlerDeprecated)
			}

		case http.MethodPut:
			if routeInfo.handler != nil {
				router.PUT(path, wrapper1.RawHTTPWrapper(reqLimiter, routeInfo.handler))
			} else {
				router.PUT(path, ginLimiter(reqLimiter), routeInfo.handlerDeprecated)
			}

		case http.MethodPatch:
			if routeInfo.handler != nil {
				router.PATCH(path, wrapper1.RawHTTPWrapper(reqLimiter, routeInfo.handler))
			} else {
				router.PATCH(path, ginLimiter(reqLimiter), routeInfo.handlerDeprecated)
			}

		case http.MethodDelete:

			if routeInfo.handler != nil {
				router.DELETE(path, wrapper1.RawHTTPWrapper(reqLimiter, routeInfo.handler))
			} else {
				router.DELETE(path, ginLimiter(reqLimiter), routeInfo.handlerDeprecated)
			}

		case http.MethodOptions:
			if routeInfo.handler != nil {
				router.OPTIONS(path, wrapper1.RawHTTPWrapper(reqLimiter, routeInfo.handler))
			} else {
				router.OPTIONS(path, ginLimiter(reqLimiter), routeInfo.handlerDeprecated)
			}
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
