package beats_output

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v -timeout 30s -run ^TestParseListen$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/beats_output
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

// go test -v -timeout 30s -run ^TestEventGet$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/beats_output
func TestEventGet(t *testing.T) {
	const testEventPiece = `
{
	"@metadata": {
	  "beat": "filebeat",
	  "type": "_doc",
	  "version": "7.13.3"
	},
	"@timestamp": "2022-04-26T14:05:23.283Z",
	"agent": {
	  "ephemeral_id": "62dfa2ab-f165-444e-a97d-1cf82ec69fa6",
	  "hostname": "MacBook-Air-2.local",
	  "id": "24627faf-48e9-4253-8c5f-d64657ee6e4e",
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

	// json to map
	var mapResult map[string]interface{}
	err := json.Unmarshal([]byte(testEventPiece), &mapResult)
	assert.NoError(t, err)

	cases := []struct {
		name   string
		path   string
		expect string
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
	}

	for _, tc := range cases {
		// start call eventGet
		t.Run(tc.name, func(t *testing.T) {
			val := eventGet(mapResult, tc.path)
			assert.Equal(t, tc.expect, val)
		})
	}
}
