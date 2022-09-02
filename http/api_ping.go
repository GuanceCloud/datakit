// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"fmt"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type ping struct {
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
	Host    string `json:"host"`
}

func apiPing(w http.ResponseWriter, r *http.Request, x ...interface{}) (interface{}, error) {
	return &ping{Version: datakit.Version, Uptime: fmt.Sprintf("%v", time.Since(uptime)), Host: datakit.DatakitHostName}, nil
}
