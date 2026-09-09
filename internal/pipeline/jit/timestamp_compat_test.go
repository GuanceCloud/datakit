// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package jit

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
)

var (
	timestampBenchmarkNanoseconds int64
	timestampBenchmarkOutput      []byte
	timestampBenchmarkError       error
)

func TestHostCompatTimestampMatchesPipelineGoCorpus(t *testing.T) {
	host := NewPipelineGoHost(nil)
	pointTime := int64(1_700_000_000_000_000_123)
	dateparseVectors := make([]timestampTestVector, len(dateparseTimestampTestVectors))
	for index, vector := range dateparseTimestampTestVectors {
		dateparseVectors[index] = normalizedDateparseTimestampVector(vector)
	}
	vectors := append(append([]timestampTestVector(nil), dateparseVectors...),
		pipelineGoTimezoneTestVectors...)
	for index, vector := range vectors {
		t.Run(fmt.Sprintf("%03d", index), func(t *testing.T) {
			expected, expectedError := funcs.TimestampHandle(vector.value, vector.timezone)
			if index < len(dateparseVectors) && (expectedError != nil) != vector.shouldFail {
				t.Fatalf("vendored dateparse corpus expectation changed for %q: %v", vector.value, expectedError)
			}
			output, err := host.invoke(hostCompatTimestampOp,
				encodeHostCompatTimestamp(vector.value, vector.timezone, pointTime))
			if err != nil {
				t.Fatalf("host protocol failed for value=%q timezone=%q: %v", vector.value, vector.timezone, err)
			}
			if expectedError == nil {
				if len(output) != 9 || output[0] != 0 {
					t.Fatalf("success response = %v", output)
				}
				if actual := int64(binary.LittleEndian.Uint64(output[1:])); actual != expected {
					t.Fatalf("timestamp = %d, want pipeline-go %d", actual, expected)
				}
				return
			}
			if len(output) < 5 || output[0] != 1 {
				t.Fatalf("error response = %v", output)
			}
			reader := hostCompatReader{data: output[1:]}
			actualError, err := reader.stringValue()
			if err != nil || !reader.done() {
				t.Fatalf("decode error response: %q, %v", actualError, err)
			}
			if actualError != expectedError.Error() {
				t.Fatalf("error = %q, want pipeline-go %q", actualError, expectedError.Error())
			}
		})
	}
}

func BenchmarkTimestampHostCompat(b *testing.B) {
	const (
		value     = "2026-05-19T13:47:01.004+0800"
		pointTime = int64(1_700_000_000_000_000_123)
	)
	host := NewPipelineGoHost(nil)
	request := encodeHostCompatTimestamp(value, "", pointTime)
	b.Run("TimestampHandle", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			timestampBenchmarkNanoseconds, timestampBenchmarkError = funcs.TimestampHandle(value, "")
		}
		if timestampBenchmarkError != nil {
			b.Fatal(timestampBenchmarkError)
		}
	})
	b.Run("HostProtocol", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			timestampBenchmarkOutput, timestampBenchmarkError = host.invoke(hostCompatTimestampOp, request)
		}
		if timestampBenchmarkError != nil {
			b.Fatal(timestampBenchmarkError)
		}
	})
}
