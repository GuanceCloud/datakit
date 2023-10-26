// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package dialtesting

import (
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
