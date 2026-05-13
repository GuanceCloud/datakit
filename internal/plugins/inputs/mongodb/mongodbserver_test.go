// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMongoServerHost(t *testing.T) {
	tests := []struct {
		name    string
		server  string
		host    string
		wantErr bool
	}{
		{
			name:   "mongodb uri",
			server: "mongodb://127.0.0.1:27017/admin?authSource=admin",
			host:   "127.0.0.1:27017",
		},
		{
			name:   "mongodb uri with credentials",
			server: "mongodb://user:pass@127.0.0.1:27017/admin?authSource=admin",
			host:   "127.0.0.1:27017",
		},
		{
			name:   "mongodb srv uri",
			server: "mongodb+srv://cluster0.example.com/admin?retryWrites=true&w=majority",
			host:   "cluster0.example.com",
		},
		{
			name:    "invalid uri without host",
			server:  "mongodb://",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, err := parseMongoServerHost(tt.server)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.host, host)
		})
	}
}

func TestSplitMongoAddr(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		host    string
		port    string
		wantErr bool
	}{
		{
			name: "host and port",
			addr: "10.10.3.33:18832",
			host: "10.10.3.33",
			port: "18832",
		},
		{
			name: "host and port with params",
			addr: "10.10.3.33:18832/?authMechanism=SCRAM-SHA-256&authSource=admin",
			host: "10.10.3.33",
			port: "18832",
		},
		{
			name: "srv host only",
			addr: "cluster0.example.com",
			host: "cluster0.example.com",
			port: "",
		},
		{
			name:    "invalid",
			addr:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port, err := splitMongoAddr(tt.addr)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.host, host)
			assert.Equal(t, tt.port, port)
		})
	}
}

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
		{
			name: "srv host only",
			args: args{
				tags: make(map[string]string),
				u:    "cluster0.example.com",
			},
			expected: map[string]string{"host": "cluster0.example.com"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setHostTagIfNotLoopback(tt.args.tags, tt.args.u)
			assert.Equal(t, tt.expected, tt.args.tags)
		})
	}
}

func TestMongodbServerGetDefaultTagsAddsServerAndDatabaseInstance(t *testing.T) {
	oldDefTags := defTags
	defer func() {
		defTags = oldDefTags
	}()

	defTags = map[string]string{"env": "testing"}
	svr := &MongodbServer{
		host:             "127.0.0.1:27017",
		databaseInstance: "mongo.example.com:27017",
	}

	tags := svr.getDefaultTags()

	assert.Equal(t, "127.0.0.1:27017", tags["server"])
	assert.Equal(t, "mongo.example.com:27017", tags["database_instance"])
	assert.Equal(t, "127.0.0.1:27017", tags["mongod_host"])
	assert.Equal(t, "testing", tags["env"])
	assert.NotContains(t, tags, "host")
}

func TestMongodbServerGetDefaultTagsFallbackAndOverride(t *testing.T) {
	oldDefTags := defTags
	defer func() {
		defTags = oldDefTags
	}()

	defTags = nil
	svr := &MongodbServer{
		host:             "10.10.3.33:18832",
		databaseInstance: "10.10.3.33:18832",
	}

	tags := svr.getDefaultTags()

	assert.Equal(t, "10.10.3.33:18832", tags["server"])
	assert.Equal(t, "10.10.3.33:18832", tags["database_instance"])
	assert.Equal(t, "10.10.3.33", tags["host"])

	defTags = map[string]string{
		"database_instance": "configured-instance",
		"server":            "configured-server",
	}
	tags = svr.getDefaultTags()

	assert.Equal(t, "configured-server", tags["server"])
	assert.Equal(t, "configured-instance", tags["database_instance"])
}
