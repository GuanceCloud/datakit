// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package windowsremote

import (
	"github.com/GuanceCloud/cliutils/point"
)

type Wmi struct {
	cfg *WmiConfig //nolint
}

func newWmi(cfg *WmiConfig) *Wmi { //nolint
	return nil
}

func (w *Wmi) CollectMetric(ip string, timestamp int64) []*point.Point {
	return []*point.Point{}
}

func (w *Wmi) CollectObject(ip string) []*point.Point {
	return []*point.Point{}
}

func (w *Wmi) CollectLogging(ip string) []*point.Point {
	return []*point.Point{}
}

func (w *Wmi) Name() string {
	return "wmi"
}
