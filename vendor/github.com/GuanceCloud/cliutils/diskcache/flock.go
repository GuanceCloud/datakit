// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diskcache

import (
	"os"
	"path/filepath"
)

type walLock struct {
	file string
	f    *os.File
}

func newFlock(path string) *walLock {
	file := filepath.Clean(filepath.Join(path, ".lock"))

	return &walLock{
		file: file,
	}
}
