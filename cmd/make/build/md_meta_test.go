// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package build

import (
	"io/ioutil"
	"os"
	"path/filepath"
	T "testing"

	"github.com/stretchr/testify/assert"
)

func Test_checkMarkdownMeta(t *T.T) {
	t.Run("basic", func(t *T.T) {
		temp := "basic"
		assert.NoError(t, os.MkdirAll(temp, os.ModePerm))

		defer t.Cleanup(func() {
			os.RemoveAll(temp)
		})

		// create meta dirs
		for _, f := range []string{
			"icon/icon.png",
			"icon/icon-dark.png",
			"monitor/meta.json",
			"dashboard/meta.json",
		} {
			assert.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(temp, f)), os.ModePerm))
			t.Logf("create %s...", filepath.Join(temp, f))
			os.Create(filepath.Join(temp, f))
			// TODO: invalid json?
			if filepath.Ext(filepath.Join(temp, f)) == ".json" {
				assert.NoError(t, ioutil.WriteFile(filepath.Join(temp, f), []byte("{}"), os.ModePerm))
			}
		}

		md := `---
title: 'title string'
summary: 'summary string'
__int_icon: 'icon'
dashboard:
  - desc: 'dashboard desc'
    path: 'dashboard'
  - desc: 'dashboard desc'
    path: 'dashboard'
  - desc: 'dashboard desc'
    path: 'dashboard'
  - desc: 'dashboard desc'
    path: 'dashboard'
monitor:
  - desc: 'monitor desc'
    path: 'monitor'
  - desc: 'monitor desc'
    path: 'monitor'
  - desc: 'monitor desc'
    path: 'monitor'
---`

		assert.NoError(t, checkMarkdownMeta([]byte(md), temp))
	})
}
