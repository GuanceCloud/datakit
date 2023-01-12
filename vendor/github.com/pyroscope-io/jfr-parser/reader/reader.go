package reader

import (
	"encoding/binary"
	"fmt"
	"io"
)

type VarReader interface {
	VarShort() (int16, error)
	VarInt() (int32, error)
	VarLong() (int64, error)
}

type Reader interface {
	Boolean() (bool, error)
	Byte() (int8, error)
	Short() (int16, error)
	Char() (uint16, error)
	Int() (int32, error)
	Long() (int64, error)
	Float() (float32, error)
	Double() (float64, error)
	String() (string, error)

	VarReader

	// TODO: Support arrays
}

type InputReader interface {
	io.Reader
	io.ByteReader
}

type reader struct {
	InputReader
	varR VarReader
}

func NewReader(r InputReader, compressed bool) Reader {
	var varR VarReader
	if compressed {
		varR = newCompressed(r)
	} else {
		varR = newUncompressed(r)
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
	return Short(r)
}

func (r reader) Char() (uint16, error) {
	var n uint16
	err := binary.Read(r, binary.BigEndian, &n)
	return n, err
}

func (r reader) Int() (int32, error) {
	return Int(r)
}

func (r reader) Long() (int64, error) {
	return Long(r)
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
func (r reader) String() (string, error) {
	enc, err := r.Byte()
	if err != nil {
		return "", err
	}
	switch enc {
	case 0:
		return "", nil
	case 1:
		return "", nil
	case 3, 4, 5:
		return r.utf8()
	default:
		// TODO
		return "", fmt.Errorf("Unsupported string type :%d", enc)
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
