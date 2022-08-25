// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dkstring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v -timeout 30s -run ^TestGetMapMD5String$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring
func TestGetMapMD5String(t *testing.T) {
	cases := []struct {
		name         string
		inM          map[string]interface{}
		inIgnoreKeys []string
		expectMD5    string
		expectOrigin string
		expectErr    error
	}{
		{
			name: "normal",
			inM: map[string]interface{}{
				"a": "aaa",
				"b": 123,
				"c": "cccc",
				"d": "ddd",
			},
			expectMD5:    "95723a74a9433ae515befb78c4065a00",
			expectOrigin: "aaaccccddd",
		},
		{
			name: "array",
			inM: map[string]interface{}{
				"a": "abc",
				"b": []int{123, 456},
				"d": 123,
				"c": []string{"apple", "orange", "banana"},
				"e": 4567,
			},
			expectMD5:    "18912494e40ef0ac9876054667ed0578",
			expectOrigin: "abcappleorangebanana",
		},
		{
			name: "ignore",
			inM: map[string]interface{}{
				"a": "aaa",
				"b": 123,
				"c": "cccc",
				"d": "ddd",
				"e": "apple",
				"f": "banana",
			},
			inIgnoreKeys: []string{"e"},
			expectMD5:    "a18451aad537381eb069dfc1c3391669",
			expectOrigin: "aaaccccdddbanana",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			md5, origin, err := GetMapMD5String(tc.inM, tc.inIgnoreKeys)
			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectOrigin, origin)
			assert.Equal(t, tc.expectMD5, md5)
		})
	}
}

// go test -v -timeout 30s -run ^TestGetStringFromInterface$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring
func TestGetStringFromInterface(t *testing.T) {
	cases := []struct {
		name string
		in   interface{}
		out  string
	}{
		{
			name: "one",
			in:   "abc",
			out:  "abc",
		},
		{
			name: "array",
			in:   []string{"abc", "bcd"},
			out:  "abcbcd",
		},
		{
			name: "not string type",
			in:   123,
			out:  "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := getStringFromInterface(tc.in)
			assert.Equal(t, tc.out, out)
		})
	}
}
