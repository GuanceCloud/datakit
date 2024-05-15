//go:build linux
// +build linux

package protodec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrefix(t *testing.T) {
	t.Run("POST", func(t *testing.T) {
		data := []byte("POST / HTTP/1.1\r\n\r\n")
		v, ok := httpMethod(data)
		if !ok {
			t.Fatal("not req start")
		}
		assert.Equal(t, "POST", v)
	})

	t.Run("GET", func(t *testing.T) {
		data := []byte("GET / HTTP/1.1\r\n\r\n")
		v, ok := httpMethod(data)
		if !ok {
			t.Fatal("not req start")
		}
		assert.Equal(t, "GET", v)
	})

	t.Run("HTTP", func(t *testing.T) {
		data := []byte("HTTP/1.1 200 OK\r\n\r\n")
		v, code, ok := httpProtoVersion(data)
		if !ok {
			t.Fatal("not resp start")
		}
		assert.Equal(t, "1.1", v)
		assert.Equal(t, 200, code)
	})
}
