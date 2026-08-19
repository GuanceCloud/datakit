package aggregate

import (
	"fmt"
	"sync"

	"github.com/klauspost/compress/zstd"
)

const (
	// PayloadCompressionNone means points_payload is raw PBPoints encoding (uncompressed).
	PayloadCompressionNone int32 = 0
	// PayloadCompressionZstd means points_payload is zstd-compressed PBPoints encoding.
	PayloadCompressionZstd int32 = 1
)

// zstdEncoderPool reuses zstd encoders to avoid re-initialization on hot paths.
var zstdEncoderPool = sync.Pool{
	New: func() any {
		enc, err := zstd.NewWriter(nil)
		if err != nil {
			panic(fmt.Sprintf("new zstd encoder: %v", err))
		}
		return enc
	},
}

// zstdDecoderPool reuses zstd decoders.
var zstdDecoderPool = sync.Pool{
	New: func() any {
		dec, err := zstd.NewReader(nil)
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
		return nil, fmt.Errorf("unsupported payload compression: %d", compression)
	}

	dec := zstdDecoderPool.Get().(*zstd.Decoder)
	decompressed, err := dec.DecodeAll(payload, nil)
	zstdDecoderPool.Put(dec)
	if err != nil {
		return nil, fmt.Errorf("zstd decode points payload: %w", err)
	}

	return decompressed, nil
}
