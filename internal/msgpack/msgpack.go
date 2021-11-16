// Package msgpack wraps msgpack related functions.
package msgpack

import (
	"errors"
	"io"
	"reflect"

	"github.com/ugorji/go/codec"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/bufpool"
)

var (
	msgpHandler codec.MsgpackHandle
	encoder     = codec.NewEncoder(nil, &msgpHandler)
	decoder     = codec.NewDecoder(nil, &msgpHandler)
)

func Marshal(src interface{}) ([]byte, error) {
	buf := bufpool.GetBuffer()
	encoder.Reset(buf)
	err := encoder.Encode(src)

	return buf.Bytes(), err
}

func Unmarshal(src io.Reader, dest interface{}) error {
	if src == nil || dest == nil || reflect.ValueOf(dest).Kind() != reflect.Ptr {
		return errors.New("invalid parameters for msgpack.Unmarshal")
	}

	decoder.Reset(src)

	return decoder.Decode(dest)
}
