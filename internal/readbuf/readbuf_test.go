package readbuf

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadLines(t *testing.T) {
	const bufSize = 30

	cases := []struct {
		rd  io.Reader
		res [][]byte
	}{
		{
			rd: bytes.NewBuffer([]byte("012345678901234567890123456789\nABCD\n")),
			res: [][]byte{
				[]byte("012345678901234567890123456789"),
				[]byte("ABCD"),
			},
		},
		{
			rd: bytes.NewBuffer([]byte("01234567890123456789\n")),
			res: [][]byte{
				[]byte("01234567890123456789"),
			},
		},
		{
			rd: bytes.NewBuffer([]byte("0123456789012345678901234567890123456789\n")),
			res: [][]byte{
				[]byte("0123456789012345678901234567890123456789"),
			},
		},
		{
			rd: bytes.NewBuffer([]byte("ABCDEFG\n1234567890\nABCDEFG\n")),
			res: [][]byte{
				[]byte("ABCDEFG"),
				[]byte("1234567890"),
				[]byte("ABCDEFG"),
			},
		},
	}

	for _, tc := range cases {
		reader := NewReadBuffer(tc.rd, bufSize)

		for i := 0; i < len(tc.res); {
			lines, err := reader.ReadLines()
			assert.NoError(t, err)

			for _, line := range lines {
				assert.Equal(t, tc.res[i], line)
				i++
			}
		}
	}
}
