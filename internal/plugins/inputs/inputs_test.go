// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package inputs

import (
	"fmt"
	"os"
	"testing"
	T "testing"

	tu "github.com/GuanceCloud/cliutils/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dkpt "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/point"
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

type taggerMock struct {
	hostTags, electionTags map[string]string
}

func (m *taggerMock) HostTags() map[string]string {
	return m.hostTags
}

func (m *taggerMock) ElectionTags() map[string]string {
	return m.electionTags
}

type ForTestInput struct {
	Election bool
	Servers  []string
	Tagger   dkpt.GlobalTagger
	Tags     map[string]string

	globalTags map[string]map[string]string // server:map[string]string
}

// go test -v -timeout 30s -run ^Test_InitGlobalTags$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/
func Test_InitGlobalTags(t *testing.T) {
	cases := []struct {
		name     string
		election bool
		servers  []string
		tagger   *taggerMock
		tags     map[string]string
		expect   map[string]map[string]string
	}{
		{
			name:     "election",
			election: true,
			servers:  []string{"1.2.3.4", "1.2.3.4:80"},
			tagger: &taggerMock{
				hostTags: map[string]string{
					"host":  "foo",
					"hello": "world",
				},

				electionTags: map[string]string{
					"project": "foo",
					"cluster": "bar",
				},
			},
			tags: map[string]string{
				"apple":       "orange",
				"information": "mark",
				"see":         "you",
			},
			expect: map[string]map[string]string{
				"1.2.3.4": {
					"project":     "foo",
					"cluster":     "bar",
					"apple":       "orange",
					"information": "mark",
					"see":         "you",
					"host":        "1.2.3.4",
				},
				"1.2.3.4:80": {
					"project":     "foo",
					"cluster":     "bar",
					"apple":       "orange",
					"information": "mark",
					"see":         "you",
					"host":        "1.2.3.4",
				},
			},
		},

		{
			name:     "not_election",
			election: false,
			servers:  []string{"1.2.3.4", "1.2.3.4:80"},
			tagger: &taggerMock{
				hostTags: map[string]string{
					"hallo": "ja",
					"hello": "world",
				},

				electionTags: map[string]string{
					"project": "foo",
					"cluster": "bar",
				},
			},
			tags: map[string]string{
				"apple":       "orange",
				"information": "mark",
				"see":         "you",
			},
			expect: map[string]map[string]string{
				"1.2.3.4": {
					"hallo":       "ja",
					"hello":       "world",
					"apple":       "orange",
					"information": "mark",
					"see":         "you",
					"host":        "1.2.3.4",
				},
				"1.2.3.4:80": {
					"hallo":       "ja",
					"hello":       "world",
					"apple":       "orange",
					"information": "mark",
					"see":         "you",
					"host":        "1.2.3.4",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ipt := &ForTestInput{
				Election: tc.election,
				Servers:  tc.servers,
				Tagger:   tc.tagger,
				Tags:     tc.tags,
			}

			ipt.globalTags = InitGlobalTags(
				ipt.Servers,
				ipt.Election,
				ipt.Tagger,
				ipt.Tags,
			)

			require.Equal(t, tc.expect, ipt.globalTags)
		})
	}
}

// go test -v -timeout 30s -run ^Test_mergeGlobalTags$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/
func Test_MergeGlobalTags(t *testing.T) {
	cases := []struct {
		name       string
		globalTags map[string]map[string]string
		in         map[string]string
		remote     string
		expect     map[string]string
	}{
		{
			name: "normal",
			globalTags: map[string]map[string]string{
				"1.2.3.4": {
					"key1": "val1",
					"key2": "val2",
				},
				"1.2.3.4:80": {
					"key3": "val3",
					"key4": "val4",
				},
			},
			in: map[string]string{
				"key5": "val5",
			},
			remote: "1.2.3.4:80",
			expect: map[string]string{
				"key3": "val3",
				"key4": "val4",
				"key5": "val5",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			MergeGlobalTags(
				tc.in,
				tc.globalTags,
				tc.remote,
			)

			require.Equal(t, tc.expect, tc.in)
		})
	}
}

////////////////////////////////////////////////////////////////////////////////
