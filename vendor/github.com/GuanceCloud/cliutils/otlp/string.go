// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package otlp

// ChunkStringByRuneLength splits s into chunks measured by rune count.
func ChunkStringByRuneLength(s string, size int) []string {
	if size <= 0 || s == "" {
		return nil
	}

	runes := []rune(s)
	chunks := make([]string, 0, (len(runes)+size-1)/size)

	for i := 0; i < len(runes); i += size {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}

	return chunks
}
