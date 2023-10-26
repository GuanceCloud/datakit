# buf-readerat

[![GoDoc](https://godoc.org/github.com/avvmoto/buf-readerat?status.svg)](https://godoc.org/github.com/avvmoto/buf-readerat)
[![Build Status](https://api.travis-ci.org/avvmoto/buf-readerat.svg?branch=master)](https://travis-ci.org/avvmoto/buf-readerat)

Package buf-readerat implements buffered io.ReaderAt. It wraps an io.ReaderAt
object, creating another io.ReaderAt object that also implements the interface
but provides buffering.

This package is especially useful for packages those who require io.ReaderAt for
initialize, like [zip.Reader](https://golang.org/pkg/archive/zip/).

## Installation

`go get github.com/avvmoto/buf-readerat`

## Usage

### Use BufReaderAt

``` go
r := bytes.NewReader([]byte("123456789"))

bra := bufra.NewBufReaderAt(r, 8)

buf := make([]byte, 4)
if _, err := bra.ReadAt(buf, 4); err != nil {
	log.Fatal(err)
}
fmt.Printf("%s\n", buf)
```


### Use Zip file on Google Cloud Storage within Google App Engine

``` go

import "cloud.google.com/go/storage"


// wrap storage.NewRangeReader to use as io.ReaderAt
type rowGCSReaderAt struct {
	ctx context.Context
	obj *storage.ObjectHandle
}

func newRowGCSReaderAt(ctx context.Context, obj *storage.ObjectHandle) *rowGCSReaderAt {
	return &rowGCSReaderAt{
		ctx: ctx,
		obj: obj,
	}
}

// Wrap NewRangeReader
func (ra *rowGCSReaderAt) ReadAt(b []byte, off int64) (n int, err error) {
	if ra == nil {
		return 0, errors.New("invalid")
	}

	rc, err := ra.obj.NewRangeReader(ra.ctx, off, int64(len(b)))
	if err != nil {
		return
	}
	defer rc.Close()

	n, err = rc.Read(b)
	return
}

type GCSReaderAt struct {
	cache *BufReaderAt
}

func NewGCSReaderAt(ctx context.Context, obj *storage.ObjectHandle) (*GCSReaderAt) {
	rowReaderAt := newRowGCSReaderAt(ctx, obj)
	cra := NewBufReaderAt(rowReaderAt, cacheSize)

	return &GCSReaderAt{
		cache: cra,
	}
}

func (ra *GCSReaderAt) ReadAt(b []byte, off int64) (n int, err error) {
	return ra.cache.ReadAt(b, off)
}
```


## License

Copyright (c) 2016-present [avvmoto](https://github.com/avvmoto)

Licensed under [MIT License](./LICENSE)
