// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"net/http/httptest"
	"strings"
	T "testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Test_trimFileName(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		src := " ./abc.1,  ./def.1   "
		arr := strings.Split(src, ",")

		assert.Equal(t, `abc.1`, trimFileName(arr[0], "./"))
		assert.Equal(t, `def.1`, trimFileName(arr[1], "./"))

		src = ` .\abc.1,  .\def.1   `
		arr = strings.Split(src, ",")

		assert.Equal(t, `abc.1`, trimFileName(arr[0], `\.`))
		assert.Equal(t, `def.1`, trimFileName(arr[1], `\.`))
	})
}

func TestCheckIsVersion(t *T.T) {
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
		t.Run(tc.ver, func(t *T.T) {
			err := checkIsNewVersion(ts.URL, tc.ver)
			if tc.fail {
				assert.Error(t, err)
				t.Logf("expect err: %s", err)

				return
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
