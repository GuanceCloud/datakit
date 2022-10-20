// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"io"
	"net/http"
	"runtime/debug"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
)

func ProtectedHandlerFunc(handler http.HandlerFunc, log *logger.Logger) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		log.Debugf("### protected mode HTTP handler for pattern: %s", req.URL.Path)

		defer func() {
			if r := recover(); r != nil {
				log.Errorf("Stack crash: %v", r)
				log.Errorf("Stack info :%s", string(debug.Stack()))
			}
		}()

		handler(resp, req)
	}
}

type validator func(req *http.Request) error

func ReadBodyHandlerFunc(handler http.HandlerFunc, log *logger.Logger, validators ...validator) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		log.Debugf("### read HTTP body and return mode for pattern: %s", req.URL.Path)

		for i := range validators {
			if err := validators[i](req); err != nil {
				log.Error(err.Error())
				resp.WriteHeader(http.StatusBadRequest)

				return
			}
		}

		pbuf := bufpool.GetBuffer()
		defer bufpool.PutBuffer(pbuf)

		_, err := io.Copy(pbuf, req.Body)
		if err != nil {
			log.Error(err.Error())
			resp.WriteHeader(http.StatusBadRequest)

			return
		}
		resp.WriteHeader(http.StatusOK)

		req.Body.Close() // nolint: errcheck,gosec
		req.Body = io.NopCloser(pbuf)
		handler(&NopResponseWriter{resp}, req)
	}
}
