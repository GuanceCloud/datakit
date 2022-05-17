// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostdir

import (
	"os"
	"runtime"
	"testing"
)

func TestInput_Collect(t *testing.T) {
	str, _ := os.Getwd()
	i := Input{Dir: str, platform: runtime.GOOS}
	if err := i.Collect(); err != nil {
		t.Error(err)
	}
	t.Log(i.collectCache[0])
}
