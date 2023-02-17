// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"net/http"
	"runtime/debug"

	"github.com/GuanceCloud/cliutils/logger"
)

func ProtectedHandlerFunc(next http.HandlerFunc, log *logger.Logger) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		log.Debugf("### protected mode HTTP handler for pattern: %s", req.URL.Path)

		defer func() {
			if r := recover(); r != nil {
				log.Errorf("Stack crash: %v", r)
				log.Errorf("Stack info :%s", string(debug.Stack()))
			}
		}()

		next(resp, req)
	}
}

func CheckExpectedHeaders(next http.HandlerFunc, log *logger.Logger, expectedHeaders map[string][]string) http.HandlerFunc {
	if len(expectedHeaders) == 0 {
		return next
	}

	return func(resp http.ResponseWriter, req *http.Request) {
		for k, v := range expectedHeaders {
			if ss, ok := req.Header[k]; ok {
				for i := range ss {
					for j := range v {
						if ss[i] == v[j] {
							next(resp, req)

							return
						}
					}
				}
			}
		}

		log.Debug("### expected HTTP header not found")
		resp.WriteHeader(http.StatusBadRequest)
	}
}
