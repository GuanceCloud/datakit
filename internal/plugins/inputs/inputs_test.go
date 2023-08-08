// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package inputs

import (
	"fmt"
	"os"
	T "testing"

	tu "github.com/GuanceCloud/cliutils/testutil"
	"github.com/stretchr/testify/assert"
)

func TestGetEnvs(t *T.T) {
	cases := []struct {
		envs   map[string]string
		expect map[string]string
	}{
		{
			envs:   map[string]string{"ENV_A": "a", "DK_B": "b", "ABC_X": "x"},
			expect: map[string]string{"ENV_A": "a", "DK_B": "b"},
		},

		{
			envs:   map[string]string{"ENV_A": "===a", "DK_B": "b", "ABC_X": "x"},
			expect: map[string]string{"ENV_A": "===a", "DK_B": "b"},
		},

		{
			envs:   map[string]string{"ENV_A": "===a", "DK_B": "-=-=-=a:b", "ABC_X": "x"},
			expect: map[string]string{"ENV_A": "===a", "DK_B": "-=-=-=a:b"},
		},

		{
			envs:   map[string]string{"ENV_": "===a", "DK_": "-=-=-=a:b", "ABC_X": "x"},
			expect: map[string]string{"ENV_": "===a", "DK_": "-=-=-=a:b"},
		},
	}

	for _, tc := range cases {
		for k, v := range tc.envs {
			if err := os.Setenv(k, v); err != nil {
				t.Error(err)
			}
		}

		envs := getEnvs()
		for k, v := range tc.expect {
			val, ok := envs[k]
			tu.Assert(t, val == v && ok, fmt.Sprintf("%s <> %s && %v", v, val, ok))
		}
	}
}

func TestMergeTags(t *T.T) {
	t.Run("empty-remote", func(t *T.T) {
		orig := map[string]string{}
		after := MergeTags(nil, orig, "")
		assert.NotContains(t, after, "host")
	})

	t.Run("basic", func(t *T.T) {
		orig := map[string]string{}
		after := MergeTags(nil, orig, "http://1.2.3.4:1234")
		assert.Equal(t, after["host"], "1.2.3.4")
	})

	t.Run("localhost", func(t *T.T) {
		orig := map[string]string{}
		after := MergeTags(nil, orig, "http://localhost:1234")
		assert.NotContains(t, after, "host")
	})

	t.Run("other-scheme", func(t *T.T) {
		orig := map[string]string{}
		after := MergeTags(nil, orig, "mysql://1.2.3.4:1234")
		assert.Equal(t, after["host"], "1.2.3.4")
	})

	t.Run("no-port", func(t *T.T) {
		orig := map[string]string{}
		after := MergeTags(nil, orig, "http://1.2.3.4")
		assert.Equal(t, after["host"], "1.2.3.4")
	})

	t.Run("no-scheme", func(t *T.T) {
		orig := map[string]string{}
		after := MergeTags(nil, orig, "1.2.3.4:1234")
		assert.Equal(t, after["host"], "1.2.3.4")
	})

	t.Run("no-scheme-no-port", func(t *T.T) {
		orig := map[string]string{}
		after := MergeTags(nil, orig, "1.2.3.4")
		assert.Equal(t, after["host"], "1.2.3.4")
	})

	t.Run("bad-remote", func(t *T.T) {
		orig := map[string]string{}
		after := MergeTags(nil, orig, "this-is-a-bad-URL")
		assert.Equal(t, after["host"], "this-is-a-bad-URL") // Yes, we accept bad URL as `host` tag
	})

	t.Run("domain", func(t *T.T) {
		orig := map[string]string{}
		after := MergeTags(nil, orig, "http://some-service/metrics")
		assert.Equal(t, after["host"], "some-service")
	})

	t.Run("domain-with-port", func(t *T.T) {
		orig := map[string]string{}
		after := MergeTags(nil, orig, "http://some-service:12345/metrics")
		assert.Equal(t, after["host"], "some-service")
	})

	t.Run("domain-with-invalid-port", func(t *T.T) {
		orig := map[string]string{}
		after := MergeTags(nil, orig, "http://some-service:abcde/metrics")
		assert.Equal(t, after["host"], "http://some-service:abcde/metrics")
	})

	t.Run("domain-with-invalid-port-2", func(t *T.T) {
		orig := map[string]string{}

		// port number beyond 64k URL parse OK!
		after := MergeTags(nil, orig, "http://some-service:66666/metrics")

		assert.Equal(t, after["host"], "some-service")
	})

	t.Run("with-user-password", func(t *T.T) {
		orig := map[string]string{}

		// port number beyond 64k URL parse OK!
		after := MergeTags(nil, orig, "mongodb://user:pswd@1.2.3.4:27017/?authMechanism=SCRAM-SHA-256&authSource=admin")

		assert.Equal(t, after["host"], "1.2.3.4")
	})

	t.Run("127.0.0.1", func(t *T.T) {
		orig := map[string]string{}

		// port number beyond 64k URL parse OK!
		after := MergeTags(nil, orig, "mongodb://user:pswd@127.0.0.1:27017/?authMechanism=SCRAM-SHA-256&authSource=admin")

		assert.NotContains(t, after, "host")
	})
}

////////////////////////////////////////////////////////////////////////////////
