// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dataway

import (
	"errors"
)

var (
	errWritePoints4XX    = errors.New("write point 4xx")
	errRequestTerminated = errors.New("no response and request maybe terminated")
)
