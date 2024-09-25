package parser

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/GuanceCloud/zipstream"
	"github.com/pierrec/lz4/v4"
	"io"
)

type CompressionType uint8

const (
	Unknown CompressionType = iota
	PlainJFR
	GZip
	ZIP
	LZ4
)

var (
	JFRMagic  = []byte{'F', 'L', 'R', 0}
	ZIPMagic  = []byte{0x50, 0x4b, 3, 4}
	LZ4Magic  = []byte{4, 34, 77, 24}
	GZipMagic = []byte{31, 139}
)

func hasMagic(buf []byte, magic []byte) bool {
	if len(buf) < len(magic) {
		return false
	}
	return bytes.Compare(buf[:len(magic)], magic) == 0
}

func GuessCompressionType(magic []byte) CompressionType {
	if len(magic) == 4 {
		if hasMagic(magic, ZIPMagic) {
			return ZIP
		} else if hasMagic(magic, LZ4Magic) {
			return LZ4
		} else if hasMagic(magic, JFRMagic) {
			return PlainJFR
		}
	}
	if len(magic) >= 2 && hasMagic(magic[:2], GZipMagic) {
		return GZip
	}
	return Unknown
}

func Decompress(r io.Reader) (io.ReadCloser, error) {
	buf := make([]byte, 4)
	n, err := io.ReadFull(r, buf)
	if n == 0 && err != nil {
		return nil, fmt.Errorf("unable to read file magic: %w", err)
	}

	buf = buf[:n]
	typ := GuessCompressionType(buf)
	r = io.MultiReader(bytes.NewReader(buf), r)

	switch typ {
	case GZip:
		return gzip.NewReader(r)
	case ZIP:
		zr := zipstream.NewReader(r)
		for {
			entry, err := zr.GetNextEntry()
			if err != nil {
				if err == io.EOF {
					return nil, fmt.Errorf("the zip archive does not contain any regular file")
				}
				return nil, fmt.Errorf("unable to resolve zip entry: %w", err)
			}
			if !entry.IsDir() {
				return entry.Open()
			}
		}
	case LZ4:
		return io.NopCloser(lz4.NewReader(r)), nil
	case PlainJFR:
		return io.NopCloser(r), nil
	default:
		return nil, errors.New("unsupported compression type")
	}
}
