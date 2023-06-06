// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_setHostTagIfNotLoopback(t *testing.T) {
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
			name: "loopback with param",
			args: args{
				tags: make(map[string]string),
				u:    "localhost:27017/?authMechanism=SCRAM-SHA-256&authSource=admin",
			},
			expected: map[string]string{},
		},
		{
			name: "loopback with param",
			args: args{
				tags: make(map[string]string),
				u:    "127.0.0.1:27017/?authMechanism=SCRAM-SHA-256&authSource=admin",
			},
			expected: map[string]string{},
		},
		{
			name: "loopback",
			args: args{
				tags: make(map[string]string),
				u:    "127.0.0.1:27017",
			},
			expected: map[string]string{},
		},
		{
			name: "normal",
			args: args{
				tags: make(map[string]string),
				u:    "10.10.3.33:18832",
			},
			expected: map[string]string{"host": "10.10.3.33"},
		},
		{
			name: "normal with param",
			args: args{
				tags: make(map[string]string),
				u:    "10.10.3.33:18832/?authMechanism=SCRAM-SHA-256&authSource=admin",
			},
			expected: map[string]string{"host": "10.10.3.33"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setHostTagIfNotLoopback(tt.args.tags, tt.args.u)
			assert.Equal(t, tt.expected, tt.args.tags)
		})
	}
}
