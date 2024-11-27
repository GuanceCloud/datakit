// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package fileprovider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanFiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "testing")
	assert.NoError(t, err)

	defer func() {
		err := os.Remove(dir)
		assert.NoError(t, err)
		t.Logf("remove directory %s", dir)
	}()

	file, err := os.CreateTemp(dir, "")
	assert.NoError(t, err)
	filename := file.Name()

	defer func() {
		err := os.Remove(filename)
		assert.NoError(t, err)
		t.Logf("remove file %s", filename)
	}()

	testcases := []struct {
		in, out []string
	}{
		{
			in:  []string{filename[:len(filename)-1] + "*"},
			out: []string{filename},
		},
		{
			in:  []string{filepath.Join(dir, "**", "*")},
			out: []string{filename},
		},
	}

	for _, tc := range testcases {
		t.Logf("glob pattern %s", tc.in)
		sc, err := NewScanner(tc.in)
		assert.NoError(t, err)

		res, err := sc.ScanFiles()
		assert.NoError(t, err)
		assert.Equal(t, tc.out, res)
	}
}
