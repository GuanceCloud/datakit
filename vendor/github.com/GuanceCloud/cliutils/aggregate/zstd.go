// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package aggregate

import (
	"errors"
	"fmt"
	"sync"

	"github.com/klauspost/compress/zstd"
)

const (
	// PayloadCompressionNone means points_payload is raw PBPoints encoding (uncompressed).
	PayloadCompressionNone int32 = 0
	// PayloadCompressionZstd means points_payload is zstd-compressed PBPoints encoding.
	PayloadCompressionZstd int32 = 1

	zstdDecoderMaxWindowBytes = 64 << 20
	zstdDecoderMaxMemoryBytes = 256 << 20
)

// ErrUnsupportedPayloadCompression indicates that a DataPacket uses a
// points_payload compression method this cliutils version cannot decode.
var ErrUnsupportedPayloadCompression = errors.New("unsupported payload compression")

// ErrPayloadDecodedSizeUnknown indicates that a compressed frame does not
// advertise its decoded size, so callers cannot reserve bounded memory first.
var ErrPayloadDecodedSizeUnknown = errors.New("payload decoded size is unknown")

// ErrPayloadEmpty indicates that a DataPacket has no points payload.
var ErrPayloadEmpty = errors.New("points payload is empty")

func unsupportedPayloadCompressionError(compression int32) error {
	return fmt.Errorf("%w: %d", ErrUnsupportedPayloadCompression, compression)
}

// zstdEncoderPool reuses zstd encoders to avoid re-initialization on hot paths.
var zstdEncoderPool = sync.Pool{
	New: func() any {
		enc, err := zstd.NewWriter(nil,
			zstd.WithEncoderConcurrency(1),
			zstd.WithEncoderLevel(zstd.SpeedFastest),
		)
		if err != nil {
			panic(fmt.Sprintf("new zstd encoder: %v", err))
		}
		return enc
	},
}

// zstdDecoderPool reuses zstd decoders.
var zstdDecoderPool = sync.Pool{
	New: func() any {
		dec, err := zstd.NewReader(nil,
			zstd.WithDecoderConcurrency(1),
			zstd.WithDecoderLowmem(true),
			zstd.WithDecoderMaxWindow(zstdDecoderMaxWindowBytes),
			zstd.WithDecoderMaxMemory(zstdDecoderMaxMemoryBytes),
		)
		if err != nil {
			panic(fmt.Sprintf("new zstd decoder: %v", err))
		}
		return dec
	},
}

// CompressPointsPayload compresses a PBPoints payload.
// It returns the compressed bytes together with the compression method;
// when compression has no benefit (result not smaller than the input),
// the original bytes and PayloadCompressionNone are returned.
func CompressPointsPayload(payload []byte) ([]byte, int32, error) {
	if len(payload) == 0 {
		return payload, PayloadCompressionNone, nil
	}

	enc := zstdEncoderPool.Get().(*zstd.Encoder)
	compressed := enc.EncodeAll(payload, nil)
	zstdEncoderPool.Put(enc)

	if len(compressed) >= len(payload) {
		return payload, PayloadCompressionNone, nil
	}

	return compressed, PayloadCompressionZstd, nil
}

// DecompressPointsPayload decompresses a PBPoints payload by the given compression method.
// PayloadCompressionNone returns the input as-is without copying.
func DecompressPointsPayload(payload []byte, compression int32) ([]byte, error) {
	if len(payload) == 0 || compression == PayloadCompressionNone {
		return payload, nil
	}

	if compression != PayloadCompressionZstd {
		return nil, unsupportedPayloadCompressionError(compression)
	}
	if _, err := decodeZstdSingleFrameHeader(payload); err != nil {
		return nil, err
	}

	dec := zstdDecoderPool.Get().(*zstd.Decoder)
	decompressed, err := dec.DecodeAll(payload, nil)
	zstdDecoderPool.Put(dec)
	if err != nil {
		return nil, fmt.Errorf("zstd decode points payload: %w", err)
	}

	return decompressed, nil
}

// PointsPayloadDecodedSize returns the number of bytes produced by decoding a
// points_payload without allocating the decoded output. Zstd frames without a
// content-size field are rejected so memory-bounded callers can fail before
// decompression.
func PointsPayloadDecodedSize(payload []byte, compression int32) (int64, error) {
	if len(payload) == 0 || compression == PayloadCompressionNone {
		return int64(len(payload)), nil
	}
	if compression != PayloadCompressionZstd {
		return 0, unsupportedPayloadCompressionError(compression)
	}

	header, err := decodeZstdSingleFrameHeader(payload)
	if err != nil {
		return 0, err
	}
	if !header.HasFCS {
		return 0, ErrPayloadDecodedSizeUnknown
	}
	if header.FrameContentSize > uint64(1<<63-1) {
		return 0, fmt.Errorf("points payload decoded size overflows int64: %d", header.FrameContentSize)
	}
	return int64(header.FrameContentSize), nil
}

func decodeZstdSingleFrameHeader(payload []byte) (zstd.Header, error) {
	var header zstd.Header
	if err := header.Decode(payload); err != nil {
		return zstd.Header{}, fmt.Errorf("decode zstd points payload header: %w", err)
	}
	frameSize, err := zstdFrameSize(payload, &header)
	if err != nil {
		return zstd.Header{}, err
	}
	if frameSize != len(payload) {
		return zstd.Header{}, fmt.Errorf("points payload contains trailing or concatenated zstd data: frame=%d payload=%d",
			frameSize, len(payload))
	}
	return header, nil
}

// zstdFrameSize returns the compressed byte length of the first zstd frame.
// It walks block headers only and does not allocate or decode the frame body.
func zstdFrameSize(payload []byte, header *zstd.Header) (int, error) {
	if header == nil || header.Skippable {
		return 0, errors.New("invalid zstd points payload frame")
	}

	offset := header.HeaderSize
	for {
		if len(payload)-offset < 3 {
			return 0, errors.New("truncated zstd points payload block header")
		}

		blockHeader := uint32(payload[offset]) |
			uint32(payload[offset+1])<<8 |
			uint32(payload[offset+2])<<16
		offset += 3

		lastBlock := blockHeader&1 != 0
		blockType := (blockHeader >> 1) & 3
		blockSize := int(blockHeader >> 3)
		switch blockType {
		case 0, 2: // raw or compressed
		case 1: // RLE stores one byte regardless of the decoded block size.
			blockSize = 1
		default:
			return 0, errors.New("zstd points payload uses a reserved block type")
		}
		if blockSize > len(payload)-offset {
			return 0, errors.New("truncated zstd points payload block")
		}
		offset += blockSize

		if lastBlock {
			break
		}
	}

	if header.HasCheckSum {
		if len(payload)-offset < 4 {
			return 0, errors.New("truncated zstd points payload checksum")
		}
		offset += 4
	}

	return offset, nil
}
