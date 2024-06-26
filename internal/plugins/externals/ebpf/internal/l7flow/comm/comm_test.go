//go:build linux
// +build linux

// Package comm stores connection information
package comm

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDump(t *testing.T) {
	d := NetwrkData{
		Conn: ConnectionInfo{
			Saddr: [4]uint32{1},
		},
		Payload: []byte("assss\n"),
	}
	d.Payload = append(d.Payload, 11)

	s, _ := json.Marshal(d)
	t.Log(string(s))
	d2 := NetwrkData{}
	json.Unmarshal(s, &d2)
	assert.Equal(t, d, d2)
	assert.Equal(t, true, bytes.Equal(d.Payload, d2.Payload))
}
