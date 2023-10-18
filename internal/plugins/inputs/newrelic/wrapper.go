// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package newrelic handle New Relic APM traces.
package newrelic

import (
	"net/http"
)

func decodingWrapper(next http.HandlerFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		printRequestInfo(req)

		encoding := req.Header.Get("Content-Encoding")
		if encoding == "" {
			next(resp, req)
		} else {
			body, err := decode(encoding, req.Body)
			if err != nil {
				log.Error(err.Error())
				body = req.Body
			}
			req.Body = body
			next(resp, req)
		}
	}
}
