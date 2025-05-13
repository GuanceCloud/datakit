// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"net/http"
	"time"
)

type ntpResp struct {
	TimestampSec int64 `json:"timestamp_sec"`
}

func apiNTP(w http.ResponseWriter, r *http.Request, x ...interface{}) (interface{}, error) {
	return &ntpResp{
		TimestampSec: time.Now().Unix(),
	}, nil
}
