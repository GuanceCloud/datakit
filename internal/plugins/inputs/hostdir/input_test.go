// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostdir

import (
	"os"
	"runtime"
	"testing"

	"github.com/GuanceCloud/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

func TestInput_Collect(t *testing.T) {
	str, _ := os.Getwd()
	i := Input{
		Dir:      str,
		platform: runtime.GOOS,
		semStop:  cliutils.NewSem(),
		feeder:   io.DefaultFeeder(),
		Tagger:   datakit.DefaultGlobalTagger(),
	}
	if err := i.collect(); err != nil {
		t.Error(err)
	}
	t.Log(i.collectCache)
}
