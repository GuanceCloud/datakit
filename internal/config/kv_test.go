// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"encoding/json"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func TestLoadKVs(t *testing.T) {
	f, err := os.MkdirTemp("./", "kv")
	assert.NoError(t, err)

	initialWorkDir := datakit.InstallDir

	t.Cleanup(func() {
		os.RemoveAll(f)
		datakit.SetupWorkDir(initialWorkDir)
	})

	datakit.SetWorkDir(f)

	testKV := map[string]interface{}{
		"version": 123456789,
		"value":   `{"token":"tkn_test_token"}`,
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
		assert.Equal(t, tc.expected, data)
	}
}

func TestKV(t *testing.T) {
	createKV := func() *KV {
		return &KV{
			watchers: map[string]*watcher{},
			kv: map[string]string{
				"name":  "name1",
				"name1": "name12",
			},
		}
	}

	t.Run("singal conf", func(t *testing.T) {
		kv := createKV()
		// register single conf watcher
		count := 0
		var wg sync.WaitGroup
		wg.Add(1)
		kv.Register("singal_conf", "{{.name}}", func(content map[string]string) error {
			defer wg.Done()
			count++ // count times +1
			assert.Equal(t, "name11", content["singal_conf"])
			return nil
		}, nil)

		// kv changed && reload
		kv.kv["name"] = "name11"
		kv.reload()

		wg.Wait()
		assert.Equal(t, 1, count)
	})

	t.Run("multi conf", func(t *testing.T) {
		// register multi conf watcher with conf name
		var wg sync.WaitGroup
		count := 0
		kv := createKV()
		wg.Add(1)
		err := kv.Register("multi_conf", "{{.name}}", func(content map[string]string) error {
			defer wg.Done()
			assert.Equal(t, "", content["multi_conf"])
			count++ // count times +1
			return nil
		}, &KVOpt{
			IsMultiConf: true,
			ConfName:    "name",
		})
		assert.NoError(t, err)

		// register multi conf watcher with conf name
		err = kv.Register("multi_conf", "{{.name1}}", nil, &KVOpt{
			IsMultiConf: true,
			ConfName:    "name1",
		})
		assert.NoError(t, err)

		// kv changed && reload
		kv.kv["name"] = "name11"
		kv.kv["name1"] = "name13"
		kv.reload()

		wg.Wait()
		assert.Equal(t, 1, count)
	})

	t.Run("multi conf auto unregister", func(t *testing.T) {
		// register multi conf watcher with conf name and auto unregister
		var wg sync.WaitGroup
		kv := createKV()
		called := false
		wg.Add(1)
		err := kv.Register("multi_conf_auto_unregister", "{{.name}}",
			func(content map[string]string) error {
				defer wg.Done()

				// auto unregister
				_, ok := kv.watchers["multi_conf_auto_unregister"]
				assert.False(t, ok)

				called = true
				return nil
			},
			&KVOpt{
				IsMultiConf:              true,
				IsUnRegisterBeforeReload: true,
				ConfName:                 "name",
			})
		assert.NoError(t, err)

		// kv changed && reload
		kv.kv["name"] = "name11"
		kv.reload()

		wg.Wait()
		assert.True(t, called)
	})
}
