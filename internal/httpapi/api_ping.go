// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"fmt"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

type Ping struct {
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
	Host    string `json:"host"`
	Commit  string `json:"commit"`
}

func apiPing(w http.ResponseWriter, r *http.Request, x ...interface{}) (interface{}, error) {
	return &Ping{
		Version: datakit.Version,
		Uptime:  fmt.Sprintf("%v", time.Since(metrics.Uptime)),
		Host:    datakit.DKHost,
		Commit:  datakit.Commit,
	}, nil
}
