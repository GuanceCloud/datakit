// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestLoadKVs(t *testing.T) {
	f, err := os.MkdirTemp("./", "kv")
	assert.NoError(t, err)
	defer os.RemoveAll(f)
	initialWorkDir := datakit.InstallDir

	defer func() {
		datakit.SetWorkDir(initialWorkDir)
	}()

	datakit.SetWorkDir(f)

	testKV := KV{
		Version: 123456789,
		Value:   `{"token":"tkn_test_token"}`,
	}

	v, err := json.Marshal(testKV)
	assert.NoError(t, err)

	err = os.WriteFile(datakit.KVFile, v, os.ModePerm)
	assert.NoError(t, err)

	defaultKV.LoadKV()

	assert.Equal(t, "tkn_test_token", defaultKV.kv["token"])
}

func TestReplaceKVs(t *testing.T) {
	cases := []struct {
		name     string
		kv       map[string]string
		template string
		expected string
	}{
		{
			name: "replace string",
			kv: map[string]string{
				"token": "tkn_test_token",
			},
			template: `token={{.token}}`,
			expected: "token=tkn_test_token",
		},
		{
			name:     "default value used",
			kv:       map[string]string{},
			template: `token={{.token|default "default_token"}}`,
			expected: "token=default_token",
		},
		{
			name: "default value not used",
			kv: map[string]string{
				"token": "tkn_test_token",
			},
			template: `token={{.token|default "default_token"}}`,
			expected: "token=tkn_test_token",
		},
	}

	for _, tc := range cases {
		kv := &KV{
			kv: tc.kv,
		}

		data, err := kv.ReplaceKV(tc.template)

		assert.NoError(t, err)
		assert.Equal(t, tc.expected, string(data))
	}
}
