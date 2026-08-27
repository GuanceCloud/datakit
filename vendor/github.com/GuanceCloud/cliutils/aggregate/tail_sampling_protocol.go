// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package aggregate

const (
	// TailSamplingPayloadCompressionHeader negotiates compression used inside
	// DataPacket.points_payload. It does not describe HTTP body compression.
	TailSamplingPayloadCompressionHeader = "Guance-Tail-Sampling-Payload-Compression"

	// TailSamplingPayloadCompressionZstd allows packets in the request to carry
	// zstd-compressed points_payload values.
	TailSamplingPayloadCompressionZstd = "zstd"
)

// SetDataPacketPayloadCompression converts a packet payload to the requested
// representation. Requesting zstd may keep a small payload uncompressed when
// compression would not reduce its size.
func SetDataPacketPayloadCompression(packet *DataPacket, compression int32) error {
	if packet == nil {
		return nil
	}

	switch compression {
	case PayloadCompressionNone:
		payload, err := DecompressPointsPayload(packet.PointsPayload, packet.PayloadCompression)
		if err != nil {
			return err
		}
		packet.PointsPayload = payload
		packet.PayloadCompression = PayloadCompressionNone
		return nil

	case PayloadCompressionZstd:
		switch packet.PayloadCompression {
		case PayloadCompressionZstd:
			return nil
		case PayloadCompressionNone:
			payload, method, err := CompressPointsPayload(packet.PointsPayload)
			if err != nil {
				return err
			}
			packet.PointsPayload = payload
			packet.PayloadCompression = method
			return nil
		default:
			return unsupportedPayloadCompressionError(packet.PayloadCompression)
		}

	default:
		return unsupportedPayloadCompressionError(compression)
	}
}
