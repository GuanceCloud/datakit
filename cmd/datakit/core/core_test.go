// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	giturls "github.com/whilp/git-urls"
)

func TestGitURL(t *testing.T) {
	const e = "https://username:password@github.com/username/repository.git"

	u, err := giturls.Parse(e)
	assert.NoError(t, err)

	assert.Equal(t, "username", u.User.Username())
	pass, ok := u.User.Password()
	assert.Equal(t, "password", pass)
	assert.True(t, ok)
}

func TestParseCurrentCPUMax(t *testing.T) {
	// Mock getNumCPU function
	originalGetNumCPU := getNumCPU
	defer func() { getNumCPU = originalGetNumCPU }()
	getNumCPU = func() int { return 4 }

	c := &Core{}

	testcases := []struct {
		inQuota, inPeriod string
		fail              bool
		value             float64
	}{
		{
			inQuota:  "100000",
			inPeriod: "100000",
			value:    1,
		},
		{
			inQuota:  "800000",
			inPeriod: "100000",
			value:    4, // max NumCPU
		},
		{
			inQuota:  "800000",
			inPeriod: "200000",
			value:    4,
		},
		{
			inQuota:  "max",
			inPeriod: "100000",
			value:    4,
		},
		{
			inQuota:  "-1",
			inPeriod: "100000",
			value:    4,
		},
		{
			inQuota:  "invalid",
			inPeriod: "",
			fail:     true,
		},
		{
			inQuota:  "",
			inPeriod: "invalid",
			fail:     true,
		},
		{
			inQuota:  "0",
			inPeriod: "100000",
			fail:     true,
		},
		{
			inQuota:  "-100000",
			inPeriod: "100000",
			fail:     true,
		},
		{
			inQuota:  "100000",
			inPeriod: "0",
			fail:     true,
		},
	}

	for _, tc := range testcases {
		res, err := c.parseCurrentCPUMax(tc.inQuota, tc.inPeriod)
		if tc.fail {
			assert.Error(t, err)
			continue
		}

		assert.Nil(t, err)
		assert.Equal(t, tc.value, res)
	}
}

func TestParseUserNamePasswd(t *testing.T) {
	gitURLs := []string{
		"https://username:password@github.com/username/repository.git",
		"https://username@github.com/username/repository.git",
	}

	for _, v := range gitURLs {
		t.Log("\n--------------------------\n")

		_, err := giturls.Parse(v)
		if err != nil {
			t.Logf("parse [%s] failed: [%v]", v, err)
			continue
		}
		t.Logf("parse [%s] ok", v)
	}

	t.Log("\n--------------------------\n")
	t.Log("parse all ok")
}
