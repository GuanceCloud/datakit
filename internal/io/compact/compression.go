// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package compact

import (
	"fmt"
	"strings"
	"sync"

	"github.com/klauspost/compress/zstd"
)

// Compression identifies the HTTP body compression used for Point data.
type Compression string

const (
	// CompressionUnknown means the body has not been prepared for upload.
	CompressionUnknown Compression = ""
	// CompressionIdentity means the body is sent without compression.
	CompressionIdentity Compression = "identity"
	// CompressionGzip means the body is gzip-compressed.
	CompressionGzip Compression = "gzip"
	// CompressionZstd means the body is zstd-compressed.
	CompressionZstd Compression = "zstd"
)

func (c Compression) String() string {
	return string(c)
}

// ParseCompression parses a persisted or configured compression name.
func ParseCompression(value string) Compression {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case CompressionIdentity.String():
		return CompressionIdentity
	case CompressionGzip.String():
		return CompressionGzip
	case CompressionZstd.String():
		return CompressionZstd
	default:
		return CompressionUnknown
	}
}

// HTTPContentEncoding returns the value for the Content-Encoding header.
func (c Compression) HTTPContentEncoding() string {
	if c == CompressionGzip || c == CompressionZstd {
		return c.String()
	}
	return ""
}

// Compressor compresses Point request bodies into caller-owned storage.
type Compressor interface {
	Encoding() Compression
	Compress(dst, src []byte) ([]byte, error)
}

// NewCompressor creates a compressor for encoding.
func NewCompressor(encoding Compression) (Compressor, error) {
	switch encoding {
	case CompressionUnknown:
		return nil, fmt.Errorf("unsupported compression %q", encoding)
	case CompressionIdentity:
		return identityCompressor{}, nil
	case CompressionGzip:
		return gzipCompressor{}, nil
	case CompressionZstd:
		compressor := &zstdCompressor{}
		writer, err := newZstdWriter()
		if err != nil {
			return nil, err
		}
		compressor.pool.Put(writer)
		return compressor, nil
	default:
		return nil, fmt.Errorf("unsupported compression %q", encoding)
	}
}

type identityCompressor struct{}

func (identityCompressor) Encoding() Compression {
	return CompressionIdentity
}

func (identityCompressor) Compress(_, src []byte) ([]byte, error) {
	return src, nil
}

type gzipCompressor struct{}

func (gzipCompressor) Encoding() Compression {
	return CompressionGzip
}

func (gzipCompressor) Compress(dst, src []byte) ([]byte, error) {
	writer := GetZipper()
	defer PutZipper(writer)

	compressed, err := writer.Zip(src)
	if err != nil {
		return nil, err
	}
	return copyCompressed(dst, compressed), nil
}

type zstdCompressor struct {
	pool sync.Pool
}

type zstdWriter struct {
	encoder *zstd.Encoder
	buf     []byte
}

func newZstdWriter() (*zstdWriter, error) {
	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderConcurrency(1))
	if err != nil {
		return nil, fmt.Errorf("create zstd encoder: %w", err)
	}
	return &zstdWriter{encoder: encoder}, nil
}

func (c *zstdCompressor) Encoding() Compression {
	return CompressionZstd
}

func (c *zstdCompressor) Compress(dst, src []byte) ([]byte, error) {
	var writer *zstdWriter
	if pooled := c.pool.Get(); pooled == nil {
		var err error
		writer, err = newZstdWriter()
		if err != nil {
			return nil, err
		}
	} else {
		writer = pooled.(*zstdWriter)
	}

	writer.buf = writer.encoder.EncodeAll(src, writer.buf[:0])
	compressed := copyCompressed(dst, writer.buf)
	c.pool.Put(writer)
	return compressed, nil
}

func copyCompressed(dst, src []byte) []byte {
	if cap(dst) < len(src) {
		dst = make([]byte, len(src))
	} else {
		dst = dst[:len(src)]
	}
	copy(dst, src)
	return dst
}
