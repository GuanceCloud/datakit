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
