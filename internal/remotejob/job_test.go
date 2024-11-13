// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

// Package remotejob is running GuanCe remote job.
package remotejob

import (
	"net/url"
	"os"
	"testing"
)

func TestUrl(t *testing.T) {
	dw := "https://openway.guance.com?token=tkn_xxxxxxxxx"
	u, err := url.ParseRequestURI(dw)
	if err != nil {
		t.Errorf("err=%v", err)
		return
	}
	// u.JoinPath("/v1/datakit/pull")
	u.Path = "/v1/datakit/pull"

	t.Logf("url string=%s", u.String())

	t.Logf("path env:%s", os.Getenv("PATH"))
}
