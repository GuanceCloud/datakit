// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineDiff(t *testing.T) {
	cases := []struct {
		in1, in2       string
		inContextLines int
		out            string
	}{
		{
			in1:            "Lorem ipsum dolor",
			in2:            "Lorem dolor sit amet",
			inContextLines: -1,
			out:            "-Lorem ipsum dolor\n+Lorem dolor sit amet",
		},
		{
			in1:            "Lorem ipsum dolor",
			in2:            "Lorem dolor sit amet",
			inContextLines: 0,
			out:            "-Lorem ipsum dolor\n+Lorem dolor sit amet",
		},
		{
			in1:            "Lorem ipsum dolor",
			in2:            "Lorem dolor sit amet",
			inContextLines: 100,
			out:            "-Lorem ipsum dolor\n+Lorem dolor sit amet",
		},
		{
			in1:            "AAAA\nLorem ipsum dolor",
			in2:            "AAAA\nLorem dolor sit amet",
			inContextLines: -1,
			out:            "AAAA\n-Lorem ipsum dolor\n+Lorem dolor sit amet",
		},
		{
			in1:            "AAAA\nLorem ipsum dolor\nBBBB",
			in2:            "AAAA\nLorem dolor sit amet\nBBBB",
			inContextLines: -1,
			out:            "AAAA\n-Lorem ipsum dolor\n+Lorem dolor sit amet\nBBBB",
		},
		{
			in1:            "AAAA\nBBBB\nCCCC\nLorem ipsum dolor\nEEEE\nFFFF\nGGGG",
			in2:            "AAAA\nBBBB\nCCCC\nLorem dolor sit amet\nEEEE\nFFFF\nGGGG",
			inContextLines: -1,
			out:            "AAAA\nBBBB\nCCCC\n-Lorem ipsum dolor\n+Lorem dolor sit amet\nEEEE\nFFFF\nGGGG",
		},
		{
			in1:            "AAAA\nBBBB\nCCCC\nLorem ipsum dolor\nEEEE\nFFFF\nGGGG",
			in2:            "AAAA\nBBBB\nCCCC\nLorem dolor sit amet\nEEEE\nFFFF\nGGGG",
			inContextLines: 1,
			out:            "CCCC\n-Lorem ipsum dolor\n+Lorem dolor sit amet\nEEEE",
		},
		{
			in1:            "AAAA\nBBBB\nCCCC\nLorem ipsum dolor\nEEEE\nFFFF\nGGGG",
			in2:            "AAAA\nBBBB\nCCCC\nLorem dolor sit amet\nEEEE\nFFFF\nGGGG",
			inContextLines: 2,
			out:            "BBBB\nCCCC\n-Lorem ipsum dolor\n+Lorem dolor sit amet\nEEEE\nFFFF",
		},
		{
			in1:            "AAAA\nBBBB\nCCCC\nLorem ipsum dolor\nDDDD\nHello World!\nEEEE\nFFFF\nGGGG",
			in2:            "AAAA\nBBBB\nCCCC\nLorem dolor sit amet\nDDDD\nHello Tony!\nEEEE\nFFFF\nGGGG",
			inContextLines: 2,
			out:            "BBBB\nCCCC\n-Lorem ipsum dolor\n+Lorem dolor sit amet\nDDDD\n-Hello World!\n+Hello Tony!\nEEEE\nFFFF",
		},
		{
			in1:            "AAAA\nBBBB\nCCCC\nLorem ipsum dolor\nDDDD\nEEEE\nHello World!\nFFFF\nGGGG",
			in2:            "AAAA\nBBBB\nCCCC\nLorem dolor sit amet\nDDDD\nEEEE\nHello Tony!\nFFFF\nGGGG",
			inContextLines: 1,
			out:            "CCCC\n-Lorem ipsum dolor\n+Lorem dolor sit amet\nDDDD\n@@ ...\nEEEE\n-Hello World!\n+Hello Tony!\nFFFF",
		},
		{
			in1:            "AAAA\nBBBB\nCCCC\nLorem ipsum dolor\nDDDD\nEEEE\nHello World!\nFFFF\nGGGG",
			in2:            "AAAA\nBBBB\nCCCC\nLorem dolor sit amet\nDDDD\nEEEE\nHello Tony!\nFFFF\nGGGG",
			inContextLines: 2,
			out:            "BBBB\nCCCC\n-Lorem ipsum dolor\n+Lorem dolor sit amet\nDDDD\nEEEE\n-Hello World!\n+Hello Tony!\nFFFF\nGGGG",
		},
		{
			in1:            "AAAA\nBBBB\nCCCC\nLorem ipsum dolor\nDDDD\nEEEE\nFFFF\nHello World!\nGGGG\nHHHH",
			in2:            "AAAA\nBBBB\nCCCC\nLorem dolor sit amet\nDDDD\nEEEE\nFFFF\nHello Tony!\nGGGG\nHHHH",
			inContextLines: 2,
			out:            "BBBB\nCCCC\n-Lorem ipsum dolor\n+Lorem dolor sit amet\nDDDD\nEEEE\n@@ ...\nFFFF\n-Hello World!\n+Hello Tony!\nGGGG\nHHHH",
		},
		{
			in1:            "AAAA\nBBBB\nCCCC\nLorem ipsum dolor\nDDDD\nEEEE\nFFFF\nGGGG\nHello World!\nHHHH\nLLLL",
			in2:            "AAAA\nBBBB\nCCCC\nLorem dolor sit amet\nDDDD\nEEEE\nFFFF\nGGGG\nHello Tony!\nHHHH\nLLLL",
			inContextLines: 2,
			out:            "BBBB\nCCCC\n-Lorem ipsum dolor\n+Lorem dolor sit amet\nDDDD\nEEEE\n@@ ...\nFFFF\nGGGG\n-Hello World!\n+Hello Tony!\nHHHH\nLLLL",
		},
		{
			in1: `
AAAA
BBBB
CCCC
Hello, World!
DDDD
EEEE
FFFF
GGGG
This is file1.
HHHH
`,
			in2: `
AAAA
BBBB
CCCC
Hello, Git Diff!
DDDD
EEEE
FFFF
GGGG
This is a modified file1.
HHHH
`,
			inContextLines: 2,
			out: `BBBB
CCCC
-Hello, World!
+Hello, Git Diff!
DDDD
EEEE
@@ ...
FFFF
GGGG
-This is file1.
+This is a modified file1.
HHHH`,
		},
	}

	for _, tc := range cases {
		res := LineDiffWithContextLines(tc.in1, tc.in2, tc.inContextLines)
		assert.Equal(t, tc.out, res)
	}
}
