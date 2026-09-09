// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package jit

import (
	"encoding/binary"
	"errors"
	"reflect"
	"testing"

	"github.com/GuanceCloud/pipeline-go/ptinput/ipdb"
)

type hostCompatIPDB struct {
	record *ipdb.IPdbRecord
	err    error
}

func (*hostCompatIPDB) Init(string, map[string]string) {}

func (database *hostCompatIPDB) Geo(string) (*ipdb.IPdbRecord, error) {
	return database.record, database.err
}

func (database *hostCompatIPDB) GeoWithChecker(value string, check ipdb.CheckData) (*ipdb.IPdbRecord, error) {
	record, err := database.Geo(value)
	if record != nil && check != nil {
		record = check(record)
	}
	return record, err
}

func (database *hostCompatIPDB) SearchIsp(string) string { return "" }

func TestHostCompatGeoIPProtocol(t *testing.T) {
	host := NewPipelineGoHost(&hostCompatIPDB{record: &ipdb.IPdbRecord{
		City: "Hangzhou", Region: "Zhejiang", Country: "CN", Isp: "example",
	}})
	output, err := host.invoke(hostCompatGeoIPOp, encodeHostCompatStrings("203.0.113.1"))
	if err != nil {
		t.Fatalf("invoke geoip: %v", err)
	}
	if len(output) == 0 || output[0] != 1 {
		t.Fatalf("geoip response kind = %v", output)
	}
	reader := hostCompatReader{data: output[1:]}
	for index, expected := range []string{"Hangzhou", "Zhejiang", "CN", "example"} {
		actual, err := reader.stringValue()
		if err != nil || actual != expected {
			t.Fatalf("geoip value %d = %q, %v; want %q", index, actual, err, expected)
		}
	}
	if !reader.done() {
		t.Fatal("geoip response has trailing bytes")
	}

	host = NewPipelineGoHost(&hostCompatIPDB{err: errors.New("not found")})
	output, err = host.invoke(hostCompatGeoIPOp, encodeHostCompatStrings("203.0.113.2"))
	if err != nil || len(output) != 1 || output[0] != 0 {
		t.Fatalf("geoip domain error response = %v, %v", output, err)
	}
}

func TestHostCompatUserAgentProtocol(t *testing.T) {
	host := NewPipelineGoHost(nil)
	output, err := host.invoke(hostCompatUserAgentOp, encodeHostCompatStrings(
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Version/17.0 Mobile/15E148 Safari/604.1",
	))
	if err != nil {
		t.Fatalf("invoke user_agent: %v", err)
	}
	if len(output) < 3 || output[0] != 1 || output[1] != 1 || output[2] != 0 {
		t.Fatalf("unexpected user-agent response prefix %v", output)
	}
	reader := hostCompatReader{data: output[3:]}
	for range 6 {
		if _, err := reader.stringValue(); err != nil {
			t.Fatalf("decode user-agent response: %v", err)
		}
	}
	if !reader.done() {
		t.Fatal("user-agent response has trailing bytes")
	}
}

func TestHostCompatTimestampProtocol(t *testing.T) {
	host := NewPipelineGoHost(nil)
	request := encodeHostCompatTimestamp("2024-01-02T03:04:05Z", "", 123)
	output, err := host.invoke(hostCompatTimestampOp, request)
	if err != nil {
		t.Fatalf("invoke timestamp: %v", err)
	}
	if len(output) != 9 || output[0] != 0 {
		t.Fatalf("timestamp success response = %v", output)
	}
	if got, want := int64(binary.LittleEndian.Uint64(output[1:])), int64(1704164645000000000); got != want {
		t.Fatalf("timestamp = %d, want %d", got, want)
	}

	output, err = host.invoke(hostCompatTimestampOp, encodeHostCompatTimestamp("invalid", "", 456))
	if err != nil {
		t.Fatalf("timestamp parse error must be a domain response: %v", err)
	}
	if len(output) < 5 || output[0] != 1 {
		t.Fatalf("timestamp parse-error response = %v", output)
	}
	reader := hostCompatReader{data: output[1:]}
	message, err := reader.stringValue()
	if err != nil || message == "" || !reader.done() {
		t.Fatalf("timestamp error = %q, %v", message, err)
	}
}

func TestHostCompatRejectsMalformedProtocol(t *testing.T) {
	host := NewPipelineGoHost(&hostCompatIPDB{})
	for _, test := range []struct {
		operation uint32
		request   []byte
	}{
		{hostCompatGeoIPOp, []byte{1, 0}},
		{hostCompatUserAgentOp, append(encodeHostCompatStrings("ok"), 0)},
		{hostCompatTimestampOp, encodeHostCompatStrings("missing fields")},
		{99, nil},
	} {
		_, err := host.invoke(test.operation, test.request)
		if err == nil || !isHostCompatProtocolError(err) {
			t.Fatalf("operation %d malformed request error = %v", test.operation, err)
		}
	}
}

func TestObservedPipelineGoHostUsesBoundedOperationNames(t *testing.T) {
	var operations []string
	host := NewPipelineGoHostObserved(nil, func(operation string) {
		operations = append(operations, operation)
	})
	for _, operation := range []uint32{
		hostCompatGeoIPOp,
		hostCompatUserAgentOp,
		hostCompatTimestampOp,
		99,
	} {
		host.observeInvoke(operation, 0)
	}
	want := []string{"geoip", "user_agent", "default_time", "unknown"}
	if !reflect.DeepEqual(operations, want) {
		t.Fatalf("observed operations = %v, want %v", operations, want)
	}
}

func TestObservedPipelineGoHostContainsObserverPanic(t *testing.T) {
	host := NewPipelineGoHostObserved(nil, func(string) {
		panic("observer must not escape into the native callback")
	})
	host.observeInvoke(hostCompatTimestampOp, 0)
}

func encodeHostCompatStrings(values ...string) []byte {
	var writer hostCompatWriter
	for _, value := range values {
		if err := writer.stringValue(value); err != nil {
			panic(err)
		}
	}
	return writer.data
}

func encodeHostCompatTimestamp(value, timezone string, pointTime int64) []byte {
	request := make([]byte, 16, 16+len(value)+len(timezone))
	binary.LittleEndian.PutUint32(request[0:4], uint32(len(value)))
	binary.LittleEndian.PutUint32(request[4:8], uint32(len(timezone)))
	binary.LittleEndian.PutUint64(request[8:16], uint64(pointTime))
	request = append(request, value...)
	request = append(request, timezone...)
	return request
}
