// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"os"
	"path/filepath"
	T "testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type mockedInput struct{}

/*

	Catalog() string
	Run()
	SampleConfig() string
*/

func (*mockedInput) Catalog() string      { return "samples" }
func (*mockedInput) Run()                 {}
func (*mockedInput) SampleConfig() string { return "test-sample" }

func Test_initDefaultEnabledPlugins(t *T.T) {
	t.Run(`basic`, func(t *T.T) {
		c := DefaultConfig()
		c.DefaultEnabledInputs = []string{"cpu", "mem"}

		dir := t.TempDir()

		ipts := map[string]inputs.Creator{
			"cpu": func() inputs.Input { return &mockedInput{} },
			"mem": func() inputs.Input { return &mockedInput{} },
		}

		c.initDefaultEnabledPlugins(dir, ipts)

		assert.FileExists(t, filepath.Join(dir, "samples", "cpu.conf"))
		assert.FileExists(t, filepath.Join(dir, "samples", "mem.conf"))
	})

	t.Run(`conf-exist-as-dir`, func(t *T.T) {
		c := DefaultConfig()
		c.DefaultEnabledInputs = []string{"cpu", "mem"}

		dir := t.TempDir()

		assert.NoError(t, os.MkdirAll(filepath.Join(dir, "samples/cpu.conf"), os.ModePerm))

		ipts := map[string]inputs.Creator{
			"cpu": func() inputs.Input { return &mockedInput{} },
			"mem": func() inputs.Input { return &mockedInput{} },
		}

		c.initDefaultEnabledPlugins(dir, ipts)

		assert.FileExists(t, filepath.Join(dir, "samples", "cpu-0xdeadbeaf.conf"))
		assert.FileExists(t, filepath.Join(dir, "samples", "mem.conf"))
	})

	t.Run(`conf-exist-and-skip`, func(t *T.T) {
		c := DefaultConfig()
		c.DefaultEnabledInputs = []string{"cpu", "mem"}

		dir := t.TempDir()

		assert.NoError(t, os.MkdirAll(filepath.Join(dir, "samples"), os.ModePerm))

		// create file with content: should not overwrite on it
		assert.NoError(t, os.WriteFile(filepath.Join(dir, "samples", "cpu.conf"), []byte(`123`), os.ModePerm))

		ipts := map[string]inputs.Creator{
			"cpu": func() inputs.Input { return &mockedInput{} },
			"mem": func() inputs.Input { return &mockedInput{} },
		}

		c.initDefaultEnabledPlugins(dir, ipts)

		assert.FileExists(t, filepath.Join(dir, "samples", "cpu.conf"))

		data, err := os.ReadFile(filepath.Join(dir, "samples", "cpu.conf"))
		assert.NoError(t, err)
		assert.Equal(t, []byte(`123`), data)

		assert.FileExists(t, filepath.Join(dir, "samples", "mem.conf"))

		data, err = os.ReadFile(filepath.Join(dir, "samples", "mem.conf"))
		assert.NoError(t, err)
		assert.Equal(t, []byte(`test-sample`), data)
	})
}
