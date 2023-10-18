// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package newrelic handle New Relic APM traces.
package newrelic

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

const (
	_encoding    unrecogObject = "encoding" // nolint: staticcheck
	_contentType               = "content type"
	_method                    = "method"
)

type unrecogObject string

func errUnrecognized(object unrecogObject, v string) error {
	return fmt.Errorf("unrecognized %s: %s", object, v)
}

func printRequestInfo(req *http.Request) {
	log.Debugf("####### Request Info(%s) ########", req.Method)

	log.Debugf("### [url] %s", req.URL.String())
	for k, v := range req.Header {
		log.Debugf("### [header] %s: %#v", k, v)
	}
	for k, v := range req.URL.Query() {
		log.Debugf("### [query] %s: %#v", k, v)
	}
}

func printQueryInfo(q *query) {
	log.Debug("@@@@@@@@@ Query Info @@@@@@@@@")

	log.Debugf("@@@ contentType: %v", q.contentType)
	log.Debugf("@@@ contentType: %v", q.contentType)
	log.Debugf("@@@ contentEncoding: %v", q.contentEncoding)
	log.Debugf("@@@ acceptEncoding: %v", q.acceptEncoding)
	log.Debugf("@@@ method: %v", q.method)
	log.Debugf("@@@ license: %v", q.license)
	log.Debugf("@@@ format: %v", q.format)
	log.Debugf("@@@ version: %v", q.version)
	log.Debugf("@@@ runID: %v", q.runID)
	if q.body == nil {
		log.Debugf("@@@ body: %v", q.body)
	} else {
		log.Debugf("@@@ body: %s", q.body.String())
	}
}

func printResponseInfo(resp http.ResponseWriter) {
	log.Debug("$$$$$$$$ Response Info $$$$$$$$")

	for k, v := range resp.Header() {
		log.Debugf("$$$ [header] %s: %#v", k, v)
	}
}

func encode(acceptEncoding string, p []byte) ([]byte, error) {
	var (
		w     = bytes.NewBuffer(nil)
		encwc io.WriteCloser
		err   error
	)
	switch acceptEncoding {
	case "gzip":
		encwc = gzip.NewWriter(w)
	case "deflate":
		encwc, err = flate.NewWriter(w, -1)
	default:
		err = errUnrecognized(_encoding, acceptEncoding)
	}
	if err != nil {
		return nil, err
	}

	if _, err = encwc.Write(p); err != nil {
		return nil, err
	} else {
		encwc.Close() // nolint:errcheck,gosec

		return w.Bytes(), nil
	}
}

func decode(contentEncoding string, body io.ReadCloser) (io.ReadCloser, error) {
	var (
		rc  io.ReadCloser
		err error
	)
	switch contentEncoding {
	case "gzip", "identity":
		rc, err = gzip.NewReader(body)
	case "deflate":
		rc, err = zlib.NewReader(body)
	default:
		err = errUnrecognized(_encoding, contentEncoding)
	}
	if err != nil {
		return nil, err
	}

	defer func() {
		rc.Close()   // nolint:errcheck,gosec
		body.Close() // nolint:errcheck,gosec
	}()

	var (
		newbody = bytes.NewBuffer(nil)
		n       int64
	)
	if n, err = io.Copy(newbody, rc); err != nil { // nolint:gosec
		return nil, err
	} else if n == 0 {
		log.Debug("got empty body")
		newbody.WriteString("[]")
	}

	return io.NopCloser(newbody), nil
}

func serialize(contentType string, v any) ([]byte, error) {
	var (
		buf []byte
		err error
	)
	switch contentType {
	case "application/json":
		buf, err = json.Marshal(v)
	default:
		err = errUnrecognized(_contentType, contentType)
	}

	return buf, err
}

// func deserialize(contentType string, body io.Reader, v any) error {
// 	var err error
// 	switch contentType {
// 	case "application/json":
// 		err = json.NewDecoder(body).Decode(v)
// 	default:
// 		err = errUnrecognized(_contentType, contentType)
// 	}

// 	return err
// }

func stdReply(acceptEncoding, contentType string, resp http.ResponseWriter, v any, err error) {
	if err != nil || v == nil {
		writeEmptyJSON(resp)

		return
	}

	var buf []byte
	if contentType == "application/octet-stream" {
		// contentType = "application/json"
		buf, err = serialize("application/json", v)
	} else {
		buf, err = serialize(contentType, v)
	}
	if err != nil {
		log.Debug(err.Error())
		writeEmptyJSON(resp)

		return
	}

	if buf, err = encode(acceptEncoding, buf); err != nil {
		log.Debug(err.Error())
		writeEmptyJSON(resp)

		return
	}

	resp.Header().Set("Content-Type", contentType)
	resp.Header().Set("Content-Encoding", acceptEncoding)
	printResponseInfo(resp)
	_, err = resp.Write(buf)
	if err != nil {
		log.Debug(err.Error())
	}
}

func writeEmptyJSON(resp http.ResponseWriter) {
	resp.Header().Set("Content-Type", "application/json")
	resp.Header().Set("Content-Encoding", "gzip")

	if buf, err := encode("gzip", []byte("{}")); err != nil {
		log.Debug()
		resp.WriteHeader(http.StatusOK)
	} else {
		resp.Write(buf) // nolint:errcheck,gosec
	}
}

func uintptr(ui uint) *uint { // nolint:predeclared
	return &ui
}

func extractAgentRunID(rawQuery string) string {
	if len(rawQuery) == 0 {
		return ""
	}

	i := strings.Index(rawQuery, "run_id=")
	s := ""
	if i > 0 {
		s = rawQuery[i+7:]
		j := strings.Index(s, "&")
		if j > 0 {
			s = s[:j]
		}
	}

	return s
}

type appWithID struct {
	id  string
	app string
}

func newAppWithID(host, version, identifier string, app string) (string, *appWithID) {
	buf := make([]byte, 30)
	_, _ = rand.Read(buf)
	id := base64.StdEncoding.EncodeToString(buf)

	return fmt.Sprintf("%s^%s^%s", host, version, identifier), &appWithID{id: id, app: app}
}

type appIDMapper struct {
	sync.Mutex
	m map[string]*appWithID
}

func (x *appIDMapper) update(host, version, identifier string, app string) string {
	x.Lock()
	defer x.Unlock()

	key, appid := newAppWithID(host, version, identifier, app)
	x.m[key] = appid

	return appid.id
}

func (x *appIDMapper) find(id string) (string, bool) {
	for _, v := range x.m {
		if v.id == id {
			return v.app, true
		}
	}

	return "", false
}
