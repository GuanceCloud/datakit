// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	T "testing"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"github.com/influxdata/toml"
	"github.com/stretchr/testify/assert"
)

func TestDuration(t *T.T) {
	t.Run("duration", func(t *T.T) {
		type Some struct {
			Duration  time.Duration   `toml:"duration"`
			Durations []time.Duration `toml:"durations"`
		}

		var x Some

		_, err := bstoml.Decode(string(`
		duration = "1m"
		durations = ["1m", "3h"]
`), &x)
		assert.NoError(t, err)

		assert.Equal(t, time.Minute, x.Duration)
		assert.Equal(t, time.Minute, x.Durations[0])
		assert.Equal(t, time.Hour*3, x.Durations[1])
	})
}

func TestTOMLParse(t *T.T) {
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
		t.Run(tc.name, func(t *T.T) {
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
