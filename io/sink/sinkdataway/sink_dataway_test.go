// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sinkdataway

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v -timeout 30s -run ^TestGetURLFromMapConfig$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkdataway
func TestGetURLFromMapConfig(t *testing.T) {
	cases := []struct {
		name string
		in   map[string]interface{}
		out  string
	}{
		{
			name: "normal",
			in: map[string]interface{}{
				"target": "dataway",
				"url":    "https://openway.guance.com?token=tkn_xxxxx",
				"proxy":  "127.0.0.1:1080",
				"host":   "",
			},
			out: "https://openway.guance.com?token=tkn_xxxxx",
		},
		{
			name: "env",
			in: map[string]interface{}{
				"target": "dataway",
				"url":    "https://openway.guance.com",
				"token":  "tkn_xxxxx",
				"proxy":  "127.0.0.1:1080",
				"host":   "",
			},
			out: "https://openway.guance.com?token=tkn_xxxxx",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := getURLFromMapConfig(tc.in)
			assert.NoError(t, err)
			assert.Equal(t, tc.out, out)
		})
	}
}
