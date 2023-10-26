// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package http wraps all HTTP releated common-used utils.
package http

import (
	"io"
	"net/http"
)

// ReadBody will automatically unzip body.
func ReadBody(req *http.Request) ([]byte, error) {
	buf, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	// as HTTP server, we do not need to close body
	switch req.Header.Get("Content-Encoding") {
	case "gzip":
		return Unzip(buf)
	default:
		return buf, err
	}
}
