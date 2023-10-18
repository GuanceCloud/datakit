// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package newrelic handle New Relic APM traces.
package newrelic

import (
	"bytes"
	"io"
	"net/http"
)

type query struct {
	contentType     string
	contentEncoding string
	acceptEncoding  string
	method          string
	license         string
	format          string
	version         string
	runID           string
	body            *bytes.Buffer
}

func httpStatusRespFunc(resp http.ResponseWriter, req *http.Request, err error) {}

func handleRawMethod(resp http.ResponseWriter, req *http.Request) {
	buf := bytes.NewBuffer(nil)
	_, err := io.Copy(buf, req.Body)
	if err != nil {
		log.Debugf(err.Error())
		writeEmptyJSON(resp)

		return
	}

	q := &query{
		contentType:     req.Header.Get("Content-Type"),
		contentEncoding: req.Header.Get("Content-Encoding"),
		acceptEncoding:  req.Header.Get("Accept-Encoding"),
		method:          req.URL.Query().Get("method"),
		license:         req.URL.Query().Get("license_key"),
		format:          req.URL.Query().Get("marshal_format"),
		version:         req.URL.Query().Get("protocol_version"),
		runID:           extractAgentRunID(req.URL.RawQuery),
		body:            buf,
	}

	proc, ok := compound[q.method]
	if !ok {
		err = errUnrecognized(_method, q.method)
		log.Debug(err.Error())
		stdReply(q.acceptEncoding, q.contentType, resp, nil, err)

		return
	}
	proc.ProcessMethod(resp, q)
}
