// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package beats_output

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v -timeout 30s -run ^TestParseListen$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/beats_output
func TestParseListen(t *testing.T) {
	cases := []struct {
		name        string
		listen      string
		expectError error
		expect      map[string]string
	}{
		{
			name:   "normal",
			listen: "tcp://0.0.0.0:5044",
			expect: map[string]string{
				"scheme": "tcp",
				"host":   "0.0.0.0:5044",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mVal, err := parseListen(tc.listen)
			assert.Equal(t, tc.expectError, err)
			assert.Equal(t, tc.expect, mVal)
		})
	}
}

// go test -v -timeout 30s -run ^TestEventGet$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/beats_output
func TestEventGet(t *testing.T) {
	const testEventPiece = `
{
  "@metadata": {
    "beat": "filebeat",
    "type": "_doc",
    "version": "7.13.3"
  },
  "@timestamp": "2022-04-29T08:28:05.060Z",
  "agent": {
    "ephemeral_id": "c11e357a-f28f-439f-8981-868db91c72ff",
    "hostname": "MacBook-Air-2.local",
    "id": "c12dca3e-5add-4cd0-9890-2a721c867ab0",
    "name": "MacBook-Air-2.local",
    "type": "filebeat",
    "version": "7.13.3"
  },
  "ecs": {
    "version": "1.8.0"
  },
  "fields": {
    "logtype": "sshd-log",
    "product": "beijing",
    "type": "sshd-log"
  },
  "host": {
    "name": "MacBook-Air-2.local"
  },
  "input": {
    "type": "filestream"
  },
  "log": {
    "file": {
      "path": "/Users/mac/Downloads/tmp/1.log"
    },
    "offset": 12
  },
  "message": "hello world"
}
`

	// json to map
	var mapResult map[string]interface{}
	err := json.Unmarshal([]byte(testEventPiece), &mapResult)
	assert.NoError(t, err)

	cases := []struct {
		name   string
		path   string
		expect interface{}
	}{
		{
			name:   "host.name",
			path:   "host.name",
			expect: "MacBook-Air-2.local",
		},
		{
			name:   "log.file.path",
			path:   "log.file.path",
			expect: "/Users/mac/Downloads/tmp/1.log",
		},
		{
			name:   "message",
			path:   "message",
			expect: "hello world",
		},
		{
			name:   "fields",
			path:   "fields",
			expect: map[string]interface{}{"logtype": "sshd-log", "product": "beijing", "type": "sshd-log"},
		},
	}

	for _, tc := range cases {
		// start call eventGet
		t.Run(tc.name, func(t *testing.T) {
			val := eventGet(mapResult, tc.path)
			assert.Equal(t, tc.expect, val)
		})
	}
}

// go test -v -timeout 30s -run ^TestGetNewTags$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/beats_output
func TestGetNewTags(t *testing.T) {
	cases := []struct {
		name   string
		ipt    *Input
		in     *DataStruct
		expect map[string]string
	}{
		{
			name: "normal",
			ipt: &Input{
				Service: "test-service",
				Tags: map[string]string{
					"tag1": "val1",
					"tag2": "val2",
				},
			},
			in: &DataStruct{
				HostName:    "MacBook-Air-2.local",
				LogFilePath: "/Users/mac/Downloads/tmp/1.log",
				Message:     "hello world",
				Fields: map[string]interface{}{
					"logtype": "sshd-log",
					"product": "beijing",
					"type":    "sshd-log",
					"int":     123,
					"int64":   int64(456),
					"int32":   int32(789),
					"float":   1.0,
				},
			},
			expect: map[string]string{
				"service":  "test-service",
				"tag1":     "val1",
				"tag2":     "val2",
				"host":     "MacBook-Air-2.local",
				"filepath": "/Users/mac/Downloads/tmp/1.log",
				"logtype":  "sshd-log",
				"product":  "beijing",
				"type":     "sshd-log",
				"int":      "123",
				"int64":    "456",
				"int32":    "789",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mVal := tc.ipt.getNewTags(tc.in)
			assert.Equal(t, tc.expect, mVal)
		})
	}
}

// go test -v -timeout 30s -run ^TestGetDataPieceFromEvent$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/beats_output
func TestGetDataPieceFromEvent(t *testing.T) {
	const normalPiece = `
{
  "@metadata": {
    "beat": "filebeat",
    "type": "_doc",
    "version": "7.13.3"
  },
  "@timestamp": "2022-04-29T08:28:05.060Z",
  "agent": {
    "ephemeral_id": "c11e357a-f28f-439f-8981-868db91c72ff",
    "hostname": "MacBook-Air-2.local",
    "id": "c12dca3e-5add-4cd0-9890-2a721c867ab0",
    "name": "MacBook-Air-2.local",
    "type": "filebeat",
    "version": "7.13.3"
  },
  "ecs": {
    "version": "1.8.0"
  },
  "fields": {
    "logtype": "sshd-log",
    "product": "beijing",
    "type": "sshd-log"
  },
  "host": {
    "name": "MacBook-Air-2.local"
  },
  "input": {
    "type": "filestream"
  },
  "log": {
    "file": {
      "path": "/Users/mac/Downloads/tmp/1.log"
    },
    "offset": 12
  },
  "message": "hello world"
}
`

	const withoutHostNamePiece = `
{
  "@metadata": {
    "beat": "filebeat",
    "type": "_doc",
    "version": "7.13.3"
  },
  "@timestamp": "2022-04-29T08:28:05.060Z",
  "agent": {
    "ephemeral_id": "c11e357a-f28f-439f-8981-868db91c72ff",
    "hostname": "MacBook-Air-2.local",
    "id": "c12dca3e-5add-4cd0-9890-2a721c867ab0",
    "name": "MacBook-Air-2.local",
    "type": "filebeat",
    "version": "7.13.3"
  },
  "ecs": {
    "version": "1.8.0"
  },
  "fields": {
    "logtype": "sshd-log",
    "product": "beijing",
    "type": "sshd-log"
  },
  "host": {},
  "input": {
    "type": "filestream"
  },
  "log": {
    "file": {
      "path": "/Users/mac/Downloads/tmp/1.log"
    },
    "offset": 12
  },
  "message": "hello world"
}
`

	const withoutFieldsPiece = `
{
  "@metadata": {
    "beat": "filebeat",
    "type": "_doc",
    "version": "7.13.3"
  },
  "@timestamp": "2022-04-29T08:28:05.060Z",
  "agent": {
    "ephemeral_id": "c11e357a-f28f-439f-8981-868db91c72ff",
    "hostname": "MacBook-Air-2.local",
    "id": "c12dca3e-5add-4cd0-9890-2a721c867ab0",
    "name": "MacBook-Air-2.local",
    "type": "filebeat",
    "version": "7.13.3"
  },
  "ecs": {
    "version": "1.8.0"
  },
  "host": {
    "name": "MacBook-Air-2.local"
  },
  "input": {
    "type": "filestream"
  },
  "log": {
    "file": {
      "path": "/Users/mac/Downloads/tmp/1.log"
    },
    "offset": 12
  },
  "message": "hello world"
}
`

	cases := []struct {
		name        string
		eventString string
		expect      *DataStruct
	}{
		{
			name:        "normal",
			eventString: normalPiece,
			expect: &DataStruct{
				HostName:    "MacBook-Air-2.local",
				LogFilePath: "/Users/mac/Downloads/tmp/1.log",
				Message:     "hello world",
				Fields: map[string]interface{}{
					"logtype": "sshd-log",
					"product": "beijing",
					"type":    "sshd-log",
				},
			},
		},
		{
			name:        "without_hostname",
			eventString: withoutHostNamePiece,
			expect: &DataStruct{
				LogFilePath: "/Users/mac/Downloads/tmp/1.log",
				Message:     "hello world",
				Fields: map[string]interface{}{
					"logtype": "sshd-log",
					"product": "beijing",
					"type":    "sshd-log",
				},
			},
		},
		{
			name:        "without_fields",
			eventString: withoutFieldsPiece,
			expect: &DataStruct{
				HostName:    "MacBook-Air-2.local",
				LogFilePath: "/Users/mac/Downloads/tmp/1.log",
				Message:     "hello world",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// json to map
			var mapResult map[string]interface{}
			err := json.Unmarshal([]byte(tc.eventString), &mapResult)
			assert.NoError(t, err)

			val := getDataPieceFromEvent(mapResult)
			assert.Equal(t, tc.expect, val)
		})
	}
}
