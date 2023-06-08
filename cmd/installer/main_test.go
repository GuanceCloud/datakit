// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	tu "github.com/GuanceCloud/cliutils/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestPromptFixVersionChecking(t *testing.T) {
	promptFixVersionChecking()
}

func TestCheckIsVersion(t *testing.T) {
	r := gin.New()
	r.GET("/v1/ping", func(c *gin.Context) {
		c.Data(200, "", []byte(`{ "content":{ "version": "1.2.3", "uptime": "30m", "host": "wtf" }}`))
	})

	_ = r

	ts := httptest.NewServer(r)
	time.Sleep(time.Second)
	defer ts.Close()

	cases := []struct {
		ver  string
		fail bool
	}{
		{
			ver: "1.2.3",
		},
		{
			ver:  "1.2.4",
			fail: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.ver, func(t *testing.T) {
			err := checkIsNewVersion(ts.URL, tc.ver)
			if tc.fail {
				tu.NotOk(t, err, "expect err, not nil")
				t.Logf("expect err: %s", err)

				promptFixVersionChecking()
				return
			} else {
				tu.Ok(t, err)
			}
		})
	}
}

func Test_checkCmd(t *testing.T) {
	cases := []struct {
		name       string
		candidates []string
		expectOut  string
		expect     error
	}{
		{
			name:       "macos",
			candidates: []string{"md5", "md6"},
			expectOut:  "md5",
		},
		{
			name:       "linux",
			candidates: []string{"md5sum", "md6sum"},
			expectOut:  "md5sum",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			switch tc.name {
			case "macos":
				if runtime.GOOS != datakit.OSDarwin {
					return
				}
			case "linux":
				if runtime.GOOS != datakit.OSLinux {
					return
				}
			}

			out, err := checkCmd(tc.candidates...)
			assert.Equal(t, tc.expect, err)
			assert.Equal(t, tc.expectOut, out)
		})
	}
}
