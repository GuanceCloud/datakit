package reader

import (
	"fmt"
	"io"
)

type VarReader interface {
	VarShort() (int16, error)
	VarInt() (int32, error)
	VarLong() (int64, error)
}

type compressed struct {
	io.ByteReader
}

func NewCompressed(r io.ByteReader) VarReader {
	return compressed{ByteReader: r}
}

func (c compressed) VarShort() (int16, error) {
	n, err := c.ulong()
	if err != nil {
		return 0, err
	}
	if (n >> 48) > 0 {
		return 0, fmt.Errorf("overflow: %d bigger than 16 bits", n)
	}
	return int16(n), nil
}

func (c compressed) VarInt() (int32, error) {
	n, err := c.ulong()
	if err != nil {
		return 0, err
	}
	if (n >> 32) > 0 {
		return 0, fmt.Errorf("overflow: %d bigger than 32 bits", n)
	}
	return int32(n), nil
}

func (c compressed) VarLong() (int64, error) {
	n, err := c.ulong()
	return int64(n), err
}

func (c compressed) ulong() (n uint64, err error) {
	s := 0
	for i := 0; i < 9; i++ {
		b, err := c.ReadByte()
		if err != nil {
			return 0, err
		}
		if b&0x80 == 0 {
			n |= uint64(b) << s
			return n, nil
		}
		if i < 8 {
			b &= 0x7f
		}
		n |= uint64(b) << s
		s += 7
	}
	return n, nil
}
