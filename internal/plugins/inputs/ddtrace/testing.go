// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ddtrace

import (
	"net/http"

	"github.com/GuanceCloud/cliutils/logger"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	itrace "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
)

func NewTestInput(feeder dkio.Feeder) *Input {
	ipt := defaultInput()
	ipt.feeder = feeder
	ipt.customTagsX = itrace.NewCustomTags([]string{}, ddTags)
	return ipt
}

func (ipt *Input) HandleTracesForTest(w http.ResponseWriter, r *http.Request) {
	if ipt.customTagsX == nil {
		ipt.customTagsX = itrace.NewCustomTags([]string{}, ddTags)
	}
	ipt.handleDDTraces(w, r)
}

func SetAfterGatherForTest(feeder dkio.Feeder) func() {
	prev := afterGatherRun
	afterGatherRun = itrace.NewAfterGather(
		itrace.WithLogger(logger.DefaultSLogger(inputName)),
		itrace.WithPointOptions(),
		itrace.WithFeeder(feeder),
	)
	return func() {
		afterGatherRun = prev
	}
}
