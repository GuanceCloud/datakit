//go:build linux
// +build linux

package protodec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAMQPHeader(t *testing.T) {
	payload := []byte{'\x41', '\x4d', '\x51', '\x50', '\x00', '\x00', '\x09', '\x01'}
	assert.Equal(t, true, checkAMQP(payload))

	data := "\x01\x00\x01\x00\x00\x00\x05\x00\x14\x00\x0a\x00\xce"

	assert.Equal(t, true, checkAMQP([]byte(data)))
}
