// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"testing"

	bstoml "github.com/BurntSushi/toml"
	"github.com/influxdata/toml"
)

func TestTOMLParse(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{
			name: `mixed-array-type`,
			data: []byte(
				`a = [
				"xyz",
				123
				]`,
			),
		},

		{
			name: `multiple-comments-in-string-arry`,
			data: []byte(
				`a = [
				# abc
				# def
				"xyz"
				]`,
			),
		},

		{
			name: `multiple-comments-in-int-arry`,
			data: []byte(
				`a = [
				"xyz"
				# abc
				# def
				]`,
			),
		},

		{
			name: `single-comment-in-arry`,
			data: []byte(
				`a = [
				  "xyz" # abc # def
				]`,
			),
		},

		{
			name: `all-commented-in-arry`,
			data: []byte(
				`a = [
					# "xyz" # abc # def
				]`,
			),
		},

		{
			name: `empty toml`,
			data: []byte(``),
		},

		{
			name: `conflict type`,
			data: []byte(`
				a = 10
				[a]
					bc = 10
			`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err1 := toml.Parse(tc.data)
			if err1 != nil {
				t.Logf("influx TOML parse: %s", err1)
			} else {
				t.Logf("influx TOML Parse ok")
			}

			var res interface{}
			_, err2 := bstoml.Decode(string(tc.data), &res)
			if err2 != nil {
				t.Logf("bstoml Parse: %s", err2)
			} else {
				t.Logf("bstoml Parse ok")
			}
		})
	}
}
