// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package labels

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSelector(t *testing.T) {
	testcases := []struct {
		inSelector string
		inLabels   map[string]string
		outMatched bool
		fail       bool
	}{
		{
			inSelector: "app=middleware-nginx",
			inLabels: map[string]string{
				"app":  "middleware-nginx",
				"type": "testing",
			},
			outMatched: true,
		},
		{
			inSelector: "app=middleware*",
			inLabels: map[string]string{
				"app":  "middleware-nginx",
				"type": "testing",
			},
			outMatched: true,
		},
		{
			inSelector: "app==middleware*", // double equals
			inLabels: map[string]string{
				"app":  "middleware-nginx",
				"type": "testing",
			},
			outMatched: true,
		},
		{
			inSelector: "app!=middleware",
			inLabels: map[string]string{
				"app":  "middleware-nginx",
				"type": "testing",
			},
			outMatched: true,
		},
		{
			inSelector: "app!=middleware-nginx",
			inLabels: map[string]string{
				"app":  "middleware-nginx",
				"type": "testing",
			},
			outMatched: false,
		},
		{
			inSelector: "app=!middleware", // unable to parse requirement: found '!', expected: identifier
			inLabels: map[string]string{
				"app":  "middleware-nginx",
				"type": "testing",
			},
			fail: true,
		},
		{
			inSelector: "app=middleware*,type!=testing", // type equal testing
			inLabels: map[string]string{
				"app":  "middleware-nginx",
				"type": "testing",
			},
			outMatched: false,
		},
		{
			inSelector: "app=middleware-(nginx|redis)", // parse lex fail
			inLabels: map[string]string{
				"app":  "middleware-nginx",
				"type": "testing",
			},
			fail: true,
		},
		{
			inSelector: "app=middleware-{nginx,redis,etcd}",
			inLabels: map[string]string{
				"app":  "middleware-redis",
				"type": "testing",
			},
			fail: true,
		},
		{
			inSelector: "app=middleware-[nginx|redis]", // Glob does not support | for matching.
			inLabels: map[string]string{
				"app":  "middleware-nginx",
				"type": "testing",
			},
			outMatched: false,
		},
		{
			inSelector: "app=middleware-[abc]",
			inLabels: map[string]string{
				"app":  "middleware-a",
				"type": "testing",
			},
			outMatched: true,
		},
		{
			inSelector: "app!=middleware-[abc]",
			inLabels: map[string]string{
				"app":  "middleware-b",
				"type": "testing",
			},
			outMatched: false,
		},
		{
			inSelector: "app!=middleware*",
			inLabels: map[string]string{
				"app":  "middleware-nginx",
				"type": "testing",
			},
			outMatched: false,
		},
		{
			inSelector: "app in (middleware-nginx)",
			inLabels: map[string]string{
				"app":  "middleware-nginx",
				"type": "testing",
			},
			outMatched: true,
		},
		{
			inSelector: `name=\ middleware`,
			fail:       true,
		},
		{
			inSelector: "name in (\\[middleware)",
			fail:       true,
		},
	}

	for idx, tc := range testcases {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			selector, err := Parse(tc.inSelector)
			if tc.fail && assert.Error(t, err) {
				// t.Log(err)
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, tc.outMatched, selector.Matches(Set(tc.inLabels)))
		})
	}
}
