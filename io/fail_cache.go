// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//nolint
package io

import (
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/tidwall/wal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/calcutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

var cacheSize uint64 // The total size of cache files, including mertics, logging...ALL of them.

type failCache struct {
	path                     string
	l                        *wal.Log
	cap                      int64
	bytesPut, bytesTruncated uint64
	sentLastIdx              uint64
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

	size, err := dirSize(path)
	if err != nil {
		return nil, err
	}

	atomic.AddUint64(&cacheSize, uint64(size)) // Load existing file size to memory.

	return &failCache{
		l:    log,
		path: path,
		cap:  cap,
	}, nil
}

type funcSend func(string, []*point.Point) ([]*point.Point, error)
type funcCall func([]byte, funcSend) error

func (fc *failCache) get(
	fnCall funcCall,
	fnSend funcSend,
) error {
	firstIdx, err := fc.l.FirstIndex()
	if err != nil {
		return err
	}

	lastIdx, err := fc.l.LastIndex()
	if err != nil {
		return err
	}

	if firstIdx == 0 && lastIdx == 0 {
		return nil
	}

	if fc.sentLastIdx == 0 || fc.sentLastIdx != firstIdx {
		data, err := fc.l.Read(firstIdx)
		if err != nil {
			if firstIdx == 0 && err.Error() == "not found" {
				return nil
			}
			return err
		}

		// send callback
		if err := fnCall(data, fnSend); err != nil {
			// If exceeded the limit, we drop it, oldest first.
			if atomic.LoadUint64(&cacheSize) <= uint64(fc.cap) {
				// Not exceeded, keep it and return.
				return err
			}
		}

		subLen := len(data)
		fc.bytesTruncated += uint64(subLen)
		calcutil.AtomicMinusUint64(&cacheSize, int64(subLen))
		fc.sentLastIdx = firstIdx // remember the index so that we can't send it twice or more.
	}

	if firstIdx != lastIdx {
		if err := fc.l.TruncateFront(firstIdx + 1); err != nil {
			return err
		}
	} else {
		// already sent, so we delete the file here.
		if err := fc.l.Close(); err != nil {
			return err
		}
		if err := os.RemoveAll(fc.path); err != nil {
			return err
		}
		newL, err := wal.Open(fc.path, &wal.Options{LogFormat: wal.Binary, NoCopy: true})
		if err != nil {
			return err
		}

		fc.l = newL
		fc.sentLastIdx = 0
	}

	return nil
}

func (fc *failCache) put(data []byte) error {
	lastIdx, err := fc.l.LastIndex()
	if err != nil {
		return err
	}

	if err := fc.l.Write(lastIdx+1, data); err != nil {
		return err
	}

	addLen := uint64(len(data))
	fc.bytesPut += addLen
	atomic.AddUint64(&cacheSize, addLen)

	return nil
}
