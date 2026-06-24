// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ingestioncanary

import (
	"testing"

	"github.com/stretchr/testify/assert"
	canary "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ingestion_canary"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

func TestInput_InitFeedFuncs(t *testing.T) {
	ipt := defaultInput()
	ipt.Categories = []string{"metric", "logging"}

	ipt.initFeedFuncs()

	assert.NotNil(t, ipt.canary)
	assert.Len(t, ipt.feedFuncs, 2)
}

func TestInput_BuildWriteURL(t *testing.T) {
	ipt := defaultInput()
	ipt.ResultWorkspace = "https://openway.example.com?token=xxx"

	url, err := ipt.buildWriteURL()

	assert.NoError(t, err)
	assert.Contains(t, url, "/v1/write/metric")
	assert.Contains(t, url, "token=xxx")
}

func TestInput_BuildResultPointAddsInputTag(t *testing.T) {
	ipt := defaultInput()
	ipt.ResultWorkspace = "https://openway.example.com?token=xxx"

	pt := ipt.buildResultPoint(canary.MetricCategory, "ok", "", 1)

	assert.Equal(t, dkio.InputSourceTagValue(inputName), pt.GetTag(dkio.InputSourceTagKey))
}
