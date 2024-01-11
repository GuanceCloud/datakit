// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildLabelsOption(t *testing.T) {
	cases := []struct {
		inAsTagKeys    []string
		inCustomerKeys []string
		out            labelsOption
	}{
		{
			inAsTagKeys:    []string{"app"},
			inCustomerKeys: []string{},
			out:            labelsOption{all: false, keys: []string{"app"}},
		},
		{
			inAsTagKeys:    []string{"app", "name"},
			inCustomerKeys: []string{"sink-project", "name"},
			out:            labelsOption{all: false, keys: []string{"app", "name", "sink-project"}},
		},
		{
			inAsTagKeys:    []string{""},
			inCustomerKeys: []string{"sink-project", "name"},
			out:            labelsOption{all: true, keys: nil},
		},
	}

	for _, tc := range cases {
		res := buildLabelsOption(tc.inAsTagKeys, tc.inCustomerKeys)
		assert.Equal(t, tc.out, res)
	}
}
