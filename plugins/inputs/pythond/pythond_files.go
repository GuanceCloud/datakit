// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package pythond contains pythond core scripts files.
package pythond

import (
	"embed"
	"io/ioutil"
	"os"
	"path/filepath"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

/*
files' location like this:

├── datakit
└── python.d
    ├── core
        └── datakit_framework.py
        └── demo.py
*/

func ReleaseFiles() error {
	// remove core dir
	if err := os.RemoveAll(datakit.PythonCoreDir); err != nil {
		return err
	}

	// generate new core dir
	if err := os.MkdirAll(datakit.PythonCoreDir, datakit.ConfPerm); err != nil {
		return err
	}

	if err := releaseEmbedFile(datakit.PythonCoreDir, "datakit_framework.py", &pyDatakitFramework); err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------------------------

//go:embed pys/datakit_framework.py
var pyDatakitFramework embed.FS

//go:embed pys/cli.py
var pyCli string

//------------------------------------------------------------------------------

func releaseEmbedFile(dir, name string, f *embed.FS) error {
	bys, err := f.ReadFile(filepath.Join("pys", name))
	if err != nil {
		return err
	}
	totalPath := filepath.Join(dir, name)
	if err := ioutil.WriteFile(totalPath, bys, datakit.ConfPerm); err != nil {
		return err
	}
	return nil
}
