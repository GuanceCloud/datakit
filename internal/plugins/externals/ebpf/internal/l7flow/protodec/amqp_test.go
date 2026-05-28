//go:build linux
// +build linux

package protodec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/l7flow/comm"
)

func TestAMQPHeader(t *testing.T) {
	payload := []byte{'\x41', '\x4d', '\x51', '\x50', '\x00', '\x00', '\x09', '\x01'}
	assert.Equal(t, true, checkAMQP(payload))

	data := "\x01\x00\x01\x00\x00\x00\x05\x00\x14\x00\x0a\x00\xce"

	assert.Equal(t, true, checkAMQP([]byte(data)))
}

func TestAMQPDecodeRejectsOversizedFrame(t *testing.T) {
	payload := []byte{0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x20, 0x00, 0x14, 0x00, 0x0a, 0xce}
	dec := &amqpDecPipe{}
	err := dec.decode(comm.NICDIngress, &comm.NetwrkData{Payload: payload}, 0, nil)
	if err == nil {
		t.Fatal("expected oversized AMQP frame to be rejected")
	}
}
