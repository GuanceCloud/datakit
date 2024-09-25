package reader

import (
	"encoding/binary"
	"io"
)

func Short(r io.Reader) (int16, error) {
	var n int16
	err := binary.Read(r, binary.BigEndian, &n)
	return n, err
}

func Int(r io.Reader) (int32, error) {
	var n int32
	err := binary.Read(r, binary.BigEndian, &n)
	return n, err
}

func Long(r io.Reader) (int64, error) {
	var n int64
	err := binary.Read(r, binary.BigEndian, &n)
	return n, err
}
