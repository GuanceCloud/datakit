// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows || arm || 386
// +build windows arm 386

package profile

import "fmt"

// run pyroscope profiling.
func (pyrs *pyroscopeOpts) run(i *Input) error {
	return fmt.Errorf("server mode is not supported on Windows")
}
