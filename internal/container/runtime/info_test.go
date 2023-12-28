// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDockerEnvs(t *testing.T) {
	cases := []struct {
		in  []string
		out map[string]string
	}{
		{
			in:  []string{"PATH=/usr/local/sbin:/usr/local/bin"},
			out: map[string]string{"PATH": "/usr/local/sbin:/usr/local/bin"},
		},
		{
			// not '='
			in:  []string{"PATH-ABC"},
			out: map[string]string{"PATH-ABC": ""},
		},
		{
			// doubel '='
			in:  []string{"PATH=/usr/local/sbin:=/usr/local/bin"},
			out: map[string]string{"PATH": "/usr/local/sbin:/usr/local/bin"},
		},
	}

	for _, tc := range cases {
		res := parseDockerEnv(tc.in)
		assert.Equal(t, tc.out, res)
	}
}

func TestParseCriInfo(t *testing.T) {
	in := `
{
  "pid": 32166,
  "config": {
    "working_dir": "/usr/local/datakit",
    "envs": [
      {
        "key": "ENV_NAME",
        "value": "abc"
      },
      {
        "key": "ENV_HTTP_PORT",
        "value": "9529"
      }
    ]
  }
}
`
	out := &criInfo{
		Pid: 32166,
	}
	out.Config.Envs = []env{
		{"ENV_NAME", "abc"},
		{"ENV_HTTP_PORT", "9529"},
	}

	t.Run("parse-info", func(t *testing.T) {
		res, err := ParseCriInfo(in)
		assert.NoError(t, err)

		assert.Equal(t, out, res)
	})

	t.Run("parse-envs", func(t *testing.T) {
		res := out.getConfigEnvs()

		assert.Equal(t, 2, len(res))
		assert.Equal(t, "abc", res["ENV_NAME"])
		assert.Equal(t, "9529", res["ENV_HTTP_PORT"])
	})
}
