// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONTaskFile(t *testing.T) {
	cases := []struct {
		name, j string
		fail    bool
	}{
		{
			j: `
{
  "HTTP": [
    {
      "name": "baidu-json-test",
      "method": "GET",
      "url": "http://baidu.com",
      "post_url": "http://testing-openway.cloudcare.cn?token=tkn_878de73a7cb411ebb24c9a711bbe15d4",
      "status": "OK",
      "frequency": "10s",
      "region": "shang_hai",
      "owner_external_id": "ak_c1imts73q2c335d71cn0-wksp_878de24e7cb411ebb24c9a711bbe15d4",
      "success_when": [
        {
          "response_time": "1000ms",
          "header": {
            "Content-Type": [
              {
                "contains": "html"
              }
            ]
          },
          "status_code": [
            {
              "is": "200"
            }
          ]
        }
      ],
      "advance_options": {
        "request_options": {
          "auth": {}
        },
        "request_body": {},
        "secret": {}
      },
      "update_time": 1645065786362746
    }
  ]
}
`,
			name: `normal case`,
		},
	}

	for _, tc := range cases {
		_ = tc
		t.Run(tc.name, func(t *testing.T) {
			i := defaultInput()
			b, err := i.getLocalJSONTasks([]byte(tc.j))
			if tc.fail {
				assert.Error(t, err)
			}

			assert.NoErrorf(t, err, "get local task: %s", string(b))
		})
	}
}

func TestGetLocalJSONTasks(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		i := defaultInput()

		b, err := i.getLocalJSONTasks([]byte(`{invalid-json}`))
		assert.Error(t, err)
		assert.Nil(t, b)
	})

	t.Run("convert task objects to json strings", func(t *testing.T) {
		i := defaultInput()

		b, err := i.getLocalJSONTasks([]byte(`{
			"HTTP": [
				{
					"name": "test-http-task",
					"external_id": "http-task",
					"method": "GET",
					"url": "http://example.com",
					"post_url": "http://example.com?token=test",
					"status": "ok",
					"frequency": "10s",
					"region": "test-region",
					"owner_external_id": "owner"
				}
			]
		}`))
		if !assert.NoError(t, err) {
			return
		}

		var resp taskPullResp
		if !assert.NoError(t, json.Unmarshal(b, &resp)) {
			return
		}

		items, ok := resp.Content["HTTP"].([]interface{})
		if !assert.True(t, ok) {
			return
		}
		if !assert.Len(t, items, 1) {
			return
		}

		taskString, ok := items[0].(string)
		if !assert.True(t, ok) {
			return
		}

		var task map[string]interface{}
		if !assert.NoError(t, json.Unmarshal([]byte(taskString), &task)) {
			return
		}

		assert.Equal(t, "test-http-task", task["name"])
		assert.Equal(t, "http-task", task["external_id"])
		assert.Equal(t, "http://example.com", task["url"])
	})

	t.Run("empty task list keeps empty content slice", func(t *testing.T) {
		i := defaultInput()

		b, err := i.getLocalJSONTasks([]byte(`{
			"HTTP": []
		}`))
		if !assert.NoError(t, err) {
			return
		}

		var resp taskPullResp
		if !assert.NoError(t, json.Unmarshal(b, &resp)) {
			return
		}

		items, ok := resp.Content["HTTP"].([]interface{})
		if !assert.True(t, ok) {
			return
		}

		assert.Empty(t, items)
	})

	t.Run("multiple task classes are converted independently", func(t *testing.T) {
		i := defaultInput()

		b, err := i.getLocalJSONTasks([]byte(`{
			"HTTP": [
				{
					"name": "http-task",
					"external_id": "http-task",
					"method": "GET",
					"url": "http://example.com"
				}
			],
			"TCP": [
				{
					"name": "tcp-task",
					"external_id": "tcp-task",
					"host": "example.com",
					"port": "80"
				}
			]
		}`))
		if !assert.NoError(t, err) {
			return
		}

		var resp taskPullResp
		if !assert.NoError(t, json.Unmarshal(b, &resp)) {
			return
		}

		httpItems, ok := resp.Content["HTTP"].([]interface{})
		if !assert.True(t, ok) {
			return
		}
		tcpItems, ok := resp.Content["TCP"].([]interface{})
		if !assert.True(t, ok) {
			return
		}

		assert.Len(t, httpItems, 1)
		assert.Len(t, tcpItems, 1)
	})
}
