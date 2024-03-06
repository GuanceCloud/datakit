//go:build linux
// +build linux

package sysmonitor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBytes(t *testing.T) {
	cases := [][2]any{
		{"2MB", float64(2 * 1000 * 1000)},
		{"2MiB", float64(2 * 1024 * 1024)},
		{"2GB", float64(2 * 1000 * 1000 * 1000)},
		{"2GiB", float64(2 * 1024 * 1024 * 1024)},
	}

	for _, v := range cases {
		t.Run(v[0].(string), func(t *testing.T) {
			r := GetBytes(v[0].(string))
			assert.Equal(t, v[1], r)
		})
	}
}
