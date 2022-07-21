// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint
package sender

import (
	"os"
	"path/filepath"

	"github.com/tidwall/wal"
)

// TODO.
type failCache struct {
	path                     string
	l                        *wal.Log
	firstIdx, lastIdx        uint64
	cap, size                int64
	bytesPut, bytesTruncated int64
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func initFailCache(path string, cap int64) (*failCache, error) {
	log, err := wal.Open(path, &wal.Options{LogFormat: wal.Binary, NoCopy: true})
	if err != nil {
		return nil, err
	}

	firstIdx, err := log.FirstIndex()
	if err != nil {
		return nil, err
	}

	lastIdx, err := log.LastIndex()
	if err != nil {
		return nil, err
	}

	size, err := dirSize(path)
	if err != nil {
		return nil, err
	}

	return &failCache{
		l:        log,
		path:     path,
		firstIdx: firstIdx,
		lastIdx:  lastIdx,
		cap:      cap,
		size:     size,
	}, nil
}

type callback func(data []byte) error

func (fc *failCache) get(fn callback) error {
	data, err := fc.l.Read(fc.firstIdx)
	if err != nil {
		return err
	}

	if err := fn(data); err != nil {
		return err
	}

	if err := fc.l.TruncateFront(fc.firstIdx + 1); err != nil {
		return err
	}

	fc.firstIdx++
	fc.bytesTruncated += int64(len(data))

	return nil
}

func (fc *failCache) put(data []byte) error {
	if err := fc.l.Write(fc.lastIdx+1, data); err != nil {
		return err
	}

	fc.lastIdx++
	fc.bytesPut += int64(len(data))
	return nil
}
