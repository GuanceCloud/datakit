// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseImage(t *testing.T) {
	cases := []struct {
		in              string
		outImageName    string
		outShortName    string
		outImageVersion string
	}{
		{
			"docker.io/library/busybox:latest",
			"docker.io/library/busybox",
			"busybox",
			"latest",
		},
	}

	for _, tc := range cases {
		imageName, shortName, imageVersion := ParseImage(tc.in)
		assert.Equal(t, tc.outImageName, imageName)
		assert.Equal(t, tc.outShortName, shortName)
		assert.Equal(t, tc.outImageVersion, imageVersion)
	}
}

func TestDigestImage(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{
			"docker.io/library/busybox:latest@sha256:7cc4b5aefd1d0cadf8d97d4350462ba51c694ebca145b08d7d41b41acc8db5aa",
			"docker.io/library/busybox:latest",
		},
	}

	for _, tc := range cases {
		image := digestImage(tc.in)
		assert.Equal(t, tc.out, image)
	}
}
