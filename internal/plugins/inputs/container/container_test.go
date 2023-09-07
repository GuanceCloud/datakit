// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitRules(t *testing.T) {
	cases := []struct {
		in  []string
		out []string
	}{
		{
			in:  []string{"image:*"},
			out: []string{"**"},
		},
		{
			in:  []string{"image:pubrepo.guance.com/datakit*", "image:*"},
			out: []string{"pubrepo.guance.com/datakit*", "**"},
		},
		{
			in:  []string{"image:pubrepo.guance.com/datakit/datakit:*"},
			out: []string{"pubrepo.guance.com/datakit/datakit:*"},
		},
	}

	for _, tc := range cases {
		res := splitRules(tc.in)
		assert.Equal(t, tc.out, res)
	}
}
