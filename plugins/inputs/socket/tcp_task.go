// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package socket

import (
	"time"
)

const (
	StatusStop = "stop"
	MaxMsgSize = 15 * 1024 * 1024
	ClassTCP   = "TCP"
)

type Task interface {
	ID() string
	Status() string
	Run() error
	Init() error
	Class() string
	GetResults() (map[string]string, map[string]interface{})

	Ticker() *time.Ticker
}
