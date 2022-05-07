// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import "gitlab.jiagouyun.com/cloudcare-tools/datakit/io/dataway"

var defPipelinePullMock pipelinePullMock = &prodPipelinePullMock{}

type pipelinePullMock interface {
	getPipelinePull(ts int64) (*dataway.PullPipelineReturn, error)
}

type prodPipelinePullMock struct{}

func (*prodPipelinePullMock) getPipelinePull(ts int64) (*dataway.PullPipelineReturn, error) {
	return defaultIO.dw.GetPipelinePull(ts)
}
