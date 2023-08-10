// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDockerEnvs(t *testing.T) {
	cases := []struct {
		in  []string
		out map[string]string
	}{
		{
			in:  []string{"PATH=/usr/local/sbin:/usr/local/bin"},
			out: map[string]string{"PATH": "/usr/local/sbin:/usr/local/bin"},
		},
		{
			// not '='
			in:  []string{"PATH-ABC"},
			out: map[string]string{"PATH-ABC": ""},
		},
		{
			// doubel '='
			in:  []string{"PATH=/usr/local/sbin:=/usr/local/bin"},
			out: map[string]string{"PATH": "/usr/local/sbin:/usr/local/bin"},
		},
	}

	for _, tc := range cases {
		res := parseDockerEnv(tc.in)
		assert.Equal(t, tc.out, res)
	}
}
