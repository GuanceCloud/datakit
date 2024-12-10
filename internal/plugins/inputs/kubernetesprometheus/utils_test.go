// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToEnvironmentValue(t *testing.T) {
	testcases := []struct {
		name             string
		envKey, envValue string
		in               string
		out              string
	}{
		{
			envKey:   "NAME",
			envValue: "hello-world",
			in:       "name=$(NAME)",
			out:      "name=hello-world",
		},
		{
			envKey:   "NAME",
			envValue: "hello-world",
			in:       "name=$(NAME-02)",
			out:      "name=$(NAME-02)",
		},
		{
			envKey:   "NAME",
			envValue: "hello-world",
			in:       "name=$(NAME),name02=$(NAME)",
			out:      "name=hello-world,name02=$(NAME)",
		},
	}

	for _, tc := range testcases {
		err := os.Setenv(tc.envKey, tc.envValue)
		assert.NoError(t, err)

		res := convertToEnvironmentValue(tc.in)
		assert.Equal(t, tc.out, res)

		err = os.Unsetenv(tc.envKey)
		assert.NoError(t, err)
	}
}
