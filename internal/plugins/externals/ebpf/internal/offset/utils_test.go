//go:build linux
// +build linux

package offset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOffsetDumpAndLoad(t *testing.T) {
	offsetExpected := OffsetGuessC{
		offset_sk_num:            _Ctype_ulonglong(1),
		offset_inet_sport:        _Ctype_ulonglong(12),
		offset_sk_family:         _Ctype_ulonglong(123),
		offset_sk_rcv_saddr:      _Ctype_ulonglong(1234),
		offset_sk_daddr:          _Ctype_ulonglong(12345),
		offset_sk_v6_rcv_saddr:   _Ctype_ulonglong(123456),
		offset_sk_v6_daddr:       _Ctype_ulonglong(1234567),
		offset_sk_dport:          _Ctype_ulonglong(12345678),
		offset_tcp_sk_srtt_us:    _Ctype_ulonglong(123456789),
		offset_tcp_sk_mdev_us:    _Ctype_ulonglong(1234567890),
		offset_flowi4_saddr:      _Ctype_ulonglong(12345678901),
		offset_flowi4_daddr:      _Ctype_ulonglong(123456789012),
		offset_flowi4_sport:      _Ctype_ulonglong(1234567890123),
		offset_flowi4_dport:      _Ctype_ulonglong(12345678901234),
		offset_flowi6_saddr:      _Ctype_ulonglong(123456789012345),
		offset_flowi6_daddr:      _Ctype_ulonglong(1234567890123456),
		offset_flowi6_sport:      _Ctype_ulonglong(12345678901234567),
		offset_flowi6_dport:      _Ctype_ulonglong(123456789012345678),
		offset_skaddr_sin_port:   _Ctype_ulonglong(1234567890123456789),
		offset_skaddr6_sin6_port: _Ctype_ulonglong(3234567890123456789),
		offset_sk_net:            _Ctype_ulonglong(5234567890123456789),
		offset_ns_common_inum:    _Ctype_ulonglong(7234567890123456789),
		offset_socket_sk:         _Ctype_ulonglong(9234567890123456789),
	}
	str, err := DumpOffset(offsetExpected)
	if err != nil {
		t.Fatal(err)
	}

	offsetActual, err := LoadOffset(str)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, offsetExpected, offsetActual)
}
