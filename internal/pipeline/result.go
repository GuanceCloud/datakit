// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package pipeline

import (
	"time"

	"github.com/GuanceCloud/pipeline-go/constants"
)

const (
	// pipeline关键字段.
	FieldTime       = constants.FieldTime
	FieldMessage    = constants.FieldMessage
	FieldStatus     = constants.FieldStatus
	PlLoggingSource = constants.PlLoggingSource

	DefaultStatus = constants.DefaultStatus
)

//nolint:structcheck,unused
type Output struct {
	Error error

	Drop bool

	Measurement string
	Time        time.Time

	Tags   map[string]string
	Fields map[string]interface{}
}

type Result struct {
	Output *Output
}

func NewResult() *Result {
	return &Result{
		Output: &Output{
			Tags:   make(map[string]string),
			Fields: make(map[string]interface{}),
		},
	}
}
