// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package otlp

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"strconv"
)

// IsZeroID reports whether id is empty or contains only zero bytes.
func IsZeroID(id []byte) bool {
	if len(id) == 0 {
		return true
	}

	for _, b := range id {
		if b != 0 {
			return false
		}
	}

	return true
}

// HexID converts an OTLP binary ID into its hex string form.
func HexID(id []byte) string {
	if IsZeroID(id) {
		return "0"
	}

	return hex.EncodeToString(id)
}

// Base64RawID converts an OTLP binary ID into raw base64 without padding.
func Base64RawID(id []byte) string {
	return base64.RawStdEncoding.EncodeToString(id)
}

// DecimalIDFromLast8Bytes converts the last 8 bytes of id as a big-endian uint64.
func DecimalIDFromLast8Bytes(id []byte) (string, bool) {
	if len(id) < 8 {
		return "", false
	}

	num := binary.BigEndian.Uint64(id[len(id)-8:])

	return strconv.FormatUint(num, 10), true
}
