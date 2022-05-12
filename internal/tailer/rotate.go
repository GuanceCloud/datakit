// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"os"
	"time"
)

func DidRotate(file *os.File, lastReadOffset int64) (bool, error) {
	f, err := os.Open(file.Name())
	if err != nil {
		return false, err
	}

	defer f.Close() //nolint:errcheck,gosec

	fi1, err := f.Stat()
	if err != nil {
		return false, err
	}

	fi2, err := file.Stat()
	if err != nil {
		return true, nil //nolint:nilerr
	}

	recreated := !os.SameFile(fi1, fi2)
	truncated := fi1.Size() < lastReadOffset

	return recreated || truncated, nil
}

func FileIsActive(fn string, lastActiveDuration time.Duration) bool {
	info, err := os.Stat(fn)
	if err != nil {
		return false
	}
	if time.Since(info.ModTime()) > lastActiveDuration {
		return false
	}
	return true
}
