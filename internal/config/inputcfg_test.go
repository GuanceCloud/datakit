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

type mockedInput struct {
	sample,
	catalog string
}

/*

	Catalog() string
	Run()
	SampleConfig() string
*/

func (i *mockedInput) Catalog() string {
	if i.catalog == "" {
		return "samples"
	}
	return i.catalog
}

func (i *mockedInput) Run() {}

func (i *mockedInput) SampleConfig() string {
	if i.sample == "" {
		return "test-sample"
	}

	return i.sample
}

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

		assert.FileExists(t, filepath.Join(dir, "cpu.conf"))
		assert.FileExists(t, filepath.Join(dir, "mem.conf"))
	})

	t.Run(`conf-exist-as-dir`, func(t *T.T) {
		c := DefaultConfig()
		c.DefaultEnabledInputs = []string{"cpu", "mem"}

		dir := t.TempDir()

		assert.NoError(t, os.MkdirAll(filepath.Join(dir, "cpu.conf"), os.ModePerm))

		ipts := map[string]inputs.Creator{
			"cpu": func() inputs.Input { return &mockedInput{} },
			"mem": func() inputs.Input { return &mockedInput{} },
		}

		c.initDefaultEnabledPlugins(dir, ipts)

		assert.FileExists(t, filepath.Join(dir, "cpu-0xdeadbeaf.conf"))
		assert.FileExists(t, filepath.Join(dir, "mem.conf"))
	})

	t.Run(`conf-exist-and-skip`, func(t *T.T) {
		c := DefaultConfig()
		c.DefaultEnabledInputs = []string{"cpu", "mem"}

		dir := t.TempDir()

		assert.NoError(t, os.MkdirAll(dir, os.ModePerm))

		// create file with content: should not overwrite on it
		assert.NoError(t, os.WriteFile(filepath.Join(dir, "cpu.conf"), []byte(`123`), os.ModePerm))

		ipts := map[string]inputs.Creator{
			"cpu": func() inputs.Input { return &mockedInput{} },
			"mem": func() inputs.Input { return &mockedInput{} },
		}

		c.initDefaultEnabledPlugins(dir, ipts)

		assert.FileExists(t, filepath.Join(dir, "cpu.conf"))

		data, err := os.ReadFile(filepath.Join(dir, "cpu.conf"))
		assert.NoError(t, err)
		assert.Equal(t, []byte(`123`), data)

		assert.FileExists(t, filepath.Join(dir, "mem.conf"))

		data, err = os.ReadFile(filepath.Join(dir, "mem.conf"))
		assert.NoError(t, err)
		assert.Equal(t, []byte(`test-sample`), data)
	})

	t.Run(`conf-legacy-exist-and-skip`, func(t *T.T) {
		c := DefaultConfig()
		c.DefaultEnabledInputs = []string{"cpu", "mem"}
		catalogHost := "host"

		dir := t.TempDir()
		assert.NoError(t, os.MkdirAll(dir, os.ModePerm))
		assert.NoError(t, os.MkdirAll(filepath.Join(dir, catalogHost), os.ModePerm))

		// create legacy conf: should not overwrite on it
		assert.NoError(t, os.WriteFile(filepath.Join(dir,
			catalogHost,
			"cpu.conf"), []byte(`123`), os.ModePerm))

		ipts := map[string]inputs.Creator{
			"cpu": func() inputs.Input { return &mockedInput{catalog: catalogHost} },
			"mem": func() inputs.Input { return &mockedInput{catalog: "not-exist-cataglog"} },
		}

		c.initDefaultEnabledPlugins(dir, ipts)

		assert.NoFileExists(t, filepath.Join(dir, "cpu.conf"))
		assert.NoFileExists(t, filepath.Join(dir, catalogHost, "mem.conf"))
		assert.FileExists(t, filepath.Join(dir, "mem.conf")) // mem.conf legacy conf not created

		data, err := os.ReadFile(filepath.Join(dir, catalogHost, "cpu.conf"))
		assert.NoError(t, err)
		assert.Equal(t, []byte(`123`), data)

		assert.FileExists(t, filepath.Join(dir, "mem.conf"))

		data, err = os.ReadFile(filepath.Join(dir, "mem.conf"))
		assert.NoError(t, err)
		assert.Equal(t, []byte(`test-sample`), data)
	})
}
