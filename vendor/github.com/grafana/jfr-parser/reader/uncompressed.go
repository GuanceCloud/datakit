package reader

import (
	"io"
)

type uncompressed struct {
	io.Reader
}

func NewUncompressed(r io.Reader) VarReader {
	return uncompressed{Reader: r}
}

func (c uncompressed) VarShort() (int16, error) {
	return Short(c)
}

func (c uncompressed) VarInt() (int32, error) {
	return Int(c)
}

func (c uncompressed) VarLong() (int64, error) {
	return Long(c)
}
