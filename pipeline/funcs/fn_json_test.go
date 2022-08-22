// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"
	"time"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestJSON(t *testing.T) {
	testCase := []*funcCase{
		{
			in: `{
			  "name": {"first": "Tom", "last": "Anderson"},
			  "age":37,
			  "children": ["Sara","Alex","Jack"],
			  "fav.movie": "Deer Hunter",
			  "friends": [
			    {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
			    {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
			    {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
			  ]
			}`,
			script:   `json(_, name) json(name, first)`,
			expected: "Tom",
			key:      "first",
		},
		{
			in: `{
			  "name": {"first": "Tom", "last": "Anderson"},
			  "age":37,
			  "children": ["Sara","Alex","Jack"],
			  "fav.movie": "Deer Hunter",
			  "friends": [
			    {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
			    {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
			    {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
			  ]
			}`,
			script:   `json(_, friends) json(friends, .[1].first, f_first)`,
			expected: "Roger",
			key:      "f_first",
		},
		{
			in: `[
				    {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
				    {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
				    {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
				]`,
			script:   `json(_, .[0].nets[-1])`,
			expected: "tw",
			key:      "[0].nets[-1]",
		},
		{
			in: `[
				    {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
				    {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
				    {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
				]`,
			script:   `json(_, .[1].age)`,
			expected: float64(68),
			key:      "[1].age",
		},
		{
			name:     "trim_space auto",
			in:       `{"item": " not_space "}`,
			script:   `json(_, item, item)`,
			key:      "item",
			expected: "not_space",
		},
		{
			name:     "trim_space disable",
			in:       `{"item": " not_space "}`,
			script:   `json(_, item, item, false)`,
			key:      "item",
			expected: " not_space ",
		},
		{
			name:     "trim_space enable",
			in:       `{"item": " not_space "}`,
			script:   `json(_, item, item, true)`,
			key:      "item",
			expected: "not_space",
		},
	}

	for idx, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			runner, err := NewTestingRunner(tc.script)
			tu.Equals(t, nil, err)
			ret, err := runner.Run("test", map[string]string{},
				map[string]interface{}{
					"message": tc.in,
				}, "message", time.Now())
			tu.Equals(t, nil, err)
			tu.Equals(t, nil, ret.Error)

			r, ok := ret.Fields[tc.key]
			tu.Equals(t, true, ok)
			if tc.key == "[2].age" {
				t.Log(1)
			}
			tu.Equals(t, tc.expected, r)

			t.Logf("[%d] PASS", idx)
		})
	}
}
