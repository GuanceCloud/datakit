// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package apache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getVersion(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{
			name: "2.4.6",
			in:   "httpd:2.4.6-alpine-server-status",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := getVersion(tc.in)
			require.Equal(t, tc.name, out)
		})
	}
}
