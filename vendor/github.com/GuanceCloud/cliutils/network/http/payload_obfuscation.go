// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import "errors"

const (
	// PayloadObfuscationHeader is the HTTP header name carrying the payload obfuscation mode.
	PayloadObfuscationHeader = "X-Payload-Obfuscation"
	// PayloadObfuscationGzipCaesarV1 identifies a mode that adds 10 modulo 256 to each of the first 256 bytes
	// of a gzip-compressed payload, or to all bytes when the payload is shorter. The transformation is in place
	// and allocation-free. It is obfuscation, not encryption, and does not protect against MITM attacks or tampering.
	PayloadObfuscationGzipCaesarV1 = "gzip-caesar-v1"
)

// ErrUnsupportedPayloadObfuscation is returned when the requested mode is not supported.
// In this case, ObfuscatePayload and DeobfuscatePayload leave body unchanged.
var ErrUnsupportedPayloadObfuscation = errors.New("unsupported payload obfuscation")

// ObfuscatePayload obfuscates body in place using mode.
// It does not compress body.
func ObfuscatePayload(mode string, body []byte) error {
	return transformPayload(mode, body, false)
}

// DeobfuscatePayload reverses ObfuscatePayload in place using mode.
// It does not validate or decompress gzip data.
func DeobfuscatePayload(mode string, body []byte) error {
	return transformPayload(mode, body, true)
}

func transformPayload(mode string, body []byte, reverse bool) error {
	const (
		PAYLOAD_OBFUSCATION_PREFIX_LENGTH = 256
		PAYLOAD_OBFUSCATION_CAESAR_SHIFT  = 10
	)

	if mode != PayloadObfuscationGzipCaesarV1 {
		return ErrUnsupportedPayloadObfuscation
	}

	for i := range body[:min(len(body), PAYLOAD_OBFUSCATION_PREFIX_LENGTH)] {
		if reverse {
			body[i] -= PAYLOAD_OBFUSCATION_CAESAR_SHIFT
		} else {
			body[i] += PAYLOAD_OBFUSCATION_CAESAR_SHIFT
		}
	}

	return nil
}
