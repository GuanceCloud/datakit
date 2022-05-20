// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package io

import (
	"encoding/json"
	"fmt"
	"time"
)

var defPipelinePullMock pipelinePullMock = &prodPipelinePullMock{}

type pipelinePullMock interface {
	getPipelinePull(ts int64) (*pullPipelineReturn, error)
}

type prodPipelinePullMock struct{}

//------------------------------------------------------------------------------
// copied from kodo project

type dataways struct {
	Regions    map[string][]string `json:"regions"`
	UpdateTime int64               `json:"update_time"` // not used
}

type pipelineUnit struct {
	Name       string `json:"name"`
	Base64Text string `json:"base64text"`
	Category   string `json:"category"`
}

type pullPipelineReturn struct {
	UpdateTime int64           `json:"update_time"`
	Pipelines  []*pipelineUnit `json:"pipelines"`
}

type pullResp struct {
	Filters      map[string][]string `json:"filters"`
	RemotePL     *pullPipelineReturn `json:"remote_pipelines"`
	Dataways     *dataways           `json:"dataways"`
	PullInterval time.Duration       `json:"pull_interval"`
}

//------------------------------------------------------------------------------

func (*prodPipelinePullMock) getPipelinePull(ts int64) (*pullPipelineReturn, error) {
	body, err := defaultIO.dw.DatakitPull(fmt.Sprintf("remote_pipelines=true&ts=%d", ts))
	if err != nil {
		return nil, err
	}
	var br pullResp
	err = json.Unmarshal(body, &br)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return br.RemotePL, err
}
