// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2026-present Guance, Inc.

package jit

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/GuanceCloud/pipeline-go/ptinput/funcs"
	"github.com/GuanceCloud/pipeline-go/ptinput/ipdb"
	"github.com/GuanceCloud/pipeline-go/ptinput/refertable"
)

const (
	hostCompatABIVersion = 1 //nolint:unused // Used by the pipeline_jit build on Linux.

	hostCompatGeoIPFlag      = 1 << 0
	hostCompatUserAgentFlag  = 1 << 1
	hostCompatTimestampFlag  = 1 << 2
	hostCompatReferTableFlag = 1 << 3
	hostCompatMaximumPayload = 1 << 20

	hostCompatGeoIPOp      = 1
	hostCompatUserAgentOp  = 2
	hostCompatTimestampOp  = 3
	hostCompatReferTableOp = 4
)

// HostCompat exposes pipeline-go operations whose behavior depends on Go-only
// parsers or DataKit state. A HostCompat is immutable and can be shared by all
// programs owned by one Runner.
type HostCompat struct {
	ipdb          ipdb.IPdb
	referTable    refertable.PlReferTables
	flags         uint16
	observeInvoke func(operation uint32, recordIndex uintptr)
}

type HostCallObserver func(operation string)

type hostCompatProtocolError struct {
	message string
}

func (e *hostCompatProtocolError) Error() string { return e.message }

func invalidHostCompatRequest(message string) error {
	return &hostCompatProtocolError{message: message}
}

func isHostCompatProtocolError(err error) bool {
	var protocolError *hostCompatProtocolError
	return errors.As(err, &protocolError)
}

// NewPipelineGoHost snapshots DataKit's current IP database and enables the
// pipeline-go 1.4.3 compatibility operations. A nil database remains a valid
// geoip host because pipeline-go defines that case as a no-op.
func NewPipelineGoHost(database ipdb.IPdb) *HostCompat {
	return NewPipelineGoHostObserved(database, nil)
}

// NewPipelineGoHostObserved constructs an immutable compatibility host whose
// observer is fixed before the host is published to native code. Operation
// names come from a bounded set so callers can safely use them as metric labels.
func NewPipelineGoHostObserved(database ipdb.IPdb, observer HostCallObserver) *HostCompat {
	return NewPipelineGoHostWithReferObserved(database, nil, observer)
}

// NewPipelineGoHostWithReferObserved shares DataKit's process-wide refer-table
// snapshot with every JIT program. It does not create a connection pool per
// script; the table's existing pull/update and locking policy remains the
// single source of truth.
func NewPipelineGoHostWithReferObserved(database ipdb.IPdb, referTable refertable.PlReferTables, observer HostCallObserver) *HostCompat {
	host := &HostCompat{
		ipdb:       database,
		referTable: referTable,
		flags:      hostCompatGeoIPFlag | hostCompatUserAgentFlag | hostCompatTimestampFlag,
	}
	if referTable != nil {
		host.flags |= hostCompatReferTableFlag
	}
	if observer != nil {
		host.observeInvoke = func(operation uint32, _ uintptr) {
			defer func() {
				_ = recover()
			}()
			observer(hostCompatOperationName(operation))
		}
	}
	return host
}

func hostCompatOperationName(operation uint32) string {
	switch operation {
	case hostCompatGeoIPOp:
		return "geoip"
	case hostCompatUserAgentOp:
		return "user_agent"
	case hostCompatTimestampOp:
		return "default_time"
	case hostCompatReferTableOp:
		return "refer_table"
	default:
		return "unknown"
	}
}

func (h *HostCompat) invoke(operation uint32, input []byte) (output []byte, err error) {
	if h == nil {
		return nil, errors.New("JIT host compatibility layer is nil")
	}
	if len(input) > hostCompatMaximumPayload {
		return nil, errors.New("JIT host compatibility request is too large")
	}
	defer func() {
		if value := recover(); value != nil {
			output = nil
			err = fmt.Errorf("JIT host compatibility operation panicked: %v", value)
		}
	}()

	switch operation {
	case hostCompatGeoIPOp:
		if h.flags&hostCompatGeoIPFlag == 0 {
			return nil, errors.New("JIT host geoip operation is disabled")
		}
		return h.invokeGeoIP(input)
	case hostCompatUserAgentOp:
		if h.flags&hostCompatUserAgentFlag == 0 {
			return nil, errors.New("JIT host user-agent operation is disabled")
		}
		return invokeUserAgent(input)
	case hostCompatTimestampOp:
		if h.flags&hostCompatTimestampFlag == 0 {
			return nil, errors.New("JIT host timestamp operation is disabled")
		}
		return invokeTimestamp(input)
	case hostCompatReferTableOp:
		if h.flags&hostCompatReferTableFlag == 0 || h.referTable == nil {
			return nil, errors.New("JIT host refer-table operation is disabled")
		}
		return h.invokeReferTable(input)
	default:
		return nil, invalidHostCompatRequest(fmt.Sprintf("unsupported JIT host compatibility operation %d", operation))
	}
}

type hostReferTableRequest struct {
	Table  string   `json:"table"`
	Keys   []string `json:"keys"`
	Values []any    `json:"values"`
}

func (h *HostCompat) invokeReferTable(input []byte) ([]byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(input))
	decoder.UseNumber()
	var request hostReferTableRequest
	if err := decoder.Decode(&request); err != nil || request.Table == "" || len(request.Keys) != len(request.Values) {
		return nil, invalidHostCompatRequest("invalid JIT host refer-table request")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, invalidHostCompatRequest("trailing JIT host refer-table request data")
	}
	for index, value := range request.Values {
		if number, ok := value.(json.Number); ok {
			if integer, err := number.Int64(); err == nil {
				request.Values[index] = integer
			} else if floating, err := number.Float64(); err == nil {
				request.Values[index] = floating
			} else {
				return nil, invalidHostCompatRequest("invalid numeric refer-table value")
			}
		}
	}
	row, ok := h.referTable.Query(request.Table, request.Keys, request.Values, nil)
	if !ok {
		return []byte("[]"), nil
	}
	return json.Marshal([]map[string]any{row})
}

func (h *HostCompat) invokeGeoIP(input []byte) ([]byte, error) {
	reader := hostCompatReader{data: input}
	ip, err := reader.stringValue()
	if err != nil || !reader.done() {
		return nil, invalidHostCompatRequest("invalid JIT host geoip request")
	}
	values, err := funcs.GeoIPHandle(h.ipdb, ip)
	if err != nil || values == nil {
		return []byte{0}, nil
	}
	writer := hostCompatWriter{data: make([]byte, 1, 1+4*4)}
	writer.data[0] = 1
	for _, key := range []string{"city", "province", "country", "isp"} {
		if err := writer.stringValue(values[key]); err != nil {
			return nil, err
		}
	}
	return writer.data, nil
}

func invokeUserAgent(input []byte) ([]byte, error) {
	reader := hostCompatReader{data: input}
	value, err := reader.stringValue()
	if err != nil || !reader.done() {
		return nil, invalidHostCompatRequest("invalid JIT host user-agent request")
	}
	values, _ := funcs.UserAgentHandle(value)
	writer := hostCompatWriter{data: make([]byte, 3, 3+6*4)}
	writer.data[0] = 1
	if mobile, ok := values["isMobile"].(bool); ok && mobile {
		writer.data[1] = 1
	}
	if bot, ok := values["isBot"].(bool); ok && bot {
		writer.data[2] = 1
	}
	for _, key := range []string{"os", "browser", "browserVer", "engine", "engineVer", "ua"} {
		value, ok := values[key].(string)
		if !ok {
			return nil, fmt.Errorf("pipeline-go user-agent result %q is not a string", key)
		}
		if err := writer.stringValue(value); err != nil {
			return nil, err
		}
	}
	return writer.data, nil
}

func invokeTimestamp(input []byte) ([]byte, error) {
	reader := hostCompatReader{data: input}
	valueLength, err := reader.uint32()
	if err != nil {
		return nil, invalidHostCompatRequest("invalid JIT host timestamp request")
	}
	timezoneLength, err := reader.uint32()
	if err != nil {
		return nil, invalidHostCompatRequest("invalid JIT host timestamp request")
	}
	if _, err := reader.int64(); err != nil { // Rust applies point_time on a parse error.
		return nil, invalidHostCompatRequest("invalid JIT host timestamp request")
	}
	value, err := reader.fixedString(valueLength)
	if err != nil {
		return nil, invalidHostCompatRequest("invalid JIT host timestamp value")
	}
	timezone, err := reader.fixedString(timezoneLength)
	if err != nil || !reader.done() {
		return nil, invalidHostCompatRequest("invalid JIT host timestamp timezone")
	}
	nanoseconds, err := funcs.TimestampHandle(value, timezone)
	if err == nil {
		output := make([]byte, 9)
		output[0] = 0
		binary.LittleEndian.PutUint64(output[1:], uint64(nanoseconds))
		return output, nil
	}
	writer := hostCompatWriter{data: []byte{1}}
	if encodeErr := writer.stringValue(err.Error()); encodeErr != nil {
		return nil, encodeErr
	}
	return writer.data, nil
}

type hostCompatReader struct {
	data   []byte
	offset int
}

func (r *hostCompatReader) uint32() (uint32, error) {
	if len(r.data)-r.offset < 4 {
		return 0, errors.New("truncated uint32")
	}
	value := binary.LittleEndian.Uint32(r.data[r.offset : r.offset+4])
	r.offset += 4
	return value, nil
}

func (r *hostCompatReader) int64() (int64, error) {
	if len(r.data)-r.offset < 8 {
		return 0, errors.New("truncated int64")
	}
	value := int64(binary.LittleEndian.Uint64(r.data[r.offset : r.offset+8]))
	r.offset += 8
	return value, nil
}

func (r *hostCompatReader) stringValue() (string, error) {
	length, err := r.uint32()
	if err != nil {
		return "", err
	}
	return r.fixedString(length)
}

func (r *hostCompatReader) fixedString(length uint32) (string, error) {
	if uint64(length) > uint64(len(r.data)-r.offset) {
		return "", errors.New("truncated string")
	}
	value := r.data[r.offset : r.offset+int(length)]
	r.offset += int(length)
	if !utf8.Valid(value) {
		return "", errors.New("string is not UTF-8")
	}
	return string(value), nil
}

func (r *hostCompatReader) done() bool {
	return r.offset == len(r.data)
}

type hostCompatWriter struct {
	data []byte
}

func (w *hostCompatWriter) stringValue(value string) error {
	if !utf8.ValidString(value) {
		return errors.New("pipeline-go host result is not UTF-8")
	}
	if len(value) > hostCompatMaximumPayload || len(w.data) > hostCompatMaximumPayload-4-len(value) {
		return errors.New("JIT host compatibility response is too large")
	}
	offset := len(w.data)
	w.data = append(w.data, 0, 0, 0, 0)
	binary.LittleEndian.PutUint32(w.data[offset:], uint32(len(value)))
	w.data = append(w.data, value...)
	return nil
}
