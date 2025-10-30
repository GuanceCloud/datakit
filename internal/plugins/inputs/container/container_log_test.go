// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinHostFilepath(t *testing.T) {
	cases := []struct {
		inHostDir, inInsideDir, inPath string
		out                            string
	}{
		{
			inHostDir:   "/var/lib/kubelet/pods/ABCDEFG012344567/volumes/kubernetes.io~empty-dir/<volume-name>/",
			inInsideDir: "/tmp/log",
			inPath:      "/tmp/log/nginx-log/a.log",
			out:         "/var/lib/kubelet/pods/ABCDEFG012344567/volumes/kubernetes.io~empty-dir/<volume-name>/nginx-log/a.log",
		},
	}

	for _, tc := range cases {
		res := joinHostFilepath(tc.inHostDir, tc.inInsideDir, tc.inPath)
		assert.Equal(t, tc.out, res)
	}
}

func TestJoinInsideFilepath(t *testing.T) {
	cases := []struct {
		inHostDir, inInsideDir, inPath string
		out                            string
	}{
		{
			inHostDir:   "/var/lib/kubelet/pods/ABCDEFG012344567/volumes/kubernetes.io~empty-dir/<volume-name>/",
			inInsideDir: "/tmp/log",
			inPath:      "/var/lib/kubelet/pods/ABCDEFG012344567/volumes/kubernetes.io~empty-dir/<volume-name>/nginx-log/a.log",
			out:         "/tmp/log/nginx-log/a.log",
		},
	}

	for _, tc := range cases {
		res := joinInsideFilepath(tc.inHostDir, tc.inInsideDir, tc.inPath)
		assert.Equal(t, tc.out, res)
	}
}
