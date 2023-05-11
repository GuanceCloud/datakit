// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nginx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getAmendName(t *testing.T) {
	cases := []struct {
		name      string
		savedName string
		in        string
	}{
		{
			name:      "nginx:vts-1.8.0-alpine",
			savedName: "nginx:vts-1.8.0-alpine____using-vts",
			in:        "nginx:vts-1.8.0-alpine____using-vts",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out1, out2 := getAmendName(tc.in)
			require.Equal(t, tc.name, out1)
			require.Equal(t, tc.savedName, out2)
		})
	}
}
