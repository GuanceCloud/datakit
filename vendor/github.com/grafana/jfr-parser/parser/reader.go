package parser

import (
	"encoding/binary"
	"fmt"
	"io"

	reader2 "github.com/grafana/jfr-parser/reader"
)

const (
	StringEncodingNull            = 0
	StringEncodingEmptyString     = 1
	StringEncodingConstantPool    = 2
	StringEncodingUtf8ByteArray   = 3
	StringEncodingCharArray       = 4
	StringEncodingLatin1ByteArray = 5
)

type Reader interface {
	Boolean() (bool, error)
	Byte() (int8, error)
	Short() (int16, error)
	Char() (uint16, error)
	Int() (int32, error)
	Long() (int64, error)
	Float() (float32, error)
	Double() (float64, error)
	String() (*String, error)

	reader2.VarReader

	// TODO: Support arrays
}

type InputReader interface {
	io.Reader
	io.ByteReader
}

type reader struct {
	InputReader
	varR reader2.VarReader
}

func NewReader(r InputReader, compressed bool) Reader {
	var varR reader2.VarReader
	if compressed {
		varR = reader2.NewCompressed(r)
	} else {
		varR = reader2.NewUncompressed(r)
	}
	return reader{
		InputReader: r,
		varR:        varR,
	}
}

func (r reader) Boolean() (bool, error) {
	var n int8
	err := binary.Read(r, binary.BigEndian, &n)
	if n == 0 {
		return false, err
	}
	return true, err
}

func (r reader) Byte() (int8, error) {
	var n int8
	err := binary.Read(r, binary.BigEndian, &n)
	return n, err
}

func (r reader) Short() (int16, error) {
	return reader2.Short(r)
}

func (r reader) Char() (uint16, error) {
	var n uint16
	err := binary.Read(r, binary.BigEndian, &n)
	return n, err
}

func (r reader) Int() (int32, error) {
	return reader2.Int(r)
}

func (r reader) Long() (int64, error) {
	return reader2.Long(r)
}

func (r reader) Float() (float32, error) {
	var n float32
	err := binary.Read(r, binary.BigEndian, &n)
	return n, err
}

func (r reader) Double() (float64, error) {
	var n float64
	err := binary.Read(r, binary.BigEndian, &n)
	return n, err
}

// TODO: Should we differentiate between null and empty?
func (r reader) String() (*String, error) {
	s := new(String)
	enc, err := r.Byte()
	if err != nil {
		return nil, err
	}
	switch enc {
	case StringEncodingNull:
		return s, nil
	case StringEncodingEmptyString:
		return s, nil
	case StringEncodingConstantPool:
		idx, err := r.VarLong()
		if err != nil {
			return nil, fmt.Errorf("unable to resolve constant refrence index: %w", err)
		}
		s.constantRef = &constantReference{index: idx}
		return s, nil
	case StringEncodingUtf8ByteArray, StringEncodingCharArray, StringEncodingLatin1ByteArray:
		str, err := r.utf8()
		if err != nil {
			return nil, err
		}
		s.s = str
		return s, nil
	default:
		return nil, fmt.Errorf("unsupported string type :%d", enc)
	}
}

func (r reader) VarShort() (int16, error) {
	return r.varR.VarShort()
}

func (r reader) VarInt() (int32, error) {
	return r.varR.VarInt()
}

func (r reader) VarLong() (int64, error) {
	return r.varR.VarLong()
}

func (r reader) utf8() (string, error) {
	n, err := r.varR.VarInt()
	if err != nil {
		return "", nil
	}
	// TODO: make sure n is reasonable
	b := make([]byte, n)
	_, err = io.ReadFull(r, b)
	return string(b), err
}
