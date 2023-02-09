// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package prom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_setHostTag(t *testing.T) {
	type args struct {
		tags map[string]string
		u    string
	}
	tests := []struct {
		name     string
		args     args
		expected map[string]string
	}{
		{
			name:     "loopback address",
			args:     args{make(map[string]string), "http://127.0.0.1:9100/metrics"},
			expected: map[string]string{},
		},
		{
			name:     "loopback address",
			args:     args{make(map[string]string), "http://localhost:1234/metrics"},
			expected: map[string]string{},
		},
		{
			name:     "normal",
			args:     args{make(map[string]string), "http://224.135.2.10:1234/metrics"},
			expected: map[string]string{"host": "224.135.2.10"},
		},
		{
			name:     "illegal url",
			args:     args{make(map[string]string), "/usr/local"},
			expected: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setHostTagIfNotLoopback(tt.args.tags, tt.args.u)
			assert.Equal(t, tt.expected, tt.args.tags)
		})
	}
}
