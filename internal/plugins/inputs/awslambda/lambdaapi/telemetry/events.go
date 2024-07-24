// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

// Package telemetry aws telemetry api related operations
package telemetry

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"
	"unsafe"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/awslambda/lambdaapi/model"
)

// TelemetryType constants for different types of telemetry records.
const (
	TypeFunctionLog                   = "function"
	TypeExtensionLog                  = "extension"
	TypePlatformInitStart             = "platform.initStart"
	TypePlatformInitRuntimeDone       = "platform.initRuntimeDone"
	TypePlatformInitReport            = "platform.initReport"
	TypePlatformRestoreRuntimeDone    = "platform.restoreRuntimeDone"
	TypePlatformRestoreReport         = "platform.restoreReport"
	TypePlatformRestoreStart          = "platform.restoreStart"
	TypePlatformStart                 = "platform.start"
	TypePlatformRuntimeDone           = "platform.runtimeDone"
	TypePlatformReport                = "platform.report"
	TypePlatformExtension             = "platform.extension"
	TypePlatformTelemetrySubscription = "platform.telemetrySubscription"
	TypePlatformLogsDropped           = "platform.logsDropped"
)

// Event is the payload received from the Telemetry API.
type Event struct {
	Time   time.Time    `json:"time"`
	Record LambdaRecord `json:"record"`
}

type LambdaLog struct {
	Fields map[string]any
}

// FunctionLog represents function log records.
type FunctionLog struct {
	LambdaLog
}

// ExtensionLog represents extension log records.
type ExtensionLog struct {
	LambdaLog
}

// PlatformInitStart represents platform init start record.
type PlatformInitStart struct {
	FunctionName       string                `json:"functionName,omitempty"`
	FunctionVersion    string                `json:"functionVersion,omitempty"`
	InitializationType model.InitType        `json:"initializationType"`
	InstanceID         string                `json:"instanceId,omitempty"`
	InstanceMaxMemory  uint32                `json:"instanceMaxMemory,omitempty"`
	Phase              model.InitPhase       `json:"phase"`
	RuntimeVersion     string                `json:"runtimeVersion,omitempty"`
	RuntimeVersionArn  string                `json:"runtimeVersionArn,omitempty"`
	Tracing            *model.TracingContext `json:"tracing,omitempty"`
}

// PlatformInitRuntimeDone represents platform init runtime done record.
type PlatformInitRuntimeDone struct {
	InitializationType model.InitType        `json:"initializationType"`
	Phase              model.InitPhase       `json:"phase"`
	Status             model.Status          `json:"status"`
	ErrorType          string                `json:"errorType,omitempty"`
	Tracing            *model.TracingContext `json:"tracing,omitempty"`
	Spans              []model.Span          `json:"spans,omitempty"`
}

// PlatformInitReport represents platform init start record.
type PlatformInitReport struct {
	InitializationType model.InitType          `json:"initializationType"`
	Phase              model.InitPhase         `json:"phase"`
	Status             model.Status            `json:"status"`
	ErrorType          string                  `json:"errorType,omitempty"`
	Metrics            model.InitReportMetrics `json:"metrics"`
	Tracing            *model.TracingContext   `json:"tracing,omitempty"`
	Spans              []model.Span            `json:"spans,omitempty"`
}

type PlatformRestoreRuntimeDone struct {
	Status    model.Status          `json:"status"`
	ErrorType string                `json:"errorType,omitempty"`
	Tracing   *model.TracingContext `json:"tracing,omitempty"`
	Spans     []model.Span          `json:"spans,omitempty"`
}

type PlatformRestoreReport struct {
	Status    model.Status               `json:"status"`
	ErrorType string                     `json:"errorType,omitempty"`
	Metrics   model.RestoreReportMetrics `json:"metrics"`
	Tracing   *model.TracingContext      `json:"tracing,omitempty"`
	Spans     []model.Span               `json:"spans,omitempty"`
}

type PlatformRestoreStart struct {
	FunctionName      string                `json:"functionName,omitempty"`
	FunctionVersion   string                `json:"functionVersion,omitempty"`
	InstanceID        string                `json:"instanceId,omitempty"`
	InstanceMaxMemory uint32                `json:"instanceMaxMemory,omitempty"`
	RuntimeVersion    string                `json:"runtimeVersion,omitempty"`
	RuntimeVersionArn string                `json:"runtimeVersionArn,omitempty"`
	Tracing           *model.TracingContext `json:"tracing,omitempty"`
}

// PlatformStart represents record marking start of an invocation.
type PlatformStart struct {
	RequestID string                `json:"requestId"`
	Version   string                `json:"version,omitempty"`
	Tracing   *model.TracingContext `json:"tracing,omitempty"`
}

// PlatformRuntimeDone represents record marking the completion of an invocation.
type PlatformRuntimeDone struct {
	RequestID string                    `json:"requestId"`
	Status    model.Status              `json:"status"`
	ErrorType string                    `json:"errorType,omitempty"`
	Metrics   *model.RuntimeDoneMetrics `json:"metrics,omitempty"`
	Tracing   *model.TracingContext     `json:"tracing,omitempty"`
	Spans     []model.Span              `json:"spans,omitempty"`
}

// PlatformReport represents platform report record.
type PlatformReport struct {
	RequestID string                `json:"requestId"`
	Status    model.Status          `json:"status"`
	ErrorType string                `json:"errorType,omitempty"`
	Metrics   model.ReportMetrics   `json:"metrics"`
	Tracing   *model.TracingContext `json:"tracing,omitempty"`
	Spans     []model.Span          `json:"spans,omitempty"`
}

// PlatformExtension represents extension-specific record.
type PlatformExtension struct {
	Name      string   `json:"name"`
	State     string   `json:"state"`
	ErrorType string   `json:"errorType,omitempty"`
	Events    []string `json:"events"`
}

// PlatformTelemetrySubscription represents telemetry processor-specific record.
type PlatformTelemetrySubscription struct {
	Name  string                        `json:"name"`
	State string                        `json:"state"`
	Types []model.SubscriptionEventType `json:"types"`
}

// PlatformLogsDropped represents record generated when the telemetry processor is falling behind.
type PlatformLogsDropped struct {
	DroppedBytes   uint64 `json:"droppedBytes"`
	DroppedRecords uint64 `json:"droppedRecords"`
	Reason         string `json:"reason"`
}

func (r *Event) UnmarshalJSON(data []byte) error {
	// Temporary structure to decode the JSON to extract the type and record fields
	temp := struct {
		Time   time.Time       `json:"time"`
		Type   string          `json:"type"`
		Record json.RawMessage `json:"record"`
	}{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	var err error
	r.Time = temp.Time
	r.Record, err = CreateRecord(temp.Type)
	if err != nil {
		return err
	}
	return json.Unmarshal(temp.Record, &r.Record)
}

func (l *LambdaLog) UnmarshalJSON(data []byte) error {
	l.Fields = map[string]any{}

	compactBuffer := new(bytes.Buffer)
	err := json.Compact(compactBuffer, data)
	if err != nil {
		return err
	}
	p := compactBuffer.Bytes()
	if len(p) >= 2 && p[0] == p[len(p)-1] && p[0] == '"' {
		p = p[1 : len(p)-1]
		l.Fields["message"] = *(*string)(unsafe.Pointer(&p)) //nolint:gosec
	} else {
		err = json.Unmarshal(data, &l.Fields)
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateRecord(typ string) (record LambdaRecord, err error) {
	switch typ {
	case TypeFunctionLog:
		record = &FunctionLog{}
	case TypeExtensionLog:
		record = &ExtensionLog{}
	case TypePlatformInitStart:
		record = &PlatformInitStart{}
	case TypePlatformInitRuntimeDone:
		record = &PlatformInitRuntimeDone{}
	case TypePlatformInitReport:
		record = &PlatformInitReport{}
	case TypePlatformRestoreStart:
		record = &PlatformRestoreStart{}
	case TypePlatformRestoreRuntimeDone:
		record = &PlatformRestoreRuntimeDone{}
	case TypePlatformRestoreReport:
		record = &PlatformRestoreReport{}
	case TypePlatformStart:
		record = &PlatformStart{}
	case TypePlatformRuntimeDone:
		record = &PlatformRuntimeDone{}
	case TypePlatformReport:
		record = &PlatformReport{}
	case TypePlatformExtension:
		record = &PlatformExtension{}
	case TypePlatformTelemetrySubscription:
		record = &PlatformTelemetrySubscription{}
	case TypePlatformLogsDropped:
		record = &PlatformLogsDropped{}
	default:
		err = errors.New("unknown type: " + typ)
		l.Warn(err)
	}
	return record, err
}

type LambdaRecord interface {
	GetType() string
}

func (l *LambdaLog) GetFields() map[string]any {
	return l.Fields
}

func (fl *FunctionLog) GetType() string                    { return TypeFunctionLog }
func (el *ExtensionLog) GetType() string                   { return TypeExtensionLog }
func (pis *PlatformInitStart) GetType() string             { return TypePlatformInitStart }
func (pird *PlatformInitRuntimeDone) GetType() string      { return TypePlatformInitRuntimeDone }
func (pir *PlatformInitReport) GetType() string            { return TypePlatformInitReport }
func (prs *PlatformRestoreStart) GetType() string          { return TypePlatformRestoreStart }
func (prrd *PlatformRestoreRuntimeDone) GetType() string   { return TypePlatformRestoreRuntimeDone }
func (prr *PlatformRestoreReport) GetType() string         { return TypePlatformRestoreReport }
func (ps *PlatformStart) GetType() string                  { return TypePlatformStart }
func (prd *PlatformRuntimeDone) GetType() string           { return TypePlatformRuntimeDone }
func (pr *PlatformReport) GetType() string                 { return TypePlatformReport }
func (pe *PlatformExtension) GetType() string              { return TypePlatformExtension }
func (pts *PlatformTelemetrySubscription) GetType() string { return TypePlatformTelemetrySubscription }
func (pld *PlatformLogsDropped) GetType() string           { return TypePlatformLogsDropped }

type LambdaLogInterface interface {
	GetFields() map[string]any
	LambdaRecord
}

type LogEvent struct {
	Time   time.Time          `json:"time"`
	Record LambdaLogInterface `json:"record"`
}

func SeparateEvents(events []*Event) (filteredEvents []*Event, logEvents []*LogEvent) {
	logEvents = make([]*LogEvent, 0, len(events)) // 预先分配容量以减少内存分配次数
	filteredEvents = make([]*Event, 0, len(events))

	for _, e := range events {
		if r, ok := e.Record.(LambdaLogInterface); ok {
			logEvents = append(logEvents, &LogEvent{
				Time:   e.Time,
				Record: r,
			})
		} else {
			filteredEvents = append(filteredEvents, e)
		}
	}

	return filteredEvents, logEvents
}
