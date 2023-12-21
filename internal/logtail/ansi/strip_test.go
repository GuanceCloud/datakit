// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ansi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStrip(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{
			in:  "\033[34mhello,world!\033[0m", // foreground-green
			out: "hello,world!",
		},
		{
			in:  "\033[0;32mfoo\033[0m", // foreground-green
			out: "foo",
		},
		{
			in:  "\033[0;31m这是中文字符\033[0m", // foreground-red
			out: "这是中文字符",
		},
	}

	for _, tc := range cases {
		res := Strip(tc.in)
		assert.Equal(t, tc.out, res)
	}
}
