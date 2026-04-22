//go:build linux && vmlinux
// +build linux,vmlinux

package offset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOffsetsFromBTF(t *testing.T) {
	kernel, source, err := GetKernelOffsetGuessFromBTF()
	if err != nil {
		t.Skipf("kernel BTF not available: %v", err)
		return
	}
	t.Logf("kernel BTF source: %s", source)

	assert.NotNil(t, kernel)
	assert.True(t, kernelOffsetsReady(kernel, false, false))

	httpOffset, source, err := GetHTTPFlowOffsetFromBTF()
	if err != nil {
		t.Skipf("httpflow offsets from BTF unavailable: %v", err)
		return
	}
	t.Logf("httpflow BTF source: %s", source)
	assert.NotZero(t, int32(httpOffset.offset_task_struct_files))
	assert.NotZero(t, int32(httpOffset.offset_files_struct_fdt))
	assert.NotZero(t, int32(httpOffset.offset_socket_file))
	assert.NotZero(t, int32(httpOffset.offset_file_private_data))

	conntrackOffset, source, err := GetConntrackOffsetFromBTF()
	if err != nil {
		t.Skipf("conntrack offsets from BTF unavailable: %v", err)
		return
	}
	t.Logf("conntrack BTF source: %s", source)
	assert.True(t, conntrackOffsetsReady(conntrackOffset))

	if _, seqOffset, err := TryGetTCPSeqOffsetFromBTF(); err != nil {
		t.Skipf("tcp seq offsets from BTF unavailable: %v", err)
	} else {
		assert.NotZero(t, int32(seqOffset.offset_copied_seq))
		assert.NotZero(t, int32(seqOffset.offset_write_seq))
	}
}
