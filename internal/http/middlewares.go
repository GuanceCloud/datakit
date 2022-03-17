// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"net/http"
	"runtime/debug"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

func ProtectedHandlerFunc(handler http.HandlerFunc, log *logger.Logger) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		log.Debugf("protected HTTP handler for pattern: %s", req.URL.Path)
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("Stack crash: %v", r)
				log.Errorf("Stack info :%s", string(debug.Stack()))
			}
		}()

		handler(resp, req)
	}
}
