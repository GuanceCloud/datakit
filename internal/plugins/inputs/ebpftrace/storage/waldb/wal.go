// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package wal implements a wal based storage for ebpftrace
package wal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rosedblabs/wal"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ebpftrace/espan"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/ebpftrace/storage"
)

var _ storage.DB = (*ESpanDB)(nil)

type WALDB struct {
	dir string
	wal *wal.WAL

	reader *wal.Reader
}

type ESpanDB struct {
	enableInsert bool

	db     *WALDB
	metaDB *WALDB
}

func newOpt(dir string) *wal.Options {
	return &wal.Options{
		DirPath:        dir,
		SegmentSize:    wal.MB * 512,
		SegmentFileExt: ".SEG",
		BlockCache:     32 * wal.KB * 10,
		Sync:           false,
		BytesPerSync:   0,
	}
}

func NewWALChunk(dir string) (*WALDB, error) {
	wal, err := wal.Open(*newOpt(dir))
	if err != nil {
		return nil, err
	}

	return &WALDB{
		dir: dir,
		wal: wal,
	}, nil
}

func NewESpanDB(dir string) (*ESpanDB, error) {
	dirDB := filepath.Join(dir, "espan/")
	dirMeta := filepath.Join(dir, "espan_meta/")
	if err := os.MkdirAll(dirDB, os.ModePerm); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dirMeta, os.ModePerm); err != nil {
		return nil, err
	}

	datadb, err := NewWALChunk(dirDB)
	if err != nil {
		return nil, err
	}
	metadb, err := NewWALChunk(dirMeta)
	if err != nil {
		return nil, err
	}

	return &ESpanDB{
		db:     datadb,
		metaDB: metadb,
	}, nil
}

func (w *WALDB) Put(buf []byte) error {
	_, err := w.wal.Write(buf)
	return err
}

// Get returns a batch of points from wal, returns EOF if wal is empty.
func (w *WALDB) Get() ([]byte, error) {
	var buf []byte

	if w.reader == nil {
		w.reader = w.wal.NewReader()
	}

	buf, _, err := w.reader.Next()
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (w *WALDB) Sync() error {
	if w.wal != nil {
		return w.wal.Sync()
	}
	return nil
}

func (w *WALDB) Close() error {
	return w.wal.Close()
}

func (db *ESpanDB) Get() ([]*point.Point, error) {
	if db.db == nil {
		return nil, fmt.Errorf("no data db")
	}
	b, err := db.db.Get()
	if err != nil {
		return nil, err
	}

	pts, err := decodePoints(b)
	if err != nil {
		return nil, err
	}

	return pts, nil
}

func (db *ESpanDB) Put(pts []*point.Point) error {
	if db.db == nil {
		return fmt.Errorf("no data db")
	}

	b, err := encodePoints(pts)
	if err != nil {
		return err
	}
	return db.db.Put(b)
}

func (db *ESpanDB) EnableInsert() {
	db.enableInsert = true
}

func (db *ESpanDB) DisableInsert() {
	db.enableInsert = false
}

func (db *ESpanDB) Sync() error {
	if db.metaDB != nil {
		_ = db.metaDB.Sync()
	}

	if db.db != nil {
		_ = db.db.Sync()
	}
	return nil
}

func (db *ESpanDB) PutMetaList(li *espan.SpanMetaList) error {
	if db.metaDB == nil {
		return fmt.Errorf("no meta db")
	}

	if li == nil {
		return nil
	}

	b, err := encodeSpanMeta(li)
	if err != nil {
		return err
	}

	return db.metaDB.Put(b)
}

func (db *ESpanDB) GetMetaList() (*espan.SpanMetaList, error) {
	if db.metaDB == nil {
		return nil, fmt.Errorf("no meta db")
	}

	b, err := db.metaDB.Get()
	if err != nil {
		return nil, err
	}

	return deocdeSpanMeta(b)
}

func (db *ESpanDB) Close() error {
	if db.db != nil {
		return db.db.Close()
	}
	if db.metaDB != nil {
		return db.metaDB.Close()
	}
	return nil
}

func (db *ESpanDB) Drop() error {
	var err1, err2 error
	if db.db != nil {
		err1 = os.RemoveAll(db.db.dir)
	}
	if db.metaDB != nil {
		err2 = os.RemoveAll(db.metaDB.dir)
	}

	switch {
	case err1 != nil && err2 != nil:
		return fmt.Errorf("failed to drop wal(meta and data): %w, %s", err1, err2.Error())
	case err1 != nil:
		return fmt.Errorf("failed to drop wal(data): %w", err1)
	case err2 != nil:
		return fmt.Errorf("failed to drop wal(meta): %w", err2)
	}

	return nil
}

func decodePoints(b []byte) ([]*point.Point, error) {
	dec := point.GetDecoder(
		point.WithDecEncoding(point.Protobuf))
	defer point.PutDecoder(dec)

	pts, err := dec.Decode(b)
	if err != nil {
		return nil, err
	}

	return pts, nil
}

func encodePoints(pts []*point.Point) ([]byte, error) {
	enc := point.GetEncoder(
		point.WithEncEncoding(point.Protobuf),
		point.WithEncBatchSize(0))

	defer point.PutEncoder(enc)

	b, err := enc.Encode(pts)
	if err != nil {
		return nil, err
	}
	if len(b) != 1 {
		return nil, fmt.Errorf("unexpected batch size: %d", len(b))
	}

	return b[0], nil
}

func deocdeSpanMeta(b []byte) (*espan.SpanMetaList, error) {
	val := &espan.SpanMetaList{}

	if err := val.Unmarshal(b); err != nil {
		return nil, err
	}

	return val, nil
}

func encodeSpanMeta(li *espan.SpanMetaList) ([]byte, error) {
	b, err := li.Marshal()
	if err != nil {
		return nil, err
	}

	return b, nil
}

func Creator(dir string) (storage.DB, error) {
	db, err := NewESpanDB(dir)
	if err != nil {
		return nil, err
	}
	return db, nil
}
