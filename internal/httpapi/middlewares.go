// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"io"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/GuanceCloud/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	"google.golang.org/protobuf/proto"
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

func GetHeader(req *http.Request, key string) string {
	return strings.ToLower(strings.TrimSpace(req.Header.Get(key)))
}

type NopResponseWriter struct {
	Raw http.ResponseWriter
}

func (nop *NopResponseWriter) Header() http.Header { return make(http.Header) }

func (nop *NopResponseWriter) Write([]byte) (int, error) { return 0, nil }

func (nop *NopResponseWriter) WriteHeader(statusCode int) {}

type HTTPStatusResponse func(resp http.ResponseWriter, req *http.Request, err error)

func HTTPStorageWrapper(key uint8,
	statRespFunc HTTPStatusResponse,
	s *storage.Storage,
	handler http.HandlerFunc,
) http.HandlerFunc {
	if s == nil || !s.Enabled() {
		return handler
	} else {
		return func(resp http.ResponseWriter, req *http.Request) {
			pbuf := bufpool.GetBuffer()
			defer bufpool.PutBuffer(pbuf)

			_, err := io.Copy(pbuf, req.Body)
			if err != nil {
				l.Error(err.Error())
				resp.WriteHeader(http.StatusBadRequest)

				return
			}

			reqpb := &storage.Request{
				Method:           req.Method,
				Url:              req.URL.String(),
				Proto:            req.Proto,
				ProtoMajor:       int32(req.ProtoMajor),
				ProtoMinor:       int32(req.ProtoMinor),
				Header:           storage.ConvertMapToMapEntries(req.Header),
				Body:             pbuf.Bytes(),
				ContentLength:    req.ContentLength,
				TransferEncoding: req.TransferEncoding,
				Close:            req.Close,
				Host:             req.Host,
				Form:             storage.ConvertMapToMapEntries(req.Form),
				PostForm:         storage.ConvertMapToMapEntries(req.PostForm),
				RemoteAddr:       req.RemoteAddr,
				RequestUri:       req.RequestURI,
			}
			buf, err := proto.Marshal(reqpb)
			if err != nil {
				l.Error(err.Error())
				resp.WriteHeader(http.StatusBadRequest)

				return
			}

			if err = s.Put(key, buf); err != nil {
				l.Error(err.Error())
				statRespFunc(resp, req, err)
			} else {
				l.Debug("HTTP wrapper: new data put into local-cache success")
				statRespFunc(resp, req, nil)
			}
		}
	}
}
