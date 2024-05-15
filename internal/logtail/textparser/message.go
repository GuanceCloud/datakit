// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package textparser wraps text parsing functions
package textparser

type LogStreamType string

const (
	Stdout LogStreamType = "stdout"
	Stderr LogStreamType = "stderr"
)

type LogMessage struct {
	// Timestamp time.Time
	Stream LogStreamType
	// If text parsing does not exist, the Log is nil
	Log       []byte
	IsPartial bool
}
