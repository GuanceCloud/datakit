// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pprof

import (
	"encoding/json"

	"github.com/GuanceCloud/cliutils/pprofparser/domain/quantity"
)

const (
	FieldQuantity     = "quantity"
	FieldValue        = "value"
	FieldUnit         = "unit"
	FieldPercent      = "percent"
	FieldFunctionName = "functionName"
	FieldLine         = "line"
	FieldFile         = "file"
	FieldDirectory    = "directory"
	FieldThreadID     = "threadID"
	FieldThreadName   = "threadName"
	FieldClass        = "class"
	FieldNamespace    = "namespace"
	FieldAssembly     = "assembly"
	FieldPackage      = "package"
	FieldPrintString  = "printString"
	FieldSubFrames    = "subFrames"
)

func GetFrameJSONFields() []string {
	return []string{
		FieldQuantity,
		FieldValue,
		FieldUnit,
		FieldPercent,
		FieldFunctionName,
		FieldLine,
		FieldFile,
		FieldDirectory,
		FieldThreadID,
		FieldThreadName,
		FieldClass,
		FieldNamespace,
		FieldAssembly,
		FieldPackage,
		FieldPrintString,
		FieldSubFrames,
	}
}

type Frame struct {
	Quantity    *quantity.Quantity `json:"quantity"`
	Value       int64              `json:"value"`
	Unit        *quantity.Unit     `json:"unit"`
	Percent     string             `json:"percent"`
	Function    string             `json:"functionName"`
	Line        int64              `json:"line,omitempty"`
	File        string             `json:"file,omitempty"`
	Directory   string             `json:"directory,omitempty"`
	ThreadID    string             `json:"threadID"`
	ThreadName  string             `json:"threadName"`
	Class       string             `json:"class,omitempty"`
	Namespace   string             `json:"namespace,omitempty"`
	Assembly    string             `json:"assembly,omitempty"`
	Package     string             `json:"package,omitempty"`
	PrintString string             `json:"printString"`
	SubFrames   SubFrames          `json:"subFrames"`
}

func (f *Frame) MarshalJSON() ([]byte, error) {
	toArr := []interface{}{
		f.Quantity,
		f.Value,
		f.Unit,
		f.Percent,
		f.Function,
		f.Line,
		f.File,
		f.Directory,
		f.ThreadID,
		f.ThreadName,
		f.Class,
		f.Namespace,
		f.Assembly,
		f.Package,
		f.PrintString,
		f.SubFrames,
	}
	return json.Marshal(toArr)
}

type SubFrames map[string]*Frame

func (sf SubFrames) MarshalJSON() ([]byte, error) {
	frames := make([]*Frame, 0, len(sf))
	for _, frame := range sf {
		frames = append(frames, frame)
	}
	return json.Marshal(frames)
}
