// Package encoding wrap internal charset encoding/decoding functions
package encoding

import (
	"errors"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
)

var (
	supportedEncodings = []string{"utf8", "utf-16le", "utf-16be", "gbk", "gb18030", "none"}

	errUnknownCharacterEncoding = errors.New("unknown character encoding")
)

type Decoder struct {
	decoder *encoding.Decoder
}

func NewDecoder(enc string) (*Decoder, error) {
	var decoder *encoding.Decoder

	switch enc {
	case "utf-8":
		decoder = unicode.UTF8.NewDecoder()
	case "utf-16le":
		decoder = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
	case "utf-16be":
		decoder = unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()
	case "gbk":
		decoder = simplifiedchinese.GBK.NewDecoder()
	case "gb18030":
		decoder = simplifiedchinese.GB18030.NewDecoder()
	case "none", "":
		decoder = encoding.Nop.NewDecoder()
	default:
		return nil, errUnknownCharacterEncoding
	}

	return &Decoder{decoder: decoder}, nil
}

func (d *Decoder) String(s string) (string, error) {
	return d.decoder.String(s)
}

func (d *Decoder) Bytes(b []byte) ([]byte, error) {
	return d.decoder.Bytes(b)
}

func DetectEncoding(s string) (string, error) {
	for _, enc := range supportedEncodings {
		d, _ := NewDecoder(enc)
		if _, err := d.String(s); err != nil {
			continue
		}
		return enc, nil
	}

	return "", errUnknownCharacterEncoding
}

func DetectEncodingForBytes(b []byte) (string, error) {
	for _, enc := range supportedEncodings {
		d, _ := NewDecoder(enc)
		if _, err := d.Bytes(b); err != nil {
			continue
		}
		return enc, nil
	}

	return "", errUnknownCharacterEncoding
}
