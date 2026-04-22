//go:build linux
// +build linux

package offset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKernelBTFCandidatePathsForRelease(t *testing.T) {
	paths := kernelBTFCandidatePathsForRelease("6.8.0-test")
	assert.NotEmpty(t, paths)
	assert.Equal(t, "/sys/kernel/btf/vmlinux", paths[0])
	assert.Contains(t, paths, "/boot/vmlinux-6.8.0-test")
	assert.Contains(t, paths, "/usr/lib/modules/6.8.0-test/vmlinux")
	assert.Contains(t, paths, "/usr/lib/debug/boot/vmlinux-6.8.0-test")
}

func TestKernelBTFDisabled(t *testing.T) {
	t.Setenv(DisableKernelBTFEnv, "1")
	assert.True(t, kernelBTFDisabled())

	_, err := FindKernelBTF()
	assert.ErrorContains(t, err, DisableKernelBTFEnv)

	_, _, err = LoadKernelBTF()
	assert.ErrorContains(t, err, DisableKernelBTFEnv)
}

func TestKernelOffsetsReadyRequiresTCPSeqWhenRequested(t *testing.T) {
	offset := &OffsetGuessC{
		offset_sk_num:          1,
		offset_inet_sport:      2,
		offset_sk_family:       3,
		offset_sk_rcv_saddr:    4,
		offset_sk_dport:        5,
		offset_tcp_sk_srtt_us:  6,
		offset_tcp_sk_mdev_us:  7,
		offset_flowi4_saddr:    8,
		offset_flowi4_daddr:    9,
		offset_flowi4_sport:    10,
		offset_flowi4_dport:    11,
		offset_sk_net:          12,
		offset_ns_common_inum:  13,
		offset_socket_sk:       14,
		offset_sk_v6_rcv_saddr: 15,
		offset_sk_v6_daddr:     16,
		offset_flowi6_saddr:    17,
		offset_flowi6_daddr:    18,
		offset_flowi6_sport:    19,
		offset_flowi6_dport:    20,
	}

	assert.True(t, kernelOffsetsReady(offset, false, false))
	assert.False(t, kernelOffsetsReady(offset, true, false))

	offset.offset_copied_seq = 21
	offset.offset_write_seq = 22
	assert.True(t, kernelOffsetsReady(offset, true, false))
}

func TestConntrackOffsetsReady(t *testing.T) {
	offset := &OffsetConntrackC{}
	assert.False(t, conntrackOffsetsReady(offset))

	offset.offset_ct_origin_tuple = 1
	offset.offset_ct_reply_tuple = 2
	offset.offset_ct_net = 3
	offset.offset_ct_ns_common_inum = 4
	assert.True(t, conntrackOffsetsReady(offset))
}

func TestKernelGuessNeedsSelectiveProbes(t *testing.T) {
	assert.True(t, kernelGuessNeedsTCP4(nil))
	assert.True(t, kernelGuessNeedsUDP4(nil))
	assert.True(t, kernelGuessNeedsTCP6(nil))
	assert.True(t, kernelGuessNeedsSocket(nil))

	offset := &OffsetGuessC{
		offset_inet_sport:      1,
		offset_sk_dport:        2,
		offset_tcp_sk_srtt_us:  3,
		offset_tcp_sk_mdev_us:  4,
		offset_sk_daddr:        5,
		offset_sk_net:          6,
		offset_ns_common_inum:  7,
		offset_sk_family:       8,
		offset_flowi4_daddr:    9,
		offset_flowi4_saddr:    10,
		offset_flowi4_dport:    11,
		offset_flowi4_sport:    12,
		offset_sk_v6_daddr:     13,
		offset_sk_v6_rcv_saddr: 14,
		offset_socket_sk:       15,
	}

	assert.False(t, kernelGuessNeedsTCP4(offset))
	assert.False(t, kernelGuessNeedsUDP4(offset))
	assert.False(t, kernelGuessNeedsTCP6(offset))
	assert.False(t, kernelGuessNeedsSocket(offset))

	offset.offset_socket_sk = 0
	assert.True(t, kernelGuessNeedsSocket(offset))
	assert.False(t, kernelGuessNeedsTCP4(offset))
}
