// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package storage

import (
	"io"
	"net/http"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
	"google.golang.org/protobuf/proto"
)

func HTTPWrapper(key uint8, s *Storage, handler http.HandlerFunc) http.HandlerFunc {
	if s == nil || !s.enabled {
		return handler
	} else {
		return func(resp http.ResponseWriter, req *http.Request) {
			pbuf := bufpool.GetBuffer()
			defer bufpool.PutBuffer(pbuf)

			_, err := io.Copy(pbuf, req.Body)
			if err != nil {
				s.log.Error(err.Error())
				resp.WriteHeader(http.StatusBadRequest)

				return
			}

			reqpb := &Request{
				Method:           req.Method,
				Url:              req.URL.String(),
				Proto:            req.Proto,
				ProtoMajor:       int32(req.ProtoMajor),
				ProtoMinor:       int32(req.ProtoMinor),
				Header:           ConvertMapToMapEntries(req.Header),
				Body:             pbuf.Bytes(),
				ContentLength:    req.ContentLength,
				TransferEncoding: req.TransferEncoding,
				Close:            req.Close,
				Host:             req.Host,
				Form:             ConvertMapToMapEntries(req.Form),
				PostForm:         ConvertMapToMapEntries(req.PostForm),
				RemoteAddr:       req.RemoteAddr,
				RequestUri:       req.RequestURI,
			}
			buf, err := proto.Marshal(reqpb)
			if err != nil {
				s.log.Error(err.Error())
				resp.WriteHeader(http.StatusBadRequest)

				return
			}

			if err = s.Put(key, buf); err != nil {
				s.log.Error(err.Error())
				resp.WriteHeader(http.StatusBadRequest)
			} else {
				s.log.Debug("HTTP wrapper: put new data into local-cache success")
				resp.WriteHeader(http.StatusOK)
			}
		}
	}
}

func ConvertMapToMapEntries(src map[string][]string) (dst []*MapEntry) {
	if len(src) == 0 {
		return nil
	}

	dst = make([]*MapEntry, len(src))
	i := 0
	for k, v := range src {
		dst[i] = &MapEntry{
			Key:   k,
			Value: v,
		}
		i++
	}

	return
}

func ConvertMapEntriesToMap(src []*MapEntry) (dst map[string][]string) {
	if len(src) == 0 {
		return nil
	}

	dst = make(map[string][]string)
	for _, v := range src {
		dst[v.Key] = v.Value
	}

	return
}
