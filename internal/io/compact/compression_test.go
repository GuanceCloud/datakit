// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package compact

import (
	"bytes"
	T "testing"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZstdCompressor(t *T.T) {
	raw := bytes.Repeat([]byte("point-data"), 1024)

	compressor, err := NewCompressor(CompressionZstd)
	require.NoError(t, err)
	compressed, err := compressor.Compress(make([]byte, 1), raw)
	require.NoError(t, err)

	decoder, err := zstd.NewReader(nil)
	require.NoError(t, err)
	defer decoder.Close()
	decoded, err := decoder.DecodeAll(compressed, nil)
	require.NoError(t, err)
	assert.Equal(t, raw, decoded)
}
