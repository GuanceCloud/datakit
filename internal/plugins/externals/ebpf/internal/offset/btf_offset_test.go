//go:build linux && vmlinux
// +build linux,vmlinux

package offset

import (
	"testing"

	"github.com/cilium/ebpf/btf"
	"github.com/stretchr/testify/assert"
)

func TestGetTCPOffsetFromBTF(t *testing.T) {
	if _, v, err := TryGetTCPSeqOffsetFromBTF(); err != nil {
		t.Skipf("BTF not available: %v", err)
	} else {
		t.Logf("TryGetTCPSeqOffsetFromBTF succeeded: v=%+v", v)
	}

	offsets, err := GetTCPOffsetFromBTF()
	if err != nil {
		t.Skipf("BTF not available: %v", err)
		return
	}

	assert.NotEmpty(t, offsets, "offsets should not be empty")

	expectedFields := []string{
		"tcp_sk_srtt_us",
		"tcp_sk_mdev_us",
		"inet_sport",
		"sk_family",
		"sk_v6_rcv_saddr",
		"sk_v6_daddr",
		"sk_net",
		"ns_common_inum",
	}

	for _, field := range expectedFields {
		val, exists := offsets[field]
		assert.True(t, exists, "field %s should exist", field)
		assert.NotEqual(t, uint64(0), val, "field %s should not be zero", field)
		t.Logf("%s: %d", field, val)
	}
}

func TestGetMemberOffset(t *testing.T) {
	spec, err := btf.LoadKernelSpec()
	if err != nil {
		t.Skipf("BTF not available: %v", err)
		return
	}

	tests := []struct {
		name        string
		typeName    string
		memberName  string
		expectError bool
	}{
		{"tcp_sock_srtt_us", "tcp_sock", "srtt_us", false},
		{"tcp_sock_mdev_us", "tcp_sock", "mdev_us", false},
		{"sock_common_skc_net", "sock_common", "skc_net", false},
		{"invalid_type", "a_invalid_type", "field", true},
		{"invalid_member", "tcp_sock", "a_invalid_field", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s *btf.Struct
			err := spec.TypeByName(tt.typeName, &s)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if err == nil && s != nil {
					found := false
					for _, m := range s.Members {
						if m.Name == tt.memberName {
							offset := uint64(m.Offset.Bytes())
							assert.NotEqual(t, uint64(0), offset, "offset should not be zero")
							t.Logf("%s.%s: %d", tt.typeName, tt.memberName, offset)
							found = true
							break
						}
					}
					if tt.name == "invalid_member" {
						assert.False(t, found, "member should not be found")
					} else {
						assert.True(t, found, "member %s should be found", tt.memberName)
					}
				} else if tt.name == "invalid_member" {
					assert.Error(t, err)
				}
			}
		})
	}
}
