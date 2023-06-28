// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package zipkin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getSplitName(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{
			name: "zipkin:agent-go",
			in:   "zipkin:agent-go{}apiV1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := getSplitName(tc.in)
			require.Equal(t, tc.name, out)
		})
	}
}
