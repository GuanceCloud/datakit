//go:build linux
// +build linux

package offset

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	bpfutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/externals/ebpf/internal/bpfutil"
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
		offset_task_struct_files: _Ctype_ulonglong(128),
		offset_files_struct_fdt:  _Ctype_ulonglong(136),
		offset_socket_file:       _Ctype_ulonglong(24),
		offset_file_private_data: _Ctype_ulonglong(168),
		offset_ct_net:            _Ctype_ulonglong(64),
		offset_ct_ns_common_inum: _Ctype_ulonglong(84),
		offset_origin_tuple:      _Ctype_ulonglong(72),
		offset_reply_tuple:       _Ctype_ulonglong(104),
	}
	str, err := dumpOffset(offsetExpected)
	if err != nil {
		t.Fatal(err)
	}

	offsetActual, err := loadOffset(str)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, offsetExpected, offsetActual)
	assert.Contains(t, str, fmt.Sprintf("\"version\":\"%s\"", offsetCacheVersion))
	assert.Contains(t, str, "\"meta\":{\"kernel_version\":")
	assert.Contains(t, str, "\"kernel\":{")
	assert.Contains(t, str, "\"tcp_seq\":{")
	assert.Contains(t, str, "\"httpflow\":{")
	assert.Contains(t, str, "\"conntrack\":{")
	assert.NotContains(t, str, "\",\"kernel_version\":\"")
	assert.NotContains(t, str, "\",\"offset_ct_net\":64,")
}

func TestLoadOffsetSupportsLegacyFlatSchema(t *testing.T) {
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
		offset_task_struct_files: _Ctype_ulonglong(128),
		offset_files_struct_fdt:  _Ctype_ulonglong(136),
		offset_socket_file:       _Ctype_ulonglong(24),
		offset_file_private_data: _Ctype_ulonglong(168),
		offset_ct_net:            _Ctype_ulonglong(64),
		offset_ct_ns_common_inum: _Ctype_ulonglong(84),
		offset_origin_tuple:      _Ctype_ulonglong(72),
		offset_reply_tuple:       _Ctype_ulonglong(104),
	}

	kernelVersion, err := currentOffsetCacheKernelVersion()
	if err != nil {
		t.Fatalf("currentOffsetCacheKernelVersion: %v", err)
	}

	legacy := newLegacyOffsetCache(offsetExpected, kernelVersion)
	str, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("marshal legacy offset cache: %v", err)
	}

	offsetActual, err := loadOffset(string(str))
	if err != nil {
		t.Fatalf("loadOffset legacy: %v", err)
	}

	assert.Equal(t, offsetExpected, offsetActual)
}

func TestApplyConstantPatches(t *testing.T) {
	offset := &OffsetGuessC{}

	ApplyConstantPatches(offset, []bpfutil.ConstantPatch{
		{Name: "offset_task_struct_files", Value: uint64(1880)},
		{Name: "offset_files_struct_fdt", Value: uint64(8)},
		{Name: "offset_socket_file", Value: uint64(24)},
		{Name: "offset_file_private_data", Value: uint64(168)},
		{Name: "offset_socket_sk", Value: uint64(32)},
		{Name: "offset_ct_net", Value: uint64(64)},
		{Name: "offset_ct_ns_common_inum", Value: uint64(84)},
		{Name: "offset_ct_origin_tuple", Value: uint64(72)},
		{Name: "offset_ct_reply_tuple", Value: uint64(104)},
	})

	assert.Equal(t, _Ctype_ulonglong(1880), offset.offset_task_struct_files)
	assert.Equal(t, _Ctype_ulonglong(8), offset.offset_files_struct_fdt)
	assert.Equal(t, _Ctype_ulonglong(24), offset.offset_socket_file)
	assert.Equal(t, _Ctype_ulonglong(168), offset.offset_file_private_data)
	assert.Equal(t, _Ctype_ulonglong(32), offset.offset_socket_sk)
	assert.Equal(t, _Ctype_ulonglong(64), offset.offset_ct_net)
	assert.Equal(t, _Ctype_ulonglong(84), offset.offset_ct_ns_common_inum)
	assert.Equal(t, _Ctype_ulonglong(72), offset.offset_origin_tuple)
	assert.Equal(t, _Ctype_ulonglong(104), offset.offset_reply_tuple)
}

func TestKernelGuessRequiredReady(t *testing.T) {
	check := &OffsetCheck{
		skDaddrOk:     MINSUCCESS + 1,
		skDportOk:     MINSUCCESS + 1,
		skFamilyOk:    MINSUCCESS + 1,
		sknetOk:       MINSUCCESS + 1,
		netnsInumOk:   MINSUCCESS + 1,
		flowi4SaddrOk: MINSUCCESS + 1,
		flowi4DaddrOk: MINSUCCESS + 1,
		flowi4DportOk: MINSUCCESS + 1,
	}

	assert.True(t, kernelGuessRequiredReady(check, true, true, false))
	assert.False(t, kernelGuessOptionalReady(check, true, false))
}

func TestKernelGuessOptionalReady(t *testing.T) {
	check := &OffsetCheck{
		skDaddrOk:     MINSUCCESS + 1,
		skDportOk:     MINSUCCESS + 1,
		skFamilyOk:    MINSUCCESS + 1,
		sknetOk:       MINSUCCESS + 1,
		netnsInumOk:   MINSUCCESS + 1,
		flowi4SaddrOk: MINSUCCESS + 1,
		flowi4DaddrOk: MINSUCCESS + 1,
		flowi4DportOk: MINSUCCESS + 1,
		inetSportOk:   MINSUCCESS + 1,
		tcpSkSrttUsOk: MINSUCCESS + 1,
		tcpSkMdevUsOk: MINSUCCESS + 1,
		socketSkOK:    MINSUCCESS + 1,
	}

	assert.True(t, kernelGuessRequiredReady(check, true, true, false))
	assert.True(t, kernelGuessOptionalReady(check, true, false))
}

func TestFinalizeKernelGuess(t *testing.T) {
	status := &OffsetGuessC{
		offset_sk_dport:          12,
		offset_sk_num:            14,
		offset_sk_daddr:          0,
		offset_sk_rcv_saddr:      4,
		offset_flowi4_saddr:      40,
		offset_flowi4_daddr:      44,
		offset_flowi4_dport:      48,
		offset_flowi4_sport:      50,
		offset_task_struct_files: 1880,
		offset_files_struct_fdt:  8,
		offset_socket_file:       24,
		offset_file_private_data: 168,
		offset_ct_net:            64,
		offset_ct_ns_common_inum: 84,
		offset_origin_tuple:      72,
		offset_reply_tuple:       104,
	}

	got := finalizeKernelGuess(status)
	assert.Equal(t, _Ctype_ulonglong(40), got.offset_flowi6_daddr)
	assert.Equal(t, _Ctype_ulonglong(56), got.offset_flowi6_saddr)
	assert.Equal(t, _Ctype_ulonglong(76), got.offset_flowi6_dport)
	assert.Equal(t, _Ctype_ulonglong(78), got.offset_flowi6_sport)
	assert.Equal(t, _Ctype_ulonglong(1880), got.offset_task_struct_files)
	assert.Equal(t, _Ctype_ulonglong(8), got.offset_files_struct_fdt)
	assert.Equal(t, _Ctype_ulonglong(24), got.offset_socket_file)
	assert.Equal(t, _Ctype_ulonglong(168), got.offset_file_private_data)
	assert.Equal(t, _Ctype_ulonglong(64), got.offset_ct_net)
	assert.Equal(t, _Ctype_ulonglong(84), got.offset_ct_ns_common_inum)
	assert.Equal(t, _Ctype_ulonglong(72), got.offset_origin_tuple)
	assert.Equal(t, _Ctype_ulonglong(104), got.offset_reply_tuple)
}

func TestDumpOffsetSkipsUnchangedWrites(t *testing.T) {
	dir := t.TempDir()
	offsetPath := filepath.Join(dir, "externals", "datakit-ebpf.offset")
	offset := &OffsetGuessC{
		offset_sk_num:            14,
		offset_sk_family:         16,
		offset_sk_rcv_saddr:      4,
		offset_sk_dport:          12,
		offset_flowi4_saddr:      40,
		offset_flowi4_daddr:      44,
		offset_flowi4_sport:      50,
		offset_flowi4_dport:      48,
		offset_task_struct_files: 2704,
		offset_files_struct_fdt:  32,
		offset_socket_file:       24,
		offset_file_private_data: 200,
	}

	if err := DumpOffset(dir, offset); err != nil {
		t.Fatalf("first DumpOffset: %v", err)
	}

	infoBefore, err := os.Stat(offsetPath)
	if err != nil {
		t.Fatalf("stat before: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	if err := DumpOffset(dir, offset); err != nil {
		t.Fatalf("second DumpOffset: %v", err)
	}

	infoAfter, err := os.Stat(offsetPath)
	if err != nil {
		t.Fatalf("stat after unchanged dump: %v", err)
	}

	if !infoAfter.ModTime().Equal(infoBefore.ModTime()) {
		t.Fatalf("expected unchanged dump to keep modtime, before=%v after=%v",
			infoBefore.ModTime(), infoAfter.ModTime())
	}

	offset.offset_file_private_data = 168

	time.Sleep(20 * time.Millisecond)

	if err := DumpOffset(dir, offset); err != nil {
		t.Fatalf("third DumpOffset: %v", err)
	}

	infoUpdated, err := os.Stat(offsetPath)
	if err != nil {
		t.Fatalf("stat after changed dump: %v", err)
	}

	if !infoUpdated.ModTime().After(infoAfter.ModTime()) {
		t.Fatalf("expected changed dump to update modtime, before=%v after=%v",
			infoAfter.ModTime(), infoUpdated.ModTime())
	}
}

func TestLoadOffsetRejectsUnsupportedVersion(t *testing.T) {
	offset := OffsetGuessC{offset_sk_num: 14}
	str, err := dumpOffset(offset)
	if err != nil {
		t.Fatalf("dumpOffset: %v", err)
	}

	str = strings.Replace(str,
		fmt.Sprintf("\"version\":\"%s\"", offsetCacheVersion),
		"\"version\":\"1\"",
		1,
	)

	if _, err := loadOffset(str); err == nil || !strings.Contains(err.Error(), "unsupported offset cache version") {
		t.Fatalf("expected version rejection, got %v", err)
	}
}

func TestLoadOffsetRejectsKernelMismatch(t *testing.T) {
	offset := OffsetGuessC{offset_sk_num: 14}
	str, err := dumpOffset(offset)
	if err != nil {
		t.Fatalf("dumpOffset: %v", err)
	}

	kernelVersion, err := currentOffsetCacheKernelVersion()
	if err != nil {
		t.Fatalf("currentOffsetCacheKernelVersion: %v", err)
	}

	replacement := "\"kernel_version\":\"0xdeadbeef\""
	current := fmt.Sprintf("\"kernel_version\":\"%s\"", kernelVersion)
	str = strings.Replace(str, current, replacement, 1)

	if _, err := loadOffset(str); err == nil || !strings.Contains(err.Error(), "offset cache kernel mismatch") {
		t.Fatalf("expected kernel mismatch rejection, got %v", err)
	}
}
